package transformer

import (
	"strings"
	"testing"

	"std-ai/internal/config"
	"std-ai/internal/parser"
	"std-ai/internal/writer"
)

func TestAllTransformersSmoke(t *testing.T) {
	cfg := &config.Config{Inject: true, InjectWhatIs: false}
	docs := []*parser.Document{
		{Type: parser.TypeRules, Name: "coding-style", Description: "Style", Body: "Use clear names.", ApplyTo: []string{"**/*.go"}},
		{Type: parser.TypeRules, Name: "always-on", AlwaysApply: true, Body: "Always."},
		{Type: parser.TypeSkills, Name: "code-review", Description: "Review code.", Body: "Steps..."},
		{Type: parser.TypeCommands, Name: "review", Description: "Run review.", Body: "Please review.", AllowedTools: []string{"Read"}},
	}

	expected := map[string]bool{
		"claude-code":  true,
		"codex":        true,
		"cursor":       true,
		"copilot":      true,
		"windsurf":     true,
		"gemini":       true,
		"aider":        false, // noop
		"cline":        true,
		"opencode":     true,
		"continue-dev": true,
		"antigravity":  true,
	}

	for name, tr := range Registry {
		t.Run(name, func(t *testing.T) {
			plan, err := tr.Plan(docs, cfg)
			if err != nil {
				t.Fatalf("%s: %v", name, err)
			}
			if plan == nil {
				t.Fatalf("%s: nil plan", name)
			}
			if expected[name] && len(plan.Files) == 0 {
				t.Errorf("%s expected to produce files, got 0", name)
			}
			if !expected[name] && len(plan.Files) != 0 {
				t.Errorf("%s expected noop, got %d files", name, len(plan.Files))
			}
			for _, f := range plan.Files {
				if f.Path == "" {
					t.Errorf("%s: empty path in fileop", name)
				}
				if len(f.Content) == 0 {
					t.Errorf("%s/%s: empty content", name, f.Path)
				}
			}
		})
	}
}

func TestClaudeCodeOutputs(t *testing.T) {
	tr := &ClaudeCode{}
	cfg := &config.Config{Inject: true, InjectWhatIs: false}
	docs := []*parser.Document{
		{Type: parser.TypeRules, Name: "style", Body: "rules"},
	}
	plan, _ := tr.Plan(docs, cfg)
	paths := pathSet(plan)
	for _, want := range []string{"CLAUDE.md", ".claude/rules/style.md"} {
		if !paths[want] {
			t.Errorf("missing %s in plan: %v", want, paths)
		}
	}
}

func TestCodexAGENTSMd(t *testing.T) {
	tr := &Codex{}
	cfg := &config.Config{Inject: true, InjectWhatIs: false}
	docs := []*parser.Document{
		{Type: parser.TypeRules, Name: "rule-a", Description: "Rule A 描述", Body: "body of a"},
		{Type: parser.TypeRules, Name: "rule-b", Description: "Rule B 描述", Body: "body of b"},
	}
	plan, _ := tr.Plan(docs, cfg)
	// 现行行为：AGENTS.md（自描述清单，不内联 ## section）+ .codex/memories/<name>.md per nonRoot
	paths := pathSet(plan)
	if !paths["AGENTS.md"] {
		t.Error("missing AGENTS.md")
	}
	if !paths[".codex/memories/rule-a.md"] || !paths[".codex/memories/rule-b.md"] {
		t.Errorf("expected nonRoot rules spilled to .codex/memories/, got %v", paths)
	}
	main, _ := contentOf(plan, "AGENTS.md")
	for _, want := range []string{
		"Project AGENTS Manifest",
		"Reference Rules",
		".codex/memories/rule-a.md",
		"Rule A 描述",
		".codex/memories/rule-b.md",
	} {
		if !strings.Contains(main, want) {
			t.Errorf("missing %q in AGENTS.md:\n%s", want, main)
		}
	}
	// 不应再 ## section 内联 nonRoot rule body
	if strings.Contains(main, "body of a") {
		t.Error("nonRoot rule body should not be inlined in AGENTS.md")
	}
}

