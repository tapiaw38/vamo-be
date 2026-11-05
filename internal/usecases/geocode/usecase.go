package geocode

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Result struct {
	Address      string  `json:"address"`
	Street       string  `json:"street"`
	HouseNumber  string  `json:"house_number"`
	LocalityName string  `json:"locality_name"`
	Lat          float64 `json:"lat"`
	Lon          float64 `json:"lon"`
}

type cacheEntry struct {
	result    Result
	results   []Result
	expiresAt time.Time
}

func (s *Service) acquire(ctx context.Context) bool {
	for {
		s.mu.Lock()
		wait := time.Second - time.Since(s.lastRequest)
		if wait <= 0 {
			s.lastRequest = time.Now()
			s.mu.Unlock()
			return true
		}
		s.mu.Unlock()
		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			timer.Stop()
			return false
		case <-timer.C:
		}
	}
}

type Service struct {
	client      *http.Client
	baseURL     string
	userAgent   string
	mu          sync.Mutex
	lastRequest time.Time
	cache       map[string]cacheEntry
}

func NewService(client *http.Client, baseURL string) *Service {
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}
	if baseURL == "" {
		baseURL = "https://nominatim.openstreetmap.org/reverse"
	}
	return &Service{client: client, baseURL: baseURL, userAgent: "VAMO/1.0 (contacto@vamo.app)", cache: make(map[string]cacheEntry)}
}

func (s *Service) Reverse(ctx context.Context, lat, lon float64) Result {
	key := fmt.Sprintf("%.5f,%.5f", lat, lon)
	for {
		s.mu.Lock()
		if cached, ok := s.cache[key]; ok && time.Now().Before(cached.expiresAt) {
			s.mu.Unlock()
			return cached.result
		}
		s.mu.Unlock()
		if !s.acquire(ctx) {
			return Result{}
		}
		break
	}

	query := url.Values{"format": {"jsonv2"}, "lat": {strconv.FormatFloat(lat, 'f', 6, 64)}, "lon": {strconv.FormatFloat(lon, 'f', 6, 64)}}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.baseURL+"?"+query.Encode(), nil)
	if err != nil {
		return Result{}
	}
	req.Header.Set("User-Agent", s.userAgent)
	response, err := s.client.Do(req)
	if err != nil {
		return Result{}
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return Result{}
	}
	var payload struct {
		DisplayName string            `json:"display_name"`
		Address     map[string]string `json:"address"`
	}
	if json.NewDecoder(response.Body).Decode(&payload) != nil {
		return Result{}
	}
	result := Result{Address: payload.DisplayName, Street: payload.Address["road"], HouseNumber: payload.Address["house_number"], LocalityName: firstAddress(payload.Address, "city", "town", "village", "municipality")}
	s.mu.Lock()
	s.cache[key] = cacheEntry{result: result, expiresAt: time.Now().Add(10 * time.Minute)}
	s.mu.Unlock()
	return result
}

func (s *Service) Forward(ctx context.Context, query string, biasLat, biasLon float64) []Result {
	query = strings.TrimSpace(query)
	if query == "" {
		return []Result{}
	}
	key := fmt.Sprintf("forward:%s:%.3f,%.3f", strings.ToLower(query), biasLat, biasLon)
	s.mu.Lock()
	if cached, ok := s.cache[key]; ok && time.Now().Before(cached.expiresAt) {
		results := append([]Result(nil), cached.results...)
		s.mu.Unlock()
		return results
	}
	s.mu.Unlock()
	if !s.acquire(ctx) {
		return []Result{}
	}
	viewbox := fmt.Sprintf("%f,%f,%f,%f", biasLon-0.5, biasLat+0.5, biasLon+0.5, biasLat-0.5)
	params := url.Values{"format": {"jsonv2"}, "q": {query}, "countrycodes": {"ar"}, "limit": {"6"}, "viewbox": {viewbox}, "bounded": {"0"}}
	endpoint := strings.TrimSuffix(s.baseURL, "/reverse") + "/search"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+params.Encode(), nil)
	if err != nil {
		return []Result{}
	}
	req.Header.Set("User-Agent", s.userAgent)
	response, err := s.client.Do(req)
	if err != nil {
		return []Result{}
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return []Result{}
	}
	var payload []struct {
		DisplayName string            `json:"display_name"`
		Lat         string            `json:"lat"`
		Lon         string            `json:"lon"`
		Address     map[string]string `json:"address"`
	}
	if json.NewDecoder(response.Body).Decode(&payload) != nil {
		return []Result{}
	}
	results := make([]Result, 0, len(payload))
	for _, item := range payload {
		lat, latErr := strconv.ParseFloat(item.Lat, 64)
		lon, lonErr := strconv.ParseFloat(item.Lon, 64)
		if latErr != nil || lonErr != nil {
			continue
		}
		results = append(results, Result{Address: item.DisplayName, Street: item.Address["road"], HouseNumber: item.Address["house_number"], LocalityName: firstAddress(item.Address, "city", "town", "village", "municipality"), Lat: lat, Lon: lon})
	}
	s.mu.Lock()
	s.cache[key] = cacheEntry{results: append([]Result(nil), results...), expiresAt: time.Now().Add(10 * time.Minute)}
	s.mu.Unlock()
	return results
}

func firstAddress(address map[string]string, keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(address[key]); value != "" {
			return value
		}
	}
	return ""
}
