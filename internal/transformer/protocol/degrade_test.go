package protocol

import (
	"strings"
	"testing"

	"std-ai/internal/config"
	"std-ai/internal/parser"
)

func TestDefaultFallbackSubdir(t *testing.T) {
	cases := map[parser.DocType]string{
		parser.TypeRules:      "",
		parser.TypeSkills:     "skills",
		parser.TypeCommands:   "commands",
		parser.TypeReferences: "references",
		parser.TypeSubagents:  "subagents",
	}
	for typ, want := range cases {
		got := defaultFallbackSubdir(typ)
		if got != want {
			t.Errorf("defaultFallbackSubdir(%q) = %q, want %q", typ, got, want)
		}
	}
}

func newTestCfg() *config.Config {
	return &config.Config{Inject: false, InjectWhatIs: false}
}

func TestBuildDegradedFileOp_PathNoPrefix(t *testing.T) {
	doc := &parser.Document{
		Type:        parser.TypeReferences,
		Name:        "transformer-design",
		Description: "transformer architecture notes",
		Body:        "BODY",
		Path:        ".stdai/standards/references/transformer-design.md",
	}
	adapter := Adapter{
		Name:                 "foo",
		RulesDir:             ".foo/rules",
		FallbackDir:          ".foo/rules",
		InjectExplainer:      true,
		InjectStdaiTypeField: true,
	}
	op := BuildDegradedFileOp(doc, adapter, newTestCfg())

	wantPath := ".foo/rules/references/transformer-design.md"
	if op.Path != wantPath {
		t.Fatalf("path = %q, want %q", op.Path, wantPath)
	}
	for _, bad := range []string{"_ref_", "_skill_", "_command_", "_subagent_"} {
		if strings.Contains(op.Path, bad) {
			t.Errorf("path %q should not contain forbidden prefix %q", op.Path, bad)
		}
	}
}

func TestBuildDegradedFileOp_StdaiTypeField(t *testing.T) {
	doc := &parser.Document{
		Type: parser.TypeReferences,
		Name: "background",
		Body: "BODY",
	}
	adapter := Adapter{
		Name:                 "foo",
		RulesDir:             ".foo/rules",
		InjectStdaiTypeField: true,
	}
	op := BuildDegradedFileOp(doc, adapter, newTestCfg())
	got := string(op.Content)
	if !strings.Contains(got, "std-ai-type: references") {
		t.Errorf("expected std-ai-type frontmatter, got:\n%s", got)
	}
}

func TestBuildDegradedFileOp_NoStdaiTypeField(t *testing.T) {
	doc := &parser.Document{
		Type: parser.TypeReferences,
		Name: "background",
		Body: "BODY",
	}
	adapter := Adapter{
		Name:                 "foo",
		RulesDir:             ".foo/rules",
		InjectStdaiTypeField: false,
	}
	op := BuildDegradedFileOp(doc, adapter, newTestCfg())
	got := string(op.Content)
	if strings.Contains(got, "std-ai-type:") {
		t.Errorf("InjectStdaiTypeField=false should skip field, got:\n%s", got)
	}
}

func TestBuildDegradedFileOp_ExplainerHeader(t *testing.T) {
	doc := &parser.Document{
		Type: parser.TypeReferences,
		Name: "ref-x",
		Body: "REAL BODY",
	}
	adapter := Adapter{
		Name:            "foo",
		RulesDir:        ".foo/rules",
		InjectExplainer: true,
	}
	op := BuildDegradedFileOp(doc, adapter, newTestCfg())
	got := string(op.Content)
	if !strings.Contains(got, "<!-- std-ai degraded references: ref-x -->") {
		t.Errorf("expected explainer comment, got:\n%s", got)
	}
	if !strings.Contains(got, "Reference is background material") {
		t.Errorf("expected references semantics, got:\n%s", got)
	}
	if !strings.Contains(got, "REAL BODY") {
		t.Error("body should still be present")
	}
}

