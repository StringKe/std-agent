package cli

import (
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/StringKe/std-agent/internal/source"
	"github.com/StringKe/std-agent/internal/state"
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
			// 与 sync 相同的保守过滤：不吃未改写的上次 sync 输出，同名外部包让位本地源
			skip := map[string]string{}
			if st, serr := state.Load(filepath.Join(root, state.StateFile)); serr == nil && st != nil {
				for _, t := range st.Targets {
					for p, sha := range t.Outputs {
						skip[filepath.ToSlash(p)] = sha
					}
				}
			}
			opts := source.ExternalScanOptions{
				SkipPaths:  skip,
				UserSkills: source.LocalSkillNames(filepath.Join(root, ".stdai/standards")),
			}
			adopted, warns, err := source.ImportExternal(root, flagDryRun, opts)
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
