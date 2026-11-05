CREATE TABLE IF NOT EXISTS bookings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trip_id UUID NOT NULL REFERENCES trips(id) ON DELETE RESTRICT,
    passenger_id TEXT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    seats INTEGER NOT NULL CHECK (seats > 0),
    unit_price_snapshot NUMERIC(14,2) NOT NULL CHECK (unit_price_snapshot > 0),
    total_price NUMERIC(14,2) NOT NULL CHECK (total_price > 0),
    currency TEXT NOT NULL DEFAULT 'ARS',
    status TEXT NOT NULL CHECK (status IN ('pending', 'confirmed', 'rejected', 'cancelled', 'expired', 'completed')),
    requested_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    confirmed_at TIMESTAMPTZ,
    cancelled_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS bookings_active_passenger_trip_idx
    ON bookings(trip_id, passenger_id)
    WHERE status IN ('pending', 'confirmed');
CREATE INDEX IF NOT EXISTS bookings_passenger_status_idx ON bookings(passenger_id, status);
CREATE INDEX IF NOT EXISTS bookings_trip_status_idx ON bookings(trip_id, status);

CREATE TABLE IF NOT EXISTS booking_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    booking_id UUID NOT NULL REFERENCES bookings(id) ON DELETE CASCADE,
    from_status TEXT,
    to_status TEXT NOT NULL,
    actor_user_id TEXT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    reason TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS booking_events_booking_idx ON booking_events(booking_id, created_at);