func TestCodexAGENTSMdAllSpilled(t *testing.T) {
	tr := &Codex{}
	cfg := &config.Config{Inject: false}
	docs := []*parser.Document{
		{Type: parser.TypeRules, Name: "aaa-tiny", Body: "tiny"},
		{Type: parser.TypeRules, Name: "zzz-huge", Body: strings.Repeat("x", 1000)},
	}
	plan, _ := tr.Plan(docs, cfg)
	paths := pathSet(plan)
	for _, want := range []string{"AGENTS.md", ".codex/memories/aaa-tiny.md", ".codex/memories/zzz-huge.md"} {
		if !paths[want] {
			t.Errorf("missing %s in plan: %v", want, paths)
		}
	}
	main, _ := contentOf(plan, "AGENTS.md")
	if !strings.Contains(main, "Reference Rules") {
		t.Error("expected Reference Rules manifest in AGENTS.md")
	}
}

func TestCodexCommandsAsSkill(t *testing.T) {
	tr := &Codex{}
	cfg := &config.Config{Inject: false}
	docs := []*parser.Document{
		{Type: parser.TypeCommands, Name: "review", Description: "Run code review", Body: "Steps..."},
	}
	plan, _ := tr.Plan(docs, cfg)
	c, ok := contentOf(plan, ".agents/skills/cmd-review/SKILL.md")
	if !ok {
		t.Fatalf("expected .agents/skills/cmd-review/SKILL.md, paths: %v", pathSet(plan))
	}
	for _, want := range []string{"name: cmd-review", "/review", "Run code review"} {
		if !strings.Contains(c, want) {
			t.Errorf("missing %q in:\n%s", want, c)
		}
	}
}

func TestCodexCommandsSkillDoesNotCollideWithSkills(t *testing.T) {
	tr := &Codex{}
	cfg := &config.Config{Inject: false}
	// 同 name 的 skill 与 command 同时存在，应输出到不同路径
	docs := []*parser.Document{
		{Type: parser.TypeSkills, Name: "review", Description: "Skill version", Body: "skill body"},
		{Type: parser.TypeCommands, Name: "review", Description: "Cmd version", Body: "cmd body"},
	}
	plan, _ := tr.Plan(docs, cfg)
	paths := pathSet(plan)
	if !paths[".agents/skills/review/SKILL.md"] {
		t.Errorf("missing skill path, paths: %v", paths)
	}
	if !paths[".agents/skills/cmd-review/SKILL.md"] {
		t.Errorf("missing command-as-skill path, paths: %v", paths)
	}
}

func TestCursorRuleModes(t *testing.T) {
	tr := &Cursor{}
	cfg := &config.Config{Inject: false}
	cases := []struct {
		name string
		doc  *parser.Document
		want string
	}{
		{"always", &parser.Document{Type: parser.TypeRules, Name: "x", AlwaysApply: true, Body: "b"}, "alwaysApply: true"},
		{"glob", &parser.Document{Type: parser.TypeRules, Name: "x", ApplyTo: []string{"**/*.go", "**/*.ts"}, Body: "b"}, "**/*.go,**/*.ts"},
		{"agent-req", &parser.Document{Type: parser.TypeRules, Name: "x", Description: "use when X", Body: "b"}, "description: use when X"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			plan, _ := tr.Plan([]*parser.Document{tc.doc}, cfg)
			if len(plan.Files) != 1 {
				t.Fatalf("expected 1 file, got %d", len(plan.Files))
			}
			c := string(plan.Files[0].Content)
			if !strings.Contains(c, tc.want) {
				t.Errorf("missing %q in:\n%s", tc.want, c)
			}
		})
	}
}

