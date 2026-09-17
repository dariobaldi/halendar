package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dariobaldi/halendar_back/internal/calendarimport"
	"github.com/dariobaldi/halendar_back/internal/data"
	"github.com/dariobaldi/halendar_back/internal/secretbox"
	"github.com/google/uuid"
	"halendar/calendar"
	"halendar/mail"
)

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
			app.logger.Error("email sync: storing message: " + err.Error())
			continue
		}
		if !inserted {
			continue // already imported by a previous, overlapping sync
		}

		app.background(func() { app.analyzeEmailMessage(account.UserID, record, record.Body) })
	}

	if err := app.models.EmailAccounts.UpdateSyncState(account.ID, highestUID, nil); err != nil {
		app.logger.Error("email sync: updating sync state: " + err.Error())
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
const eventExtractionPromptTmpl = `You are screening one email received by our user on %s (YYYY-MM-DD, this is "today" for resolving relative dates).

Decide whether the SENDER is inviting or asking OUR USER (the reader, "you") to participate in a meeting, call, or event. This is NOT the case when the sender is only mentioning an event they themselves are attending, describing something that already happened, or sending a newsletter/automated notice -- only a genuine ask for the reader to attend or schedule something counts.

If, and only if, that's the case, extract every candidate date/time the sender proposed or asked about, resolving relative dates using the reference date above. If the sender asks an open question with no explicit date/time ("when are you free?"), leave "slots" empty.

Respond with ONLY this JSON object and nothing else -- no explanation, no markdown fences:
{"requests_participation": true, "title": "short event title, or empty string", "location": "place or link mentioned, or empty string", "slots": [{"date": "YYYY-MM-DD", "start": "HH:MM", "end": "HH:MM"}]}

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
func (app *app) analyzeEmailMessage(userID uuid.UUID, msg data.EmailMessage, body string) {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	referenceDate := msg.ReceivedAt
	if referenceDate.IsZero() {
		referenceDate = time.Now()
	}
	prompt := fmt.Sprintf(eventExtractionPromptTmpl, referenceDate.Format("2006-01-02"), msg.Subject, body)
	response, err := app.ollama.Generate(ctx, prompt)
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
		if len(slots) > 0 {
			app.draftEventReply(ctx, userID, calSource, loc, msg, &event)
		}
		if err := app.models.EmailEvents.Upsert(&event); err != nil {
			app.logger.Error("email analysis: storing extracted event: " + err.Error())
		}
	} else if err := app.models.EmailEvents.DeleteForMessage(msg.ID); err != nil {
		// Only matters on a re-analysis where a previous pass had wrongly flagged an
		// event -- harmless no-op otherwise.
		app.logger.Error("email analysis: clearing stale event: " + err.Error())
	}

	app.SendToWsUser(userID, app.retriveWebSocket("halendar"), envelope{"type": "email_messages", "refresh": true})
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
			app.background(func() { app.analyzeEmailMessage(user.ID, msg, msg.Body) })
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
		app.analyzeEmailMessage(userID, msg, full.Text)
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

// replyPromptTmpl asks the local model to write the reply's body text. The actual
// decision (accept/decline, which slot, which alternatives) is made in Go from real
// calendar data beforehand -- the model only handles phrasing, which is far more
// reliable for a small local model than trusting it to reason about availability
// itself. The signature is appended separately in Go, guaranteed present regardless
// of whether the model follows the instruction not to sign.
const replyPromptTmpl = `You are the Halendar Assistant, an AI scheduling assistant writing an email on %s's behalf, replying to a message about "%s".

Write the reply in the same language as the original email below. Make it clear in the body that you (the Halendar Assistant) are writing on %s's behalf, not %s personally -- for example "%s asked me to let you know...". Keep it short and polite. Do not sign or sign off the email yourself -- a signature is added automatically afterwards.

%s

Respond with ONLY the email body text -- no subject line, no signature, no explanation, no markdown.

Original email:
"""
%s
"""`

// draftEventReply decides the outcome from event.Slots (already checked against the
// calendar) -- confirm the first free one, or decline and suggest alternatives found
// on the calendar when none were -- and asks the local model to phrase it. Sets
// event.ResponseKind/ResponseDraft, and appends any suggested alternatives to
// event.Slots.
func (app *app) draftEventReply(ctx context.Context, userID uuid.UUID, source calendarimport.Source, loc *time.Location, msg data.EmailMessage, event *data.EmailMessageEvent) {
	user, err := app.models.Users.Get(userID)
	if err != nil {
		app.logger.Error("draft reply: loading user: " + err.Error())
		return
	}

	var outcome string
	kind := data.ResponseKindDecline
	if accepted := firstFreeSlot(event.Slots); accepted != nil {
		kind = data.ResponseKindAccept
		outcome = fmt.Sprintf("%s is available at %s and confirms this works.", user.Name, formatSlot(*accepted, loc))
	} else {
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

	prompt := fmt.Sprintf(replyPromptTmpl, user.Name, msg.Subject, user.Name, user.Name, user.Name, outcome, msg.Body)
	response, err := app.ollama.Generate(ctx, prompt)
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
