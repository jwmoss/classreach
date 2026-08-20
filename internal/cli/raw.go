package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
)

func newRawCommand(rc *runtime) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "raw",
		Short: "Send a read-only raw HTTP request",
	}
	cmd.AddCommand(newRawGetCommand(rc))
	return cmd
}

func newRawGetCommand(rc *runtime) *cobra.Command {
	var queryFlag []string
	cmd := &cobra.Command{
		Use:   "get <path>",
		Short: "Send a raw GET request",
		Args:  usageArgs(cobra.ExactArgs(1)),
		RunE: func(cmd *cobra.Command, args []string) error {
			query, err := parseQuery(queryFlag)
			if err != nil {
				return err
			}
			resp, err := rc.client.Do(cmd.Context(), http.MethodGet, args[0], query, nil)
			if err != nil {
				return err
			}
			if rc.out.IsJSON() && json.Valid(resp) {
				var decoded any
				if err := json.Unmarshal(resp, &decoded); err != nil {
					return err
				}
				return rc.out.JSON(decoded)
			}
			if len(resp) == 0 {
				return nil
			}
			rc.out.Printf("%s", string(resp))
			if !strings.HasSuffix(string(resp), "\n") {
				rc.out.Printf("\n")
			}
			return nil
		},
	}
	cmd.Flags().StringArrayVar(&queryFlag, "query", nil, "query parameter in key=value form")
	return cmd
}

func parseQuery(items []string) (url.Values, error) {
	values := url.Values{}
	for _, item := range items {
		key, value, ok := strings.Cut(item, "=")
		if !ok || strings.TrimSpace(key) == "" {
			return nil, fmt.Errorf("%w: query must be key=value", errUsage)
		}
		values.Add(key, value)
	}
	return values, nil
}
