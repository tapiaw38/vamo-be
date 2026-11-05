CREATE TABLE IF NOT EXISTS trips (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    publisher_type TEXT NOT NULL CHECK (publisher_type IN ('driver', 'company')),
    driver_profile_id UUID REFERENCES driver_profiles(id) ON DELETE RESTRICT,
    company_id UUID,
    vehicle_id UUID NOT NULL REFERENCES vehicles(id) ON DELETE RESTRICT,
    origin_locality_id TEXT NOT NULL REFERENCES localities(id) ON DELETE RESTRICT,
    destination_locality_id TEXT NOT NULL REFERENCES localities(id) ON DELETE RESTRICT,
    departure_at TIMESTAMPTZ NOT NULL,
    arrival_at TIMESTAMPTZ,
    timezone TEXT NOT NULL DEFAULT 'America/Argentina/Buenos_Aires',
    price_per_seat NUMERIC(14,2) NOT NULL CHECK (price_per_seat > 0),
    currency TEXT NOT NULL DEFAULT 'ARS',
    total_seats INTEGER NOT NULL CHECK (total_seats > 0),
    available_seats INTEGER NOT NULL CHECK (available_seats >= 0 AND available_seats <= total_seats),
    booking_mode TEXT NOT NULL DEFAULT 'automatic' CHECK (booking_mode IN ('automatic', 'manual')),
    status TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'published', 'full', 'departed', 'completed', 'cancelled')),
    meeting_point TEXT NOT NULL DEFAULT '',
    notes TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT trips_different_localities CHECK (origin_locality_id <> destination_locality_id),
    CONSTRAINT trips_exactly_one_publisher CHECK (
        (publisher_type = 'driver' AND driver_profile_id IS NOT NULL AND company_id IS NULL)
        OR (publisher_type = 'company' AND driver_profile_id IS NULL AND company_id IS NOT NULL)
    )
);

CREATE INDEX IF NOT EXISTS trips_search_idx ON trips(origin_locality_id, destination_locality_id, departure_at, status);
CREATE INDEX IF NOT EXISTS trips_driver_idx ON trips(driver_profile_id);
CREATE INDEX IF NOT EXISTS trips_departure_idx ON trips(departure_at);
