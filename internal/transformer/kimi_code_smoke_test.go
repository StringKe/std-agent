package transformer

import (
	"strings"
	"testing"

	"github.com/StringKe/std-agent/internal/config"
	"github.com/StringKe/std-agent/internal/parser"
)

func TestKimiCodeOutputs(t *testing.T) {
	tr := &KimiCode{}
	cfg := &config.Config{Inject: false}
	docs := []*parser.Document{
		{Type: parser.TypeRules, Name: "naming", Description: "命名", Body: "Use clear names."},
		{Type: parser.TypeSkills, Name: "review", Description: "Review code.", Body: "Steps..."},
		{Type: parser.TypeCommands, Name: "deploy", Description: "Deploy", Body: "go"},
	}
	plan, err := tr.Plan(docs, cfg)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	paths := pathSet(plan)
	if !paths["AGENTS.md"] {
		t.Errorf("missing root AGENTS.md, paths: %v", paths)
	}
	// 无 RulesDir：nonRoot rules 全 inline 到 AGENTS.md
	root, _ := contentOf(plan, "AGENTS.md")
	if !strings.Contains(root, "Use clear names.") {
		t.Errorf("nonRoot rule should be inlined into AGENTS.md:\n%s", root)
	}
	// skills 原生 .agents/skills/（kimi-code 官方 Project 层扫描路径之一）
	if !paths[".agents/skills/review/SKILL.md"] {
		t.Errorf("missing native skill at .agents/skills/review/SKILL.md, paths: %v", paths)
	}
	// commands 无独立机制，降级为 skill（kimi 自动注册 /skill:<name>）
	if !paths[".agents/skills/commands/deploy/SKILL.md"] {
		t.Errorf("missing command-as-skill at .agents/skills/commands/deploy/SKILL.md, paths: %v", paths)
	}
}

func TestKimiCodeFallback(t *testing.T) {
	tr := &KimiCode{}
	cfg := &config.Config{Inject: false}
	docs := []*parser.Document{
		{Type: parser.TypeReferences, Name: "api", Description: "API ref", Body: "details"},
		{Type: parser.TypeSubagents, Name: "linter", Description: "Lint", Body: "lint body"},
	}
	plan, _ := tr.Plan(docs, cfg)
	paths := pathSet(plan)
	// references / subagents 走私有 fallback（explainer 含 target 名，
	// 不能落共享 .agents/ 否则与 codex 的降级产物互相改写）
	if !paths[".kimi-code/rules/references/api.md"] {
		t.Errorf("missing references fallback at .kimi-code/rules/references/api.md, paths: %v", paths)
	}
	if !paths[".kimi-code/agents/linter.md"] {
		t.Errorf("missing native subagent at .kimi-code/agents/linter.md, paths: %v", paths)
	}
}

func TestKimiCodeNestedRoot(t *testing.T) {
	tr := &KimiCode{}
	cfg := &config.Config{Inject: false}
	docs := []*parser.Document{
		{Type: parser.TypeRules, Name: "auth", Root: true, NestedPath: "src/auth", Body: "# Auth\nbody"},
	}
	plan, _ := tr.Plan(docs, cfg)
	if !pathSet(plan)["src/auth/AGENTS.md"] {
		t.Errorf("kimi-code 层级发现应产出嵌套 AGENTS.md, paths: %v", pathSet(plan))
	}
}

// TestKimiCodeCodexSkillByteIdentical 保证 kimi-code 与 codex 对同一 skill /
// command 产出字节一致——五个 target（codex / amp / warp / antigravity /
// kimi-code）共享 .agents/skills/ 落点，靠 writer unchanged 去重，
// 内容不一致会导致每次 sync 互相改写（flip-flop）。
func TestKimiCodeCodexSkillByteIdentical(t *testing.T) {
	cfg := &config.Config{Inject: true, InjectWhatIs: false}
	docs := []*parser.Document{
		{Type: parser.TypeSkills, Name: "review", Description: "Review code", WhenToUse: "on review", Body: "skill body"},
		{Type: parser.TypeCommands, Name: "deploy", Description: "Deploy", Body: "deploy body"},
	}
	kimiPlan, err := (&KimiCode{}).Plan(docs, cfg)
	if err != nil {
		t.Fatal(err)
	}
	codexPlan, err := (&Codex{}).Plan(docs, cfg)
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{".agents/skills/review/SKILL.md", ".agents/skills/commands/deploy/SKILL.md"} {
		k, kok := contentOf(kimiPlan, p)
		c, cok := contentOf(codexPlan, p)
		if !kok || !cok {
			t.Fatalf("both targets must produce %s (kimi-code=%v codex=%v)", p, kok, cok)
		}
		if k != c {
			t.Errorf("kimi-code and codex output differ for shared path %s:\nkimi-code:\n%s\ncodex:\n%s", p, k, c)
		}
	}
}
