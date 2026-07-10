package transformer

import (
	"strings"
	"testing"

	"github.com/StringKe/std-agent/internal/config"
	"github.com/StringKe/std-agent/internal/parser"
)

func TestRegistry(t *testing.T) {
	if _, ok := Get("claude-code"); !ok {
		t.Error("claude-code transformer not registered")
	}
	if _, ok := Get("codex"); !ok {
		t.Error("codex transformer not registered")
	}
	if _, ok := Get("nonexistent"); ok {
		t.Error("nonexistent should not be registered")
	}
	if len(Names()) < 2 {
		t.Errorf("Names() = %v, want at least 2", Names())
	}
}

func TestWindsurfRuleOverLimitWarn(t *testing.T) {
	tr := &Windsurf{}
	cfg := &config.Config{Inject: false}
	bigBody := strings.Repeat("a", 13000)
	docs := []*parser.Document{
		{Type: parser.TypeRules, Name: "huge", AlwaysApply: true, Body: bigBody},
	}
	plan, _ := tr.Plan(docs, cfg)
	found := false
	for _, f := range plan.Files {
		if strings.HasPrefix(f.Reason, "WARN") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected WARN for over-limit windsurf rule")
	}
}

func TestAntigravityRuleOverLimitWarn(t *testing.T) {
	tr := &Antigravity{}
	cfg := &config.Config{Inject: false}
	bigBody := strings.Repeat("b", 13000)
	docs := []*parser.Document{
		{Type: parser.TypeRules, Name: "huge", AlwaysApply: true, Body: bigBody},
	}
	plan, _ := tr.Plan(docs, cfg)
	found := false
	for _, f := range plan.Files {
		if strings.HasPrefix(f.Reason, "WARN") {
			found = true
		}
	}
	if !found {
		t.Error("expected WARN for over-limit antigravity rule")
	}
}

// TestCodexAGENTSMdInlinesLargeRules：超大 rule 也全文 inline 到 AGENTS.md
// （.codex/memories spill 已废弃；总体积超 32768 由 budget root-file 检查提醒）
func TestCodexAGENTSMdInlinesLargeRules(t *testing.T) {
	tr := &Codex{}
	cfg := &config.Config{Inject: false}
	bigBody := strings.Repeat("c ", 18000)
	docs := []*parser.Document{
		{Type: parser.TypeRules, Name: "aaa-tiny", Body: "tiny"},
		{Type: parser.TypeRules, Name: "bbb-huge", Body: bigBody},
		{Type: parser.TypeRules, Name: "ccc-huger", Body: bigBody},
	}
	plan, _ := tr.Plan(docs, cfg)
	paths := pathSet(plan)
	if len(paths) != 1 || !paths["AGENTS.md"] {
		t.Errorf("expected AGENTS.md as the only output, paths: %v", paths)
	}
	main, _ := contentOf(plan, "AGENTS.md")
	for _, want := range []string{"tiny", strings.TrimSpace(bigBody)} {
		if !strings.Contains(main, want) {
			t.Error("rule body missing from inlined AGENTS.md")
		}
	}
}

func TestCursorRuleVeryLongBodyDoesNotPanic(t *testing.T) {
	tr := &Cursor{}
	cfg := &config.Config{Inject: false}
	docs := []*parser.Document{
		{
			Type: parser.TypeRules, Name: "long", AlwaysApply: true,
			Body: strings.Repeat("x ", 50000),
		},
	}
	plan, err := tr.Plan(docs, cfg)
	if err != nil {
		t.Fatal(err)
	}
	// 至少 1 个 rule 文件（cursor 默认 InjectTypeGlossary=true 会额外产 glossary.md）
	hasRule := false
	for _, f := range plan.Files {
		if strings.HasSuffix(f.Path, "/long.mdc") {
			hasRule = true
			break
		}
	}
	if !hasRule {
		t.Errorf("expected long.mdc rule file, got %d files", len(plan.Files))
	}
}
