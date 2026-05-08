package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestInstallHookSucceedsInGitRepo(t *testing.T) {
	old := flagConfig
	defer func() { flagConfig = old }()
	flagConfig = ".stdai/config.toml"

	tmp := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmp, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(tmp)

	cmd := newInstallHookCmd()
	var out, errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	cmd.SetArgs(nil)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	hookPath := filepath.Join(tmp, ".git", "hooks", "pre-commit")
	info, err := os.Stat(hookPath)
	if err != nil {
		t.Fatalf("hook missing: %v", err)
	}
	// Unix executable bit；Windows NTFS 不支持 Unix mode bits 跳过此检查
	if runtime.GOOS != "windows" && info.Mode().Perm()&0o100 == 0 {
		t.Errorf("hook not executable, mode = %v", info.Mode())
	}
	content, _ := os.ReadFile(hookPath)
	if !strings.Contains(string(content), "stdagent status --strict") {
		t.Errorf("hook content missing strict status: %s", content)
	}
}

func TestInstallHookFailsOutsideGitRepo(t *testing.T) {
	old := flagConfig
	defer func() { flagConfig = old }()
	flagConfig = ".stdai/config.toml"

	tmp := t.TempDir()
	t.Chdir(tmp)

	cmd := newInstallHookCmd()
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs(nil)
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "not a git repository") {
		t.Errorf("expected git repo error, got %v", err)
	}
}

func TestInstallHookForceOverwrites(t *testing.T) {
	old := flagConfig
	defer func() { flagConfig = old }()
	flagConfig = ".stdai/config.toml"

	tmp := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmp, ".git/hooks"), 0o755); err != nil {
		t.Fatal(err)
	}
	hookPath := filepath.Join(tmp, ".git/hooks/pre-commit")
	if err := os.WriteFile(hookPath, []byte("# old hook"), 0o755); err != nil { //nolint:gosec
		t.Fatal(err)
	}
	t.Chdir(tmp)

	// 不带 --force 应失败
	cmd := newInstallHookCmd()
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs(nil)
	if err := cmd.Execute(); err == nil {
		t.Error("expected error without --force")
	}

	// 带 --force 应成功覆盖
	cmd2 := newInstallHookCmd()
	cmd2.SetOut(new(bytes.Buffer))
	cmd2.SetErr(new(bytes.Buffer))
	cmd2.SetArgs([]string{"--force"})
	if err := cmd2.Execute(); err != nil {
		t.Fatalf("execute --force: %v", err)
	}
	content, _ := os.ReadFile(hookPath)
	if !strings.Contains(string(content), "stdagent status --strict") {
		t.Errorf("hook not overwritten: %s", content)
	}
}
