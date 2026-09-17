package transformer

import (
	"strings"
	"testing"

	"github.com/StringKe/std-agent/internal/config"
	"github.com/StringKe/std-agent/internal/parser"
)

// TestMuseOutputs 验证 muse adapter 主路径：
//   - 根 AGENTS.md（nonRoot rules inline，RulesDir 为空）
//   - .agents/skills/<name>/SKILL.md（官方项目 skills 路径）
//   - .agents/skills/commands/<name>/SKILL.md（commands 无原生目录，降级为 skill）
//   - .agents/memory/<name>.md（官方项目记忆目录）
func TestMuseOutputs(t *testing.T) {
	tr := &Muse{}
	cfg := &config.Config{Inject: true, InjectWhatIs: false}
	docs := []*parser.Document{
		{Type: parser.TypeRules, Name: "coding-style", Description: "Style", Body: "Use clear names."},
		{Type: parser.TypeSkills, Name: "review", Description: "Review code.", Body: "Steps..."},
		{Type: parser.TypeCommands, Name: "deploy", Description: "Deploy.", Body: "Run deploy."},
		{Type: parser.TypeReferences, Name: "api-spec", Description: "API spec", Body: "spec body"},
	}
	plan, err := tr.Plan(docs, cfg)
	if err != nil {
		t.Fatalf("Plan err: %v", err)
	}
	paths := pathSet(plan)
	for _, want := range []string{
		"AGENTS.md",
		".agents/skills/review/SKILL.md",
		".agents/skills/commands/deploy/SKILL.md",
		".agents/memory/api-spec.md",
	} {
		if !paths[want] {
			t.Errorf("missing %q in plan, got: %v", want, paths)
		}
	}

	// 根文件应含 inline nonRoot rule body（RulesDir 为空 -> 全文 inline）
	main, _ := contentOf(plan, "AGENTS.md")
	if !strings.Contains(main, "Use clear names.") {
		t.Errorf("expected nonRoot rule body inlined in AGENTS.md, got:\n%s", main)
	}

	// SKILL.md 应含共享字段集 frontmatter（与 codex / goose / zed 字节一致）
	skill, _ := contentOf(plan, ".agents/skills/review/SKILL.md")
	for _, want := range []string{"name: review", "description: Review code."} {
		if !strings.Contains(skill, want) {
			t.Errorf("missing %q in SKILL.md:\n%s", want, skill)
		}
	}
}

// TestMuseSubagentsFallback 验证 subagents 无原生 markdown 格式，走
// FallbackDir .muse 子目录隔离：.muse/subagents/<name>.md
func TestMuseSubagentsFallback(t *testing.T) {
	tr := &Muse{}
	cfg := &config.Config{Inject: false}
	docs := []*parser.Document{
		{Type: parser.TypeSubagents, Name: "reviewer", Description: "Reviewer", Body: "do review"},
	}
	plan, err := tr.Plan(docs, cfg)
	if err != nil {
		t.Fatalf("Plan err: %v", err)
	}
	paths := pathSet(plan)
	if !paths[".muse/subagents/reviewer.md"] {
		t.Errorf("expected subagent fallback to .muse/subagents/reviewer.md, got: %v", paths)
	}
}
