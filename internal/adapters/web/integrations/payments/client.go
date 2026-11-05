package payments

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type PreferenceRequest struct {
	Items             []PreferenceItem `json:"items"`
	PayerEmail        string           `json:"payer_email"`
	ExternalReference string           `json:"external_reference"`
	UserID            string           `json:"user_id"`
	BackURLs          BackURLs         `json:"back_urls"`
	NotificationURL   string           `json:"notification_url,omitempty"`
}
type PreferenceItem struct {
	Title      string `json:"title"`
	Quantity   int    `json:"quantity"`
	UnitPrice  string `json:"unit_price"`
	CurrencyID string `json:"currency_id"`
}
type BackURLs struct {
	Success string `json:"success"`
	Failure string `json:"failure"`
	Pending string `json:"pending"`
}
type PreferenceResponse struct {
	PreferenceID     string `json:"preference_id"`
	InitPoint        string `json:"init_point"`
	SandboxInitPoint string `json:"sandbox_init_point"`
	PaymentID        *int64 `json:"payment_id"`
}
type Client interface {
	CreatePreference(context.Context, PreferenceRequest) (PreferenceResponse, error)
}
type client struct {
	url, apiKey string
	httpClient  *http.Client
}

func NewClient(url, apiKey string, timeout time.Duration) Client {
	return &client{url: url, apiKey: apiKey, httpClient: &http.Client{Timeout: timeout}}
}
func (c *client) CreatePreference(ctx context.Context, input PreferenceRequest) (PreferenceResponse, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return PreferenceResponse{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, bytes.NewReader(body))
	if err != nil {
		return PreferenceResponse{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", c.apiKey)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return PreferenceResponse{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return PreferenceResponse{}, fmt.Errorf("payments upstream status %d", resp.StatusCode)
	}
	var out PreferenceResponse
	if err = json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return PreferenceResponse{}, err
	}
	return out, nil
}
