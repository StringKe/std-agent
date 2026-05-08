package source

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadIgnoreMissingFileReturnsEmpty(t *testing.T) {
	ig, err := LoadIgnoreFile(filepath.Join(t.TempDir(), "missing"))
	if err != nil {
		t.Fatal(err)
	}
	if ig.Match("anything") {
		t.Error("empty ignore should not match")
	}
}

func TestParseIgnoreSkipsBlankAndComments(t *testing.T) {
	src := `# comment line
rules/draft-*.md

# another comment
skills/wip/**

`
	ig, err := parseIgnore(strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}
	got := ig.Patterns()
	want := []string{"rules/draft-*.md", "skills/wip/**"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i, p := range want {
		if got[i] != p {
			t.Errorf("pattern[%d] = %q, want %q", i, got[i], p)
		}
	}
}

func TestIgnoreMatchSimple(t *testing.T) {
	ig, _ := parseIgnore(strings.NewReader("rules/draft-*.md\nskills/wip/**\n"))
	cases := map[string]bool{
		"rules/draft-1.md":            true,
		"rules/draft-foo.md":          true,
		"rules/coding.md":             false,
		"skills/wip/foo/SKILL.md":     true,
		"skills/wip/scripts/x.sh":     true,
		"skills/code-review/SKILL.md": false,
		"references/note.md":          false,
	}
	for in, want := range cases {
		if got := ig.Match(in); got != want {
			t.Errorf("Match(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestIgnoreMatchDoublestarDeep(t *testing.T) {
	ig, _ := parseIgnore(strings.NewReader("**/*.tmp\n"))
	cases := map[string]bool{
		"foo.tmp":          true,
		"a/b/c/x.tmp":      true,
		"rules/foo.md":     false,
		"skills/wip/x.tmp": true,
	}
	for in, want := range cases {
		if got := ig.Match(in); got != want {
			t.Errorf("Match(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestLoadIgnoreFileFromDisk(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, ".stdaiignore")
	content := "# project ignore\nrules/private-*.md\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	ig, err := LoadIgnoreFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !ig.Match("rules/private-keys.md") {
		t.Error("should match rules/private-keys.md")
	}
	if ig.Match("rules/public.md") {
		t.Error("should not match rules/public.md")
	}
}

func TestNilIgnoreMatch(t *testing.T) {
	var ig *Ignore
	if ig.Match("anything") {
		t.Error("nil ignore should not match")
	}
	if ig.Patterns() != nil {
		t.Error("nil ignore Patterns should return nil")
	}
}
