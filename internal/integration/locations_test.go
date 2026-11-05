package integration

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"os"
	"testing"
)

func TestSeededLocalitiesBelongToExistingProvince(t *testing.T) {
	url := os.Getenv("VAMO_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("set VAMO_TEST_DATABASE_URL after migrations and Georef seed")
	}
	db, err := pgxpool.New(context.Background(), url)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var orphaned int
	if err := db.QueryRow(context.Background(), `SELECT COUNT(*) FROM localities l LEFT JOIN provinces p ON p.id = l.province_id WHERE p.id IS NULL`).Scan(&orphaned); err != nil {
		t.Fatal(err)
	}
	if orphaned != 0 {
		t.Fatalf("found %d orphaned localities", orphaned)
	}
}
