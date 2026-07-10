package writer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMergeJSONCreateFromFragment(t *testing.T) {
	out, err := MergeJSON(nil, []byte(`{"options":{"skills_paths":[".crush/skills"]}}`))
	if err != nil {
		t.Fatal(err)
	}
	s := string(out)
	if !strings.Contains(s, `".crush/skills"`) || !strings.Contains(s, `"options"`) {
		t.Errorf("unexpected merged output: %s", s)
	}
}

func TestMergeJSONDeepMergePreservesUserConfig(t *testing.T) {
	existing := []byte(`{"options":{"skills_paths":["/global/skills"],"theme":"dark"},"model":"gpt"}`)
	fragment := []byte(`{"options":{"skills_paths":[".crush/skills"],"theme":"light"}}`)
	out, err := MergeJSON(existing, fragment)
	if err != nil {
		t.Fatal(err)
	}
	s := string(out)
	// array 并集：用户已有项保留在前，新项追加
	if !strings.Contains(s, `"/global/skills"`) || !strings.Contains(s, `".crush/skills"`) {
		t.Errorf("expected union of skills_paths, got: %s", s)
	}
	// scalar：existing 优先，不覆盖用户取值
	if !strings.Contains(s, `"theme": "dark"`) {
		t.Errorf("expected user theme preserved, got: %s", s)
	}
	if !strings.Contains(s, `"model": "gpt"`) {
		t.Errorf("expected untouched user key preserved, got: %s", s)
	}
}

func TestMergeJSONIdempotent(t *testing.T) {
	fragment := []byte(`{"instructions":[".kilo/rules/*.md"]}`)
	first, err := MergeJSON(nil, fragment)
	if err != nil {
		t.Fatal(err)
	}
	second, err := MergeJSON(first, fragment)
	if err != nil {
		t.Fatal(err)
	}
	if string(first) != string(second) {
		t.Errorf("merge not idempotent:\nfirst: %s\nsecond: %s", first, second)
	}
}

func TestMergeJSONRejectsJSONC(t *testing.T) {
	existing := []byte("{\n  // user comment\n  \"instructions\": []\n}")
	if _, err := MergeJSON(existing, []byte(`{"instructions":["x"]}`)); err == nil {
		t.Fatal("expected error for JSONC existing content")
	}
}

func TestApplyJSONMergeWritesAndSkipsUnchanged(t *testing.T) {
	dir := t.TempDir()
	w := NewWriter(dir, false)
	op := FileOp{Path: "crush.json", Content: []byte(`{"options":{"skills_paths":[".crush/skills"]}}`), JSONMerge: true}

	written, _, err := w.Apply(&Plan{Target: "crush", Files: []FileOp{op}})
	if err != nil {
		t.Fatal(err)
	}
	if written != 1 {
		t.Fatalf("expected 1 written, got %d", written)
	}

	// 第二次 apply 应 unchanged skip
	_, skipped, err := w.Apply(&Plan{Target: "crush", Files: []FileOp{op}})
	if err != nil {
		t.Fatal(err)
	}
	if skipped != 1 {
		t.Fatalf("expected 1 skipped on second apply, got %d", skipped)
	}
}

func TestApplyJSONMergeSkipsJSONCWithWarn(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "kilo.jsonc")
	if err := os.WriteFile(target, []byte("{\n  // keep my comment\n  \"instructions\": []\n}"), 0o600); err != nil {
		t.Fatal(err)
	}
	w := NewWriter(dir, false)
	plan := &Plan{Target: "kilo-code", Files: []FileOp{{
		Path: "kilo.jsonc", Content: []byte(`{"instructions":[".kilo/rules/*.md"]}`), JSONMerge: true,
	}}}
	written, skipped, err := w.Apply(plan)
	if err != nil {
		t.Fatal(err)
	}
	if written != 0 || skipped != 1 {
		t.Fatalf("expected skip, got written=%d skipped=%d", written, skipped)
	}
	if !strings.HasPrefix(plan.Files[0].Reason, "WARN") {
		t.Errorf("expected WARN reason, got %q", plan.Files[0].Reason)
	}
	// 用户文件必须原样保留
	after, _ := os.ReadFile(target) //nolint:gosec
	if !strings.Contains(string(after), "keep my comment") {
		t.Errorf("user JSONC file was modified: %s", after)
	}
}
