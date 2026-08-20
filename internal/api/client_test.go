package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestDoAddsQueryAndDecodesJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("page"); got != "1" {
			t.Fatalf("query page = %q", got)
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"ok": "true"})
	}))
	defer server.Close()

	client := New(server.URL)
	query := url.Values{"page": []string{"1"}}
	var out map[string]string
	request := JSONRequest{Method: http.MethodGet, Path: "/test", Query: query}
	err := client.DoJSON(context.Background(), request, &out)
	if err != nil {
		t.Fatal(err)
	}
	if out["ok"] != "true" {
		t.Fatalf("decoded output = %#v", out)
	}
}

func TestAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"message":"bad token"}`))
	}))
	defer server.Close()

	client := New(server.URL)
	_, err := client.Do(context.Background(), http.MethodGet, "/private", nil, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("error type = %T", err)
	}
	if apiErr.Status != http.StatusUnauthorized || !strings.Contains(apiErr.Error(), "bad token") {
		t.Fatalf("api error = %v", apiErr)
	}
}

func TestAPIErrorDoesNotPrintHTML(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_, _ = w.Write([]byte(`<html>private student data</html>`))
	}))
	defer server.Close()

	_, err := New(server.URL).Do(context.Background(), http.MethodGet, "/private", nil, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if strings.Contains(err.Error(), "private student data") {
		t.Fatalf("error exposed response body: %v", err)
	}
}

func TestDryRunBlocksMutations(t *testing.T) {
	client := New("https://example.invalid", WithDryRun(true), WithTimeout(time.Millisecond))
	body := map[string]string{"x": "y"}
	if _, err := client.Do(context.Background(), http.MethodPost, "/mutate", nil, body); err == nil {
		t.Fatal("expected dry-run error")
	}
}
