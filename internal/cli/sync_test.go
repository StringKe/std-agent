package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupSyncProject(t *testing.T, tmp string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(tmp, ".stdai/standards/rules"), 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := `version = "1.0"
inject = false
backup = false
auto_pull = false

[targets]
claude-code = { enabled = true, convert = true }
codex = { enabled = true, convert = true }
`
	if err := os.WriteFile(filepath.Join(tmp, ".stdai/config.toml"), []byte(cfg), 0o600); err != nil {
		t.Fatal(err)
	}
	rule := `---
type: rules
name: style
description: Test
---
Use clear names.
`
	if err := os.WriteFile(filepath.Join(tmp, ".stdai/standards/rules/style.md"), []byte(rule), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestSyncCommandActuallyWrites(t *testing.T) {
	old := flagConfig
	defer func() { flagConfig = old }()
	flagConfig = ".stdai/config.toml"

	tmp := t.TempDir()
	setupSyncProject(t, tmp)
	t.Chdir(tmp)

	cmd := newSyncCmd()
	var out, errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !strings.Contains(out.String(), "[done]") {
		t.Errorf("missing [done] in output:\n%s", out.String())
	}
	for _, want := range []string{"CLAUDE.md", "AGENTS.md", ".claude/rules/style.md"} {
		if _, err := os.Stat(filepath.Join(tmp, want)); err != nil {
			t.Errorf("expected %s: %v", want, err)
		}
	}
}

func TestSyncCommandTargetFilter(t *testing.T) {
	old := flagConfig
	defer func() { flagConfig = old }()
	flagConfig = ".stdai/config.toml"

	tmp := t.TempDir()
	setupSyncProject(t, tmp)
	t.Chdir(tmp)

	cmd := newSyncCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs([]string{"--target", "claude-code"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tmp, "CLAUDE.md")); err != nil {
		t.Errorf("CLAUDE.md should exist: %v", err)
	}
	// codex 没在 --target 列表，AGENTS.md 不应生成
	if _, err := os.Stat(filepath.Join(tmp, "AGENTS.md")); err == nil {
		t.Error("AGENTS.md should not exist when --target only claude-code")
	}
}

func TestFixCommandIsAliasOfSync(t *testing.T) {
	old := flagConfig
	defer func() { flagConfig = old }()
	flagConfig = ".stdai/config.toml"

	tmp := t.TempDir()
	setupSyncProject(t, tmp)
	t.Chdir(tmp)

	cmd := newFixCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(new(bytes.Buffer))
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute fix: %v", err)
	}
	if !strings.Contains(out.String(), "[fix]") {
		t.Errorf("missing [fix] prefix:\n%s", out.String())
	}
	if _, err := os.Stat(filepath.Join(tmp, "CLAUDE.md")); err != nil {
		t.Errorf("fix should also produce CLAUDE.md: %v", err)
	}
}

func TestStatusCommandAfterSync(t *testing.T) {
	old := flagConfig
	defer func() { flagConfig = old }()
	flagConfig = ".stdai/config.toml"

	tmp := t.TempDir()
	setupSyncProject(t, tmp)
	t.Chdir(tmp)

	// 先 sync
	syncCmd := newSyncCmd()
	syncCmd.SetOut(new(bytes.Buffer))
	syncCmd.SetErr(new(bytes.Buffer))
	if err := syncCmd.Execute(); err != nil {
		t.Fatalf("sync: %v", err)
	}

	// 再 status；无 drift 应当返回 nil
	statusCmd := newStatusCmd()
	var out bytes.Buffer
	statusCmd.SetOut(&out)
	statusCmd.SetErr(new(bytes.Buffer))
	if err := statusCmd.Execute(); err != nil {
		t.Errorf("status after fresh sync should not error: %v", err)
	}
	for _, want := range []string{"claude-code", "tracked=", "drift=0"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("status missing %q in:\n%s", want, out.String())
		}
	}
}

func TestStatusCommandJSON(t *testing.T) {
	old := flagConfig
	defer func() { flagConfig = old }()
	flagConfig = ".stdai/config.toml"

	tmp := t.TempDir()
	setupSyncProject(t, tmp)
	t.Chdir(tmp)

	syncCmd := newSyncCmd()
	syncCmd.SetOut(new(bytes.Buffer))
	syncCmd.SetErr(new(bytes.Buffer))
	if err := syncCmd.Execute(); err != nil {
		t.Fatal(err)
	}

	statusCmd := newStatusCmd()
	var out bytes.Buffer
	statusCmd.SetOut(&out)
	statusCmd.SetErr(new(bytes.Buffer))
	statusCmd.SetArgs([]string{"--json"})
	if err := statusCmd.Execute(); err != nil {
		t.Fatalf("status --json: %v", err)
	}
	var reports []targetReport
	if err := json.Unmarshal(out.Bytes(), &reports); err != nil {
		t.Fatalf("not valid JSON: %v\n%s", err, out.String())
	}
	if len(reports) < 2 {
		t.Errorf("expected >= 2 reports, got %d", len(reports))
	}
	for _, r := range reports {
		if r.Name == "claude-code" && r.Drift != 0 {
			t.Errorf("claude-code drift = %d after fresh sync", r.Drift)
		}
	}
}

func TestStatusCommandDetectsDrift(t *testing.T) {
	old := flagConfig
	defer func() { flagConfig = old }()
	flagConfig = ".stdai/config.toml"

	tmp := t.TempDir()
	setupSyncProject(t, tmp)
	t.Chdir(tmp)

	syncCmd := newSyncCmd()
	syncCmd.SetOut(new(bytes.Buffer))
	syncCmd.SetErr(new(bytes.Buffer))
	if err := syncCmd.Execute(); err != nil {
		t.Fatal(err)
	}

	// 手动篡改一个生成文件
	if err := os.WriteFile(filepath.Join(tmp, "CLAUDE.md"), []byte("tampered"), 0o600); err != nil {
		t.Fatal(err)
	}

	statusCmd := newStatusCmd()
	var out bytes.Buffer
	statusCmd.SetOut(&out)
	statusCmd.SetErr(new(bytes.Buffer))
	err := statusCmd.Execute()
	if err == nil {
		t.Error("status should error when drift detected")
	}
	if !strings.Contains(out.String(), "drift=1") && !strings.Contains(err.Error(), "drift") {
		t.Errorf("expected drift report in output or error:\nout:%s\nerr:%v", out.String(), err)
	}
}
