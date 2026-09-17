package cli

import (
	"github.com/spf13/cobra"

	"github.com/StringKe/std-agent/internal/runner"
)

// newFixCmd is a semantic alias of sync: drift auto-fix re-syncs to overwrite drifted files
func newFixCmd() *cobra.Command {
	var targets []string
	cmd := &cobra.Command{
		Use:   "fix",
		Short: "Re-sync to fix drift (equivalent to sync)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfgPath, root := resolveConfigPath()
			res, err := runner.Sync(runner.Options{
				ProjectRoot: root,
				ConfigPath:  cfgPath,
				Targets:     targets,
				Version:     versionStr,
			})
			if err != nil {
				return err
			}
			cmd.Printf("[fix] %d written, %d skipped\n", res.Written, res.Skipped)
			for _, w := range res.Warnings {
				cmd.PrintErrln("[warn]", w)
			}
			return nil
		},
	}
	cmd.Flags().StringSliceVar(&targets, "target", nil, "Limit to the given target(s)")
	return cmd
}
