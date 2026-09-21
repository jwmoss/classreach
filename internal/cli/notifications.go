package cli

import "github.com/spf13/cobra"

func newNotificationsCommand(rc *runtime) *cobra.Command {
	cmd := &cobra.Command{Use: "notifications", Short: "Read notification counts"}
	var term string
	counts := &cobra.Command{
		Use: "counts", Short: "Read notification counts for an academic term as JSON",
		RunE: func(cmd *cobra.Command, args []string) error {
			data, err := rc.client.GetNotificationCounts(cmd.Context(), term)
			if err != nil {
				return err
			}
			return rc.out.JSON(data)
		},
	}
	counts.Flags().StringVar(&term, "term", "", "academic term ID from the overview section data")
	_ = counts.MarkFlagRequired("term")
	cmd.AddCommand(counts)
	return cmd
}
