package transformer

import (
	"strings"
	"testing"

	"github.com/StringKe/std-agent/internal/config"
	"github.com/StringKe/std-agent/internal/parser"
)

func TestAmpOutputs(t *testing.T) {
	tr := &Amp{}
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
	// 不应存在 .amp/rules 下的独立 rule 文件（RulesDir=""，nonRoot 全 inline）
	for _, f := range plan.Files {
		if strings.HasPrefix(f.Path, ".amp/rules/") && !strings.Contains(f.Path, "/skills/") &&
			!strings.Contains(f.Path, "/commands/") && !strings.Contains(f.Path, "/references/") &&
			!strings.Contains(f.Path, "/subagents/") {
			t.Errorf("unexpected rule fan-out: %s", f.Path)
		}
	}
}

func TestAmpNativeSkillsAndFallback(t *testing.T) {
	tr := &Amp{}
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
	// skills 原生 .agents/skills/（与 codex 共享落点）；commands 官方已移除并入
	// skills，同 codex 降级为 .agents/skills/commands/<n>/SKILL.md
	wants := []string{
		".agents/skills/review/SKILL.md",
		".agents/skills/commands/deploy/SKILL.md",
		".amp/rules/references/api.md",
		".amp/rules/subagents/tester.md",
	}
	for _, w := range wants {
		if !paths[w] {
			t.Errorf("missing %s in plan: %v", w, paths)
		}
	}
	// 旧降级路径不应再产出
	for p := range paths {
		if strings.HasPrefix(p, ".amp/rules/skills/") || strings.HasPrefix(p, ".amp/rules/commands/") {
			t.Errorf("stale degraded path still produced: %s", p)
		}
	}
	// skill 原生包不含 degraded 标记
	if c, ok := contentOf(plan, ".agents/skills/review/SKILL.md"); ok {
		if strings.Contains(c, "std-agent degraded") || strings.Contains(c, "std-agent-type") {
			t.Errorf("native skill should not carry degraded markers:\n%s", c)
		}
	}
}

// TestAmpCodexSkillByteIdentical 保证 amp 与 codex 对同一 skill / command 产出
// 字节一致——两个 target 共享 .agents/skills/ 落点，靠 writer unchanged 去重，
// 内容不一致会导致每次 sync 互相改写（flip-flop）。
func TestAmpCodexSkillByteIdentical(t *testing.T) {
	cfg := &config.Config{Inject: true, InjectWhatIs: false}
	docs := []*parser.Document{
		{Type: parser.TypeSkills, Name: "review", Description: "Review code", WhenToUse: "on review", Body: "skill body"},
		{Type: parser.TypeCommands, Name: "deploy", Description: "Deploy", Body: "deploy body"},
	}
	ampPlan, err := (&Amp{}).Plan(docs, cfg)
	if err != nil {
		t.Fatal(err)
	}
	codexPlan, err := (&Codex{}).Plan(docs, cfg)
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{".agents/skills/review/SKILL.md", ".agents/skills/commands/deploy/SKILL.md"} {
		a, aok := contentOf(ampPlan, p)
		c, cok := contentOf(codexPlan, p)
		if !aok || !cok {
			t.Fatalf("both targets must produce %s (amp=%v codex=%v)", p, aok, cok)
		}
		if a != c {
			t.Errorf("amp and codex output differ for shared path %s:\namp:\n%s\ncodex:\n%s", p, a, c)
		}
	}
}
