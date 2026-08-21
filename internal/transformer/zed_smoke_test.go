package transformer

import (
	"strings"
	"testing"

	"github.com/StringKe/std-agent/internal/config"
	"github.com/StringKe/std-agent/internal/parser"
)

func TestZedOutputs(t *testing.T) {
	plan, err := (&Zed{}).Plan([]*parser.Document{
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
		t.Errorf("rules should inline into AGENTS.md:\n%s", root)
	}
	if !paths[".agents/skills/review/SKILL.md"] {
		t.Errorf("missing shared skill, paths: %v", paths)
	}
	if paths[".agents/skills/commands/deploy/SKILL.md"] {
		t.Error("zed skills are flat; commands must not nest under .agents/skills/commands/")
	}
}

func TestZedFallbackPrivate(t *testing.T) {
	plan, err := (&Zed{}).Plan([]*parser.Document{
		{Type: parser.TypeCommands, Name: "deploy", Description: "Deploy", Body: "go"},
		{Type: parser.TypeReferences, Name: "api", Description: "API ref", Body: "details"},
		{Type: parser.TypeSubagents, Name: "linter", Description: "Lint", Body: "lint body"},
	}, &config.Config{Inject: false})
	if err != nil {
		t.Fatal(err)
	}
	paths := pathSet(plan)
	for _, want := range []string{".zed/commands/deploy.md", ".zed/references/api.md", ".zed/subagents/linter.md"} {
		if !paths[want] {
			t.Errorf("missing private fallback %s, paths: %v", want, paths)
		}
	}
}

func TestZedCodexSkillByteIdentical(t *testing.T) {
	cfg := &config.Config{Inject: true, InjectWhatIs: false}
	docs := []*parser.Document{
		{Type: parser.TypeSkills, Name: "review", Description: "Review code", Body: "skill body"},
	}
	zedPlan, err := (&Zed{}).Plan(docs, cfg)
	if err != nil {
		t.Fatal(err)
	}
	codexPlan, err := (&Codex{}).Plan(docs, cfg)
	if err != nil {
		t.Fatal(err)
	}
	z, zok := contentOf(zedPlan, ".agents/skills/review/SKILL.md")
	c, cok := contentOf(codexPlan, ".agents/skills/review/SKILL.md")
	if !zok || !cok {
		t.Fatalf("both targets must produce shared skill (zed=%v codex=%v)", zok, cok)
	}
	if z != c {
		t.Errorf("zed and codex differ for shared skill:\nzed:\n%s\ncodex:\n%s", z, c)
	}
}
