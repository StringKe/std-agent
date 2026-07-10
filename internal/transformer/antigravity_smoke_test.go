package transformer

import (
	"strings"
	"testing"

	"github.com/StringKe/std-agent/internal/config"
	"github.com/StringKe/std-agent/internal/parser"
)

func TestAntigravityRuleTriggers(t *testing.T) {
	tr := &Antigravity{}
	cfg := &config.Config{Inject: false}
	cases := []struct {
		name string
		doc  *parser.Document
		want string
	}{
		{"always", &parser.Document{Type: parser.TypeRules, Name: "x", AlwaysApply: true, Body: "b"}, "trigger: always_on"},
		{"glob", &parser.Document{Type: parser.TypeRules, Name: "x", ApplyTo: []string{"**/*.go"}, Body: "b"}, "trigger: glob"},
		{"model-decision", &parser.Document{Type: parser.TypeRules, Name: "x", Description: "use", Body: "b"}, "trigger: model_decision"},
		{"manual", &parser.Document{Type: parser.TypeRules, Name: "x", Body: "b"}, "trigger: manual"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			plan, _ := tr.Plan([]*parser.Document{tc.doc}, cfg)
			c := string(plan.Files[0].Content)
			if !strings.Contains(c, tc.want) {
				t.Errorf("missing %q in:\n%s", tc.want, c)
			}
			if plan.Files[0].Path != ".agents/rules/x.md" {
				t.Errorf("path = %s, want .agents/rules/x.md", plan.Files[0].Path)
			}
		})
	}
}

func TestAntigravityWorkflowAndSkill(t *testing.T) {
	tr := &Antigravity{}
	cfg := &config.Config{Inject: false}
	docs := []*parser.Document{
		{Type: parser.TypeSkills, Name: "review", Description: "Review", Body: "steps"},
		{Type: parser.TypeCommands, Name: "deploy", Description: "Deploy", Body: "go"},
	}
	plan, _ := tr.Plan(docs, cfg)
	paths := pathSet(plan)
	// v3：skill 走 Agent Skills 标准 fallback（子目录隔离，无 std-agent 私有前缀）
	if !paths[".agents/rules/skills/review/SKILL.md"] {
		t.Errorf("missing skill fallback (Agent Skills standard path), paths: %v", paths)
	}
	if !paths[".agents/workflows/deploy.md"] {
		t.Errorf("missing workflow, paths: %v", paths)
	}
}
