package cli

import (
	"os"
	"path/filepath"
)

func sliceContains(ss []string, s string) bool {
	for _, x := range ss {
		if x == s {
			return true
		}
	}
	return false
}

// resolveConfigPath decides which .stdai/config.toml and project root to use.
//
// With an explicit --config, root is the config's grandparent dir
// (path/.stdai/config.toml -> path). By default, walk up from the cwd to the
// nearest .stdai/config.toml; fall back to the cwd when none is found.
//
// The walk-up behavior supports running stdagent from a monorepo
// subdirectory, auto-locating the enclosing project root.
func resolveConfigPath() (cfgPath, root string) {
	const defaultRel = ".stdai/config.toml"

	if flagConfig != defaultRel {
		abs, _ := filepath.Abs(flagConfig)
		return abs, filepath.Dir(filepath.Dir(abs))
	}

	cwd, err := os.Getwd()
	if err != nil {
		return flagConfig, "."
	}

	dir := cwd
	for {
		candidate := filepath.Join(dir, defaultRel)
		if _, err := os.Stat(candidate); err == nil {
			return candidate, dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	abs, _ := filepath.Abs(flagConfig)
	return abs, cwd
}
