-- Tracks the user's decision on a proposal (the frontend's "needs action" vs
-- "history" split) and which slot they've settled on, if they've overridden the
-- one auto-selected at analysis time.
ALTER TABLE email_message_events ADD COLUMN IF NOT EXISTS status text NOT NULL DEFAULT 'pending';
ALTER TABLE email_message_events ADD COLUMN IF NOT EXISTS selected_slot_id UUID REFERENCES email_event_slots ON DELETE SET NULL;

-- Once true, re-selecting a slot no longer silently overwrites the user's own
-- wording -- mirrors Proposal.draftEditedByUser on the frontend.
ALTER TABLE email_message_events ADD COLUMN IF NOT EXISTS draft_edited_by_user boolean NOT NULL DEFAULT FALSE;
