package parser

import (
	"strings"
	"testing"
)

func TestParseFullFrontmatter(t *testing.T) {
	src := `---
type: rules
name: coding-style
version: "1.2"
description: General style
targets:
  - claude-code
  - codex
priority: high
tags: [style]
applyTo:
  - "**/*.go"
alwaysApply: false
---

# Coding Style

Always use meaningful names.
`
	doc, err := Parse("rules/coding-style.md", []byte(src))
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if doc.Type != TypeRules {
		t.Errorf("Type = %s, want rules", doc.Type)
	}
	if doc.Name != "coding-style" {
		t.Errorf("Name = %s", doc.Name)
	}
	if doc.Priority != PriorityHigh {
		t.Errorf("Priority = %s", doc.Priority)
	}
	if len(doc.Targets) != 2 || doc.Targets[0] != "claude-code" {
		t.Errorf("Targets = %v", doc.Targets)
	}
	if !strings.Contains(doc.Body, "Always use meaningful names") {
		t.Errorf("Body missing rule text: %q", doc.Body)
	}
}

func TestParseNoFrontmatter(t *testing.T) {
	src := `# Just markdown

No frontmatter at all.`
	doc, err := Parse("rules/free.md", []byte(src))
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if doc.Type != TypeRules {
		t.Errorf("Type = %s, want fallback rules", doc.Type)
	}
	if doc.Name != "free" {
		t.Errorf("Name = %s", doc.Name)
	}
}

func TestParseInferNameFromPath(t *testing.T) {
	src := `---
type: skills
description: Foo skill
---
body`
	doc, err := Parse("skills/code-review/SKILL.md", []byte(src))
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if doc.Type != TypeSkills {
		t.Errorf("Type = %s", doc.Type)
	}
	if doc.Name != "skill" {
		t.Errorf("Name = %s, want fallback skill", doc.Name)
	}
}

func TestParseInvalidType(t *testing.T) {
	src := `---
type: invalid
name: foo
---
body`
	if _, err := Parse("rules/x.md", []byte(src)); err == nil {
		t.Error("expected error on invalid type")
	}
}

func TestParseConflictTargetsExclude(t *testing.T) {
	src := `---
type: rules
name: foo
targets: [claude-code]
exclude_targets: [codex]
---
body`
	if _, err := Parse("rules/x.md", []byte(src)); err == nil {
		t.Error("expected error on targets+exclude_targets conflict")
	}
}

func TestParseInvalidName(t *testing.T) {
	src := `---
type: rules
name: Bad_Name
---
body`
	if _, err := Parse("rules/x.md", []byte(src)); err == nil {
		t.Error("expected error on invalid name")
	}
}

func TestParseInvalidPriority(t *testing.T) {
	src := `---
type: rules
name: foo
priority: critical
---
body`
	if _, err := Parse("rules/x.md", []byte(src)); err == nil {
		t.Error("expected error on invalid priority")
	}
}

func TestSplitFrontmatterLF(t *testing.T) {
	src := []byte("---\nfoo: bar\n---\nbody\n")
	front, body, ok := splitFrontmatter(src)
	if !ok {
		t.Fatal("expected ok")
	}
	if !strings.Contains(string(front), "foo: bar") {
		t.Errorf("front = %q", front)
	}
	if string(body) != "body\n" {
		t.Errorf("body = %q", body)
	}
}

func TestSplitFrontmatterCRLF(t *testing.T) {
	src := []byte("---\r\nfoo: bar\r\n---\r\nbody\r\n")
	_, body, ok := splitFrontmatter(src)
	if !ok {
		t.Fatal("expected ok")
	}
	if !strings.Contains(string(body), "body") {
		t.Errorf("body = %q", body)
	}
}

func TestSplitFrontmatterBOM(t *testing.T) {
	src := append([]byte{0xEF, 0xBB, 0xBF}, []byte("---\nx: 1\n---\nbody\n")...)
	_, body, ok := splitFrontmatter(src)
	if !ok {
		t.Fatal("expected ok with BOM")
	}
	if !strings.Contains(string(body), "body") {
		t.Errorf("body = %q", body)
	}
}

