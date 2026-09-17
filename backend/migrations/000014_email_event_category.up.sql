-- Set only when the AI decided a message isn't a genuine meeting request, so the
-- Messages page can tell "not a meeting, but a real person wrote it" apart from
-- "promotional/automated, probably safe to skip outright" (see Proposal.SuggestedSkip
-- in internal/data/proposals.go). Empty string for a genuine meeting request, and for
-- rows written before this column existed until they're re-analyzed.
ALTER TABLE email_message_events ADD COLUMN IF NOT EXISTS category text NOT NULL DEFAULT '';
