package transformerutil

import (
	"strings"
	"testing"

	"std-ai/internal/parser"
	"std-ai/internal/writer"
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

func TestBuildMarkdownFileFrontmatterFirstLine(t *testing.T) {
	opts := writer.FooterOptions{Inject: true, Version: "test", SourcePath: "skills/x/SKILL.md"}
	fm := "---\nname: x\ndescription: y\n---\n"
	op := BuildMarkdownFile(".claude/skills/x/SKILL.md", fm, "body text", opts)
	out := string(op.Content)

	// frontmatter 必须占首行，否则不是合法 YAML frontmatter
	if !strings.HasPrefix(out, "---\nname: x\n") {
		t.Fatalf("frontmatter must start at line 1, got:\n%s", out)
	}
	// header marker 必须出现在 frontmatter 闭合之后、body 之前
	markerIdx := strings.Index(out, writer.MarkerStart)
	if markerIdx < 0 {
		t.Fatal("header marker missing")
	}
	fmCloseIdx := strings.Index(out, "\n---\n") // 闭合的 ---
	bodyIdx := strings.Index(out, "body text")
	if fmCloseIdx >= markerIdx || markerIdx >= bodyIdx {
		t.Errorf("marker must sit between frontmatter close and body; fmClose=%d marker=%d body=%d\n%s",
			fmCloseIdx, markerIdx, bodyIdx, out)
	}
}

func TestBuildMarkdownFileNoFrontmatterMarkerFirst(t *testing.T) {
	opts := writer.FooterOptions{Inject: true, Version: "test", SourcePath: "rules/x.md"}
	op := BuildMarkdownFile(".claude/rules/x.md", "", "body text", opts)
	out := string(op.Content)
	// 无 frontmatter 时 header marker 仍在文件最前
	if !strings.HasPrefix(out, writer.MarkerStart) {
		t.Errorf("marker should be first when no frontmatter, got:\n%s", out)
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
