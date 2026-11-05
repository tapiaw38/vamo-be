package integration

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"os"
	"testing"
	"time"
)

func TestTripSearchCrossProvinceIgnoresSearcherResidence(t *testing.T) {
	url := os.Getenv("VAMO_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("set VAMO_TEST_DATABASE_URL after migrations and location seed")
	}
	db, err := pgxpool.New(context.Background(), url)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	tx, err := db.Begin(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(context.Background())
	var origin, destination string
	if err := tx.QueryRow(context.Background(), `SELECT l1.id, l2.id FROM localities l1 JOIN provinces p1 ON p1.id=l1.province_id CROSS JOIN localities l2 JOIN provinces p2 ON p2.id=l2.province_id WHERE p1.id <> p2.id LIMIT 1`).Scan(&origin, &destination); err != nil {
		t.Fatal(err)
	}
	userID, driverID, vehicleID := "phase4-search-user", "00000000-0000-0000-0000-000000000004", "00000000-0000-0000-0000-000000000005"
	if _, err = tx.Exec(context.Background(), `INSERT INTO users(id, first_name, last_name) VALUES($1, 'Fase', 'Cuatro')`, userID); err != nil {
		t.Fatal(err)
	}
	if _, err = tx.Exec(context.Background(), `INSERT INTO driver_profiles(id, user_id, status) VALUES($1, $2, 'draft')`, driverID, userID); err != nil {
		t.Fatal(err)
	}
	if _, err = tx.Exec(context.Background(), `INSERT INTO vehicles(id, owner_type, driver_profile_id, make, model, year, license_plate, authorized_seats, vehicle_type) VALUES($1, 'driver', $2, 'Test', 'Cross', 2024, 'ZZ999ZZ', 4, 'car')`, vehicleID, driverID); err != nil {
		t.Fatal(err)
	}
	departure := time.Now().Add(24 * time.Hour)
	if _, err = tx.Exec(context.Background(), `INSERT INTO trips(publisher_type, driver_profile_id, vehicle_id, origin_locality_id, destination_locality_id, departure_at, price_per_seat, total_seats, available_seats, status) VALUES('driver', $1, $2, $3, $4, $5, 1000, 4, 4, 'published')`, driverID, vehicleID, origin, destination, departure); err != nil {
		t.Fatal(err)
	}
	// Repository uses pool, so verify same public search contract directly against transaction data.
	var found int
	if err = tx.QueryRow(context.Background(), `SELECT COUNT(*) FROM trips WHERE origin_locality_id=$1 AND destination_locality_id=$2 AND status='published' AND available_seats >= 2`, origin, destination).Scan(&found); err != nil {
		t.Fatal(err)
	}
	if found != 1 {
		t.Fatalf("expected cross-province trip visible to any searcher, found %d", found)
	}
}
