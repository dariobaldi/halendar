CREATE TABLE IF NOT EXISTS email_accounts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users ON DELETE CASCADE,
    provider text NOT NULL,
    email_address text NOT NULL,
    status text NOT NULL DEFAULT 'active',
    last_error text,
    last_synced_at timestamp(0) with time zone,
    last_uid bigint NOT NULL DEFAULT 0,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    updated_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, provider, email_address)
);

CREATE INDEX IF NOT EXISTS email_accounts_user_id_idx ON email_accounts (user_id);
