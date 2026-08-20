package cli

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jwmoss/classreach/internal/api"
	"github.com/jwmoss/classreach/internal/output"
)

func TestAgendaDownloadExtractsFiles(t *testing.T) {
	archive := testZIP(t, map[string]string{
		"A/B - sheet.pdf": "assignment",
		"notice.pdf":      "notice",
	})
	server := agendaTestServer(t, archive)
	defer server.Close()

	outputDir := filepath.Join(t.TempDir(), "agenda")
	var stdout bytes.Buffer
	rc := &runtime{
		ctx: context.Background(), stdout: &stdout, stderr: &bytes.Buffer{},
		g: &globals{}, client: api.New(server.URL),
	}
	rc.out = output.New(rc.stdout, rc.stderr, false, true, false, true)
	cmd := newAgendaDownloadCommand(rc)
	cmd.SetArgs([]string{"--week", "2026-08-17", "--output", outputDir})
	if err := cmd.ExecuteContext(context.Background()); err != nil {
		t.Fatal(err)
	}

	wantPath := filepath.Join(outputDir, "A-B - sheet.pdf")
	got, err := os.ReadFile(wantPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "assignment" {
		t.Fatalf("file contents = %q", got)
	}
	if !strings.Contains(stdout.String(), wantPath+"\n") {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestWriteAgendaArchivePreservesBytes(t *testing.T) {
	want := []byte{'P', 'K', 0x03, 0x04, 0xff}
	outputPath := filepath.Join(t.TempDir(), "agenda.zip")
	paths, err := writeAgenda(want, outputPath, false)
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) || len(paths) != 1 || paths[0] != outputPath {
		t.Fatalf("archive = %v, paths = %v", got, paths)
	}
}

func TestExtractAgendaRejectsPathEscape(t *testing.T) {
	archive := testZIP(t, map[string]string{"../evil.pdf": "nope"})
	_, err := extractAgenda(archive, filepath.Join(t.TempDir(), "agenda"), false)
	if err == nil || !strings.Contains(err.Error(), "unsafe ZIP entry") {
		t.Fatalf("error = %v", err)
	}
}

func agendaTestServer(t *testing.T, archive []byte) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/Home/GetQuickView":
			_, _ = fmt.Fprint(w, `{
				"DownloadAgendaForWeekUrl":"/Agenda/DownloadAgendaForWeek?weekDate=2026-08-17"
			}`)
		case "/Agenda/DownloadAgendaForWeek":
			_, _ = w.Write(archive)
		default:
			http.NotFound(w, r)
		}
	}))
}

func testZIP(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var data bytes.Buffer
	writer := zip.NewWriter(&data)
	for name, contents := range files {
		entry, err := writer.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := entry.Write([]byte(contents)); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return data.Bytes()
}
