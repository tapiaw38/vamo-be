ALTER TABLE trips
    DROP CONSTRAINT IF EXISTS trips_meeting_coordinates_pair,
    DROP COLUMN IF EXISTS meeting_point,
    DROP COLUMN IF EXISTS meeting_point_lat,
    DROP COLUMN IF EXISTS meeting_point_lon;
