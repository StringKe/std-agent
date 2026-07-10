package transformer

import (
	"strings"
	"testing"

	"github.com/StringKe/std-agent/internal/config"
	"github.com/StringKe/std-agent/internal/parser"
)

func TestQwenCodeOutputs(t *testing.T) {
	tr := &QwenCode{}
	cfg := &config.Config{Inject: false}
	docs := []*parser.Document{
		{Type: parser.TypeRules, Name: "naming", Description: "命名", Body: "Use clear names."},
		{Type: parser.TypeCommands, Name: "deploy", Description: "Deploy", Body: "go"},
	}
	plan, err := tr.Plan(docs, cfg)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	paths := pathSet(plan)
	if !paths["QWEN.md"] {
		t.Errorf("missing root QWEN.md, paths: %v", paths)
	}
	if !paths[".qwen/commands/deploy.md"] {
		t.Errorf("missing .qwen/commands/deploy.md, paths: %v", paths)
	}
	// root QWEN.md 应 inline 含 nonRoot rule body（RulesDir=""）
	root, _ := contentOf(plan, "QWEN.md")
	if !strings.Contains(root, "Use clear names.") {
		t.Errorf("root QWEN.md should inline rule body, got:\n%s", root)
	}
}

func TestQwenCodeFallback(t *testing.T) {
	tr := &QwenCode{}
	cfg := &config.Config{Inject: false}
	docs := []*parser.Document{
		{Type: parser.TypeSkills, Name: "review", Description: "Review code.", Body: "Steps..."},
		{Type: parser.TypeReferences, Name: "api", Description: "API ref", Body: "details"},
		{Type: parser.TypeSubagents, Name: "linter", Description: "Lint", Body: "lint body"},
	}
	plan, _ := tr.Plan(docs, cfg)
	paths := pathSet(plan)
	// skills fallback to BuildDegradedSkillPackage -> .qwen/rules/skills/<name>/SKILL.md
	if !paths[".qwen/rules/skills/review/SKILL.md"] {
		t.Errorf("missing skill fallback at .qwen/rules/skills/review/SKILL.md, paths: %v", paths)
	}
	// references fallback to .qwen/rules/references/<name>.md
	if !paths[".qwen/rules/references/api.md"] {
		t.Errorf("missing references fallback at .qwen/rules/references/api.md, paths: %v", paths)
	}
	// subagents fallback to .qwen/rules/subagents/<name>.md
	if !paths[".qwen/rules/subagents/linter.md"] {
		t.Errorf("missing subagents fallback at .qwen/rules/subagents/linter.md, paths: %v", paths)
	}
}
