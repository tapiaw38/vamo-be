package payments

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/gin-gonic/gin"
	uc "github.com/tapiaw38/vamo/backend/internal/usecases/payments"
	"io"
	"net/http"
)

func NewStart(u uc.Usecases) gin.HandlerFunc {
	return func(c *gin.Context) {
		o, e := u.Start(c.Request.Context(), c.GetString("vamo.user_id"), c.Param("id"))
		if e != nil {
			c.JSON(e.Status, e)
			return
		}
		c.JSON(200, gin.H{"data": o})
	}
}
func NewWebhook(u uc.Usecases, secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		body, e := io.ReadAll(c.Request.Body)
		if e != nil {
			c.Status(400)
			return
		}
		sig := c.GetHeader("X-Payments-Signature")
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write(body)
		expected := hex.EncodeToString(mac.Sum(nil))
		if secret == "" || sig == "" || !hmac.Equal([]byte(sig), []byte(expected)) {
			c.JSON(http.StatusUnauthorized, gin.H{"code": "payments:invalid-signature", "message": "valid payment signature required"})
			return
		}
		var in struct {
			BookingID         string `json:"booking_id"`
			ProviderPaymentID string `json:"provider_payment_id"`
			Status            string `json:"status"`
		}
		if json.Unmarshal(body, &in) != nil {
			c.Status(400)
			return
		}
		if e := u.Webhook(c.Request.Context(), in.BookingID, in.ProviderPaymentID, in.Status); e != nil {
			c.JSON(e.Status, e)
			return
		}
		c.Status(http.StatusNoContent)
	}
}
