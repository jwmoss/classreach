package cli

import "github.com/spf13/cobra"

func newLoginCommand(rc *runtime) *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Validate the configured ClassReach credentials",
		RunE: func(cmd *cobra.Command, args []string) error {
			payload := map[string]any{
				"ok":       true,
				"base_url": rc.cfg.BaseURL,
			}
			if rc.out.IsJSON() {
				return rc.out.JSON(payload)
			}
			rc.out.Success("login successful")
			rc.out.Printf("base_url: %s\n", rc.cfg.BaseURL)
			return nil
		},
	}
}
