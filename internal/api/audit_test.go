package api

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestLoginRejectsCrossOriginRedirect(t *testing.T) {
	for _, status := range []int{http.StatusTemporaryRedirect, http.StatusPermanentRedirect} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			var received atomic.Int32
			sink := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { received.Add(1); fmt.Fprint(w, "<html>done</html>") }))
			defer sink.Close()
			source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodGet {
					fmt.Fprint(w, testLoginForm)
					return
				}
				http.Redirect(w, r, sink.URL, status)
			}))
			defer source.Close()
			err := New(source.URL).Login(context.Background(), "synthetic-user", "synthetic-password")
			if err == nil || received.Load() != 0 {
				t.Fatalf("err=%v destination requests=%d", err, received.Load())
			}
		})
	}
}

func TestLoginRejectsUnauthenticatedHTML(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			fmt.Fprint(w, testLoginForm)
			return
		}
		fmt.Fprint(w, "<html>Maintenance</html>")
	}))
	defer server.Close()
	if err := New(server.URL).Login(context.Background(), "synthetic-user", "synthetic-password"); err == nil {
		t.Fatal("maintenance page accepted as authenticated")
	}
}

func TestDryRunMakesNoRequests(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { requests.Add(1); fmt.Fprint(w, testLoginForm) }))
	defer server.Close()
	client := New(server.URL, WithDryRun(true))
	if err := client.Login(context.Background(), "synthetic-user", "synthetic-password"); err == nil {
		t.Error("dry-run login must not run")
	}
	if _, err := client.Do(context.Background(), http.MethodGet, "/read", nil, nil); err == nil {
		t.Error("dry-run GET must not run")
	}
	if requests.Load() != 0 {
		t.Fatalf("made %d requests", requests.Load())
	}
}

func TestRedirectPolicy(t *testing.T) {
	origin, _ := http.NewRequest(http.MethodPost, "https://tenant.classreach.com/Login", nil)
	for _, target := range []string{"http://tenant.classreach.com/Login", "https://other.classreach.com/Login", "https://tenant.classreach.com:8443/Login"} {
		req, _ := http.NewRequest(http.MethodPost, target, nil)
		if err := sameOriginRedirect(req, []*http.Request{origin}); err == nil {
			t.Errorf("redirect accepted: %s", target)
		}
	}
	same, _ := http.NewRequest(http.MethodGet, "https://tenant.classreach.com/", nil)
	if err := sameOriginRedirect(same, []*http.Request{origin}); err != nil {
		t.Fatal(err)
	}
	if err := sameOriginRedirect(same, make([]*http.Request, 10)); err == nil {
		t.Fatal("redirect limit not enforced")
	}
}

func TestResponseLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		block := make([]byte, 1<<20)
		for range 65 {
			if _, err := w.Write(block); err != nil {
				return
			}
		}
	}))
	defer server.Close()
	if _, err := New(server.URL).Do(context.Background(), http.MethodGet, "/large", nil, nil); err == nil {
		t.Fatal("oversized response accepted")
	}
}
