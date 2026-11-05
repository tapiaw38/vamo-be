package company

import (
	"context"
	"encoding/json"
	"github.com/jackc/pgx/v5"
	"github.com/tapiaw38/vamo/backend/internal/domain"
)

const companyCols = `id, legal_name, display_name, tax_id, province_id, locality_id, status, verification_status, contact_email, contact_phone, created_at, updated_at`
const companyColsAliased = `c.id, c.legal_name, c.display_name, c.tax_id, c.province_id, c.locality_id, c.status, c.verification_status, c.contact_email, c.contact_phone, c.created_at, c.updated_at`

func scanCompany(row interface{ Scan(...any) error }) (*domain.Company, error) {
	var c domain.Company
	err := row.Scan(&c.ID, &c.LegalName, &c.DisplayName, &c.TaxID, &c.ProvinceID, &c.LocalityID, &c.Status, &c.VerificationStatus, &c.ContactEmail, &c.ContactPhone, &c.CreatedAt, &c.UpdatedAt)
	return &c, err
}
func (r *repository) Create(ctx context.Context, c domain.Company, owner string) (*domain.Company, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	item, err := scanCompany(tx.QueryRow(ctx, `INSERT INTO companies(legal_name,display_name,tax_id,province_id,locality_id,contact_email,contact_phone,status) VALUES($1,$2,$3,$4,$5,$6,$7,'draft') RETURNING `+companyCols, c.LegalName, c.DisplayName, c.TaxID, c.ProvinceID, c.LocalityID, c.ContactEmail, c.ContactPhone))
	if err != nil {
		return nil, err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO company_members(company_id,user_id,role) VALUES($1,$2,'owner')`, item.ID, owner); err != nil {
		return nil, err
	}
	if err = insertAudit(ctx, tx, owner, "company.created", "company", item.ID); err != nil {
		return nil, err
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return item, nil
}
func (r *repository) Get(ctx context.Context, id, user string) (*domain.Company, error) {
	var allowed bool
	if err := r.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM company_members WHERE company_id=$1 AND user_id=$2 AND status='active')`, id, user).Scan(&allowed); err != nil {
		return nil, err
	}
	if !allowed {
		return nil, ErrForbidden
	}
	item, err := scanCompany(r.db.QueryRow(ctx, `SELECT `+companyCols+` FROM companies WHERE id=$1`, id))
	if noRows(err) {
		return nil, ErrNotFound
	}
	return item, err
}
func (r *repository) Update(ctx context.Context, c domain.Company, user string) (*domain.Company, error) {
	ok, err := r.CanManage(ctx, c.ID, user)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrForbidden
	}
	item, err := scanCompany(r.db.QueryRow(ctx, `UPDATE companies SET legal_name=$2,display_name=$3,tax_id=$4,province_id=$5,locality_id=$6,contact_email=$7,contact_phone=$8,updated_at=NOW() WHERE id=$1 RETURNING `+companyCols, c.ID, c.LegalName, c.DisplayName, c.TaxID, c.ProvinceID, c.LocalityID, c.ContactEmail, c.ContactPhone))
	return item, err
}
func (r *repository) ListMine(ctx context.Context, user string) ([]domain.Company, error) {
	rows, err := r.db.Query(ctx, `SELECT `+companyColsAliased+` FROM companies c JOIN company_members m ON m.company_id=c.id WHERE m.user_id=$1 AND m.status='active' ORDER BY c.display_name`, user)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Company{}
	for rows.Next() {
		item, e := scanCompany(rows)
		if e != nil {
			return nil, e
		}
		out = append(out, *item)
	}
	return out, rows.Err()
}
func insertAudit(ctx context.Context, tx pgx.Tx, actor, action, typ, id string) error {
	metadata, _ := json.Marshal(map[string]any{})
	_, err := tx.Exec(ctx, `INSERT INTO audit_logs(actor_user_id,action,entity_type,entity_id,metadata) VALUES($1,$2,$3,$4,$5)`, actor, action, typ, id, metadata)
	return err
}
