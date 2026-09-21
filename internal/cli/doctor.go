package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newDoctorCommand(rc *runtime) *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Verify configuration and API connectivity",
		RunE: func(cmd *cobra.Command, args []string) error {
			view, err := rc.client.GetQuickView(cmd.Context(), defaultWeek())
			if err != nil {
				return err
			}
			if view.Students == nil {
				return fmt.Errorf("quick view is missing UserInfos; authentication or the API contract has changed")
			}
			payload := map[string]any{
				"ok":       true,
				"base_url": rc.cfg.BaseURL,
			}
			if rc.out.IsJSON() {
				return rc.out.JSON(payload)
			}
			rc.out.Success("API reachable")
			rc.out.Printf("base_url: %s\n", rc.cfg.BaseURL)
			return nil
		},
	}
}
