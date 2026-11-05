ALTER TABLE trips DROP CONSTRAINT IF EXISTS trips_exactly_one_publisher;
ALTER TABLE trips ADD CONSTRAINT trips_exactly_one_publisher CHECK (
    (publisher_type = 'driver' AND driver_profile_id IS NOT NULL AND company_id IS NULL)
    OR (publisher_type = 'company' AND driver_profile_id IS NULL AND company_id IS NOT NULL)
);
ALTER TABLE trips DROP CONSTRAINT IF EXISTS trips_company_fk;
ALTER TABLE vehicles DROP CONSTRAINT IF EXISTS vehicles_company_fk;
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS company_members;
DROP TABLE IF EXISTS companies;
