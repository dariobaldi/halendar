package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

// Provider values for AIProviderSettings.
const (
	AIProviderOllama = "ollama"
	AIProviderClaude = "claude"
)

// AIProviderSettings is what the frontend needs to render the AI settings screen --
// the key itself is never exposed, only whether one is on file.
type AIProviderSettings struct {
	Provider  string `json:"provider"`
	HasAPIKey bool   `json:"has_api_key"`
}

type AIProviderSettingsModel struct {
	DB *sql.DB
}

// Get returns the user's AI settings, defaulting to Ollama with no key if they've
// never touched this (no row exists yet -- there's nothing to migrate old users into).
func (m AIProviderSettingsModel) Get(userID uuid.UUID) (AIProviderSettings, error) {
	query := `SELECT provider, encrypted_claude_api_key FROM ai_provider_settings WHERE user_id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var settings AIProviderSettings
	var key []byte
	err := m.DB.QueryRowContext(ctx, query, userID).Scan(&settings.Provider, &key)
	if errors.Is(err, sql.ErrNoRows) {
		return AIProviderSettings{Provider: AIProviderOllama}, nil
	}
	if err != nil {
		return AIProviderSettings{}, err
	}
	settings.HasAPIKey = len(key) > 0
	return settings, nil
}

// SetAPIKey stores (inserting or replacing) the user's encrypted Claude API key.
// Doesn't change which provider is active -- see SetProvider.
func (m AIProviderSettingsModel) SetAPIKey(userID uuid.UUID, encryptedKey []byte) error {
	query := `
		INSERT INTO ai_provider_settings (user_id, encrypted_claude_api_key)
		VALUES ($1, $2)
		ON CONFLICT (user_id) DO UPDATE SET
			encrypted_claude_api_key = EXCLUDED.encrypted_claude_api_key,
			updated_at = NOW()
	`
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, userID, encryptedKey)
	return err
}

// ClearAPIKey removes the stored key and forces the provider back to Ollama -- there's
// nothing left to authenticate Claude requests with otherwise.
func (m AIProviderSettingsModel) ClearAPIKey(userID uuid.UUID) error {
	query := `
		UPDATE ai_provider_settings
		SET encrypted_claude_api_key = NULL, provider = $2, updated_at = NOW()
		WHERE user_id = $1
	`
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, userID, AIProviderOllama)
	return err
}

// SetProvider switches which model future analysis uses for this user.
func (m AIProviderSettingsModel) SetProvider(userID uuid.UUID, provider string) error {
	query := `
		INSERT INTO ai_provider_settings (user_id, provider)
		VALUES ($1, $2)
		ON CONFLICT (user_id) DO UPDATE SET
			provider = EXCLUDED.provider,
			updated_at = NOW()
	`
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, userID, provider)
	return err
}

// GetEncryptedAPIKey returns the still-encrypted key, or ErrRecordNotFound if none is
// on file.
func (m AIProviderSettingsModel) GetEncryptedAPIKey(userID uuid.UUID) ([]byte, error) {
	query := `SELECT encrypted_claude_api_key FROM ai_provider_settings WHERE user_id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var key []byte
	err := m.DB.QueryRowContext(ctx, query, userID).Scan(&key)
	if errors.Is(err, sql.ErrNoRows) || (err == nil && len(key) == 0) {
		return nil, ErrRecordNotFound
	}
	if err != nil {
		return nil, err
	}
	return key, nil
}
