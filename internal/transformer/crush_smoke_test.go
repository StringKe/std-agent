package transformer

import (
	"strings"
	"testing"

	"std-ai/internal/config"
	"std-ai/internal/parser"
)

func TestCrushOutputs(t *testing.T) {
	tr := &Crush{}
	cfg := &config.Config{Inject: true, InjectWhatIs: false}
	docs := []*parser.Document{
		{Type: parser.TypeRules, Name: "naming", Description: "命名规范", Body: "naming body"},
		{Type: parser.TypeSkills, Name: "code-review", Description: "Review code.", Body: "Steps..."},
	}
	plan, err := tr.Plan(docs, cfg)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	paths := pathSet(plan)
	if !paths["CRUSH.md"] {
		t.Errorf("missing CRUSH.md, paths: %v", paths)
	}
	if !paths[".crush/skills/code-review/SKILL.md"] {
		t.Errorf("missing .crush/skills/code-review/SKILL.md, paths: %v", paths)
	}
	// RulesDir 为空 -> nonRoot rule 应 inline 到 CRUSH.md，不 fan-out 到子目录
	root, _ := contentOf(plan, "CRUSH.md")
	for _, want := range []string{"naming body", "命名规范"} {
		if !strings.Contains(root, want) {
			t.Errorf("CRUSH.md missing %q:\n%s", want, root)
		}
	}
	skill, _ := contentOf(plan, ".crush/skills/code-review/SKILL.md")
	for _, want := range []string{"name: code-review", "Review code."} {
		if !strings.Contains(skill, want) {
			t.Errorf("SKILL.md missing %q:\n%s", want, skill)
		}
	}
}

func TestCrushFallbackPaths(t *testing.T) {
	tr := &Crush{}
	cfg := &config.Config{Inject: false}
	docs := []*parser.Document{
		{Type: parser.TypeCommands, Name: "deploy", Description: "Deploy", Body: "go"},
		{Type: parser.TypeReferences, Name: "api-spec", Description: "API spec", Body: "ref body"},
		{Type: parser.TypeSubagents, Name: "reviewer", Description: "Reviewer", Body: "agent body"},
	}
	plan, err := tr.Plan(docs, cfg)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	paths := pathSet(plan)
	// v3 修订：references 走独立 references/ 子目录（不再借 SkillsDir SKILL.md 路径，
	// 那会让 crush 按 skill 触发逻辑加载 references 破坏"按需查阅"语义）
	for _, want := range []string{
		".crush/rules/commands/deploy.md",
		".crush/rules/references/api-spec.md",
		".crush/rules/subagents/reviewer.md",
	} {
		if !paths[want] {
			t.Errorf("missing fallback path %s, paths: %v", want, paths)
		}
	}
}
