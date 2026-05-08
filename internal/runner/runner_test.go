package runner

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSyncEndToEnd(t *testing.T) {
	tmp := t.TempDir()
	stdai := filepath.Join(tmp, ".stdai")
	mustMkdir(t, filepath.Join(stdai, "standards/rules"))
	mustMkdir(t, filepath.Join(stdai, "standards/skills/code-review"))

	mustWrite(t, filepath.Join(stdai, "standards/rules/coding-style.md"), `---
type: rules
name: coding-style
description: General style
priority: high
---

# Coding Style

Use meaningful variable names.
`)
	mustWrite(t, filepath.Join(stdai, "standards/skills/code-review/SKILL.md"), `---
type: skills
name: code-review
description: Review code carefully
---

# Code Review

Steps to review.
`)

	// 最简 config，仅 enable claude-code + codex，不配 sources
	mustWrite(t, filepath.Join(stdai, "config.toml"), `version = "1.0"
inject = true
inject_whatis = false
backup = false
auto_pull = false

[targets]
claude-code = { enabled = true, convert = true }
codex = { enabled = true, convert = true }
`)

	res, err := Sync(Options{
		ProjectRoot: tmp,
		ConfigPath:  filepath.Join(stdai, "config.toml"),
		Version:     "test",
	})
	if err != nil {
		t.Fatalf("Sync: %v", err)
	}
	if res.SourceFiles != 2 {
		t.Errorf("SourceFiles = %d, want 2", res.SourceFiles)
	}
	if res.Docs != 2 {
		t.Errorf("Docs = %d, want 2", res.Docs)
	}
	if res.Written < 2 {
		t.Errorf("Written = %d, want >= 2", res.Written)
	}

	// 验证关键扩散文件存在
	for _, want := range []string{
		"CLAUDE.md",
		filepath.Join(".claude", "rules", "coding-style.md"),
		filepath.Join(".claude", "skills", "code-review", "SKILL.md"),
		"AGENTS.md",
		filepath.Join(".agents", "skills", "code-review", "SKILL.md"),
	} {
		full := filepath.Join(tmp, want)
		if _, err := os.Stat(full); err != nil {
			t.Errorf("expected output %s: %v", want, err)
		}
	}

	// 验证 state.json 写入
	if _, err := os.Stat(filepath.Join(stdai, "state.json")); err != nil {
		t.Errorf("expected state.json: %v", err)
	}
}

func TestSyncDryRunDoesNotWrite(t *testing.T) {
	tmp := t.TempDir()
	stdai := filepath.Join(tmp, ".stdai")
	mustMkdir(t, filepath.Join(stdai, "standards/rules"))
	mustWrite(t, filepath.Join(stdai, "standards/rules/x.md"), `---
type: rules
name: x
---
body
`)
	mustWrite(t, filepath.Join(stdai, "config.toml"), `version = "1.0"
inject = false
backup = false
auto_pull = false

[targets]
claude-code = { enabled = true, convert = true }
`)

	res, err := Sync(Options{
		ProjectRoot: tmp,
		ConfigPath:  filepath.Join(stdai, "config.toml"),
		DryRun:      true,
		Version:     "test",
	})
	if err != nil {
		t.Fatalf("Sync: %v", err)
	}
	if res.Written != 0 {
		t.Errorf("dry-run Written = %d, want 0", res.Written)
	}
	if _, err := os.Stat(filepath.Join(tmp, "CLAUDE.md")); err == nil {
		t.Error("dry-run should not produce CLAUDE.md")
	}
}

func TestSyncSkillPackageWithSubdirs(t *testing.T) {
	tmp := t.TempDir()
	stdai := filepath.Join(tmp, ".stdai")
	mustMkdir(t, filepath.Join(stdai, "standards/skills/code-review/scripts"))
	mustMkdir(t, filepath.Join(stdai, "standards/skills/code-review/references"))

	mustWrite(t, filepath.Join(stdai, "standards/skills/code-review/SKILL.md"), `---
type: skills
name: code-review
description: Review skill
license: MIT
metadata:
  author: foo
---
body
`)
	mustWrite(t, filepath.Join(stdai, "standards/skills/code-review/scripts/lint.sh"), "#!/bin/sh\necho lint")
	mustWrite(t, filepath.Join(stdai, "standards/skills/code-review/references/checklist.md"), "checklist content")

	mustWrite(t, filepath.Join(stdai, "config.toml"), `version = "1.0"
inject = false
backup = false
auto_pull = false

[targets]
claude-code = { enabled = true, convert = true }
`)

	res, err := Sync(Options{
		ProjectRoot: tmp,
		ConfigPath:  filepath.Join(stdai, "config.toml"),
		Version:     "test",
	})
	if err != nil {
		t.Fatalf("Sync: %v", err)
	}
	if res.Docs != 1 {
		t.Errorf("Docs = %d, want 1 SKILL.md doc", res.Docs)
	}
	for _, want := range []string{
		filepath.Join(".claude/skills/code-review/SKILL.md"),
		filepath.Join(".claude/skills/code-review/scripts/lint.sh"),
		filepath.Join(".claude/skills/code-review/references/checklist.md"),
	} {
		if _, err := os.Stat(filepath.Join(tmp, want)); err != nil {
			t.Errorf("expected %s: %v", want, err)
		}
	}
	skillContent, _ := os.ReadFile(filepath.Join(tmp, ".claude/skills/code-review/SKILL.md"))
	for _, want := range []string{"license: MIT", "metadata:", "author: foo"} {
		if !strings.Contains(string(skillContent), want) {
			t.Errorf("SKILL.md missing %q in:\n%s", want, skillContent)
		}
	}
}

