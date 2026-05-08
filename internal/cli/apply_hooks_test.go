package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestApplyHooksClaudeFreshSettings(t *testing.T) {
	old := flagConfig
	defer func() { flagConfig = old }()
	flagConfig = ".stdai/config.toml"

	tmp := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmp, ".stdai"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmp, ".stdai/config.toml"), []byte(`version = "1.0"`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(tmp, ".claude"), 0o755); err != nil {
		t.Fatal(err)
	}
	hooks := `{"hooks":{"PreToolUse":[{"matcher":"Bash","type":"command","command":"echo pre"}]}}`
	if err := os.WriteFile(filepath.Join(tmp, ".claude/stdagent-hooks.json"), []byte(hooks), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Chdir(tmp)

	cmd := newApplyHooksCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(new(bytes.Buffer))
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	settingsData, err := os.ReadFile(filepath.Join(tmp, ".claude/settings.json"))
	if err != nil {
		t.Fatal(err)
	}
	var settings map[string]json.RawMessage
	if err := json.Unmarshal(settingsData, &settings); err != nil {
		t.Fatalf("settings not valid JSON: %v\n%s", err, settingsData)
	}
	if _, ok := settings["hooks"]; !ok {
		t.Errorf("settings.json should contain hooks: %s", settingsData)
	}
}

func TestApplyHooksClaudePreservesExisting(t *testing.T) {
	old := flagConfig
	defer func() { flagConfig = old }()
	flagConfig = ".stdai/config.toml"

	tmp := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmp, ".stdai"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmp, ".stdai/config.toml"), []byte(`version = "1.0"`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(tmp, ".claude"), 0o755); err != nil {
		t.Fatal(err)
	}
	// 用户已有 settings.json 含 model 字段
	existing := `{"model":"claude-sonnet-4-6","permissions":{"allow":["Read"]}}`
	if err := os.WriteFile(filepath.Join(tmp, ".claude/settings.json"), []byte(existing), 0o600); err != nil {
		t.Fatal(err)
	}
	hooks := `{"hooks":{"PostToolUse":[{"matcher":"Edit","type":"command","command":"echo post"}]}}`
	if err := os.WriteFile(filepath.Join(tmp, ".claude/stdagent-hooks.json"), []byte(hooks), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Chdir(tmp)

	cmd := newApplyHooksCmd()
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	settingsData, _ := os.ReadFile(filepath.Join(tmp, ".claude/settings.json"))
	s := string(settingsData)
	for _, want := range []string{"claude-sonnet-4-6", "permissions", "PostToolUse", "echo post"} {
		if !strings.Contains(s, want) {
			t.Errorf("missing %q in merged settings:\n%s", want, s)
		}
	}
}

func TestApplyHooksMissingFile(t *testing.T) {
	old := flagConfig
	defer func() { flagConfig = old }()
	flagConfig = ".stdai/config.toml"

	tmp := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmp, ".stdai"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmp, ".stdai/config.toml"), []byte(`version = "1.0"`), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Chdir(tmp)

	cmd := newApplyHooksCmd()
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got %v", err)
	}
}

func TestApplyHooksUnsupportedTarget(t *testing.T) {
	old := flagConfig
	defer func() { flagConfig = old }()
	flagConfig = ".stdai/config.toml"

	tmp := t.TempDir()
	t.Chdir(tmp)

	cmd := newApplyHooksCmd()
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs([]string{"--target", "unknown"})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "unsupported target") {
		t.Errorf("expected unsupported target error, got %v", err)
	}
}
