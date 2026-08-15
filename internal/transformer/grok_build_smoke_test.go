package transformer

import (
	"strings"
	"testing"

	"github.com/StringKe/std-agent/internal/config"
	"github.com/StringKe/std-agent/internal/parser"
)

func TestGrokBuildOutputs(t *testing.T) {
	tr := &GrokBuild{}
	cfg := &config.Config{Inject: true, InjectWhatIs: false}
	docs := []*parser.Document{
		{Type: parser.TypeRules, Name: "naming", Description: "命名规范", Body: "naming body"},
	}
	plan, err := tr.Plan(docs, cfg)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	paths := pathSet(plan)
	if !paths["AGENTS.md"] {
		t.Errorf("missing AGENTS.md, paths: %v", paths)
	}
	// RulesDir 为空 -> nonRoot rule 应 inline 到 AGENTS.md，不 fan-out 到子目录
	root, _ := contentOf(plan, "AGENTS.md")
	for _, want := range []string{"naming body", "命名规范"} {
		if !strings.Contains(root, want) {
			t.Errorf("AGENTS.md missing %q:\n%s", want, root)
		}
	}
}

func TestGrokBuildFallbackPaths(t *testing.T) {
	tr := &GrokBuild{}
	cfg := &config.Config{Inject: false}
	docs := []*parser.Document{
		{Type: parser.TypeSkills, Name: "code-review", Description: "Review code.", Body: "Steps..."},
		{Type: parser.TypeCommands, Name: "deploy", Description: "Deploy", Body: "go"},
		{Type: parser.TypeReferences, Name: "api-spec", Description: "API spec", Body: "ref body"},
		{Type: parser.TypeSubagents, Name: "reviewer", Description: "Reviewer", Body: "agent body"},
	}
	plan, err := tr.Plan(docs, cfg)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	paths := pathSet(plan)
	// skills 原生 .grok/skills；commands 原生 .grok/commands；
	// subagents 原生 .grok/agents；references 仍落 .grok/docs
	for _, want := range []string{
		".grok/skills/code-review/SKILL.md",
		".grok/commands/deploy.md",
		".grok/docs/references/api-spec.md",
		".grok/agents/reviewer.md",
	} {
		if !paths[want] {
			t.Errorf("missing path %s, paths: %v", want, paths)
		}
	}
	for p := range paths {
		if strings.HasPrefix(p, ".grok/rules/") {
			t.Errorf("degraded output must stay out of session-loaded .grok/rules/: %s", p)
		}
	}
}
