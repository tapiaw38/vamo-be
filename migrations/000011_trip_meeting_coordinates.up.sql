ALTER TABLE trips
    ADD COLUMN IF NOT EXISTS meeting_point_lat NUMERIC(9,6),
    ADD COLUMN IF NOT EXISTS meeting_point_lon NUMERIC(9,6);

ALTER TABLE trips
    ADD CONSTRAINT trips_meeting_coordinates_pair
    CHECK ((meeting_point_lat IS NULL AND meeting_point_lon IS NULL)
        OR (meeting_point_lat IS NOT NULL AND meeting_point_lon IS NOT NULL));
