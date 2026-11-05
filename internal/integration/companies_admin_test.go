package integration

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	adminrepo "github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/admin"
	companyrepo "github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/company"
	driverrepo "github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/driver"
	triprepo "github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/trip"
	vehiclerepo "github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/vehicle"
	companyh "github.com/tapiaw38/vamo/backend/internal/adapters/web/handlers/companies"
	"github.com/tapiaw38/vamo/backend/internal/domain"
	"github.com/tapiaw38/vamo/backend/internal/platform/auth"
	companyuc "github.com/tapiaw38/vamo/backend/internal/usecases/companies"
	tripuc "github.com/tapiaw38/vamo/backend/internal/usecases/trips"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

type phase6Fixture struct {
	db                                                             *pgxpool.Pool
	company, owner, outsider, driver, driverProfile, vehicle, trip string
	province, locality, destination                                string
}

func phase6DB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("VAMO_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("set VAMO_TEST_DATABASE_URL after migrations")
	}
	db, e := pgxpool.New(context.Background(), url)
	if e != nil {
		t.Fatal(e)
	}
	return db
}
func newPhase6Fixture(t *testing.T) phase6Fixture {
	db := phase6DB(t)
	f := phase6Fixture{db: db, company: "00000000-0000-0000-0000-000000000009", owner: "phase6-owner", outsider: "phase6-outsider", driver: "phase6-driver", driverProfile: "00000000-0000-0000-0000-000000000010", vehicle: "00000000-0000-0000-0000-000000000011", trip: "00000000-0000-0000-0000-000000000012"}
	cleanupPhase6(t, f)
	if e := db.QueryRow(context.Background(), `SELECT id FROM provinces ORDER BY id LIMIT 1`).Scan(&f.province); e != nil {
		t.Fatal(e)
	}
	if e := db.QueryRow(context.Background(), `SELECT id FROM localities WHERE province_id=$1 ORDER BY id LIMIT 1`, f.province).Scan(&f.locality); e != nil {
		t.Fatal(e)
	}
	if e := db.QueryRow(context.Background(), `SELECT id FROM localities WHERE id<>$1 ORDER BY id LIMIT 1`, f.locality).Scan(&f.destination); e != nil {
		t.Fatal(e)
	}
	for _, u := range []string{f.owner, f.outsider, f.driver} {
		if _, e := db.Exec(context.Background(), `INSERT INTO users(id) VALUES($1)`, u); e != nil {
			t.Fatal(e)
		}
	}
	if _, e := db.Exec(context.Background(), `INSERT INTO driver_profiles(id,user_id,status) VALUES($1,$2,'draft')`, f.driverProfile, f.driver); e != nil {
		t.Fatal(e)
	}
	if _, e := db.Exec(context.Background(), `INSERT INTO companies(id,legal_name,display_name,tax_id,province_id,locality_id) VALUES($1,'Fase 6 SA','Fase 6',$2,$3,$4)`, f.company, "30-99999999-1", f.province, f.locality); e != nil {
		t.Fatal(e)
	}
	if _, e := db.Exec(context.Background(), `INSERT INTO company_members(company_id,user_id,role) VALUES($1,$2,'owner')`, f.company, f.owner); e != nil {
		t.Fatal(e)
	}
	if _, e := db.Exec(context.Background(), `INSERT INTO vehicles(id,owner_type,company_id,make,model,year,license_plate,authorized_seats,vehicle_type) VALUES($1,'company',$2,'Test','Company',2024,'F6C0001',4,'van')`, f.vehicle, f.company); e != nil {
		t.Fatal(e)
	}
	return f
}
func cleanupPhase6(t *testing.T, f phase6Fixture) {
	if f.db == nil {
		return
	}
	ctx := context.Background()
	_, _ = f.db.Exec(ctx, `DELETE FROM audit_logs WHERE entity_id IN ($1,$2,$3) OR actor_user_id IN ($4,$5,$6)`, f.company, f.driverProfile, f.vehicle, f.owner, f.outsider, f.driver)
	_, _ = f.db.Exec(ctx, `DELETE FROM bookings WHERE trip_id=$1`, f.trip)
	_, _ = f.db.Exec(ctx, `DELETE FROM trips WHERE id=$1`, f.trip)
	_, _ = f.db.Exec(ctx, `DELETE FROM vehicles WHERE id=$1`, f.vehicle)
	_, _ = f.db.Exec(ctx, `DELETE FROM company_members WHERE company_id=$1`, f.company)
	_, _ = f.db.Exec(ctx, `DELETE FROM companies WHERE id=$1`, f.company)
	_, _ = f.db.Exec(ctx, `DELETE FROM driver_profiles WHERE id=$1`, f.driverProfile)
	_, _ = f.db.Exec(ctx, `DELETE FROM user_roles WHERE user_id IN ($1,$2,$3)`, f.owner, f.outsider, f.driver)
	_, _ = f.db.Exec(ctx, `DELETE FROM users WHERE id IN ($1,$2,$3)`, f.owner, f.outsider, f.driver)
}

