package transformer

import (
	"strings"
	"testing"

	"github.com/StringKe/std-agent/internal/config"
	"github.com/StringKe/std-agent/internal/parser"
)

func TestFactoryOutputs(t *testing.T) {
	tr := &Factory{}
	cfg := &config.Config{Inject: true, InjectWhatIs: false}
	docs := []*parser.Document{
		{Type: parser.TypeRules, Name: "style", Description: "Style 描述", Body: "rule body"},
		{Type: parser.TypeSkills, Name: "review", Description: "Review skill", Body: "skill body"},
		{Type: parser.TypeSubagents, Name: "code-reviewer", Description: "Reviews code", Model: "claude-sonnet-4-5", AllowedTools: []string{"Read", "Grep"}, Body: "You are a code reviewer..."},
	}
	plan, _ := tr.Plan(docs, cfg)
	paths := pathSet(plan)
	for _, want := range []string{
		"AGENTS.md",
		".factory/rules/style.md",
		".factory/skills/review/SKILL.md",
		".factory/droids/code-reviewer.md",
	} {
		if !paths[want] {
			t.Errorf("missing %s in plan: %v", want, paths)
		}
	}
	// 验证 subagent 内容
	c, _ := contentOf(plan, ".factory/droids/code-reviewer.md")
	for _, want := range []string{"name: code-reviewer", "description: Reviews code", "model: claude-sonnet-4-5", "Read", "Grep", "You are a code reviewer"} {
		if !strings.Contains(c, want) {
			t.Errorf("missing %q in droid file:\n%s", want, c)
		}
	}
}

func TestFactoryReferencesOutsideRules(t *testing.T) {
	plan, err := (&Factory{}).Plan([]*parser.Document{{
		Type:        parser.TypeReferences,
		Name:        "design",
		Description: "Design",
		Body:        "ref",
	}}, &config.Config{Inject: false})
	if err != nil {
		t.Fatal(err)
	}
	if !pathSet(plan)[".factory/references/design.md"] {
		t.Errorf("missing isolated references, paths: %v", pathSet(plan))
	}
	if pathSet(plan)[".factory/rules/references/design.md"] {
		t.Error("references must not land under .factory/rules/")
	}
}

func TestFactoryNoGlobsFrontmatter(t *testing.T) {
	tr := &Factory{}
	cfg := &config.Config{Inject: false}
	docs := []*parser.Document{
		{Type: parser.TypeRules, Name: "go-only", ApplyTo: []string{"**/*.go"}, AlwaysApply: false, Body: "go rule body"},
	}
	plan, _ := tr.Plan(docs, cfg)
	c, ok := contentOf(plan, ".factory/rules/go-only.md")
	if !ok {
		t.Fatalf("expected .factory/rules/go-only.md, paths: %v", pathSet(plan))
	}
	// GlobsFieldName="" -> 应跳过 globs / paths / applyTo 字段
	for _, forbid := range []string{"globs:", "paths:", "applyTo:"} {
		if strings.Contains(c, forbid) {
			t.Errorf("factory rule frontmatter should not contain %q (GlobsFieldName=\"\"):\n%s", forbid, c)
		}
	}
}

func TestFactorySkillInvocationFields(t *testing.T) {
	falseVal := false
	plan, err := (&Factory{}).Plan([]*parser.Document{{
		Type:                   parser.TypeSkills,
		Name:                   "review",
		Description:            "Review",
		AllowedTools:           []string{"Read"},
		DisableModelInvocation: true,
		UserInvocable:          &falseVal,
		Body:                   "body",
	}}, &config.Config{Inject: false})
	if err != nil {
		t.Fatal(err)
	}
	c, ok := contentOf(plan, ".factory/skills/review/SKILL.md")
	if !ok {
		t.Fatalf("missing factory skill, paths: %v", pathSet(plan))
	}
	for _, want := range []string{
		"allowed-tools:",
		"- Read",
		"disable-model-invocation: true",
		"user-invocable: false",
	} {
		if !strings.Contains(c, want) {
			t.Errorf("missing %q in:\n%s", want, c)
		}
	}
}
