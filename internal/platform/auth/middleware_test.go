package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestMiddlewareRejectsInvalidAndExpiredTokens(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for name, token := range map[string]string{
		"invalid": "not-a-jwt",
		"expired": func() string {
			value, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"user_id": "u1", "exp": time.Now().Add(-time.Minute).Unix()}).SignedString([]byte("secret"))
			return value
		}(),
	} {
		t.Run(name, func(t *testing.T) {
			router := gin.New()
			router.Use(Middleware("secret"))
			router.GET("/private", func(c *gin.Context) { c.Status(http.StatusNoContent) })
			response := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/private", nil)
			request.Header.Set("Authorization", "Bearer "+token)
			router.ServeHTTP(response, request)
			if response.Code != http.StatusUnauthorized {
				t.Fatalf("expected 401, got %d", response.Code)
			}
		})
	}
}

func TestMiddlewareExtractsAuthAPIClaimsAndRoleGuards(t *testing.T) {
	gin.SetMode(gin.TestMode)
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": "driver-1",
		"roles":   []map[string]string{{"name": "user"}, {"name": "admin"}},
		"exp":     time.Now().Add(time.Hour).Unix(),
	}).SignedString([]byte("secret"))
	if err != nil {
		t.Fatal(err)
	}
	router := gin.New()
	router.Use(Middleware("secret"), RequireConductor())
	router.GET("/private", func(c *gin.Context) { c.Status(http.StatusNoContent) })
	request := httptest.NewRequest(http.MethodGet, "/private", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusNoContent {
		t.Fatalf("expected conductor claim to pass, got %d", response.Code)
	}
}
