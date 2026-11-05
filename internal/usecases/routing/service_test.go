package routing

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

type failingRoundTripper struct{}

func (failingRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, context.DeadlineExceeded
}

func TestRouteProviderFailureReturnsEmptyCoordinates(t *testing.T) {
	service := NewService(&http.Client{Transport: failingRoundTripper{}}, "http://osrm.test")
	result := service.Route(context.Background(), -28, -67, -28.1, -67.1)
	if len(result.Coordinates) != 0 {
		t.Fatalf("expected empty route, got %+v", result)
	}
}

func TestRouteCachesProviderResponse(t *testing.T) {
	count := 0
	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		count++
		body := `{"routes":[{"distance":120,"duration":30,"geometry":{"coordinates":[[-67,-28],[-67.1,-28.1]]}}]}`
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})}
	service := NewService(client, "http://osrm.test")
	first := service.Route(context.Background(), -28, -67, -28.1, -67.1)
	second := service.Route(context.Background(), -28, -67, -28.1, -67.1)
	if count != 1 || len(first.Coordinates) != 2 || len(second.Coordinates) != 2 || first.Coordinates[0] != [2]float64{-28, -67} {
		t.Fatalf("expected cached route, count=%d first=%+v second=%+v", count, first, second)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) { return f(request) }
