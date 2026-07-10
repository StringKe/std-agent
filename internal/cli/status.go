package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"

	"github.com/StringKe/std-agent/internal/config"
	"github.com/StringKe/std-agent/internal/state"
	"github.com/StringKe/std-agent/internal/writer"
)

type targetReport struct {
	Name         string `json:"name"`
	Enabled      bool   `json:"enabled"`
	Convert      bool   `json:"convert"`
	LastSync     string `json:"last_sync,omitempty"`
	FilesTracked int    `json:"files_tracked"`
	Drift        int    `json:"drift"`
	Missing      int    `json:"missing"`
}

func newStatusCmd() *cobra.Command {
	var targets []string
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "status",
		Short: "显示 targets 状态、drift 与最后同步时间",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfgPath, root := resolveConfigPath()
			cfg, err := config.Load(cfgPath)
			if err != nil {
				return err
			}
			st, err := state.Load(filepath.Join(root, state.StateFile))
			if err != nil {
				return err
			}

			var reports []targetReport
			for name, t := range cfg.Targets {
				if len(targets) > 0 && !sliceContains(targets, name) {
					continue
				}
				r := targetReport{Name: name, Enabled: t.Enabled, Convert: t.Convert}
				if ts, ok := st.Targets[name]; ok {
					r.LastSync = ts.LastSync.UTC().Format("2006-01-02T15:04:05Z")
					r.FilesTracked = len(ts.Outputs)
					for path, want := range ts.Outputs {
						actual, rerr := os.ReadFile(filepath.Join(root, path)) //nolint:gosec
						if rerr != nil {
							r.Missing++
							continue
						}
						if writer.Checksum(actual) != want {
							r.Drift++
						}
					}
				}
				reports = append(reports, r)
			}
			sort.Slice(reports, func(i, j int) bool { return reports[i].Name < reports[j].Name })

			if asJSON {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(reports)
			}

			driftTotal := 0
			for _, r := range reports {
				flag := "[disabled]"
				if r.Enabled {
					flag = "[enabled] "
				}
				lastSync := r.LastSync
				if lastSync == "" {
					lastSync = "never"
				}
				cmd.Printf("%s %-12s last_sync=%s tracked=%d drift=%d missing=%d\n",
					flag, r.Name, lastSync, r.FilesTracked, r.Drift, r.Missing)
				driftTotal += r.Drift + r.Missing
			}
			if driftTotal > 0 {
				return fmt.Errorf("%d files drifted or missing", driftTotal)
			}
			return nil
		},
	}
	cmd.Flags().StringSliceVar(&targets, "target", nil, "限定 target")
	cmd.Flags().BoolVar(&asJSON, "json", false, "结构化 JSON 输出")
	return cmd
}
