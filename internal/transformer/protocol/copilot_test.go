package protocol

import (
	"encoding/json"
	"strings"
	"testing"

	"std-ai/internal/config"
	"std-ai/internal/parser"
	"std-ai/internal/writer"
)

// newCopilotAdapter 是 Phase 2.6 测试用 adapter（默认形态 A：SubagentInvokeCmd 空）。
// Phase 3 真正的 transformer copilotAdapter 会复用相同字段（届时 Phase 4.2 加
// SubagentInvokeCmd="claude --agent {name}"）。
func newCopilotAdapter() Adapter {
	return Adapter{
		Name:                 "copilot",
		RootFileName:         ".github/copilot-instructions.md",
		ManifestSection:      "Path-Specific Instructions",
		RulesDir:             ".github/instructions",
		CommandsDir:          ".github/prompts",
		SubagentsDir:         ".github/agents",
		FallbackDir:          ".github/instructions",
		InjectExplainer:      true,
		InjectStdaiTypeField: true,
		GlobsFieldName:       "applyTo",
		GlobsFieldFormat:     GlobsCommaString,
		RuleTriggerMode:      TriggerApplyTo,
		MCPPath:              ".vscode/mcp.json",
		MCPTopKey:            "servers",
		MaxBytesPerFile:      8000,
		SoftBytes:            4000,
	}
}

func pathsOf(plan *writer.Plan) map[string]string {
	out := make(map[string]string, len(plan.Files))
	for _, f := range plan.Files {
		out[f.Path] = string(f.Content)
	}
	return out
}

func TestCopilot_Disabled(t *testing.T) {
	a := newCopilotAdapter()
	a.Disabled = true
	plan, err := Copilot{}.Plan(nil, a, &config.Config{})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(plan.Files) != 0 {
		t.Errorf("Disabled adapter should yield 0 files, got %d", len(plan.Files))
	}
	if plan.Target != "copilot" {
		t.Errorf("Target = %q, want copilot", plan.Target)
	}
}

func TestCopilot_MissingName(t *testing.T) {
	_, err := Copilot{}.Plan(nil, Adapter{}, &config.Config{})
	if err == nil {
		t.Fatal("expected error on empty Name")
	}
}

func TestCopilot_RootFile_GlossaryAndManifest(t *testing.T) {
	a := newCopilotAdapter()
	a.InjectTypeGlossary = true
	docs := []*parser.Document{
		{Type: parser.TypeRules, Name: "root", Root: true, Body: "# Project rules\n\nbe careful."},
		{Type: parser.TypeRules, Name: "general", Body: "general guideline"},
		{Type: parser.TypeRules, Name: "go-only", ApplyTo: []string{"**/*.go"}, Description: "Go specific", Body: "go body"},
	}
	plan, err := Copilot{}.Plan(docs, a, &config.Config{Inject: false})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	files := pathsOf(plan)
	rootContent, ok := files[".github/copilot-instructions.md"]
	if !ok {
		t.Fatalf("missing root, paths=%v", keys(files))
	}
	if !strings.Contains(rootContent, "std-ai 类型速查") {
		t.Error("expected glossary section in root")
	}
	if !strings.Contains(rootContent, "be careful") {
		t.Error("expected root rule body")
	}
	if !strings.Contains(rootContent, "general guideline") {
		t.Error("expected non-root general rule in AGENTS-style body")
	}
	if !strings.Contains(rootContent, "## Path-Specific Instructions") {
		t.Error("expected manifest section title")
	}
	if !strings.Contains(rootContent, ".github/instructions/go-only.instructions.md") {
		t.Error("expected manifest to list path-specific rule")
	}
	// Root file should have IsRoot=true
	for _, f := range plan.Files {
		if f.Path == ".github/copilot-instructions.md" && !f.IsRoot {
			t.Error("root file should have IsRoot=true")
		}
	}
}

func TestCopilot_PathSpecificInstruction(t *testing.T) {
	a := newCopilotAdapter()
	docs := []*parser.Document{
		{Type: parser.TypeRules, Name: "ts-rules", ApplyTo: []string{"**/*.ts", "**/*.tsx"}, Description: "TypeScript", Body: "use strict"},
	}
	plan, err := Copilot{}.Plan(docs, a, &config.Config{})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	files := pathsOf(plan)
	body, ok := files[".github/instructions/ts-rules.instructions.md"]
	if !ok {
		t.Fatalf("missing ts-rules instruction, paths=%v", keys(files))
	}
	// applyTo comma string
	if !strings.Contains(body, `applyTo: "**/*.ts,**/*.tsx"`) {
		t.Errorf("expected comma-joined applyTo, got:\n%s", body)
	}
	// description should be merged into body
	if !strings.Contains(body, "TypeScript") {
		t.Error("expected description in body")
	}
	if !strings.Contains(body, "use strict") {
		t.Error("expected body content")
	}
}

