package cli

import (
	"os"

	"github.com/jwmoss/classreach/internal/privatefile"
)

func fileExists(path string) bool {
	_, err := os.Lstat(path)
	return err == nil
}

func writeDownload(rc *runtime, path string, data []byte, force bool) error {
	if err := privatefile.Write(path, data, force); err != nil {
		return err
	}
	if rc.out.IsJSON() {
		return rc.out.JSON(map[string]any{"path": path, "bytes": len(data)})
	}
	rc.out.Printf("%s\n", path)
	return nil
}
