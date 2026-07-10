package transformer

import (
	"testing"

	"github.com/StringKe/std-agent/internal/config"
	"github.com/StringKe/std-agent/internal/parser"
)

func TestCopilotRulesSplit(t *testing.T) {
	tr := &Copilot{}
	cfg := &config.Config{Inject: false}
	docs := []*parser.Document{
		{Type: parser.TypeRules, Name: "general-a", Body: "a"},
		{Type: parser.TypeRules, Name: "ts-only", ApplyTo: []string{"**/*.ts"}, Body: "ts"},
	}
	plan, _ := tr.Plan(docs, cfg)
	paths := pathSet(plan)
	if !paths[".github/copilot-instructions.md"] {
		t.Errorf("missing copilot-instructions.md, paths: %v", paths)
	}
	if !paths[".github/instructions/ts-only.instructions.md"] {
		t.Errorf("missing ts-only.instructions.md, paths: %v", paths)
	}
}
