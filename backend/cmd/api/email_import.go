package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dariobaldi/halendar_back/internal/calendarimport"
	"github.com/dariobaldi/halendar_back/internal/claude"
	"github.com/dariobaldi/halendar_back/internal/data"
	"github.com/dariobaldi/halendar_back/internal/gemini"
	"github.com/dariobaldi/halendar_back/internal/push"
	"github.com/dariobaldi/halendar_back/internal/secretbox"
	"github.com/google/uuid"
	"halendar/calendar"
	"halendar/mail"
)

// aiClient is the shape both internal/ollama.Client and internal/claude.Client
// implement, letting the analysis pipeline below use whichever a user has chosen
// (see aiClientFor) without caring which it's actually talking to.
type aiClient interface {
	Generate(ctx context.Context, prompt string) (string, error)
	GenerateDeterministic(ctx context.Context, prompt string) (string, error)
}

// aiClientFor returns the AI client to use for userID's analysis: their own Claude or
// Gemini key if they've connected and activated one, falling back to the shared local
// Ollama instance otherwise -- including on any lookup/decryption error, so a
// misconfigured key degrades to "use the local model" rather than breaking analysis
// outright.
func (app *app) aiClientFor(userID uuid.UUID) aiClient {
	settings, err := app.models.AISettings.Get(userID)
	if err != nil {
		app.logger.Error("ai settings: loading: " + err.Error())
		return app.ollama
	}

	var model string
	switch settings.Provider {
	case data.AIProviderClaude:
		model = app.config.claude.model
	case data.AIProviderGemini:
		model = app.config.gemini.model
	default:
		return app.ollama
	}

	encrypted, err := app.models.AISettings.GetEncryptedAPIKey(userID, settings.Provider)
	if err != nil {
		app.logger.Error("ai settings: loading " + settings.Provider + " key: " + err.Error())
		return app.ollama
	}
	plaintext, err := secretbox.Open(app.encryptionKey, encrypted)
	if err != nil {
		app.logger.Error("ai settings: decrypting " + settings.Provider + " key: " + err.Error())
		return app.ollama
	}

	if settings.Provider == data.AIProviderGemini {
		return gemini.New(string(plaintext), model)
	}
	return claude.New(string(plaintext), model)
}

// emailSyncLoop periodically checks every connected account for new mail. Each
// account is synced independently and in its own goroutine, so one slow or broken
// account never delays the others.
func (app *app) emailSyncLoop() {
	if len(app.emailProviders) == 0 {
		return // no provider configured (e.g. GOOGLE_OAUTH_CLIENT_ID unset): nothing to sync
	}
	app.background(func() {
		ticker := time.NewTicker(app.config.emailSync.interval)
		defer ticker.Stop()
		for range ticker.C {
			app.syncAllEmailAccounts()
		}
	})
}

func (app *app) syncAllEmailAccounts() {
	accounts, err := app.models.EmailAccounts.GetActive()
	if err != nil {
		app.logger.Error("email sync: listing accounts: " + err.Error())
		return
	}
	for _, account := range accounts {
		app.background(func() { app.syncEmailAccount(account) })
	}
}

// syncEmailAccount fetches whatever arrived since the last sync, stores each message
// as history, and kicks off AI analysis for it. Any failure (expired grant, IMAP
// hiccup, ...) is recorded on the account instead of propagated -- the next tick tries
// again.
func (app *app) syncEmailAccount(account data.EmailAccount) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	mailbox, err := app.connectMailbox(ctx, account)
	if err != nil {
		app.recordSyncFailure(account, err)
		return
	}

	msgs, highestUID, err := mailbox.NewSince(ctx, account.LastUID)
	if err != nil {
		app.recordSyncFailure(account, err)
		return
	}

	app.importMessages(account, msgs)

	if err := app.models.EmailAccounts.UpdateSyncState(account.ID, highestUID, nil); err != nil {
		app.logger.Error("email sync: updating sync state: " + err.Error())
	}
}

// backfillLimit caps the initial import so connecting an account with years of mail
// doesn't try to analyze all of it at once -- unread mail is what's actually likely to
// still need a reply, and 100 is already generous for that.
const backfillLimit = 100

