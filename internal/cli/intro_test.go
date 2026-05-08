package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestIntroCommandText(t *testing.T) {
	cmd := newIntroCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(new(bytes.Buffer))
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	got := out.String()
	for _, want := range []string{
		"stdagent AI 助手提示词",
		".stdai/standards/",
		"frontmatter",
		"rules",
		"skills",
		"commands",
		"references",
		"claude-code",
		"applyTo",
		"stdagent sync",
		"stdagent budget",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in intro output", want)
		}
	}
}

func TestIntroCommandJSON(t *testing.T) {
	old := versionStr
	defer func() { versionStr = old }()
	versionStr = "1.0.0"

	cmd := newIntroCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs([]string{"--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	var p introPayload
	if err := json.Unmarshal(out.Bytes(), &p); err != nil {
		t.Fatalf("not valid JSON: %v\n%s", err, out.String())
	}
	if p.Version != "1.0.0" {
		t.Errorf("version = %s", p.Version)
	}
	if !strings.Contains(p.Prompt, "stdagent AI 助手提示词") {
		t.Error("prompt missing header")
	}
}

func TestIntroPromptHelper(t *testing.T) {
	got := IntroPrompt()
	if len(got) < 1000 {
		t.Errorf("prompt too short: %d chars", len(got))
	}
	for _, want := range []string{".stdai/standards/", "stdagent sync"} {
		if !strings.Contains(got, want) {
			t.Errorf("prompt should mention %q", want)
		}
	}
}

func TestIntroCommandCopyMode(t *testing.T) {
	cmd := newIntroCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs([]string{"--json", "--copy"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	got := out.String()
	// --copy 输出原始 markdown，不应是 JSON 包装
	if strings.HasPrefix(got, "{") {
		t.Error("--copy should output raw markdown not JSON")
	}
	if !strings.Contains(got, "stdagent AI 助手提示词") {
		t.Error("missing prompt content")
	}
}
