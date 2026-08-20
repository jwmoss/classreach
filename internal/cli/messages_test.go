package cli

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spf13/cobra"

	"github.com/jwmoss/classreach/internal/api"
)

func TestFindMessageFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `{"MessageThreadViewModel":{"Messages":[`+
			`{"Files":[{"ID":"file-1","Name":"Notice",`+
			`"FileDownloadLink":"/file"}]}]}}`)
	}))
	defer server.Close()

	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())
	rc := &runtime{client: api.New(server.URL)}
	file, err := findMessageFile(cmd, rc, "thread-1", "file-1")
	if err != nil {
		t.Fatal(err)
	}
	if file.DownloadURL() != "/file" {
		t.Fatalf("file = %#v", file)
	}
}
