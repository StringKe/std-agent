package transformer

import (
	"strings"
	"testing"

	"std-ai/internal/config"
	"std-ai/internal/parser"
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
	// 迁移到 WindsurfStyle 协议后，skills 不再降级为 skill-<name>.md，
	// 改走 Agent Skills 标准 fallback：.continue/rules/skills/<name>/SKILL.md
	for _, want := range []string{
		".continue/rules/ts-style.md",
		".continue/rules/skills/review/SKILL.md",
		".continue/prompts/explain.prompt.md",
	} {
		if !paths[want] {
			t.Errorf("missing %s, paths: %v", want, paths)
		}
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
