package transformer

import (
	"strings"
	"testing"

	"std-ai/internal/config"
	"std-ai/internal/parser"
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
