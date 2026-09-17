CREATE TABLE IF NOT EXISTS calendar_accounts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users ON DELETE CASCADE,
    provider text NOT NULL,               -- 'google', 'caldav' (Apple iCloud, La Suite, Nextcloud, ...)
    display_name text NOT NULL,           -- e.g. "user@gmail.com" or "Home (iCloud)"
    config jsonb NOT NULL DEFAULT '{}',   -- provider-specific, non-secret settings (caldav: url/user/calendar/timezone)
    status text NOT NULL DEFAULT 'active',
    last_error text,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    updated_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, provider, display_name)
);

CREATE INDEX IF NOT EXISTS calendar_accounts_user_id_idx ON calendar_accounts (user_id);

-- One row per account. "encrypted_secret" holds an AES-256-GCM sealed blob of a small
-- JSON document whose shape depends on auth_type: for 'oauth2', {"refresh_token": "..."};
-- for 'basic' (CalDAV), {"password": "..."}.
CREATE TABLE IF NOT EXISTS calendar_account_credentials (
    calendar_account_id UUID PRIMARY KEY REFERENCES calendar_accounts ON DELETE CASCADE,
    auth_type text NOT NULL,
    encrypted_secret bytea NOT NULL,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    updated_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);
