package transformer

import (
	"strings"
	"testing"

	"github.com/StringKe/std-agent/internal/config"
	"github.com/StringKe/std-agent/internal/parser"
)

func TestJunieOutputs(t *testing.T) {
	plan, err := (&Junie{}).Plan([]*parser.Document{
		{Type: parser.TypeRules, Name: "root-style", Root: true, Body: "Always follow style."},
		{Type: parser.TypeRules, Name: "go-style", Description: "Go style", ApplyTo: []string{"**/*.go"}, Body: "Use gofmt."},
		{Type: parser.TypeSkills, Name: "review", Description: "Review code.", Body: "Steps..."},
		{Type: parser.TypeCommands, Name: "deploy", Description: "Deploy", Body: "go"},
	}, &config.Config{Inject: false})
	if err != nil {
		t.Fatal(err)
	}
	paths := pathSet(plan)
	if !paths["AGENTS.md"] {
		t.Errorf("missing AGENTS.md, paths: %v", paths)
	}
	if paths[".junie/AGENTS.md"] {
		t.Error("must not duplicate shared AGENTS.md into .junie/AGENTS.md")
	}
	root, _ := contentOf(plan, "AGENTS.md")
	if !strings.Contains(root, "Always follow style.") {
		t.Errorf("always-on rule should land in AGENTS.md:\n%s", root)
	}
	foundRule := false
	for p := range paths {
		if strings.HasPrefix(p, ".junie/rules/") && strings.Contains(p, "go-style") {
			foundRule = true
			break
		}
	}
	if !foundRule {
		t.Errorf("path-scoped rule should land in .junie/rules/, paths: %v", paths)
	}
	if !paths[".junie/skills/review/SKILL.md"] {
		t.Errorf("missing junie skill, paths: %v", paths)
	}
}

func TestJunieFallbackPrivate(t *testing.T) {
	plan, err := (&Junie{}).Plan([]*parser.Document{
		{Type: parser.TypeCommands, Name: "deploy", Description: "Deploy", Body: "go"},
		{Type: parser.TypeReferences, Name: "api", Description: "API ref", Body: "details"},
		{Type: parser.TypeSubagents, Name: "linter", Description: "Lint", Body: "lint body"},
	}, &config.Config{Inject: false})
	if err != nil {
		t.Fatal(err)
	}
	paths := pathSet(plan)
	for _, want := range []string{".junie/commands/deploy.md", ".junie/references/api.md", ".junie/subagents/linter.md"} {
		if !paths[want] {
			t.Errorf("missing private fallback %s, paths: %v", want, paths)
		}
	}
}