// backfillNewAccount runs once, right after an account is connected: rather than only
// watching for mail from this point on, it imports the account's current unread
// messages (capped at backfillLimit, most recent first) so the app already has
// something useful to show before anything new even arrives. The regular sync
// baseline (the UID high-water mark) is established separately and afterwards, since
// backfilling only unread mail can't be used to infer it -- an unread message older
// than the newest read one would otherwise leave the baseline too low, causing the
// next regular sync to "rediscover" mail that was intentionally left out here.
func (app *app) backfillNewAccount(account data.EmailAccount) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	mailbox, err := app.connectMailbox(ctx, account)
	if err != nil {
		app.recordSyncFailure(account, err)
		return
	}

	msgs, err := mailbox.Search(ctx, mail.SearchQuery{Unread: true, Max: backfillLimit})
	if err != nil {
		app.logger.Error("email backfill: searching unread mail: " + err.Error())
	} else {
		app.importMessages(account, msgs)
	}

	// lastUID = 0 never imports anything on its own (see NewSince's doc comment) --
	// it's only used here to read the mailbox's current highest UID as the baseline
	// for future incremental syncs.
	_, highestUID, err := mailbox.NewSince(ctx, 0)
	if err != nil {
		app.recordSyncFailure(account, err)
		return
	}
	if err := app.models.EmailAccounts.UpdateSyncState(account.ID, highestUID, nil); err != nil {
		app.logger.Error("email backfill: updating sync state: " + err.Error())
	}
}

// importMessages stores each message as history (skipping ones already imported by an
// overlapping sync) and kicks off AI analysis for the new ones.
func (app *app) importMessages(account data.EmailAccount, msgs []mail.Message) {
	for _, msg := range msgs {
		record := data.EmailMessage{
			EmailAccountID:    account.ID,
			ProviderMessageID: msg.ID,
			IMAPUID:           msg.UID,
			FromAddress:       msg.From,
			FromName:          msg.FromName,
			Subject:           msg.Subject,
			ReceivedAt:        msg.Date,
			Snippet:           snippet(msg.Text),
			Body:              msg.Text,
		}
		inserted, err := app.models.EmailMessages.Insert(&record)
		if err != nil {
			app.logger.Error("email import: storing message: " + err.Error())
			continue
		}
		if !inserted {
			continue // already imported by a previous, overlapping sync
		}

		app.background(func() { app.analyzeEmailMessage(account.UserID, record, record.Body, true) })
	}
}

func (app *app) recordSyncFailure(account data.EmailAccount, err error) {
	app.logger.Error("email sync: "+account.EmailAddress+": "+err.Error(), "account_id", account.ID)
	if updateErr := app.models.EmailAccounts.UpdateSyncState(account.ID, account.LastUID, err); updateErr != nil {
		app.logger.Error("email sync: recording failure: " + updateErr.Error())
	}
	app.SendToWsUser(account.UserID, app.retriveWebSocket("halendar"), envelope{"type": "email_accounts", "refresh": true})
}

// connectMailbox refreshes an account's access token and builds a *mail.Mailbox ready
// to use. Shared by the regular sync pass and on-demand re-analysis.
func (app *app) connectMailbox(ctx context.Context, account data.EmailAccount) (*mail.Mailbox, error) {
	provider, ok := app.emailProviders.Get(account.Provider)
	if !ok {
		return nil, fmt.Errorf("provider %q not configured", account.Provider)
	}
	refreshToken, err := app.decryptRefreshToken(account.ID)
	if err != nil {
		return nil, err
	}
	token, err := provider.Refresh(ctx, refreshToken)
	if err != nil {
		return nil, err
	}
	return provider.Mailbox(account.EmailAddress, token), nil
}

// decryptRefreshToken loads and decrypts the OAuth2 refresh token stored for an
// account at connect time.
func (app *app) decryptRefreshToken(accountID uuid.UUID) (string, error) {
	cred, err := app.models.EmailAccounts.GetCredential(accountID)
	if err != nil {
		return "", fmt.Errorf("loading credential: %w", err)
	}
	plaintext, err := secretbox.Open(app.encryptionKey, cred.EncryptedSecret)
	if err != nil {
		return "", fmt.Errorf("decrypting credential: %w", err)
	}
	var secret oauth2Secret
	if err := json.Unmarshal(plaintext, &secret); err != nil {
		return "", fmt.Errorf("parsing credential: %w", err)
	}
	return secret.RefreshToken, nil
}