func TestParseV12SkillFields(t *testing.T) {
	src := `---
type: skills
name: code-review
description: Review skill
when_to_use: When user asks for code review or PR review
arguments:
  - filename
  - severity
effort: high
context: fork
agent: review-subagent
shell: bash
license: MIT
compatibility: claude-code-4+
metadata:
  author: foo
  version: "1.0"
  tags:
    - quality
    - security
hooks:
  pre_skill:
    - echo start
---
body
`
	doc, err := Parse("skills/code-review/SKILL.md", []byte(src))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if doc.WhenToUse == "" {
		t.Error("WhenToUse should be set")
	}
	if len(doc.Arguments) != 2 || doc.Arguments[0] != "filename" {
		t.Errorf("Arguments = %v", doc.Arguments)
	}
	if doc.Effort != "high" {
		t.Errorf("Effort = %s", doc.Effort)
	}
	if doc.SkillContext != "fork" {
		t.Errorf("SkillContext = %s", doc.SkillContext)
	}
	if doc.Agent != "review-subagent" {
		t.Errorf("Agent = %s", doc.Agent)
	}
	if doc.Shell != "bash" {
		t.Errorf("Shell = %s", doc.Shell)
	}
	if doc.License != "MIT" {
		t.Errorf("License = %s", doc.License)
	}
	if doc.Compatibility != "claude-code-4+" {
		t.Errorf("Compatibility = %s", doc.Compatibility)
	}
	if doc.Metadata["author"] != "foo" {
		t.Errorf("Metadata[author] = %v", doc.Metadata["author"])
	}
	if doc.Hooks == nil {
		t.Error("Hooks should be set")
	}
}

func TestParseSubagentOfficialFields(t *testing.T) {
	src := `---
type: subagents
name: code-reviewer
description: Reviews code
isolation: worktree
memory: project
permission_mode: plan
max_turns: 8
preload_skills:
  - api-conventions
background: true
---
body
`
	doc, err := Parse("subagents/code-reviewer.md", []byte(src))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if doc.Isolation != "worktree" {
		t.Errorf("Isolation = %s", doc.Isolation)
	}
	if doc.Memory != "project" {
		t.Errorf("Memory = %s", doc.Memory)
	}
	if doc.PermissionMode != "plan" {
		t.Errorf("PermissionMode = %s", doc.PermissionMode)
	}
	if doc.MaxTurns != 8 {
		t.Errorf("MaxTurns = %d", doc.MaxTurns)
	}
	if len(doc.PreloadSkills) != 1 || doc.PreloadSkills[0] != "api-conventions" {
		t.Errorf("PreloadSkills = %v", doc.PreloadSkills)
	}
	if !doc.Background {
		t.Error("Background should be true")
	}
}

func TestParseEmptyV12FieldsDefault(t *testing.T) {
	src := `---
type: skills
name: foo
description: x
---
body`
	doc, err := Parse("skills/foo/SKILL.md", []byte(src))
	if err != nil {
		t.Fatal(err)
	}
	if doc.WhenToUse != "" || doc.Effort != "" || doc.License != "" {
		t.Error("v1.2 fields should be empty default")
	}
	if doc.Metadata != nil {
		t.Error("Metadata should be nil default")
	}
	if doc.Arguments != nil {
		t.Error("Arguments should be nil default")
	}
}

func TestPriorityRank(t *testing.T) {
	cases := []struct {
		p Priority
		r int
	}{
		{PriorityHigh, 100},
		{PriorityNormal, 500},
		{"", 500},
		{PriorityLow, 900},
		{"unknown", 500},
	}
	for _, tc := range cases {
		if got := PriorityRank(tc.p); got != tc.r {
			t.Errorf("PriorityRank(%q) = %d, want %d", tc.p, got, tc.r)
		}
	}
}

