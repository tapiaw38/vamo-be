package authapi

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGetByEmailPropagatesBearerAndDistinguishesNotFound(t *testing.T) {
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.Header.Get("Authorization") != "Bearer access" {
			t.Errorf("authorization not propagated")
		}
		if request.URL.Path == "/missing" {
			response.WriteHeader(http.StatusNotFound)
			return
		}
		response.Header().Set("Content-Type", "application/json")
		_, _ = response.Write([]byte(`{"data":{"id":"u1","first_name":"Ada","last_name":"Lovelace","email":"ada@example.com"}}`))
	}))
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	server.Listener = listener
	server.Start()
	defer server.Close()
	client := NewClient(server.URL, time.Second).(*client)
	user, err := client.GetByEmail(t.Context(), "Bearer access", "ada@example.com")
	if err != nil || user == nil || user.Username != "u1" {
		t.Fatalf("unexpected user: %#v, %v", user, err)
	}
	missing, err := client.getUser(t.Context(), "Bearer access", "/missing", false)
	if err != nil || missing != nil {
		t.Fatalf("expected nil not-found user, got %#v, %v", missing, err)
	}
}
