package health

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGetHealth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/health", NewGetHandler())
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/health", nil)
	router.ServeHTTP(recorder, request)
	if recorder.Code != 200 {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
}
