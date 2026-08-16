package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadValid(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "config.toml")
	content := `version = "1.0"
name = "test"
inject = true
inject_whatis = true

[targets]
claude-code = { enabled = true, convert = true }
codex = { enabled = true, convert = true }

[sources.default]
url = "https://github.com/foo/bar.git"
branch = "main"
enabled = true
paths = ["standards/"]
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if cfg.Version != "1.0" {
		t.Errorf("Version = %s", cfg.Version)
	}
	if !cfg.Targets["claude-code"].Enabled {
		t.Error("claude-code should be enabled")
	}
	if cfg.Sources["default"].URL == "" {
		t.Error("source URL missing")
	}
	if cfg.BackupKeep != 1 {
		t.Errorf("BackupKeep = %d, want 1 (normalized)", cfg.BackupKeep)
	}
	if cfg.Gitignore != GitignoreGenerated {
		t.Errorf("empty gitignore should normalize to generated, got %q", cfg.Gitignore)
	}
}

func TestLoadEnvOverride(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "config.toml")
	content := `version = "1.0"
auto_pull = true

[targets]

[sources.default]
url = "https://x.com/y.git"
paths = ["standards/"]
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("STDAI_DRY_RUN", "1")
	t.Setenv("STDAI_VERBOSE", "1")
	t.Setenv("STDAI_NO_PULL", "1")
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !cfg.DryRun {
		t.Error("DryRun should be true via env")
	}
	if !cfg.Verbose {
		t.Error("Verbose should be true via env")
	}
	if cfg.AutoPull {
		t.Error("AutoPull should be false via STDAI_NO_PULL=1")
	}
}

func TestValidateUnknownTarget(t *testing.T) {
	cfg := &Config{
		Version: "1.0",
		Targets: map[string]TargetConfig{
			"unknown-tool": {Enabled: true},
		},
	}
	if err := Validate(cfg); err == nil {
		t.Error("expected error on unknown target")
	}
}

func TestValidateInvalidVersion(t *testing.T) {
	cfg := &Config{Version: "9.9"}
	if err := Validate(cfg); err == nil {
		t.Error("expected error on invalid version")
	}
}

func TestValidateSourceMissingURL(t *testing.T) {
	cfg := &Config{
		Version: "1.0",
		Sources: map[string]SourceConfig{
			"x": {Paths: []string{"a/"}},
		},
	}
	if err := Validate(cfg); err == nil {
		t.Error("expected error on missing source URL")
	}
}

func TestValidateHTTPSTokenMissingEnv(t *testing.T) {
	cfg := &Config{
		Version: "1.0",
		Sources: map[string]SourceConfig{
			"x": {URL: "https://...", Paths: []string{"a/"}, Auth: "https-token"},
		},
	}
	if err := Validate(cfg); err == nil {
		t.Error("expected error on https-token missing token_env")
	}
}

func TestDefault(t *testing.T) {
	cfg := Default()
	if cfg.Version != "1.0" {
		t.Errorf("Version = %s", cfg.Version)
	}
	if !cfg.Targets["claude-code"].Enabled {
		t.Error("claude-code should default enabled")
	}
	if cfg.Targets["cursor"].Enabled {
		t.Error("cursor should default disabled")
	}
	if cfg.InjectTypeGlossary {
		t.Error("InjectTypeGlossary should default false")
	}
	if cfg.Gitignore != GitignoreGenerated {
		t.Errorf("Gitignore default = %q, want generated", cfg.Gitignore)
	}
}

func TestLoadInjectTypeGlossaryOverride(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "config.toml")
	content := `version = "1.0"
inject_type_glossary = false

[targets]

[sources.default]
url = "https://x.com/y.git"
paths = ["standards/"]
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.InjectTypeGlossary {
		t.Error("InjectTypeGlossary should be false after toml override")
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "config.toml")

	cfg := Default()
	if err := Save(path, cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.Version != cfg.Version {
		t.Errorf("Version: %s vs %s", loaded.Version, cfg.Version)
	}
	if len(loaded.Targets) != len(cfg.Targets) {
		t.Errorf("Targets count: %d vs %d", len(loaded.Targets), len(cfg.Targets))
	}
	if loaded.Inject != cfg.Inject {
		t.Errorf("Inject: %v vs %v", loaded.Inject, cfg.Inject)
	}
	if !loaded.Targets["claude-code"].Enabled {
		t.Error("claude-code should be enabled after round-trip")
	}
	if loaded.Gitignore != GitignoreGenerated {
		t.Errorf("Gitignore after round-trip = %q, want generated", loaded.Gitignore)
	}
}

func TestValidateInvalidGitignore(t *testing.T) {
	cfg := &Config{Version: "1.0", Gitignore: "all"}
	if err := Validate(cfg); err == nil {
		t.Error("expected error on invalid gitignore")
	}
}

func TestValidateNormalizesEmptyGitignore(t *testing.T) {
	cfg := &Config{Version: "1.0"}
	if err := Validate(cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.Gitignore != GitignoreGenerated {
		t.Errorf("empty gitignore should become generated, got %q", cfg.Gitignore)
	}
}

func TestValidateNormalizesBackupKeep(t *testing.T) {
	cfg := &Config{Version: "1.0", BackupKeep: 0}
	if err := Validate(cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.BackupKeep != 1 {
		t.Errorf("BackupKeep should normalize to 1, got %d", cfg.BackupKeep)
	}
}

func TestValidateInvalidAuth(t *testing.T) {
	cfg := &Config{
		Version: "1.0",
		Sources: map[string]SourceConfig{
			"x": {URL: "https://x.com/y.git", Paths: []string{"a/"}, Auth: "weird"},
		},
	}
	if err := Validate(cfg); err == nil {
		t.Error("expected error on invalid auth")
	}
}

func TestIsValidTarget(t *testing.T) {
	if !IsValidTarget("claude-code") {
		t.Error("claude-code should be valid")
	}
	if IsValidTarget("foo") {
		t.Error("foo should be invalid")
	}
}
