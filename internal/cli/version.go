package cli

import (
	"encoding/json"
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	versionStr = "dev"
	commitStr  = "none"
	dateStr    = "unknown"
)

// SetVersion 由 main 注入 ldflags 值；空值不覆盖现有
func SetVersion(v, c, d string) {
	if v != "" {
		versionStr = v
	}
	if c != "" {
		commitStr = c
	}
	if d != "" {
		dateStr = d
	}
}

func newVersionCmd() *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "version",
		Short: "打印版本与构建信息",
		RunE: func(cmd *cobra.Command, _ []string) error {
			info := map[string]string{
				"version": versionStr,
				"commit":  commitStr,
				"built":   dateStr,
				"go":      runtime.Version(),
				"os":      runtime.GOOS,
				"arch":    runtime.GOARCH,
			}
			if asJSON {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(info)
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(),
				"stdagent %s\n  commit:  %s\n  built:   %s\n  go:      %s\n  os/arch: %s/%s\n",
				versionStr, commitStr, dateStr, runtime.Version(), runtime.GOOS, runtime.GOARCH)
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "结构化 JSON 输出")
	return cmd
}
