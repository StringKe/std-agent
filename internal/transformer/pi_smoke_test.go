package transformer

import (
	"strings"
	"testing"

	"std-ai/internal/config"
	"std-ai/internal/parser"
)

// TestPiOutputs 验证 pi adapter 主路径：
//   - 根 AGENTS.md
//   - .pi/skills/<name>/SKILL.md（Agent Skills 原生包）
//   - .pi/prompts/<name>.md（命令降级前的原生 commands 落点）
func TestPiOutputs(t *testing.T) {
	tr := &Pi{}
	cfg := &config.Config{Inject: true, InjectWhatIs: false}
	docs := []*parser.Document{
		{Type: parser.TypeRules, Name: "coding-style", Description: "Style", Body: "Use clear names."},
		{Type: parser.TypeSkills, Name: "review", Description: "Review code.", Body: "Steps..."},
		{Type: parser.TypeCommands, Name: "deploy", Description: "Deploy.", Body: "Run deploy."},
	}
	plan, err := tr.Plan(docs, cfg)
	if err != nil {
		t.Fatalf("Plan err: %v", err)
	}
	paths := pathSet(plan)
	for _, want := range []string{
		"AGENTS.md",
		".pi/skills/review/SKILL.md",
		".pi/prompts/deploy.md",
	} {
		if !paths[want] {
			t.Errorf("missing %q in plan, got: %v", want, paths)
		}
	}

	// 根文件应含 glossary（InjectTypeGlossary=true）与 inline nonRoot rule body
	// （RulesDir 为空 -> nonRoot rule 直接 inline）
	main, _ := contentOf(plan, "AGENTS.md")
	if !strings.Contains(main, "Use clear names.") {
		t.Errorf("expected nonRoot rule body inlined in AGENTS.md, got:\n%s", main)
	}

	// SKILL.md 应含白名单 frontmatter
	skill, _ := contentOf(plan, ".pi/skills/review/SKILL.md")
	for _, want := range []string{"name: review", "description: Review code."} {
		if !strings.Contains(skill, want) {
			t.Errorf("missing %q in SKILL.md:\n%s", want, skill)
		}
	}
}

// TestPiReferencesAndSubagentsFallback 验证 fallback 行为：
//   - subagents 无 SubagentsDir -> 落到 .pi/rules/subagents/<name>.md
//   - references 无 ReferencesDir -> v3 走独立 references/ 子目录
//     .pi/rules/references/<name>.md（不再借 SkillsDir SKILL.md 路径，避免被 pi
//     按 skill 触发加载破坏"按需查阅"语义）
func TestPiReferencesAndSubagentsFallback(t *testing.T) {
	tr := &Pi{}
	cfg := &config.Config{Inject: false}
	docs := []*parser.Document{
		{Type: parser.TypeReferences, Name: "api-spec", Description: "API spec", Body: "spec body"},
		{Type: parser.TypeSubagents, Name: "reviewer", Description: "Reviewer", Body: "do review"},
	}
	plan, _ := tr.Plan(docs, cfg)
	paths := pathSet(plan)
	if !paths[".pi/rules/references/api-spec.md"] {
		t.Errorf("expected reference fallback to .pi/rules/references/api-spec.md, got: %v", paths)
	}
	if !paths[".pi/rules/subagents/reviewer.md"] {
		t.Errorf("expected subagent fallback to .pi/rules/subagents/reviewer.md, got: %v", paths)
	}
}
