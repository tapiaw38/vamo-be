ALTER TABLE pricing_settings
    DROP COLUMN IF EXISTS urban_base_price,
    DROP COLUMN IF EXISTS urban_price_per_km;

ALTER TABLE ride_requests
    DROP COLUMN IF EXISTS origin_lat,
    DROP COLUMN IF EXISTS origin_lon,
    DROP COLUMN IF EXISTS destination_lat,
    DROP COLUMN IF EXISTS destination_lon,
    DROP COLUMN IF EXISTS origin_address,
    DROP COLUMN IF EXISTS destination_address;

ALTER TABLE ride_requests
    ADD CONSTRAINT ride_requests_different_localities CHECK (origin_locality_id <> destination_locality_id);
