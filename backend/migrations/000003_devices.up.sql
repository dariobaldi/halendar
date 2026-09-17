CREATE TABLE IF NOT EXISTS devices (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users ON DELETE CASCADE,
    push_token text UNIQUE NOT NULL,
    platform text NOT NULL,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    last_seen_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS devices_user_id_idx ON devices (user_id);
