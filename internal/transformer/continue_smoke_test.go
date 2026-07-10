package transformer

import (
	"strings"
	"testing"

	"github.com/StringKe/std-agent/internal/config"
	"github.com/StringKe/std-agent/internal/parser"
)

func TestContinueDevOutputs(t *testing.T) {
	tr := &ContinueDev{}
	cfg := &config.Config{Inject: false}
	docs := []*parser.Document{
		{Type: parser.TypeRules, Name: "ts-style", ApplyTo: []string{"**/*.ts"}, Body: "use interfaces"},
		{Type: parser.TypeSkills, Name: "review", Description: "Code review", Body: "steps"},
		{Type: parser.TypeCommands, Name: "explain", Description: "Explain code", Body: "Please explain.", Version: "1.0"},
	}
	plan, _ := tr.Plan(docs, cfg)
	paths := pathSet(plan)
	// skills 原生 .continue/skills/<name>/SKILL.md（continuedev/continue#9353 GA）
	for _, want := range []string{
		".continue/rules/ts-style.md",
		".continue/skills/review/SKILL.md",
		".continue/prompts/explain.prompt.md",
	} {
		if !paths[want] {
			t.Errorf("missing %s, paths: %v", want, paths)
		}
	}
	if paths[".continue/rules/skills/review/SKILL.md"] {
		t.Errorf("stale degraded skill path still produced, paths: %v", paths)
	}
	rule, _ := contentOf(plan, ".continue/rules/ts-style.md")
	if !strings.Contains(rule, "globs:") || !strings.Contains(rule, "**/*.ts") {
		t.Errorf("ts-style missing globs: %s", rule)
	}
	prompt, _ := contentOf(plan, ".continue/prompts/explain.prompt.md")
	if !strings.Contains(prompt, "invokable: true") {
		t.Errorf("prompt missing invokable: %s", prompt)
	}
}

// TestContinueDevNestedRulesMd 验证嵌套目录说明写 <NestedPath>/rules.md
// （continue 只认固定文件名 rules.md 的 colocated rule，不读嵌套 AGENTS.md）
func TestContinueDevNestedRulesMd(t *testing.T) {
	tr := &ContinueDev{}
	cfg := &config.Config{Inject: false}
	docs := []*parser.Document{
		{Type: parser.TypeRules, Name: "web-root", NestedPath: "apps/web", Body: "web guidance"},
	}
	plan, _ := tr.Plan(docs, cfg)
	c, ok := contentOf(plan, "apps/web/rules.md")
	if !ok {
		t.Fatalf("missing apps/web/rules.md, paths: %v", pathSet(plan))
	}
	if !strings.Contains(c, "web guidance") {
		t.Errorf("nested body missing:\n%s", c)
	}
	if pathSet(plan)[".continue/rules/web-root.md"] {
		t.Errorf("nested doc should not also land in .continue/rules, paths: %v", pathSet(plan))
	}
}
