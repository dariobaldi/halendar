package data

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

// Status and auth-type values, mirroring EmailAccount's.
const (
	CalendarAccountStatusActive  = "active"
	CalendarAccountStatusError   = "error"
	CalendarAccountStatusRevoked = "revoked"

	CalendarAuthTypeOAuth2 = "oauth2"
	CalendarAuthTypeBasic  = "basic" // CalDAV: a URL + username + password
)

// CalendarAccount is one calendar a user has connected (Google Calendar, a CalDAV
// server such as iCloud or a groupware's calendar, ...). Credentials are never on
// this struct -- see CalendarAccountCredential. Config holds whatever non-secret,
// provider-specific settings that provider needs (e.g. a CalDAV URL).
type CalendarAccount struct {
	ID          uuid.UUID       `json:"id"`
	UserID      uuid.UUID       `json:"-"`
	Provider    string          `json:"provider"`
	DisplayName string          `json:"display_name"`
	Config      json.RawMessage `json:"-"`
	Status      string          `json:"status"`
	LastError   *string         `json:"last_error,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// CalendarAccountCredential holds the still-encrypted secret for one account.
type CalendarAccountCredential struct {
	CalendarAccountID uuid.UUID
	AuthType          string
	EncryptedSecret   []byte
}

type CalendarAccountModel struct {
	DB *sql.DB
}

// Upsert creates the account if it doesn't already exist for (user, provider, display
// name), or reactivates and refreshes its config if it does.
func (m CalendarAccountModel) Upsert(account *CalendarAccount) error {
	if account.Config == nil {
		account.Config = json.RawMessage("{}")
	}
	query := `
		INSERT INTO calendar_accounts (user_id, provider, display_name, config)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, provider, display_name) DO UPDATE SET
			config = EXCLUDED.config,
			status = $5,
			last_error = NULL,
			updated_at = NOW()
		RETURNING id, status, created_at, updated_at
	`
	args := []interface{}{account.UserID, account.Provider, account.DisplayName, account.Config, CalendarAccountStatusActive}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&account.ID, &account.Status, &account.CreatedAt, &account.UpdatedAt)
}

// GetForUser lists every calendar a user has connected, most recently created first.
func (m CalendarAccountModel) GetForUser(userID uuid.UUID) ([]CalendarAccount, error) {
	query := `
		SELECT id, user_id, provider, display_name, config, status, last_error, created_at, updated_at
		FROM calendar_accounts
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	accounts := []CalendarAccount{}
	for rows.Next() {
		var a CalendarAccount
		if err := rows.Scan(&a.ID, &a.UserID, &a.Provider, &a.DisplayName, &a.Config, &a.Status, &a.LastError, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		accounts = append(accounts, a)
	}
	return accounts, rows.Err()
}

// GetActiveForUser returns the user's active calendars -- what the email-analysis
// pipeline checks slots against.
func (m CalendarAccountModel) GetActiveForUser(userID uuid.UUID) ([]CalendarAccount, error) {
	accounts, err := m.GetForUser(userID)
	if err != nil {
		return nil, err
	}
	active := accounts[:0]
	for _, a := range accounts {
		if a.Status == CalendarAccountStatusActive {
			active = append(active, a)
		}
	}
	return active, nil
}

// Get returns one account, scoped to its owner so one user can't reach another's.
func (m CalendarAccountModel) Get(id, userID uuid.UUID) (*CalendarAccount, error) {
	query := `
		SELECT id, user_id, provider, display_name, config, status, last_error, created_at, updated_at
		FROM calendar_accounts
		WHERE id = $1 AND user_id = $2
	`
	var a CalendarAccount

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id, userID).Scan(
		&a.ID, &a.UserID, &a.Provider, &a.DisplayName, &a.Config, &a.Status, &a.LastError, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}
	return &a, nil
}

// UpdateStatus records the outcome of trying to use an account (e.g. a busy-check),
// clearing any previous error on success.
func (m CalendarAccountModel) UpdateStatus(id uuid.UUID, useErr error) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if useErr == nil {
		_, err := m.DB.ExecContext(ctx, `
			UPDATE calendar_accounts SET status = $2, last_error = NULL, updated_at = NOW() WHERE id = $1
		`, id, CalendarAccountStatusActive)
		return err
	}
	_, err := m.DB.ExecContext(ctx, `
		UPDATE calendar_accounts SET status = $2, last_error = $3, updated_at = NOW() WHERE id = $1
	`, id, CalendarAccountStatusError, useErr.Error())
	return err
}

// Delete removes an account (and, via cascade, its credentials), scoped to its owner.
func (m CalendarAccountModel) Delete(id, userID uuid.UUID) error {
	query := `DELETE FROM calendar_accounts WHERE id = $1 AND user_id = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, id, userID)
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

// PutCredential stores (inserting or replacing) the encrypted secret for an account.
func (m CalendarAccountModel) PutCredential(cred CalendarAccountCredential) error {
	query := `
		INSERT INTO calendar_account_credentials (calendar_account_id, auth_type, encrypted_secret)
		VALUES ($1, $2, $3)
		ON CONFLICT (calendar_account_id) DO UPDATE SET
			auth_type = EXCLUDED.auth_type,
			encrypted_secret = EXCLUDED.encrypted_secret,
			updated_at = NOW()
	`
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, cred.CalendarAccountID, cred.AuthType, cred.EncryptedSecret)
	return err
}

// GetCredential returns the still-encrypted secret for an account.
func (m CalendarAccountModel) GetCredential(accountID uuid.UUID) (*CalendarAccountCredential, error) {
	query := `
		SELECT calendar_account_id, auth_type, encrypted_secret
		FROM calendar_account_credentials
		WHERE calendar_account_id = $1
	`
	var cred CalendarAccountCredential

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, accountID).Scan(&cred.CalendarAccountID, &cred.AuthType, &cred.EncryptedSecret)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}
	return &cred, nil
}
