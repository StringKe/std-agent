package transformer

import (
	"strings"
	"testing"

	"github.com/StringKe/std-agent/internal/config"
	"github.com/StringKe/std-agent/internal/parser"
)

func TestGooseOutputs(t *testing.T) {
	plan, err := (&Goose{}).Plan([]*parser.Document{
		{Type: parser.TypeRules, Name: "naming", Description: "命名", Body: "Use clear names."},
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
	root, _ := contentOf(plan, "AGENTS.md")
	if !strings.Contains(root, "Use clear names.") {
		t.Errorf("nonRoot rule should inline into AGENTS.md:\n%s", root)
	}
	if !paths[".agents/skills/review/SKILL.md"] {
		t.Errorf("missing shared skill, paths: %v", paths)
	}
	if !paths[".agents/skills/commands/deploy/SKILL.md"] {
		t.Errorf("missing command-as-skill, paths: %v", paths)
	}
	if paths[".goosehints"] {
		t.Error("must not write .goosehints; AGENTS.md is the official default context file")
	}
}

func TestGooseFallbackPrivate(t *testing.T) {
	plan, _ := (&Goose{}).Plan([]*parser.Document{
		{Type: parser.TypeReferences, Name: "api", Description: "API ref", Body: "details"},
		{Type: parser.TypeSubagents, Name: "linter", Description: "Lint", Body: "lint body"},
	}, &config.Config{Inject: false})
	paths := pathSet(plan)
	if !paths[".goose/references/api.md"] {
		t.Errorf("missing private references fallback, paths: %v", paths)
	}
	if !paths[".goose/subagents/linter.md"] {
		t.Errorf("missing private subagents fallback, paths: %v", paths)
	}
}

func TestGooseCodexSkillByteIdentical(t *testing.T) {
	cfg := &config.Config{Inject: true, InjectWhatIs: false}
	docs := []*parser.Document{
		{Type: parser.TypeSkills, Name: "review", Description: "Review code", WhenToUse: "on review", Body: "skill body"},
		{Type: parser.TypeCommands, Name: "deploy", Description: "Deploy", Body: "deploy body"},
	}
	goosePlan, err := (&Goose{}).Plan(docs, cfg)
	if err != nil {
		t.Fatal(err)
	}
	codexPlan, err := (&Codex{}).Plan(docs, cfg)
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{".agents/skills/review/SKILL.md", ".agents/skills/commands/deploy/SKILL.md"} {
		g, gok := contentOf(goosePlan, p)
		c, cok := contentOf(codexPlan, p)
		if !gok || !cok {
			t.Fatalf("both targets must produce %s (goose=%v codex=%v)", p, gok, cok)
		}
		if g != c {
			t.Errorf("goose and codex differ for shared %s:\ngoose:\n%s\ncodex:\n%s", p, g, c)
		}
	}
}
