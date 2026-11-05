package integration

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	userrepo "github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/user"
)

func TestUserProfileResolvesResidenceInSingleRepositoryRead(t *testing.T) {
	url := os.Getenv("VAMO_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("set VAMO_TEST_DATABASE_URL after migrations")
	}
	db, err := pgxpool.New(context.Background(), url)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	ctx := context.Background()
	provinceID, localityID, userID := "profile-test-province", "profile-test-locality", "profile-test-user"
	_, _ = db.Exec(ctx, `DELETE FROM users WHERE id=$1`, userID)
	_, _ = db.Exec(ctx, `DELETE FROM localities WHERE id=$1`, localityID)
	_, _ = db.Exec(ctx, `DELETE FROM provinces WHERE id=$1`, provinceID)
	if _, err = db.Exec(ctx, `INSERT INTO provinces(id,name,code) VALUES($1,$2,$3)`, provinceID, "Provincia Test", "PT"); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ctx, `INSERT INTO localities(id,province_id,name,normalized_name) VALUES($1,$2,$3,$4)`, localityID, provinceID, "Localidad Test", "localidad test"); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_, _ = db.Exec(ctx, `DELETE FROM users WHERE id=$1`, userID)
		_, _ = db.Exec(ctx, `DELETE FROM localities WHERE id=$1`, localityID)
		_, _ = db.Exec(ctx, `DELETE FROM provinces WHERE id=$1`, provinceID)
	}()
	if _, err = db.Exec(ctx, `INSERT INTO users(id,first_name,last_name,residence_locality_id) VALUES($1,$2,$3,$4)`, userID, "Perfil", "Test", localityID); err != nil {
		t.Fatal(err)
	}
	profile, err := userrepo.NewRepository(db).Get(ctx, userID)
	if err != nil {
		t.Fatal(err)
	}
	if profile.ResidenceLocalityName == nil || *profile.ResidenceLocalityName != "Localidad Test" || profile.ResidenceProvinceName == nil || *profile.ResidenceProvinceName != "Provincia Test" || profile.ResidenceProvinceID == nil || *profile.ResidenceProvinceID != provinceID {
		t.Fatalf("unexpected resolved residence: %+v", profile)
	}
}
