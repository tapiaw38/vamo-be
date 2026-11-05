package routing

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
	Coordinates [][2]float64 `json:"coordinates"`
	DistanceM   float64      `json:"distance_m"`
	DurationS   float64      `json:"duration_s"`
}

type cacheEntry struct {
	result    Result
	expiresAt time.Time
}

type Service struct {
	client      *http.Client
	baseURL     string
	mu          sync.Mutex
	lastRequest time.Time
	cache       map[string]cacheEntry
}

func NewService(client *http.Client, baseURL string) *Service {
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}
	if baseURL == "" {
		baseURL = "https://router.project-osrm.org"
	}
	return &Service{client: client, baseURL: strings.TrimRight(baseURL, "/"), cache: make(map[string]cacheEntry)}
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

func (s *Service) Route(ctx context.Context, originLat, originLon, destinationLat, destinationLon float64) Result {
	key := fmt.Sprintf("%.5f,%.5f;%.5f,%.5f", originLat, originLon, destinationLat, destinationLon)
	s.mu.Lock()
	if cached, ok := s.cache[key]; ok && time.Now().Before(cached.expiresAt) {
		s.mu.Unlock()
		return cached.result
	}
	s.mu.Unlock()
	if !s.acquire(ctx) {
		return Result{Coordinates: [][2]float64{}}
	}
	coordinates := url.PathEscape(strconv.FormatFloat(originLon, 'f', 6, 64) + "," + strconv.FormatFloat(originLat, 'f', 6, 64) + ";" + strconv.FormatFloat(destinationLon, 'f', 6, 64) + "," + strconv.FormatFloat(destinationLat, 'f', 6, 64))
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, s.baseURL+"/route/v1/driving/"+coordinates+"?overview=full&geometries=geojson", nil)
	if err != nil {
		return Result{Coordinates: [][2]float64{}}
	}
	response, err := s.client.Do(request)
	if err != nil {
		return Result{Coordinates: [][2]float64{}}
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return Result{Coordinates: [][2]float64{}}
	}
	var payload struct {
		Routes []struct {
			Distance float64 `json:"distance"`
			Duration float64 `json:"duration"`
			Geometry struct {
				Coordinates [][2]float64 `json:"coordinates"`
			} `json:"geometry"`
		} `json:"routes"`
	}
	if json.NewDecoder(response.Body).Decode(&payload) != nil || len(payload.Routes) == 0 {
		return Result{Coordinates: [][2]float64{}}
	}
	route := payload.Routes[0]
	result := Result{Coordinates: make([][2]float64, 0, len(route.Geometry.Coordinates)), DistanceM: route.Distance, DurationS: route.Duration}
	for _, coordinate := range route.Geometry.Coordinates {
		result.Coordinates = append(result.Coordinates, [2]float64{coordinate[1], coordinate[0]})
	}
	s.mu.Lock()
	s.cache[key] = cacheEntry{result: result, expiresAt: time.Now().Add(time.Hour)}
	s.mu.Unlock()
	return result
}
