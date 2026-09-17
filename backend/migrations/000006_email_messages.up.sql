-- History of every message imported from a connected account, and what the AI
-- analysis pipeline did with it.
CREATE TABLE IF NOT EXISTS email_messages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email_account_id UUID NOT NULL REFERENCES email_accounts ON DELETE CASCADE,
    provider_message_id text NOT NULL,
    imap_uid bigint,
    from_address text NOT NULL,
    from_name text,
    subject text,
    received_at timestamp(0) with time zone,
    snippet text,
    analysis_status text NOT NULL DEFAULT 'pending',
    has_event boolean,
    analysis_result jsonb,
    analyzed_at timestamp(0) with time zone,
    analysis_error text,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    UNIQUE (email_account_id, provider_message_id)
);

CREATE INDEX IF NOT EXISTS email_messages_account_id_idx ON email_messages (email_account_id);
CREATE INDEX IF NOT EXISTS email_messages_status_idx ON email_messages (analysis_status);
