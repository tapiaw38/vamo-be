package integration

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	bookingrepo "github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/booking"
	"os"
	"sync"
	"testing"
	"time"
)

func TestConcurrentBookingsOnlyOneGetsLastSeat(t *testing.T) {
	url := os.Getenv("VAMO_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("set VAMO_TEST_DATABASE_URL after migrations and location seed")
	}
	db, err := pgxpool.New(context.Background(), url)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx := context.Background()
	driverUser, passengerOne, passengerTwo := "phase5-driver", "phase5-passenger-one", "phase5-passenger-two"
	driverID, vehicleID, tripID := "00000000-0000-0000-0000-000000000006", "00000000-0000-0000-0000-000000000007", "00000000-0000-0000-0000-000000000008"
	cleanup := func() {
		_, _ = db.Exec(ctx, `DELETE FROM bookings WHERE trip_id=$1`, tripID)
		_, _ = db.Exec(ctx, `DELETE FROM trips WHERE id=$1`, tripID)
		_, _ = db.Exec(ctx, `DELETE FROM vehicles WHERE id=$1`, vehicleID)
		_, _ = db.Exec(ctx, `DELETE FROM driver_profiles WHERE id=$1`, driverID)
		_, _ = db.Exec(ctx, `DELETE FROM users WHERE id IN ($1,$2,$3)`, driverUser, passengerOne, passengerTwo)
	}
	cleanup()
	defer cleanup()
	var origin, destination string
	if err = db.QueryRow(ctx, `SELECT l1.id, l2.id FROM localities l1 CROSS JOIN localities l2 WHERE l1.id <> l2.id LIMIT 1`).Scan(&origin, &destination); err != nil {
		t.Fatal(err)
	}
	for _, userID := range []string{driverUser, passengerOne, passengerTwo} {
		if _, err = db.Exec(ctx, `INSERT INTO users(id) VALUES($1)`, userID); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = db.Exec(ctx, `INSERT INTO driver_profiles(id,user_id,status) VALUES($1,$2,'draft')`, driverID, driverUser); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ctx, `INSERT INTO vehicles(id,owner_type,driver_profile_id,make,model,year,license_plate,authorized_seats,vehicle_type) VALUES($1,'driver',$2,'Test','Booking',2024,'P555PAA',1,'car')`, vehicleID, driverID); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ctx, `INSERT INTO trips(id,publisher_type,driver_profile_id,vehicle_id,origin_locality_id,destination_locality_id,departure_at,price_per_seat,total_seats,available_seats,status) VALUES($1,'driver',$2,$3,$4,$5,$6,1000,1,1,'published')`, tripID, driverID, vehicleID, origin, destination, time.Now().Add(24*time.Hour)); err != nil {
		t.Fatal(err)
	}
	repo := bookingrepo.NewRepository(db)
	var wg sync.WaitGroup
	results := make(chan error, 2)
	for _, passengerID := range []string{passengerOne, passengerTwo} {
		wg.Add(1)
		go func(id string) { defer wg.Done(); _, callErr := repo.Create(ctx, id, tripID, 1); results <- callErr }(passengerID)
	}
	wg.Wait()
	close(results)
	wins := 0
	for callErr := range results {
		if callErr == nil {
			wins++
		}
	}
	if wins != 1 {
		t.Fatalf("expected one successful booking for last seat, got %d", wins)
	}
	var available int
	if err = db.QueryRow(ctx, `SELECT available_seats FROM trips WHERE id=$1`, tripID).Scan(&available); err != nil {
		t.Fatal(err)
	}
	if available != 0 {
		t.Fatalf("expected zero available seats, got %d", available)
	}
}
