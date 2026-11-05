ALTER TABLE users ADD COLUMN IF NOT EXISTS email TEXT NOT NULL DEFAULT '';
CREATE INDEX IF NOT EXISTS users_search_name_idx ON users (LOWER(first_name), LOWER(last_name));
CREATE INDEX IF NOT EXISTS users_search_email_idx ON users (LOWER(email));
