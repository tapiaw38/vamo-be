ALTER TABLE ride_requests
    DROP CONSTRAINT IF EXISTS ride_requests_status_check;

WITH ranked AS (
    SELECT id,
           ROW_NUMBER() OVER (PARTITION BY passenger_id ORDER BY created_at DESC, id DESC) AS position
    FROM ride_requests
    WHERE status IN ('open', 'accepted')
)
UPDATE ride_requests AS rr
SET status = 'cancelled', updated_at = NOW()
FROM ranked
WHERE rr.id = ranked.id
  AND ranked.position > 1;

ALTER TABLE ride_requests
    ADD CONSTRAINT ride_requests_status_check
    CHECK (status IN ('open', 'accepted', 'cancelled', 'expired', 'completed'));

CREATE UNIQUE INDEX IF NOT EXISTS ride_requests_one_active_per_passenger
    ON ride_requests(passenger_id)
    WHERE status IN ('open', 'accepted');
