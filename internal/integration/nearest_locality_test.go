package integration

import (
	"context"
	"testing"

	locationrepo "github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/location"
	locations "github.com/tapiaw38/vamo/backend/internal/usecases/locations"
)

func TestNearestLocalityAgainstPostgres(t *testing.T) {
	db := phase6DB(t)
	defer db.Close()
	ctx := context.Background()
	rows, err := db.Query(ctx, `SELECT id, lat::double precision, lon::double precision FROM localities l WHERE province_id = '06' AND lat IS NOT NULL AND lon IS NOT NULL AND (SELECT COUNT(*) FROM localities same WHERE same.lat = l.lat AND same.lon = l.lon) = 1 ORDER BY id LIMIT 2`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	tested := 0
	for rows.Next() {
		var localityID string
		var lat, lon float64
		if err := rows.Scan(&localityID, &lat, &lon); err != nil {
			t.Fatal(err)
		}
		output, appErr := locations.NewNearestUsecase(locationrepo.NewRepository(db)).Execute(ctx, lat, lon)
		if appErr != nil {
			t.Fatalf("nearest failed: %s", appErr.Code)
		}
		if output.ID != localityID || output.Lat == nil || output.Lon == nil || *output.Lat != lat || *output.Lon != lon {
			t.Fatalf("unexpected nearest locality: got=%+v expected_id=%s query=(%f,%f)", output, localityID, lat, lon)
		}
		tested++
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	if tested != 2 {
		t.Fatalf("expected two distinct Buenos Aires localities with precise coordinates, got %d", tested)
	}
}
