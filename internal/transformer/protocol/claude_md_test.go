package protocol

import (
	"strings"
	"testing"

	"github.com/StringKe/std-agent/internal/config"
	"github.com/StringKe/std-agent/internal/parser"
	"github.com/StringKe/std-agent/internal/writer"
)

// claudeTestAdapter 返回与 claude-code transformer 等价的 Adapter literal。
func claudeTestAdapter() Adapter {
	return Adapter{
		Name:                 "claude-code",
		RootFileName:         "CLAUDE.md",
		ManifestSection:      "Imported Rules",
		NestedSupported:      true,
		NestedFileName:       "CLAUDE.md",
		RulesDir:             ".claude/rules",
		SkillsDir:            ".claude/skills",
		CommandsDir:          ".claude/commands",
		SubagentsDir:         ".claude/agents",
		GlobsFieldName:       "paths",
		GlobsFieldFormat:     GlobsList,
		SupportsDescription:  true,
		InjectExplainer:      true,
		InjectStdaiTypeField: true,
		InjectTypeGlossary:   true,
		MCPPath:              ".mcp.json",
		MCPTopKey:            "mcpServers",
	}
}

// claudeTestCfg 构造禁用 marker 注入的最小 Config，便于断言原始 body。
func claudeTestCfg() *config.Config {
	return &config.Config{Inject: false, InjectWhatIs: false, InjectTypeGlossary: true}
}

// claudeFileByPath 在 plan.Files 中查特定路径，返回 (op, ok)。
func claudeFileByPath(plan *writer.Plan, p string) (writer.FileOp, bool) {
	for _, op := range plan.Files {
		if op.Path == p {
			return op, true
		}
	}
	return writer.FileOp{}, false
}

func TestClaudeMD_Disabled(t *testing.T) {
	adapter := claudeTestAdapter()
	adapter.Disabled = true
	plan, err := ClaudeMD{}.Plan([]*parser.Document{
		{Type: parser.TypeRules, Name: "x", Body: "b"},
	}, adapter, claudeTestCfg())
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Files) != 0 {
		t.Errorf("Disabled=true must produce 0 files, got %d", len(plan.Files))
	}
}

func TestClaudeMD_RootWithManifestAndGlossary(t *testing.T) {
	adapter := claudeTestAdapter()
	docs := []*parser.Document{
		{Type: parser.TypeRules, Name: "project", Root: true, Body: "# Project Overview\n\nstd-agent is..."},
		{Type: parser.TypeRules, Name: "style", Description: "code style", ApplyTo: []string{"**/*.go"}, Body: "use gofmt"},
	}
	plan, err := ClaudeMD{}.Plan(docs, adapter, claudeTestCfg())
	if err != nil {
		t.Fatal(err)
	}
	op, ok := claudeFileByPath(plan, "CLAUDE.md")
	if !ok {
		t.Fatalf("expected CLAUDE.md, got files: %v", claudePlanPaths(plan))
	}
	if !op.IsRoot {
		t.Error("CLAUDE.md must have IsRoot=true")
	}
	c := string(op.Content)
	if !strings.Contains(c, "std-agent 类型速查") {
		t.Errorf("expected glossary section in root, got:\n%s", c)
	}
	if !strings.Contains(c, "# Project Overview") {
		t.Errorf("expected root rule body, got:\n%s", c)
	}
	if !strings.Contains(c, "## Imported Rules") {
		t.Errorf("expected manifest section, got:\n%s", c)
	}
	if !strings.Contains(c, "@.claude/rules/style.md") {
		t.Errorf("expected @-import path in manifest, got:\n%s", c)
	}
	if !strings.Contains(c, "code style") {
		t.Errorf("expected rule description in manifest, got:\n%s", c)
	}
}

func TestClaudeMD_RootGlossaryDisabled(t *testing.T) {
	adapter := claudeTestAdapter()
	adapter.InjectTypeGlossary = false
	docs := []*parser.Document{
		{Type: parser.TypeRules, Name: "project", Root: true, Body: "# Title\n\nbody"},
	}
	plan, err := ClaudeMD{}.Plan(docs, adapter, claudeTestCfg())
	if err != nil {
		t.Fatal(err)
	}
	op, _ := claudeFileByPath(plan, "CLAUDE.md")
	c := string(op.Content)
	if strings.Contains(c, "std-agent 类型速查") {
		t.Errorf("InjectTypeGlossary=false should suppress glossary, got:\n%s", c)
	}
	if strings.Contains(c, "auto-injected") {
		t.Errorf("InjectTypeGlossary=false should suppress marker comment, got:\n%s", c)
	}
}

