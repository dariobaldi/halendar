package data

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// Status values for a Proposal -- the frontend's "needs action" (pending) vs
// "history" (confirmed/rejected) split.
const (
	ProposalStatusPending   = "pending"
	ProposalStatusConfirmed = "confirmed"
	ProposalStatusRejected  = "rejected"
)

// Proposal is the read shape the frontend's Messages page needs: an analyzed message
// joined with its source email, ready to display without further lookups. Every
// imported, analyzed message gets one of these -- not just ones that turned out to be
// meeting requests -- so IsMeetingRequest tells the frontend whether to show the
// extracted slots/draft or just the email itself. It's a view over the same
// email_message_events/email_event_slots tables EmailEventModel writes -- kept
// separate because the two have different shapes for different purposes (write-time
// extraction result vs. read-time display+review).
type Proposal struct {
	ID               uuid.UUID `json:"id"`
	SenderName       string    `json:"sender_name"`
	SenderEmail      string    `json:"sender_email"`
	Subject          string    `json:"subject"`
	ReceivedAt       time.Time `json:"received_at"`
	EmailExcerpt     string    `json:"email_excerpt"`
	IsMeetingRequest bool      `json:"is_meeting_request"`
	// SuggestedSkip and SkipReason are set for a non-meeting message that's very
	// likely not worth reading at all -- a promotional/automated category from the
	// AI, or a sender address that plainly can't receive a reply -- so the Messages
	// page can offer a fast, no-confirmation skip instead of the normal one. Never
	// set for a genuine meeting request, however it was sent.
	SuggestedSkip     bool             `json:"suggested_skip,omitempty"`
	SkipReason        string           `json:"skip_reason,omitempty"`
	Slots             []EmailEventSlot `json:"slots"`
	SelectedSlotID    *uuid.UUID       `json:"selected_slot_id,omitempty"`
	NeedsManualReview bool             `json:"needs_manual_review"`
	ResponseDraft     string           `json:"response_draft"`
	DraftEditedByUser bool             `json:"draft_edited_by_user"`
	Status            string           `json:"status"`

	// Backend-only, needed to actually send the reply, book the event, and re-run
	// analysis -- not something the frontend has any use for.
	EmailMessageID uuid.UUID `json:"-"`
	EmailAccountID uuid.UUID `json:"-"`
	IMAPUID        uint32    `json:"-"`
	Title          string    `json:"-"`
	Location       string    `json:"-"`
	category       string    // the AI's raw category, only used to derive SuggestedSkip
}

// noReplyPatterns match the local part of a sender address that plainly can't (or
// isn't meant to) receive a reply -- a strong, free, deterministic signal that costs
// no AI call, independent of whatever category the AI assigned the message.
var noReplyPatterns = []string{"no-reply", "noreply", "do-not-reply", "donotreply", "mailer-daemon", "postmaster"}

func isNoReplyAddress(address string) bool {
	local, _, found := strings.Cut(address, "@")
	if !found {
		local = address
	}
	local = strings.ToLower(local)
	for _, pattern := range noReplyPatterns {
		if strings.Contains(local, pattern) {
			return true
		}
	}
	return false
}

// deriveSuggestedSkip fills in SuggestedSkip/SkipReason for a proposal that isn't a
// meeting request, from whichever signal applies (checked in order of how confident
// it is). A genuine meeting request is never suggested for skipping, regardless of
// the sender address -- e.g. a shared team inbox that happens to look automated.
func (p *Proposal) deriveSuggestedSkip() {
	if p.IsMeetingRequest {
		return
	}
	switch {
	case isNoReplyAddress(p.SenderEmail):
		p.SuggestedSkip = true
		p.SkipReason = "Sent from an address that can't receive replies."
	case p.category == CategoryPromotional:
		p.SuggestedSkip = true
		p.SkipReason = "Looks like a promotional email."
	case p.category == CategoryAutomated:
		p.SuggestedSkip = true
		p.SkipReason = "Looks like an automated notification."
	}
}

// EffectiveSlot returns the slot to treat as "the one the user is going with": the
// one they explicitly picked (SelectedSlotID), or, absent that, the first free one --
// the same default the frontend shows before any explicit selection is made. Nil
// means there's nothing to book (decline, or no slots at all).
func (p Proposal) EffectiveSlot() *EmailEventSlot {
	if p.SelectedSlotID != nil {
		for i := range p.Slots {
			if p.Slots[i].ID == *p.SelectedSlotID {
				return &p.Slots[i]
			}
		}
	}
	for i := range p.Slots {
		if p.Slots[i].Availability == SlotAvailabilityFree {
			return &p.Slots[i]
		}
	}
	return nil
}

type ProposalModel struct {
	DB *sql.DB
}

// GetForUser returns every proposal across all of a user's connected accounts, most
// recently received first. The frontend splits pending from confirmed/rejected
// itself, the same way it already does with the mock data.
func (m ProposalModel) GetForUser(userID uuid.UUID) ([]Proposal, error) {
	return m.list("ea.user_id = $1", userID)
}

// GetByID returns one proposal, scoped to its owner so one user can't reach another's.
func (m ProposalModel) GetByID(id, userID uuid.UUID) (*Proposal, error) {
	proposals, err := m.list("ev.id = $1 AND ea.user_id = $2", id, userID)
	if err != nil {
		return nil, err
	}
	if len(proposals) == 0 {
		return nil, ErrRecordNotFound
	}
	return &proposals[0], nil
}

