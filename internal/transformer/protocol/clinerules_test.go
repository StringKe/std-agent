package protocol

import (
	"fmt"
	"strings"
	"testing"

	"github.com/StringKe/std-agent/internal/parser"
)

// priorityToPrefix 复现 cline.go 的 priority -> 数字前缀映射，仅供测试使用
func priorityToPrefix(d *parser.Document) string {
	prefix := 500
	switch d.Priority {
	case parser.PriorityHigh:
		prefix = 100
	case parser.PriorityLow:
		prefix = 900
	}
	return fmt.Sprintf("%03d-", prefix)
}

// clineLikeAdapter 是 cline 风格的 adapter（带 RulePrefix）
func clineLikeAdapter() Adapter {
	return Adapter{
		Name:                 "cline",
		RulesDir:             ".clinerules",
		FallbackDir:          ".clinerules",
		GlobsFieldName:       "paths",
		GlobsFieldFormat:     GlobsList,
		RulePrefix:           priorityToPrefix,
		InjectExplainer:      true,
		InjectStdaiTypeField: true,
	}
}

// rooLikeAdapter 是 roo-code 风格的 adapter（无 RulePrefix，子目录隔离）
func rooLikeAdapter() Adapter {
	return Adapter{
		Name:                 "roo-code",
		RulesDir:             ".roo/rules",
		FallbackDir:          ".roo/rules",
		GlobsFieldName:       "paths",
		GlobsFieldFormat:     GlobsList,
		RulePrefix:           nil,
		InjectExplainer:      true,
		InjectStdaiTypeField: true,
	}
}

func TestClinerules_RuleWithPriorityPrefix(t *testing.T) {
	docs := []*parser.Document{
		{
			Type:     parser.TypeRules,
			Name:     "style",
			Priority: parser.PriorityHigh,
			ApplyTo:  []string{"**/*.go"},
			Body:     "BODY",
			Path:     ".stdai/standards/rules/style.md",
		},
	}
	plan, err := Clinerules{}.Plan(docs, clineLikeAdapter(), newTestCfg())
	if err != nil {
		t.Fatalf("Plan() error: %v", err)
	}
	if len(plan.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(plan.Files))
	}
	op := plan.Files[0]
	if op.Path != ".clinerules/100-style.md" {
		t.Errorf("path = %q, want .clinerules/100-style.md", op.Path)
	}
	got := string(op.Content)
	if !strings.Contains(got, "paths:\n  - \"**/*.go\"") {
		t.Errorf("expected paths frontmatter, got:\n%s", got)
	}
	if !strings.Contains(got, "BODY") {
		t.Errorf("expected body, got:\n%s", got)
	}
}

func TestClinerules_RulePriorityVariants(t *testing.T) {
	cases := []struct {
		name     string
		priority parser.Priority
		want     string
	}{
		{"high", parser.PriorityHigh, ".clinerules/100-a.md"},
		{"normal", parser.PriorityNormal, ".clinerules/500-a.md"},
		{"low", parser.PriorityLow, ".clinerules/900-a.md"},
		{"empty defaults to normal", "", ".clinerules/500-a.md"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			docs := []*parser.Document{
				{Type: parser.TypeRules, Name: "a", Priority: tc.priority, Body: "B"},
			}
			plan, err := Clinerules{}.Plan(docs, clineLikeAdapter(), newTestCfg())
			if err != nil {
				t.Fatalf("Plan() error: %v", err)
			}
			if plan.Files[0].Path != tc.want {
				t.Errorf("path = %q, want %q", plan.Files[0].Path, tc.want)
			}
		})
	}
}

func TestClinerules_RuleNoPrefixWhenAdapterNil(t *testing.T) {
	docs := []*parser.Document{
		{Type: parser.TypeRules, Name: "style", Priority: parser.PriorityHigh, Body: "B"},
	}
	plan, err := Clinerules{}.Plan(docs, rooLikeAdapter(), newTestCfg())
	if err != nil {
		t.Fatalf("Plan() error: %v", err)
	}
	if plan.Files[0].Path != ".roo/rules/style.md" {
		t.Errorf("path = %q, want .roo/rules/style.md (no numeric prefix)", plan.Files[0].Path)
	}
}

