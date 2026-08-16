package protocol

import (
	"strings"
	"testing"

	"github.com/StringKe/std-agent/internal/config"
	"github.com/StringKe/std-agent/internal/parser"
)

func TestBuildNativeSkillPackage_OptionalInvocationFields(t *testing.T) {
	falseVal := false
	d := &parser.Document{
		Type:                   parser.TypeSkills,
		Name:                   "review",
		Description:            "Review code",
		ArgumentHint:           "[path]",
		DisableModelInvocation: true,
		UserInvocable:          &falseVal,
		Body:                   "body",
	}
	ops := BuildNativeSkillPackage(d, Adapter{
		Name:      "copilot",
		SkillsDir: ".github/skills",
		SkillSupportedFields: []string{
			"name", "description", "argument-hint",
			"user-invocable", "disable-model-invocation",
		},
	}, &config.Config{Inject: false})
	var content string
	for _, op := range ops {
		if strings.HasSuffix(op.Path, "/SKILL.md") {
			content = string(op.Content)
			break
		}
	}
	if content == "" {
		t.Fatal("missing SKILL.md")
	}
	for _, want := range []string{
		"name: review",
		"description: Review code",
		"argument-hint: \"[path]\"",
		"disable-model-invocation: true",
		"user-invocable: false",
	} {
		if !strings.Contains(content, want) {
			t.Errorf("missing %q in:\n%s", want, content)
		}
	}
}

func TestBuildNativeSkillPackage_DefaultUserInvocableOmitted(t *testing.T) {
	d := &parser.Document{
		Type:        parser.TypeSkills,
		Name:        "review",
		Description: "Review code",
		Body:        "body",
	}
	ops := BuildNativeSkillPackage(d, Adapter{
		Name:                 "copilot",
		SkillsDir:            ".github/skills",
		SkillSupportedFields: []string{"name", "description", "user-invocable"},
	}, &config.Config{Inject: false})
	content := string(ops[0].Content)
	if strings.Contains(content, "user-invocable") {
		t.Errorf("unset user_invocable must not render, got:\n%s", content)
	}
}
