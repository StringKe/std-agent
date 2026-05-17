package transformer

import (
	"strings"
	"testing"

	"std-ai/internal/config"
	"std-ai/internal/parser"
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

func TestCodexAGENTSMdMultipleSpills(t *testing.T) {
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
	if !paths["AGENTS.md"] {
		t.Error("AGENTS.md missing")
	}
	spillCount := 0
	for p := range paths {
		if strings.HasPrefix(p, ".codex/memories/") {
			spillCount++
		}
	}
	if spillCount < 2 {
		t.Errorf("expected at least 2 spills, got %d, paths: %v", spillCount, paths)
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
	if len(plan.Files) != 1 {
		t.Errorf("expected 1 file, got %d", len(plan.Files))
	}
}
