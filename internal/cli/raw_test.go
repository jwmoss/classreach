package cli

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jwmoss/classreach/internal/api"
	"github.com/jwmoss/classreach/internal/output"
)

func TestRawGetPreservesBinaryResponse(t *testing.T) {
	want := []byte{'P', 'K', 0x03, 0x04, 0xff}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(want)
	}))
	defer server.Close()

	var stdout bytes.Buffer
	rc := &runtime{client: api.New(server.URL), stdout: &stdout, stderr: &bytes.Buffer{}}
	rc.out = output.New(rc.stdout, rc.stderr, false, false, false, true)
	cmd := newRawGetCommand(rc)
	cmd.SetArgs([]string{"/archive"})
	if err := cmd.ExecuteContext(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(stdout.Bytes(), want) {
		t.Fatalf("raw response = %v, want %v", stdout.Bytes(), want)
	}
}
