package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

// Availability values for an EmailEventSlot, as checked against the calendar at
// extraction time. Unknown means the calendar check itself failed (e.g. CalDAV was
// unreachable) -- distinct from Free/Busy, and not the same as needing manual review.
const (
	SlotAvailabilityFree    = "free"
	SlotAvailabilityBusy    = "busy"
	SlotAvailabilityUnknown = "unknown"
)

// Source values for an EmailEventSlot.
const (
	SlotSourceSender    = "sender"    // one of the times the sender proposed
	SlotSourceSuggested = "suggested" // an alternative free time we found and offered instead
)

// Response kinds for an EmailMessageEvent, once its slots have been checked.
const (
	ResponseKindAccept    = "accept"
	ResponseKindDecline   = "decline"
	ResponseKindOpenEnded = "open_ended" // asked to participate, but named no specific time to accept or decline
)

// EmailEventSlot is one candidate date/time, either proposed by the sender or
// suggested by us as an alternative.
type EmailEventSlot struct {
	ID           uuid.UUID `json:"id"`
	StartAt      time.Time `json:"start_at"`
	EndAt        time.Time `json:"end_at"`
	Availability string    `json:"availability"`
	Source       string    `json:"source"`
	Position     int       `json:"position"`
}

// EmailMessageEvent is the structured event extracted from a message, created only
// when the sender is asking our user to participate in something, plus the drafted
// reply once its slots have been checked against the calendar.
type EmailMessageEvent struct {
	ID                uuid.UUID        `json:"id"`
	EmailMessageID    uuid.UUID        `json:"email_message_id"`
	Title             string           `json:"title"`
	Location          string           `json:"location,omitempty"`
	NeedsManualReview bool             `json:"needs_manual_review"`
	ResponseKind      *string          `json:"response_kind,omitempty"`
	ResponseDraft     *string          `json:"response_draft,omitempty"`
	CreatedAt         time.Time        `json:"created_at"`
	Slots             []EmailEventSlot `json:"slots"`
}

type EmailEventModel struct {
	DB *sql.DB
}

// Upsert stores the extracted event and its candidate slots as one unit, replacing
// whatever was there before for this message -- analysis can be re-run (e.g. after
// improving the prompt) without first checking whether a row already exists.
func (m EmailEventModel) Upsert(event *EmailMessageEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// response_kind/response_draft are only overwritten by a fresh analysis when the
	// user hasn't hand-edited the draft yet -- otherwise re-analysis (e.g. after
	// improving the prompt) would silently clobber their own wording.
	query := `
		INSERT INTO email_message_events (email_message_id, title, location, needs_manual_review, response_kind, response_draft)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (email_message_id) DO UPDATE SET
			title = EXCLUDED.title,
			location = EXCLUDED.location,
			needs_manual_review = EXCLUDED.needs_manual_review,
			response_kind = CASE WHEN email_message_events.draft_edited_by_user THEN email_message_events.response_kind ELSE EXCLUDED.response_kind END,
			response_draft = CASE WHEN email_message_events.draft_edited_by_user THEN email_message_events.response_draft ELSE EXCLUDED.response_draft END
		RETURNING id, created_at
	`
	err = tx.QueryRowContext(ctx, query,
		event.EmailMessageID, event.Title, event.Location, event.NeedsManualReview, event.ResponseKind, event.ResponseDraft,
	).Scan(&event.ID, &event.CreatedAt)
	if err != nil {
		return err
	}

	// Simplest way to keep slots consistent with a fresh extraction: drop whatever was
	// there and re-insert, rather than trying to diff against the previous set.
	if _, err := tx.ExecContext(ctx, `DELETE FROM email_event_slots WHERE email_message_event_id = $1`, event.ID); err != nil {
		return err
	}

	slotQuery := `
		INSERT INTO email_event_slots (email_message_event_id, start_at, end_at, availability, source, position)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`
	for i := range event.Slots {
		slot := &event.Slots[i]
		if err := tx.QueryRowContext(ctx, slotQuery, event.ID, slot.StartAt, slot.EndAt, slot.Availability, slot.Source, slot.Position).Scan(&slot.ID); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// DeleteForMessage removes the extracted event (and its slots) for a message, if any.
// Used when re-analysis decides a message no longer qualifies as a participation
// request, so a stale event doesn't linger.
func (m EmailEventModel) DeleteForMessage(messageID uuid.UUID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, `DELETE FROM email_message_events WHERE email_message_id = $1`, messageID)
	return err
}

// GetForMessage returns the extracted event and its slots (sender-proposed slots
// first, then any suggested alternatives, each ordered by position) for a message,
// if any.
func (m EmailEventModel) GetForMessage(messageID uuid.UUID) (*EmailMessageEvent, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var event EmailMessageEvent
	query := `
		SELECT id, email_message_id, title, location, needs_manual_review, response_kind, response_draft, created_at
		FROM email_message_events
		WHERE email_message_id = $1
	`
	err := m.DB.QueryRowContext(ctx, query, messageID).Scan(
		&event.ID, &event.EmailMessageID, &event.Title, &event.Location, &event.NeedsManualReview,
		&event.ResponseKind, &event.ResponseDraft, &event.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	rows, err := m.DB.QueryContext(ctx, `
		SELECT id, start_at, end_at, availability, source, position
		FROM email_event_slots
		WHERE email_message_event_id = $1
		ORDER BY CASE source WHEN 'sender' THEN 0 ELSE 1 END, position
	`, event.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var slot EmailEventSlot
		if err := rows.Scan(&slot.ID, &slot.StartAt, &slot.EndAt, &slot.Availability, &slot.Source, &slot.Position); err != nil {
			return nil, err
		}
		event.Slots = append(event.Slots, slot)
	}
	return &event, rows.Err()
}
