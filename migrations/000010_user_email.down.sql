DROP INDEX IF EXISTS users_search_email_idx;
DROP INDEX IF EXISTS users_search_name_idx;
ALTER TABLE users DROP COLUMN IF EXISTS email;
