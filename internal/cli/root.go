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
		Short:         "std-agent multi-AI CLI config synchronizer",
		Long:          "stdagent uses .stdai/ as the single source of truth, syncing YAML frontmatter + Markdown into each AI CLI tool's native config.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	pf := root.PersistentFlags()
	pf.StringVar(&flagConfig, "config", ".stdai/config.toml", "Path to the config file")
	pf.BoolVar(&flagDryRun, "dry-run", false, "Print what would be done without writing")
	pf.CountVarP(&flagVerbose, "verbose", "v", "Verbose logging (-v info, -vv debug)")
	pf.BoolVarP(&flagQuiet, "quiet", "q", false, "Errors only")
	pf.BoolVar(&flagNoColor, "no-color", false, "Disable colors")

	root.AddCommand(
		newInitCmd(),
		newPullCmd(),
		newSyncCmd(),
		newImportCmd(),
		newFixCmd(),
		newStatusCmd(),
		newCleanCmd(),
		newBudgetCmd(),
		newWhichCmd(),
		newExplainCmd(),
		newIntroCmd(),
		newUpgradeCmd(),
		newVersionCmd(),
	)
	return root
}

// Execute is the main cmd entrypoint
func Execute() error {
	return newRootCmd().Execute()
}
