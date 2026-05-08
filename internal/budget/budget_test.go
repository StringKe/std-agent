package budget

import (
	"strings"
	"testing"

	"std-ai/internal/parser"
)

func TestCheckDocumentRuleSoftWarn(t *testing.T) {
	d := &parser.Document{
		Path: "rules/big.md",
		Type: parser.TypeRules,
		Body: strings.Repeat("a", 9000),
	}
	out := CheckDocument(d)
	if len(out) == 0 {
		t.Fatal("expected SOFT warn for >8000 char rule")
	}
	if !strings.Contains(out[0], "SOFT") {
		t.Errorf("expected SOFT prefix, got %q", out[0])
	}
}

func TestCheckDocumentRuleHardWarn(t *testing.T) {
	d := &parser.Document{
		Path: "rules/giant.md",
		Type: parser.TypeRules,
		Body: strings.Repeat("a", 13000), // > windsurf hard 12000
	}
	out := CheckDocument(d)
	hard := false
	for _, msg := range out {
		if strings.Contains(msg, "HARD") && strings.Contains(msg, "windsurf") {
			hard = true
		}
	}
	if !hard {
		t.Errorf("expected HARD windsurf warn for > 12000, got %v", out)
	}
}

func TestCheckDocumentSkillBudget(t *testing.T) {
	d := &parser.Document{
		Path: "skills/foo/SKILL.md",
		Type: parser.TypeSkills,
		Body: strings.Repeat("s", 25000),
	}
	out := CheckDocument(d)
	if len(out) == 0 {
		t.Error("expected skill budget warn")
	}
	if !strings.Contains(out[0], "skill") {
		t.Errorf("expected skill kind, got %q", out[0])
	}
}

func TestCheckDocumentCommandBudget(t *testing.T) {
	d := &parser.Document{
		Path: "commands/x.md",
		Type: parser.TypeCommands,
		Body: strings.Repeat("c", 5000),
	}
	out := CheckDocument(d)
	if len(out) == 0 {
		t.Error("expected command budget warn for >4000 chars")
	}
}

func TestCheckDocumentSmallBodyNoWarn(t *testing.T) {
	d := &parser.Document{
		Path: "rules/small.md",
		Type: parser.TypeRules,
		Body: "tiny",
	}
	if out := CheckDocument(d); len(out) != 0 {
		t.Errorf("small body should not warn, got %v", out)
	}
}

func TestCheckDocumentReferencesNotChecked(t *testing.T) {
	d := &parser.Document{
		Path: "references/big.md",
		Type: parser.TypeReferences,
		Body: strings.Repeat("r", 100000),
	}
	if out := CheckDocument(d); len(out) != 0 {
		t.Errorf("references should not be budget-checked, got %v", out)
	}
}

func TestCheckDocumentNil(t *testing.T) {
	if out := CheckDocument(nil); out != nil {
		t.Errorf("nil should return nil, got %v", out)
	}
}

func TestCheckDocumentBodyBytesOverridesString(t *testing.T) {
	d := &parser.Document{
		Path:      "rules/x.md",
		Type:      parser.TypeRules,
		Body:      "tiny",
		BodyBytes: 9000, // 显式覆盖
	}
	out := CheckDocument(d)
	if len(out) == 0 {
		t.Error("BodyBytes should take precedence over Body length")
	}
}

func TestCheckTotalRulesOverCodexLimit(t *testing.T) {
	docs := []*parser.Document{
		{Type: parser.TypeRules, Body: strings.Repeat("a", 20000)},
		{Type: parser.TypeRules, Body: strings.Repeat("b", 20000)},
		{Type: parser.TypeSkills, Body: strings.Repeat("c", 20000)}, // 不计入
	}
	out := CheckTotalRules(docs)
	if len(out) == 0 {
		t.Error("expected codex AGENTS.md total warn")
	}
	if !strings.Contains(out[0], "codex") {
		t.Errorf("expected codex target, got %q", out[0])
	}
}

func TestCheckTotalRulesUnderLimit(t *testing.T) {
	docs := []*parser.Document{
		{Type: parser.TypeRules, Body: "small"},
	}
	if out := CheckTotalRules(docs); len(out) != 0 {
		t.Errorf("under limit should not warn, got %v", out)
	}
}
