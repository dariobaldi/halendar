-- The structured event a message was extracted to represent, created only when the AI
-- decides the sender is asking our user to participate in something -- as opposed to
-- merely mentioning an event of the sender's own, a past event, or an automated notice.
CREATE TABLE IF NOT EXISTS email_message_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email_message_id UUID NOT NULL UNIQUE REFERENCES email_messages ON DELETE CASCADE,
    title text NOT NULL DEFAULT '',
    location text NOT NULL DEFAULT '',
    needs_manual_review boolean NOT NULL DEFAULT FALSE,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);

-- One row per candidate date/time the sender proposed or asked about, in the order the
-- AI returned them (its implied preference), each checked against the calendar.
CREATE TABLE IF NOT EXISTS email_event_slots (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email_message_event_id UUID NOT NULL REFERENCES email_message_events ON DELETE CASCADE,
    start_at timestamp(0) with time zone NOT NULL,
    end_at timestamp(0) with time zone NOT NULL,
    availability text NOT NULL DEFAULT 'unknown',
    position smallint NOT NULL DEFAULT 0,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS email_event_slots_event_id_idx ON email_event_slots (email_message_event_id);
