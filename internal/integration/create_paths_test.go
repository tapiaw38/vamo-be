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
	companyrepo "github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/company"
	driverrepo "github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/driver"
	triprepo "github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/trip"
	vehiclerepo "github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/vehicle"
	triphandler "github.com/tapiaw38/vamo/backend/internal/adapters/web/handlers/trips"
	"github.com/tapiaw38/vamo/backend/internal/platform/auth"
	companyuc "github.com/tapiaw38/vamo/backend/internal/usecases/companies"
	tripuc "github.com/tapiaw38/vamo/backend/internal/usecases/trips"
)

func TestTripCreateParticularAgainstPostgres(t *testing.T) {
	db := phase6DB(t)
	defer db.Close()
	ctx := context.Background()
	const userID = "sql-trip-create-user"
	const driverID = "00000000-0000-0000-0000-000000000021"
	const vehicleID = "00000000-0000-0000-0000-000000000022"
	defer func() {
		_, _ = db.Exec(ctx, `DELETE FROM trips WHERE driver_profile_id=$1`, driverID)
		_, _ = db.Exec(ctx, `DELETE FROM vehicles WHERE id=$1`, vehicleID)
		_, _ = db.Exec(ctx, `DELETE FROM driver_profiles WHERE id=$1`, driverID)
		_, _ = db.Exec(ctx, `DELETE FROM users WHERE id=$1`, userID)
	}()

	var origin, destination string
	if err := db.QueryRow(ctx, `SELECT l1.id, l2.id FROM localities l1 JOIN provinces p1 ON p1.id=l1.province_id CROSS JOIN localities l2 JOIN provinces p2 ON p2.id=l2.province_id WHERE p1.id <> p2.id LIMIT 1`).Scan(&origin, &destination); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(ctx, `INSERT INTO users(id) VALUES($1)`, userID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(ctx, `INSERT INTO driver_profiles(id,user_id,status) VALUES($1,$2,'draft')`, driverID, userID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(ctx, `INSERT INTO vehicles(id,owner_type,driver_profile_id,make,model,year,license_plate,authorized_seats,vehicle_type) VALUES($1,'driver',$2,'Integration','Trip',2025,'SQLTRIP21',4,'car')`, vehicleID, driverID); err != nil {
		t.Fatal(err)
	}

	input := tripuc.Input{
		VehicleID: vehicleID, OriginLocalityID: origin, DestinationLocalityID: destination,
		DepartureAt: time.Now().Add(24 * time.Hour).UTC().Format(time.RFC3339), PricePerSeat: "1500",
		TotalSeats: 2, BookingMode: "automatic",
	}
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(auth.Middleware("integration-secret"))
	r.POST("/trips", auth.RequireConductor(), triphandler.NewCreateHandler(tripuc.NewCreateUsecase(driverrepo.NewRepository(db), vehiclerepo.NewRepository(db), triprepo.NewRepository(db))))
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"roles":   []map[string]any{{"name": "user"}, {"name": "admin"}},
		"exp":     time.Now().Add(time.Hour).Unix(),
	})
	signed, err := token.SignedString([]byte("integration-secret"))
	if err != nil {
		t.Fatal(err)
	}
	body, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/trips", strings.NewReader(string(body)))
	req.Header.Set("Authorization", "Bearer "+signed)
	req.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	r.ServeHTTP(response, req)
	if response.Code != http.StatusCreated {
		t.Fatalf("trip POST failed: status=%d body=%s", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), `"origin_locality_id":"`+origin+`"`) || !strings.Contains(response.Body.String(), `"destination_locality_id":"`+destination+`"`) {
		t.Fatalf("unexpected created trip response: %s", response.Body.String())
	}
}

func TestCompanyListMineAfterCreateAgainstPostgres(t *testing.T) {
	db := phase6DB(t)
	defer db.Close()
	ctx := context.Background()
	const userID = "sql-company-list-user"
	const taxID = "SQL-LIST-TEST-22"
	defer func() {
		_, _ = db.Exec(ctx, `DELETE FROM audit_logs WHERE actor_user_id=$1`, userID)
		_, _ = db.Exec(ctx, `DELETE FROM company_members WHERE user_id=$1`, userID)
		_, _ = db.Exec(ctx, `DELETE FROM companies WHERE tax_id=$1`, taxID)
		_, _ = db.Exec(ctx, `DELETE FROM users WHERE id=$1`, userID)
	}()
	var provinceID, localityID string
	if err := db.QueryRow(ctx, `SELECT p.id, l.id FROM provinces p JOIN localities l ON l.province_id=p.id ORDER BY p.id, l.id LIMIT 1`).Scan(&provinceID, &localityID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(ctx, `INSERT INTO users(id) VALUES($1)`, userID); err != nil {
		t.Fatal(err)
	}
	repo := companyrepo.NewRepository(db)
	created, appErr := (companyuc.Usecases{Companies: repo}).Create(ctx, userID, companyuc.CompanyInput{
		LegalName: "SQL List Test SA", DisplayName: "SQL List Test", TaxID: taxID,
		ProvinceID: provinceID, LocalityID: localityID,
	})
	if appErr != nil {
		t.Fatalf("company create failed: %s", appErr.Code)
	}
	items, appErr := (companyuc.Usecases{Companies: repo}).List(ctx, userID)
	if appErr != nil {
		t.Fatalf("company list failed: %s", appErr.Code)
	}
	for _, item := range items {
		if item.ID == created.ID && item.TaxID == taxID {
			return
		}
	}
	t.Fatalf("created company not found in list: %+v", items)
}
