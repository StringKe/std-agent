package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestVersionCommandText(t *testing.T) {
	old := versionStr
	defer func() { versionStr = old }()
	SetVersion("1.2.3", "abc1234", "2026-05-08T00:00:00Z")

	cmd := newVersionCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(new(bytes.Buffer))
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	s := out.String()
	for _, want := range []string{"stdagent 1.2.3", "abc1234", "2026-05-08T00:00:00Z"} {
		if !strings.Contains(s, want) {
			t.Errorf("missing %q in:\n%s", want, s)
		}
	}
}

func TestVersionCommandJSON(t *testing.T) {
	oldV, oldC, oldD := versionStr, commitStr, dateStr
	defer func() { versionStr, commitStr, dateStr = oldV, oldC, oldD }()
	SetVersion("9.9.9", "deadbeef", "2026-01-01T00:00:00Z")

	cmd := newVersionCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs([]string{"--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	var data map[string]string
	if err := json.Unmarshal(out.Bytes(), &data); err != nil {
		t.Fatalf("not valid JSON: %v\n%s", err, out.String())
	}
	if data["version"] != "9.9.9" {
		t.Errorf("version = %s", data["version"])
	}
	if data["commit"] != "deadbeef" {
		t.Errorf("commit = %s", data["commit"])
	}
	if data["built"] != "2026-01-01T00:00:00Z" {
		t.Errorf("built = %s", data["built"])
	}
	if data["go"] == "" || data["os"] == "" || data["arch"] == "" {
		t.Errorf("runtime fields empty: %v", data)
	}
}

func TestSetVersionEmptyKeepsDefault(t *testing.T) {
	old := versionStr
	defer func() { versionStr = old }()
	SetVersion("0.5.0", "x", "y")
	SetVersion("", "", "")
	if versionStr != "0.5.0" {
		t.Errorf("empty SetVersion should not overwrite, got %s", versionStr)
	}
}
