package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"
)

const (
	DefaultBaseURL   = "https://providencewilmington.classreach.com"
	DefaultUserAgent = "classreach/dev"
)

type Client struct {
	baseURL          string
	antiForgeryToken string
	httpClient       *http.Client
	userAgent        string
	trace            func(method, path string, status int, duration time.Duration)
	dryRun           bool
}

type Option func(*Client)

func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		if httpClient != nil {
			c.httpClient = httpClient
		}
	}
}

func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		if timeout > 0 {
			c.httpClient.Timeout = timeout
		}
	}
}

func WithUserAgent(userAgent string) Option {
	return func(c *Client) {
		if strings.TrimSpace(userAgent) != "" {
			c.userAgent = userAgent
		}
	}
}

func WithTrace(trace func(method, path string, status int, duration time.Duration)) Option {
	return func(c *Client) {
		c.trace = trace
	}
}

func WithDryRun(dryRun bool) Option {
	return func(c *Client) {
		c.dryRun = dryRun
	}
}

func New(baseURL string, opts ...Option) *Client {
	jar, _ := cookiejar.New(nil)
	c := &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Jar:     jar,
			Timeout: 30 * time.Second,
		},
		userAgent: DefaultUserAgent,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

type APIError struct {
	Status  int
	Method  string
	Path    string
	Body    []byte
	Message string
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("%s %s: HTTP %d: %s", e.Method, e.Path, e.Status, e.Message)
	}
	return fmt.Sprintf("%s %s: HTTP %d", e.Method, e.Path, e.Status)
}

func (c *Client) Do(
	ctx context.Context,
	method, requestPath string,
	query url.Values,
	body any,
) ([]byte, error) {
	method = strings.ToUpper(strings.TrimSpace(method))
	if err := c.validateMethod(method, requestPath); err != nil {
		return nil, err
	}
	endpoint, err := c.url(requestPath, query)
	if err != nil {
		return nil, err
	}
	req, err := c.newAPIRequest(ctx, method, endpoint, body)
	if err != nil {
		return nil, err
	}
	start := time.Now()
	data, status, err := c.send(req)
	if err != nil {
		return nil, err
	}
	if c.trace != nil {
		c.trace(method, req.URL.Path, status, time.Since(start))
	}
	if status < 200 || status >= 300 {
		return data, &APIError{
			Status: status, Method: method, Path: req.URL.Path, Body: data,
			Message: extractErrorMessage(data),
		}
	}
	return data, nil
}

func (c *Client) validateMethod(method, requestPath string) error {
	if method == "" {
		return fmt.Errorf("method is required")
	}
	if c.dryRun && method != http.MethodGet {
		return fmt.Errorf("dry-run: refusing %s %s", method, requestPath)
	}
	return nil
}

func (c *Client) newAPIRequest(
	ctx context.Context,
	method, endpoint string,
	body any,
) (*http.Request, error) {
	reader, err := jsonReader(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, reader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("__RequestVerificationToken", c.antiForgeryToken)
	}
	return req, nil
}

func jsonReader(body any) (io.Reader, error) {
	if body == nil {
		return nil, nil
	}
	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("encode request body: %w", err)
	}
	return bytes.NewReader(data), nil
}

func (c *Client) send(req *http.Request) ([]byte, int, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("read response: %w", err)
	}
	return data, resp.StatusCode, nil
}

type JSONRequest struct {
	Body   any
	Method string
	Path   string
	Query  url.Values
}

func (c *Client) DoJSON(
	ctx context.Context,
	request JSONRequest,
	out any,
) error {
	data, err := c.Do(ctx, request.Method, request.Path, request.Query, request.Body)
	if err != nil {
		return err
	}
	if out == nil || len(bytes.TrimSpace(data)) == 0 {
		return nil
	}
	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("decode response JSON: %w", err)
	}
	return nil
}

func (c *Client) url(requestPath string, query url.Values) (string, error) {
	if strings.HasPrefix(requestPath, "http://") || strings.HasPrefix(requestPath, "https://") {
		u, err := url.Parse(requestPath)
		if err != nil {
			return "", fmt.Errorf("parse URL: %w", err)
		}
		if len(query) > 0 {
			u.RawQuery = mergeQuery(u.Query(), query).Encode()
		}
		return u.String(), nil
	}
	if c.baseURL == "" {
		return "", fmt.Errorf("base URL is required")
	}
	joined := c.baseURL + "/" + strings.TrimLeft(requestPath, "/")
	u, err := url.Parse(joined)
	if err != nil {
		return "", fmt.Errorf("parse URL: %w", err)
	}
	if len(query) > 0 {
		u.RawQuery = mergeQuery(u.Query(), query).Encode()
	}
	return u.String(), nil
}

func mergeQuery(dst, src url.Values) url.Values {
	for key, values := range src {
		for _, value := range values {
			dst.Add(key, value)
		}
	}
	return dst
}

func extractErrorMessage(data []byte) string {
	var generic map[string]any
	if err := json.Unmarshal(data, &generic); err != nil {
		return ""
	}
	for _, key := range []string{"message", "error", "detail"} {
		if value, ok := generic[key].(string); ok {
			return limitErrorMessage(value)
		}
	}
	if errorsValue, ok := generic["errors"]; ok {
		return limitErrorMessage(fmt.Sprint(errorsValue))
	}
	return ""
}

func limitErrorMessage(message string) string {
	message = strings.Join(strings.Fields(message), " ")
	const maxLength = 300
	if len(message) <= maxLength {
		return message
	}
	return message[:maxLength] + "..."
}
