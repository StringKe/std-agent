package protocol

import (
	"strings"
	"testing"

	"std-ai/internal/config"
	"std-ai/internal/parser"
	"std-ai/internal/writer"
)

// cursorTestAdapter 模拟 cursor target 的 adapter（实际 cursorAdapter 在
// Phase 3.5 切换 transformer 委托时定义在 transformer 包）
func cursorTestAdapter() Adapter {
	return Adapter{
		Name:                 "cursor",
		RulesDir:             ".cursor/rules",
		SkillsDir:            ".cursor/skills",
		CommandsDir:          ".cursor/commands",
		MCPPath:              ".cursor/mcp.json",
		MCPTopKey:            "mcpServers",
		GlobsFieldName:       "globs",
		GlobsFieldFormat:     GlobsCommaString,
		SupportsAlwaysApply:  true,
		SupportsDescription:  true,
		InjectExplainer:      true,
		InjectStdaiTypeField: true,
	}
}

func cursorCfg() *config.Config {
	return &config.Config{Inject: false, InjectWhatIs: false}
}

// cursorFindFile 在 plan.Files 中找指定路径；找不到返回 nil
func cursorFindFile(files []writer.FileOp, path string) *writer.FileOp {
	for i := range files {
		if files[i].Path == path {
			return &files[i]
		}
	}
	return nil
}

// cursorPlanPaths 把 plan.Files 列成易读字符串供 t.Fatalf 用
func cursorPlanPaths(files []writer.FileOp) string {
	var b strings.Builder
	for _, op := range files {
		b.WriteString("\n  - ")
		b.WriteString(op.Path)
	}
	return b.String()
}

func TestCursor_RuleFile_MdcSuffix_GlobsComma_AlwaysApplyBool(t *testing.T) {
	doc := &parser.Document{
		Type:        parser.TypeRules,
		Name:        "test-requirements",
		Description: "Always run tests before claiming done",
		AlwaysApply: true,
		ApplyTo:     []string{"**/*.go", "**/*.md"},
		Body:        "RULE BODY",
		Path:        ".stdai/standards/rules/test-requirements.md",
	}
	plan, err := Cursor{}.Plan([]*parser.Document{doc}, cursorTestAdapter(), cursorCfg())
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	want := ".cursor/rules/test-requirements.mdc"
	op := cursorFindFile(plan.Files, want)
	if op == nil {
		t.Fatalf("expected file %q, files:%s", want, cursorPlanPaths(plan.Files))
	}
	s := string(op.Content)
	if !strings.Contains(s, "alwaysApply: true") {
		t.Errorf("expected alwaysApply bool, got:\n%s", s)
	}
	if !strings.Contains(s, "globs:") {
		t.Errorf("expected globs field, got:\n%s", s)
	}
	// globs 是逗号分隔字符串（不是 YAML list）
	if !strings.Contains(s, "**/*.go,**/*.md") {
		t.Errorf("expected comma-joined globs, got:\n%s", s)
	}
	if strings.Contains(s, "globs:\n  -") {
		t.Errorf("globs should be comma string, not YAML list, got:\n%s", s)
	}
	if !strings.Contains(s, "description: Always run tests before claiming done") {
		t.Errorf("expected description frontmatter, got:\n%s", s)
	}
}

func TestCursor_SkillFile_SKILLmdStandard(t *testing.T) {
	doc := &parser.Document{
		Type:        parser.TypeSkills,
		Name:        "code-review",
		Description: "Review code thoroughly",
		WhenToUse:   "when user asks for review",
		License:     "MIT",
		Body:        "SKILL BODY",
		Path:        ".stdai/standards/skills/code-review.md",
	}
	plan, _ := Cursor{}.Plan([]*parser.Document{doc}, cursorTestAdapter(), cursorCfg())
	want := ".cursor/skills/code-review/SKILL.md"
	op := cursorFindFile(plan.Files, want)
	if op == nil {
		t.Fatalf("expected %q, files:%s", want, cursorPlanPaths(plan.Files))
	}
	s := string(op.Content)
	if !strings.Contains(s, "name: code-review") {
		t.Errorf("expected name frontmatter, got:\n%s", s)
	}
	if !strings.Contains(s, "Review code thoroughly when user asks for review") {
		t.Errorf("expected merged description, got:\n%s", s)
	}
	if !strings.Contains(s, "license: MIT") {
		t.Errorf("expected license, got:\n%s", s)
	}
}

