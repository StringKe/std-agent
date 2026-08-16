package transformer

import (
	"strings"
	"testing"

	"github.com/StringKe/std-agent/internal/config"
	"github.com/StringKe/std-agent/internal/parser"
)

func TestCopilotRulesSplit(t *testing.T) {
	tr := &Copilot{}
	cfg := &config.Config{Inject: false}
	docs := []*parser.Document{
		{Type: parser.TypeRules, Name: "general-a", Body: "a"},
		{Type: parser.TypeRules, Name: "ts-only", ApplyTo: []string{"**/*.ts"}, Body: "ts"},
	}
	plan, _ := tr.Plan(docs, cfg)
	paths := pathSet(plan)
	if !paths[".github/copilot-instructions.md"] {
		t.Errorf("missing copilot-instructions.md, paths: %v", paths)
	}
	if !paths[".github/instructions/ts-only.instructions.md"] {
		t.Errorf("missing ts-only.instructions.md, paths: %v", paths)
	}
}

func TestCopilotNativeSkillInvocationFields(t *testing.T) {
	falseVal := false
	plan, err := (&Copilot{}).Plan([]*parser.Document{{
		Type:                   parser.TypeSkills,
		Name:                   "review",
		Description:            "Review",
		ArgumentHint:           "[file]",
		DisableModelInvocation: true,
		UserInvocable:          &falseVal,
		Body:                   "body",
	}}, &config.Config{Inject: false})
	if err != nil {
		t.Fatal(err)
	}
	c, ok := contentOf(plan, ".github/skills/review/SKILL.md")
	if !ok {
		t.Fatalf("missing native skill, paths: %v", pathSet(plan))
	}
	for _, want := range []string{
		"argument-hint: \"[file]\"",
		"disable-model-invocation: true",
		"user-invocable: false",
	} {
		if !strings.Contains(c, want) {
			t.Errorf("missing %q in:\n%s", want, c)
		}
	}
}
