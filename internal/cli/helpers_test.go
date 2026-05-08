package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSliceContains(t *testing.T) {
	cases := []struct {
		ss   []string
		s    string
		want bool
	}{
		{[]string{"a", "b", "c"}, "b", true},
		{[]string{"a"}, "z", false},
		{nil, "a", false},
		{[]string{}, "a", false},
	}
	for _, c := range cases {
		if got := sliceContains(c.ss, c.s); got != c.want {
			t.Errorf("sliceContains(%v, %q) = %v, want %v", c.ss, c.s, got, c.want)
		}
	}
}

func TestResolveConfigPathExplicit(t *testing.T) {
	old := flagConfig
	defer func() { flagConfig = old }()

	tmp := t.TempDir()
	custom := filepath.Join(tmp, "custom", "config.toml")
	flagConfig = custom

	cfg, root := resolveConfigPath()
	if cfg != custom {
		t.Errorf("cfgPath = %s, want %s", cfg, custom)
	}
	// /tmp/.../custom/config.toml -> /tmp/.../custom -> /tmp/...
	if root != tmp {
		t.Errorf("root = %s, want %s", root, tmp)
	}
}

func TestResolveConfigPathWalkUp(t *testing.T) {
	old := flagConfig
	defer func() { flagConfig = old }()
	flagConfig = ".stdai/config.toml"

	tmp := t.TempDir()
	sub := filepath.Join(tmp, "sub", "deeper")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(tmp, ".stdai"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmp, ".stdai/config.toml"), []byte("v=1"), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Chdir(sub)

	cfgPath, root := resolveConfigPath()
	if !strings.HasSuffix(filepath.ToSlash(cfgPath), ".stdai/config.toml") {
		t.Errorf("cfgPath = %s", cfgPath)
	}
	// 解析 symlink 后 root 应当指向 tmp（macOS /private/var 链接处理）
	rootAbs, _ := filepath.EvalSymlinks(root)
	tmpAbs, _ := filepath.EvalSymlinks(tmp)
	if rootAbs != tmpAbs {
		t.Errorf("root = %s, want %s", rootAbs, tmpAbs)
	}
}

func TestResolveConfigPathNoStdai(t *testing.T) {
	old := flagConfig
	defer func() { flagConfig = old }()
	flagConfig = ".stdai/config.toml"

	tmp := t.TempDir()
	t.Chdir(tmp)

	// 没有任何 .stdai/，fallback 到 cwd 自身
	_, root := resolveConfigPath()
	rootAbs, _ := filepath.EvalSymlinks(root)
	tmpAbs, _ := filepath.EvalSymlinks(tmp)
	if rootAbs != tmpAbs {
		t.Errorf("fallback root = %s, want %s", rootAbs, tmpAbs)
	}
}