func (m ProposalModel) list(where string, args ...any) ([]Proposal, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := `
		SELECT ev.id, em.from_name, em.from_address, em.subject, em.received_at, em.body,
			em.has_event, ev.needs_manual_review, ev.category, COALESCE(ev.response_draft, ''), ev.draft_edited_by_user,
			ev.status, ev.selected_slot_id, em.id, em.email_account_id, em.imap_uid, ev.title, ev.location
		FROM email_message_events ev
		INNER JOIN email_messages em ON em.id = ev.email_message_id
		INNER JOIN email_accounts ea ON ea.id = em.email_account_id
		WHERE ` + where + `
		ORDER BY em.received_at DESC NULLS LAST
	`
	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	proposals := []Proposal{}
	byID := make(map[uuid.UUID]*Proposal)
	var ids []uuid.UUID
	for rows.Next() {
		p := Proposal{Slots: []EmailEventSlot{}} // marshals as [] rather than null when there are none
		var isMeetingRequest sql.NullBool
		if err := rows.Scan(
			&p.ID, &p.SenderName, &p.SenderEmail, &p.Subject, &p.ReceivedAt, &p.EmailExcerpt,
			&isMeetingRequest, &p.NeedsManualReview, &p.category, &p.ResponseDraft, &p.DraftEditedByUser, &p.Status, &p.SelectedSlotID,
			&p.EmailMessageID, &p.EmailAccountID, &p.IMAPUID, &p.Title, &p.Location,
		); err != nil {
			return nil, err
		}
		p.IsMeetingRequest = isMeetingRequest.Valid && isMeetingRequest.Bool
		p.deriveSuggestedSkip()
		proposals = append(proposals, p)
		ids = append(ids, p.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for i := range proposals {
		byID[proposals[i].ID] = &proposals[i]
	}

	if len(ids) == 0 {
		return proposals, nil
	}

	slotRows, err := m.DB.QueryContext(ctx, `
		SELECT email_message_event_id, id, start_at, end_at, availability, source, position
		FROM email_event_slots
		WHERE email_message_event_id = ANY($1)
		ORDER BY CASE source WHEN 'sender' THEN 0 ELSE 1 END, position
	`, pq.Array(uuidArray(ids)))
	if err != nil {
		return nil, err
	}
	defer slotRows.Close()

	for slotRows.Next() {
		var eventID uuid.UUID
		var slot EmailEventSlot
		if err := slotRows.Scan(&eventID, &slot.ID, &slot.StartAt, &slot.EndAt, &slot.Availability, &slot.Source, &slot.Position); err != nil {
			return nil, err
		}
		if p, ok := byID[eventID]; ok {
			p.Slots = append(p.Slots, slot)
		}
	}
	return proposals, slotRows.Err()
}

// uuidArray adapts a []uuid.UUID for use with Postgres' ANY($1) against a uuid[]
// column, since lib/pq doesn't marshal uuid.UUID slices on its own.
func uuidArray(ids []uuid.UUID) []string {
	out := make([]string, len(ids))
	for i, id := range ids {
		out[i] = id.String()
	}
	return out
}

// ownedProposalQuery is the WHERE clause shared by every mutation below, scoping the
// target event to one the requesting user actually owns (through
// message -> account -> user), so one user can't act on another's proposal.
const ownedProposalFrom = `
	FROM email_messages em
	INNER JOIN email_accounts ea ON ea.id = em.email_account_id
	WHERE email_message_events.email_message_id = em.id
		AND ea.user_id = $2
		AND email_message_events.id = $1
`

// SetSelectedSlot records which slot the user wants to go with -- slot must belong to
// this proposal, or the update is rejected. Also updates the draft text in the same
// call, since the frontend regenerates it locally whenever the user hasn't hand-edited
// it yet (see draftEditedByUser) and sends both together.
func (m ProposalModel) SetSelectedSlot(id, userID, slotID uuid.UUID, responseDraft string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, `
		UPDATE email_message_events
		SET selected_slot_id = $3, response_draft = $4, draft_edited_by_user = FALSE
		`+ownedProposalFrom+`
			AND EXISTS (SELECT 1 FROM email_event_slots s WHERE s.id = $3 AND s.email_message_event_id = email_message_events.id)
	`, id, userID, slotID, responseDraft)
	return checkOwnedUpdate(result, err)
}

// SetDraft records the user's own edited reply text.
func (m ProposalModel) SetDraft(id, userID uuid.UUID, responseDraft string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, `
		UPDATE email_message_events
		SET response_draft = $3, draft_edited_by_user = TRUE
		`+ownedProposalFrom, id, userID, responseDraft)
	return checkOwnedUpdate(result, err)
}

// SetStatus moves a proposal to any status unconditionally -- used to revert a
// ClaimPending back to pending when sending the reply afterward fails, so the user
// can retry. Confirming/rejecting for real goes through ClaimPending instead, which
// guards against acting twice.
func (m ProposalModel) SetStatus(id, userID uuid.UUID, status string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, `
		UPDATE email_message_events
		SET status = $3
		`+ownedProposalFrom, id, userID, status)
	return checkOwnedUpdate(result, err)
}

// ClaimPending moves a proposal from pending to newStatus, but only if it's still
// pending -- the atomic "claim" step before actually sending a reply or booking an
// event, so two concurrent requests for the same proposal can't both go through.
// Whichever call flips the row first proceeds; the other gets ErrRecordNotFound.
func (m ProposalModel) ClaimPending(id, userID uuid.UUID, newStatus string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, `
		UPDATE email_message_events
		SET status = $3
		`+ownedProposalFrom+`
			AND email_message_events.status = '`+ProposalStatusPending+`'
	`, id, userID, newStatus)
	return checkOwnedUpdate(result, err)
}

func checkOwnedUpdate(result sql.Result, err error) error {
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrRecordNotFound
	}
	return nil
}