func TestCursor_CommandFile_MdSuffix(t *testing.T) {
	doc := &parser.Document{
		Type:        parser.TypeCommands,
		Name:        "release-patch",
		Description: "Cut a patch release",
		Body:        "CMD BODY",
		Path:        ".stdai/standards/commands/release-patch.md",
	}
	plan, _ := Cursor{}.Plan([]*parser.Document{doc}, cursorTestAdapter(), cursorCfg())
	want := ".cursor/commands/release-patch.md"
	op := cursorFindFile(plan.Files, want)
	if op == nil {
		t.Fatalf("expected %q, files:%s", want, cursorPlanPaths(plan.Files))
	}
	s := string(op.Content)
	if !strings.Contains(s, "Cut a patch release") {
		t.Errorf("expected description in body, got:\n%s", s)
	}
	if !strings.Contains(s, "CMD BODY") {
		t.Errorf("expected body, got:\n%s", s)
	}
}

func TestCursor_MCPJSON_McpServersTopKey(t *testing.T) {
	cfg := cursorCfg()
	cfg.MCP = &config.MCPConfig{
		Servers: map[string]config.MCPServer{
			"github": {Command: "uvx", Args: []string{"mcp-server-github"}},
		},
	}
	plan, _ := Cursor{}.Plan(nil, cursorTestAdapter(), cfg)
	want := ".cursor/mcp.json"
	op := cursorFindFile(plan.Files, want)
	if op == nil {
		t.Fatalf("expected %q, files:%s", want, cursorPlanPaths(plan.Files))
	}
	s := string(op.Content)
	if !strings.Contains(s, `"mcpServers"`) {
		t.Errorf("expected mcpServers top key, got:\n%s", s)
	}
}

func TestCursor_Glossary_FallsTo_RulesDir_NoPrefix(t *testing.T) {
	adapter := cursorTestAdapter()
	adapter.InjectTypeGlossary = true
	plan, _ := Cursor{}.Plan(nil, adapter, cursorCfg())
	want := ".cursor/rules/glossary.md"
	op := cursorFindFile(plan.Files, want)
	if op == nil {
		t.Fatalf("expected %q, files:%s", want, cursorPlanPaths(plan.Files))
	}
	// 校验无下划线私有前缀
	for _, file := range plan.Files {
		if strings.Contains(file.Path, "_glossary") {
			t.Errorf("glossary should not use underscore prefix, got path %q", file.Path)
		}
	}
	s := string(op.Content)
	if !strings.Contains(s, "std-ai-type: glossary") {
		t.Errorf("expected frontmatter std-ai-type: glossary, got:\n%s", s)
	}
	if !strings.Contains(s, "std-ai 类型速查") {
		t.Errorf("expected glossary body content, got:\n%s", s)
	}
}

func TestCursor_Glossary_Disabled_NoFile(t *testing.T) {
	adapter := cursorTestAdapter()
	adapter.InjectTypeGlossary = false
	plan, _ := Cursor{}.Plan(nil, adapter, cursorCfg())
	for _, op := range plan.Files {
		if strings.HasSuffix(op.Path, "glossary.md") || strings.Contains(op.Path, "glossary") {
			t.Errorf("InjectTypeGlossary=false should not emit glossary, got %q", op.Path)
		}
	}
}

func TestCursor_ReferencesFallback_GoesToSkillsDir(t *testing.T) {
	doc := &parser.Document{
		Type:        parser.TypeReferences,
		Name:        "transformer-design",
		Description: "Architecture notes",
		Body:        "REF BODY",
		Path:        ".stdai/standards/references/transformer-design.md",
	}
	plan, _ := Cursor{}.Plan([]*parser.Document{doc}, cursorTestAdapter(), cursorCfg())
	want := ".cursor/skills/transformer-design/SKILL.md"
	op := cursorFindFile(plan.Files, want)
	if op == nil {
		t.Fatalf("references fallback should land at %q, files:%s", want, cursorPlanPaths(plan.Files))
	}
	s := string(op.Content)
	if !strings.Contains(s, "std-ai-type: references") {
		t.Errorf("expected std-ai-type: references frontmatter, got:\n%s", s)
	}
	if !strings.Contains(s, "<!-- std-ai degraded references") {
		t.Errorf("expected explainer comment, got:\n%s", s)
	}
	// 路径不应含私有前缀
	for _, file := range plan.Files {
		for _, bad := range []string{"_ref_", "_skill_", "_command_", "_subagent_"} {
			if strings.Contains(file.Path, bad) {
				t.Errorf("path %q must not contain forbidden prefix %q", file.Path, bad)
			}
		}
	}
}

