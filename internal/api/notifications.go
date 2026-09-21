package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// GetNotificationCounts preserves the provider's JSON without assuming count fields.
func (c *Client) GetNotificationCounts(ctx context.Context, termID string) (json.RawMessage, error) {
	if strings.TrimSpace(termID) == "" {
		return nil, fmt.Errorf("academic term ID is required")
	}
	data, err := c.Do(ctx, http.MethodGet, "/Notifications/GetNotificationCounts",
		url.Values{"academicTermID": {termID}}, nil)
	if err != nil {
		return nil, err
	}
	if !json.Valid(data) {
		return nil, fmt.Errorf("notification counts response is not JSON")
	}
	return json.RawMessage(data), nil
}
