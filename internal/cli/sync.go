package cli

import (
	"github.com/spf13/cobra"

	"std-ai/internal/runner"
)

func newSyncCmd() *cobra.Command {
	var targets []string
	var noPull, noBackup, strict bool
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "核心同步：pull -> parse -> convert -> 向外扩散",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfgPath, root := resolveConfigPath()
			res, err := runner.Sync(runner.Options{
				ProjectRoot: root,
				ConfigPath:  cfgPath,
				DryRun:      flagDryRun,
				NoPull:      noPull,
				NoBackup:    noBackup,
				Strict:      strict,
				Targets:     targets,
				Version:     versionStr,
			})
			if err != nil {
				return err
			}
			cmd.Printf("[parse] %d source files -> %d docs\n", res.SourceFiles, res.Docs)
			for _, p := range res.Plans {
				cmd.Printf("[%s] %d files\n", p.Target, len(p.Files))
			}
			cmd.Printf("[done] %d written, %d skipped\n", res.Written, res.Skipped)
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
	f.StringSliceVar(&targets, "target", nil, "限定 sync 的 target，可重复")
	f.BoolVar(&noPull, "no-pull", false, "跳过 pull")
	f.BoolVar(&noBackup, "no-backup", false, "跳过 backup")
	f.BoolVar(&strict, "strict", false, "任何 warn 升级为 error")
	return cmd
}
