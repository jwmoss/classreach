package api

import (
	"context"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

var (
	inputPattern     = regexp.MustCompile(`(?i)<input\b[^>]*>`)
	attributePattern = regexp.MustCompile(`(?i)([a-z][a-z0-9_-]*)\s*=\s*["']([^"']*)["']`)
)

func (c *Client) Login(ctx context.Context, username, password string) error {
	if strings.TrimSpace(username) == "" {
		return fmt.Errorf("username is required")
	}
	if password == "" {
		return fmt.Errorf("password is required")
	}

	loginURL, err := c.url("/Login", url.Values{"ReturnUrl": {"/"}})
	if err != nil {
		return err
	}
	page, err := c.sendLoginRequest(ctx, http.MethodGet, loginURL, nil)
	if err != nil {
		return fmt.Errorf("load login page: %w", err)
	}
	token, err := antiForgeryToken(page)
	if err != nil {
		return err
	}
	c.antiForgeryToken = token

	form := url.Values{
		"Username":                   {username},
		"Password":                   {password},
		"__RequestVerificationToken": {token},
	}
	response, err := c.sendLoginRequest(
		ctx,
		http.MethodPost,
		loginURL,
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return fmt.Errorf("submit login: %w", err)
	}
	if loginFormPresent(response) {
		return fmt.Errorf("login failed: check the configured username and password")
	}
	return nil
}

func (c *Client) sendLoginRequest(
	ctx context.Context,
	method string,
	endpoint string,
	body io.Reader,
) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return nil, fmt.Errorf("create login request: %w", err)
	}
	req.Header.Set("Accept", "text/html")
	req.Header.Set("User-Agent", c.userAgent)
	if body != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("login request failed: %w", err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read login response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return data, nil
}

func antiForgeryToken(data []byte) (string, error) {
	for _, input := range inputPattern.FindAll(data, -1) {
		attributes := map[string]string{}
		for _, match := range attributePattern.FindAllSubmatch(input, -1) {
			attributes[strings.ToLower(string(match[1]))] = html.UnescapeString(string(match[2]))
		}
		if attributes["name"] == "__RequestVerificationToken" && attributes["value"] != "" {
			return attributes["value"], nil
		}
	}
	return "", fmt.Errorf("login page did not contain an anti-forgery token")
}

func loginFormPresent(data []byte) bool {
	inputNames := map[string]bool{}
	for _, input := range inputPattern.FindAll(data, -1) {
		for _, match := range attributePattern.FindAllSubmatch(input, -1) {
			if strings.EqualFold(string(match[1]), "name") {
				inputNames[strings.ToLower(string(match[2]))] = true
			}
		}
	}
	return inputNames["username"] && inputNames["password"]
}