func TestCopilot_PathSpecificInstruction_GlobsList(t *testing.T) {
	a := newCopilotAdapter()
	a.GlobsFieldFormat = GlobsList
	docs := []*parser.Document{
		{Type: parser.TypeRules, Name: "py-rules", ApplyTo: []string{"**/*.py"}, Body: "PEP8"},
	}
	plan, _ := Copilot{}.Plan(docs, a, &config.Config{})
	files := pathsOf(plan)
	body := files[".github/instructions/py-rules.instructions.md"]
	if !strings.Contains(body, "applyTo:\n  - \"**/*.py\"") {
		t.Errorf("expected YAML list applyTo, got:\n%s", body)
	}
}

func TestCopilot_Prompt(t *testing.T) {
	a := newCopilotAdapter()
	docs := []*parser.Document{
		{
			Type:         parser.TypeCommands,
			Name:         "release-patch",
			Description:  "Cut a patch release",
			ArgumentHint: "<version>",
			AllowedTools: []string{"Bash", "Edit"},
			Model:        "claude-sonnet-4-5",
			Body:         "## Steps\n1. bump\n",
		},
	}
	plan, _ := Copilot{}.Plan(docs, a, &config.Config{})
	files := pathsOf(plan)
	body, ok := files[".github/prompts/release-patch.prompt.md"]
	if !ok {
		t.Fatalf("missing prompt file, paths=%v", keys(files))
	}
	for _, want := range []string{
		"description: Cut a patch release",
		"argument-hint:",
		"tools:",
		"- Bash",
		"- Edit",
		"model:",
		"## Steps",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("missing %q in prompt body:\n%s", want, body)
		}
	}
}

func TestCopilot_Subagent_Native(t *testing.T) {
	a := newCopilotAdapter()
	// Phase 2.6 默认 SubagentInvokeCmd 空，走原生
	docs := []*parser.Document{
		{
			Type:        parser.TypeSubagents,
			Name:        "reviewer",
			Description: "Code reviewer",
			WhenToUse:   "Use after diff is ready",
			Body:        "You are a strict reviewer.",
		},
	}
	plan, _ := Copilot{}.Plan(docs, a, &config.Config{})
	files := pathsOf(plan)
	body, ok := files[".github/agents/reviewer.agent.md"]
	if !ok {
		t.Fatalf("missing agent file, paths=%v", keys(files))
	}
	if !strings.Contains(body, "description: Code reviewer Use after diff is ready") {
		t.Errorf("expected merged description, got:\n%s", body)
	}
	if !strings.Contains(body, "You are a strict reviewer.") {
		t.Error("expected body")
	}
	if strings.Contains(body, "How to spawn") {
		t.Error("native mode should not have shell invocation section")
	}
}

func TestCopilot_Subagent_CLIInvoke(t *testing.T) {
	a := newCopilotAdapter()
	a.SubagentInvokeCmd = "claude --agent {name}"
	docs := []*parser.Document{
		{
			Type:        parser.TypeSubagents,
			Name:        "reviewer",
			Description: "Code reviewer",
			Body:        "You are a strict reviewer.",
		},
	}
	plan, _ := Copilot{}.Plan(docs, a, &config.Config{})
	files := pathsOf(plan)
	body, ok := files[".github/agents/reviewer.agent.md"]
	if !ok {
		t.Fatalf("missing agent file, paths=%v", keys(files))
	}
	if !strings.Contains(body, "claude --agent reviewer") {
		t.Errorf("expected substituted CLI command, got:\n%s", body)
	}
	if !strings.Contains(body, "```bash") {
		t.Error("expected shell fenced block")
	}
	if !strings.Contains(body, "## How to spawn") {
		t.Error("expected spawn section")
	}
	if !strings.Contains(body, "std-ai-type: subagents") {
		t.Error("expected std-ai-type frontmatter under CLI invoke mode")
	}
}

func TestCopilot_SkillsFallback(t *testing.T) {
	a := newCopilotAdapter()
	docs := []*parser.Document{
		{
			Type:        parser.TypeSkills,
			Name:        "code-review",
			Description: "review code",
			Body:        "## Activation\nApply when reviewing.",
		},
	}
	plan, _ := Copilot{}.Plan(docs, a, &config.Config{})
	files := pathsOf(plan)
	body, ok := files[".github/instructions/skills/code-review.instructions.md"]
	if !ok {
		t.Fatalf("expected skills fallback path, paths=%v", keys(files))
	}
	// Path must not contain any private prefix
	for _, bad := range []string{"_skill_", "_ref_", "_command_", "_subagent_"} {
		for p := range files {
			if strings.Contains(p, bad) {
				t.Errorf("path %q must not contain forbidden prefix %q", p, bad)
			}
		}
	}
	if !strings.Contains(body, "std-ai-type: skills") {
		t.Error("expected std-ai-type frontmatter")
	}
	if !strings.Contains(body, `applyTo: ""`) {
		t.Errorf("expected empty applyTo frontmatter, got:\n%s", body)
	}
	if !strings.Contains(body, "<!-- std-ai degraded skills: code-review -->") {
		t.Error("expected explainer comment")
	}
	if !strings.Contains(body, "Skill is an on-demand capability pack") {
		t.Error("expected skills semantics text")
	}
}

