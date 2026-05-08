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
