ALTER TABLE driver_profiles ADD COLUMN IF NOT EXISTS is_available BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS photo_url TEXT;

CREATE TABLE IF NOT EXISTS ride_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    passenger_id TEXT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    origin_locality_id TEXT NOT NULL REFERENCES localities(id) ON DELETE RESTRICT,
    destination_locality_id TEXT NOT NULL REFERENCES localities(id) ON DELETE RESTRICT,
    requested_at TIMESTAMPTZ NOT NULL,
    max_price_per_seat NUMERIC(14,2) NOT NULL CHECK (max_price_per_seat > 0),
    seats INTEGER NOT NULL CHECK (seats > 0),
    status TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open','accepted','cancelled','expired')),
    accepted_by_driver_profile_id UUID REFERENCES driver_profiles(id) ON DELETE RESTRICT,
    accepted_trip_id UUID REFERENCES trips(id) ON DELETE RESTRICT,
    notes TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT ride_requests_different_localities CHECK (origin_locality_id <> destination_locality_id)
);
CREATE INDEX IF NOT EXISTS ride_requests_open_origin_idx ON ride_requests(origin_locality_id, destination_locality_id, requested_at) WHERE status='open';
