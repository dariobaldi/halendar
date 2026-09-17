package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Provider values for AIProviderSettings. Ollama needs no key; Claude and Gemini each
// need their own, connected separately (see SetAPIKey) -- a user can have both on
// file and switch between them without re-entering either.
const (
	AIProviderOllama = "ollama"
	AIProviderClaude = "claude"
	AIProviderGemini = "gemini"
)

// AIProviderSettings is what the frontend needs to render the AI settings screen --
// keys themselves are never exposed, only whether one is on file for each provider.
type AIProviderSettings struct {
	Provider     string `json:"provider"`
	HasClaudeKey bool   `json:"has_claude_key"`
	HasGeminiKey bool   `json:"has_gemini_key"`
}

type AIProviderSettingsModel struct {
	DB *sql.DB
}

// keyColumn maps a key-bearing provider to its column in ai_provider_settings.
// Ollama needs no key and isn't a valid argument here.
func keyColumn(provider string) (string, error) {
	switch provider {
	case AIProviderClaude:
		return "encrypted_claude_api_key", nil
	case AIProviderGemini:
		return "encrypted_gemini_api_key", nil
	default:
		return "", fmt.Errorf("data: %q is not a key-bearing AI provider", provider)
	}
}

// Get returns the user's AI settings, defaulting to Ollama with no keys if they've
// never touched this (no row exists yet -- there's nothing to migrate old users into).
func (m AIProviderSettingsModel) Get(userID uuid.UUID) (AIProviderSettings, error) {
	query := `SELECT provider, encrypted_claude_api_key, encrypted_gemini_api_key FROM ai_provider_settings WHERE user_id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var settings AIProviderSettings
	var claudeKey, geminiKey []byte
	err := m.DB.QueryRowContext(ctx, query, userID).Scan(&settings.Provider, &claudeKey, &geminiKey)
	if errors.Is(err, sql.ErrNoRows) {
		return AIProviderSettings{Provider: AIProviderOllama}, nil
	}
	if err != nil {
		return AIProviderSettings{}, err
	}
	settings.HasClaudeKey = len(claudeKey) > 0
	settings.HasGeminiKey = len(geminiKey) > 0
	return settings, nil
}

// SetAPIKey stores (inserting or replacing) the user's encrypted API key for provider
// (claude or gemini). Doesn't change which provider is active -- see SetProvider.
func (m AIProviderSettingsModel) SetAPIKey(userID uuid.UUID, provider string, encryptedKey []byte) error {
	column, err := keyColumn(provider)
	if err != nil {
		return err
	}
	query := fmt.Sprintf(`
		INSERT INTO ai_provider_settings (user_id, %s)
		VALUES ($1, $2)
		ON CONFLICT (user_id) DO UPDATE SET
			%s = EXCLUDED.%s,
			updated_at = NOW()
	`, column, column, column)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err = m.DB.ExecContext(ctx, query, userID, encryptedKey)
	return err
}

// ClearAPIKey removes provider's stored key. If provider was the active one, the
// active provider falls back to Ollama -- there's nothing left to authenticate its
// requests with otherwise; if the user had switched to a different provider already
// (e.g. they have both Claude and Gemini connected and are using Gemini), that choice
// is left alone.
func (m AIProviderSettingsModel) ClearAPIKey(userID uuid.UUID, provider string) error {
	column, err := keyColumn(provider)
	if err != nil {
		return err
	}
	query := fmt.Sprintf(`
		UPDATE ai_provider_settings
		SET %s = NULL, provider = CASE WHEN provider = $2 THEN $3 ELSE provider END, updated_at = NOW()
		WHERE user_id = $1
	`, column)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err = m.DB.ExecContext(ctx, query, userID, provider, AIProviderOllama)
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

// GetEncryptedAPIKey returns provider's still-encrypted key, or ErrRecordNotFound if
// none is on file.
func (m AIProviderSettingsModel) GetEncryptedAPIKey(userID uuid.UUID, provider string) ([]byte, error) {
	column, err := keyColumn(provider)
	if err != nil {
		return nil, err
	}
	query := fmt.Sprintf(`SELECT %s FROM ai_provider_settings WHERE user_id = $1`, column)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var key []byte
	err = m.DB.QueryRowContext(ctx, query, userID).Scan(&key)
	if errors.Is(err, sql.ErrNoRows) || (err == nil && len(key) == 0) {
		return nil, ErrRecordNotFound
	}
	if err != nil {
		return nil, err
	}
	return key, nil
}
