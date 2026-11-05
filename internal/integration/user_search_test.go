package integration

import (
	"context"
	"testing"

	userrepo "github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/user"
	useruc "github.com/tapiaw38/vamo/backend/internal/usecases/users"
)

func TestUserSearchReturnsMinimalFieldsAgainstPostgres(t *testing.T) {
	db := phase6DB(t)
	defer db.Close()
	ctx := context.Background()
	const id = "user-search-real"
	defer func() { _, _ = db.Exec(ctx, `DELETE FROM users WHERE id=$1`, id) }()
	if _, err := db.Exec(ctx, `INSERT INTO users(id,first_name,last_name,email) VALUES($1,'Search','Target','target-search@example.test') ON CONFLICT(id) DO UPDATE SET first_name=EXCLUDED.first_name,last_name=EXCLUDED.last_name,email=EXCLUDED.email`, id); err != nil {
		t.Fatal(err)
	}
	items, appErr := useruc.NewSearchUsecase(userrepo.NewRepository(db)).Execute(ctx, "target-search@example")
	if appErr != nil || len(items) != 1 {
		t.Fatalf("unexpected search result: %+v %v", items, appErr)
	}
	if items[0].ID != id || items[0].Email != "target-search@example.test" {
		t.Fatalf("unexpected minimal result: %+v", items[0])
	}
}