func TestParseGlobsAlias(t *testing.T) {
	raw := []byte(`---
type: rules
name: g
globs:
  - "**/*.go"
  - "**/*.ts"
---
body
`)
	d, err := Parse("rules/g.md", raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(d.ApplyTo) != 2 || d.ApplyTo[0] != "**/*.go" {
		t.Errorf("globs not merged into ApplyTo: %v", d.ApplyTo)
	}
}

func TestParseApplyToAndGlobsMerged(t *testing.T) {
	raw := []byte(`---
type: rules
name: g
applyTo:
  - "**/*.go"
globs:
  - "**/*.go"
  - "**/*.ts"
---
body
`)
	d, _ := Parse("rules/g.md", raw)
	// applyTo 优先 + globs 兜底，去重后应剩 2 个
	if len(d.ApplyTo) != 2 {
		t.Errorf("expected dedup to 2, got %v", d.ApplyTo)
	}
}

func TestParseTargetSpecificPaths(t *testing.T) {
	raw := []byte(`---
type: rules
name: g
applyTo:
  - "**/*.go"
claudecode:
  paths:
    - "**/*Service.java"
    - "**/*Logic.java"
cursor:
  paths:
    - "src/**/*.ts"
---
body
`)
	d, err := Parse("rules/g.md", raw)
	if err != nil {
		t.Fatal(err)
	}
	if got := d.TargetPaths["claudecode"]; len(got) != 2 || got[0] != "**/*Service.java" {
		t.Errorf("claudecode.paths not parsed: %v", d.TargetPaths)
	}
	if got := d.TargetPaths["cursor"]; len(got) != 1 || got[0] != "src/**/*.ts" {
		t.Errorf("cursor.paths not parsed: %v", d.TargetPaths)
	}
	// 全局 ApplyTo 仍保留
	if len(d.ApplyTo) != 1 || d.ApplyTo[0] != "**/*.go" {
		t.Errorf("global applyTo lost: %v", d.ApplyTo)
	}
}

func TestParseTargetPathsIgnoresUnknownKeys(t *testing.T) {
	raw := []byte(`---
type: rules
name: g
custom-tool:
  paths: ["x"]
---
body
`)
	d, _ := Parse("rules/g.md", raw)
	if _, ok := d.TargetPaths["custom-tool"]; ok {
		t.Error("unknown target key should be ignored")
	}
}

func TestParseSubagentsType(t *testing.T) {
	raw := []byte(`---
type: subagents
name: code-reviewer
description: Reviews code carefully
model: claude-sonnet-4-5
allowed_tools: [Read, Grep]
---
You are a strict code reviewer.
`)
	d, err := Parse("agents/code-reviewer.md", raw)
	if err != nil {
		t.Fatal(err)
	}
	if d.Type != TypeSubagents {
		t.Errorf("type = %s, want subagents", d.Type)
	}
	if d.Name != "code-reviewer" {
		t.Errorf("name = %s", d.Name)
	}
	if len(d.AllowedTools) != 2 {
		t.Errorf("allowed_tools = %v", d.AllowedTools)
	}
}

func TestParseNestedRoot(t *testing.T) {
	raw := []byte(`---
type: rules
name: auth-module
description: Auth 模块说明
---
# Auth 模块
内容
`)
	d, err := Parse("nested/igx-modules/auth/root.md", raw)
	if err != nil {
		t.Fatal(err)
	}
	if !d.Root {
		t.Error("nested root should set Root=true")
	}
	if d.NestedPath != "igx-modules/auth" {
		t.Errorf("NestedPath = %q, want igx-modules/auth", d.NestedPath)
	}
}

func TestParseNestedRootNoFrontmatter(t *testing.T) {
	raw := []byte(`# 子模块
直接内容
`)
	d, err := Parse("nested/src/api/v1/root.md", raw)
	if err != nil {
		t.Fatal(err)
	}
	if !d.Root || d.NestedPath != "src/api/v1" {
		t.Errorf("Root=%v NestedPath=%q", d.Root, d.NestedPath)
	}
}

func TestParseTopRootNotNested(t *testing.T) {
	d, _ := Parse("root.md", []byte("body"))
	if !d.Root {
		t.Error("top root.md should set Root=true")
	}
	if d.NestedPath != "" {
		t.Errorf("top root should have empty NestedPath, got %q", d.NestedPath)
	}
}
