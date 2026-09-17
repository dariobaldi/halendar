-- One row per account. "encrypted_secret" holds an AES-256-GCM sealed blob (nonce
-- prepended) of a small JSON document whose shape depends on auth_type, e.g. for
-- 'oauth2': {"refresh_token": "..."}. Never queried on, so one opaque column is enough.
CREATE TABLE IF NOT EXISTS email_account_credentials (
    email_account_id UUID PRIMARY KEY REFERENCES email_accounts ON DELETE CASCADE,
    auth_type text NOT NULL,
    encrypted_secret bytea NOT NULL,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    updated_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);
