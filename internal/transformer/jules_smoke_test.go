package transformer

import (
	"strings"
	"testing"

	"github.com/StringKe/std-agent/internal/config"
	"github.com/StringKe/std-agent/internal/parser"
)

func TestJulesOutputs(t *testing.T) {
	tr := &Jules{}
	cfg := &config.Config{Inject: true, InjectWhatIs: false, InjectTypeGlossary: true}
	docs := []*parser.Document{
		{Type: parser.TypeRules, Name: "root", Root: true, Body: "# Project AGENTS"},
		{Type: parser.TypeRules, Name: "naming", Description: "命名规范", Body: "naming body"},
	}
	plan, err := tr.Plan(docs, cfg)
	if err != nil {
		t.Fatal(err)
	}
	main, ok := contentOf(plan, "AGENTS.md")
	if !ok {
		t.Fatalf("expected AGENTS.md, paths: %v", pathSet(plan))
	}
	// inline rules（无 RulesDir，nonRoot rule body 直接拼进 AGENTS.md）
	if !strings.Contains(main, "Project AGENTS") {
		t.Errorf("root rule body missing in AGENTS.md:\n%s", main)
	}
	if !strings.Contains(main, "naming body") {
		t.Errorf("nonRoot rule should be inlined into AGENTS.md:\n%s", main)
	}
	// glossary 注入
	if !strings.Contains(main, "std-agent 类型速查") {
		t.Errorf("glossary missing in AGENTS.md:\n%s", main)
	}
	// 不应存在 .jules/rules 下的独立 rule 文件（RulesDir=""，nonRoot 全 inline）
	for _, f := range plan.Files {
		if strings.HasPrefix(f.Path, ".jules/rules/") && !strings.Contains(f.Path, "/skills/") &&
			!strings.Contains(f.Path, "/commands/") && !strings.Contains(f.Path, "/references/") &&
			!strings.Contains(f.Path, "/subagents/") {
			t.Errorf("unexpected rule fan-out: %s", f.Path)
		}
	}
}

func TestJulesFallback(t *testing.T) {
	tr := &Jules{}
	cfg := &config.Config{Inject: false}
	docs := []*parser.Document{
		{Type: parser.TypeSkills, Name: "review", Description: "Review code", Body: "skill body"},
		{Type: parser.TypeCommands, Name: "deploy", Description: "Deploy", Body: "deploy body"},
		{Type: parser.TypeReferences, Name: "api", Description: "API ref", Body: "api body"},
		{Type: parser.TypeSubagents, Name: "tester", Description: "Tester", Body: "tester body"},
	}
	plan, err := tr.Plan(docs, cfg)
	if err != nil {
		t.Fatal(err)
	}
	paths := pathSet(plan)
	wants := []string{
		".jules/rules/skills/review/SKILL.md",
		".jules/rules/commands/deploy.md",
		".jules/rules/references/api.md",
		".jules/rules/subagents/tester.md",
	}
	for _, w := range wants {
		if !paths[w] {
			t.Errorf("missing %s in plan: %v", w, paths)
		}
	}
}