func TestBuildDegradedFileOp_ExplainerOverride(t *testing.T) {
	doc := &parser.Document{
		Type: parser.TypeCommands,
		Name: "c",
		Body: "BODY",
	}
	adapter := Adapter{
		Name:            "foo",
		RulesDir:        ".foo/rules",
		InjectExplainer: true,
		InjectExplainerOverride: map[parser.DocType]bool{
			parser.TypeCommands: false,
		},
	}
	op := BuildDegradedFileOp(doc, adapter, newTestCfg())
	if strings.Contains(string(op.Content), "<!-- std-ai degraded") {
		t.Error("override=false should suppress explainer")
	}
}

func TestBuildDegradedFileOp_FallbackSubdirOverride(t *testing.T) {
	doc := &parser.Document{Type: parser.TypeSkills, Name: "x", Body: "B"}
	adapter := Adapter{
		Name:        "foo",
		RulesDir:    ".foo/rules",
		FallbackDir: ".foo/extra",
		FallbackSubdir: map[parser.DocType]string{
			parser.TypeSkills: "abilities",
		},
	}
	op := BuildDegradedFileOp(doc, adapter, newTestCfg())
	if op.Path != ".foo/extra/abilities/x.md" {
		t.Errorf("path = %q", op.Path)
	}
}

func TestBuildDegradedFileOp_SubagentCLIBody(t *testing.T) {
	doc := &parser.Document{
		Type:        parser.TypeSubagents,
		Name:        "reviewer",
		Description: "Code reviewer",
		Body:        "You are a strict reviewer.",
	}
	adapter := Adapter{
		Name:              "copilot",
		RulesDir:          ".github/instructions",
		InjectExplainer:   true,
		SubagentInvokeCmd: "claude --agent {name}",
	}
	op := BuildDegradedFileOp(doc, adapter, newTestCfg())
	got := string(op.Content)
	if !strings.Contains(got, "claude --agent reviewer") {
		t.Errorf("expected CLI invocation with substituted name, got:\n%s", got)
	}
	if !strings.Contains(got, "```bash") {
		t.Errorf("expected shell fenced block, got:\n%s", got)
	}
	if !strings.Contains(got, "## How to spawn") {
		t.Errorf("expected spawn instructions section, got:\n%s", got)
	}
}

func TestBuildDegradedSkillPackage_Path(t *testing.T) {
	doc := &parser.Document{
		Type:        parser.TypeReferences,
		Name:        "design",
		Description: "design notes",
		Body:        "BODY",
	}
	adapter := Adapter{
		Name:                 "claude-code",
		SkillsDir:            ".claude/skills",
		InjectExplainer:      true,
		InjectStdaiTypeField: true,
	}
	ops := BuildDegradedSkillPackage(doc, adapter, newTestCfg())
	if len(ops) != 1 {
		t.Fatalf("expected 1 op for skill without SkillFiles, got %d", len(ops))
	}
	if ops[0].Path != ".claude/skills/design/SKILL.md" {
		t.Errorf("path = %q", ops[0].Path)
	}
	got := string(ops[0].Content)
	if !strings.Contains(got, "name: design") {
		t.Errorf("expected name frontmatter, got:\n%s", got)
	}
	if !strings.Contains(got, "std-ai-type: references") {
		t.Errorf("expected std-ai-type frontmatter, got:\n%s", got)
	}
	if !strings.Contains(got, "<!-- std-ai degraded references: design") {
		t.Errorf("expected explainer comment, got:\n%s", got)
	}
}

func TestBuildDegradedSkillPackage_FallbackSkillsDir(t *testing.T) {
	doc := &parser.Document{Type: parser.TypeReferences, Name: "x", Body: "B"}
	adapter := Adapter{
		Name:     "tool",
		RulesDir: ".tool/rules",
		// SkillsDir 空，应走 fallback：<FallbackDir or RulesDir>/skills/<name>/SKILL.md
	}
	ops := BuildDegradedSkillPackage(doc, adapter, newTestCfg())
	if ops[0].Path != ".tool/rules/skills/x/SKILL.md" {
		t.Errorf("path = %q", ops[0].Path)
	}
}