func TestCursor_SubagentsFallback_GoesToRulesSubdir(t *testing.T) {
	doc := &parser.Document{
		Type:        parser.TypeSubagents,
		Name:        "reviewer",
		Description: "Code reviewer",
		Body:        "SUBAGENT BODY",
		Path:        ".stdai/standards/subagents/reviewer.md",
	}
	plan, _ := Cursor{}.Plan([]*parser.Document{doc}, cursorTestAdapter(), cursorCfg())
	want := ".cursor/rules/subagents/reviewer.md"
	op := cursorFindFile(plan.Files, want)
	if op == nil {
		t.Fatalf("subagent fallback should land at %q, files:%s", want, cursorPlanPaths(plan.Files))
	}
	s := string(op.Content)
	if !strings.Contains(s, "std-ai-type: subagents") {
		t.Errorf("expected std-ai-type: subagents frontmatter, got:\n%s", s)
	}
	if !strings.Contains(s, "<!-- std-ai degraded subagents") {
		t.Errorf("expected explainer comment, got:\n%s", s)
	}
	if !strings.Contains(s, "SUBAGENT BODY") {
		t.Errorf("expected body, got:\n%s", s)
	}
	// 路径不应含私有前缀
	for _, file := range plan.Files {
		for _, bad := range []string{"_subagent_", "_ref_", "_skill_", "_command_"} {
			if strings.Contains(file.Path, bad) {
				t.Errorf("path %q must not contain forbidden prefix %q", file.Path, bad)
			}
		}
	}
}

func TestCursor_Disabled_ReturnsEmptyPlan(t *testing.T) {
	adapter := cursorTestAdapter()
	adapter.Disabled = true
	doc := &parser.Document{Type: parser.TypeRules, Name: "x", Body: "B"}
	plan, _ := Cursor{}.Plan([]*parser.Document{doc}, adapter, cursorCfg())
	if len(plan.Files) != 0 {
		t.Errorf("Disabled=true should produce zero files, got %d:%s",
			len(plan.Files), cursorPlanPaths(plan.Files))
	}
	if plan.Target != "cursor" {
		t.Errorf("plan.Target = %q, want %q", plan.Target, "cursor")
	}
}

func TestCursor_FullFixture_AllTypesPresent(t *testing.T) {
	docs := []*parser.Document{
		{Type: parser.TypeRules, Name: "r1", AlwaysApply: true, Body: "RB1", Path: "p1"},
		{Type: parser.TypeSkills, Name: "s1", Description: "skill 1", Body: "SB1", Path: "p2"},
		{Type: parser.TypeCommands, Name: "c1", Body: "CB1", Path: "p3"},
		{Type: parser.TypeReferences, Name: "ref1", Body: "RFB1", Path: "p4"},
		{Type: parser.TypeSubagents, Name: "sa1", Body: "SAB1", Path: "p5"},
	}
	adapter := cursorTestAdapter()
	adapter.InjectTypeGlossary = true
	cfg := cursorCfg()
	cfg.MCP = &config.MCPConfig{Servers: map[string]config.MCPServer{
		"x": {Command: "y"},
	}}
	plan, _ := Cursor{}.Plan(docs, adapter, cfg)

	wantPaths := []string{
		".cursor/mcp.json",
		".cursor/rules/glossary.md",
		".cursor/rules/r1.mdc",
		".cursor/skills/s1/SKILL.md",
		".cursor/commands/c1.md",
		".cursor/skills/ref1/SKILL.md",
		".cursor/rules/subagents/sa1.md",
	}
	for _, want := range wantPaths {
		if cursorFindFile(plan.Files, want) == nil {
			t.Errorf("missing expected path %q. have:%s", want, cursorPlanPaths(plan.Files))
		}
	}
}
