package transformer

import (
	"strings"
	"testing"

	"std-ai/internal/config"
	"std-ai/internal/parser"
)

func makeDoc(t parser.DocType, name, prio string, targets, exclude []string) *parser.Document {
	return &parser.Document{
		Type:           t,
		Name:           name,
		Priority:       parser.Priority(prio),
		Targets:        targets,
		ExcludeTargets: exclude,
		Body:           "body of " + name,
	}
}

func TestFilterDocsTargetWhitelist(t *testing.T) {
	docs := []*parser.Document{
		makeDoc(parser.TypeRules, "a", "", []string{"claude-code"}, nil),
		makeDoc(parser.TypeRules, "b", "", []string{"codex"}, nil),
		makeDoc(parser.TypeRules, "c", "", nil, nil),
	}
	out := FilterDocs(docs, "claude-code")
	names := docNames(out)
	want := []string{"a", "c"}
	if !equalSlices(names, want) {
		t.Errorf("names = %v, want %v", names, want)
	}
}

func TestFilterDocsExcludeTargets(t *testing.T) {
	docs := []*parser.Document{
		makeDoc(parser.TypeRules, "a", "", nil, []string{"codex"}),
		makeDoc(parser.TypeRules, "b", "", nil, []string{"claude-code"}),
	}
	out := FilterDocs(docs, "claude-code")
	if len(out) != 1 || out[0].Name != "a" {
		t.Errorf("filter exclude = %v", docNames(out))
	}
}

