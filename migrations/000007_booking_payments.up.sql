CREATE TABLE IF NOT EXISTS booking_payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    booking_id UUID NOT NULL UNIQUE REFERENCES bookings(id) ON DELETE RESTRICT,
    provider TEXT NOT NULL,
    external_reference TEXT NOT NULL UNIQUE,
    provider_payment_id TEXT,
    amount NUMERIC(14,2) NOT NULL CHECK (amount > 0),
    currency TEXT NOT NULL DEFAULT 'ARS',
    status TEXT NOT NULL CHECK (status IN ('initiated', 'pending', 'approved', 'rejected', 'cancelled', 'refunded')),
    idempotency_key TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS booking_payments_status_idx ON booking_payments(status);
