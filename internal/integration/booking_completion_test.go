package integration

import (
	"context"
	"testing"
	"time"

	bookingrepo "github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/booking"
)

func TestLazyBookingCompletionAgainstPostgres(t *testing.T) {
	db := phase6DB(t)
	defer db.Close()
	ctx := context.Background()
	const driverUser = "lazy-completion-driver"
	const pastPassenger = "lazy-completion-past"
	const futurePassenger = "lazy-completion-future"
	const pendingPassenger = "lazy-completion-pending"
	const driverID = "00000000-0000-0000-0000-000000000041"
	const vehicleID = "00000000-0000-0000-0000-000000000042"
	const pastTripID = "00000000-0000-0000-0000-000000000043"
	const futureTripID = "00000000-0000-0000-0000-000000000044"
	const pastBookingID = "00000000-0000-0000-0000-000000000045"
	const futureBookingID = "00000000-0000-0000-0000-000000000046"
	const pendingBookingID = "00000000-0000-0000-0000-000000000047"
	cleanup := func() {
		_, _ = db.Exec(ctx, `DELETE FROM booking_events WHERE booking_id IN ($1,$2,$3)`, pastBookingID, futureBookingID, pendingBookingID)
		_, _ = db.Exec(ctx, `DELETE FROM bookings WHERE id IN ($1,$2,$3)`, pastBookingID, futureBookingID, pendingBookingID)
		_, _ = db.Exec(ctx, `DELETE FROM trips WHERE id IN ($1,$2)`, pastTripID, futureTripID)
		_, _ = db.Exec(ctx, `DELETE FROM vehicles WHERE id=$1`, vehicleID)
		_, _ = db.Exec(ctx, `DELETE FROM driver_profiles WHERE id=$1`, driverID)
		_, _ = db.Exec(ctx, `DELETE FROM users WHERE id IN ($1,$2,$3,$4)`, driverUser, pastPassenger, futurePassenger, pendingPassenger)
	}
	cleanup()
	defer cleanup()

	var origin, destination string
	if err := db.QueryRow(ctx, `SELECT l1.id, l2.id FROM localities l1 CROSS JOIN localities l2 WHERE l1.id <> l2.id LIMIT 1`).Scan(&origin, &destination); err != nil {
		t.Fatal(err)
	}
	for _, userID := range []string{driverUser, pastPassenger, futurePassenger, pendingPassenger} {
		if _, err := db.Exec(ctx, `INSERT INTO users(id) VALUES($1)`, userID); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := db.Exec(ctx, `INSERT INTO driver_profiles(id,user_id,status) VALUES($1,$2,'verified')`, driverID, driverUser); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(ctx, `INSERT INTO vehicles(id,owner_type,driver_profile_id,make,model,year,license_plate,authorized_seats,vehicle_type) VALUES($1,'driver',$2,'Lazy','Completion',2024,'LAZY0041',4,'car')`, vehicleID, driverID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(ctx, `INSERT INTO trips(id,publisher_type,driver_profile_id,vehicle_id,origin_locality_id,destination_locality_id,departure_at,price_per_seat,total_seats,available_seats,status) VALUES($1,'driver',$2,$3,$4,$5,$6,1000,4,4,'published'),($7,'driver',$2,$3,$4,$5,$8,1000,4,4,'published')`, pastTripID, driverID, vehicleID, origin, destination, time.Now().Add(-time.Hour), futureTripID, time.Now().Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	for _, item := range []struct{ id, trip, passenger, status string }{
		{pastBookingID, pastTripID, pastPassenger, "confirmed"},
		{futureBookingID, futureTripID, futurePassenger, "confirmed"},
		{pendingBookingID, pastTripID, pendingPassenger, "pending"},
	} {
		if _, err := db.Exec(ctx, `INSERT INTO bookings(id,trip_id,passenger_id,seats,unit_price_snapshot,total_price,status) VALUES($1,$2,$3,1,1000,1000,$4)`, item.id, item.trip, item.passenger, item.status); err != nil {
			t.Fatal(err)
		}
	}

	repo := bookingrepo.NewRepository(db)
	mine, err := repo.ListMine(ctx, pastPassenger, "")
	if err != nil || len(mine) != 1 || mine[0].Status != "completed" {
		t.Fatalf("past booking did not complete on ListMine: err=%v items=%+v", err, mine)
	}
	future, err := repo.ListMine(ctx, futurePassenger, "")
	if err != nil || len(future) != 1 || future[0].Status != "confirmed" {
		t.Fatalf("future booking changed too early: err=%v items=%+v", err, future)
	}
	pending, err := repo.ListMine(ctx, pendingPassenger, "")
	if err != nil || len(pending) != 1 || pending[0].Status != "pending" {
		t.Fatalf("pending booking changed unexpectedly: err=%v items=%+v", err, pending)
	}
	received, err := repo.ListReceived(ctx, driverUser, "completed")
	if err != nil || len(received) != 1 || received[0].ID != pastBookingID {
		t.Fatalf("ListReceived did not expose completed booking: err=%v items=%+v", err, received)
	}
	detail, err := repo.GetVisible(ctx, pastBookingID, pastPassenger)
	if err != nil || detail.Status != "completed" {
		t.Fatalf("GetVisible did not preserve completed status: err=%v item=%+v", err, detail)
	}
}