func TestSortDocsByPriorityThenName(t *testing.T) {
	docs := []*parser.Document{
		makeDoc(parser.TypeRules, "z", "low", nil, nil),
		makeDoc(parser.TypeRules, "a", "high", nil, nil),
		makeDoc(parser.TypeRules, "m", "", nil, nil),
		makeDoc(parser.TypeRules, "b", "high", nil, nil),
	}
	SortDocs(docs)
	got := docNames(docs)
	want := []string{"a", "b", "m", "z"}
	if !equalSlices(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestFilterByType(t *testing.T) {
	docs := []*parser.Document{
		makeDoc(parser.TypeRules, "r1", "", nil, nil),
		makeDoc(parser.TypeSkills, "s1", "", nil, nil),
		makeDoc(parser.TypeCommands, "c1", "", nil, nil),
	}
	rules := FilterByType(docs, parser.TypeRules)
	if len(rules) != 1 || rules[0].Name != "r1" {
		t.Errorf("filter rules = %v", docNames(rules))
	}
}

func TestFmBuilderEmpty(t *testing.T) {
	var fm FmBuilder
	if fm.String() != "" {
		t.Errorf("empty fm should return empty: %q", fm.String())
	}
}

func TestFmBuilderFields(t *testing.T) {
	var fm FmBuilder
	fm.Add("name", "coding-style")
	fm.Add("description", "general style")
	fm.AddList("applyTo", []string{"**/*.go", "**/*.ts"})
	fm.AddBool("alwaysApply", true)
	out := fm.String()
	if !strings.HasPrefix(out, "---\n") || !strings.HasSuffix(out, "---\n") {
		t.Errorf("malformed fm: %q", out)
	}
	for _, want := range []string{
		"name: coding-style", "description: general style",
		"applyTo:", `  - "**/*.go"`, `  - "**/*.ts"`,
		"alwaysApply: true",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
}

func TestYAMLScalarQuoting(t *testing.T) {
	if YAMLScalar("simple") != "simple" {
		t.Error("simple should not quote")
	}
	q := YAMLScalar("has: colon")
	if !strings.HasPrefix(q, `"`) || !strings.HasSuffix(q, `"`) {
		t.Errorf("colon should quote: %q", q)
	}
	if !strings.Contains(YAMLScalar(`with "quote"`), `\"`) {
		t.Error("quote should escape")
	}
	for _, s := range []string{"**/*.go", "*Service.java", "[draft]", "{a,b}", "a&b", "!important"} {
		q := YAMLScalar(s)
		if !strings.HasPrefix(q, `"`) || !strings.HasSuffix(q, `"`) {
			t.Errorf("YAML-special %q should be quoted, got %q", s, q)
		}
	}
}

func TestJoinAGENTSStyle(t *testing.T) {
	docs := []*parser.Document{
		{Name: "rule-a", Body: "content of a", Description: "desc a"},
		{Name: "rule-b", Body: "content of b"},
	}
	out := JoinAGENTSStyle("Project AGENTS Manifest", docs)
	for _, want := range []string{
		"# Project AGENTS Manifest", "## rule-a", "desc a", "content of a",
		"## rule-b", "content of b",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
}

func TestWindsurfRuleOverLimitWarn(t *testing.T) {
	tr := &Windsurf{}
	cfg := &config.Config{Inject: false}
	bigBody := strings.Repeat("a", 13000)
	docs := []*parser.Document{
		{Type: parser.TypeRules, Name: "huge", AlwaysApply: true, Body: bigBody},
	}
	plan, _ := tr.Plan(docs, cfg)
	found := false
	for _, f := range plan.Files {
		if strings.HasPrefix(f.Reason, "WARN") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected WARN for over-limit windsurf rule")
	}
}

func TestAntigravityRuleOverLimitWarn(t *testing.T) {
	tr := &Antigravity{}
	cfg := &config.Config{Inject: false}
	bigBody := strings.Repeat("b", 13000)
	docs := []*parser.Document{
		{Type: parser.TypeRules, Name: "huge", AlwaysApply: true, Body: bigBody},
	}
	plan, _ := tr.Plan(docs, cfg)
	found := false
	for _, f := range plan.Files {
		if strings.HasPrefix(f.Reason, "WARN") {
			found = true
		}
	}
	if !found {
		t.Error("expected WARN for over-limit antigravity rule")
	}
}

func TestCodexAGENTSMdMultipleSpills(t *testing.T) {
	tr := &Codex{}
	cfg := &config.Config{Inject: false}
	bigBody := strings.Repeat("c ", 18000)
	docs := []*parser.Document{
		{Type: parser.TypeRules, Name: "aaa-tiny", Body: "tiny"},
		{Type: parser.TypeRules, Name: "bbb-huge", Body: bigBody},
		{Type: parser.TypeRules, Name: "ccc-huger", Body: bigBody},
	}
	plan, _ := tr.Plan(docs, cfg)
	paths := pathSet(plan)
	if !paths["AGENTS.md"] {
		t.Error("AGENTS.md missing")
	}
	spillCount := 0
	for p := range paths {
		if strings.HasPrefix(p, ".codex/memories/") {
			spillCount++
		}
	}
	if spillCount < 2 {
		t.Errorf("expected at least 2 spills, got %d, paths: %v", spillCount, paths)
	}
}

func TestCursorRuleVeryLongBodyDoesNotPanic(t *testing.T) {
	tr := &Cursor{}
	cfg := &config.Config{Inject: false}
	docs := []*parser.Document{
		{
			Type: parser.TypeRules, Name: "long", AlwaysApply: true,
			Body: strings.Repeat("x ", 50000),
		},
	}
	plan, err := tr.Plan(docs, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Files) != 1 {
		t.Errorf("expected 1 file, got %d", len(plan.Files))
	}
}

func TestFilterDocsAllExcluded(t *testing.T) {
	docs := []*parser.Document{
		{Type: parser.TypeRules, Name: "a", ExcludeTargets: []string{"claude-code", "codex"}},
		{Type: parser.TypeRules, Name: "b", Targets: []string{"cursor"}},
	}
	if got := FilterDocs(docs, "claude-code"); len(got) != 0 {
		t.Errorf("claude-code: got %d, want 0", len(got))
	}
	// cursor: doc a 不 exclude cursor（仅 exclude claude-code/codex），doc b targets cursor
	if got := FilterDocs(docs, "cursor"); len(got) != 2 {
		t.Errorf("cursor: got %d, want 2", len(got))
	}
	// codex: doc a exclude codex；doc b targets 仅 cursor 不含 codex -> 0
	if got := FilterDocs(docs, "codex"); len(got) != 0 {
		t.Errorf("codex: got %d, want 0", len(got))
	}
}

func TestRegistry(t *testing.T) {
	if _, ok := Get("claude-code"); !ok {
		t.Error("claude-code transformer not registered")
	}
	if _, ok := Get("codex"); !ok {
		t.Error("codex transformer not registered")
	}
	if _, ok := Get("nonexistent"); ok {
		t.Error("nonexistent should not be registered")
	}
	if len(Names()) < 2 {
		t.Errorf("Names() = %v, want at least 2", Names())
	}
}

func docNames(docs []*parser.Document) []string {
	out := make([]string, len(docs))
	for i, d := range docs {
		out[i] = d.Name
	}
	return out
}

func equalSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestEffectiveApplyToFallsBackToGlobal(t *testing.T) {
	d := &parser.Document{ApplyTo: []string{"**/*.go"}}
	got := EffectiveApplyTo(d, "claude-code")
	if len(got) != 1 || got[0] != "**/*.go" {
		t.Errorf("expected global ApplyTo fallback, got %v", got)
	}
}

func TestEffectiveApplyToUsesTargetSpecific(t *testing.T) {
	d := &parser.Document{
		ApplyTo: []string{"**/*.go"},
		TargetPaths: map[string][]string{
			"claudecode": {"**/*Service.java"},
		},
	}
	got := EffectiveApplyTo(d, "claude-code")
	if len(got) != 1 || got[0] != "**/*Service.java" {
		t.Errorf("expected claudecode override, got %v", got)
	}
	// 其他 target 落到全局
	got2 := EffectiveApplyTo(d, "cursor")
	if len(got2) != 1 || got2[0] != "**/*.go" {
		t.Errorf("expected global fallback for cursor, got %v", got2)
	}
}

func TestEffectiveApplyToNilDoc(t *testing.T) {
	if got := EffectiveApplyTo(nil, "claude-code"); got != nil {
		t.Errorf("nil doc should return nil, got %v", got)
	}
}
