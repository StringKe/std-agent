package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/StringKe/std-agent/internal/parser"
)

func TestWhichMatchesApplyToGlob(t *testing.T) {
	root := t.TempDir()
	bootstrapWhichFixture(t, root)

	out, err := runWhichCmd(root, []string{"internal/runner/runner.go"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "test-requirements") {
		t.Errorf("expected test-requirements in output, got:\n%s", out)
	}
	if strings.Contains(out, "no-history-rewrite") {
		t.Error("no-history-rewrite has no applyTo, should NOT show without --include-global")
	}
}

func TestWhichJSONOutput(t *testing.T) {
	root := t.TempDir()
	bootstrapWhichFixture(t, root)

	out, err := runWhichCmd(root, []string{"internal/runner/runner.go", "--json"})
	if err != nil {
		t.Fatal(err)
	}
	var payload whichPayload
	if jerr := json.Unmarshal([]byte(out), &payload); jerr != nil {
		t.Fatalf("invalid JSON: %v\n%s", jerr, out)
	}
	if payload.Target != "internal/runner/runner.go" {
		t.Errorf("Target = %q", payload.Target)
	}
	if len(payload.Matches) != 1 {
		t.Errorf("want 1 match, got %d", len(payload.Matches))
	}
	if payload.Matches[0].Name != "test-requirements" {
		t.Errorf("match[0].Name = %q", payload.Matches[0].Name)
	}
	if len(payload.Matches[0].MatchedGlobs) == 0 {
		t.Error("matched_globs should not be empty")
	}
}

func TestWhichPathsOnly(t *testing.T) {
	root := t.TempDir()
	bootstrapWhichFixture(t, root)

	out, err := runWhichCmd(root, []string{"internal/runner/runner.go", "--paths"})
	if err != nil {
		t.Fatal(err)
	}
	got := strings.TrimSpace(out)
	want := filepath.ToSlash(filepath.Join(".stdai/standards", "rules/test-requirements.md"))
	// 把输出按行 split，规范化
	lines := strings.Split(got, "\n")
	found := false
	for _, l := range lines {
		if filepath.ToSlash(strings.TrimSpace(l)) == want {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected %s in --paths output, got:\n%s", want, got)
	}
}

func TestWhichIncludeGlobalAddsGlobalDocs(t *testing.T) {
	root := t.TempDir()
	bootstrapWhichFixture(t, root)

	out, err := runWhichCmd(root, []string{"internal/runner/runner.go", "--include-global"})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"test-requirements", "no-history-rewrite", "commit-style"} {
		if !strings.Contains(out, want) {
			t.Errorf("--include-global should list %s, got:\n%s", want, out)
		}
	}
}

func TestWhichTypeFilter(t *testing.T) {
	root := t.TempDir()
	bootstrapWhichFixture(t, root)
	// 加一条 reference doc，applyTo 匹配同样路径
	mustMkdirAll(t, filepath.Join(root, ".stdai/standards/references"))
	mustWriteFile(t, filepath.Join(root, ".stdai/standards/references/runner-arch.md"), `---
type: references
name: runner-arch
description: runner 架构参考
applyTo:
  - "internal/runner/**/*.go"
---
body
`)

	// 不过滤：应有 test-requirements(rules) + runner-arch(references)
	all, err := runWhichCmd(root, []string{"internal/runner/runner.go"})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"test-requirements", "runner-arch"} {
		if !strings.Contains(all, want) {
			t.Errorf("default output missing %s", want)
		}
	}

	// 仅 references
	refs, err := runWhichCmd(root, []string{"internal/runner/runner.go", "--type=references"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(refs, "runner-arch") {
		t.Errorf("--type=references missing runner-arch")
	}
	if strings.Contains(refs, "test-requirements") {
		t.Error("--type=references should not include rules")
	}
}

// runWhichCmd 在 isolated cwd 跑 which，捕获 stdout
func runWhichCmd(root string, args []string) (string, error) {
	cwd, _ := os.Getwd()
	defer os.Chdir(cwd) //nolint:errcheck // test-only
	if err := os.Chdir(root); err != nil {
		return "", err
	}

	// 重置 flag 状态（cobra command 复用全局 root flag，per-test 重新构造 root）
	saveConfig := flagConfig
	flagConfig = ".stdai/config.toml"
	defer func() { flagConfig = saveConfig }()

	rootCmd := newRootCmd()
	rootCmd.SetArgs(append([]string{"which"}, args...))
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	err := rootCmd.Execute()
	return buf.String(), err
}

func bootstrapWhichFixture(t *testing.T, root string) {
	t.Helper()
	stdai := filepath.Join(root, ".stdai")
	mustMkdirAll(t, filepath.Join(stdai, "standards/rules"))
	mustWriteFile(t, filepath.Join(stdai, "config.toml"), `version = "1.0"
[targets]
claude-code = { enabled = true, convert = true }
`)
	// 有 applyTo（runner 路径专属）
	mustWriteFile(t, filepath.Join(stdai, "standards/rules/test-requirements.md"), `---
type: rules
name: test-requirements
description: 改 runner 必带测试
priority: high
applyTo:
  - "internal/runner/**/*.go"
---
body
`)
	// 无 applyTo（全局，默认不出现，--include-global 才出现）
	mustWriteFile(t, filepath.Join(stdai, "standards/rules/no-history-rewrite.md"), `---
type: rules
name: no-history-rewrite
description: 禁止改写历史
priority: high
---
body
`)
	mustWriteFile(t, filepath.Join(stdai, "standards/rules/commit-style.md"), `---
type: rules
name: commit-style
description: Conventional Commits
priority: high
---
body
`)
	// sanity check: parser 接受 fixture
	files, _ := os.ReadDir(filepath.Join(stdai, "standards/rules"))
	if len(files) != 3 {
		t.Fatalf("fixture: want 3 rule files, got %d", len(files))
	}
	_ = parser.TypeRules
}

func mustMkdirAll(t *testing.T, p string) {
	t.Helper()
	if err := os.MkdirAll(p, 0o755); err != nil {
		t.Fatal(err)
	}
}

func mustWriteFile(t *testing.T, p, content string) {
	t.Helper()
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}
