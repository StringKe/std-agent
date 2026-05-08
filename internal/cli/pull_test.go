package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPullCommandNoSourcesConfigured(t *testing.T) {
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

	cmd := newPullCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(new(bytes.Buffer))
	if err := cmd.Execute(); err != nil {
		t.Fatalf("pull: %v", err)
	}
	if !strings.Contains(out.String(), "0 sources pulled") {
		t.Errorf("expected '0 sources pulled': %s", out.String())
	}
}

func TestPullCommandDisabledSourceSkipped(t *testing.T) {
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

[sources.default]
url = "/nonexistent"
enabled = false
paths = ["standards/"]
`), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Chdir(tmp)

	cmd := newPullCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(new(bytes.Buffer))
	if err := cmd.Execute(); err != nil {
		t.Fatalf("pull: %v", err)
	}
	if !strings.Contains(out.String(), "0 sources pulled") {
		t.Errorf("disabled source should be skipped, got: %s", out.String())
	}
}

func TestDefaultBranchHelper(t *testing.T) {
	if defaultBranch("") != "main" {
		t.Error("empty should default to main")
	}
	if defaultBranch("dev") != "dev" {
		t.Error("explicit branch should be returned")
	}
}
