package transformer

import (
	"testing"

	"github.com/StringKe/std-agent/internal/config"
	"github.com/StringKe/std-agent/internal/parser"
)

func TestClinePriorityPrefix(t *testing.T) {
	tr := &Cline{}
	cfg := &config.Config{Inject: false}
	docs := []*parser.Document{
		{Type: parser.TypeRules, Name: "a", Priority: parser.PriorityHigh, Body: "h"},
		{Type: parser.TypeRules, Name: "b", Priority: parser.PriorityNormal, Body: "n"},
		{Type: parser.TypeRules, Name: "c", Priority: parser.PriorityLow, Body: "l"},
	}
	plan, _ := tr.Plan(docs, cfg)
	paths := pathSet(plan)
	for _, want := range []string{".clinerules/100-a.md", ".clinerules/500-b.md", ".clinerules/900-c.md"} {
		if !paths[want] {
			t.Errorf("missing %s, paths: %v", want, paths)
		}
	}
}

func TestClineNativeSkillsDir(t *testing.T) {
	tr := &Cline{}
	cfg := &config.Config{Inject: false}
	docs := []*parser.Document{
		{Type: parser.TypeSkills, Name: "code-review", Description: "review code", Body: "skill body"},
	}
	plan, err := tr.Plan(docs, cfg)
	if err != nil {
		t.Fatal(err)
	}
	paths := pathSet(plan)
	if !paths[".cline/skills/code-review/SKILL.md"] {
		t.Errorf("expected official .cline/skills path, got %v", paths)
	}
	if paths[".clinerules/skills/code-review/SKILL.md"] {
		t.Errorf("legacy fallback skill path must not be written, got %v", paths)
	}
}

func TestClineReferencesOutsideRules(t *testing.T) {
	plan, err := (&Cline{}).Plan([]*parser.Document{{
		Type:        parser.TypeReferences,
		Name:        "design",
		Description: "Design",
		Body:        "ref",
	}}, &config.Config{Inject: false})
	if err != nil {
		t.Fatal(err)
	}
	if !pathSet(plan)[".cline/references/design.md"] {
		t.Errorf("missing isolated references, paths: %v", pathSet(plan))
	}
	if pathSet(plan)[".clinerules/references/design.md"] {
		t.Error("references must not land under .clinerules/")
	}
}
