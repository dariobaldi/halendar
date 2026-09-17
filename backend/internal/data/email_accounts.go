package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

// Status values for an EmailAccount.
const (
	EmailAccountStatusActive  = "active"
	EmailAccountStatusError   = "error"
	EmailAccountStatusRevoked = "revoked"
)

// EmailAccount is one messaging account a user has connected (e.g. a Gmail inbox).
// Credentials are never on this struct -- see EmailAccountCredential.
type EmailAccount struct {
	ID           uuid.UUID  `json:"id"`
	UserID       uuid.UUID  `json:"-"`
	Provider     string     `json:"provider"`
	EmailAddress string     `json:"email_address"`
	Status       string     `json:"status"`
	LastError    *string    `json:"last_error,omitempty"`
	LastSyncedAt *time.Time `json:"last_synced_at,omitempty"`
	LastUID      uint32     `json:"-"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// EmailAccountCredential holds the still-encrypted secret for one account. AuthType
// determines how Secret (once decrypted) should be interpreted, e.g. "oauth2" for a
// JSON document {"refresh_token": "..."}.
type EmailAccountCredential struct {
	EmailAccountID  uuid.UUID
	AuthType        string
	EncryptedSecret []byte
}

const AuthTypeOAuth2 = "oauth2"

type EmailAccountModel struct {
	DB *sql.DB
}

// Upsert creates the account if it doesn't already exist for (user, provider, email),
// or returns the existing one, reactivating it if it had errored out or been revoked.
func (m EmailAccountModel) Upsert(account *EmailAccount) error {
	query := `
		INSERT INTO email_accounts (user_id, provider, email_address)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, provider, email_address) DO UPDATE SET
			status = $4,
			last_error = NULL,
			updated_at = NOW()
		RETURNING id, status, last_synced_at, last_uid, created_at, updated_at
	`
	args := []interface{}{account.UserID, account.Provider, account.EmailAddress, EmailAccountStatusActive}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(
		&account.ID, &account.Status, &account.LastSyncedAt, &account.LastUID, &account.CreatedAt, &account.UpdatedAt,
	)
}

// GetForUser lists every account a user has connected, most recently created first.
func (m EmailAccountModel) GetForUser(userID uuid.UUID) ([]EmailAccount, error) {
	query := `
		SELECT id, user_id, provider, email_address, status, last_error, last_synced_at, last_uid, created_at, updated_at
		FROM email_accounts
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

	accounts := []EmailAccount{}
	for rows.Next() {
		var a EmailAccount
		if err := rows.Scan(&a.ID, &a.UserID, &a.Provider, &a.EmailAddress, &a.Status, &a.LastError, &a.LastSyncedAt, &a.LastUID, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		accounts = append(accounts, a)
	}
	return accounts, rows.Err()
}

// GetActive returns every account currently eligible for background syncing.
func (m EmailAccountModel) GetActive() ([]EmailAccount, error) {
	query := `
		SELECT id, user_id, provider, email_address, status, last_error, last_synced_at, last_uid, created_at, updated_at
		FROM email_accounts
		WHERE status = $1
	`
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, EmailAccountStatusActive)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	accounts := []EmailAccount{}
	for rows.Next() {
		var a EmailAccount
		if err := rows.Scan(&a.ID, &a.UserID, &a.Provider, &a.EmailAddress, &a.Status, &a.LastError, &a.LastSyncedAt, &a.LastUID, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		accounts = append(accounts, a)
	}
	return accounts, rows.Err()
}

// Get returns one account, scoped to its owner so one user can't reach another's.
func (m EmailAccountModel) Get(id, userID uuid.UUID) (*EmailAccount, error) {
	query := `
		SELECT id, user_id, provider, email_address, status, last_error, last_synced_at, last_uid, created_at, updated_at
		FROM email_accounts
		WHERE id = $1 AND user_id = $2
	`
	var a EmailAccount

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id, userID).Scan(
		&a.ID, &a.UserID, &a.Provider, &a.EmailAddress, &a.Status, &a.LastError, &a.LastSyncedAt, &a.LastUID, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}
	return &a, nil
}

// UpdateSyncState records the outcome of a sync pass: the new UID high-water mark on
// success (err == nil), or the error and an 'error' status otherwise.
func (m EmailAccountModel) UpdateSyncState(id uuid.UUID, lastUID uint32, syncErr error) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if syncErr == nil {
		query := `
			UPDATE email_accounts
			SET last_uid = $2, last_synced_at = NOW(), status = $3, last_error = NULL, updated_at = NOW()
			WHERE id = $1
		`
		_, err := m.DB.ExecContext(ctx, query, id, lastUID, EmailAccountStatusActive)
		return err
	}

	query := `
		UPDATE email_accounts
		SET status = $2, last_error = $3, updated_at = NOW()
		WHERE id = $1
	`
	_, err := m.DB.ExecContext(ctx, query, id, EmailAccountStatusError, syncErr.Error())
	return err
}

// Delete removes an account (and, via cascade, its credentials and message history),
// scoped to its owner.
func (m EmailAccountModel) Delete(id, userID uuid.UUID) error {
	query := `DELETE FROM email_accounts WHERE id = $1 AND user_id = $2`

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
func (m EmailAccountModel) PutCredential(cred EmailAccountCredential) error {
	query := `
		INSERT INTO email_account_credentials (email_account_id, auth_type, encrypted_secret)
		VALUES ($1, $2, $3)
		ON CONFLICT (email_account_id) DO UPDATE SET
			auth_type = EXCLUDED.auth_type,
			encrypted_secret = EXCLUDED.encrypted_secret,
			updated_at = NOW()
	`
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, cred.EmailAccountID, cred.AuthType, cred.EncryptedSecret)
	return err
}

// GetCredential returns the still-encrypted secret for an account.
func (m EmailAccountModel) GetCredential(accountID uuid.UUID) (*EmailAccountCredential, error) {
	query := `
		SELECT email_account_id, auth_type, encrypted_secret
		FROM email_account_credentials
		WHERE email_account_id = $1
	`
	var cred EmailAccountCredential

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, accountID).Scan(&cred.EmailAccountID, &cred.AuthType, &cred.EncryptedSecret)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}
	return &cred, nil
}
