package integration

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	driverrepo "github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/driver"
	pricingrepo "github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/pricing"
	riderepo "github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/riderequest"
	triprepo "github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/trip"
	userrepo "github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/user"
	pricinguc "github.com/tapiaw38/vamo/backend/internal/usecases/pricing"
	profilebootstrap "github.com/tapiaw38/vamo/backend/internal/usecases/profilebootstrap"
	rideuc "github.com/tapiaw38/vamo/backend/internal/usecases/riderequests"
	tripuc "github.com/tapiaw38/vamo/backend/internal/usecases/trips"
)

func TestRideRequestsRequireAvailableDriverAndAcceptOnce(t *testing.T) {
	db := phase6DB(t)
	defer db.Close()
	ctx := context.Background()
	const passenger = "ride-passenger-real"
	const driverOne = "ride-driver-one-real"
	const driverTwo = "ride-driver-two-real"
	const profileOne = "00000000-0000-0000-0000-000000000031"
	const profileTwo = "00000000-0000-0000-0000-000000000032"
	const vehicleOne = "00000000-0000-0000-0000-000000000033"
	const vehicleTwo = "00000000-0000-0000-0000-000000000034"
	defer func() {
		_, _ = db.Exec(ctx, `DELETE FROM booking_events WHERE booking_id IN (SELECT id FROM bookings WHERE passenger_id=$1)`, passenger)
		_, _ = db.Exec(ctx, `DELETE FROM bookings WHERE passenger_id=$1`, passenger)
		_, _ = db.Exec(ctx, `DELETE FROM ride_requests WHERE passenger_id=$1`, passenger)
		_, _ = db.Exec(ctx, `DELETE FROM trips WHERE driver_profile_id IN ($1,$2)`, profileOne, profileTwo)
		_, _ = db.Exec(ctx, `DELETE FROM vehicles WHERE id IN ($1,$2)`, vehicleOne, vehicleTwo)
		_, _ = db.Exec(ctx, `DELETE FROM driver_profiles WHERE id IN ($1,$2)`, profileOne, profileTwo)
		_, _ = db.Exec(ctx, `DELETE FROM users WHERE id IN ($1,$2,$3)`, passenger, driverOne, driverTwo)
	}()
	var origin, destination string
	if err := db.QueryRow(ctx, `SELECT l1.id,l2.id FROM localities l1 CROSS JOIN localities l2 WHERE l1.id<>l2.id ORDER BY l1.id,l2.id LIMIT 1`).Scan(&origin, &destination); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{passenger, driverOne, driverTwo} {
		if _, err := db.Exec(ctx, `INSERT INTO users(id) VALUES($1)`, id); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := db.Exec(ctx, `INSERT INTO driver_profiles(id,user_id,status,is_available) VALUES($1,$2,'draft',FALSE),($3,$4,'draft',TRUE)`, profileOne, driverOne, profileTwo, driverTwo); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(ctx, `INSERT INTO vehicles(id,owner_type,driver_profile_id,make,model,year,license_plate,authorized_seats,vehicle_type) VALUES($1,'driver',$2,'Ride','One',2025,'RIDEREQ31',4,'car'),($3,'driver',$4,'Ride','Two',2025,'RIDEREQ32',4,'car')`, vehicleOne, profileOne, vehicleTwo, profileTwo); err != nil {
		t.Fatal(err)
	}

	repo := riderepo.NewRepository(db)
	pricing := pricinguc.NewUsecases(pricingrepo.NewRepository(db))
	uc := rideuc.Usecases{Repo: repo, Drivers: driverrepo.NewRepository(db), Pricing: pricing}
	requested := time.Now().Add(2 * time.Hour).UTC().Format(time.RFC3339)
	created, appErr := uc.Create(ctx, passenger, rideuc.Input{OriginLocalityID: origin, DestinationLocalityID: destination, RequestedAt: requested, MaxPricePerSeat: "1800", Seats: 1})
	if appErr != nil {
		t.Fatalf("create ride request: %s", appErr.Code)
	}
	if _, appErr = uc.List(ctx, driverOne, rideuc.SearchInput{OriginLocalityID: origin}); appErr == nil || appErr.Code != "ride-requests:driver-unavailable" {
		t.Fatalf("expected unavailable driver rejection, got %v", appErr)
	}

	// Make both available, then race two accepts against one locked request.
	if _, err := db.Exec(ctx, `UPDATE driver_profiles SET is_available=TRUE WHERE id IN ($1,$2)`, profileOne, profileTwo); err != nil {
		t.Fatal(err)
	}
	if _, appErr = uc.Accept(ctx, driverOne, created.ID, vehicleTwo); appErr == nil {
		t.Fatal("expected vehicle owned by another driver to be rejected")
	}
	var wg sync.WaitGroup
	results := make(chan *rideuc.AcceptData, 2)
	errors := make(chan *string, 2)
	for _, user := range []string{driverOne, driverTwo} {
		wg.Add(1)
		go func(actor string) {
			defer wg.Done()
			result, e := uc.Accept(ctx, actor, created.ID)
			if e != nil {
				code := e.Code
				errors <- &code
				return
			}
			results <- &result
		}(user)
	}
	wg.Wait()
	close(results)
	close(errors)
	if len(results) != 1 || len(errors) != 1 {
		codes := []string{}
		for code := range errors {
			codes = append(codes, *code)
		}
		t.Fatalf("expected one acceptance and one conflict, accepted=%d errors=%d codes=%v", len(results), len(errors), codes)
	}
	var status string
	var tripID, bookingID string
	if err := db.QueryRow(ctx, `SELECT status,accepted_trip_id FROM ride_requests WHERE id=$1`, created.ID).Scan(&status, &tripID); err != nil {
		t.Fatal(err)
	}
	if status != "accepted" || tripID == "" {
		t.Fatalf("unexpected request state: %s %s", status, tripID)
	}
	if err := db.QueryRow(ctx, `SELECT id FROM bookings WHERE trip_id=$1 AND passenger_id=$2`, tripID, passenger).Scan(&bookingID); err != nil {
		t.Fatal(err)
	}
	var bookingStatus string
	if err := db.QueryRow(ctx, `SELECT status FROM bookings WHERE id=$1`, bookingID).Scan(&bookingStatus); err != nil {
		t.Fatal(err)
	}
	if bookingStatus != "confirmed" {
		t.Fatalf("expected confirmed booking, got %s", bookingStatus)
	}
	active, activeErr := repo.GetActive(ctx, passenger)
	if activeErr != nil {
		t.Fatalf("get active ride request: %v", activeErr)
	}
	if active == nil || active.Status != "accepted" || active.AcceptedTripID != tripID {
		t.Fatalf("unexpected active request: %#v", active)
	}
	activeData, activeAppErr := uc.Active(ctx, passenger)
	if activeAppErr != nil || activeData == nil || !strings.Contains(activeData.AcceptedVehicleLicensePlate, "*") {
		t.Fatalf("expected accepted active request with masked plate, data=%#v err=%v", activeData, activeAppErr)
	}
	var tripPrice string
	var totalSeats, availableSeats int
	if err := db.QueryRow(ctx, `SELECT price_per_seat::text,total_seats,available_seats FROM trips WHERE id=$1`, tripID).Scan(&tripPrice, &totalSeats, &availableSeats); err != nil {
		t.Fatal(err)
	}
	if totalSeats != 4 || availableSeats != 3 || tripPrice == "1800" {
		t.Fatalf("ride acceptance did not use calculated price/capacity: price=%s total=%d available=%d", tripPrice, totalSeats, availableSeats)
	}
}

func TestRideRequestBootstrapsMissingPassengerProfileAgainstPostgres(t *testing.T) {
	db := phase6DB(t)
	defer db.Close()
	ctx := context.Background()
	const passenger = "ride-passenger-lazy-profile-real"
	defer func() {
		_, _ = db.Exec(ctx, `DELETE FROM ride_requests WHERE passenger_id=$1`, passenger)
		_, _ = db.Exec(ctx, `DELETE FROM users WHERE id=$1`, passenger)
	}()

	var origin, destination string
	if err := db.QueryRow(ctx, `SELECT l1.id,l2.id FROM localities l1 CROSS JOIN localities l2 WHERE l1.id<>l2.id AND l1.lat IS NOT NULL AND l2.lat IS NOT NULL ORDER BY l1.id,l2.id LIMIT 1`).Scan(&origin, &destination); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE id=$1`, passenger).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("test passenger unexpectedly exists before lazy bootstrap")
	}

	uc := rideuc.Usecases{
		Repo:    riderepo.NewRepository(db),
		Pricing: pricinguc.NewUsecases(pricingrepo.NewRepository(db)),
		Profile: profilebootstrap.Service{Profile: userrepo.NewRepository(db)},
	}
	created, appErr := uc.Create(ctx, passenger, rideuc.Input{
		OriginLocalityID:      origin,
		DestinationLocalityID: destination,
		RequestedAt:           time.Now().Add(time.Hour).UTC().Format(time.RFC3339),
	})
	if appErr != nil {
		t.Fatalf("create ride request with lazy profile: %s", appErr.Code)
	}
	if created.PassengerID != passenger {
		t.Fatalf("unexpected passenger id: %s", created.PassengerID)
	}
	if created.Seats != 1 {
		t.Fatalf("expected default seats=1, got %d", created.Seats)
	}
	if err := db.QueryRow(ctx, `SELECT id FROM users WHERE id=$1`, passenger).Scan(new(string)); err != nil {
		t.Fatalf("lazy profile was not persisted: %v", err)
	}
}

func TestRideRequestGraceActiveUniquenessAndCancellationAgainstPostgres(t *testing.T) {
	db := phase6DB(t)
	defer db.Close()
	ctx := context.Background()
	const passenger = "ride-request-active-real"
	defer func() {
		_, _ = db.Exec(ctx, `DELETE FROM ride_requests WHERE passenger_id=$1`, passenger)
		_, _ = db.Exec(ctx, `DELETE FROM users WHERE id=$1`, passenger)
	}()
	var origin, destination string
	if err := db.QueryRow(ctx, `SELECT l1.id,l2.id FROM localities l1 JOIN localities l2 ON l1.id<>l2.id WHERE l1.lat IS NOT NULL AND l2.lat IS NOT NULL ORDER BY l1.id,l2.id LIMIT 1`).Scan(&origin, &destination); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(ctx, `INSERT INTO users(id) VALUES($1)`, passenger); err != nil {
		t.Fatal(err)
	}
	repo := riderepo.NewRepository(db)
	pricing := pricinguc.NewUsecases(pricingrepo.NewRepository(db))
	uc := rideuc.Usecases{Repo: repo, Pricing: pricing}
	first, appErr := uc.Create(ctx, passenger, rideuc.Input{OriginLocalityID: origin, DestinationLocalityID: destination, RequestedAt: time.Now().Add(time.Hour).UTC().Format(time.RFC3339)})
	if appErr != nil {
		t.Fatal(appErr)
	}
	active, err := repo.GetActive(ctx, passenger)
	if err != nil || active == nil || active.Status != "open" {
		t.Fatalf("immediate request was expired or missing: active=%#v err=%v", active, err)
	}
	if _, appErr = uc.Create(ctx, passenger, rideuc.Input{OriginLocalityID: origin, DestinationLocalityID: destination, RequestedAt: time.Now().Add(time.Hour).UTC().Format(time.RFC3339)}); appErr == nil || appErr.Status != 409 || appErr.Data == nil {
		t.Fatalf("expected active conflict with id, got %#v", appErr)
	}
	if _, err = db.Exec(ctx, `UPDATE ride_requests SET requested_at=NOW()-INTERVAL '16 minutes' WHERE id=$1`, first.ID); err != nil {
		t.Fatal(err)
	}
	if _, err = repo.ListOpen(ctx, origin, ""); err != nil {
		t.Fatal(err)
	}
	var status string
	if err = db.QueryRow(ctx, `SELECT status FROM ride_requests WHERE id=$1`, first.ID).Scan(&status); err != nil || status != "expired" {
		t.Fatalf("expected expired after grace window, status=%s err=%v", status, err)
	}
	if _, appErr = uc.Create(ctx, passenger, rideuc.Input{OriginLocalityID: origin, DestinationLocalityID: destination, RequestedAt: time.Now().Add(time.Hour).UTC().Format(time.RFC3339)}); appErr != nil {
		t.Fatalf("create after expiration: %v", appErr)
	}
	if active, err = repo.GetActive(ctx, "passenger-without-active"); err != nil || active != nil {
		t.Fatalf("expected no active request, active=%#v err=%v", active, err)
	}
}

func TestRecentRideRequestDestinationsAgainstPostgres(t *testing.T) {
	db := phase6DB(t)
	defer db.Close()
	ctx := context.Background()
	const passenger = "ride-request-recent-real"
	defer func() {
		_, _ = db.Exec(ctx, `DELETE FROM ride_requests WHERE passenger_id=$1`, passenger)
		_, _ = db.Exec(ctx, `DELETE FROM users WHERE id=$1`, passenger)
	}()
	var origin, destination string
	if err := db.QueryRow(ctx, `SELECT l1.id,l2.id FROM localities l1 JOIN localities l2 ON l1.id<>l2.id LIMIT 1`).Scan(&origin, &destination); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(ctx, `INSERT INTO users(id) VALUES($1)`, passenger); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 7; i++ {
		address := fmt.Sprintf("Destino %d", i)
		if _, err := db.Exec(ctx, `INSERT INTO ride_requests(passenger_id,origin_locality_id,destination_locality_id,requested_at,max_price_per_seat,seats,status,destination_address) VALUES($1,$2,$3,NOW()+INTERVAL '1 hour',1000,1,'cancelled',$4)`, passenger, origin, destination, address); err != nil {
			t.Fatal(err)
		}
	}
	items, err := riderepo.NewRepository(db).RecentDestinations(ctx, passenger)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 5 || items[0].DestinationAddress != "Destino 6" {
		t.Fatalf("expected five recent distinct destinations, got %+v", items)
	}
}

func TestRideRequestPriceIgnoresClientPriceAgainstPostgres(t *testing.T) {
	db := phase6DB(t)
	defer db.Close()
	ctx := context.Background()
	const passenger = "ride-pricing-passenger-real"
	defer func() {
		_, _ = db.Exec(ctx, `DELETE FROM ride_requests WHERE passenger_id=$1`, passenger)
		_, _ = db.Exec(ctx, `DELETE FROM users WHERE id=$1`, passenger)
	}()
	var origin, destination string
	if err := db.QueryRow(ctx, `SELECT l1.id,l2.id FROM localities l1 JOIN localities l2 ON l1.id<>l2.id WHERE l1.lat IS NOT NULL AND l1.lon IS NOT NULL AND l2.lat IS NOT NULL AND l2.lon IS NOT NULL ORDER BY l1.id,l2.id LIMIT 1`).Scan(&origin, &destination); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(ctx, `INSERT INTO users(id) VALUES($1)`, passenger); err != nil {
		t.Fatal(err)
	}
	pricing := pricinguc.NewUsecases(pricingrepo.NewRepository(db))
	quote, quoteErr := pricing.Quote(ctx, origin, destination)
	if quoteErr != nil {
		t.Fatal(quoteErr)
	}
	if quote.FareKind != "intercity" || quote.DistanceKM <= 0 {
		t.Fatalf("expected intercity quote with positive distance, got %+v", quote)
	}
	created, appErr := (rideuc.Usecases{Repo: riderepo.NewRepository(db), Drivers: driverrepo.NewRepository(db), Pricing: pricing}).Create(ctx, passenger, rideuc.Input{OriginLocalityID: origin, DestinationLocalityID: destination, RequestedAt: time.Now().Add(time.Hour).UTC().Format(time.RFC3339), MaxPricePerSeat: "1.00", Seats: 1})
	if appErr != nil {
		t.Fatal(appErr)
	}
	if created.MaxPricePerSeat != quote.PricePerSeat {
		t.Fatalf("expected calculated price %s, got %s", quote.PricePerSeat, created.MaxPricePerSeat)
	}
}

func TestRideRequestNowEmptyAndDefaultSeatsAgainstPostgres(t *testing.T) {
	db := phase6DB(t)
	defer db.Close()
	ctx := context.Background()
	const passenger = "ride-request-time-defaults-real"
	defer func() {
		_, _ = db.Exec(ctx, `DELETE FROM ride_requests WHERE passenger_id=$1`, passenger)
		_, _ = db.Exec(ctx, `DELETE FROM users WHERE id=$1`, passenger)
	}()
	var origin, destination string
	if err := db.QueryRow(ctx, `SELECT l1.id,l2.id FROM localities l1 JOIN localities l2 ON l1.id<>l2.id WHERE l1.lat IS NOT NULL AND l2.lat IS NOT NULL ORDER BY l1.id,l2.id LIMIT 1`).Scan(&origin, &destination); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(ctx, `INSERT INTO users(id) VALUES($1)`, passenger); err != nil {
		t.Fatal(err)
	}
	pricing := pricinguc.NewUsecases(pricingrepo.NewRepository(db))
	uc := rideuc.Usecases{Repo: riderepo.NewRepository(db), Drivers: driverrepo.NewRepository(db), Pricing: pricing}
	nowCreated, appErr := uc.Create(ctx, passenger, rideuc.Input{OriginLocalityID: origin, DestinationLocalityID: destination, RequestedAt: time.Now().UTC().Format(time.RFC3339)})
	if appErr != nil {
		t.Fatalf("now request should be accepted: %s %s", appErr.Code, appErr.Message)
	}
	if nowCreated.Seats != 1 {
		t.Fatalf("expected default seats=1, got %d", nowCreated.Seats)
	}
	if _, err := db.Exec(ctx, `UPDATE ride_requests SET status='cancelled' WHERE id=$1`, nowCreated.ID); err != nil {
		t.Fatal(err)
	}
	before := time.Now()
	emptyCreated, appErr := uc.Create(ctx, passenger, rideuc.Input{OriginLocalityID: origin, DestinationLocalityID: destination})
	if appErr != nil {
		t.Fatalf("empty requested_at should be accepted: %s %s", appErr.Code, appErr.Message)
	}
	if emptyCreated.RequestedAt.Before(before.Add(-time.Second)) || emptyCreated.RequestedAt.After(time.Now().Add(time.Second)) {
		t.Fatalf("empty requested_at was not server time: %s", emptyCreated.RequestedAt)
	}
	if _, appErr = uc.Create(ctx, passenger, rideuc.Input{OriginLocalityID: origin, DestinationLocalityID: destination, RequestedAt: time.Now().Add(-time.Hour).UTC().Format(time.RFC3339)}); appErr == nil || appErr.Message != "la fecha de salida es inválida o quedó demasiado en el pasado" {
		t.Fatalf("expected old date message, got %#v", appErr)
	}
	if _, appErr = uc.Create(ctx, passenger, rideuc.Input{DestinationLocalityID: destination}); appErr == nil || appErr.Message != "falta el origen" {
		t.Fatalf("expected missing origin message, got %#v", appErr)
	}
	if _, appErr = uc.Create(ctx, passenger, rideuc.Input{OriginLocalityID: origin}); appErr == nil || appErr.Message != "falta el destino" {
		t.Fatalf("expected missing destination message, got %#v", appErr)
	}
}

func TestRideRequestUrbanPinsAndAddressesAgainstPostgres(t *testing.T) {
	db := phase6DB(t)
	defer db.Close()
	ctx := context.Background()
	const passenger = "ride-urban-passenger-real"
	defer func() {
		_, _ = db.Exec(ctx, `DELETE FROM ride_requests WHERE passenger_id=$1`, passenger)
		_, _ = db.Exec(ctx, `DELETE FROM users WHERE id=$1`, passenger)
	}()
	var locality string
	var centerLat, centerLon float64
	if err := db.QueryRow(ctx, `SELECT id,lat,lon FROM localities WHERE lat IS NOT NULL AND lon IS NOT NULL ORDER BY id LIMIT 1`).Scan(&locality, &centerLat, &centerLon); err != nil {
		t.Fatal(err)
	}
	pricing := pricinguc.NewUsecases(pricingrepo.NewRepository(db))
	quote, appErr := pricing.Quote(ctx, locality, locality, centerLat, centerLon, centerLat+0.01, centerLon+0.01)
	if appErr != nil {
		t.Fatal(appErr)
	}
	if quote.FareKind != "urban" || quote.DistanceKM <= 0 {
		t.Fatalf("unexpected urban quote: %+v", quote)
	}
	if _, err := db.Exec(ctx, `INSERT INTO users(id) VALUES($1)`, passenger); err != nil {
		t.Fatal(err)
	}
	created, createErr := (rideuc.Usecases{Repo: riderepo.NewRepository(db), Drivers: driverrepo.NewRepository(db), Pricing: pricing}).Create(ctx, passenger, rideuc.Input{OriginLocalityID: locality, DestinationLocalityID: locality, OriginLat: &centerLat, OriginLon: &centerLon, DestinationLat: ptrFloat(centerLat + 0.01), DestinationLon: ptrFloat(centerLon + 0.01), OriginAddress: "Plaza principal", DestinationAddress: "Calle 25 de Mayo 100", RequestedAt: time.Now().Add(time.Hour).UTC().Format(time.RFC3339), MaxPricePerSeat: "1", Seats: 1})
	if createErr != nil {
		t.Fatal(createErr)
	}
	var storedPrice, originAddress, destinationAddress string
	var storedOriginLat, storedDestinationLat float64
	if err := db.QueryRow(ctx, `SELECT max_price_per_seat::text,origin_lat,origin_address,destination_lat,destination_address FROM ride_requests WHERE id=$1`, created.ID).Scan(&storedPrice, &storedOriginLat, &originAddress, &storedDestinationLat, &destinationAddress); err != nil {
		t.Fatal(err)
	}
	if storedPrice != quote.PricePerSeat || storedOriginLat != centerLat || storedDestinationLat != centerLat+0.01 || originAddress != "Plaza principal" || destinationAddress != "Calle 25 de Mayo 100" {
		t.Fatalf("urban ride request data mismatch: price=%s origin=%v/%s destination=%v/%s", storedPrice, storedOriginLat, originAddress, storedDestinationLat, destinationAddress)
	}
}

func ptrFloat(value float64) *float64 { return &value }

func TestRideRequestPassengerCannotCancelForeignRequest(t *testing.T) {
	db := phase6DB(t)
	defer db.Close()
	ctx := context.Background()
	const owner = "ride-cancel-owner-real"
	const outsider = "ride-cancel-outsider-real"
	defer func() {
		_, _ = db.Exec(ctx, `DELETE FROM ride_requests WHERE passenger_id IN ($1,$2)`, owner, outsider)
		_, _ = db.Exec(ctx, `DELETE FROM users WHERE id IN ($1,$2)`, owner, outsider)
	}()
	var origin, destination string
	if err := db.QueryRow(ctx, `SELECT id FROM localities ORDER BY id LIMIT 1`).Scan(&origin); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(ctx, `SELECT id FROM localities WHERE id<>$1 ORDER BY id LIMIT 1`, origin).Scan(&destination); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{owner, outsider} {
		if _, err := db.Exec(ctx, `INSERT INTO users(id) VALUES($1)`, id); err != nil {
			t.Fatal(err)
		}
	}
	pricing := pricinguc.NewUsecases(pricingrepo.NewRepository(db))
	id, appErr := (rideuc.Usecases{Repo: riderepo.NewRepository(db), Drivers: driverrepo.NewRepository(db), Pricing: pricing}).Create(ctx, owner, rideuc.Input{OriginLocalityID: origin, DestinationLocalityID: destination, RequestedAt: time.Now().Add(time.Hour).UTC().Format(time.RFC3339), MaxPricePerSeat: "1000", Seats: 1})
	if appErr != nil {
		t.Fatal(appErr)
	}
	_, appErr = (rideuc.Usecases{Repo: riderepo.NewRepository(db), Drivers: driverrepo.NewRepository(db), Pricing: pricing}).Cancel(ctx, outsider, id.ID)
	if appErr == nil || appErr.Code != "ride-requests:forbidden" {
		t.Fatalf("expected foreign cancellation rejection, got %v", appErr)
	}
}

func TestPublicTripMasksVehiclePlate(t *testing.T) {
	db := phase6DB(t)
	defer db.Close()
	ctx := context.Background()
	const user = "ride-mask-user"
	const profile = "00000000-0000-0000-0000-000000000035"
	const vehicle = "00000000-0000-0000-0000-000000000036"
	defer func() {
		_, _ = db.Exec(ctx, `DELETE FROM trips WHERE driver_profile_id=$1`, profile)
		_, _ = db.Exec(ctx, `DELETE FROM vehicles WHERE id=$1`, vehicle)
		_, _ = db.Exec(ctx, `DELETE FROM driver_profiles WHERE id=$1`, profile)
		_, _ = db.Exec(ctx, `DELETE FROM users WHERE id=$1`, user)
	}()
	var origin, destination string
	if err := db.QueryRow(ctx, `SELECT id FROM localities ORDER BY id LIMIT 1`).Scan(&origin); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(ctx, `SELECT id FROM localities WHERE id<>$1 ORDER BY id LIMIT 1`, origin).Scan(&destination); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(ctx, `INSERT INTO users(id) VALUES($1)`, user); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(ctx, `INSERT INTO driver_profiles(id,user_id,status) VALUES($1,$2,'draft')`, profile, user); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(ctx, `INSERT INTO vehicles(id,owner_type,driver_profile_id,make,model,year,license_plate,authorized_seats,vehicle_type) VALUES($1,'driver',$2,'Mask','Car',2025,'AB123CD',4,'car')`, vehicle, profile); err != nil {
		t.Fatal(err)
	}
	var tripID string
	if err := db.QueryRow(ctx, `INSERT INTO trips(publisher_type,driver_profile_id,vehicle_id,origin_locality_id,destination_locality_id,departure_at,price_per_seat,total_seats,available_seats,status) VALUES('driver',$1,$2,$3,$4,NOW()+INTERVAL '2 hours',1000,4,4,'published') RETURNING id`, profile, vehicle, origin, destination).Scan(&tripID); err != nil {
		t.Fatal(err)
	}
	output, appErr := tripuc.NewGetUsecase(triprepo.NewRepository(db)).Execute(ctx, tripID)
	if appErr != nil {
		t.Fatal(appErr)
	}
	if output.VehicleLicensePlate == "AB123CD" || output.VehicleLicensePlate != "AB1****" {
		t.Fatalf("unexpected masked plate: %q", output.VehicleLicensePlate)
	}
}
