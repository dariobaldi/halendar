-- The drafted reply for an extracted event, once its slots have been checked against
-- the calendar: 'accept' (a proposed slot was free) or 'decline' (none were, possibly
-- with alternative free slots suggested instead -- see email_event_slots.source).
ALTER TABLE email_message_events ADD COLUMN IF NOT EXISTS response_kind text;
ALTER TABLE email_message_events ADD COLUMN IF NOT EXISTS response_draft text;

-- 'sender': one of the times the sender proposed. 'suggested': an alternative free
-- time we found on the calendar and offered instead, when none of the sender's were
-- free.
ALTER TABLE email_event_slots ADD COLUMN IF NOT EXISTS source text NOT NULL DEFAULT 'sender';
