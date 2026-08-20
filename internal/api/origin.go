package api

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"strings"
)

func WithOriginHost(originHost string) Option {
	return func(c *Client) {
		baseURL, err := url.Parse(c.baseURL)
		if err != nil || !isClassReachHost(baseURL.Hostname()) || strings.TrimSpace(originHost) == "" {
			return
		}
		transport := http.DefaultTransport.(*http.Transport).Clone()
		dialer := &net.Dialer{}
		transport.DialContext = func(ctx context.Context, network, address string) (net.Conn, error) {
			target := originAddress(address, baseURL.Hostname(), strings.TrimSpace(originHost))
			return dialer.DialContext(ctx, network, target)
		}
		c.httpClient.Transport = transport
	}
}

func isClassReachHost(host string) bool {
	return strings.EqualFold(host, "classreach.com") ||
		strings.HasSuffix(strings.ToLower(host), ".classreach.com")
}

func originAddress(address, tenantHost, originHost string) string {
	host, port, err := net.SplitHostPort(address)
	if err != nil || !strings.EqualFold(host, tenantHost) {
		return address
	}
	return net.JoinHostPort(originHost, port)
}
