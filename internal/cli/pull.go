package cli

import (
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"

	"github.com/StringKe/std-agent/internal/config"
	"github.com/StringKe/std-agent/internal/source"
)

func newPullCmd() *cobra.Command {
	var sourceName string
	var all bool
	cmd := &cobra.Command{
		Use:   "pull",
		Short: "更新 .stdai/cache/ 中的 Git 源",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfgPath, root := resolveConfigPath()
			cfg, err := config.Load(cfgPath)
			if err != nil {
				return err
			}

			names := make([]string, 0, len(cfg.Sources))
			for n := range cfg.Sources {
				names = append(names, n)
			}
			sort.Strings(names)

			pulled := 0
			for _, name := range names {
				src := cfg.Sources[name]
				if sourceName != "" && name != sourceName {
					continue
				}
				if !src.Enabled && !all {
					continue
				}
				g := &source.Git{
					NameValue: name,
					URL:       src.URL,
					Branch:    src.Branch,
					CacheDir:  filepath.Join(root, ".stdai/cache", name),
					Paths:     src.Paths,
					Auth:      src.Auth,
					TokenEnv:  src.TokenEnv,
				}
				cmd.Printf("[pull] %s <- %s (branch %s)\n", name, src.URL, defaultBranch(src.Branch))
				if perr := g.Pull(); perr != nil {
					cmd.PrintErrf("  failed: %v\n", perr)
					continue
				}
				pulled++
			}
			cmd.Printf("[done] %d sources pulled\n", pulled)
			return nil
		},
	}
	cmd.Flags().StringVar(&sourceName, "source", "", "仅 pull 指定 source")
	cmd.Flags().BoolVar(&all, "all", false, "即使 enabled=false 也 pull")
	return cmd
}

func defaultBranch(b string) string {
	if b == "" {
		return "main"
	}
	return b
}
