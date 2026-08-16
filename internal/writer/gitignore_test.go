package writer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/StringKe/std-agent/internal/config"
)

func TestGitignoreEntriesGeneratedIncludesShared(t *testing.T) {
	got := GitignoreEntries(config.GitignoreGenerated, []string{"claude-code", "codex"}, nil)
	for _, want := range []string{
		".stdai/state.json",
		"CLAUDE.md",
		".claude/",
		"AGENTS.md",
		".agents/",
		".claude/settings.local.json",
	} {
		if !containsStr(got, want) {
			t.Errorf("generated missing %q in %v", want, got)
		}
	}
	if containsStr(got, "crush.json") || containsStr(got, "kilo.jsonc") {
		t.Errorf("must not ignore user JSON merge files: %v", got)
	}
}

func TestGitignoreEntriesPortableKeepsAgents(t *testing.T) {
	got := GitignoreEntries(config.GitignorePortable, []string{"claude-code", "codex", "goose"}, nil)
	if containsStr(got, "AGENTS.md") || containsStr(got, ".agents/") {
		t.Errorf("portable must keep AGENTS.md and .agents/: %v", got)
	}
	for _, want := range []string{"CLAUDE.md", ".claude/", ".codex/", ".goose/", ".stdai/backups/"} {
		if !containsStr(got, want) {
			t.Errorf("portable missing %q in %v", want, got)
		}
	}
}

func TestGitignoreEntriesOffEmpty(t *testing.T) {
	if got := GitignoreEntries(config.GitignoreOff, []string{"codex"}, nil); len(got) != 0 {
		t.Errorf("off must be empty, got %v", got)
	}
}

func TestGitignoreEntriesCollapsesPlanPaths(t *testing.T) {
	plans := []*Plan{{
		Target: "claude-code",
		Files: []FileOp{
			{Path: ".claude/rules/style.md"},
			{Path: "frontend/CLAUDE.md"},
			{Path: ".github/instructions/go.instructions.md"},
			{Path: ".vscode/mcp.json"},
			{Path: "crush.json", JSONMerge: true},
		},
	}}
	got := GitignoreEntries(config.GitignoreGenerated, nil, plans)
	for _, want := range []string{".claude/", "CLAUDE.md", ".github/instructions/", ".vscode/mcp.json"} {
		if !containsStr(got, want) {
			t.Errorf("expected %q from plan, got %v", want, got)
		}
	}
	if containsStr(got, ".github/") {
		t.Error("must not ignore entire .github/")
	}
	if containsStr(got, "crush.json") {
		t.Error("JSONMerge path must not be ignored")
	}
}

func TestGitignoreEntriesPortableSkipsPlanShared(t *testing.T) {
	plans := []*Plan{{
		Files: []FileOp{
			{Path: "AGENTS.md"},
			{Path: ".agents/skills/x/SKILL.md"},
			{Path: "frontend/AGENTS.md"},
		},
	}}
	got := GitignoreEntries(config.GitignorePortable, []string{"codex"}, plans)
	if containsStr(got, "AGENTS.md") || containsStr(got, ".agents/") {
		t.Errorf("portable must keep AGENTS.md and .agents/ even from plan: %v", got)
	}
}

func TestGitignorePrefixesCoverAllTargets(t *testing.T) {
	for _, name := range config.ValidTargets {
		if _, ok := targetIgnorePrefixes[name]; !ok {
			t.Errorf("target %s missing gitignore prefixes", name)
		}
	}
}

func TestUpsertGitignoreCreatesAndUpdatesBlock(t *testing.T) {
	tmp := t.TempDir()
	gi := filepath.Join(tmp, ".gitignore")
	if err := os.WriteFile(gi, []byte("bin/\n.stdai/cache/\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	changed, err := UpsertGitignore(tmp, []string{".stdai/cache/", "CLAUDE.md"}, false)
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("first upsert should change")
	}
	raw, _ := os.ReadFile(gi)
	s := string(raw)
	if !strings.Contains(s, gitignoreBegin) || !strings.Contains(s, "CLAUDE.md") {
		t.Fatalf("missing managed block:\n%s", s)
	}
	if strings.Count(s, ".stdai/cache/") != 1 {
		t.Fatalf("duplicate cache line:\n%s", s)
	}
	if !strings.Contains(s, "bin/") {
		t.Fatalf("user line dropped:\n%s", s)
	}
	changed, err = UpsertGitignore(tmp, []string{".stdai/cache/", "CLAUDE.md"}, false)
	if err != nil {
		t.Fatal(err)
	}
	if changed {
		t.Fatal("second upsert should be unchanged")
	}
}

func TestUpsertGitignoreEmptyIdempotent(t *testing.T) {
	tmp := t.TempDir()
	changed, err := UpsertGitignore(tmp, []string{"CLAUDE.md"}, false)
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("create on missing .gitignore should change")
	}
	changed, err = UpsertGitignore(tmp, []string{"CLAUDE.md"}, false)
	if err != nil {
		t.Fatal(err)
	}
	if changed {
		raw, _ := os.ReadFile(filepath.Join(tmp, ".gitignore"))
		t.Fatalf("empty-file upsert must be idempotent:\n%s", raw)
	}
}

func TestUpsertGitignoreDryRun(t *testing.T) {
	tmp := t.TempDir()
	changed, err := UpsertGitignore(tmp, []string{"AGENTS.md"}, true)
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("dry-run should report change")
	}
	if _, err := os.Stat(filepath.Join(tmp, ".gitignore")); !os.IsNotExist(err) {
		t.Fatal("dry-run must not write .gitignore")
	}
}

func containsStr(items []string, want string) bool {
	for _, s := range items {
		if s == want {
			return true
		}
	}
	return false
}
