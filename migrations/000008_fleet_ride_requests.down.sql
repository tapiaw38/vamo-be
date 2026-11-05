DROP TABLE IF EXISTS ride_requests;
ALTER TABLE vehicles DROP COLUMN IF EXISTS photo_url;
ALTER TABLE driver_profiles DROP COLUMN IF EXISTS is_available;
