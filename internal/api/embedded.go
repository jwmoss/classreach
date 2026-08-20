package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

func (c *Client) getEmbeddedJSON(
	ctx context.Context,
	path, marker string,
	destination any,
) error {
	body, err := c.Do(ctx, http.MethodGet, path, nil, nil)
	if err != nil {
		return err
	}
	encoded, err := findEmbeddedJSON(body, marker)
	if err != nil {
		return fmt.Errorf("read %s model: %w", path, err)
	}
	var decoded string
	encoded = strings.ReplaceAll(encoded, `\'`, `'`)
	if err := json.Unmarshal([]byte(`"`+encoded+`"`), &decoded); err != nil {
		return fmt.Errorf("decode %s JavaScript string: %w", marker, err)
	}
	if err := json.Unmarshal([]byte(decoded), destination); err != nil {
		return fmt.Errorf("decode %s JSON: %w", marker, err)
	}
	return nil
}

func findEmbeddedJSON(body []byte, marker string) (string, error) {
	pattern := `(?s)` + regexp.QuoteMeta(marker) +
		`\s*=\s*[^;]*?JSON\.parse\('((?:\\.|[^'])*)'\)`
	match := regexp.MustCompile(pattern).FindSubmatch(body)
	if len(match) != 2 {
		return "", fmt.Errorf("marker %q was not found", marker)
	}
	return string(match[1]), nil
}
