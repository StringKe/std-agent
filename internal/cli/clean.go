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
		Short: "Remove generated files from root and platform dirs, keep .stdai/",
		RunE: func(cmd *cobra.Command, _ []string) error {
			_ = keepBackups // v1.0: clean always keeps .stdai/backups/
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
	f.StringSliceVar(&targets, "target", nil, "Only clean the given target(s)")
	f.BoolVar(&keepBackups, "keep-backups", true, "Keep .stdai/backups/")
	f.BoolVarP(&yes, "yes", "y", false, "Skip confirmation")
	return cmd
}

// cleanEmptyDirs removes parent dirs of cleaned files when they are empty.
//
// The exit condition filepath.Dir(dir) == dir detects the filesystem root
// and is safe across OSes:
// Linux/macOS: filepath.Dir("/") == "/"
// Windows:     filepath.Dir("C:\\") == "C:\\"
func cleanEmptyDirs(paths []string) {
	dirs := map[string]bool{}
	for _, p := range paths {
		dir := filepath.Dir(p)
		for dir != "." && dir != "" {
			parent := filepath.Dir(dir)
			if parent == dir {
				break // reached filesystem root
			}
			dirs[dir] = true
			dir = parent
		}
	}
	dirList := make([]string, 0, len(dirs))
	for d := range dirs {
		dirList = append(dirList, d)
	}
	// Sort by depth so deeper dirs are removed first
	sort.Slice(dirList, func(i, j int) bool {
		return len(dirList[i]) > len(dirList[j])
	})
	for _, d := range dirList {
		_ = os.Remove(d) // non-empty dirs fail naturally, ignored
	}
}
