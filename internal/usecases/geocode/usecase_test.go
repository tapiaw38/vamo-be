package geocode

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type failingRoundTripper struct{}

func (failingRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, context.DeadlineExceeded
}

func TestForwardReturnsCoordinatesAndUsesSearchEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/search" || r.Header.Get("User-Agent") == "" {
			t.Fatalf("unexpected request: %s user-agent=%q", r.URL.String(), r.Header.Get("User-Agent"))
		}
		_ = json.NewEncoder(w).Encode([]map[string]any{{"display_name": "Calle 1, Tinogasta", "lat": "-28.063", "lon": "-67.564", "address": map[string]string{"road": "Calle 1", "house_number": "10", "city": "Tinogasta"}}})
	}))
	defer server.Close()
	service := NewService(server.Client(), server.URL+"/reverse")
	results := service.Forward(context.Background(), "Calle 1", -28, -67)
	if len(results) != 1 || results[0].Lat != -28.063 || results[0].Lon != -67.564 {
		t.Fatalf("unexpected forward result: %+v", results)
	}
}

func TestForwardEmptyQueryReturnsEmptyWithoutUpstream(t *testing.T) {
	service := NewService(nil, "http://nominatim.test/reverse")
	if results := service.Forward(context.Background(), "  ", 0, 0); len(results) != 0 {
		t.Fatalf("expected empty results, got %+v", results)
	}
}

func TestReverseUpstreamFailureReturnsEmptyResult(t *testing.T) {
	service := NewService(&http.Client{Transport: failingRoundTripper{}}, "http://nominatim.test/reverse")
	result := service.Reverse(context.Background(), -27.7, -67.2)
	if result.Address != "" || result.Street != "" || result.LocalityName != "" {
		t.Fatalf("expected empty fallback, got %+v", result)
	}
}
