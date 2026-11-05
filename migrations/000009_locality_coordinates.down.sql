DROP INDEX IF EXISTS localities_coordinates_idx;
ALTER TABLE localities DROP COLUMN IF EXISTS lat;
ALTER TABLE localities DROP COLUMN IF EXISTS lon;