func TestClaudeMD_RootNoRulesPlaceholder(t *testing.T) {
	adapter := claudeTestAdapter()
	plan, err := ClaudeMD{}.Plan([]*parser.Document{
		{Type: parser.TypeSkills, Name: "x", Body: "b"},
	}, adapter, claudeTestCfg())
	if err != nil {
		t.Fatal(err)
	}
	op, _ := claudeFileByPath(plan, "CLAUDE.md")
	c := string(op.Content)
	if !strings.Contains(c, "Project CLAUDE Manifest") {
		t.Errorf("expected placeholder title when no root rule, got:\n%s", c)
	}
	if !strings.Contains(c, "No rules synced.") {
		t.Errorf("expected 'No rules synced.' line, got:\n%s", c)
	}
}

func TestClaudeMD_NestedRoot(t *testing.T) {
	adapter := claudeTestAdapter()
	docs := []*parser.Document{
		{Type: parser.TypeRules, Name: "nested-root", NestedPath: "frontend", Body: "frontend module overview"},
	}
	plan, err := ClaudeMD{}.Plan(docs, adapter, claudeTestCfg())
	if err != nil {
		t.Fatal(err)
	}
	op, ok := claudeFileByPath(plan, "frontend/CLAUDE.md")
	if !ok {
		t.Fatalf("expected frontend/CLAUDE.md, got files: %v", claudePlanPaths(plan))
	}
	if !op.IsRoot {
		t.Error("nested CLAUDE.md must have IsRoot=true")
	}
	c := string(op.Content)
	if strings.Contains(c, "std-agent 类型速查") {
		t.Error("nested CLAUDE.md must not contain glossary")
	}
	if strings.Contains(c, "## Imported Rules") {
		t.Error("nested CLAUDE.md must not contain manifest section")
	}
	if !strings.Contains(c, "frontend module overview") {
		t.Errorf("nested body missing, got:\n%s", c)
	}
}

func TestClaudeMD_RuleFilePathsFrontmatter(t *testing.T) {
	adapter := claudeTestAdapter()
	docs := []*parser.Document{
		{
			Type:        parser.TypeRules,
			Name:        "go-style",
			Description: "Go coding style",
			ApplyTo:     []string{"**/*.go"},
			Body:        "use gofmt",
		},
	}
	plan, err := ClaudeMD{}.Plan(docs, adapter, claudeTestCfg())
	if err != nil {
		t.Fatal(err)
	}
	op, ok := claudeFileByPath(plan, ".claude/rules/go-style.md")
	if !ok {
		t.Fatalf("expected .claude/rules/go-style.md, got files: %v", claudePlanPaths(plan))
	}
	c := string(op.Content)
	if !strings.Contains(c, "paths:") {
		t.Errorf("expected 'paths:' frontmatter (Anthropic dialect), got:\n%s", c)
	}
	for _, bad := range []string{"applyTo:", "globs:", "alwaysApply:"} {
		if strings.Contains(c, bad) {
			t.Errorf("must not contain %q (foreign dialect), got:\n%s", bad, c)
		}
	}
	if !strings.Contains(c, "**/*.go") {
		t.Errorf("expected glob value in paths, got:\n%s", c)
	}
	if !strings.Contains(c, "description: Go coding style") {
		t.Errorf("expected description frontmatter, got:\n%s", c)
	}
}

func TestClaudeMD_SkillPrivateFields(t *testing.T) {
	adapter := claudeTestAdapter()
	docs := []*parser.Document{
		{
			Type:                   parser.TypeSkills,
			Name:                   "code-review",
			Description:            "Review code",
			WhenToUse:              "when user asks for review",
			ArgumentHint:           "[scope]",
			AllowedTools:           []string{"Read", "Grep"},
			Effort:                 "high",
			SkillContext:           "fork",
			Agent:                  "reviewer",
			Shell:                  "bash",
			Arguments:              []string{"scope"},
			Model:                  "claude-sonnet-4-5",
			ApplyTo:                []string{"**/*.go"},
			DisableModelInvocation: true,
			License:                "MIT",
			Compatibility:          ">=1.0",
			Metadata:               map[string]interface{}{"author": "std-agent"},
			Hooks:                  map[string]interface{}{"pre": "echo ready"},
			Body:                   "Do a thorough review.",
		},
	}
	plan, err := ClaudeMD{}.Plan(docs, adapter, claudeTestCfg())
	if err != nil {
		t.Fatal(err)
	}
	op, ok := claudeFileByPath(plan, ".claude/skills/code-review/SKILL.md")
	if !ok {
		t.Fatalf("expected SKILL.md, got files: %v", claudePlanPaths(plan))
	}
	c := string(op.Content)
	mustContain := []string{
		"name: code-review",
		"description: Review code",
		"when_to_use:",
		"argument-hint:",
		"effort: high",
		"context: fork",
		"agent: reviewer",
		"shell: bash",
		"allowed-tools:",
		"  - Read",
		"  - Grep",
		"paths:",
		"  - \"**/*.go\"",
		"arguments:",
		"  - scope",
		"model: claude-sonnet-4-5",
		"disable-model-invocation: true",
		"license: MIT",
		"compatibility:",
		"metadata:",
		"  author: std-agent",
		"hooks:",
		"Do a thorough review.",
	}
	for _, want := range mustContain {
		if !strings.Contains(c, want) {
			t.Errorf("expected %q in SKILL.md, got:\n%s", want, c)
		}
	}
}

