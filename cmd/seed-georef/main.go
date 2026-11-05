package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
	"unicode"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	provincesURL  = "https://apis.datos.gob.ar/georef/api/provincias?max=1000"
	localitiesURL = "https://apis.datos.gob.ar/georef/api/localidades?max=5000"
)

type provinceResponse struct {
	Provinces []struct {
		ID   string `json:"id"`
		Name string `json:"nombre"`
	} `json:"provincias"`
}
type localityResponse struct {
	Localities []struct {
		ID       string `json:"id"`
		Name     string `json:"nombre"`
		Province struct {
			ID string `json:"id"`
		} `json:"provincia"`
		Centroide struct {
			Lat float64 `json:"lat"`
			Lon float64 `json:"lon"`
		} `json:"centroide"`
		// Centroid is accepted for compatibility with alternate Georef payloads.
		Centroid struct {
			Lat float64 `json:"lat"`
			Lon float64 `json:"lon"`
		} `json:"centroid"`
	} `json:"localidades"`
}

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		panic("DATABASE_URL is required")
	}
	ctx := context.Background()
	db, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	client := &http.Client{Timeout: 60 * time.Second}
	var provinces provinceResponse
	fetchJSON(client, provincesURL, &provinces)
	for _, province := range provinces.Provinces {
		_, err = db.Exec(ctx, `INSERT INTO provinces (id, name, code) VALUES ($1, $2, $1) ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, code = EXCLUDED.code`, province.ID, province.Name)
		if err != nil {
			panic(err)
		}
	}
	var localities localityResponse
	fetchJSON(client, localitiesURL, &localities)
	for _, locality := range localities.Localities {
		if locality.Province.ID == "" || locality.Name == "" {
			continue
		}
		lat, lon := locality.Centroide.Lat, locality.Centroide.Lon
		if lat == 0 && lon == 0 {
			lat, lon = locality.Centroid.Lat, locality.Centroid.Lon
		}
		_, err = db.Exec(ctx, `INSERT INTO localities (id, province_id, name, normalized_name, official_code, lat, lon) VALUES ($1, $2, $3, $4, $1, NULLIF($5::numeric, 0), NULLIF($6::numeric, 0)) ON CONFLICT (province_id, normalized_name) DO UPDATE SET name = EXCLUDED.name, official_code = EXCLUDED.official_code, lat = EXCLUDED.lat, lon = EXCLUDED.lon`, locality.ID, locality.Province.ID, locality.Name, normalize(locality.Name), lat, lon)
		if err != nil {
			panic(err)
		}
	}
	fmt.Printf("Georef seed completed: %d provinces, %d localities\n", len(provinces.Provinces), len(localities.Localities))
}

func fetchJSON(client *http.Client, url string, target any) {
	response, err := client.Get(url)
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()
	if response.StatusCode >= 300 {
		panic(fmt.Sprintf("Georef returned %s", response.Status))
	}
	if err := json.NewDecoder(response.Body).Decode(target); err != nil {
		panic(err)
	}
}

func normalize(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.NewReplacer("á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u", "ü", "u", "ñ", "n").Replace(value)
	var result []rune
	previousSpace := false
	for _, character := range value {
		if unicode.IsLetter(character) || unicode.IsNumber(character) {
			result = append(result, character)
			previousSpace = false
		} else if !previousSpace {
			result = append(result, ' ')
			previousSpace = true
		}
	}
	return strings.TrimSpace(string(result))
}
