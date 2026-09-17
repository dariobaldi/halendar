package data

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

// Status values for an EmailMessage's analysis.
const (
	AnalysisStatusPending  = "pending"
	AnalysisStatusAnalyzed = "analyzed"
	AnalysisStatusError    = "error"
)

// EmailMessage is one imported message and the outcome of analyzing it -- the history
// the user can look back on, independent of whatever a future "proposal" screen does
// with it.
type EmailMessage struct {
	ID                uuid.UUID       `json:"id"`
	EmailAccountID    uuid.UUID       `json:"email_account_id"`
	ProviderMessageID string          `json:"-"`
	IMAPUID           uint32          `json:"-"`
	FromAddress       string          `json:"from_address"`
	FromName          string          `json:"from_name,omitempty"`
	Subject           string          `json:"subject"`
	ReceivedAt        time.Time       `json:"received_at"`
	Snippet           string          `json:"snippet,omitempty"`
	Body              string          `json:"body,omitempty"` // full plain-text body, for the frontend to show when deciding on a proposal
	AnalysisStatus    string          `json:"analysis_status"`
	HasEvent          *bool           `json:"has_event,omitempty"`
	AnalysisResult    json.RawMessage `json:"analysis_result,omitempty"`
	AnalyzedAt        *time.Time      `json:"analyzed_at,omitempty"`
	AnalysisError     *string         `json:"analysis_error,omitempty"`
	CreatedAt         time.Time       `json:"created_at"`
}

type EmailMessageModel struct {
	DB *sql.DB
}

// Insert records a newly-imported message. If the account already has a message with
// the same ProviderMessageID (a re-sync overlap), it is left untouched and ok is false.
func (m EmailMessageModel) Insert(msg *EmailMessage) (ok bool, err error) {
	query := `
		INSERT INTO email_messages (email_account_id, provider_message_id, imap_uid, from_address, from_name, subject, received_at, snippet, body)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (email_account_id, provider_message_id) DO NOTHING
		RETURNING id, analysis_status, created_at
	`
	args := []interface{}{msg.EmailAccountID, msg.ProviderMessageID, msg.IMAPUID, msg.FromAddress, msg.FromName, msg.Subject, msg.ReceivedAt, msg.Snippet, msg.Body}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = m.DB.QueryRowContext(ctx, query, args...).Scan(&msg.ID, &msg.AnalysisStatus, &msg.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// SetAnalysis records the outcome of running the AI analysis on a message.
func (m EmailMessageModel) SetAnalysis(id uuid.UUID, hasEvent bool, result json.RawMessage) error {
	query := `
		UPDATE email_messages
		SET analysis_status = $2, has_event = $3, analysis_result = $4, analyzed_at = NOW(), analysis_error = NULL
		WHERE id = $1
	`
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, id, AnalysisStatusAnalyzed, hasEvent, result)
	return err
}

// SetAnalysisError records that analysis failed for a message (e.g. Ollama unreachable).
func (m EmailMessageModel) SetAnalysisError(id uuid.UUID, analysisErr error) error {
	query := `
		UPDATE email_messages
		SET analysis_status = $2, analysis_error = $3
		WHERE id = $1
	`
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, id, AnalysisStatusError, analysisErr.Error())
	return err
}

// GetForUser returns the message history across every account belonging to a user,
// most recent first.
func (m EmailMessageModel) GetForUser(userID uuid.UUID, limit int) ([]EmailMessage, error) {
	if limit <= 0 {
		limit = 50
	}
	query := `
		SELECT em.id, em.email_account_id, em.from_address, em.from_name, em.subject,
			em.received_at, em.snippet, em.body, em.analysis_status, em.has_event,
			COALESCE(em.analysis_result, 'null'), em.analyzed_at, em.analysis_error, em.created_at
		FROM email_messages em
		INNER JOIN email_accounts ea ON ea.id = em.email_account_id
		WHERE ea.user_id = $1
		ORDER BY em.received_at DESC NULLS LAST
		LIMIT $2
	`
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages := []EmailMessage{}
	for rows.Next() {
		var msg EmailMessage
		if err := rows.Scan(
			&msg.ID, &msg.EmailAccountID, &msg.FromAddress, &msg.FromName, &msg.Subject,
			&msg.ReceivedAt, &msg.Snippet, &msg.Body, &msg.AnalysisStatus, &msg.HasEvent, &msg.AnalysisResult,
			&msg.AnalyzedAt, &msg.AnalysisError, &msg.CreatedAt,
		); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}
	return messages, rows.Err()
}

// ListForReanalysis returns every message belonging to a user, with its stored body
// when there is one. Messages imported before "body" existed have none -- the caller
// falls back to re-fetching those from the provider using EmailAccountID/IMAPUID.
func (m EmailMessageModel) ListForReanalysis(userID uuid.UUID) ([]EmailMessage, error) {
	query := `
		SELECT em.id, em.email_account_id, em.imap_uid, em.subject, em.received_at, em.body
		FROM email_messages em
		INNER JOIN email_accounts ea ON ea.id = em.email_account_id
		WHERE ea.user_id = $1
		ORDER BY em.email_account_id, em.imap_uid
	`
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages := []EmailMessage{}
	for rows.Next() {
		var msg EmailMessage
		if err := rows.Scan(&msg.ID, &msg.EmailAccountID, &msg.IMAPUID, &msg.Subject, &msg.ReceivedAt, &msg.Body); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}
	return messages, rows.Err()
}
