package transformer

import (
	"strings"
	"testing"

	"std-ai/internal/config"
	"std-ai/internal/parser"
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
	// xAI 官方 SkillsDir=.grok/skills（Agent Skills 标准）；其他 type 无原生 -> fallback
	for _, want := range []string{
		".grok/skills/code-review/SKILL.md", // 原生 SkillsDir
		".grok/rules/commands/deploy.md",
		".grok/rules/references/api-spec.md",
		".grok/rules/subagents/reviewer.md",
	} {
		if !paths[want] {
			t.Errorf("missing fallback path %s, paths: %v", want, paths)
		}
	}
}
