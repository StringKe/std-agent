package budget

import (
	"strings"
	"testing"

	"github.com/StringKe/std-agent/internal/parser"
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

func TestCheckRootFileSoftWarnClaudeCode(t *testing.T) {
	out := CheckRootFile("claude-code", "CLAUDE.md", 20000)
	if len(out) != 1 || !strings.Contains(out[0], "SOFT") {
		t.Errorf("expected SOFT for 20000 byte CLAUDE.md, got %v", out)
	}
}

func TestCheckRootFileHardWarnCodex(t *testing.T) {
	out := CheckRootFile("codex", "AGENTS.md", 40000)
	if len(out) != 1 || !strings.Contains(out[0], "HARD") {
		t.Errorf("expected HARD for 40000 byte AGENTS.md, got %v", out)
	}
}

func TestCheckRootFileCopilotSoftOnly(t *testing.T) {
	// 2026-06-12 官方移除截断规则后 copilot 根文件只剩通用软建议，无 HARD
	if out := CheckRootFile("copilot", ".github/copilot-instructions.md", 5000); len(out) != 0 {
		t.Errorf("5000 bytes is under the 8000 soft threshold, got %v", out)
	}
	soft := CheckRootFile("copilot", ".github/copilot-instructions.md", 10000)
	if len(soft) != 1 || !strings.Contains(soft[0], "SOFT") {
		t.Errorf("expected SOFT (not HARD) for 10000 byte copilot instructions, got %v", soft)
	}
}

func TestCheckRootFileNoWarnSmall(t *testing.T) {
	for _, target := range []string{"claude-code", "codex", "gemini", "copilot"} {
		if out := CheckRootFile(target, "x", 500); len(out) != 0 {
			t.Errorf("small file should not warn for %s, got %v", target, out)
		}
	}
}

func TestCheckRootFileUnknownTarget(t *testing.T) {
	if out := CheckRootFile("aider", "x", 99999); len(out) != 0 {
		t.Errorf("aider has no root-file limit, got %v", out)
	}
}

func TestCheckDocumentSkillNameHard(t *testing.T) {
	d := &parser.Document{
		Path: "skills/x/SKILL.md",
		Type: parser.TypeSkills,
		Name: strings.Repeat("a", 65),
		Body: "short",
	}
	out := CheckDocument(d)
	found := false
	for _, msg := range out {
		if strings.Contains(msg, "HARD") && strings.Contains(msg, "skill-name") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected HARD skill-name warn for 65-char name, got %v", out)
	}
}

func TestCheckDocumentSkillDescriptionHard(t *testing.T) {
	d := &parser.Document{
		Path:        "skills/x/SKILL.md",
		Type:        parser.TypeSkills,
		Name:        "x",
		Description: strings.Repeat("d", 1025),
		Body:        "short",
	}
	out := CheckDocument(d)
	found := false
	for _, msg := range out {
		if strings.Contains(msg, "HARD") && strings.Contains(msg, "skill-description") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected HARD skill-description warn for 1025-char description, got %v", out)
	}
}

func TestCheckDocumentSkillListingClaude(t *testing.T) {
	// description + when_to_use 合计超 1536 触发 claude-code listing 截断提醒
	d := &parser.Document{
		Path:        "skills/x/SKILL.md",
		Type:        parser.TypeSkills,
		Name:        "x",
		Description: strings.Repeat("d", 1000),
		WhenToUse:   strings.Repeat("w", 600),
		Body:        "short",
	}
	out := CheckDocument(d)
	found := false
	for _, msg := range out {
		if strings.Contains(msg, "HARD") && strings.Contains(msg, "skill-listing") && strings.Contains(msg, "claude-code") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected claude-code skill-listing warn for 1600-char combined, got %v", out)
	}
}

func TestCheckDocumentSubagentCopilotHard(t *testing.T) {
	d := &parser.Document{
		Path: "subagents/big.md",
		Type: parser.TypeSubagents,
		Body: strings.Repeat("s", 30001),
	}
	out := CheckDocument(d)
	found := false
	for _, msg := range out {
		if strings.Contains(msg, "HARD") && strings.Contains(msg, "copilot") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected copilot subagent HARD warn for >30000, got %v", out)
	}
}

func TestCheckDocumentClaudeRootNoFakeHard(t *testing.T) {
	// Claude 官方明确根文件全量加载不截断，root-file 不应有 claude-code HARD
	for _, l := range Limits {
		if l.Target == "claude-code" && l.Kind == "root-file" && l.Hard != 0 {
			t.Errorf("claude-code root-file must not carry a fabricated Hard limit (official: loaded in full), got %d", l.Hard)
		}
		if l.Target == "gemini" && l.Kind == "root-file" && l.Hard != 0 {
			t.Errorf("gemini root-file has no documented Hard limit, got %d", l.Hard)
		}
	}
}

func TestCheckTotalRulesAugment(t *testing.T) {
	docs := []*parser.Document{
		{Type: parser.TypeRules, Path: "rules/a.md", Body: strings.Repeat("a", 30000)},
		{Type: parser.TypeRules, Path: "rules/b.md", Body: strings.Repeat("b", 25000)},
	}
	out := CheckTotalRules(docs)
	foundAugment, foundCodex := false, false
	for _, msg := range out {
		if strings.Contains(msg, "augment-code") {
			foundAugment = true
		}
		if strings.Contains(msg, "codex") {
			foundCodex = true
		}
	}
	if !foundAugment {
		t.Errorf("expected augment-code rules-total HARD for 55000 > 49512, got %v", out)
	}
	if !foundCodex {
		t.Errorf("expected codex agents-md-total HARD for 55000 > 32768, got %v", out)
	}
}

func TestCheckRootFileCursorHard(t *testing.T) {
	out := CheckRootFile("cursor", "AGENTS.md", 100001)
	if len(out) == 0 || !strings.Contains(out[0], "HARD") {
		t.Errorf("expected cursor root-file HARD for >100000, got %v", out)
	}
}
