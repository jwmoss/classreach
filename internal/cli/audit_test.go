package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
)

func TestMain(m *testing.M) {
	// Tests must never inherit a real tenant or credentials from the caller.
	for _, key := range []string{"USERNAME", "PASSWORD", "BASE_URL", "ORIGIN_HOST", "DRY_RUN", "OUTPUT", "TIMEOUT"} {
		_ = os.Unsetenv("CLASSREACH_" + key)
	}
	os.Exit(m.Run())
}

func auditExecute(args ...string) (int, string, string) {
	var out, err bytes.Buffer
	code := Execute(context.Background(), args, strings.NewReader(""), &out, &err)
	return code, out.String(), err.String()
}

func auditServer(t *testing.T) (string, *atomic.Int32) {
	t.Helper()
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		switch {
		case r.URL.Path == "/Login" && r.Method == http.MethodGet:
			fmt.Fprint(w, `<input name="__RequestVerificationToken" value="token"><input name="Username"><input name="Password">`)
		case r.URL.Path == "/Login" && r.Method == http.MethodPost:
			http.SetCookie(w, &http.Cookie{Name: ".AspNet.SharedCookie", Value: "synthetic-session", Path: "/"})
			fmt.Fprint(w, `<input name="__RequestVerificationToken" value="authenticated-token">`)
		case r.URL.Path == "/Home/GetQuickView":
			fmt.Fprint(w, `{"UserInfos":[],"Announcements":[]}`)
		case r.URL.Path == "/SchoolDocuments":
			fmt.Fprint(w, `{"SchoolDocumentsListItems":[{"ID":"doc","FileInfo":{"DownloadUrl":"/file"}}]}`)
		case r.URL.Path == "/Messages/GetThreadMessages":
			fmt.Fprint(w, `{"MessageThreadViewModel":{"Messages":[{"Files":[{"ID":"file","Url":"/file"}]}]}}`)
		case r.URL.Path == "/file":
			fmt.Fprint(w, "synthetic file")
		case r.URL.Path == "/numbers":
			fmt.Fprint(w, `{"id":9007199254740993}`)
		default:
			fmt.Fprint(w, "<html>Maintenance</html>")
		}
	}))
	t.Cleanup(server.Close)
	cfg := filepath.Join(t.TempDir(), "config.yaml")
	content := fmt.Sprintf("base_url: %s\norigin_host: localhost\nusername: synthetic-user\npassword: synthetic-password\n", server.URL)
	if err := os.WriteFile(cfg, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"USERNAME", "PASSWORD", "BASE_URL", "ORIGIN_HOST", "DRY_RUN", "OUTPUT", "TIMEOUT"} {
		t.Setenv("CLASSREACH_"+key, "")
	}
	return cfg, &requests
}

func TestInvalidInvocationDoesNotAuthenticate(t *testing.T) {
	cfg, requests := auditServer(t)
	for _, args := range [][]string{{"assignments", "list"}, {"attendance", "list", "--student", "student"}, {"overview", "--week", "bad"}, {"calendar", "list", "--start", "2026-09-20", "--end", "2026-09-01"}, {"documents", "download", "doc"}, {"messages", "list", "--page", "0"}, {"raw", "get", "/", "--query", "bad"}, {"overview", "extra"}, {"--timeout", "0s", "overview"}} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			requests.Store(0)
			code, _, stderr := auditExecute(append([]string{"--config", cfg}, args...)...)
			if code != exitUsage || requests.Load() != 0 {
				t.Fatalf("code=%d requests=%d stderr=%s", code, requests.Load(), stderr)
			}
		})
	}
}

func TestDryRunPrintsPlanWithoutNetworkOrWrites(t *testing.T) {
	cfg, requests := auditServer(t)
	output := filepath.Join(t.TempDir(), "download")
	for _, args := range [][]string{{"messages", "list"}, {"documents", "download", "doc", "--output", output}, {"config", "init", "--force"}} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			requests.Store(0)
			before, _ := os.ReadFile(cfg)
			code, stdout, stderr := auditExecute(append([]string{"--config", cfg, "--json", "--dry-run"}, args...)...)
			var got map[string]any
			if code != exitOK || json.Unmarshal([]byte(stdout), &got) != nil || got["dry_run"] != true || requests.Load() != 0 {
				t.Fatalf("code=%d requests=%d stdout=%s stderr=%s", code, requests.Load(), stdout, stderr)
			}
			after, _ := os.ReadFile(cfg)
			if !bytes.Equal(before, after) {
				t.Error("config changed")
			}
			if _, err := os.Stat(output); !os.IsNotExist(err) {
				t.Error("download created")
			}
		})
	}
}

