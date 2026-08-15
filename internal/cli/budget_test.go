package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBudgetCommandText(t *testing.T) {
	old := flagConfig
	defer func() { flagConfig = old }()
	flagConfig = ".stdai/config.toml"

	tmp := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmp, ".stdai/standards/rules"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmp, ".stdai/config.toml"), []byte(`version = "1.0"`), 0o600); err != nil {
		t.Fatal(err)
	}
	body := "---\ntype: rules\nname: big\n---\n" + strings.Repeat("a", 9000)
	if err := os.WriteFile(filepath.Join(tmp, ".stdai/standards/rules/big.md"), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Chdir(tmp)

	cmd := newBudgetCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(new(bytes.Buffer))
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	got := out.String()
	for _, want := range []string{"rules/big.md", "rules", "SOFT", "BYTES", "~TOKENS"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
}

func TestBudgetCommandJSON(t *testing.T) {
	old := flagConfig
	defer func() { flagConfig = old }()
	flagConfig = ".stdai/config.toml"

	tmp := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmp, ".stdai/standards/rules"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmp, ".stdai/config.toml"), []byte(`version = "1.0"`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmp, ".stdai/standards/rules/x.md"),
		[]byte("---\ntype: rules\nname: x\n---\nbody"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Chdir(tmp)

	cmd := newBudgetCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs([]string{"--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var r budgetReport
	if err := json.Unmarshal(out.Bytes(), &r); err != nil {
		t.Fatalf("not valid JSON: %v\n%s", err, out.String())
	}
	if len(r.Docs) != 1 {
		t.Errorf("got %d docs, want 1", len(r.Docs))
	}
	if r.TotalTokens <= 0 {
		t.Errorf("TotalTokens should be > 0")
	}
	if strings.Contains(out.String(), "rendered_targets") {
		t.Errorf("default JSON should not include rendered targets:\n%s", out.String())
	}
}

func TestBudgetCommandRenderedTargets(t *testing.T) {
	old := flagConfig
	defer func() { flagConfig = old }()
	flagConfig = ".stdai/config.toml"

	tmp := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmp, ".stdai/standards/rules"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(tmp, ".stdai/standards/commands"), 0o755); err != nil {
		t.Fatal(err)
	}
	configBody := `version = "1.0"
inject = false
inject_type_glossary = false

[targets]
codex = { enabled = true, convert = true }
factory = { enabled = true, convert = true }
`
	if err := os.WriteFile(filepath.Join(tmp, ".stdai/config.toml"), []byte(configBody), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmp, ".stdai/standards/root.md"),
		[]byte("# Project\n\nProject overview."), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmp, ".stdai/standards/rules/style.md"),
		[]byte("---\ntype: rules\nname: style\n---\nUse clear names."), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmp, ".stdai/standards/commands/review.md"),
		[]byte("---\ntype: commands\nname: review\n---\nReview the diff."), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Chdir(tmp)

	cmd := newBudgetCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs([]string{"--rendered", "--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var r budgetReport
	if err := json.Unmarshal(out.Bytes(), &r); err != nil {
		t.Fatalf("not valid JSON: %v\n%s", err, out.String())
	}
	if len(r.SourceLayers) != 2 {
		t.Fatalf("source layers = %d, want 2", len(r.SourceLayers))
	}
	if len(r.RenderedTargets) != 2 {
		t.Fatalf("rendered targets = %d, want 2", len(r.RenderedTargets))
	}
	if r.RenderedTargets[0].RootBytes == 0 ||
		r.RenderedTargets[0].RootBytes != r.RenderedTargets[1].RootBytes {
		t.Errorf("shared root bytes should be equal and nonzero: %+v", r.RenderedTargets)
	}
	for _, target := range r.RenderedTargets {
		if len(target.RootFiles) != 1 || target.RootFiles[0].Path != "AGENTS.md" {
			t.Errorf("%s root files = %+v, want shared AGENTS.md", target.Target, target.RootFiles)
		}
	}
}

func TestBudgetCommandNoStandards(t *testing.T) {
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

	cmd := newBudgetCmd()
	var out, errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !strings.Contains(errOut.String(), "no standards") {
		t.Errorf("expected 'no standards' message in stderr: %s", errOut.String())
	}
}

func TestBudgetCommandSkipsSkillSubdirMarkdown(t *testing.T) {
	old := flagConfig
	defer func() { flagConfig = old }()
	flagConfig = ".stdai/config.toml"

	tmp := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmp, ".stdai/standards/skills/foo/refs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmp, ".stdai/config.toml"), []byte(`version = "1.0"`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmp, ".stdai/standards/skills/foo/SKILL.md"),
		[]byte("---\ntype: skills\nname: foo\n---\nbody"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmp, ".stdai/standards/skills/foo/refs/check.md"),
		[]byte("aux"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Chdir(tmp)

	cmd := newBudgetCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs([]string{"--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	var r budgetReport
	if err := json.Unmarshal(out.Bytes(), &r); err != nil {
		t.Fatal(err)
	}
	if len(r.Docs) != 1 {
		t.Errorf("got %d docs, want 1 (refs/check.md should be skipped)", len(r.Docs))
	}
}

func TestIsMarkdownFile(t *testing.T) {
	cases := map[string]bool{
		"a.md":         true,
		"a.MD":         true,
		"a.markdown":   true,
		"A.MARKDOWN":   true,
		"a.txt":        false,
		"a":            false,
		"path/to/x.md": true,
	}
	for in, want := range cases {
		if got := isMarkdownFile(in); got != want {
			t.Errorf("isMarkdownFile(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestIsSkillSubdirMarkdownFile(t *testing.T) {
	cases := map[string]bool{
		"skills/foo/SKILL.md":         false, // 顶层
		"skills/foo/refs/x.md":        true,
		"skills/foo/scripts/sub/y.md": true,
		"rules/x.md":                  false,
		"skills/foo":                  false,
	}
	for in, want := range cases {
		if got := isSkillSubdirMarkdownFile(in); got != want {
			t.Errorf("isSkillSubdirMarkdownFile(%q) = %v, want %v", in, got, want)
		}
	}
}
