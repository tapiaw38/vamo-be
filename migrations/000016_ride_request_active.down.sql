DROP INDEX IF EXISTS ride_requests_one_active_per_passenger;

ALTER TABLE ride_requests
    DROP CONSTRAINT IF EXISTS ride_requests_status_check;

UPDATE ride_requests SET status = 'cancelled', updated_at = NOW() WHERE status = 'completed';

ALTER TABLE ride_requests
    ADD CONSTRAINT ride_requests_status_check
    CHECK (status IN ('open', 'accepted', 'cancelled', 'expired'));
