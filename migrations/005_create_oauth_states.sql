-- migrations/005_create_oauth_states.sql

CREATE TABLE IF NOT EXISTS oauth_states (
    state_token TEXT PRIMARY KEY, -- The random state string
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL
);

-- Index for cleaning up expired tokens later (optional)
CREATE INDEX IF NOT EXISTS idx_oauth_states_expires_at ON oauth_states(expires_at);