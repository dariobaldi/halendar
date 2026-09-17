ALTER TABLE email_event_slots DROP COLUMN IF EXISTS source;
ALTER TABLE email_message_events DROP COLUMN IF EXISTS response_draft;
ALTER TABLE email_message_events DROP COLUMN IF EXISTS response_kind;
