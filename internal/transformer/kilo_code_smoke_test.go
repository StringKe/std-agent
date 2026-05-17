package transformer

import (
	"strings"
	"testing"

	"std-ai/internal/config"
	"std-ai/internal/parser"
)

// TestKiloCodeOutputs 验证 kilo-code adapter 的 4 类 type 落点：
//   - rules    -> .kilo/rules/<name>.md（无数字前缀）
//   - commands -> .kilo/rules/workflows/<name>.md
//   - skills   -> .kilo/rules/skills/<n>/SKILL.md（Agent Skills 标准 fallback）
//   - references / subagents -> .kilo/rules/<sub>/<name>.md
//
// 关键断言：路径无 std-ai 私有前缀；rule 文件名不含 100/500/900 等 cline 数字前缀。
func TestKiloCodeOutputs(t *testing.T) {
	tr := &KiloCode{}
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

	wantPaths := []string{
		".kilo/rules/style.md",
		".kilo/rules/always.md",
		".kilo/rules/workflows/release.md",
		".kilo/rules/skills/code-review/SKILL.md",
		".kilo/rules/references/spec.md",
		".kilo/rules/subagents/qa.md",
		".kilo/rules/glossary.md",
	}
	for _, p := range wantPaths {
		if !paths[p] {
			t.Errorf("missing %s, paths: %v", p, paths)
		}
	}

	// rule 文件名不应有 cline 风格的 100- / 500- / 900- 前缀
	for p := range paths {
		if strings.HasPrefix(p, ".kilo/rules/100-") ||
			strings.HasPrefix(p, ".kilo/rules/500-") ||
			strings.HasPrefix(p, ".kilo/rules/900-") {
			t.Errorf("kilo-code should not have cline-style numeric prefix: %s", p)
		}
	}

	// 路径不应包含 std-ai 私有前缀
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

// TestKiloCodeFallbackSubdirs 单独验证非原生 type 都落到独立子目录，
// 不会和 rules 文件名冲突。
func TestKiloCodeFallbackSubdirs(t *testing.T) {
	tr := &KiloCode{}
	cfg := &config.Config{Inject: false}
	docs := []*parser.Document{
		{Type: parser.TypeSkills, Name: "foo", Body: "x"},
		{Type: parser.TypeReferences, Name: "foo", Body: "x"},
		{Type: parser.TypeSubagents, Name: "foo", Body: "x"},
	}
	plan, _ := tr.Plan(docs, cfg)
	paths := pathSet(plan)
	for _, want := range []string{
		".kilo/rules/skills/foo/SKILL.md",
		".kilo/rules/references/foo.md",
		".kilo/rules/subagents/foo.md",
	} {
		if !paths[want] {
			t.Errorf("missing %s, paths: %v", want, paths)
		}
	}
}
