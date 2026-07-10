package cli

import (
	"github.com/spf13/cobra"

	"github.com/StringKe/std-agent/internal/runner"
)

// newFixCmd 是 sync 的语义别名：drift auto-fix 即重新 sync 覆盖 drift 文件
func newFixCmd() *cobra.Command {
	var targets []string
	cmd := &cobra.Command{
		Use:   "fix",
		Short: "重新 sync 修复 drift（等价于 sync）",
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
	cmd.Flags().StringSliceVar(&targets, "target", nil, "限定 target")
	return cmd
}