func snippet(text string) string {
	text = strings.TrimSpace(text)
	const max = 280
	if len(text) <= max {
		return text
	}
	return text[:max] + "…"
}

// eventExtractionPromptTmpl asks the local model to decide whether the SENDER is
// asking OUR USER to participate in something -- as opposed to merely mentioning an
// event the sender themselves is attending, a past event, a newsletter, or an
// automated notification, none of which call for a reply -- and if so, to extract
// every candidate date/time proposed. The reference date lets it resolve relative
// phrasing ("next Tuesday", "tomorrow afternoon") against when the mail arrived.
//
// Tuned against gemma3:1b with a small suite of representative emails (see the now-
// deleted cmd/prompttest harness -- git history has it if this needs revisiting): a
// worked weekday-math example and two few-shot examples (one accept, one reject)
// measurably improved both classification accuracy and JSON validity over a plainer
// version of these instructions. Generated with GenerateDeterministic (temperature 0)
// -- for this small a model, disabling sampling made a bigger difference than any
// further prompt wording, fixing most malformed-JSON responses outright.
const eventExtractionPromptTmpl = `You are screening one email received by our user. Today's date is %s (YYYY-MM-DD).

Weekday math example: if today is Monday 2026-09-14, then "tomorrow" = 2026-09-15, "Thursday" or "next Thursday" = 2026-09-17, "next Monday" = 2026-09-21. Always count forward from today.

TASK 1 -- decide requests_participation: true only if the SENDER is personally asking OUR USER (the reader, "you") to attend or schedule a meeting/call/event with them. Answer false for: the sender merely mentioning an event they themselves are attending, a newsletter, marketing email, or automated notification -- even if it contains dates or invites you to "join" a broadcast/webinar.

TASK 2 -- only if requests_participation is true, extract every candidate date/time the sender proposed, resolving relative dates against today's date. Each slot needs a "date" and a "start" time. Only set "end" if the sender stated an explicit end time or duration; otherwise omit "end" entirely (do not guess or repeat the start time). If the sender asks an open question with no explicit date/time ("when are you free?"), leave "slots" as an empty array.

Respond with ONLY a single JSON object, nothing before or after it -- no markdown fences, no comments, no explanation. Use real values, never the literal example text.

Example 1 (genuine request, today=2026-09-14):
Email: "Can we do a 30 min call tomorrow at 3pm about the budget?"
{"requests_participation": true, "title": "Budget call", "location": "", "slots": [{"date": "2026-09-15", "start": "15:00", "end": "15:30"}]}

Example 2 (not a request -- newsletter):
Email: "Join our free webinar next Tuesday at 2pm! Register now."
{"requests_participation": false, "title": "", "location": "", "slots": []}

Now analyze this email:

Subject: %s

Body:
"""
%s
"""`

// eventExtraction is the JSON shape eventExtractionPromptTmpl asks the model to reply
// with.
type eventExtraction struct {
	RequestsParticipation bool            `json:"requests_participation"`
	Title                 string          `json:"title"`
	Location              string          `json:"location"`
	Slots                 []extractedSlot `json:"slots"`
}

// extractedSlot is one candidate date/time as the model reports it, before it's been
// resolved to a concrete time or checked against the calendar.
type extractedSlot struct {
	Date  string `json:"date"`
	Start string `json:"start"`
	End   string `json:"end"`
}

