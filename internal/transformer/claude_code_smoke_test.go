package transformer

import (
	"strings"
	"testing"

	"github.com/StringKe/std-agent/internal/config"
	"github.com/StringKe/std-agent/internal/parser"
)

func TestClaudeCodeOutputs(t *testing.T) {
	tr := &ClaudeCode{}
	cfg := &config.Config{Inject: true, InjectWhatIs: false}
	docs := []*parser.Document{
		{Type: parser.TypeRules, Name: "style", Body: "rules"},
	}
	plan, _ := tr.Plan(docs, cfg)
	paths := pathSet(plan)
	for _, want := range []string{"CLAUDE.md", ".claude/rules/style.md"} {
		if !paths[want] {
			t.Errorf("missing %s in plan: %v", want, paths)
		}
	}
}

func TestClaudeCodeSubagentOutput(t *testing.T) {
	tr := &ClaudeCode{}
	cfg := &config.Config{Inject: false}
	docs := []*parser.Document{
		{Type: parser.TypeSubagents, Name: "code-reviewer", Description: "Reviews code", Model: "claude-sonnet-4-5", AllowedTools: []string{"Read", "Grep"}, Body: "You are a code reviewer..."},
	}
	plan, _ := tr.Plan(docs, cfg)
	c, ok := contentOf(plan, ".claude/agents/code-reviewer.md")
	if !ok {
		t.Fatalf("expected .claude/agents/code-reviewer.md, paths: %v", pathSet(plan))
	}
	for _, want := range []string{"name: code-reviewer", "description: Reviews code", "model: claude-sonnet-4-5", "Read", "Grep", "You are a code reviewer"} {
		if !strings.Contains(c, want) {
			t.Errorf("missing %q in subagent file:\n%s", want, c)
		}
	}
}
