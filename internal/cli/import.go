package cli

import (
	"github.com/spf13/cobra"

	"github.com/StringKe/std-agent/internal/source"
)

// newImportCmd adopts external artifacts (.ai/*, .agents/skills, root .mcp.json)
// into .stdai/standards/external/ (plus mcp.json merge) for review and
// customization before sync. Sync already scans externals in memory by default;
// import materializes them on disk for pinning and further edits.
func newImportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import",
		Short: "Adopt external artifacts into .stdai/standards/external/ (.ai/*, .agents/skills, .mcp.json)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfgPath, root := resolveConfigPath()
			_ = cfgPath
			adopted, warns, err := source.ImportExternal(root, flagDryRun)
			if err != nil {
				return err
			}
			if len(adopted) == 0 {
				cmd.Println("[import] no external artifacts found")
			} else if flagDryRun {
				cmd.Printf("[import] %d files would be adopted:\n", len(adopted))
			} else {
				cmd.Printf("[import] %d files adopted:\n", len(adopted))
			}
			for _, p := range adopted {
				cmd.Printf("  %s\n", p)
			}
			for _, w := range warns {
				cmd.PrintErrln("[warn]", w)
			}
			return nil
		},
	}
	return cmd
}
