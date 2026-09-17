-- One row per user, created on first use of any AI setting (see
-- AIProviderSettingsModel). "provider" picks which model email analysis uses for
-- this user; "encrypted_claude_api_key" is only set once they've connected one, and
-- provider can only be 'claude' while it's non-null (enforced in Go, not here).
CREATE TABLE IF NOT EXISTS ai_provider_settings (
    user_id UUID PRIMARY KEY REFERENCES users ON DELETE CASCADE,
    provider text NOT NULL DEFAULT 'ollama',
    encrypted_claude_api_key bytea,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    updated_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);
