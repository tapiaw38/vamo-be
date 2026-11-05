package admin

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct{ db *pgxpool.Pool }

func NewRepository(db *pgxpool.Pool) *Repository { return &Repository{db} }
func (r *Repository) Dashboard(ctx context.Context) (map[string]any, error) {
	out := map[string]any{}
	queries := map[string]string{"users": "SELECT COUNT(*) FROM users", "drivers": "SELECT COUNT(*) FROM driver_profiles", "companies": "SELECT COUNT(*) FROM companies", "published_trips": "SELECT COUNT(*) FROM trips WHERE status IN ('published','full')", "bookings": "SELECT COUNT(*) FROM bookings"}
	for k, q := range queries {
		var n int64
		if err := r.db.QueryRow(ctx, q).Scan(&n); err != nil {
			return nil, err
		}
		out[k] = n
	}
	rows, err := r.db.Query(ctx, `SELECT p.name,COUNT(*) FROM trips t JOIN localities l ON l.id=t.origin_locality_id JOIN provinces p ON p.id=l.province_id GROUP BY p.name ORDER BY p.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	byProvince := map[string]int64{}
	for rows.Next() {
		var name string
		var n int64
		if err := rows.Scan(&name, &n); err != nil {
			return nil, err
		}
		byProvince[name] = n
	}
	out["trips_by_province"] = byProvince
	return out, rows.Err()
}
func (r *Repository) List(ctx context.Context, kind string) ([]map[string]any, error) {
	var q string
	switch kind {
	case "users":
		q = `SELECT id,first_name,last_name,phone,created_at FROM users ORDER BY created_at DESC`
	case "drivers":
		q = `SELECT id,user_id,status,created_at FROM driver_profiles ORDER BY created_at DESC`
	case "companies":
		q = `SELECT id,legal_name,display_name,tax_id,status,verification_status,created_at FROM companies ORDER BY created_at DESC`
	case "trips":
		q = `SELECT id,publisher_type,company_id,driver_profile_id,departure_at,status,price_per_seat::text,total_seats,available_seats FROM trips ORDER BY departure_at DESC`
	case "bookings":
		q = `SELECT id,trip_id,passenger_id,seats,total_price::text,status,requested_at FROM bookings ORDER BY requested_at DESC`
	default:
		return nil, errors.New("invalid admin list")
	}
	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	fields := rows.FieldDescriptions()
	out := []map[string]any{}
	for rows.Next() {
		vals, err := rows.Values()
		if err != nil {
			return nil, err
		}
		item := map[string]any{}
		for i, f := range fields {
			item[string(f.Name)] = vals[i]
		}
		out = append(out, item)
	}
	return out, rows.Err()
}
func (r *Repository) Verify(ctx context.Context, actor, kind, id, status string) error {
	allowed := map[string]bool{"driver": true, "vehicle": true, "company": true}
	if !allowed[kind] || (status != "verified" && status != "suspended") {
		return errors.New("invalid verification target")
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var table string
	switch kind {
	case "driver":
		table = "driver_profiles"
	case "vehicle":
		table = "vehicles"
	case "company":
		table = "companies"
	}
	var q string
	if kind == "driver" {
		q = `UPDATE driver_profiles SET status=$2,updated_at=NOW() WHERE id=$1`
	} else {
		q = `UPDATE ` + table + ` SET verification_status=$2,updated_at=NOW() WHERE id=$1`
	}
	tag, err := tx.Exec(ctx, q, id, status)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return pgx.ErrNoRows
	}
	meta, _ := json.Marshal(map[string]any{"status": status})
	if _, err = tx.Exec(ctx, `INSERT INTO audit_logs(actor_user_id,action,entity_type,entity_id,metadata) VALUES($1,$2,$3,$4,$5)`, actor, "admin.verification.updated", kind, id, meta); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
