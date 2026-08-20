package cli

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jwmoss/classreach/internal/config"
)

func TestVersionJSON(t *testing.T) {
	var stdout, stderr bytes.Buffer
	args := []string{"--json", "version"}
	code := Execute(context.Background(), args, strings.NewReader(""), &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("code = %d stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), `"version": "dev"`) {
		t.Fatalf("stdout = %s", stdout.String())
	}
}

func TestVersionFlag(t *testing.T) {
	var stdout, stderr bytes.Buffer
	args := []string{"--version"}
	code := Execute(context.Background(), args, strings.NewReader(""), &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("code = %d stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "version dev") {
		t.Fatalf("stdout = %s", stdout.String())
	}
}

func TestDoctor(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/Login" && r.Method == http.MethodGet:
			_, _ = fmt.Fprint(
				w,
				`<input name="__RequestVerificationToken" value="test-token">`+
					`<input name="Username"><input name="Password">`,
			)
		case r.URL.Path == "/Login" && r.Method == http.MethodPost:
			http.Redirect(w, r, "/", http.StatusFound)
		default:
			_, _ = w.Write([]byte(`{"status":"ok"}`))
		}
	}))
	defer server.Close()
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	if err := config.Save(configPath, config.Config{
		BaseURL:    server.URL,
		OriginHost: config.DefaultOriginHost,
		Username:   "guardian",
		Password:   "secret",
	}); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	args := []string{"--config", configPath, "--json", "doctor"}
	code := Execute(context.Background(), args, strings.NewReader(""), &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("code = %d stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), `"ok": true`) {
		t.Fatalf("stdout = %s", stdout.String())
	}
}

func TestRawUsage(t *testing.T) {
	var stdout, stderr bytes.Buffer
	args := []string{"raw", "get"}
	code := Execute(context.Background(), args, strings.NewReader(""), &stdout, &stderr)
	if code != exitUsage {
		t.Fatalf("code = %d stderr = %s", code, stderr.String())
	}
}
