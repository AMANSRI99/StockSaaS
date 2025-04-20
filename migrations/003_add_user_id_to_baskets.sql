-- migrations/003_add_user_id_to_baskets.sql

-- Add the user_id column to the baskets table
-- Allow NULL initially to handle existing rows (if any)
ALTER TABLE baskets
ADD COLUMN IF NOT EXISTS user_id UUID;

-- IMPORTANT: Decide how to handle existing baskets without a user_id.
-- Option 1: Delete them (if this is acceptable in development)
-- DELETE FROM basket_items WHERE basket_id IN (SELECT id FROM baskets WHERE user_id IS NULL); -- Delete items first due to FK if needed later
-- DELETE FROM baskets WHERE user_id IS NULL;

-- Option 2: Assign them to a default user (Requires creating a default user first in migration 002 or manually)
-- UPDATE baskets SET user_id = 'your-default-user-uuid' WHERE user_id IS NULL;

-- Option 3: Keep this migration simple for now if NO data exists or deletion is OK.

-- Once existing rows are handled (or if none exist), make the column NOT NULL
-- If you have existing rows you didn't delete/update, this command WILL FAIL.
-- COMMENT OUT the following line if you need to handle existing data later.
ALTER TABLE baskets
ALTER COLUMN user_id SET NOT NULL;

-- Add the foreign key constraint to link baskets to users
-- ON DELETE CASCADE means if a user is deleted, their baskets (and items via the other cascade) are also deleted.
ALTER TABLE baskets
ADD CONSTRAINT fk_baskets_user_id
FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- Add an index for faster lookups of baskets by user_id
CREATE INDEX IF NOT EXISTS idx_baskets_user_id ON baskets(user_id);