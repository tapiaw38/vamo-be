package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	ratingrepo "github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/rating"
	triprepo "github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/trip"
	ratinghandler "github.com/tapiaw38/vamo/backend/internal/adapters/web/handlers/ratings"
	triph "github.com/tapiaw38/vamo/backend/internal/adapters/web/handlers/trips"
	"github.com/tapiaw38/vamo/backend/internal/platform/auth"
	ratinguc "github.com/tapiaw38/vamo/backend/internal/usecases/ratings"
	tripuc "github.com/tapiaw38/vamo/backend/internal/usecases/trips"
)

func TestRatingsFlowAndTripStatsAgainstPostgres(t *testing.T) {
	db := phase6DB(t)
	defer db.Close()
	ctx := context.Background()
	const driverUser = "ratings-driver-user"
	const passengerOne = "ratings-passenger-one"
	const passengerTwo = "ratings-passenger-two"
	const outsider = "ratings-outsider"
	const driverID = "00000000-0000-0000-0000-000000000031"
	const vehicleID = "00000000-0000-0000-0000-000000000032"
	const tripID = "00000000-0000-0000-0000-000000000033"
	const bookingOne = "00000000-0000-0000-0000-000000000034"
	const bookingTwo = "00000000-0000-0000-0000-000000000035"
	const bookingPending = "00000000-0000-0000-0000-000000000036"
	cleanup := func() {
		_, _ = db.Exec(ctx, `DELETE FROM trip_ratings WHERE trip_id=$1`, tripID)
		_, _ = db.Exec(ctx, `DELETE FROM booking_events WHERE booking_id IN ($1,$2,$3)`, bookingOne, bookingTwo, bookingPending)
		_, _ = db.Exec(ctx, `DELETE FROM bookings WHERE id IN ($1,$2,$3)`, bookingOne, bookingTwo, bookingPending)
		_, _ = db.Exec(ctx, `DELETE FROM trips WHERE id=$1`, tripID)
		_, _ = db.Exec(ctx, `DELETE FROM vehicles WHERE id=$1`, vehicleID)
		_, _ = db.Exec(ctx, `DELETE FROM driver_profiles WHERE id=$1`, driverID)
		_, _ = db.Exec(ctx, `DELETE FROM users WHERE id IN ($1,$2,$3,$4)`, driverUser, passengerOne, passengerTwo, outsider)
	}
	cleanup()
	defer cleanup()

	var origin, destination string
	if err := db.QueryRow(ctx, `SELECT l1.id, l2.id FROM localities l1 CROSS JOIN localities l2 WHERE l1.id <> l2.id LIMIT 1`).Scan(&origin, &destination); err != nil {
		t.Fatal(err)
	}
	for _, userID := range []string{driverUser, passengerOne, passengerTwo, outsider} {
		if _, err := db.Exec(ctx, `INSERT INTO users(id) VALUES($1)`, userID); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := db.Exec(ctx, `INSERT INTO driver_profiles(id,user_id,status) VALUES($1,$2,'verified')`, driverID, driverUser); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(ctx, `INSERT INTO vehicles(id,owner_type,driver_profile_id,make,model,year,license_plate,authorized_seats,vehicle_type) VALUES($1,'driver',$2,'Ratings','Car',2024,'RATING31',4,'car')`, vehicleID, driverID); err != nil {
		t.Fatal(err)
	}
	departure := time.Now().Add(-2 * time.Hour).UTC()
	if _, err := db.Exec(ctx, `INSERT INTO trips(id,publisher_type,driver_profile_id,vehicle_id,origin_locality_id,destination_locality_id,departure_at,price_per_seat,total_seats,available_seats,status) VALUES($1,'driver',$2,$3,$4,$5,$6,1000,4,4,'full')`, tripID, driverID, vehicleID, origin, destination, departure); err != nil {
		t.Fatal(err)
	}
	for _, item := range []struct{ id, passenger, status string }{
		{bookingOne, passengerOne, "completed"},
		{bookingTwo, passengerTwo, "confirmed"},
		{bookingPending, passengerOne, "pending"},
	} {
		if _, err := db.Exec(ctx, `INSERT INTO bookings(id,trip_id,passenger_id,seats,unit_price_snapshot,total_price,status) VALUES($1,$2,$3,1,1000,1000,$4)`, item.id, tripID, item.passenger, item.status); err != nil {
			t.Fatal(err)
		}
	}

	ratingUsecases := ratinguc.Usecases{Repo: ratingrepo.NewRepository(db)}
	gin.SetMode(gin.TestMode)
	r := gin.New()
	protected := r.Group("")
	protected.Use(auth.Middleware("ratings-secret"))
	protected.POST("/bookings/:id/rating", ratinghandler.NewCreate(ratingUsecases))
	r.GET("/drivers/:id/rating", ratinghandler.NewDriverStats(ratingUsecases))

	postRating := func(actor, booking string, score int) *httptest.ResponseRecorder {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"user_id": actor, "exp": time.Now().Add(time.Hour).Unix()})
		signed, err := token.SignedString([]byte("ratings-secret"))
		if err != nil {
			t.Fatal(err)
		}
		body, _ := json.Marshal(map[string]any{"score": score})
		req := httptest.NewRequest(http.MethodPost, "/bookings/"+booking+"/rating", strings.NewReader(string(body)))
		req.Header.Set("Authorization", "Bearer "+signed)
		req.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		r.ServeHTTP(response, req)
		return response
	}

	if response := postRating(passengerOne, bookingOne, 5); response.Code != http.StatusCreated {
		t.Fatalf("first rating failed: status=%d body=%s", response.Code, response.Body.String())
	}
	if response := postRating(passengerOne, bookingOne, 4); response.Code != http.StatusConflict {
		t.Fatalf("duplicate rating status=%d body=%s", response.Code, response.Body.String())
	}
	if response := postRating(outsider, bookingOne, 5); response.Code != http.StatusForbidden {
		t.Fatalf("foreign rating status=%d body=%s", response.Code, response.Body.String())
	}
	if response := postRating(passengerOne, bookingPending, 5); response.Code != http.StatusBadRequest {
		t.Fatalf("pending rating status=%d body=%s", response.Code, response.Body.String())
	}
	if response := postRating(passengerTwo, bookingTwo, 4); response.Code != http.StatusCreated {
		t.Fatalf("second rating failed: status=%d body=%s", response.Code, response.Body.String())
	}

	statsReq := httptest.NewRequest(http.MethodGet, "/drivers/"+driverID+"/rating", nil)
	statsResponse := httptest.NewRecorder()
	r.ServeHTTP(statsResponse, statsReq)
	if statsResponse.Code != http.StatusOK || !strings.Contains(statsResponse.Body.String(), `"average_score":4.5`) || !strings.Contains(statsResponse.Body.String(), `"total_ratings":2`) {
		t.Fatalf("unexpected rating stats: status=%d body=%s", statsResponse.Code, statsResponse.Body.String())
	}

	tripRouter := gin.New()
	tripRouter.GET("/trips/:id", triph.NewGetHandler(tripuc.NewGetUsecase(triprepo.NewRepository(db))))
	tripRouter.GET("/trips", triph.NewSearchHandler(tripuc.NewSearchUsecase(triprepo.NewRepository(db))))
	for _, path := range []string{"/trips/" + tripID, "/trips?origin_locality_id=" + origin + "&destination_locality_id=" + destination} {
		request := httptest.NewRequest(http.MethodGet, path, nil)
		response := httptest.NewRecorder()
		tripRouter.ServeHTTP(response, request)
		if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"average_score":4.5`) || !strings.Contains(response.Body.String(), `"total_ratings":2`) {
			t.Fatalf("trip rating stats missing for %s: status=%d body=%s", path, response.Code, response.Body.String())
		}
	}
}
