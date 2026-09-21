package privatefile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWritePreservesExistingFileUnlessForced(t *testing.T) {
	path := filepath.Join(t.TempDir(), "private")
	if err := Write(path, []byte("old"), false); err != nil {
		t.Fatal(err)
	}
	if err := Write(path, []byte("new"), false); err == nil {
		t.Fatal("overwrote existing file")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "old" {
		t.Fatal("old file changed")
	}
	if err := Write(path, []byte("new"), true); err != nil {
		t.Fatal(err)
	}
	data, err = os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "new" {
		t.Fatal("replacement incomplete")
	}
	matches, err := filepath.Glob(filepath.Join(filepath.Dir(path), ".classreach-*"))
	if err != nil || len(matches) != 0 {
		t.Fatalf("temporary files: %v error: %v", matches, err)
	}
}

func TestWriteRejectsDirectory(t *testing.T) {
	dir := t.TempDir()
	if err := Write(dir, []byte("data"), true); err == nil {
		t.Fatal("directory replaced")
	}
}
