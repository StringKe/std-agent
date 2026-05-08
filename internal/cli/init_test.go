package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitCommandCreatesStdaiTree(t *testing.T) {
	tmp := t.TempDir()
	t.Chdir(tmp)

	cmd := newInitCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(new(bytes.Buffer))
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	for _, want := range []string{
		".stdai/config.toml",
		".stdai/standards/rules/.gitkeep",
		".stdai/standards/skills/.gitkeep",
		".stdai/standards/commands/.gitkeep",
		".stdai/standards/references/.gitkeep",
		".stdai/cache/.gitkeep",
		".stdai/backups/.gitkeep",
		".stdai/standards/rules/example.md",
		".stdai/standards/skills/code-review/SKILL.md",
	} {
		full := filepath.Join(tmp, want)
		if _, err := os.Stat(full); err != nil {
			t.Errorf("missing %s: %v", want, err)
		}
	}
	gi, err := os.ReadFile(filepath.Join(tmp, ".gitignore"))
	if err != nil {
		t.Fatalf("missing .gitignore: %v", err)
	}
	if !strings.Contains(string(gi), ".stdai/cache/") {
		t.Error(".gitignore missing .stdai/cache/")
	}
}

func TestInitCommandFailsIfExistsWithoutForce(t *testing.T) {
	tmp := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmp, ".stdai"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(tmp)

	cmd := newInitCmd()
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))
	if err := cmd.Execute(); err == nil {
		t.Error("expected error when .stdai exists without --force")
	}
}

func TestInitCommandForceBackupsOld(t *testing.T) {
	tmp := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmp, ".stdai"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmp, ".stdai/marker.txt"), []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Chdir(tmp)

	cmd := newInitCmd()
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs([]string{"--force"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute --force: %v", err)
	}

	entries, _ := os.ReadDir(tmp)
	backupFound := false
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".stdai-backup-") && e.IsDir() {
			backupFound = true
		}
	}
	if !backupFound {
		t.Error("--force should create timestamped backup directory")
	}
}

func TestInitCommandMinimal(t *testing.T) {
	tmp := t.TempDir()
	t.Chdir(tmp)

	cmd := newInitCmd()
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs([]string{"--minimal"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute --minimal: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tmp, ".stdai/standards/rules/example.md")); err == nil {
		t.Error("--minimal should not write example.md")
	}
	if _, err := os.Stat(filepath.Join(tmp, ".stdai/config.toml")); err != nil {
		t.Errorf("config.toml should still be created with --minimal: %v", err)
	}
}

func TestInitCommandSourceURL(t *testing.T) {
	tmp := t.TempDir()
	t.Chdir(tmp)

	cmd := newInitCmd()
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs([]string{"--source", "https://github.com/foo/bar.git"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute --source: %v", err)
	}
	cfg, err := os.ReadFile(filepath.Join(tmp, ".stdai/config.toml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(cfg), "https://github.com/foo/bar.git") {
		t.Errorf("config.toml missing source URL:\n%s", cfg)
	}
}
