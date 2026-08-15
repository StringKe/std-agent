package transformer

import (
	"strings"
	"testing"

	"github.com/StringKe/std-agent/internal/config"
	"github.com/StringKe/std-agent/internal/parser"
)

func TestWindsurfRuleTriggers(t *testing.T) {
	tr := &Windsurf{}
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
			// rules 双写：.windsurf/rules/ + .devin/rules/ 镜像
			if len(plan.Files) != 2 {
				t.Fatalf("expected 2 files (windsurf + devin mirror), got %d", len(plan.Files))
			}
			c, ok := contentOf(plan, ".windsurf/rules/x.md")
			if !ok {
				t.Fatalf("missing .windsurf/rules/x.md, paths: %v", pathSet(plan))
			}
			if !strings.Contains(c, tc.want) {
				t.Errorf("missing %q in:\n%s", tc.want, c)
			}
		})
	}
}

// TestWindsurfDevinMirror 验证 .windsurf/rules/ 产物镜像到 .devin/rules/（字节一致）
func TestWindsurfDevinMirror(t *testing.T) {
	tr := &Windsurf{}
	cfg := &config.Config{Inject: false}
	docs := []*parser.Document{
		{Type: parser.TypeRules, Name: "style", Description: "Style", Body: "rule body"},
		{Type: parser.TypeSkills, Name: "review", Description: "Review", Body: "skill body"},
	}
	plan, err := tr.Plan(docs, cfg)
	if err != nil {
		t.Fatal(err)
	}
	ws, wok := contentOf(plan, ".windsurf/rules/style.md")
	dv, dok := contentOf(plan, ".devin/rules/style.md")
	if !wok || !dok {
		t.Fatalf("expected rule in both .windsurf and .devin, paths: %v", pathSet(plan))
	}
	if ws != dv {
		t.Errorf("devin mirror must be byte-identical:\nwindsurf:\n%s\ndevin:\n%s", ws, dv)
	}
	// skills 不镜像（devin 迁移只涉及 rules 目录）
	if pathSet(plan)[".devin/skills/review/SKILL.md"] {
		t.Errorf("skills should not be mirrored to .devin, paths: %v", pathSet(plan))
	}
}

func TestWindsurfNativeSubagent(t *testing.T) {
	tr := &Windsurf{}
	cfg := &config.Config{Inject: false}
	plan, _ := tr.Plan([]*parser.Document{
		{Type: parser.TypeSubagents, Name: "reviewer", Description: "Review", Body: "body"},
	}, cfg)
	if !pathSet(plan)[".devin/agents/reviewer.md"] {
		t.Errorf("missing Devin Local native subagent, paths: %v", pathSet(plan))
	}
}
