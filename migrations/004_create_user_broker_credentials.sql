-- migrations/004_create_user_broker_credentials.sql

CREATE TABLE IF NOT EXISTS user_broker_credentials (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    broker VARCHAR(50) NOT NULL DEFAULT 'kite', -- In case you add other brokers later
    kite_user_id TEXT, -- Store the UserID provided by Kite Connect
    public_token TEXT,
    access_token_encrypted BYTEA NOT NULL, -- Store encrypted token as raw bytes
    token_type VARCHAR(50), -- e.g., "bearer" (if provided by Kite)
    -- expires_at TIMESTAMPTZ, -- Add if Kite provides expiry info with the token
    -- refresh_token_encrypted BYTEA, -- Add if using refresh tokens
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Ensure a user can only link one account per broker type
    UNIQUE (user_id, broker)
);

-- Index for faster lookups by user_id
CREATE INDEX IF NOT EXISTS idx_user_broker_credentials_user_id ON user_broker_credentials(user_id);

-- Add the updated_at trigger (using the existing function)
CREATE TRIGGER update_user_broker_credentials_updated_at
BEFORE UPDATE ON user_broker_credentials
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();