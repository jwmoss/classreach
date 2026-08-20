package api

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

const testLoginForm = `<form method="post">` +
	`<input value="test-token" name="__RequestVerificationToken">` +
	`<input name="Username"><input name="Password" type="password"></form>`

func TestLogin(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(loginTestHandler))
	defer server.Close()

	client := New(server.URL)
	if err := client.Login(context.Background(), "guardian", "secret"); err != nil {
		t.Fatal(err)
	}
	body := map[string]string{"read": "only"}
	if _, err := client.Do(context.Background(), http.MethodPost, "/read", nil, body); err != nil {
		t.Fatal(err)
	}
}

func loginTestHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/Login" {
		handleLoginTestRequest(w, r)
		return
	}
	switch r.URL.Path {
	case "/":
		logoutForm := `<form><input name="__RequestVerificationToken" ` +
			`value="logout-token"></form>`
		_, _ = fmt.Fprint(w, logoutForm)
	case "/read":
		if r.Header.Get("__RequestVerificationToken") != "test-token" {
			http.Error(w, "missing anti-forgery token", http.StatusBadRequest)
			return
		}
		_, _ = fmt.Fprint(w, `{}`)
	default:
		http.NotFound(w, r)
	}
}

func handleLoginTestRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		http.SetCookie(w, &http.Cookie{Name: "antiforgery", Value: "cookie-token"})
		_, _ = fmt.Fprint(w, testLoginForm)
		return
	}
	if !validLoginRequest(r) {
		http.Error(w, "bad login request", http.StatusBadRequest)
		return
	}
	http.SetCookie(w, &http.Cookie{Name: "auth", Value: "session"})
	http.Redirect(w, r, "/", http.StatusFound)
}

func validLoginRequest(r *http.Request) bool {
	_ = r.ParseForm()
	_, cookieErr := r.Cookie("antiforgery")
	return cookieErr == nil &&
		r.Form.Get("__RequestVerificationToken") == "test-token" &&
		r.Form.Get("Username") == "guardian" &&
		r.Form.Get("Password") == "secret"
}

func TestLoginRejectsInvalidCredentials(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, testLoginForm)
	}))
	defer server.Close()

	client := New(server.URL)
	err := client.Login(context.Background(), "guardian", "wrong")
	if err == nil {
		t.Fatal("expected login error")
	}
}

func TestAntiForgeryTokenRequiresToken(t *testing.T) {
	if _, err := antiForgeryToken([]byte(`<form></form>`)); err == nil {
		t.Fatal("expected missing token error")
	}
}
