ALTER TABLE trips DROP CONSTRAINT IF EXISTS trips_meeting_coordinates_pair;
ALTER TABLE trips DROP COLUMN IF EXISTS meeting_point_lat;
ALTER TABLE trips DROP COLUMN IF EXISTS meeting_point_lon;