func TestClaudeMD_SkillWithPackageFiles(t *testing.T) {
	adapter := claudeTestAdapter()
	docs := []*parser.Document{
		{
			Type:        parser.TypeSkills,
			Name:        "tool-x",
			Description: "tool x",
			Body:        "skill body",
			SkillFiles: []parser.SkillFile{
				{Path: "scripts/run.sh", Raw: []byte("#!/bin/bash\necho hi\n")},
			},
		},
	}
	plan, err := ClaudeMD{}.Plan(docs, adapter, claudeTestCfg())
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := claudeFileByPath(plan, ".claude/skills/tool-x/SKILL.md"); !ok {
		t.Error("missing SKILL.md")
	}
	op, ok := claudeFileByPath(plan, ".claude/skills/tool-x/scripts/run.sh")
	if !ok {
		t.Fatalf("expected packaged script, got files: %v", claudePlanPaths(plan))
	}
	if !strings.Contains(string(op.Content), "echo hi") {
		t.Errorf("packaged script content wrong, got:\n%s", op.Content)
	}
}

func TestClaudeMD_CommandFile(t *testing.T) {
	adapter := claudeTestAdapter()
	docs := []*parser.Document{
		{
			Type:         parser.TypeCommands,
			Name:         "release",
			Description:  "Cut a release",
			ArgumentHint: "[version]",
			AllowedTools: []string{"Bash"},
			Model:        "claude-opus-4-1",
			Body:         "Run the release flow.",
		},
	}
	plan, err := ClaudeMD{}.Plan(docs, adapter, claudeTestCfg())
	if err != nil {
		t.Fatal(err)
	}
	op, ok := claudeFileByPath(plan, ".claude/commands/release.md")
	if !ok {
		t.Fatalf("expected commands file, got files: %v", claudePlanPaths(plan))
	}
	c := string(op.Content)
	for _, want := range []string{
		"description: Cut a release",
		"argument-hint:",
		"allowed-tools:",
		"  - Bash",
		"model: claude-opus-4-1",
		"Run the release flow.",
	} {
		if !strings.Contains(c, want) {
			t.Errorf("expected %q in command file, got:\n%s", want, c)
		}
	}
}

func TestClaudeMD_SubagentFile(t *testing.T) {
	adapter := claudeTestAdapter()
	docs := []*parser.Document{
		{
			Type:            parser.TypeSubagents,
			Name:            "code-reviewer",
			Description:     "Reviews code",
			Model:           "claude-sonnet-4-5",
			AllowedTools:    []string{"Read", "Grep"},
			DisallowedTools: []string{"Write"},
			Background:      true,
			Effort:          "high",
			Isolation:       "worktree",
			Memory:          "project",
			PermissionMode:  "plan",
			MaxTurns:        8,
			PreloadSkills:   []string{"api-conventions"},
			Body:            "You are a strict reviewer.",
		},
	}
	plan, err := ClaudeMD{}.Plan(docs, adapter, claudeTestCfg())
	if err != nil {
		t.Fatal(err)
	}
	op, ok := claudeFileByPath(plan, ".claude/agents/code-reviewer.md")
	if !ok {
		t.Fatalf("expected subagent file, got files: %v", claudePlanPaths(plan))
	}
	c := string(op.Content)
	for _, want := range []string{
		"name: code-reviewer",
		"description: Reviews code",
		"model: claude-sonnet-4-5",
		"tools:",
		"  - Read",
		"  - Grep",
		"disallowedTools:",
		"  - Write",
		"background: true",
		"effort: high",
		"isolation: worktree",
		"memory: project",
		"permissionMode: plan",
		"maxTurns: 8",
		"skills:",
		"  - api-conventions",
		"You are a strict reviewer.",
	} {
		if !strings.Contains(c, want) {
			t.Errorf("expected %q in subagent file, got:\n%s", want, c)
		}
	}
}

