CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS driver_profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'pending_review', 'verified', 'suspended')),
    license_number TEXT,
    sensitive_metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    verification_notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS driver_profiles_status_idx ON driver_profiles(status);

CREATE TABLE IF NOT EXISTS vehicles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_type TEXT NOT NULL CHECK (owner_type IN ('driver', 'company')),
    driver_profile_id UUID REFERENCES driver_profiles(id) ON DELETE RESTRICT,
    company_id UUID,
    make TEXT NOT NULL,
    model TEXT NOT NULL,
    year INTEGER NOT NULL CHECK (year BETWEEN 1900 AND 2100),
    license_plate TEXT NOT NULL,
    authorized_seats INTEGER NOT NULL CHECK (authorized_seats > 0),
    vehicle_type TEXT NOT NULL,
    verification_status TEXT NOT NULL DEFAULT 'pending' CHECK (verification_status IN ('pending', 'verified', 'rejected', 'suspended')),
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT vehicles_exactly_one_owner CHECK (
        (owner_type = 'driver' AND driver_profile_id IS NOT NULL AND company_id IS NULL)
        OR (owner_type = 'company' AND driver_profile_id IS NULL AND company_id IS NOT NULL)
    )
);

CREATE UNIQUE INDEX IF NOT EXISTS vehicles_active_license_plate_idx ON vehicles (LOWER(license_plate)) WHERE active;
CREATE INDEX IF NOT EXISTS vehicles_driver_profile_idx ON vehicles(driver_profile_id);

CREATE TABLE IF NOT EXISTS vehicle_documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vehicle_id UUID NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
    document_type TEXT NOT NULL,
    external_url TEXT NOT NULL,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    verification_status TEXT NOT NULL DEFAULT 'pending' CHECK (verification_status IN ('pending', 'verified', 'rejected')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS vehicle_documents_vehicle_idx ON vehicle_documents(vehicle_id);
