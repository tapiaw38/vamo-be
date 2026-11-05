package integration

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	paymentrepo "github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/bookingpayment"
	"os"
	"testing"
	"time"
)

func TestBookingPaymentExternalReferenceIsIdempotent(t *testing.T) {
	url := os.Getenv("VAMO_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("set VAMO_TEST_DATABASE_URL after migrations")
	}
	db, e := pgxpool.New(context.Background(), url)
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	ctx := context.Background()
	user, driver, dp, vehicle, trip, booking := "phase7-user", "phase7-driver", "00000000-0000-0000-0000-000000000014", "00000000-0000-0000-0000-000000000015", "00000000-0000-0000-0000-000000000016", "00000000-0000-0000-0000-000000000017"
	cleanup := func() {
		_, _ = db.Exec(ctx, `DELETE FROM booking_payments WHERE booking_id=$1`, booking)
		_, _ = db.Exec(ctx, `DELETE FROM bookings WHERE id=$1`, booking)
		_, _ = db.Exec(ctx, `DELETE FROM trips WHERE id=$1`, trip)
		_, _ = db.Exec(ctx, `DELETE FROM vehicles WHERE id=$1`, vehicle)
		_, _ = db.Exec(ctx, `DELETE FROM driver_profiles WHERE id=$1`, dp)
		_, _ = db.Exec(ctx, `DELETE FROM users WHERE id IN ($1,$2)`, user, driver)
	}
	cleanup()
	defer cleanup()
	for _, id := range []string{user, driver} {
		if _, e = db.Exec(ctx, `INSERT INTO users(id) VALUES($1)`, id); e != nil {
			t.Fatal(e)
		}
	}
	var origin, destination string
	if e = db.QueryRow(ctx, `SELECT id FROM localities ORDER BY id LIMIT 1`).Scan(&origin); e != nil {
		t.Fatal(e)
	}
	if e = db.QueryRow(ctx, `SELECT id FROM localities WHERE id<>$1 ORDER BY id LIMIT 1`, origin).Scan(&destination); e != nil {
		t.Fatal(e)
	}
	if _, e = db.Exec(ctx, `INSERT INTO driver_profiles(id,user_id) VALUES($1,$2)`, dp, driver); e != nil {
		t.Fatal(e)
	}
	if _, e = db.Exec(ctx, `INSERT INTO vehicles(id,owner_type,driver_profile_id,make,model,year,license_plate,authorized_seats,vehicle_type) VALUES($1,'driver',$2,'T','P',2024,'P700PAA',2,'car')`, vehicle, dp); e != nil {
		t.Fatal(e)
	}
	if _, e = db.Exec(ctx, `INSERT INTO trips(id,publisher_type,driver_profile_id,vehicle_id,origin_locality_id,destination_locality_id,departure_at,price_per_seat,total_seats,available_seats,status) VALUES($1,'driver',$2,$3,$4,$5,$6,1500,2,1,'published')`, trip, dp, vehicle, origin, destination, time.Now().Add(24*time.Hour)); e != nil {
		t.Fatal(e)
	}
	if _, e = db.Exec(ctx, `INSERT INTO bookings(id,trip_id,passenger_id,seats,unit_price_snapshot,total_price,currency,status) VALUES($1,$2,$3,1,1500,1500,'ARS','confirmed')`, booking, trip, user); e != nil {
		t.Fatal(e)
	}
	repo := paymentrepo.NewRepository(db)
	first, created, e := repo.Prepare(ctx, booking, user)
	if e != nil || !created {
		t.Fatalf("first prepare failed created=%v err=%v", created, e)
	}
	second, created, e := repo.Prepare(ctx, booking, user)
	if e != nil || created {
		t.Fatalf("second prepare not idempotent created=%v err=%v", created, e)
	}
	if first.ID != second.ID || first.ExternalReference != "vamo:booking:"+booking {
		t.Fatalf("unexpected idempotency result first=%+v second=%+v", first, second)
	}
	var count int
	if e = db.QueryRow(ctx, `SELECT COUNT(*) FROM booking_payments WHERE booking_id=$1`, booking).Scan(&count); e != nil {
		t.Fatal(e)
	}
	if count != 1 {
		t.Fatalf("expected one booking payment, got %d", count)
	}
}
