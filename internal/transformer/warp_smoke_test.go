package transformer

import (
	"strings"
	"testing"

	"github.com/StringKe/std-agent/internal/config"
	"github.com/StringKe/std-agent/internal/parser"
)

// TestWarpOutputs 验证根 AGENTS.md：inline nonRoot rules + glossary + manifest 头
func TestWarpOutputs(t *testing.T) {
	tr := &Warp{}
	cfg := &config.Config{Inject: false, InjectWhatIs: false}
	docs := []*parser.Document{
		{Type: parser.TypeRules, Name: "style", Description: "代码风格", Body: "Use clear names."},
		{Type: parser.TypeRules, Name: "naming", Description: "命名规范", Body: "Use snake_case."},
	}
	plan, err := tr.Plan(docs, cfg)
	if err != nil {
		t.Fatalf("plan err: %v", err)
	}
	main, ok := contentOf(plan, "AGENTS.md")
	if !ok {
		t.Fatalf("missing AGENTS.md, paths=%v", pathSet(plan))
	}
	// glossary 头
	if !strings.Contains(main, "std-agent type glossary") {
		t.Errorf("missing glossary header:\n%s", main)
	}
	// nonRoot rules 应 inline 进根文件
	for _, want := range []string{"Use clear names.", "Use snake_case."} {
		if !strings.Contains(main, want) {
			t.Errorf("missing inline rule body %q in:\n%s", want, main)
		}
	}
	// RulesDir 为空 + rule 未超阈值时不应在 .warp/rules/ 产生 rule 文件
	paths := pathSet(plan)
	if paths[".warp/rules/style.md"] || paths[".warp/rules/naming.md"] {
		t.Errorf("inline rules should not spill to .warp/rules/, paths=%v", paths)
	}
}

// TestWarpNestedAGENTSMd 验证嵌套子目录 AGENTS.md 自动叠加
func TestWarpNestedAGENTSMd(t *testing.T) {
	tr := &Warp{}
	cfg := &config.Config{Inject: false}
	docs := []*parser.Document{
		{Type: parser.TypeRules, Name: "auth", Root: true, NestedPath: "src/auth", Body: "# Auth\nbody"},
		{Type: parser.TypeRules, Name: "naming", Description: "命名", Body: "naming body"},
	}
	plan, err := tr.Plan(docs, cfg)
	if err != nil {
		t.Fatalf("plan err: %v", err)
	}
	paths := pathSet(plan)
	if !paths["src/auth/AGENTS.md"] {
		t.Errorf("expected src/auth/AGENTS.md, paths=%v", paths)
	}
	if !paths["AGENTS.md"] {
		t.Errorf("top-level AGENTS.md should still exist, paths=%v", paths)
	}
	nested, _ := contentOf(plan, "src/auth/AGENTS.md")
	// 嵌套根不带 manifest / glossary
	if strings.Contains(nested, "std-agent type glossary") {
		t.Errorf("nested AGENTS.md should NOT contain glossary:\n%s", nested)
	}
	if strings.Contains(nested, "## Reference Rules") {
		t.Errorf("nested AGENTS.md should NOT contain manifest:\n%s", nested)
	}
}

// TestWarpFallbacks 验证 commands / references / subagents 全 fallback 到 .warp/rules/
func TestWarpFallbacks(t *testing.T) {
	tr := &Warp{}
	cfg := &config.Config{Inject: false}
	docs := []*parser.Document{
		{Type: parser.TypeCommands, Name: "deploy", Description: "Deploy", Body: "go"},
		{Type: parser.TypeReferences, Name: "design", Description: "Design notes", Body: "background"},
		{Type: parser.TypeSubagents, Name: "reviewer", Description: "Code reviewer", Body: "instructions"},
	}
	plan, err := tr.Plan(docs, cfg)
	if err != nil {
		t.Fatalf("plan err: %v", err)
	}
	paths := pathSet(plan)
	for _, want := range []string{
		".warp/rules/commands/deploy.md",
		".warp/rules/references/design.md",
		".warp/rules/subagents/reviewer.md",
	} {
		if !paths[want] {
			t.Errorf("missing fallback %s, paths=%v", want, paths)
		}
	}
	// 每个 fallback 文件应含 std-agent-type 字段 + explainer 头
	for _, p := range []string{
		".warp/rules/commands/deploy.md",
		".warp/rules/references/design.md",
		".warp/rules/subagents/reviewer.md",
	} {
		c, _ := contentOf(plan, p)
		if !strings.Contains(c, "std-agent-type:") {
			t.Errorf("%s missing std-agent-type frontmatter:\n%s", p, c)
		}
		if !strings.Contains(c, "<!-- std-agent degraded") {
			t.Errorf("%s missing explainer header:\n%s", p, c)
		}
	}
}

// TestWarpNativeSkill 验证 skill 走 Warp 原生推荐路径 .agents/skills/<name>/SKILL.md
// （https://docs.warp.dev/agent-platform/capabilities/skills/，与 codex 共享落点，
// 内容字节一致由 writer unchanged 去重）
func TestWarpNativeSkill(t *testing.T) {
	tr := &Warp{}
	cfg := &config.Config{Inject: false}
	docs := []*parser.Document{
		{Type: parser.TypeSkills, Name: "review", Description: "Code review skill", Body: "Steps..."},
	}
	plan, err := tr.Plan(docs, cfg)
	if err != nil {
		t.Fatalf("plan err: %v", err)
	}
	c, ok := contentOf(plan, ".agents/skills/review/SKILL.md")
	if !ok {
		t.Fatalf("missing native skill output, paths=%v", pathSet(plan))
	}
	if !strings.Contains(c, "Code review skill") {
		t.Errorf("skill description should be preserved:\n%s", c)
	}
	if paths := pathSet(plan); paths[".warp/rules/skills/review/SKILL.md"] {
		t.Errorf("stale degraded skill path still produced: %v", paths)
	}
}
