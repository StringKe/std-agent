package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestExplainAll(t *testing.T) {
	cmd := newExplainCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(new(bytes.Buffer))
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	got := out.String()
	for _, want := range []string{
		"std-agent 类型",
		"## rules",
		"## skills",
		"## commands",
		"## references",
		"## subagents",
		"## 选择标准",
		"stdagent budget --rendered",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in default explain output", want)
		}
	}
}

func TestExplainSingle(t *testing.T) {
	cmd := newExplainCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs([]string{"rules"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	got := out.String()
	if !strings.Contains(got, "## rules") {
		t.Error("rules section missing")
	}
	if !strings.Contains(got, "applyTo") {
		t.Error("rules example frontmatter missing applyTo")
	}
	// 不应含其他 type 段
	for _, other := range []string{"## skills", "## commands", "## references", "## subagents"} {
		if strings.Contains(got, other) {
			t.Errorf("single explain rules should not contain %q", other)
		}
	}
}

func TestExplainSingleAllTypes(t *testing.T) {
	types := []string{"rules", "skills", "commands", "references", "subagents"}
	for _, ty := range types {
		t.Run(ty, func(t *testing.T) {
			cmd := newExplainCmd()
			var out bytes.Buffer
			cmd.SetOut(&out)
			cmd.SetErr(new(bytes.Buffer))
			cmd.SetArgs([]string{ty})
			if err := cmd.Execute(); err != nil {
				t.Fatal(err)
			}
			got := out.String()
			if !strings.Contains(got, "## "+ty) {
				t.Errorf("missing ## %s header", ty)
			}
		})
	}
}

func TestExplainUnknownType(t *testing.T) {
	cmd := newExplainCmd()
	var out, errBuf bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)
	cmd.SetArgs([]string{"bogus"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for unknown type")
	}
	if !strings.Contains(err.Error(), "unknown type") {
		t.Errorf("expected 'unknown type' in error, got: %v", err)
	}
}

func TestExplainJSON(t *testing.T) {
	cmd := newExplainCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs([]string{"--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	var arr []explainType
	if err := json.Unmarshal(out.Bytes(), &arr); err != nil {
		t.Fatalf("not valid JSON: %v\n%s", err, out.String())
	}
	if len(arr) != 5 {
		t.Fatalf("expected 5 types, got %d", len(arr))
	}
	wantTypes := map[string]bool{
		"rules": false, "skills": false, "commands": false,
		"references": false, "subagents": false,
	}
	for _, e := range arr {
		if _, ok := wantTypes[e.Type]; !ok {
			t.Errorf("unexpected type %q", e.Type)
			continue
		}
		wantTypes[e.Type] = true
		if e.Semantics == "" || e.WhenToUse == "" || e.WhenNot == "" || e.ExampleFM == "" {
			t.Errorf("type %s missing required fields", e.Type)
		}
		if !strings.Contains(e.ExampleFM, "type: "+e.Type) {
			t.Errorf("type %s example_frontmatter should contain 'type: %s', got: %s", e.Type, e.Type, e.ExampleFM)
		}
	}
	for ty, seen := range wantTypes {
		if !seen {
			t.Errorf("missing type %q in JSON output", ty)
		}
	}
}

func TestExplainJSONSingle(t *testing.T) {
	cmd := newExplainCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs([]string{"rules", "--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	var arr []explainType
	if err := json.Unmarshal(out.Bytes(), &arr); err != nil {
		t.Fatalf("not valid JSON: %v\n%s", err, out.String())
	}
	if len(arr) != 1 {
		t.Fatalf("expected 1 item, got %d", len(arr))
	}
	if arr[0].Type != "rules" {
		t.Errorf("expected rules, got %s", arr[0].Type)
	}
}

func TestExplainJSONUnknownType(t *testing.T) {
	cmd := newExplainCmd()
	var out, errBuf bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)
	cmd.SetArgs([]string{"nope", "--json"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error for unknown type with --json")
	}
}

func TestExplainTooManyArgs(t *testing.T) {
	cmd := newExplainCmd()
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs([]string{"rules", "skills"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error for too many args")
	}
}
