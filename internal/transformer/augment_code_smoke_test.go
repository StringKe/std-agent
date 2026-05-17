package transformer

import (
	"strings"
	"testing"

	"std-ai/internal/config"
	"std-ai/internal/parser"
)

func TestAugmentCodeOutputs(t *testing.T) {
	tr := &AugmentCode{}
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
			if len(plan.Files) != 1 {
				t.Fatalf("expected 1 file, got %d", len(plan.Files))
			}
			c := string(plan.Files[0].Content)
			if !strings.Contains(c, tc.want) {
				t.Errorf("missing %q in:\n%s", tc.want, c)
			}
			if plan.Files[0].Path != ".augment/rules/x.md" {
				t.Errorf("path = %s, want .augment/rules/x.md", plan.Files[0].Path)
			}
		})
	}
}

func TestAugmentCodeFallback(t *testing.T) {
	tr := &AugmentCode{}
	cfg := &config.Config{Inject: false}
	docs := []*parser.Document{
		{Type: parser.TypeSkills, Name: "review", Description: "Review code.", Body: "Steps"},
		{Type: parser.TypeReferences, Name: "api-ref", Description: "API ref", Body: "ref body"},
		{Type: parser.TypeSubagents, Name: "qa", Description: "QA agent", Body: "qa body"},
		{Type: parser.TypeCommands, Name: "deploy", Description: "Deploy", Body: "go"},
	}
	plan, _ := tr.Plan(docs, cfg)
	paths := pathSet(plan)

	// skills 走 Agent Skills 标准 fallback：<RulesDir>/skills/<name>/SKILL.md
	if !paths[".augment/rules/skills/review/SKILL.md"] {
		t.Errorf("missing skills fallback, paths: %v", paths)
	}
	// references 走 <RulesDir>/references/<name>.md
	if !paths[".augment/rules/references/api-ref.md"] {
		t.Errorf("missing references fallback, paths: %v", paths)
	}
	// subagents 走 <RulesDir>/subagents/<name>.md
	if !paths[".augment/rules/subagents/qa.md"] {
		t.Errorf("missing subagents fallback, paths: %v", paths)
	}
	// commands 走 .augment/rules/workflows/<name>.md
	if !paths[".augment/rules/workflows/deploy.md"] {
		t.Errorf("missing workflow, paths: %v", paths)
	}

	// 验证 fallback 文件含 std-ai-type frontmatter
	refContent, _ := contentOf(plan, ".augment/rules/references/api-ref.md")
	if !strings.Contains(refContent, "std-ai-type: references") {
		t.Errorf("references fallback missing std-ai-type:\n%s", refContent)
	}
	subContent, _ := contentOf(plan, ".augment/rules/subagents/qa.md")
	if !strings.Contains(subContent, "std-ai-type: subagents") {
		t.Errorf("subagents fallback missing std-ai-type:\n%s", subContent)
	}
}