func TestWindsurfRuleTriggers(t *testing.T) {
	tr := &Windsurf{}
	cfg := &config.Config{Inject: false}
	cases := []struct {
		name string
		doc  *parser.Document
		want string
	}{
		{"always", &parser.Document{Type: parser.TypeRules, Name: "x", AlwaysApply: true, Body: "b"}, "trigger: always_on"},
		{"glob", &parser.Document{Type: parser.TypeRules, Name: "x", ApplyTo: []string{"**/*.go"}, Body: "b"}, "trigger: glob"},
		{"model-decision", &parser.Document{Type: parser.TypeRules, Name: "x", Description: "use", Body: "b"}, "trigger: model_decision"},
		{"manual", &parser.Document{Type: parser.TypeRules, Name: "x", Body: "b"}, "trigger: manual"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			plan, _ := tr.Plan([]*parser.Document{tc.doc}, cfg)
			if len(plan.Files) != 1 {
				t.Fatalf("expected 1 file")
			}
			c := string(plan.Files[0].Content)
			if !strings.Contains(c, tc.want) {
				t.Errorf("missing %q in:\n%s", tc.want, c)
			}
		})
	}
}

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

func TestGeminiCommandToml(t *testing.T) {
	tr := &Gemini{}
	cfg := &config.Config{Inject: true}
	docs := []*parser.Document{
		{Type: parser.TypeCommands, Name: "review", Description: "Run review", Body: "Please review the diff."},
	}
	plan, _ := tr.Plan(docs, cfg)
	c, ok := contentOf(plan, ".gemini/commands/review.toml")
	if !ok {
		t.Fatalf("missing .gemini/commands/review.toml")
	}
	if !strings.Contains(c, `description = "Run review"`) {
		t.Errorf("missing description: %s", c)
	}
	if !strings.Contains(c, "prompt = '''") {
		t.Errorf("missing prompt body: %s", c)
	}
}

func TestAiderNoop(t *testing.T) {
	tr := &Aider{}
	cfg := &config.Config{Inject: true}
	docs := []*parser.Document{{Type: parser.TypeRules, Name: "x", Body: "y"}}
	plan, _ := tr.Plan(docs, cfg)
	if len(plan.Files) != 0 {
		t.Errorf("aider should be noop, got %d files", len(plan.Files))
	}
}

func TestMCPDispatch(t *testing.T) {
	mcp := &config.MCPConfig{
		Version: "1.0",
		Servers: map[string]config.MCPServer{
			"github": {
				Type:    "stdio",
				Command: "gh",
				Args:    []string{"api"},
				Env:     map[string]string{"TOKEN": "${env:GH_TOKEN}"},
			},
		},
	}
	cfg := &config.Config{Inject: false, MCP: mcp}
	docs := []*parser.Document{
		{Type: parser.TypeRules, Name: "x", Body: "y"},
	}

	cases := map[string]struct {
		tr     Transformer
		path   string
		topKey string
	}{
		"claude-code": {tr: &ClaudeCode{}, path: ".mcp.json", topKey: `"mcpServers"`},
		"cursor":      {tr: &Cursor{}, path: ".cursor/mcp.json", topKey: `"mcpServers"`},
		"copilot":     {tr: &Copilot{}, path: ".vscode/mcp.json", topKey: `"servers"`},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			plan, _ := tc.tr.Plan(docs, cfg)
			c, ok := contentOf(plan, tc.path)
			if !ok {
				t.Fatalf("%s: missing %s in plan", name, tc.path)
			}
			if !strings.Contains(c, tc.topKey) {
				t.Errorf("%s: missing top-level key %s in:\n%s", name, tc.topKey, c)
			}
			if !strings.Contains(c, `"github"`) {
				t.Errorf("%s: missing github server in:\n%s", name, c)
			}
		})
	}
}

func TestMCPNotEmittedWhenAbsent(t *testing.T) {
	cfg := &config.Config{Inject: false, MCP: nil}
	docs := []*parser.Document{{Type: parser.TypeRules, Name: "x", Body: "y"}}
	for _, tr := range []Transformer{&ClaudeCode{}, &Cursor{}, &Copilot{}} {
		plan, _ := tr.Plan(docs, cfg)
		for _, f := range plan.Files {
			if strings.HasSuffix(f.Path, "mcp.json") {
				t.Errorf("%s emitted %s when MCP nil", tr.Name(), f.Path)
			}
		}
	}
}

