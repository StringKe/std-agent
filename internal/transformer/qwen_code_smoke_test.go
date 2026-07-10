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
	// nonRoot rules 落原生 .qwen/rules/（源码 loadRules，支持 paths 条件规则），
	// 不再全量 inline 进 QWEN.md
	rule, ok := contentOf(plan, ".qwen/rules/naming.md")
	if !ok {
		t.Fatalf("missing native rule .qwen/rules/naming.md, paths: %v", paths)
	}
	if !strings.Contains(rule, "Use clear names.") {
		t.Errorf("rule body missing:\n%s", rule)
	}
	// root QWEN.md 引用 manifest 而不是 inline 全文
	root, _ := contentOf(plan, "QWEN.md")
	if !strings.Contains(root, ".qwen/rules/naming.md") {
		t.Errorf("root QWEN.md should reference rule in manifest, got:\n%s", root)
	}
}

func TestQwenCodeRulePathsFrontmatter(t *testing.T) {
	tr := &QwenCode{}
	cfg := &config.Config{Inject: false}
	docs := []*parser.Document{
		{Type: parser.TypeRules, Name: "go-style", Description: "Go style", ApplyTo: []string{"**/*.go"}, Body: "gofmt."},
	}
	plan, _ := tr.Plan(docs, cfg)
	c, ok := contentOf(plan, ".qwen/rules/go-style.md")
	if !ok {
		t.Fatalf("missing .qwen/rules/go-style.md, paths: %v", pathSet(plan))
	}
	if !strings.Contains(c, "paths:") || !strings.Contains(c, "**/*.go") {
		t.Errorf("qwen rule should render paths frontmatter (lazy conditional rule):\n%s", c)
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
	// skills 原生 .qwen/skills/<name>/SKILL.md（官方 GA）
	if !paths[".qwen/skills/review/SKILL.md"] {
		t.Errorf("missing native skill at .qwen/skills/review/SKILL.md, paths: %v", paths)
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
