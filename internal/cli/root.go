package cli

import "github.com/spf13/cobra"

var (
	flagConfig  string
	flagDryRun  bool
	flagVerbose int
	flagQuiet   bool
	flagNoColor bool
)

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "stdagent",
		Short:         "std-ai 多 AI CLI 配置同步器",
		Long:          "stdagent 以 .stdai/ 为内部单一真相源，把 YAML frontmatter + Markdown 同步到 9 个 AI CLI 工具的扩散文件。",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	pf := root.PersistentFlags()
	pf.StringVar(&flagConfig, "config", ".stdai/config.toml", "配置文件路径")
	pf.BoolVar(&flagDryRun, "dry-run", false, "只输出将做什么，不写盘")
	pf.CountVarP(&flagVerbose, "verbose", "v", "详细日志（-v info, -vv debug）")
	pf.BoolVarP(&flagQuiet, "quiet", "q", false, "仅输出错误")
	pf.BoolVar(&flagNoColor, "no-color", false, "禁用颜色")

	root.AddCommand(
		newInitCmd(),
		newPullCmd(),
		newSyncCmd(),
		newFixCmd(),
		newStatusCmd(),
		newCleanCmd(),
		newBudgetCmd(),
		newIntroCmd(),
		newUpgradeCmd(),
		newVersionCmd(),
	)
	return root
}

// Execute 是 cmd 主入口
func Execute() error {
	return newRootCmd().Execute()
}
