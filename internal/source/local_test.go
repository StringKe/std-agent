package source

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func TestLocalFiles(t *testing.T) {
	tmp := t.TempDir()
	mkAll(t, filepath.Join(tmp, "rules"))
	mkAll(t, filepath.Join(tmp, "skills/code-review"))
	mustWrite(t, filepath.Join(tmp, "rules/coding.md"), "rule")
	mustWrite(t, filepath.Join(tmp, "skills/code-review/SKILL.md"), "skill")
	mustWrite(t, filepath.Join(tmp, "rules/note.txt"), "ignored")

	l := NewLocal(tmp)
	files, err := l.Files()
	if err != nil {
		t.Fatal(err)
	}
	paths := make([]string, len(files))
	for i, f := range files {
		paths[i] = f.Path
	}
	sort.Strings(paths)
	want := []string{"rules/coding.md", "skills/code-review/SKILL.md"}
	if len(paths) != len(want) {
		t.Fatalf("got %v, want %v", paths, want)
	}
	for i, w := range want {
		if paths[i] != w {
			t.Errorf("path[%d] = %q, want %q", i, paths[i], w)
		}
	}
}

func TestLocalMissingDir(t *testing.T) {
	l := NewLocal(filepath.Join(t.TempDir(), "nope"))
	files, err := l.Files()
	if err != nil {
		t.Fatalf("expect nil err on missing, got %v", err)
	}
	if files != nil {
		t.Errorf("expect nil files, got %v", files)
	}
}

func TestLocalNamesIsLocal(t *testing.T) {
	if (&Local{}).Name() != "local" {
		t.Error("Name should be 'local'")
	}
}

func TestLocalSkillsSubtreeIncludesNonMarkdown(t *testing.T) {
	tmp := t.TempDir()
	mkAll(t, filepath.Join(tmp, "skills/code-review/scripts"))
	mkAll(t, filepath.Join(tmp, "skills/code-review/references"))
	mkAll(t, filepath.Join(tmp, "skills/code-review/assets"))
	mkAll(t, filepath.Join(tmp, "rules"))
	mustWrite(t, filepath.Join(tmp, "skills/code-review/SKILL.md"), "skill")
	mustWrite(t, filepath.Join(tmp, "skills/code-review/scripts/lint.sh"), "#!/bin/sh")
	mustWrite(t, filepath.Join(tmp, "skills/code-review/references/checklist.md"), "ref")
	mustWrite(t, filepath.Join(tmp, "skills/code-review/assets/template.json"), "{}")
	mustWrite(t, filepath.Join(tmp, "rules/coding.txt"), "ignored")

	files, err := NewLocal(tmp).Files()
	if err != nil {
		t.Fatal(err)
	}
	got := make(map[string]bool, len(files))
	for _, f := range files {
		got[f.Path] = true
	}
	for _, want := range []string{
		"skills/code-review/SKILL.md",
		"skills/code-review/scripts/lint.sh",
		"skills/code-review/references/checklist.md",
		"skills/code-review/assets/template.json",
	} {
		if !got[want] {
			t.Errorf("missing %s", want)
		}
	}
	if got["rules/coding.txt"] {
		t.Error("non-markdown outside skills/ should NOT be included")
	}
}

func TestIsMarkdownExtensions(t *testing.T) {
	for _, name := range []string{"a.md", "a.MD", "x.markdown", "X.Markdown"} {
		if !isMarkdown(name) {
			t.Errorf("%s should be markdown", name)
		}
	}
	for _, name := range []string{"a.txt", "x.json", "noext", "x.mdown"} {
		if isMarkdown(name) {
			t.Errorf("%s should NOT be markdown", name)
		}
	}
}

func TestShouldIncludeRules(t *testing.T) {
	cases := []struct {
		rel, name string
		want      bool
	}{
		{"rules/style.md", "style.md", true},
		{"rules/style.txt", "style.txt", false},
		{"skills/foo/SKILL.md", "SKILL.md", true},
		{"skills/foo/scripts/x.sh", "x.sh", true},
		{"skills/foo/assets/data.json", "data.json", true},
	}
	for _, c := range cases {
		if got := shouldInclude(c.rel, c.name); got != c.want {
			t.Errorf("shouldInclude(%q, %q) = %v, want %v", c.rel, c.name, got, c.want)
		}
	}
}

func mkAll(t *testing.T, p string) {
	t.Helper()
	if err := os.MkdirAll(p, 0o755); err != nil {
		t.Fatal(err)
	}
}

func mustWrite(t *testing.T, p, c string) {
	t.Helper()
	if err := os.WriteFile(p, []byte(c), 0o600); err != nil {
		t.Fatal(err)
	}
}
