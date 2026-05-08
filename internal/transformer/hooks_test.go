package transformer

import (
	"strings"
	"testing"

	"std-ai/internal/config"
	"std-ai/internal/parser"
)

func TestClaudeCodeHooksOutput(t *testing.T) {
	cfg := &config.Config{
		Hooks: &config.HooksConfig{
			Version: "1.0",
			Hooks: map[string][]config.HookEntry{
				"PreToolUse": {{Matcher: "Bash", Type: "command", Command: "echo hi"}},
			},
		},
	}
	tr := &ClaudeCode{}
	plan, err := tr.Plan(nil, cfg)
	if err != nil {
		t.Fatal(err)
	}
	var found string
	for _, f := range plan.Files {
		if f.Path == ".claude/stdagent-hooks.json" {
			found = string(f.Content)
		}
	}
	if found == "" {
		t.Fatal("expected .claude/stdagent-hooks.json in plan")
	}
	for _, want := range []string{`"PreToolUse"`, `"matcher": "Bash"`, `"command": "echo hi"`} {
		if !strings.Contains(found, want) {
			t.Errorf("hooks json missing %q in:\n%s", want, found)
		}
	}
}

func TestCodexHooksOutputWithoutDocs(t *testing.T) {
	cfg := &config.Config{
		Hooks: &config.HooksConfig{
			Version: "1.0",
			Hooks: map[string][]config.HookEntry{
				"PreToolUse": {{Matcher: "Bash", Type: "command", Command: "echo cod"}},
			},
		},
	}
	tr := &Codex{}
	plan, err := tr.Plan(nil, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Files) == 0 {
		t.Fatal("hooks-only plan should still have files")
	}
	var found bool
	for _, f := range plan.Files {
		if f.Path == ".codex/stdagent-hooks.json" {
			found = true
		}
	}
	if !found {
		t.Error("expected .codex/stdagent-hooks.json")
	}
}

func TestHooksEmptyMapNoOutput(t *testing.T) {
	cfg := &config.Config{
		Hooks: &config.HooksConfig{Hooks: map[string][]config.HookEntry{}},
	}
	tr := &ClaudeCode{}
	docs := []*parser.Document{{Type: parser.TypeRules, Name: "x", Body: "y"}}
	plan, _ := tr.Plan(docs, cfg)
	for _, f := range plan.Files {
		if strings.Contains(f.Path, "stdagent-hooks.json") {
			t.Error("empty hooks map should not produce hooks file")
		}
	}
}
