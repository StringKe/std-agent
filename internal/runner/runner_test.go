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
		".codex/memories/style.md", // rule 索引引用
		"## Slash Commands",        // commands 段
		"### /review",              // command name
		"Run code review",          // command description
		"Please review the diff.",  // command body
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

// TestSyncPrunesOrphans 复现"规则从 30 优化为 15"的典型场景：
// 第一次 sync 写入完整集合，删源文件后第二次 sync 应该把过时输出删掉。
func TestSyncPrunesOrphans(t *testing.T) {
	tmp := t.TempDir()
	stdai := filepath.Join(tmp, ".stdai")
	mustMkdir(t, filepath.Join(stdai, "standards/rules"))

	for _, name := range []string{"keep-a", "keep-b", "drop-c", "drop-d"} {
		mustWrite(t, filepath.Join(stdai, "standards/rules", name+".md"), `---
type: rules
name: `+name+`
---
body of `+name+`
`)
	}
	mustWrite(t, filepath.Join(stdai, "config.toml"), `version = "1.0"
inject = false
backup = false
auto_pull = false

[targets]
claude-code = { enabled = true, convert = true }
`)

	if _, err := Sync(Options{
		ProjectRoot: tmp,
		ConfigPath:  filepath.Join(stdai, "config.toml"),
		Version:     "test",
	}); err != nil {
		t.Fatalf("first Sync: %v", err)
	}
	for _, n := range []string{"keep-a", "keep-b", "drop-c", "drop-d"} {
		if _, err := os.Stat(filepath.Join(tmp, ".claude/rules", n+".md")); err != nil {
			t.Fatalf("first sync missing %s: %v", n, err)
		}
	}

	mustRemove(t, filepath.Join(stdai, "standards/rules/drop-c.md"))
	mustRemove(t, filepath.Join(stdai, "standards/rules/drop-d.md"))

	res, err := Sync(Options{
		ProjectRoot: tmp,
		ConfigPath:  filepath.Join(stdai, "config.toml"),
		Version:     "test",
	})
	if err != nil {
		t.Fatalf("second Sync: %v", err)
	}
	if res.Pruned != 2 {
		t.Errorf("Pruned = %d, want 2", res.Pruned)
	}
	for _, n := range []string{"drop-c", "drop-d"} {
		if _, err := os.Stat(filepath.Join(tmp, ".claude/rules", n+".md")); err == nil {
			t.Errorf(".claude/rules/%s.md should have been pruned", n)
		}
	}
	for _, n := range []string{"keep-a", "keep-b"} {
		if _, err := os.Stat(filepath.Join(tmp, ".claude/rules", n+".md")); err != nil {
			t.Errorf(".claude/rules/%s.md should survive: %v", n, err)
		}
	}
}

// TestSyncNoPruneKeepsOrphans 验证 --no-prune 跳过删除
func TestSyncNoPruneKeepsOrphans(t *testing.T) {
	tmp := t.TempDir()
	stdai := filepath.Join(tmp, ".stdai")
	mustMkdir(t, filepath.Join(stdai, "standards/rules"))
	mustWrite(t, filepath.Join(stdai, "standards/rules/keep.md"), `---
type: rules
name: keep
---
body
`)
	mustWrite(t, filepath.Join(stdai, "standards/rules/drop.md"), `---
type: rules
name: drop
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

	if _, err := Sync(Options{
		ProjectRoot: tmp,
		ConfigPath:  filepath.Join(stdai, "config.toml"),
		Version:     "test",
	}); err != nil {
		t.Fatal(err)
	}
	mustRemove(t, filepath.Join(stdai, "standards/rules/drop.md"))

	res, err := Sync(Options{
		ProjectRoot: tmp,
		ConfigPath:  filepath.Join(stdai, "config.toml"),
		NoPrune:     true,
		Version:     "test",
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Pruned != 0 {
		t.Errorf("Pruned = %d, want 0 with NoPrune", res.Pruned)
	}
	if _, err := os.Stat(filepath.Join(tmp, ".claude/rules/drop.md")); err != nil {
		t.Errorf("drop.md should remain under NoPrune: %v", err)
	}
}

// TestSyncDryRunReportsPrune DryRun 模式只汇报、不真删
func TestSyncDryRunReportsPrune(t *testing.T) {
	tmp := t.TempDir()
	stdai := filepath.Join(tmp, ".stdai")
	mustMkdir(t, filepath.Join(stdai, "standards/rules"))
	mustWrite(t, filepath.Join(stdai, "standards/rules/keep.md"), `---
type: rules
name: keep
---
body
`)
	mustWrite(t, filepath.Join(stdai, "standards/rules/drop.md"), `---
type: rules
name: drop
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
	if _, err := Sync(Options{
		ProjectRoot: tmp,
		ConfigPath:  filepath.Join(stdai, "config.toml"),
		Version:     "test",
	}); err != nil {
		t.Fatal(err)
	}
	mustRemove(t, filepath.Join(stdai, "standards/rules/drop.md"))

	res, err := Sync(Options{
		ProjectRoot: tmp,
		ConfigPath:  filepath.Join(stdai, "config.toml"),
		DryRun:      true,
		Version:     "test",
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Pruned != 1 {
		t.Errorf("Pruned = %d, want 1 (dry-run reports)", res.Pruned)
	}
	if _, err := os.Stat(filepath.Join(tmp, ".claude/rules/drop.md")); err != nil {
		t.Error("dry-run should NOT actually delete drop.md")
	}
}

// TestSyncUnchangedStillTrackedInState 修复预存 bug：上次写入 + 这次内容不变（unchanged skip）
// 的文件必须保留在 state.Outputs 中。否则之后一旦源被删，第三次 sync 无从识别孤儿。
func TestSyncUnchangedStillTrackedInState(t *testing.T) {
	tmp := t.TempDir()
	stdai := filepath.Join(tmp, ".stdai")
	mustMkdir(t, filepath.Join(stdai, "standards/rules"))
	mustWrite(t, filepath.Join(stdai, "standards/rules/stable.md"), `---
type: rules
name: stable
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
	for i := 0; i < 2; i++ {
		if _, err := Sync(Options{
			ProjectRoot: tmp,
			ConfigPath:  filepath.Join(stdai, "config.toml"),
			Version:     "test",
		}); err != nil {
			t.Fatalf("sync %d: %v", i, err)
		}
	}

	mustRemove(t, filepath.Join(stdai, "standards/rules/stable.md"))
	res, err := Sync(Options{
		ProjectRoot: tmp,
		ConfigPath:  filepath.Join(stdai, "config.toml"),
		Version:     "test",
	})
	if err != nil {
		t.Fatal(err)
	}
	// 关键断言：stable.md 必须是 orphan 之一（即 unchanged 路径在 state 里被跟踪了）。
	// 不限定 Pruned 总数，因为 transformer 可能同时 prune 其他衍生文件（whatis 索引等）。
	found := false
	for _, p := range res.PrunedPaths {
		if strings.HasSuffix(p, ".claude/rules/stable.md") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("stable.md not in PrunedPaths = %v", res.PrunedPaths)
	}
	if _, err := os.Stat(filepath.Join(tmp, ".claude/rules/stable.md")); err == nil {
		t.Error("stable.md should be pruned from disk")
	}
}

func mustRemove(t *testing.T, p string) {
	t.Helper()
	if err := os.Remove(p); err != nil {
		t.Fatal(err)
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
