package transformer

import (
	"strings"
	"testing"

	"github.com/StringKe/std-agent/internal/config"
	"github.com/StringKe/std-agent/internal/parser"
)

func TestKiroSteeringInclusion(t *testing.T) {
	tr := &Kiro{}
	cfg := &config.Config{Inject: false}
	docs := []*parser.Document{
		{Type: parser.TypeRules, Name: "project", Root: true, Body: "# Project\nroot body"},
		{Type: parser.TypeRules, Name: "always", AlwaysApply: true, Body: "always body"},
		{Type: parser.TypeRules, Name: "go-only", ApplyTo: []string{"**/*.go"}, Body: "go body"},
		{Type: parser.TypeRules, Name: "api", Description: "REST API patterns", Body: "api body"},
		{Type: parser.TypeRules, Name: "manual", Body: "manual body"},
	}
	plan, err := tr.Plan(docs, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if !pathSet(plan)["AGENTS.md"] {
		t.Fatalf("missing AGENTS.md, paths: %v", pathSet(plan))
	}
	root, _ := contentOf(plan, "AGENTS.md")
	if !strings.Contains(root, "root body") {
		t.Errorf("root rule must be in AGENTS.md:\n%s", root)
	}
	if strings.Contains(root, "always body") {
		t.Errorf("nonRoot steering must not be inlined into AGENTS.md:\n%s", root)
	}

	cases := []struct {
		path string
		want string
	}{
		{".kiro/steering/always.md", "inclusion: always"},
		{".kiro/steering/go-only.md", "fileMatchPattern:"},
		{".kiro/steering/api.md", "inclusion: auto"},
		{".kiro/steering/manual.md", "inclusion: manual"},
	}
	for _, tc := range cases {
		c, ok := contentOf(plan, tc.path)
		if !ok {
			t.Errorf("missing %s, paths: %v", tc.path, pathSet(plan))
			continue
		}
		if !strings.Contains(c, tc.want) {
			t.Errorf("%s missing %q:\n%s", tc.path, tc.want, c)
		}
	}
}

func TestKiroNativeSidecars(t *testing.T) {
	plan, err := (&Kiro{}).Plan([]*parser.Document{
		{Type: parser.TypeSkills, Name: "review", Description: "Review", Body: "skill"},
		{Type: parser.TypeCommands, Name: "ship", Description: "Ship", Body: "cmd"},
		{Type: parser.TypeSubagents, Name: "reviewer", Description: "Reviewer", Model: "claude-sonnet-5", AllowedTools: []string{"read"}, Body: "You review."},
		{Type: parser.TypeReferences, Name: "design", Description: "Design", Body: "ref"},
	}, &config.Config{Inject: false})
	if err != nil {
		t.Fatal(err)
	}
	paths := pathSet(plan)
	for _, want := range []string{
		".kiro/skills/review/SKILL.md",
		".kiro/skills/commands/ship/SKILL.md",
		".kiro/agents/reviewer.md",
		".kiro/references/design.md",
	} {
		if !paths[want] {
			t.Errorf("missing %s, paths: %v", want, paths)
		}
	}
	if paths[".kiro/steering/references/design.md"] {
		t.Error("references must not land under steering (CLI loads all steering)")
	}
	agent, _ := contentOf(plan, ".kiro/agents/reviewer.md")
	for _, want := range []string{"name: reviewer", "description: Reviewer", "model: claude-sonnet-5", "You review."} {
		if !strings.Contains(agent, want) {
			t.Errorf("agent missing %q:\n%s", want, agent)
		}
	}
}

func TestKiroNestedAgents(t *testing.T) {
	plan, _ := (&Kiro{}).Plan([]*parser.Document{
		{Type: parser.TypeRules, Name: "api", Root: true, NestedPath: "services/api", Body: "api root"},
	}, &config.Config{Inject: false})
	if !pathSet(plan)["services/api/AGENTS.md"] {
		t.Errorf("missing nested AGENTS.md, paths: %v", pathSet(plan))
	}
}
