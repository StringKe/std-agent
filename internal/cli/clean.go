package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"

	"github.com/StringKe/std-agent/internal/state"
)

func newCleanCmd() *cobra.Command {
	var targets []string
	var keepBackups, yes bool
	cmd := &cobra.Command{
		Use:   "clean",
		Short: "清空根目录与平台目录的生成文件，保留 .stdai/",
		RunE: func(cmd *cobra.Command, _ []string) error {
			_ = keepBackups // 当前 v1.0：clean 永远保留 .stdai/backups/
			_, root := resolveConfigPath()
			st, err := state.Load(filepath.Join(root, state.StateFile))
			if err != nil {
				return err
			}
			var paths []string
			for name, t := range st.Targets {
				if len(targets) > 0 && !sliceContains(targets, name) {
					continue
				}
				for p := range t.Outputs {
					paths = append(paths, p)
				}
			}
			sort.Strings(paths)
			if len(paths) == 0 {
				cmd.Println("nothing to clean")
				return nil
			}
			cmd.Printf("Will delete %d generated files:\n", len(paths))
			for _, p := range paths {
				cmd.Printf("  %s\n", p)
			}
			if !yes {
				cmd.Print("Proceed? [y/N]: ")
				var answer string
				_, _ = fmt.Fscanln(os.Stdin, &answer)
				if answer != "y" && answer != "Y" {
					return errors.New("user cancelled")
				}
			}
			removed := 0
			absPaths := make([]string, 0, len(paths))
			for _, p := range paths {
				full := filepath.Join(root, p)
				absPaths = append(absPaths, full)
				if rerr := os.Remove(full); rerr != nil {
					cmd.PrintErrf("warn: %v\n", rerr)
					continue
				}
				removed++
			}
			cleanEmptyDirs(absPaths)
			cmd.Printf("Removed %d files\n", removed)
			return nil
		},
	}
	f := cmd.Flags()
	f.StringSliceVar(&targets, "target", nil, "仅清理指定 target")
	f.BoolVar(&keepBackups, "keep-backups", true, "保留 .stdai/backups/")
	f.BoolVarP(&yes, "yes", "y", false, "跳过确认")
	return cmd
}

// cleanEmptyDirs 删除被清理文件的父目录（若已空）
//
// 退出条件用 filepath.Dir(dir) == dir 判定 root，跨 OS 安全：
// Linux/macOS: filepath.Dir("/") == "/"
// Windows:     filepath.Dir("C:\\") == "C:\\"
func cleanEmptyDirs(paths []string) {
	dirs := map[string]bool{}
	for _, p := range paths {
		dir := filepath.Dir(p)
		for dir != "." && dir != "" {
			parent := filepath.Dir(dir)
			if parent == dir {
				break // 已到 filesystem root
			}
			dirs[dir] = true
			dir = parent
		}
	}
	dirList := make([]string, 0, len(dirs))
	for d := range dirs {
		dirList = append(dirList, d)
	}
	// 按深度排序，深的先删
	sort.Slice(dirList, func(i, j int) bool {
		return len(dirList[i]) > len(dirList[j])
	})
	for _, d := range dirList {
		_ = os.Remove(d) // 非空目录会自然失败，忽略
	}
}