// analyzeEmailMessage runs the local AI over a newly-imported message, records
// whether it's a genuine participation request, and -- when it is -- extracts the
// candidate time slots, checks each against the calendar, and drafts a reply:
// confirming a free slot, or declining and suggesting alternatives when none were.
// notify controls whether a push notification is sent for a newly-found proposal --
// true for a message seen for the first time, false for a re-analysis (the user
// already saw it, so re-running the prompt shouldn't notify them again).
func (app *app) analyzeEmailMessage(userID uuid.UUID, msg data.EmailMessage, body string, notify bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	client := app.aiClientFor(userID)

	referenceDate := msg.ReceivedAt
	if referenceDate.IsZero() {
		referenceDate = time.Now()
	}
	prompt := fmt.Sprintf(eventExtractionPromptTmpl, referenceDate.Format("2006-01-02"), msg.Subject, body)
	response, err := client.GenerateDeterministic(ctx, prompt)
	if err != nil {
		if setErr := app.models.EmailMessages.SetAnalysisError(msg.ID, err); setErr != nil {
			app.logger.Error("email analysis: recording error: " + setErr.Error())
		}
		return
	}

	extraction, parsed := parseEventExtraction(response)
	hasEvent := parsed && extraction.RequestsParticipation

	result, _ := json.Marshal(map[string]any{"raw_response": strings.TrimSpace(response)})
	if err := app.models.EmailMessages.SetAnalysis(msg.ID, hasEvent, result); err != nil {
		app.logger.Error("email analysis: recording result: " + err.Error())
		return
	}

	if hasEvent {
		calSource, loc := app.userCalendarSource(ctx, userID)
		slots := app.resolveEventSlots(ctx, calSource, loc, extraction.Slots)
		event := data.EmailMessageEvent{
			EmailMessageID:    msg.ID,
			Title:             firstNonEmpty(extraction.Title, msg.Subject),
			Location:          extraction.Location,
			NeedsManualReview: len(slots) == 0, // asked to participate, but no usable time could be pinned down
			Slots:             slots,
		}
		app.draftEventReply(ctx, client, userID, calSource, loc, msg, &event)
		if err := app.models.EmailEvents.Upsert(&event); err != nil {
			app.logger.Error("email analysis: storing extracted event: " + err.Error())
		} else if notify {
			app.notifyNewProposal(userID, event, msg)
		}
	} else if err := app.models.EmailEvents.DeleteForMessage(msg.ID); err != nil {
		// Only matters on a re-analysis where a previous pass had wrongly flagged an
		// event -- harmless no-op otherwise.
		app.logger.Error("email analysis: clearing stale event: " + err.Error())
	}

	app.SendToWsUser(userID, app.retriveWebSocket("halendar"), envelope{"type": "email_messages", "refresh": true})
}

