package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	pricingrepo "github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/pricing"
	pricinghandlers "github.com/tapiaw38/vamo/backend/internal/adapters/web/handlers/pricing"
	"github.com/tapiaw38/vamo/backend/internal/platform/auth"
	pricinguc "github.com/tapiaw38/vamo/backend/internal/usecases/pricing"
)

func TestPricingAdminValidationAuthorizationAndAuditAgainstPostgres(t *testing.T) {
	db := phase6DB(t)
	defer db.Close()
	ctx := context.Background()
	const actor = "pricing-admin-real"
	defer func() {
		_, _ = db.Exec(ctx, `DELETE FROM audit_logs WHERE actor_user_id=$1`, actor)
		_, _ = db.Exec(ctx, `DELETE FROM users WHERE id=$1`, actor)
	}()
	if _, err := db.Exec(ctx, `INSERT INTO users(id) VALUES($1)`, actor); err != nil {
		t.Fatal(err)
	}
	repository := pricingrepo.NewRepository(db)
	current, err := repository.Get(ctx)
	if err != nil {
		t.Fatal(err)
	}
	pricing := pricinguc.NewUsecases(repository)
	var locality string
	var lat, lon float64
	if err := db.QueryRow(ctx, `SELECT id,lat,lon FROM localities WHERE lat IS NOT NULL AND lon IS NOT NULL ORDER BY id LIMIT 1`).Scan(&locality, &lat, &lon); err != nil {
		t.Fatal(err)
	}
	quoteRouter := gin.New()
	quoteRouter.GET("/quote", pricinghandlers.NewQuote(pricing))
	quoteRequest := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/quote?origin_locality_id=%s&destination_locality_id=%s&origin_lat=%f&origin_lon=%f&destination_lat=%f&destination_lon=%f", locality, locality, lat, lon, lat+0.01, lon+0.01), nil)
	quoteResponse := httptest.NewRecorder()
	quoteRouter.ServeHTTP(quoteResponse, quoteRequest)
	if quoteResponse.Code != http.StatusOK || !strings.Contains(quoteResponse.Body.String(), `"fare_kind":"urban"`) {
		t.Fatalf("urban quote expected 200/urban, got %d body=%s", quoteResponse.Code, quoteResponse.Body.String())
	}
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(auth.Middleware("pricing-secret"), auth.AdminMiddleware())
	router.PUT("/admin/pricing", pricinghandlers.NewUpdate(pricing))
	request := func(roles []map[string]any, input map[string]string) *httptest.ResponseRecorder {
		payload, _ := json.Marshal(input)
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"user_id": actor, "roles": roles, "exp": time.Now().Add(time.Hour).Unix()})
		signed, _ := token.SignedString([]byte("pricing-secret"))
		req := httptest.NewRequest(http.MethodPut, "/admin/pricing", strings.NewReader(string(payload)))
		req.Header.Set("Authorization", "Bearer "+signed)
		req.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		return response
	}
	baseInput := map[string]string{"base_price": current.BasePrice, "price_per_km": current.PricePerKm, "urban_base_price": current.UrbanBasePrice, "urban_price_per_km": current.UrbanPricePerKm, "currency": current.Currency}
	for _, field := range []string{"base_price", "price_per_km", "urban_base_price", "urban_price_per_km"} {
		input := map[string]string{}
		for key, value := range baseInput {
			input[key] = value
		}
		input[field] = "-1"
		if response := request([]map[string]any{{"name": "superadmin"}}, input); response.Code != http.StatusBadRequest {
			t.Fatalf("negative %s expected 400, got %d body=%s", field, response.Code, response.Body.String())
		}
	}
	if response := request([]map[string]any{{"name": "admin"}}, baseInput); response.Code != http.StatusForbidden {
		t.Fatalf("non-superadmin expected 403, got %d", response.Code)
	}
	input := map[string]string{"base_price": current.BasePrice, "price_per_km": current.PricePerKm, "urban_base_price": "1300.00", "urban_price_per_km": "450.00", "currency": current.Currency}
	if response := request([]map[string]any{{"name": "superadmin"}}, input); response.Code != http.StatusOK {
		t.Fatalf("superadmin update expected 200, got %d body=%s", response.Code, response.Body.String())
	}
	defer func() {
		_, _ = db.Exec(ctx, `UPDATE pricing_settings SET base_price=$1::numeric, price_per_km=$2::numeric, urban_base_price=$3::numeric, urban_price_per_km=$4::numeric, currency=$5, updated_at=$6, updated_by=$7 WHERE id=1`, current.BasePrice, current.PricePerKm, current.UrbanBasePrice, current.UrbanPricePerKm, current.Currency, current.UpdatedAt, current.UpdatedBy)
	}()
	var metadata string
	if err := db.QueryRow(ctx, `SELECT metadata::text FROM audit_logs WHERE actor_user_id=$1 AND action='admin.pricing.updated' ORDER BY created_at DESC LIMIT 1`, actor).Scan(&metadata); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(metadata, `"old_urban_base_price"`) || !strings.Contains(metadata, `"urban_base_price": "1300.00"`) || !strings.Contains(metadata, `"urban_price_per_km": "450.00"`) {
		t.Fatalf("urban pricing audit metadata incomplete: %s", metadata)
	}
}
