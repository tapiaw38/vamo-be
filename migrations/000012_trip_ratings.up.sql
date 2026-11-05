CREATE TABLE IF NOT EXISTS trip_ratings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trip_id UUID NOT NULL REFERENCES trips(id) ON DELETE CASCADE,
    booking_id UUID NOT NULL UNIQUE REFERENCES bookings(id) ON DELETE CASCADE,
    passenger_id TEXT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    driver_profile_id UUID NOT NULL REFERENCES driver_profiles(id) ON DELETE RESTRICT,
    score INTEGER NOT NULL CHECK (score BETWEEN 1 AND 5),
    comment TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS trip_ratings_driver_idx ON trip_ratings(driver_profile_id);