func TestClinerules_CommandsAsWorkflows(t *testing.T) {
	docs := []*parser.Document{
		{
			Type:        parser.TypeCommands,
			Name:        "release-patch",
			Description: "release a patch",
			Body:        "STEP 1",
		},
	}
	plan, err := Clinerules{}.Plan(docs, clineLikeAdapter(), newTestCfg())
	if err != nil {
		t.Fatalf("Plan() error: %v", err)
	}
	if len(plan.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(plan.Files))
	}
	op := plan.Files[0]
	if op.Path != ".clinerules/workflows/release-patch.md" {
		t.Errorf("path = %q, want .clinerules/workflows/release-patch.md", op.Path)
	}
	got := string(op.Content)
	if !strings.Contains(got, "release a patch") {
		t.Errorf("expected description prepended, got:\n%s", got)
	}
	if !strings.Contains(got, "STEP 1") {
		t.Errorf("expected body, got:\n%s", got)
	}
}

func TestClinerules_SkillsFallbackToAgentSkillsPackage(t *testing.T) {
	docs := []*parser.Document{
		{
			Type:        parser.TypeSkills,
			Name:        "code-review",
			Description: "review code",
			Body:        "skill body",
		},
	}
	plan, err := Clinerules{}.Plan(docs, clineLikeAdapter(), newTestCfg())
	if err != nil {
		t.Fatalf("Plan() error: %v", err)
	}
	if len(plan.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(plan.Files))
	}
	op := plan.Files[0]
	want := ".clinerules/skills/code-review/SKILL.md"
	if op.Path != want {
		t.Errorf("path = %q, want %q", op.Path, want)
	}
	got := string(op.Content)
	if !strings.Contains(got, "name: code-review") {
		t.Errorf("expected name frontmatter, got:\n%s", got)
	}
	if !strings.Contains(got, "std-agent-type: skills") {
		t.Errorf("expected std-agent-type frontmatter, got:\n%s", got)
	}
	if !strings.Contains(got, "<!-- std-agent degraded skills: code-review") {
		t.Errorf("expected explainer header, got:\n%s", got)
	}
}

func TestClinerules_ReferencesFallbackToSubdir(t *testing.T) {
	docs := []*parser.Document{
		{
			Type:        parser.TypeReferences,
			Name:        "transformer-design",
			Description: "design notes",
			Body:        "ref body",
		},
	}
	plan, err := Clinerules{}.Plan(docs, clineLikeAdapter(), newTestCfg())
	if err != nil {
		t.Fatalf("Plan() error: %v", err)
	}
	if len(plan.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(plan.Files))
	}
	op := plan.Files[0]
	want := ".clinerules/references/transformer-design.md"
	if op.Path != want {
		t.Errorf("path = %q, want %q", op.Path, want)
	}
	for _, bad := range []string{"_ref_", "_skill_", "_command_", "_subagent_", "900-"} {
		if strings.Contains(op.Path, bad) {
			t.Errorf("path %q should not contain forbidden token %q", op.Path, bad)
		}
	}
	got := string(op.Content)
	if !strings.Contains(got, "std-agent-type: references") {
		t.Errorf("expected std-agent-type frontmatter, got:\n%s", got)
	}
	if !strings.Contains(got, "<!-- std-agent degraded references: transformer-design") {
		t.Errorf("expected explainer header, got:\n%s", got)
	}
}