func TestCompanyAdminCannotAccessForeignCompany(t *testing.T) {
	f := newPhase6Fixture(t)
	defer func() { cleanupPhase6(t, f); f.db.Close() }()
	_, e := companyrepo.NewRepository(f.db).Get(context.Background(), f.company, f.outsider)
	if e != companyrepo.ErrForbidden {
		t.Fatalf("expected foreign company forbidden, got %v", e)
	}
}
func TestCompanyPatchRequiresOwnerOrAdmin(t *testing.T) {
	f := newPhase6Fixture(t)
	defer func() { cleanupPhase6(t, f); f.db.Close() }()
	repo := companyrepo.NewRepository(f.db)
	_, err := repo.Update(context.Background(), domain.Company{ID: f.company, LegalName: "Changed", DisplayName: "Changed", TaxID: "30-99999999-1", ProvinceID: f.province, LocalityID: f.locality}, f.outsider)
	if err != companyrepo.ErrForbidden {
		t.Fatalf("expected foreign PATCH forbidden, got %v", err)
	}
	if _, err = repo.Update(context.Background(), domain.Company{ID: f.company, LegalName: "Changed", DisplayName: "Changed", TaxID: "30-99999999-1", ProvinceID: f.province, LocalityID: f.locality}, f.owner); err != nil {
		t.Fatalf("owner PATCH failed: %v", err)
	}
}
func TestCompanyVehiclesListHandlesNullableOwnersAgainstPostgres(t *testing.T) {
	f := newPhase6Fixture(t)
	defer func() { cleanupPhase6(t, f); f.db.Close() }()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) { c.Set("vamo.user_id", f.owner); c.Next() })
	r.GET("/companies/:id/vehicles", companyh.NewVehicles(companyuc.Usecases{Companies: companyrepo.NewRepository(f.db), VehicleRepo: vehiclerepo.NewRepository(f.db)}))
	req := httptest.NewRequest(http.MethodGet, "/companies/"+f.company+"/vehicles", nil)
	response := httptest.NewRecorder()
	r.ServeHTTP(response, req)
	if response.Code != http.StatusOK {
		t.Fatalf("company vehicle GET failed: status=%d body=%s", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), "F6C0001") {
		t.Fatalf("company vehicle missing from response: %s", response.Body.String())
	}
}
func TestCompanyTripRejectsForeignDriver(t *testing.T) {
	f := newPhase6Fixture(t)
	defer func() { cleanupPhase6(t, f); f.db.Close() }()
	repo := companyrepo.NewRepository(f.db)
	vehicles := vehiclerepo.NewRepository(f.db)
	drivers := driverrepo.NewRepository(f.db)
	trips := triprepo.NewRepository(f.db)
	_, e := tripuc.NewCreateUsecase(drivers, vehicles, trips, repo).Execute(context.Background(), f.owner, tripuc.Input{PublisherType: "company", CompanyID: f.company, DriverProfileID: f.driverProfile, VehicleID: f.vehicle, OriginLocalityID: f.locality, DestinationLocalityID: f.destination, DepartureAt: time.Now().Add(24 * time.Hour).UTC().Format(time.RFC3339), PricePerSeat: "1000", TotalSeats: 2, BookingMode: "automatic"})
	if e == nil || e.Code != "companies:driver-forbidden" {
		t.Fatalf("expected foreign driver rejection, got %v", e)
	}
}
func TestCompanyTripRejectsVehicleOutsideCompany(t *testing.T) {
	f := newPhase6Fixture(t)
	defer func() { cleanupPhase6(t, f); f.db.Close() }()
	ctx := context.Background()
	const foreignVehicle = "00000000-0000-0000-0000-000000000014"
	defer func() { _, _ = f.db.Exec(ctx, `DELETE FROM vehicles WHERE id=$1`, foreignVehicle) }()
	if _, e := f.db.Exec(ctx, `INSERT INTO company_members(company_id,user_id,role) VALUES($1,$2,'driver')`, f.company, f.driver); e != nil {
		t.Fatal(e)
	}
	if _, e := f.db.Exec(ctx, `INSERT INTO vehicles(id,owner_type,driver_profile_id,make,model,year,license_plate,authorized_seats,vehicle_type) VALUES($1,'driver',$2,'Foreign','Vehicle',2024,'F6FOREIGN',4,'car')`, foreignVehicle, f.driverProfile); e != nil {
		t.Fatal(e)
	}
	_, appErr := tripuc.NewCreateUsecase(driverrepo.NewRepository(f.db), vehiclerepo.NewRepository(f.db), triprepo.NewRepository(f.db), companyrepo.NewRepository(f.db)).Execute(ctx, f.owner, tripuc.Input{PublisherType: "company", CompanyID: f.company, DriverProfileID: f.driverProfile, VehicleID: foreignVehicle, OriginLocalityID: f.locality, DestinationLocalityID: f.destination, DepartureAt: time.Now().Add(24 * time.Hour).UTC().Format(time.RFC3339), PricePerSeat: "1000", TotalSeats: 2, BookingMode: "automatic"})
	if appErr == nil || appErr.Code != "vehicles:not-found" {
		t.Fatalf("expected foreign vehicle rejection, got %v", appErr)
	}
}
func TestNonAdminCannotAccessAdminEndpoint(t *testing.T) {
	f := newPhase6Fixture(t)
	defer func() { cleanupPhase6(t, f); f.db.Close() }()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(auth.Middleware("phase6-secret"), auth.AdminMiddleware())
	r.GET("/admin/dashboard", func(c *gin.Context) { c.Status(200) })
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"user_id": f.outsider, "exp": time.Now().Add(time.Hour).Unix()})
	signed, e := token.SignedString([]byte("phase6-secret"))
	if e != nil {
		t.Fatal(e)
	}
	req := httptest.NewRequest(http.MethodGet, "/admin/dashboard", nil)
	req.Header.Set("Authorization", "Bearer "+signed)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d body=%s", w.Code, w.Body.String())
	}
}
func TestAdminVerificationCreatesAuditLog(t *testing.T) {
	f := newPhase6Fixture(t)
	defer func() { cleanupPhase6(t, f); f.db.Close() }()
	ctx := context.Background()
	roleID := "00000000-0000-0000-0000-000000000013"
	_, e := f.db.Exec(ctx, `INSERT INTO roles(id,code) VALUES($1,'ADMIN') ON CONFLICT(code) DO NOTHING`, roleID)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = f.db.Exec(ctx, `INSERT INTO user_roles(user_id,role_id) SELECT $1,id FROM roles WHERE code='ADMIN'`, f.owner); e != nil {
		t.Fatal(e)
	}
	if e = adminrepo.NewRepository(f.db).Verify(ctx, f.owner, "driver", f.driverProfile, "verified"); e != nil {
		t.Fatal(e)
	}
	var count int
	if e = f.db.QueryRow(ctx, `SELECT COUNT(*) FROM audit_logs WHERE actor_user_id=$1 AND action='admin.verification.updated' AND entity_id=$2`, f.owner, f.driverProfile).Scan(&count); e != nil {
		t.Fatal(e)
	}
	if count != 1 {
		t.Fatalf("expected one verification audit, got %d", count)
	}
}

var _ = pgx.ErrNoRows
