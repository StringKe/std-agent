package transformer

import (
	"testing"

	"github.com/StringKe/std-agent/internal/parser"
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
