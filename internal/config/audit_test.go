package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestSaveReplacesPublicFilePrivately(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix permission bits")
	}
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte("old"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0644); err != nil {
		t.Fatal(err)
	}
	if err := Save(path, Config{Password: "synthetic-password"}, true); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0600 {
		t.Fatalf("mode=%o", info.Mode().Perm())
	}
}

func TestSaveRejectsSymlink(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "target")
	link := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(target, []byte("unchanged"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	if err := Save(link, Config{Password: "synthetic-password"}, true); err == nil {
		t.Error("symlink accepted")
	}
	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "unchanged" {
		t.Fatal("symlink target changed")
	}
}

func TestValidateRejectsUnsafeURL(t *testing.T) {
	for _, base := range []string{"http://school.classreach.com", "https://user:password@school.classreach.com", "https://school.classreach.com?secret=1", "https://school.classreach.com/#fragment", "not a URL"} {
		cfg := Config{BaseURL: base, OriginHost: "origin", Username: "user", Password: "synthetic-password"}
		if err := cfg.Validate(); err == nil {
			t.Errorf("accepted unsafe URL")
		}
	}
}
