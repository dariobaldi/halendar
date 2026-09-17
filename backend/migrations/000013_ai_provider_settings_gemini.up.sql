-- A second key column alongside encrypted_claude_api_key, so a user can connect
-- Gemini too and switch between Ollama/Claude/Gemini without re-entering either key
-- (see AIProviderSettingsModel).
ALTER TABLE ai_provider_settings ADD COLUMN IF NOT EXISTS encrypted_gemini_api_key bytea;
