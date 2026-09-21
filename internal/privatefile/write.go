// Package privatefile writes private files without following destination symlinks.
package privatefile

import (
	"fmt"
	"os"
	"path/filepath"
)

// Write publishes a complete 0600 file. Without overwrite, an existing path wins.
func Write(name string, data []byte, overwrite bool) error {
	if err := Check(name, overwrite); err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(name), ".classreach-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err := f.Write(data); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	if overwrite {
		// Rename replaces the directory entry; it never follows a raced symlink.
		return os.Rename(f.Name(), name)
	}
	// Link publishes atomically and fails if another writer creates the path.
	return os.Link(f.Name(), name)
}

// Check rejects destination symlinks, non-files, and unwanted overwrites.
func Check(name string, overwrite bool) error {
	info, err := os.Lstat(name)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("refusing non-regular output path %s", name)
	}
	if !overwrite {
		return fmt.Errorf("output exists at %s; use --force to overwrite", name)
	}
	return nil
}
