package cli

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"strings"
	"testing"
	"time"

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

func TestGlobalEnv(t *testing.T) {
	t.Setenv(config.EnvPrefix+"_TIMEOUT", "45s")
	t.Setenv(config.EnvPrefix+"_DRY_RUN", "true")
	t.Setenv(config.EnvPrefix+"_OUTPUT", "json")

	var stdout, stderr bytes.Buffer
	rc := &runtime{
		ctx:    context.Background(),
		stdin:  strings.NewReader(""),
		stdout: &stdout,
		stderr: &stderr,
		g:      &globals{timeout: 30 * time.Second},
	}
	cmd := newRootCommand(rc)
	cmd.SetArgs([]string{"version"})
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	if err := cmd.ExecuteContext(rc.ctx); err != nil {
		t.Fatal(err)
	}
	if rc.g.timeout != 45*time.Second || !rc.g.dryRun || !rc.g.asJSON {
		t.Fatalf("environment was not applied: %#v", rc.g)
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

func TestInstalledModuleVersion(t *testing.T) {
	const module = "github.com/jwmoss/classreach"
	const moduleVersion = "v1.2.3"
	dir := t.TempDir()
	proxy := filepath.Join(dir, "proxy")
	moduleDir := filepath.Join(proxy, filepath.FromSlash(module), "@v")
	if err := os.MkdirAll(moduleDir, 0700); err != nil {
		t.Fatal(err)
	}
	root := filepath.Join("..", "..")
	mod, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	for ext, data := range map[string][]byte{
		"mod":  mod,
		"info": []byte(`{"Version":"v1.2.3","Time":"2026-01-01T00:00:00Z"}`),
	} {
		if err := os.WriteFile(filepath.Join(moduleDir, moduleVersion+"."+ext), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	var archive bytes.Buffer
	zw := zip.NewWriter(&archive)
	err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if entry.IsDir() {
			if rel == "." || rel == "cmd" || rel == "internal" || strings.HasPrefix(rel, "cmd/") || strings.HasPrefix(rel, "internal/") {
				return nil
			}
			return filepath.SkipDir
		}
		if rel != "go.mod" && rel != "go.sum" && (!strings.HasSuffix(rel, ".go") || strings.HasSuffix(rel, "_test.go")) {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		file, err := zw.Create(module + "@" + moduleVersion + "/" + rel)
		if err != nil {
			return err
		}
		_, err = file.Write(data)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	// The installed fixture replaces only the HTTP boundary. It needs no TCP listener.
	file, err := zw.Create(module + "@" + moduleVersion + "/cmd/classreach/transport_fixture.go")
	if err != nil {
		t.Fatal(err)
	}
	_, err = file.Write([]byte(`package main
import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)
type fixtureTransport struct{}
func (fixtureTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	fmt.Fprintln(os.Stderr, "captured user-agent:", req.UserAgent())
	return &http.Response{StatusCode: http.StatusUnauthorized, Header: make(http.Header), Body: io.NopCloser(strings.NewReader("denied")), Request: req}, nil
}
func init() { http.DefaultTransport = fixtureTransport{} }
`))
	if err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(moduleDir, moduleVersion+".zip"), archive.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	cache, err := exec.Command("go", "env", "GOMODCACHE").Output()
	if err != nil {
		t.Fatal(err)
	}
	// File proxies use the synthetic module and cached dependencies. No external API is available.
	fileURL := func(path string) string {
		path = filepath.ToSlash(path)
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}
		return (&url.URL{Scheme: "file", Path: path}).String()
	}
	for _, test := range []struct {
		name, ldflags, version, commit, date string
	}{
		{"module", "", moduleVersion, "unknown", "unknown"},
		{"release", "-X " + module + "/internal/cli.version=v9.8.7 -X " + module + "/internal/cli.commit=release-commit -X " + module + "/internal/cli.date=release-date", "v9.8.7", "release-commit", "release-date"},
	} {
		t.Run(test.name, func(t *testing.T) {
			binDir := filepath.Join(dir, test.name)
			t.Setenv("GOBIN", binDir)
			t.Setenv("GOMODCACHE", filepath.Join(dir, "modcache"))
			t.Setenv("GOPROXY", fileURL(proxy)+","+fileURL(filepath.Join(strings.TrimSpace(string(cache)), "cache", "download")))
			for key, value := range map[string]string{"GOSUMDB": "off", "GOTOOLCHAIN": "local", "GOWORK": "off", "GOFLAGS": "-modcacherw", "GOPRIVATE": "", "GONOPROXY": "", "GONOSUMDB": "", "GOVCS": "*:off", "GOOS": goruntime.GOOS, "GOARCH": goruntime.GOARCH} {
				t.Setenv(key, value)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			defer cancel()
			install := exec.CommandContext(ctx, "go", "install", "-ldflags="+test.ldflags, module+"/cmd/classreach@"+moduleVersion)
			install.Dir = dir
			if output, err := install.CombinedOutput(); err != nil {
				t.Fatalf("install: %v\n%s", err, output)
			}
			binary := filepath.Join(binDir, "classreach")
			if goruntime.GOOS == "windows" {
				binary += ".exe"
			}
			run := func(args ...string) string {
				t.Helper()
				var stdout, stderr bytes.Buffer
				cmd := exec.CommandContext(ctx, binary, args...)
				cmd.Stdout, cmd.Stderr = &stdout, &stderr
				if err := cmd.Run(); err != nil || stderr.Len() != 0 {
					t.Fatalf("%v: err=%v stderr=%s", args, err, &stderr)
				}
				return stdout.String()
			}
			var got map[string]string
			output := run("--json", "version")
			if err := json.Unmarshal([]byte(output), &got); err != nil || len(got) != 3 || got["version"] != test.version || got["commit"] != test.commit || got["date"] != test.date {
				t.Errorf("version JSON = %s; error = %v", output, err)
			}
			for _, check := range []struct {
				args []string
				want string
			}{
				{[]string{"version"}, fmt.Sprintf("classreach version %s\ncommit: %s\nbuilt:  %s\n", test.version, test.commit, test.date)},
				{[]string{"--plain", "version"}, test.version + "\n"},
				{[]string{"--version"}, "classreach version " + test.version + "\n"},
				{[]string{"--version", "overview"}, "classreach version " + test.version + "\n"},
			} {
				if got := run(check.args...); got != check.want {
					t.Errorf("%v output = %q; want %q", check.args, got, check.want)
				}
			}
			t.Setenv("CLASSREACH_USERNAME", "synthetic-user")
			t.Setenv("CLASSREACH_PASSWORD", "synthetic-password")
			cmd := exec.CommandContext(ctx, binary, "--config", filepath.Join(dir, "absent.yaml"), "--base-url", "http://127.0.0.1:1", "doctor")
			outputBytes, err := cmd.CombinedOutput()
			exit, ok := err.(*exec.ExitError)
			if !ok || exit.ExitCode() != exitErr || !strings.Contains(string(outputBytes), "captured user-agent: classreach/"+test.version+"\n") {
				t.Errorf("mock HTTP request: error=%v output=%s", err, outputBytes)
			}
		})
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
			http.SetCookie(w, &http.Cookie{Name: ".AspNet.SharedCookie", Value: "session", Path: "/"})
			http.Redirect(w, r, "/", http.StatusFound)
		case r.URL.Path == "/Home/GetQuickView":
			_, _ = w.Write([]byte(`{"UserInfos":[],"Announcements":[]}`))
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
	}, false); err != nil {
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
