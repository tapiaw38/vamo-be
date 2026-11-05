ALTER TABLE ride_requests
    DROP CONSTRAINT IF EXISTS ride_requests_different_localities;

ALTER TABLE ride_requests
    ADD COLUMN IF NOT EXISTS origin_lat NUMERIC(9,6),
    ADD COLUMN IF NOT EXISTS origin_lon NUMERIC(9,6),
    ADD COLUMN IF NOT EXISTS destination_lat NUMERIC(9,6),
    ADD COLUMN IF NOT EXISTS destination_lon NUMERIC(9,6),
    ADD COLUMN IF NOT EXISTS origin_address TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS destination_address TEXT NOT NULL DEFAULT '';

ALTER TABLE pricing_settings
    ADD COLUMN IF NOT EXISTS urban_base_price NUMERIC(14,2) NOT NULL DEFAULT 1200.00,
    ADD COLUMN IF NOT EXISTS urban_price_per_km NUMERIC(14,2) NOT NULL DEFAULT 400.00;
