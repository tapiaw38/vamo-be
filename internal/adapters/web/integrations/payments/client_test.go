package payments

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestCreatePreferenceSendsAPIKeyAndStringAmount(t *testing.T) {
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-API-Key") != "test-key" {
			t.Fatalf("missing api key")
		}
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), `"unit_price":"25000.00"`) {
			t.Fatalf("amount was not sent as string: %s", body)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"preference_id":"pref-1","init_point":"https://example.test"}`))
	}))
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	server.Listener = listener
	server.Start()
	defer server.Close()
	client := NewClient(server.URL, "test-key", time.Second)
	out, err := client.CreatePreference(context.Background(), PreferenceRequest{Items: []PreferenceItem{{Title: "trip", Quantity: 1, UnitPrice: "25000.00", CurrencyID: "ARS"}}, ExternalReference: "vamo:booking:1", UserID: "user-1"})
	if err != nil || out.PreferenceID != "pref-1" {
		t.Fatalf("unexpected result=%+v err=%v", out, err)
	}
}
