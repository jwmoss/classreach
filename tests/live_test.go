//go:build live

package tests

import (
	"context"
	"encoding/json"
	"os/exec"
	"testing"
	"time"
)

// TestLiveReadOnly requires the real provider and the user's existing configuration.
// It never changes provider state and never logs response bodies or credentials.
func TestLiveReadOnly(t *testing.T) {
	cases := []struct {
		name  string
		args  []string
		valid func(any) bool
	}{
		{"authenticated doctor", []string{"doctor"}, func(v any) bool {
			data, ok := v.(map[string]any)
			return ok && data["ok"] == true
		}},
		{"visible students", []string{"students", "list"}, func(v any) bool {
			data, ok := v.([]any)
			return ok && len(data) > 0
		}},
	}
	for _, check := range cases {
		t.Run(check.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
			defer cancel()
			args := append([]string{"--json"}, check.args...)
			data, err := exec.CommandContext(ctx, "../bin/classreach", args...).Output()
			if err != nil {
				t.Fatalf("live CLI failed: %v; check provider access and configuration", err)
			}
			var result any
			if err := json.Unmarshal(data, &result); err != nil {
				t.Fatal("live CLI returned invalid JSON")
			}
			if !check.valid(result) {
				t.Fatal("live response does not satisfy the command contract")
			}
		})
	}
}
