-- The full plain-text body, kept alongside the short "snippet" used for list views --
-- the frontend needs the whole thing to show the user what they'd be accepting or
-- declining, and it also lets re-analysis run without re-fetching from the provider.
ALTER TABLE email_messages ADD COLUMN IF NOT EXISTS body text NOT NULL DEFAULT '';
