package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/jwmoss/classreach/internal/output"
)

func configureCommands(cmd *cobra.Command) {
	cmd.Version = version
	cmd.SetVersionTemplate("classreach version {{.Version}}\n")
	if cmd.Args == nil {
		cmd.Args = usageArgs(cobra.NoArgs)
	}
	for _, child := range cmd.Commands() {
		configureCommands(child)
	}
}

func (rc *runtime) prepareCommand(cmd *cobra.Command, args []string) error {
	if err := applyGlobalEnv(cmd, rc.g); err != nil {
		return fmt.Errorf("%w: %v", errUsage, err)
	}
	if rc.g.asJSON && rc.g.plain {
		return fmt.Errorf("%w: choose only one of --json or --plain", errUsage)
	}
	if rc.g.timeout <= 0 {
		return fmt.Errorf("%w: --timeout must be greater than zero", errUsage)
	}
	if err := validateCommandFlags(cmd); err != nil {
		return fmt.Errorf("%w: %v", errUsage, err)
	}
	rc.out = output.New(rc.stdout, rc.stderr, rc.g.asJSON, rc.g.plain, rc.g.quiet, rc.g.noColor)
	if rc.g.dryRun && (!commandSkipsClient(cmd) || cmd.CommandPath() == "classreach config init") {
		if rc.out.IsJSON() {
			if err := rc.out.JSON(map[string]any{"dry_run": true, "command": cmd.CommandPath()}); err != nil {
				return err
			}
		} else {
			rc.out.Printf("Would run %s; no network requests or file changes.\n", cmd.CommandPath())
		}
		return errDryRun
	}
	if commandSkipsClient(cmd) {
		return nil
	}
	return rc.initClient()
}

func validateCommandFlags(cmd *cobra.Command) error {
	if err := cmd.ValidateRequiredFlags(); err != nil {
		return err
	}
	flags := cmd.Flags()
	if flags.Lookup("term") != nil {
		value, _ := flags.GetString("term")
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("--term is required")
		}
	}
	if flags.Lookup("section") != nil {
		for _, name := range []string{"student", "section"} {
			value, _ := flags.GetString(name)
			if strings.TrimSpace(value) == "" {
				return fmt.Errorf("--%s is required", name)
			}
		}
	}
	if flags.Lookup("output") != nil {
		value, _ := flags.GetString("output")
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("--output is required")
		}
	}
	if err := validateDateFlags(cmd); err != nil {
		return err
	}
	for _, name := range []string{"page", "per-page"} {
		if flags.Lookup(name) != nil {
			value, _ := flags.GetInt(name)
			if value < 1 {
				return fmt.Errorf("--%s must be at least 1", name)
			}
		}
	}
	if flags.Lookup("query") != nil {
		values, _ := flags.GetStringArray("query")
		_, err := parseQuery(values)
		return err
	}
	return nil
}

func validateDateFlags(cmd *cobra.Command) error {
	flags := cmd.Flags()
	if flags.Lookup("week") != nil {
		value, _ := flags.GetString("week")
		if _, err := time.Parse(dateLayout, value); err != nil {
			return fmt.Errorf("invalid week date: use YYYY-MM-DD")
		}
	}
	if flags.Lookup("start") != nil {
		start, _ := flags.GetString("start")
		end, _ := flags.GetString("end")
		return validateDateRange(start, end)
	}
	return nil
}
