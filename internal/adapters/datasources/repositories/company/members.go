package company

import (
	"context"
	"github.com/tapiaw38/vamo/backend/internal/domain"
)

func (r *repository) CanManage(ctx context.Context, companyID, userID string) (bool, error) {
	var ok bool
	err := r.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM company_members WHERE company_id=$1 AND user_id=$2 AND status='active' AND role IN ('owner','admin'))`, companyID, userID).Scan(&ok)
	return ok, err
}
func (r *repository) CanDispatch(ctx context.Context, companyID, userID string) (bool, error) {
	var ok bool
	err := r.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM company_members WHERE company_id=$1 AND user_id=$2 AND status='active' AND role IN ('owner','admin','dispatcher'))`, companyID, userID).Scan(&ok)
	return ok, err
}
func (r *repository) DriverBelongs(ctx context.Context, companyID, driverID string) (bool, error) {
	var ok bool
	err := r.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM company_members m JOIN driver_profiles d ON d.user_id=m.user_id WHERE m.company_id=$1 AND d.id=$2 AND m.role='driver' AND m.status='active')`, companyID, driverID).Scan(&ok)
	return ok, err
}
func (r *repository) ListMembers(ctx context.Context, companyID, userID string) ([]domain.CompanyMember, error) {
	ok, err := r.CanManage(ctx, companyID, userID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrForbidden
	}
	rows, err := r.db.Query(ctx, `SELECT m.company_id,m.user_id,m.role,m.status,COALESCE(d.id::text,''),m.created_at,m.updated_at FROM company_members m LEFT JOIN driver_profiles d ON d.user_id=m.user_id WHERE m.company_id=$1 ORDER BY m.created_at`, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.CompanyMember{}
	for rows.Next() {
		var m domain.CompanyMember
		if err := rows.Scan(&m.CompanyID, &m.UserID, &m.Role, &m.Status, &m.DriverProfileID, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}
func (r *repository) AddMember(ctx context.Context, companyID string, m domain.CompanyMember, actor string) error {
	ok, err := r.CanManage(ctx, companyID, actor)
	if err != nil {
		return err
	}
	if !ok {
		return ErrForbidden
	}
	_, err = r.db.Exec(ctx, `INSERT INTO company_members(company_id,user_id,role,status) VALUES($1,$2,$3,'active') ON CONFLICT(company_id,user_id) DO UPDATE SET role=EXCLUDED.role,status='active',updated_at=NOW()`, companyID, m.UserID, m.Role)
	return err
}
func (r *repository) RemoveMember(ctx context.Context, companyID, userID, actor string) error {
	ok, err := r.CanManage(ctx, companyID, actor)
	if err != nil {
		return err
	}
	if !ok {
		return ErrForbidden
	}
	_, err = r.db.Exec(ctx, `UPDATE company_members SET status='inactive',updated_at=NOW() WHERE company_id=$1 AND user_id=$2 AND role <> 'owner'`, companyID, userID)
	return err
}