func TestClinerules_SubagentsFallbackToSubdir(t *testing.T) {
	docs := []*parser.Document{
		{
			Type:        parser.TypeSubagents,
			Name:        "reviewer",
			Description: "code reviewer",
			Body:        "you are a reviewer",
		},
	}
	plan, err := Clinerules{}.Plan(docs, clineLikeAdapter(), newTestCfg())
	if err != nil {
		t.Fatalf("Plan() error: %v", err)
	}
	if len(plan.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(plan.Files))
	}
	op := plan.Files[0]
	want := ".clinerules/subagents/reviewer.md"
	if op.Path != want {
		t.Errorf("path = %q, want %q", op.Path, want)
	}
	got := string(op.Content)
	if !strings.Contains(got, "std-agent-type: subagents") {
		t.Errorf("expected std-agent-type frontmatter, got:\n%s", got)
	}
}

func TestClinerules_GlossaryWhenEnabled(t *testing.T) {
	adapter := clineLikeAdapter()
	adapter.InjectTypeGlossary = true
	docs := []*parser.Document{
		{Type: parser.TypeRules, Name: "a", Priority: parser.PriorityHigh, Body: "B"},
	}
	cfg := newTestCfg()
	cfg.InjectTypeGlossary = true
	plan, err := Clinerules{}.Plan(docs, adapter, cfg)
	if err != nil {
		t.Fatalf("Plan() error: %v", err)
	}
	var glossaryOp *struct {
		path    string
		content string
	}
	for _, f := range plan.Files {
		if f.Path == ".clinerules/glossary.md" {
			glossaryOp = &struct {
				path    string
				content string
			}{f.Path, string(f.Content)}
		}
	}
	if glossaryOp == nil {
		paths := make([]string, 0, len(plan.Files))
		for _, f := range plan.Files {
			paths = append(paths, f.Path)
		}
		t.Fatalf("expected .clinerules/glossary.md, files: %v", paths)
	}
	if !strings.Contains(glossaryOp.content, "std-agent-type: glossary") {
		t.Errorf("expected std-agent-type frontmatter, got:\n%s", glossaryOp.content)
	}
	if !strings.Contains(glossaryOp.content, "std-agent 类型速查") {
		t.Errorf("expected glossary body title, got:\n%s", glossaryOp.content)
	}
}

func TestClinerules_GlossaryDisabledByConfig(t *testing.T) {
	t.Parallel()

	adapter := clineLikeAdapter()
	adapter.InjectTypeGlossary = true
	plan, err := Clinerules{}.Plan(nil, adapter, newTestCfg())
	if err != nil {
		t.Fatal(err)
	}
	for _, op := range plan.Files {
		if strings.Contains(op.Path, "glossary") {
			t.Fatalf("config-disabled glossary should not be emitted: %s", op.Path)
		}
	}
}

func TestClinerules_GlossarySkippedWhenDisabled(t *testing.T) {
	adapter := clineLikeAdapter()
	adapter.InjectTypeGlossary = false
	docs := []*parser.Document{
		{Type: parser.TypeRules, Name: "a", Body: "B"},
	}
	cfg := newTestCfg()
	cfg.InjectTypeGlossary = true
	plan, err := Clinerules{}.Plan(docs, adapter, cfg)
	if err != nil {
		t.Fatalf("Plan() error: %v", err)
	}
	for _, f := range plan.Files {
		if f.Path == ".clinerules/glossary.md" {
			t.Errorf("glossary should be absent when InjectTypeGlossary=false, got %s", f.Path)
		}
	}
}

func TestClinerules_Disabled(t *testing.T) {
	adapter := clineLikeAdapter()
	adapter.Disabled = true
	docs := []*parser.Document{
		{Type: parser.TypeRules, Name: "a", Body: "B"},
	}
	cfg := newTestCfg()
	cfg.InjectTypeGlossary = true
	plan, err := Clinerules{}.Plan(docs, adapter, cfg)
	if err != nil {
		t.Fatalf("Plan() error: %v", err)
	}
	if len(plan.Files) != 0 {
		t.Errorf("Disabled adapter should produce 0 files, got %d", len(plan.Files))
	}
}

