package payments

import (
	"github.com/gin-gonic/gin"
	uc "github.com/tapiaw38/vamo/backend/internal/usecases/payments"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWebhookRejectsMissingSignature(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/webhook", NewWebhook(uc.Usecases{}, "secret"))
	req := httptest.NewRequest(http.MethodPost, "/webhook", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}
