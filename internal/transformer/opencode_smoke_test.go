package transformer

import (
	"strings"
	"testing"

	"github.com/StringKe/std-agent/internal/config"
	"github.com/StringKe/std-agent/internal/parser"
)

func TestOpenCodeSkillsAndCommands(t *testing.T) {
	tr := &OpenCode{}
	cfg := &config.Config{Inject: false}
	docs := []*parser.Document{
		{Type: parser.TypeSkills, Name: "review", Description: "Review", Body: "steps"},
		{Type: parser.TypeCommands, Name: "deploy", Description: "Deploy", Body: "go"},
		{Type: parser.TypeRules, Name: "ignored", Body: "rules go to AGENTS.md instead"},
	}
	plan, _ := tr.Plan(docs, cfg)
	paths := pathSet(plan)
	// skills 原生 Agent Skills 标准包（官方 GA，旧的 mode: subagent 降级方案已废弃）
	if !paths[".opencode/skills/review/SKILL.md"] {
		t.Errorf("missing native skill package, paths: %v", paths)
	}
	if paths[".opencode/agents/review.md"] {
		t.Errorf("stale skill-as-subagent output still produced, paths: %v", paths)
	}
	if !paths[".opencode/commands/deploy.md"] {
		t.Errorf("missing command file, paths: %v", paths)
	}
	for p := range paths {
		if strings.HasPrefix(p, "AGENTS") || strings.HasPrefix(p, ".opencode/rules") {
			t.Errorf("opencode should not write rules, got %s", p)
		}
	}
	// 原生 skill 不携带 mode: subagent frontmatter
	if c, ok := contentOf(plan, ".opencode/skills/review/SKILL.md"); ok {
		if strings.Contains(c, "mode: subagent") {
			t.Errorf("native skill should not carry subagent frontmatter:\n%s", c)
		}
		if !strings.Contains(c, "name: review") {
			t.Errorf("native skill missing Agent Skills frontmatter:\n%s", c)
		}
	}
}

func TestOpenCodeNativeSubagent(t *testing.T) {
	tr := &OpenCode{}
	cfg := &config.Config{Inject: false}
	plan, _ := tr.Plan([]*parser.Document{
		{Type: parser.TypeSubagents, Name: "reviewer", Description: "Review", Body: "body"},
	}, cfg)
	c, ok := contentOf(plan, ".opencode/agents/reviewer.md")
	if !ok {
		t.Fatalf("missing native subagent, paths: %v", pathSet(plan))
	}
	if !strings.Contains(c, "mode: subagent") {
		t.Errorf("opencode subagent should set mode: subagent:\n%s", c)
	}
}
