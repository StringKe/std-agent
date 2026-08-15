package transformer

import (
	"strings"
	"testing"

	"github.com/StringKe/std-agent/internal/config"
	"github.com/StringKe/std-agent/internal/parser"
	"github.com/StringKe/std-agent/internal/writer"
)

func TestCanonicalizeSharedAGENTS(t *testing.T) {
	t.Parallel()

	shared := &parser.Document{Path: "rules/shared.md", Type: parser.TypeRules, Name: "shared", Body: "Shared rule."}
	codexOnly := &parser.Document{
		Path: "rules/codex.md", Type: parser.TypeRules, Name: "codex",
		Targets: []string{"codex"}, Body: "Codex rule.",
	}
	other := &parser.Document{
		Path: "rules/other.md", Type: parser.TypeRules, Name: "other",
		Targets: []string{"claude-code"}, Body: "Other rule.",
	}
	plans := []*writer.Plan{
		{Target: "codex", Files: []writer.FileOp{
			{Path: "AGENTS.md", Content: []byte("codex")},
			{Path: ".agents/skills/foo/SKILL.md", Content: []byte("skill")},
		}},
		{Target: "factory", Files: []writer.FileOp{
			{Path: "AGENTS.md", Content: []byte("factory")},
			{Path: ".factory/rules/shared.md", Content: []byte("rule")},
		}},
	}

	if err := CanonicalizeSharedAGENTS(plans, []*parser.Document{shared, codexOnly, other}, &config.Config{}); err != nil {
		t.Fatal(err)
	}

	var first []byte
	for _, plan := range plans {
		var agentsCount int
		for _, op := range plan.Files {
			if op.Path != "AGENTS.md" {
				continue
			}
			agentsCount++
			if first == nil {
				first = op.Content
			} else if string(first) != string(op.Content) {
				t.Fatalf("shared AGENTS.md differs between targets:\n%s\n---\n%s", first, op.Content)
			}
		}
		if agentsCount != 1 {
			t.Fatalf("%s AGENTS.md count = %d, want 1", plan.Target, agentsCount)
		}
	}

	got := string(first)
	for _, want := range []string{"Shared rule.", "Codex rule."} {
		if !strings.Contains(got, want) {
			t.Errorf("canonical AGENTS.md missing %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "Other rule.") {
		t.Errorf("canonical AGENTS.md contains rule for disabled consumer:\n%s", got)
	}
}