func TestClinerules_EmptyDocs(t *testing.T) {
	plan, err := Clinerules{}.Plan(nil, clineLikeAdapter(), newTestCfg())
	if err != nil {
		t.Fatalf("Plan() error: %v", err)
	}
	if plan.Target != "cline" {
		t.Errorf("Target = %q, want cline", plan.Target)
	}
	if len(plan.Files) != 0 {
		t.Errorf("empty docs should produce 0 files (no glossary either since InjectTypeGlossary defaults false), got %d", len(plan.Files))
	}
}

func TestClinerules_SortingByPriorityThenName(t *testing.T) {
	docs := []*parser.Document{
		{Type: parser.TypeRules, Name: "zeta", Priority: parser.PriorityNormal, Body: "B"},
		{Type: parser.TypeRules, Name: "alpha", Priority: parser.PriorityHigh, Body: "B"},
		{Type: parser.TypeRules, Name: "beta", Priority: parser.PriorityHigh, Body: "B"},
	}
	plan, err := Clinerules{}.Plan(docs, clineLikeAdapter(), newTestCfg())
	if err != nil {
		t.Fatalf("Plan() error: %v", err)
	}
	wantOrder := []string{
		".clinerules/100-alpha.md",
		".clinerules/100-beta.md",
		".clinerules/500-zeta.md",
	}
	if len(plan.Files) != len(wantOrder) {
		t.Fatalf("expected %d files, got %d", len(wantOrder), len(plan.Files))
	}
	for i, w := range wantOrder {
		if plan.Files[i].Path != w {
			t.Errorf("file[%d].Path = %q, want %q", i, plan.Files[i].Path, w)
		}
	}
}

func TestClinerules_CombinedFanoutContainsAllTypes(t *testing.T) {
	adapter := clineLikeAdapter()
	adapter.InjectTypeGlossary = true
	docs := []*parser.Document{
		{Type: parser.TypeRules, Name: "r1", Priority: parser.PriorityHigh, Body: "RULE"},
		{Type: parser.TypeCommands, Name: "c1", Body: "CMD"},
		{Type: parser.TypeSkills, Name: "s1", Description: "skill", Body: "SK"},
		{Type: parser.TypeReferences, Name: "ref1", Description: "ref", Body: "REF"},
		{Type: parser.TypeSubagents, Name: "sub1", Description: "sub", Body: "SUB"},
	}
	cfg := newTestCfg()
	cfg.InjectTypeGlossary = true
	plan, err := Clinerules{}.Plan(docs, adapter, cfg)
	if err != nil {
		t.Fatalf("Plan() error: %v", err)
	}
	wantPaths := map[string]bool{
		".clinerules/glossary.md":        false,
		".clinerules/100-r1.md":          false,
		".clinerules/workflows/c1.md":    false,
		".clinerules/skills/s1/SKILL.md": false,
		".clinerules/references/ref1.md": false,
		".clinerules/subagents/sub1.md":  false,
	}
	for _, f := range plan.Files {
		if _, ok := wantPaths[f.Path]; ok {
			wantPaths[f.Path] = true
		}
	}
	for p, found := range wantPaths {
		if !found {
			t.Errorf("expected %s in plan", p)
		}
	}
}

func TestClinerules_RulePrefixCalled(t *testing.T) {
	called := false
	customPrefix := func(_ *parser.Document) string {
		called = true
		return "XX-"
	}
	adapter := clineLikeAdapter()
	adapter.RulePrefix = customPrefix
	docs := []*parser.Document{
		{Type: parser.TypeRules, Name: "a", Body: "B"},
	}
	plan, err := Clinerules{}.Plan(docs, adapter, newTestCfg())
	if err != nil {
		t.Fatalf("Plan() error: %v", err)
	}
	if !called {
		t.Error("RulePrefix func should be called for each rule")
	}
	if plan.Files[0].Path != ".clinerules/XX-a.md" {
		t.Errorf("path = %q, want .clinerules/XX-a.md", plan.Files[0].Path)
	}
}
