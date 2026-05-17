package transformer

import (
	"strings"
	"testing"

	"std-ai/internal/config"
	"std-ai/internal/parser"
)

func TestOpenCodeSkillsAndCommands(t *testing.T) {
	tr := &OpenCode{}
	cfg := &config.Config{Inject: false}
	docs := []*parser.Document{
		{Type: parser.TypeSkills, Name: "review", Description: "Review", Body: "steps"},
		{Type: parser.TypeCommands, Name: "deploy", Description: "Deploy", Body: "go"},
		{Type: parser.TypeRules, Name: "ignored", Body: "rules go to AGENTS.md instead"},
	}
	plan, _ := tr.Plan(docs, cfg)
	paths := pathSet(plan)
	if !paths[".opencode/agents/review.md"] {
		t.Errorf("missing agent file, paths: %v", paths)
	}
	if !paths[".opencode/commands/deploy.md"] {
		t.Errorf("missing command file, paths: %v", paths)
	}
	for p := range paths {
		if strings.HasPrefix(p, "AGENTS") || strings.HasPrefix(p, ".opencode/rules") {
			t.Errorf("opencode should not write rules, got %s", p)
		}
	}
}