// notifyNewProposal pushes a real OS-level notification (not just the in-app
// websocket refresh) to every device registered for the user, carrying the
// proposal's id so tapping it can open the app straight to that item. Uses FCM's
// "notification" payload, which Android's system displays automatically even while
// the app is backgrounded or fully closed -- no extra app code needed for that part.
func (app *app) notifyNewProposal(userID uuid.UUID, event data.EmailMessageEvent, msg data.EmailMessage) {
	devices, err := app.models.Devices.GetForUser(userID)
	if err != nil {
		app.logger.Error("push notify: listing devices: " + err.Error())
		return
	}
	if len(devices) == 0 {
		return
	}

	body := firstNonEmpty(msg.FromName, msg.FromAddress)
	if msg.Subject != "" {
		body += ": " + msg.Subject
	}
	notification := push.Notification{
		Title: "New meeting request",
		Body:  body,
		Data:  map[string]string{"type": "proposal", "proposal_id": event.ID.String()},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	for _, d := range devices {
		if err := app.push.Send(ctx, d.PushToken, notification); err != nil {
			app.logger.Error("push notify: sending to device: " + err.Error())
		}
	}
}

// reanalyzeEmailMessagesHandler re-runs AI analysis on every message already stored
// for the current user -- e.g. after improving the extraction prompt. Messages that
// already have a stored body are re-analyzed directly; older ones (imported before
// "body" existed) are re-fetched from the provider first, grouped by account so each
// account's mailbox is only connected once no matter how many of its messages need it.
func (app *app) reanalyzeEmailMessagesHandler(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)

	messages, err := app.models.EmailMessages.ListForReanalysis(user.ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	needsRefetch := make(map[uuid.UUID][]data.EmailMessage)
	for _, msg := range messages {
		if msg.Body != "" {
			msg := msg
			app.background(func() { app.analyzeEmailMessage(user.ID, msg, msg.Body, false) })
			continue
		}
		needsRefetch[msg.EmailAccountID] = append(needsRefetch[msg.EmailAccountID], msg)
	}
	for accountID, accountMessages := range needsRefetch {
		accountID, accountMessages := accountID, accountMessages
		app.background(func() { app.reanalyzeAccountMessages(user.ID, accountID, accountMessages) })
	}

	if err := app.writeJSON(w, http.StatusAccepted, envelope{"status": "reanalyzing", "count": len(messages)}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// reanalyzeAccountMessages re-fetches (from the provider, since these predate the
// stored body) and re-analyzes one account's messages, one at a time -- sequentially,
// since they share the one local Ollama model and this isn't the time-sensitive
// regular sync path.
func (app *app) reanalyzeAccountMessages(userID, accountID uuid.UUID, messages []data.EmailMessage) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	account, err := app.models.EmailAccounts.Get(accountID, userID)
	if err != nil {
		app.logger.Error("reanalyze: loading account: " + err.Error())
		return
	}
	mailbox, err := app.connectMailbox(ctx, *account)
	if err != nil {
		app.logger.Error("reanalyze: connecting mailbox: " + err.Error())
		return
	}

	for _, msg := range messages {
		full, err := mailbox.Read(ctx, msg.IMAPUID)
		if err != nil {
			app.logger.Error(fmt.Sprintf("reanalyze: reading uid %d: %s", msg.IMAPUID, err.Error()))
			continue
		}
		app.analyzeEmailMessage(userID, msg, full.Text, false)
	}
}

// parseEventExtraction pulls the JSON object out of the model's response (small local
// models sometimes wrap it in prose or markdown fences despite instructions not to)
// and decodes it. ok is false if no valid JSON object could be found.
func parseEventExtraction(response string) (extraction eventExtraction, ok bool) {
	start := strings.Index(response, "{")
	end := strings.LastIndex(response, "}")
	if start < 0 || end <= start {
		return extraction, false
	}
	if err := json.Unmarshal([]byte(response[start:end+1]), &extraction); err != nil {
		return extraction, false
	}
	return extraction, true
}

// resolveEventSlots turns the model's raw date/start/end strings into concrete times
// in loc, dropping any it can't parse, and checks each of the remaining ones against
// source (nil if the user has no calendar connected -- they're still parsed and kept,
// just with "unknown" availability).
func (app *app) resolveEventSlots(ctx context.Context, source calendarimport.Source, loc *time.Location, raw []extractedSlot) []data.EmailEventSlot {
	var slots []data.EmailEventSlot
	for _, r := range raw {
		start, err := calendar.ParseDate(r.Date+"T"+r.Start, loc)
		if err != nil {
			continue
		}
		end, err := calendar.ParseDate(r.Date+"T"+r.End, loc)
		if err != nil || !end.After(start) {
			continue
		}

		slots = append(slots, data.EmailEventSlot{
			StartAt:      start,
			EndAt:        end,
			Availability: app.checkAvailability(ctx, source, start, end),
			Source:       data.SlotSourceSender,
			Position:     len(slots),
		})
	}
	return slots
}

// checkAvailability reports a slot's availability against source, or "unknown" when
// there's no calendar connected or the check itself fails.
func (app *app) checkAvailability(ctx context.Context, source calendarimport.Source, start, end time.Time) string {
	if source == nil {
		return data.SlotAvailabilityUnknown
	}
	busy, err := source.Busy(ctx, start, end)
	if err != nil {
		return data.SlotAvailabilityUnknown
	}
	if busy {
		return data.SlotAvailabilityBusy
	}
	return data.SlotAvailabilityFree
}

// userCalendarSource returns the user's first active connected calendar and its
// timezone. With none connected (or a connection error), it returns a nil Source and
// a default timezone -- slots are still parsed and stored, just with "unknown"
// availability, rather than dropped.
func (app *app) userCalendarSource(ctx context.Context, userID uuid.UUID) (calendarimport.Source, *time.Location) {
	defaultLoc, _ := time.LoadLocation("Europe/Paris")

	accounts, err := app.models.CalendarAccounts.GetActiveForUser(userID)
	if err != nil || len(accounts) == 0 {
		return nil, defaultLoc
	}
	source, err := app.calendarSourceFor(ctx, accounts[0])
	if err != nil {
		app.logger.Error("email analysis: connecting calendar: " + err.Error())
		return nil, defaultLoc
	}
	return source, source.Timezone()
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

// alternativeSearchDays and maxAlternatives bound how far ahead -- and how many
// options -- draftEventReply looks for a free slot when none of the sender's
// proposed times are, checking the same time of day on each of the next weekdays.
// A simple, explainable heuristic rather than a full scheduling search.
const (
	alternativeSearchDays = 7
	maxAlternatives       = 3
)

// detectLanguagePromptTmpl asks the model to name the email's language, so
// replyPromptTmpl can be told explicitly rather than asked to infer "the same
// language as the original" itself. Prompt-tuning found the small model reliably
// matches the source language when told outright, but frequently defaults to English
// when left to infer it, even from otherwise-clear non-English text.
const detectLanguagePromptTmpl = `What language is the following text written in? Respond with ONLY the English name of the language (e.g. English, French, Spanish), nothing else.

Text:
"""
%s
"""`

// detectLanguage returns the model's best guess at text's language as an English
// name (e.g. "French"), or "the original email's language" if detection fails --
// still a valid, if less reliable, instruction for replyPromptTmpl to follow.
func (app *app) detectLanguage(ctx context.Context, client aiClient, text string) string {
	prompt := fmt.Sprintf(detectLanguagePromptTmpl, text)
	response, err := client.GenerateDeterministic(ctx, prompt)
	if err != nil {
		app.logger.Error("detect language: " + err.Error())
		return "the original email's language"
	}
	return strings.TrimSpace(response)
}

// replyPromptTmpl asks the local model to write the reply's body text. The actual
// decision (accept/decline, which slot, which alternatives) is made in Go from real
// calendar data beforehand -- the model only handles phrasing, which is far more
// reliable for a small local model than trusting it to reason about availability
// itself. The signature is appended separately in Go, guaranteed present regardless
// of whether the model follows the instruction not to sign.
//
// Tuned alongside eventExtractionPromptTmpl (see its comment): explicit rules and
// brevity fixed the model routinely signing off despite being told not to, and
// naming the language explicitly (via detectLanguage) fixed it defaulting to English
// for non-English originals. Generated with GenerateDeterministic for the same
// consistency reasons as the extraction prompt.
const replyPromptTmpl = `You are the Halendar Assistant, an AI scheduling assistant writing an email on %s's behalf, replying to a message about "%s". The original email is in %s -- write your reply in %s too.

%s

Rules:
- Make clear you're writing on %s's behalf, not as %s -- start with something like "%s asked me to let you know...".
- Keep it to 2-3 short sentences.
- Do NOT end with a sign-off, closing phrase, or name (no "Best,", "Regards,", "Sincerely," "Cheers," or similar, and no name on its own line) -- one is appended automatically after your text.
- Output ONLY the email body -- no subject line, no markdown, no explanation.

Original email:
"""
%s
"""`

// draftEventReply decides the outcome from event.Slots (already checked against the
// calendar) -- confirm the first free one, or decline and suggest alternatives found
// on the calendar when none were -- and asks the local model to phrase it. Sets
// event.ResponseKind/ResponseDraft, and appends any suggested alternatives to
// event.Slots.
func (app *app) draftEventReply(ctx context.Context, client aiClient, userID uuid.UUID, source calendarimport.Source, loc *time.Location, msg data.EmailMessage, event *data.EmailMessageEvent) {
	user, err := app.models.Users.Get(userID)
	if err != nil {
		app.logger.Error("draft reply: loading user: " + err.Error())
		return
	}

	var outcome string
	kind := data.ResponseKindDecline
	accepted := firstFreeSlot(event.Slots)
	switch {
	case len(event.Slots) == 0:
		// Asked to participate, but no usable date/time could be pinned down at all --
		// nothing to accept or decline, just ask the sender to suggest times.
		kind = data.ResponseKindOpenEnded
		outcome = fmt.Sprintf("%s would like to participate but no specific time was mentioned. Ask the sender to suggest some times that work for them.", user.Name)

	case accepted != nil:
		kind = data.ResponseKindAccept
		outcome = fmt.Sprintf("%s is available at %s and confirms this works.", user.Name, formatSlot(*accepted, loc))

	default:
		proposed := formatSlotList(event.Slots, loc) // all sender-proposed at this point, no alternatives added yet
		alternatives := app.findAlternativeSlots(ctx, source, loc, event.Slots)
		event.Slots = append(event.Slots, alternatives...)
		if len(alternatives) > 0 {
			outcome = fmt.Sprintf(
				"%s is not available at the proposed time(s) (%s). Politely decline those, then suggest these alternative times instead and ask which works best: %s.",
				user.Name, proposed, formatSlotList(alternatives, loc),
			)
		} else {
			outcome = fmt.Sprintf(
				"%s is not available at the proposed time(s) (%s). Politely decline and ask the sender to suggest other times.",
				user.Name, proposed,
			)
		}
	}

	language := app.detectLanguage(ctx, client, msg.Body)
	prompt := fmt.Sprintf(replyPromptTmpl, user.Name, msg.Subject, language, language, outcome, user.Name, user.Name, user.Name, msg.Body)
	response, err := client.GenerateDeterministic(ctx, prompt)
	if err != nil {
		app.logger.Error("draft reply: generating: " + err.Error())
		return
	}

	draft := strings.TrimSpace(response) + "\n\n—\nSent by the Halendar Assistant on behalf of " + user.Name + "."
	event.ResponseKind = &kind
	event.ResponseDraft = &draft
}

// firstFreeSlot returns the first (in the sender's own proposed order) free slot, if
// any.
func firstFreeSlot(slots []data.EmailEventSlot) *data.EmailEventSlot {
	for i := range slots {
		if slots[i].Source == data.SlotSourceSender && slots[i].Availability == data.SlotAvailabilityFree {
			return &slots[i]
		}
	}
	return nil
}

// findAlternativeSlots looks for a free slot of the same duration as the sender's
// first proposed time, at that same time of day on each of the next few weekdays.
func (app *app) findAlternativeSlots(ctx context.Context, source calendarimport.Source, loc *time.Location, proposed []data.EmailEventSlot) []data.EmailEventSlot {
	if source == nil || len(proposed) == 0 {
		return nil
	}
	first := proposed[0]
	duration := first.EndAt.Sub(first.StartAt)
	if duration <= 0 {
		duration = time.Hour
	}

	var alternatives []data.EmailEventSlot
	for dayOffset := 1; dayOffset <= alternativeSearchDays && len(alternatives) < maxAlternatives; dayOffset++ {
		start := first.StartAt.AddDate(0, 0, dayOffset)
		if start.Weekday() == time.Saturday || start.Weekday() == time.Sunday {
			continue
		}
		end := start.Add(duration)
		if app.checkAvailability(ctx, source, start, end) != data.SlotAvailabilityFree {
			continue
		}
		alternatives = append(alternatives, data.EmailEventSlot{
			StartAt:      start,
			EndAt:        end,
			Availability: data.SlotAvailabilityFree,
			Source:       data.SlotSourceSuggested,
			Position:     len(alternatives),
		})
	}
	return alternatives
}

// formatSlot renders one slot in loc as e.g. "Tuesday, Sep 22 14:00-14:30".
func formatSlot(s data.EmailEventSlot, loc *time.Location) string {
	start, end := s.StartAt.In(loc), s.EndAt.In(loc)
	return fmt.Sprintf("%s %s-%s", start.Format("Monday, Jan 2"), start.Format("15:04"), end.Format("15:04"))
}

// formatSlotList renders each slot as a comma-separated list.
func formatSlotList(slots []data.EmailEventSlot, loc *time.Location) string {
	parts := make([]string, len(slots))
	for i, s := range slots {
		parts[i] = formatSlot(s, loc)
	}
	return strings.Join(parts, ", ")
}
