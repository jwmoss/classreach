package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNotificationCountsUsesKnownTermContract(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/Notifications/GetNotificationCounts" || r.URL.Query().Get("academicTermID") != "term-1" {
			http.Error(w, "unexpected request", 400)
			return
		}
		w.Write([]byte(`{"syntheticCount":9007199254740993}`))
	}))
	defer server.Close()
	data, err := New(server.URL).GetNotificationCounts(context.Background(), "term-1")
	if err != nil || !strings.Contains(string(data), "9007199254740993") {
		t.Fatalf("data=%s err=%v", data, err)
	}
	if _, err := New(server.URL).GetNotificationCounts(context.Background(), ""); err == nil {
		t.Fatal("missing term accepted")
	}
}
