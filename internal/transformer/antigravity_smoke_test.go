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
	// skills 原生 .agents/skills/（antigravity.google/docs/skills，workspace 固定路径）
	if !paths[".agents/skills/review/SKILL.md"] {
		t.Errorf("missing native skill path, paths: %v", paths)
	}
	if paths[".agents/rules/skills/review/SKILL.md"] {
		t.Errorf("stale degraded skill path still produced, paths: %v", paths)
	}
	if !paths[".agents/workflows/deploy.md"] {
		t.Errorf("missing workflow, paths: %v", paths)
	}
}

func TestAntigravityNativeSubagent(t *testing.T) {
	tr := &Antigravity{}
	cfg := &config.Config{Inject: false}
	plan, _ := tr.Plan([]*parser.Document{
		{Type: parser.TypeSubagents, Name: "reviewer", Description: "Review", Body: "body"},
	}, cfg)
	if !pathSet(plan)[".agents/agents/reviewer.md"] {
		t.Errorf("missing native subagent, paths: %v", pathSet(plan))
	}
}

// TestAntigravityCodexSkillByteIdentical 保证 antigravity 与 codex 对同一 skill
// 产出字节一致（共享 .agents/skills/ 落点，writer unchanged 去重）。
func TestAntigravityCodexSkillByteIdentical(t *testing.T) {
	cfg := &config.Config{Inject: true, InjectWhatIs: false}
	docs := []*parser.Document{
		{Type: parser.TypeSkills, Name: "review", Description: "Review", WhenToUse: "on review", Body: "steps"},
	}
	agPlan, _ := (&Antigravity{}).Plan(docs, cfg)
	codexPlan, _ := (&Codex{}).Plan(docs, cfg)
	a, aok := contentOf(agPlan, ".agents/skills/review/SKILL.md")
	c, cok := contentOf(codexPlan, ".agents/skills/review/SKILL.md")
	if !aok || !cok {
		t.Fatalf("both targets must produce the shared skill (antigravity=%v codex=%v)", aok, cok)
	}
	if a != c {
		t.Errorf("antigravity and codex output differ for shared path:\nantigravity:\n%s\ncodex:\n%s", a, c)
	}
}
