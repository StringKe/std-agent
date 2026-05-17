package transformer

import (
	"strings"
	"testing"

	"std-ai/internal/config"
	"std-ai/internal/parser"
)

// TestRooCodeOutputs 验证 roo-code adapter 的 4 类 type 落点：
//   - rules    -> .roo/rules/<name>.md（无数字前缀；ApplyTo 走 paths frontmatter list）
//   - commands -> .roo/rules/workflows/<name>.md
//   - skills   -> .roo/rules/skills/<name>/SKILL.md（Agent Skills 标准 fallback）
//   - references / subagents -> .roo/rules/<sub>/<name>.md
//
// 关键断言：路径无 std-ai 私有前缀；rule 文件名不含 100/500/900 等 cline 数字前缀。
func TestRooCodeOutputs(t *testing.T) {
	tr := &RooCode{}
	cfg := &config.Config{Inject: true, InjectWhatIs: false, InjectTypeGlossary: true}
	docs := []*parser.Document{
		{Type: parser.TypeRules, Name: "style", Description: "Style", Body: "body-style", ApplyTo: []string{"**/*.go"}},
		{Type: parser.TypeRules, Name: "always", AlwaysApply: true, Body: "body-always"},
		{Type: parser.TypeCommands, Name: "release", Description: "Cut release", Body: "body-release"},
		{Type: parser.TypeSkills, Name: "code-review", Description: "review", Body: "body-review"},
		{Type: parser.TypeReferences, Name: "spec", Description: "spec", Body: "body-spec"},
		{Type: parser.TypeSubagents, Name: "qa", Description: "qa agent", Body: "body-qa"},
	}

	plan, err := tr.Plan(docs, cfg)
	if err != nil {
		t.Fatalf("plan error: %v", err)
	}

	paths := pathSet(plan)

	// rules：.roo/rules/<name>.md（无数字前缀）
	wantPaths := []string{
		".roo/rules/style.md",
		".roo/rules/always.md",
		".roo/rules/workflows/release.md",
		".roo/rules/skills/code-review/SKILL.md",
		".roo/rules/references/spec.md",
		".roo/rules/subagents/qa.md",
		".roo/rules/glossary.md",
	}
	for _, p := range wantPaths {
		if !paths[p] {
			t.Errorf("missing %s, paths: %v", p, paths)
		}
	}

	// rule 文件名不应有 cline 风格的 100- / 500- / 900- 前缀
	for p := range paths {
		if strings.HasPrefix(p, ".roo/rules/100-") ||
			strings.HasPrefix(p, ".roo/rules/500-") ||
			strings.HasPrefix(p, ".roo/rules/900-") {
			t.Errorf("roo-code should not have cline-style numeric prefix: %s", p)
		}
	}

	// rule 文件含 paths frontmatter（ApplyTo 渲染为 YAML list）
	styleBody, ok := contentOf(plan, ".roo/rules/style.md")
	if !ok {
		t.Fatal("missing .roo/rules/style.md")
	}
	if !strings.Contains(styleBody, "paths:") {
		t.Errorf("style.md missing paths frontmatter:\n%s", styleBody)
	}
	if !strings.Contains(styleBody, "**/*.go") {
		t.Errorf("style.md missing applyTo value:\n%s", styleBody)
	}

	// 路径不应包含 std-ai 私有前缀（如 "std-"）
	for p := range paths {
		base := p
		if idx := strings.LastIndex(p, "/"); idx >= 0 {
			base = p[idx+1:]
		}
		if strings.HasPrefix(base, "std-") && base != "std-ai-type" {
			t.Errorf("unexpected std-ai private prefix in filename: %s", p)
		}
	}
}

// TestRooCodeFallbackSubdirs 单独验证非原生 type 都落到独立子目录，
// 不会和 rules 文件名冲突。
func TestRooCodeFallbackSubdirs(t *testing.T) {
	tr := &RooCode{}
	cfg := &config.Config{Inject: false}
	docs := []*parser.Document{
		{Type: parser.TypeSkills, Name: "foo", Body: "x"},
		{Type: parser.TypeReferences, Name: "foo", Body: "x"},
		{Type: parser.TypeSubagents, Name: "foo", Body: "x"},
	}
	plan, _ := tr.Plan(docs, cfg)
	paths := pathSet(plan)
	for _, want := range []string{
		".roo/rules/skills/foo/SKILL.md",
		".roo/rules/references/foo.md",
		".roo/rules/subagents/foo.md",
	} {
		if !paths[want] {
			t.Errorf("missing %s, paths: %v", want, paths)
		}
	}
}
