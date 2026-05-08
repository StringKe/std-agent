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

// resolveConfigPath 决定使用哪个 .stdai/config.toml 与 project root。
//
// 用户显式 --config 时，root = config 的祖父目录（path/.stdai/config.toml -> path）。
// 默认情况下，从 cwd 向上 walk，找到最近的 .stdai/config.toml；找不到则 fallback 到 cwd。
//
// 该 walkUp 行为支持 monorepo 子目录运行 stdagent，自动定位上层项目根。
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
