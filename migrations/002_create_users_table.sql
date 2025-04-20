-- migrations/002_create_users_table.sql

-- Create the users table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    email TEXT UNIQUE NOT NULL CHECK (email <> ''), -- Ensure email is unique and not empty
    password_hash TEXT NOT NULL CHECK (password_hash <> ''), -- Store the bcrypt hash
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Add an index on the lowercase version of the email for case-insensitive uniqueness checks and faster lookups
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_lower_email ON users(LOWER(email));

-- Add the trigger function for updated_at (if not already created globally)
-- If you created this in migration 001, you might not need it again,
-- but it's safe to include CREATE OR REPLACE FUNCTION.
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
   NEW.updated_at = NOW();
   RETURN NEW;
END;
$$ language 'plpgsql';

-- Add the trigger to the new users table
CREATE TRIGGER update_users_updated_at
BEFORE UPDATE ON users
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();