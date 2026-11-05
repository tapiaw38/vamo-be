package integration

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"os"
	"testing"
)

func TestActiveVehicleLicensePlateIsUnique(t *testing.T) {
	url := os.Getenv("VAMO_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("set VAMO_TEST_DATABASE_URL after migrations")
	}
	db, err := pgxpool.New(context.Background(), url)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	tx, err := db.Begin(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(context.Background())
	userID := "phase3-test-user"
	driverID := "00000000-0000-0000-0000-000000000003"
	_, _ = tx.Exec(context.Background(), `DELETE FROM users WHERE id = $1`, userID)
	if _, err = tx.Exec(context.Background(), `INSERT INTO users(id) VALUES($1)`, userID); err != nil {
		t.Fatal(err)
	}
	if _, err = tx.Exec(context.Background(), `INSERT INTO driver_profiles(id, user_id) VALUES($1, $2)`, driverID, userID); err != nil {
		t.Fatal(err)
	}
	if _, err = tx.Exec(context.Background(), `INSERT INTO vehicles(owner_type, driver_profile_id, make, model, year, license_plate, authorized_seats, vehicle_type) VALUES('driver', $1, 'A', 'B', 2020, 'AA123AA', 4, 'car')`, driverID); err != nil {
		t.Fatal(err)
	}
	if _, err = tx.Exec(context.Background(), `INSERT INTO vehicles(owner_type, driver_profile_id, make, model, year, license_plate, authorized_seats, vehicle_type) VALUES('driver', $1, 'C', 'D', 2021, 'aa123aa', 4, 'car')`, driverID); err == nil || !isUniqueViolation(err) {
		t.Fatalf("expected active plate unique violation, got %v", err)
	}
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