// TestClaudeMD_ReferencesGoToSubdirNotSkills 验证 references 走 .claude/references/ 子目录
// v3：不借 .claude/skills/<name>/SKILL.md 路径（那会让 Claude Code 按 skill 触发加载，
// 破坏 references "按需查阅，不自动加载" 语义）。
func TestClaudeMD_ReferencesGoToSubdirNotSkills(t *testing.T) {
	adapter := claudeTestAdapter()
	docs := []*parser.Document{
		{
			Type:        parser.TypeReferences,
			Name:        "transformer-design",
			Description: "transformer architecture notes",
			Body:        "Reference body content.",
		},
	}
	plan, err := ClaudeMD{}.Plan(docs, adapter, claudeTestCfg())
	if err != nil {
		t.Fatal(err)
	}
	op, ok := claudeFileByPath(plan, ".claude/rules/references/transformer-design.md")
	if !ok {
		t.Fatalf("references must go to subdir (not SkillsDir), got files: %v", claudePlanPaths(plan))
	}
	c := string(op.Content)
	if !strings.Contains(c, "std-agent-type: references") {
		t.Errorf("expected std-agent-type: references frontmatter, got:\n%s", c)
	}
	if !strings.Contains(c, "<!-- std-agent degraded references: transformer-design") {
		t.Errorf("expected explainer comment, got:\n%s", c)
	}
	if !strings.Contains(c, "Reference body content.") {
		t.Errorf("expected original body, got:\n%s", c)
	}
	// 反向：references 必须不出现在 .claude/skills/ 路径下（旧设计错误，会被 Claude 当 skill 加载）
	for _, p := range claudePlanPaths(plan) {
		if strings.HasPrefix(p, ".claude/skills/transformer-design") {
			t.Errorf("references must NOT go to skills path (would be auto-loaded as skill), got %q", p)
		}
	}
}

func TestClaudeMD_MCPJSONTopKey(t *testing.T) {
	adapter := claudeTestAdapter()
	cfg := claudeTestCfg()
	cfg.MCP = &config.MCPConfig{
		Servers: map[string]config.MCPServer{
			"fs": {Type: "stdio", Command: "mcp-fs", Args: []string{"--root", "/"}},
		},
	}
	plan, err := ClaudeMD{}.Plan(nil, adapter, cfg)
	if err != nil {
		t.Fatal(err)
	}
	op, ok := claudeFileByPath(plan, ".mcp.json")
	if !ok {
		t.Fatalf("expected .mcp.json, got files: %v", claudePlanPaths(plan))
	}
	c := string(op.Content)
	if !strings.Contains(c, "\"mcpServers\"") {
		t.Errorf("expected top-level mcpServers key (not 'servers'), got:\n%s", c)
	}
	if strings.Contains(c, "\"servers\":") {
		t.Errorf("must not use Copilot 'servers' top key, got:\n%s", c)
	}
	if !strings.Contains(c, "\"fs\"") {
		t.Errorf("expected server name in payload, got:\n%s", c)
	}
}

func TestClaudeMD_NoMCPWhenAbsent(t *testing.T) {
	adapter := claudeTestAdapter()
	plan, err := ClaudeMD{}.Plan(nil, adapter, claudeTestCfg())
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := claudeFileByPath(plan, ".mcp.json"); ok {
		t.Error("must not emit .mcp.json without MCP config")
	}
	if len(plan.Files) != 0 {
		t.Errorf("empty docs + no MCP must produce 0 files, got %d: %v", len(plan.Files), claudePlanPaths(plan))
	}
}

func TestClaudeMD_RootRuleNotFannedOut(t *testing.T) {
	adapter := claudeTestAdapter()
	docs := []*parser.Document{
		{Type: parser.TypeRules, Name: "project", Root: true, Body: "root body"},
	}
	plan, err := ClaudeMD{}.Plan(docs, adapter, claudeTestCfg())
	if err != nil {
		t.Fatal(err)
	}
	// root rule 不应产生 .claude/rules/project.md
	if _, ok := claudeFileByPath(plan, ".claude/rules/project.md"); ok {
		t.Error("root rule must not be fanned out to .claude/rules/")
	}
}

// claudePlanPaths 收集 plan 中全部路径，便于断言失败时打印
func claudePlanPaths(plan *writer.Plan) []string {
	out := make([]string, 0, len(plan.Files))
	for _, op := range plan.Files {
		out = append(out, op.Path)
	}
	return out
}