func TestSyncCodexCommandsInAGENTSMd(t *testing.T) {
	tmp := t.TempDir()
	stdai := filepath.Join(tmp, ".stdai")
	mustMkdir(t, filepath.Join(stdai, "standards/commands"))
	mustMkdir(t, filepath.Join(stdai, "standards/rules"))
	mustWrite(t, filepath.Join(stdai, "standards/rules/style.md"), `---
type: rules
name: style
---
Use clear names.
`)
	mustWrite(t, filepath.Join(stdai, "standards/commands/review.md"), `---
type: commands
name: review
description: Run code review
---
Please review the diff.
`)
	mustWrite(t, filepath.Join(stdai, "config.toml"), `version = "1.0"
inject = false
backup = false
auto_pull = false

[targets]
codex = { enabled = true, convert = true }
`)

	_, err := Sync(Options{
		ProjectRoot: tmp,
		ConfigPath:  filepath.Join(stdai, "config.toml"),
		Version:     "test",
	})
	if err != nil {
		t.Fatalf("Sync: %v", err)
	}
	agentsContent, err := os.ReadFile(filepath.Join(tmp, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	s := string(agentsContent)
	for _, want := range []string{
		"## style",                // rule 段
		"Use clear names.",        // rule body
		"## Slash Commands",       // commands 段
		"### /review",             // command name
		"Run code review",         // command description
		"Please review the diff.", // command body
	} {
		if !strings.Contains(s, want) {
			t.Errorf("AGENTS.md missing %q", want)
		}
	}
}

func TestSyncCollectsCopilotWARN(t *testing.T) {
	tmp := t.TempDir()
	stdai := filepath.Join(tmp, ".stdai")
	mustMkdir(t, filepath.Join(stdai, "standards/skills/foo/scripts"))
	mustWrite(t, filepath.Join(stdai, "standards/skills/foo/SKILL.md"), `---
type: skills
name: foo
description: Test
---
body
`)
	mustWrite(t, filepath.Join(stdai, "standards/skills/foo/scripts/x.sh"), "#!/bin/sh")
	mustWrite(t, filepath.Join(stdai, "config.toml"), `version = "1.0"
inject = false
backup = false
auto_pull = false

[targets]
copilot = { enabled = true, convert = true }
`)

	res, err := Sync(Options{
		ProjectRoot: tmp,
		ConfigPath:  filepath.Join(stdai, "config.toml"),
		Version:     "test",
	})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, w := range res.Warnings {
		if strings.Contains(w, "copilot") && strings.Contains(w, "WARN") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected copilot WARN in res.Warnings, got %v", res.Warnings)
	}
}

func TestSyncStrictParserError(t *testing.T) {
	tmp := t.TempDir()
	stdai := filepath.Join(tmp, ".stdai")
	mustMkdir(t, filepath.Join(stdai, "standards/rules"))
	// 非法 name（非 kebab-case）
	mustWrite(t, filepath.Join(stdai, "standards/rules/bad.md"), `---
type: rules
name: Bad_Name
---
body`)
	mustWrite(t, filepath.Join(stdai, "config.toml"), `version = "1.0"
inject = false
backup = false
auto_pull = false

[targets]
claude-code = { enabled = true, convert = true }
`)

	// 非 strict 应吞掉 parser 错误，记到 Warnings
	res, err := Sync(Options{
		ProjectRoot: tmp,
		ConfigPath:  filepath.Join(stdai, "config.toml"),
		Version:     "test",
	})
	if err != nil {
		t.Errorf("non-strict should not error, got %v", err)
	}
	if len(res.Warnings) == 0 {
		t.Error("expected warning for bad name")
	}

	// strict 模式应抛错
	if _, err := Sync(Options{
		ProjectRoot: tmp,
		ConfigPath:  filepath.Join(stdai, "config.toml"),
		Strict:      true,
		Version:     "test",
	}); err == nil {
		t.Error("strict mode should error on bad parser")
	}
}

func TestSyncUnknownTargetWithStrict(t *testing.T) {
	tmp := t.TempDir()
	stdai := filepath.Join(tmp, ".stdai")
	mustMkdir(t, filepath.Join(stdai, "standards/rules"))
	mustWrite(t, filepath.Join(stdai, "standards/rules/x.md"), `---
type: rules
name: x
---
body`)
	mustWrite(t, filepath.Join(stdai, "config.toml"), `version = "1.0"
inject = false
backup = false
auto_pull = false

[targets]
claude-code = { enabled = true, convert = true }
`)

	// 显式指定不存在的 target
	if _, err := Sync(Options{
		ProjectRoot: tmp,
		ConfigPath:  filepath.Join(stdai, "config.toml"),
		Targets:     []string{"nonexistent-target"},
		Strict:      true,
		Version:     "test",
	}); err == nil {
		t.Error("strict + unknown target should error")
	}
}

func TestSyncIgnoreFilter(t *testing.T) {
	tmp := t.TempDir()
	stdai := filepath.Join(tmp, ".stdai")
	mustMkdir(t, filepath.Join(stdai, "standards/rules"))
	mustWrite(t, filepath.Join(stdai, "standards/rules/keep.md"), `---
type: rules
name: keep
---
real`)
	mustWrite(t, filepath.Join(stdai, "standards/rules/draft-skip.md"), `---
type: rules
name: draft-skip
---
ignored`)
	mustWrite(t, filepath.Join(tmp, ".stdaiignore"), "rules/draft-*.md\n")
	mustWrite(t, filepath.Join(stdai, "config.toml"), `version = "1.0"
inject = false
backup = false
auto_pull = false

[targets]
claude-code = { enabled = true, convert = true }
`)

	res, err := Sync(Options{
		ProjectRoot: tmp,
		ConfigPath:  filepath.Join(stdai, "config.toml"),
		Version:     "test",
	})
	if err != nil {
		t.Fatalf("Sync: %v", err)
	}
	if res.Docs != 1 {
		t.Errorf("Docs = %d, want 1 (draft-skip ignored)", res.Docs)
	}
	if _, err := os.Stat(filepath.Join(tmp, ".claude/rules/draft-skip.md")); err == nil {
		t.Error("draft-skip.md should have been filtered out")
	}
	if _, err := os.Stat(filepath.Join(tmp, ".claude/rules/keep.md")); err != nil {
		t.Errorf("keep.md should exist: %v", err)
	}
}

func TestSyncHooksConversion(t *testing.T) {
	tmp := t.TempDir()
	stdai := filepath.Join(tmp, ".stdai")
	mustMkdir(t, filepath.Join(stdai, "standards/rules"))
	mustWrite(t, filepath.Join(stdai, "standards/rules/x.md"), `---
type: rules
name: x
---
body`)
	mustWrite(t, filepath.Join(stdai, "standards/hooks.json"), `{
  "version": "1.0",
  "hooks": {
    "PreToolUse": [
      {"matcher": "Bash", "type": "command", "command": "echo hi"}
    ]
  }
}
`)
	mustWrite(t, filepath.Join(stdai, "config.toml"), `version = "1.0"
inject = false
backup = false
auto_pull = false

[targets]
claude-code = { enabled = true, convert = true }
codex = { enabled = true, convert = true }
`)

	_, err := Sync(Options{
		ProjectRoot: tmp,
		ConfigPath:  filepath.Join(stdai, "config.toml"),
		Version:     "test",
	})
	if err != nil {
		t.Fatalf("Sync: %v", err)
	}

	for _, want := range []string{
		".claude/stdagent-hooks.json",
		".codex/stdagent-hooks.json",
	} {
		full := filepath.Join(tmp, want)
		data, rerr := os.ReadFile(full) //nolint:gosec
		if rerr != nil {
			t.Errorf("%s should exist: %v", want, rerr)
			continue
		}
		s := string(data)
		for _, sub := range []string{"PreToolUse", "echo hi", "Bash"} {
			if !strings.Contains(s, sub) {
				t.Errorf("%s missing %q in:\n%s", want, sub, s)
			}
		}
	}
}

func TestSyncHooksMalformedNonStrict(t *testing.T) {
	tmp := t.TempDir()
	stdai := filepath.Join(tmp, ".stdai")
	mustMkdir(t, filepath.Join(stdai, "standards/rules"))
	mustWrite(t, filepath.Join(stdai, "standards/rules/x.md"), `---
type: rules
name: x
---
body`)
	mustWrite(t, filepath.Join(stdai, "standards/hooks.json"), `{not json`)
	mustWrite(t, filepath.Join(stdai, "config.toml"), `version = "1.0"
inject = false
backup = false
auto_pull = false

[targets]
claude-code = { enabled = true, convert = true }
`)

	res, err := Sync(Options{
		ProjectRoot: tmp,
		ConfigPath:  filepath.Join(stdai, "config.toml"),
		Version:     "test",
	})
	if err != nil {
		t.Fatalf("non-strict should swallow parse err: %v", err)
	}
	found := false
	for _, w := range res.Warnings {
		if strings.Contains(w, "hooks.json") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected hooks.json warning, got %v", res.Warnings)
	}
}

func mustMkdir(t *testing.T, p string) {
	t.Helper()
	if err := os.MkdirAll(p, 0o755); err != nil {
		t.Fatal(err)
	}
}

func mustWrite(t *testing.T, p, content string) {
	t.Helper()
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}