func TestCopilot_ReferencesFallback(t *testing.T) {
	a := newCopilotAdapter()
	docs := []*parser.Document{
		{
			Type:        parser.TypeReferences,
			Name:        "transformer-design",
			Description: "transformer architecture",
			Body:        "Architecture notes",
		},
	}
	plan, _ := Copilot{}.Plan(docs, a, &config.Config{})
	files := pathsOf(plan)
	body, ok := files[".github/instructions/references/transformer-design.instructions.md"]
	if !ok {
		t.Fatalf("expected references fallback path, paths=%v", keys(files))
	}
	if !strings.Contains(body, "std-ai-type: references") {
		t.Error("expected std-ai-type frontmatter")
	}
	if !strings.Contains(body, "Reference is background material") {
		t.Error("expected references semantics text")
	}
}

func TestCopilot_MCP_TopKeyServers(t *testing.T) {
	a := newCopilotAdapter()
	cfg := &config.Config{
		MCP: &config.MCPConfig{
			Servers: map[string]config.MCPServer{
				"fs": {Command: "npx", Args: []string{"@modelcontextprotocol/server-filesystem"}},
			},
		},
	}
	plan, _ := Copilot{}.Plan(nil, a, cfg)
	files := pathsOf(plan)
	raw, ok := files[".vscode/mcp.json"]
	if !ok {
		t.Fatalf("missing mcp.json, paths=%v", keys(files))
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, raw)
	}
	if _, has := parsed["servers"]; !has {
		t.Errorf("expected top-level 'servers' key (≠ Claude's mcpServers), got: %v", parsed)
	}
	if _, has := parsed["mcpServers"]; has {
		t.Error("must not use Claude's 'mcpServers' top-level key")
	}
}

func TestCopilot_MCP_AbsentWhenNoServers(t *testing.T) {
	a := newCopilotAdapter()
	plan, _ := Copilot{}.Plan(nil, a, &config.Config{MCP: &config.MCPConfig{}})
	files := pathsOf(plan)
	if _, ok := files[".vscode/mcp.json"]; ok {
		t.Error("mcp.json should not be produced with empty servers")
	}
}

func TestCopilot_RootBytesOverSoft_WarnByRunner(t *testing.T) {
	// 验证：超出 SoftBytes 时 protocol 仍正常产 FileOp，不做 WARN（WARN 由 runner / budget 检查）。
	a := newCopilotAdapter()
	big := strings.Repeat("x", 5000)
	docs := []*parser.Document{
		{Type: parser.TypeRules, Name: "huge", Root: true, Body: big},
	}
	plan, err := Copilot{}.Plan(docs, a, &config.Config{})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	files := pathsOf(plan)
	body, ok := files[".github/copilot-instructions.md"]
	if !ok {
		t.Fatal("expected root file even when over SoftBytes")
	}
	if len(body) < 5000 {
		t.Errorf("body should contain all content, got %d bytes", len(body))
	}
}

func TestCopilot_EmptyDocs_NoMCP(t *testing.T) {
	a := newCopilotAdapter()
	plan, err := Copilot{}.Plan(nil, a, &config.Config{})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(plan.Files) != 0 {
		t.Errorf("expected 0 files for empty input, got %d", len(plan.Files))
	}
}

func TestCopilot_RootFile_OnlyPathSpecific_ManifestStillEmitted(t *testing.T) {
	// 当只有 path-specific rule 时，根文件应至少包含 manifest 段
	a := newCopilotAdapter()
	docs := []*parser.Document{
		{Type: parser.TypeRules, Name: "go-only", ApplyTo: []string{"**/*.go"}, Body: "go"},
	}
	plan, _ := Copilot{}.Plan(docs, a, &config.Config{})
	files := pathsOf(plan)
	root, ok := files[".github/copilot-instructions.md"]
	if !ok {
		t.Fatal("expected root file referencing manifest")
	}
	if !strings.Contains(root, "go-only.instructions.md") {
		t.Errorf("expected manifest to list path-specific rule, got:\n%s", root)
	}
}

func TestCopilot_SortDeterministic(t *testing.T) {
	a := newCopilotAdapter()
	docs := []*parser.Document{
		{Type: parser.TypeRules, Name: "zebra", ApplyTo: []string{"**/*.z"}, Body: "z"},
		{Type: parser.TypeRules, Name: "alpha", ApplyTo: []string{"**/*.a"}, Body: "a"},
	}
	plan, _ := Copilot{}.Plan(docs, a, &config.Config{})
	var order []string
	for _, f := range plan.Files {
		if strings.HasPrefix(f.Path, ".github/instructions/") && strings.HasSuffix(f.Path, ".instructions.md") {
			order = append(order, f.Path)
		}
	}
	if len(order) < 2 {
		t.Fatalf("expected 2 instruction files, got %v", order)
	}
	if order[0] >= order[1] {
		t.Errorf("expected sorted order alpha < zebra, got %v", order)
	}
}

func keys(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