func TestContinueDevOutputs(t *testing.T) {
	tr := &ContinueDev{}
	cfg := &config.Config{Inject: false}
	docs := []*parser.Document{
		{Type: parser.TypeRules, Name: "ts-style", ApplyTo: []string{"**/*.ts"}, Body: "use interfaces"},
		{Type: parser.TypeSkills, Name: "review", Description: "Code review", Body: "steps"},
		{Type: parser.TypeCommands, Name: "explain", Description: "Explain code", Body: "Please explain.", Version: "1.0"},
	}
	plan, _ := tr.Plan(docs, cfg)
	paths := pathSet(plan)
	for _, want := range []string{
		".continue/rules/ts-style.md",
		".continue/rules/skill-review.md",
		".continue/prompts/explain.prompt.md",
	} {
		if !paths[want] {
			t.Errorf("missing %s, paths: %v", want, paths)
		}
	}
	rule, _ := contentOf(plan, ".continue/rules/ts-style.md")
	if !strings.Contains(rule, "globs:") || !strings.Contains(rule, "**/*.ts") {
		t.Errorf("ts-style missing globs: %s", rule)
	}
	prompt, _ := contentOf(plan, ".continue/prompts/explain.prompt.md")
	if !strings.Contains(prompt, "invokable: true") {
		t.Errorf("prompt missing invokable: %s", prompt)
	}
}

func TestAntigravityRuleTriggers(t *testing.T) {
	tr := &Antigravity{}
	cfg := &config.Config{Inject: false}
	cases := []struct {
		name string
		doc  *parser.Document
		want string
	}{
		{"always", &parser.Document{Type: parser.TypeRules, Name: "x", AlwaysApply: true, Body: "b"}, "trigger: always_on"},
		{"glob", &parser.Document{Type: parser.TypeRules, Name: "x", ApplyTo: []string{"**/*.go"}, Body: "b"}, "trigger: glob"},
		{"model-decision", &parser.Document{Type: parser.TypeRules, Name: "x", Description: "use", Body: "b"}, "trigger: model_decision"},
		{"manual", &parser.Document{Type: parser.TypeRules, Name: "x", Body: "b"}, "trigger: manual"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			plan, _ := tr.Plan([]*parser.Document{tc.doc}, cfg)
			c := string(plan.Files[0].Content)
			if !strings.Contains(c, tc.want) {
				t.Errorf("missing %q in:\n%s", tc.want, c)
			}
			if plan.Files[0].Path != ".agents/rules/x.md" {
				t.Errorf("path = %s, want .agents/rules/x.md", plan.Files[0].Path)
			}
		})
	}
}

func TestAntigravityWorkflowAndSkill(t *testing.T) {
	tr := &Antigravity{}
	cfg := &config.Config{Inject: false}
	docs := []*parser.Document{
		{Type: parser.TypeSkills, Name: "review", Description: "Review", Body: "steps"},
		{Type: parser.TypeCommands, Name: "deploy", Description: "Deploy", Body: "go"},
	}
	plan, _ := tr.Plan(docs, cfg)
	paths := pathSet(plan)
	if !paths[".agents/rules/skill-review.md"] {
		t.Errorf("missing skill-as-rule, paths: %v", paths)
	}
	if !paths[".agents/workflows/deploy.md"] {
		t.Errorf("missing workflow, paths: %v", paths)
	}
}

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
	if !paths[".opencode/agents/review.md"] {
		t.Errorf("missing agent file, paths: %v", paths)
	}
	if !paths[".opencode/commands/deploy.md"] {
		t.Errorf("missing command file, paths: %v", paths)
	}
	for p := range paths {
		if strings.HasPrefix(p, "AGENTS") || strings.HasPrefix(p, ".opencode/rules") {
			t.Errorf("opencode should not write rules, got %s", p)
		}
	}
}

