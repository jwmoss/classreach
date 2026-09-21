package cli

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/jwmoss/classreach/internal/privatefile"
)

type agendaFile struct {
	file *zip.File
	path string
}

const maxAgendaBytes = 128 << 20

func newAgendaCommand(rc *runtime) *cobra.Command {
	cmd := &cobra.Command{Use: "agenda", Short: "Download weekly assignment sheets"}
	cmd.AddCommand(newAgendaDownloadCommand(rc))
	return cmd
}

func newAgendaDownloadCommand(rc *runtime) *cobra.Command {
	week, outputPath, force := defaultWeek(), "", false
	cmd := &cobra.Command{
		Use:   "download",
		Short: "Download the weekly agenda",
		Args:  usageArgs(cobra.NoArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			if outputPath == "" {
				return fmt.Errorf("%w: --output is required", errUsage)
			}
			if _, err := time.Parse(dateLayout, week); err != nil {
				return fmt.Errorf("invalid week date %q: use YYYY-MM-DD", week)
			}
			data, err := rc.client.DownloadAgenda(cmd.Context(), week)
			if err != nil {
				return err
			}
			paths, err := writeAgenda(data, outputPath, force)
			if err != nil {
				return err
			}
			if rc.out.IsJSON() {
				return rc.out.JSON(paths)
			}
			for _, writtenPath := range paths {
				rc.out.Printf("%s\n", writtenPath)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&week, "week", week, "week date in YYYY-MM-DD format")
	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "output directory or .zip file path")
	cmd.Flags().BoolVar(&force, "force", false, "overwrite existing files")
	return cmd
}

func writeAgenda(data []byte, outputPath string, force bool) ([]string, error) {
	if strings.EqualFold(filepath.Ext(outputPath), ".zip") {
		if err := privatefile.Write(outputPath, data, force); err != nil {
			return nil, fmt.Errorf("write agenda %s: %w", outputPath, err)
		}
		return []string{outputPath}, nil
	}
	return extractAgenda(data, outputPath, force)
}

func extractAgenda(data []byte, outputDir string, force bool) ([]string, error) {
	if info, err := os.Lstat(outputDir); err == nil && !info.IsDir() {
		return nil, fmt.Errorf("agenda output must be a directory, not a file or symlink")
	}
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("read agenda ZIP: %w", err)
	}
	files, err := agendaFiles(reader.File, outputDir, force)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(outputDir, 0700); err != nil {
		return nil, fmt.Errorf("create agenda directory %s: %w", outputDir, err)
	}
	paths := make([]string, 0, len(files))
	for _, file := range files {
		if err := extractAgendaFile(file, force); err != nil {
			return nil, err
		}
		paths = append(paths, file.path)
	}
	return paths, nil
}

func agendaFiles(entries []*zip.File, outputDir string, force bool) ([]agendaFile, error) {
	if len(entries) > 1000 {
		return nil, fmt.Errorf("agenda ZIP exceeds the 1000-entry limit")
	}
	files := make([]agendaFile, 0, len(entries))
	seen := map[string]bool{}
	var total uint64
	for _, entry := range entries {
		if entry.FileInfo().IsDir() {
			continue
		}
		if entry.UncompressedSize64 > maxAgendaBytes-total {
			return nil, fmt.Errorf("agenda ZIP exceeds the 128 MiB expanded-size limit")
		}
		total += entry.UncompressedSize64
		cleanName := path.Clean(strings.ReplaceAll(entry.Name, `\`, "/"))
		if cleanName == "." || path.IsAbs(cleanName) ||
			cleanName == ".." || strings.HasPrefix(cleanName, "../") {
			return nil, fmt.Errorf("unsafe ZIP entry %q escapes the output directory", entry.Name)
		}
		outputName := strings.ReplaceAll(cleanName, "/", "-")
		outputPath := filepath.Join(outputDir, outputName)
		if seen[outputPath] {
			return nil, fmt.Errorf("ZIP entries resolve to the same output path %s", outputPath)
		}
		if err := privatefile.Check(outputPath, force); err != nil {
			return nil, err
		}
		seen[outputPath] = true
		files = append(files, agendaFile{file: entry, path: outputPath})
	}
	return files, nil
}

func extractAgendaFile(file agendaFile, force bool) error {
	reader, err := file.file.Open()
	if err != nil {
		return fmt.Errorf("open agenda entry %q: %w", file.file.Name, err)
	}
	data, readErr := io.ReadAll(io.LimitReader(reader, maxAgendaBytes+1))
	closeErr := reader.Close()
	if readErr != nil {
		return fmt.Errorf("read agenda entry %q: %w", file.file.Name, readErr)
	}
	if len(data) > maxAgendaBytes {
		return fmt.Errorf("agenda file exceeds the 128 MiB expanded-size limit")
	}
	if closeErr != nil {
		return fmt.Errorf("close agenda entry %q: %w", file.file.Name, closeErr)
	}
	if err := privatefile.Write(file.path, data, force); err != nil {
		return fmt.Errorf("write agenda file %s: %w", file.path, err)
	}
	return nil
}
