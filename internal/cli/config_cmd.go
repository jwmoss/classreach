package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jwmoss/classreach/internal/config"
)

func newConfigCommand(rc *runtime) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Inspect and initialize configuration",
	}
	cmd.AddCommand(newConfigShowCommand(rc))
	cmd.AddCommand(newConfigInitCommand(rc))
	return cmd
}

func newConfigShowCommand(rc *runtime) *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show effective configuration with secrets redacted",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(rc.g.configPath)
			if err != nil {
				return err
			}
			if rc.g.baseURL != "" {
				cfg.BaseURL = rc.g.baseURL
			}
			if rc.out.IsJSON() {
				return rc.out.JSON(cfg.Redacted())
			}
			redacted := cfg.Redacted()
			rc.out.Table([]string{"KEY", "VALUE"}, [][]string{
				{"base_url", cfg.BaseURL},
				{"origin_host", cfg.OriginHost},
				{"username", redacted["username"]},
				{"password", redacted["password"]},
				{"path", config.DefaultPath()},
			})
			return nil
		},
	}
}

func newConfigInitCommand(rc *runtime) *cobra.Command {
	var (
		baseURL       string
		originHost    string
		username      string
		passwordStdin bool
		force         bool
	)
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Create a config file",
		RunE: func(cmd *cobra.Command, args []string) error {
			path := rc.g.configPath
			if path == "" {
				path = config.DefaultPath()
			}
			if !force && fileExists(path) {
				return fmt.Errorf("config already exists at %s; use --force to overwrite", path)
			}
			cfg := config.Default()
			cfg.BaseURL = baseURL
			cfg.OriginHost = originHost
			cfg.Username = username
			if passwordStdin {
				data, err := io.ReadAll(cmd.InOrStdin())
				if err != nil {
					return fmt.Errorf("read password from stdin: %w", err)
				}
				cfg.Password = strings.TrimSpace(string(data))
			}
			if err := config.Save(path, cfg); err != nil {
				return err
			}
			rc.out.Success("config written")
			rc.out.Printf("%s\n", path)
			return nil
		},
	}
	cmd.Flags().StringVar(&baseURL, "base-url", config.DefaultBaseURL, "ClassReach tenant URL")
	cmd.Flags().StringVar(
		&originHost,
		"origin-host",
		config.DefaultOriginHost,
		"ClassReach Azure origin host",
	)
	cmd.Flags().StringVar(&username, "username", "", "ClassReach username or email")
	cmd.Flags().BoolVar(&passwordStdin, "password-stdin", false, "read the password from stdin")
	cmd.Flags().BoolVar(&force, "force", false, "overwrite an existing config file")
	return cmd
}
