package config

import "testing"

func TestLoadRequiresDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	if _, err := Load(); err == nil {
		t.Fatal("expected missing DATABASE_URL error")
	}
}

func TestLoadDefaults(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost/vamo")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Address() != "0.0.0.0:8080" {
		t.Fatalf("unexpected address: %s", cfg.Address())
	}
}
