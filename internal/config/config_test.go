package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadEnvOverridesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	contents := "base_url: https://file.example\n" +
		"origin_host: file-origin.example\n" +
		"username: guardian\n" +
		"password: secret\n"
	if err := os.WriteFile(path, []byte(contents), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv(EnvPrefix+"_BASE_URL", "https://env.example/")
	t.Setenv(EnvPrefix+"_ORIGIN_HOST", "env-origin.example")

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.BaseURL != "https://env.example" {
		t.Fatalf("BaseURL = %q", cfg.BaseURL)
	}
	if cfg.OriginHost != "env-origin.example" {
		t.Fatalf("OriginHost = %q", cfg.OriginHost)
	}
	if cfg.Username != "guardian" || cfg.Password != "secret" {
		t.Fatalf("credentials were not loaded")
	}
}

func TestRedactedHidesCredentials(t *testing.T) {
	cfg := Config{Username: "guardian@example.com", Password: "secret"}
	redacted := cfg.Redacted()
	if redacted["username"] != "redacted" || redacted["password"] != "redacted" {
		t.Fatalf("credentials were not redacted: %#v", redacted)
	}
}

func TestSaveWritesConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "config.yaml")
	cfg := Config{
		BaseURL:    "https://tenant.classreach.com",
		OriginHost: "classreach.azurewebsites.net",
		Username:   "guardian",
		Password:   "secret",
	}
	if err := Save(path, cfg); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0600 {
		t.Fatalf("mode = %v", got)
	}
}
