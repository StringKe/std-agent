package cli

import (
	"github.com/spf13/cobra"

	"github.com/StringKe/std-agent/internal/runner"
)

func newSyncCmd() *cobra.Command {
	var targets []string
	var noPull, noBackup, noPrune, noExternal, strict bool
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Core sync: pull -> parse -> convert -> fan out (prunes previously written but no longer produced files by default)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfgPath, root := resolveConfigPath()
			res, err := runner.Sync(runner.Options{
				ProjectRoot: root,
				ConfigPath:  cfgPath,
				DryRun:      flagDryRun,
				NoPull:      noPull,
				NoBackup:    noBackup,
				NoPrune:     noPrune,
				NoExternal:  noExternal,
				Strict:      strict,
				Targets:     targets,
				Version:     versionStr,
			})
			if err != nil {
				return err
			}
			cmd.Printf("[parse] %d source files -> %d docs\n", res.SourceFiles, res.Docs)
			if res.ExternalFiles > 0 {
				cmd.Printf("[external] %d adopted files (.ai/.agents)\n", res.ExternalFiles)
			}
			for _, p := range res.Plans {
				cmd.Printf("[%s] %d files\n", p.Target, len(p.Files))
			}
			cmd.Printf("[done] %d written, %d skipped\n", res.Written, res.Skipped)
			if res.GitignoreUpdated {
				if flagDryRun {
					cmd.Println("[gitignore] would update managed block")
				} else {
					cmd.Println("[gitignore] updated managed block")
				}
			}
			if res.Pruned > 0 {
				if flagDryRun {
					cmd.Printf("[prune] %d orphans would be removed:\n", res.Pruned)
				} else {
					cmd.Printf("[prune] %d orphans removed:\n", res.Pruned)
				}
				for _, p := range res.PrunedPaths {
					cmd.Printf("  %s\n", p)
				}
			}
			if res.BackupDir != "" {
				cmd.Printf("[backup] %s\n", res.BackupDir)
			}
			for _, w := range res.Warnings {
				cmd.PrintErrln("[warn]", w)
			}
			return nil
		},
	}
	f := cmd.Flags()
	f.StringSliceVar(&targets, "target", nil, "Limit sync to the given target(s), repeatable")
	f.BoolVar(&noPull, "no-pull", false, "Skip pull")
	f.BoolVar(&noBackup, "no-backup", false, "Skip backup")
	f.BoolVar(&noPrune, "no-prune", false, "Keep orphan files written last time but no longer produced (deleted by default)")
	f.BoolVar(&noExternal, "no-external", false, "Skip external artifact auto-adopt (.ai/*, .agents/skills; enabled by default)")
	f.BoolVar(&strict, "strict", false, "Promote any warning to an error")
	return cmd
}