func pathSet(plan *writer.Plan) map[string]bool {
	out := make(map[string]bool, len(plan.Files))
	for _, f := range plan.Files {
		out[f.Path] = true
	}
	return out
}

func contentOf(plan *writer.Plan, path string) (string, bool) {
	for _, f := range plan.Files {
		if f.Path == path {
			return string(f.Content), true
		}
	}
	return "", false
}

func TestClaudeCodeSubagentOutput(t *testing.T) {
	tr := &ClaudeCode{}
	cfg := &config.Config{Inject: false}
	docs := []*parser.Document{
		{Type: parser.TypeSubagents, Name: "code-reviewer", Description: "Reviews code", Model: "claude-sonnet-4-5", AllowedTools: []string{"Read", "Grep"}, Body: "You are a code reviewer..."},
	}
	plan, _ := tr.Plan(docs, cfg)
	c, ok := contentOf(plan, ".claude/agents/code-reviewer.md")
	if !ok {
		t.Fatalf("expected .claude/agents/code-reviewer.md, paths: %v", pathSet(plan))
	}
	for _, want := range []string{"name: code-reviewer", "description: Reviews code", "model: claude-sonnet-4-5", "Read", "Grep", "You are a code reviewer"} {
		if !strings.Contains(c, want) {
			t.Errorf("missing %q in subagent file:\n%s", want, c)
		}
	}
}

func TestCodexAGENTSMdManifestWhenNoRoot(t *testing.T) {
	tr := &Codex{}
	cfg := &config.Config{Inject: false}
	docs := []*parser.Document{
		{Type: parser.TypeRules, Name: "naming", Description: "命名规范", Body: "naming body"},
	}
	plan, _ := tr.Plan(docs, cfg)
	main, _ := contentOf(plan, "AGENTS.md")
	if !strings.Contains(main, "Reference Rules") {
		t.Errorf("无 root rule 时应自动追加 Reference Rules:\n%s", main)
	}
}

func TestNestedRootOutputsToSubpath(t *testing.T) {
	tr := &ClaudeCode{}
	cfg := &config.Config{Inject: false}
	docs := []*parser.Document{
		{Type: parser.TypeRules, Name: "auth", Root: true, NestedPath: "igx-modules/auth", Body: "# Auth 模块\nbody"},
		{Type: parser.TypeRules, Name: "naming", Description: "命名", Body: "naming body"},
	}
	plan, _ := tr.Plan(docs, cfg)
	paths := pathSet(plan)
	if !paths["igx-modules/auth/CLAUDE.md"] {
		t.Errorf("expected igx-modules/auth/CLAUDE.md, paths: %v", paths)
	}
	if !paths["CLAUDE.md"] {
		t.Error("top-level CLAUDE.md should still exist")
	}
	// nested CLAUDE.md 不应含 manifest
	nested, _ := contentOf(plan, "igx-modules/auth/CLAUDE.md")
	if strings.Contains(nested, "Imported Rules") {
		t.Errorf("nested CLAUDE.md should not have manifest:\n%s", nested)
	}
	// 顶级 CLAUDE.md 应含 manifest
	top, _ := contentOf(plan, "CLAUDE.md")
	if !strings.Contains(top, "Imported Rules") {
		t.Error("top CLAUDE.md should have manifest section")
	}
	// nested rule 不应再 fan-out 到 .claude/rules/
	if paths[".claude/rules/auth.md"] {
		t.Error("nested root should not fan-out to .claude/rules/")
	}
}

func TestNestedAGENTSMd(t *testing.T) {
	tr := &Codex{}
	cfg := &config.Config{Inject: false}
	docs := []*parser.Document{
		{Type: parser.TypeRules, Name: "auth", Root: true, NestedPath: "src/auth", Body: "# Auth\nbody"},
		{Type: parser.TypeRules, Name: "x", Description: "x rule", Body: "x body"},
	}
	plan, _ := tr.Plan(docs, cfg)
	paths := pathSet(plan)
	if !paths["src/auth/AGENTS.md"] {
		t.Errorf("expected src/auth/AGENTS.md, paths: %v", paths)
	}
}
