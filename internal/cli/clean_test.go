package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCleanCommandRemovesGeneratedFiles(t *testing.T) {
	old := flagConfig
	defer func() { flagConfig = old }()
	flagConfig = ".stdai/config.toml"

	tmp := t.TempDir()
	setupSyncProject(t, tmp)
	t.Chdir(tmp)

	// first sync to generate CLAUDE.md / AGENTS.md / .claude/rules/...
	syncCmd := newSyncCmd()
	syncCmd.SetOut(new(bytes.Buffer))
	syncCmd.SetErr(new(bytes.Buffer))
	if err := syncCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(tmp, "CLAUDE.md")); err != nil {
		t.Fatal("CLAUDE.md should exist after sync")
	}

	// clean --yes skips interactive confirmation
	cleanCmd := newCleanCmd()
	var out bytes.Buffer
	cleanCmd.SetOut(&out)
	cleanCmd.SetErr(new(bytes.Buffer))
	cleanCmd.SetArgs([]string{"--yes"})
	if err := cleanCmd.Execute(); err != nil {
		t.Fatalf("clean: %v", err)
	}

	// generated files should be removed
	for _, gone := range []string{"CLAUDE.md", "AGENTS.md", ".claude/rules/style.md"} {
		if _, err := os.Stat(filepath.Join(tmp, gone)); err == nil {
			t.Errorf("%s should be removed after clean", gone)
		}
	}
	// .stdai/ should be preserved
	if _, err := os.Stat(filepath.Join(tmp, ".stdai/config.toml")); err != nil {
		t.Errorf(".stdai should be preserved: %v", err)
	}
	if !strings.Contains(out.String(), "Removed") {
		t.Errorf("expected 'Removed' message: %s", out.String())
	}
}

func TestCleanCommandNothingToClean(t *testing.T) {
	old := flagConfig
	defer func() { flagConfig = old }()
	flagConfig = ".stdai/config.toml"

	tmp := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmp, ".stdai"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmp, ".stdai/config.toml"), []byte(`version = "1.0"
[targets]
claude-code = { enabled = true, convert = true }
`), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Chdir(tmp)

	cleanCmd := newCleanCmd()
	var out bytes.Buffer
	cleanCmd.SetOut(&out)
	cleanCmd.SetErr(new(bytes.Buffer))
	cleanCmd.SetArgs([]string{"--yes"})
	if err := cleanCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "nothing to clean") {
		t.Errorf("expected 'nothing to clean' in: %s", out.String())
	}
}

func TestCleanCommandTargetFilter(t *testing.T) {
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

	// only clean claude-code
	cleanCmd := newCleanCmd()
	cleanCmd.SetOut(new(bytes.Buffer))
	cleanCmd.SetErr(new(bytes.Buffer))
	cleanCmd.SetArgs([]string{"--target", "claude-code", "--yes"})
	if err := cleanCmd.Execute(); err != nil {
		t.Fatal(err)
	}

	// CLAUDE.md is removed
	if _, err := os.Stat(filepath.Join(tmp, "CLAUDE.md")); err == nil {
		t.Error("CLAUDE.md should be removed (target=claude-code)")
	}
	// AGENTS.md stays (target limited to claude-code)
	if _, err := os.Stat(filepath.Join(tmp, "AGENTS.md")); err != nil {
		t.Errorf("AGENTS.md should remain when --target claude-code: %v", err)
	}
}