func TestVersionFlagWithSubcommandDoesNotPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("version flag panic: %v", r)
		}
	}()
	code, stdout, stderr := auditExecute("--version", "overview")
	if code != exitOK || !strings.Contains(stdout, "classreach") {
		t.Fatalf("code=%d stdout=%s stderr=%s", code, stdout, stderr)
	}
}

func TestOutputAndUsageContracts(t *testing.T) {
	for _, args := range [][]string{{"--unknown"}, {"unknown"}, {"--json", "--plain", "version"}} {
		code, _, stderr := auditExecute(args...)
		if code != exitUsage {
			t.Errorf("args=%v code=%d stderr=%s", args, code, stderr)
		}
	}
}

func TestConfigShowUsesEffectivePathAndOrigin(t *testing.T) {
	cfg, _ := auditServer(t)
	code, stdout, stderr := auditExecute("--config", cfg, "--origin-host", "override.example", "--json", "config", "show")
	var got map[string]string
	if code != exitOK || json.Unmarshal([]byte(stdout), &got) != nil || got["path"] != cfg || got["origin_host"] != "override.example" {
		t.Fatalf("code=%d stdout=%s stderr=%s", code, stdout, stderr)
	}
}

func TestDownloadsEmitJSON(t *testing.T) {
	cfg, _ := auditServer(t)
	for _, args := range [][]string{{"documents", "download", "doc"}, {"messages", "download", "thread", "file"}} {
		out := filepath.Join(t.TempDir(), "file")
		args = append(args, "--output", out)
		code, stdout, stderr := auditExecute(append([]string{"--config", cfg, "--json"}, args...)...)
		var got map[string]any
		if code != exitOK || json.Unmarshal([]byte(stdout), &got) != nil || got["path"] != out {
			t.Errorf("code=%d stdout=%s stderr=%s", code, stdout, stderr)
		}
	}
}

func TestRawJSONPreservesNumbersAndRejectsHTML(t *testing.T) {
	cfg, _ := auditServer(t)
	code, stdout, stderr := auditExecute("--config", cfg, "--json", "raw", "get", "/numbers")
	if code != exitOK || !strings.Contains(stdout, "9007199254740993") {
		t.Errorf("code=%d stdout=%s stderr=%s", code, stdout, stderr)
	}
	code, stdout, _ = auditExecute("--config", cfg, "--json", "raw", "get", "/maintenance")
	if code == exitOK || stdout != "" {
		t.Errorf("non-JSON accepted: code=%d stdout=%s", code, stdout)
	}
}

func TestDoctorRejectsAuthenticatedMaintenancePage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/Login" && r.Method == http.MethodGet {
			fmt.Fprint(w, `<input name="__RequestVerificationToken" value="token"><input name="Username"><input name="Password">`)
			return
		}
		if r.URL.Path == "/Login" {
			http.SetCookie(w, &http.Cookie{Name: ".AspNet.SharedCookie", Value: "synthetic-session", Path: "/"})
		}
		fmt.Fprint(w, "<html>Maintenance</html>")
	}))
	defer server.Close()
	cfg := filepath.Join(t.TempDir(), "config.yaml")
	data := fmt.Sprintf("base_url: %s\norigin_host: localhost\nusername: synthetic-user\npassword: synthetic-password\n", server.URL)
	if err := os.WriteFile(cfg, []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
	code, stdout, _ := auditExecute("--config", cfg, "--json", "doctor")
	if code == exitOK || stdout != "" {
		t.Fatalf("code=%d stdout=%s", code, stdout)
	}
}

func TestConfigInitJSONPreservesPasswordSpaces(t *testing.T) {
	cfg := filepath.Join(t.TempDir(), "config.yaml")
	var stdout, stderr bytes.Buffer
	code := Execute(context.Background(), []string{"--config", cfg, "--json", "config", "init", "--username", "synthetic-user", "--password-stdin"}, strings.NewReader(" synthetic-password \n"), &stdout, &stderr)
	if code != exitOK || !json.Valid(stdout.Bytes()) {
		t.Fatalf("code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	data, err := os.ReadFile(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(data, []byte(" synthetic-password ")) {
		t.Fatal("password spaces changed")
	}
}
