package runner

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/StringKe/std-agent/internal/writer"
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

	if !res.GitignoreUpdated {
		t.Error("first sync should write managed gitignore block")
	}
	gi, err := os.ReadFile(filepath.Join(tmp, ".gitignore"))
	if err != nil {
		t.Fatalf(".gitignore: %v", err)
	}
	for _, want := range []string{"# BEGIN stdagent", "CLAUDE.md", "AGENTS.md", ".agents/", ".claude/"} {
		if !strings.Contains(string(gi), want) {
			t.Errorf(".gitignore missing %q:\n%s", want, gi)
		}
	}
	res2, err := Sync(Options{
		ProjectRoot: tmp,
		ConfigPath:  filepath.Join(stdai, "config.toml"),
		Version:     "test",
	})
	if err != nil {
		t.Fatalf("repeat Sync: %v", err)
	}
	if res2.GitignoreUpdated {
		t.Error("repeat sync should leave gitignore unchanged")
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
	if !res.GitignoreUpdated {
		t.Error("dry-run should report gitignore change")
	}
	if _, err := os.Stat(filepath.Join(tmp, ".gitignore")); !os.IsNotExist(err) {
		t.Error("dry-run must not write .gitignore")
	}
}

func TestSyncGitignoreOffLeavesFileUntouched(t *testing.T) {
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
gitignore = "off"

[targets]
claude-code = { enabled = true, convert = true }
`)
	gi := filepath.Join(tmp, ".gitignore")
	mustWrite(t, gi, "bin/\n")
	res, err := Sync(Options{
		ProjectRoot: tmp,
		ConfigPath:  filepath.Join(stdai, "config.toml"),
		Version:     "test",
	})
	if err != nil {
		t.Fatalf("Sync: %v", err)
	}
	if res.GitignoreUpdated {
		t.Error("off must not update gitignore")
	}
	got, err := os.ReadFile(gi)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "bin/\n" {
		t.Fatalf("off must not touch .gitignore, got:\n%s", got)
	}
}

func TestSyncGitignorePortableKeepsAgents(t *testing.T) {
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
gitignore = "portable"

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
	if !res.GitignoreUpdated {
		t.Fatal("portable sync should write managed gitignore block")
	}
	gi, err := os.ReadFile(filepath.Join(tmp, ".gitignore"))
	if err != nil {
		t.Fatal(err)
	}
	s := string(gi)
	if hasGitignoreLine(s, "AGENTS.md") || hasGitignoreLine(s, ".agents/") {
		t.Fatalf("portable must keep AGENTS.md and .agents/:\n%s", s)
	}
	for _, want := range []string{"# BEGIN stdagent", "CLAUDE.md", ".claude/"} {
		if !strings.Contains(s, want) {
			t.Errorf("portable missing %q:\n%s", want, s)
		}
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

func TestSyncCodexCommandsStayOutOfSharedAGENTSMd(t *testing.T) {
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
	for _, want := range []string{"Use clear names."} {
		if !strings.Contains(s, want) {
			t.Errorf("AGENTS.md missing %q", want)
		}
	}
	if strings.Contains(s, "## Slash Commands") {
		t.Errorf("shared AGENTS.md should not contain target-specific commands:\n%s", s)
	}
	commandContent, err := os.ReadFile(filepath.Join(tmp, ".agents/skills/commands/review/SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Run code review", "Please review the diff."} {
		if !strings.Contains(string(commandContent), want) {
			t.Errorf("command skill missing %q", want)
		}
	}
}

func TestSyncSharedAGENTSStableAcrossTargets(t *testing.T) {
	tmp := t.TempDir()
	stdai := filepath.Join(tmp, ".stdai")
	mustMkdir(t, filepath.Join(stdai, "standards/rules"))
	mustMkdir(t, filepath.Join(stdai, "standards/commands"))
	mustWrite(t, filepath.Join(stdai, "standards/root.md"), `# Project

Project overview.
`)
	mustWrite(t, filepath.Join(stdai, "standards/rules/style.md"), `---
type: rules
name: style
description: Shared style
---
Use clear names.
`)
	mustWrite(t, filepath.Join(stdai, "standards/commands/review.md"), `---
type: commands
name: review
description: Run review
---
Review the current diff.
`)
	mustWrite(t, filepath.Join(stdai, "config.toml"), `version = "1.0"
inject = false
inject_type_glossary = false
backup = false
auto_pull = false

[targets]
codex = { enabled = true, convert = true }
factory = { enabled = true, convert = true }
`)

	first, err := Sync(Options{
		ProjectRoot: tmp,
		ConfigPath:  filepath.Join(stdai, "config.toml"),
		Version:     "test",
	})
	if err != nil {
		t.Fatalf("first Sync: %v", err)
	}
	if first.Written == 0 {
		t.Fatal("first sync should write outputs")
	}

	content, err := os.ReadFile(filepath.Join(tmp, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	got := string(content)
	for _, want := range []string{"Project overview.", "Use clear names."} {
		if !strings.Contains(got, want) {
			t.Errorf("shared AGENTS.md missing %q:\n%s", want, got)
		}
	}
	for _, unwanted := range []string{".factory/rules", "## Slash Commands"} {
		if strings.Contains(got, unwanted) {
			t.Errorf("shared AGENTS.md should not contain target-specific %q:\n%s", unwanted, got)
		}
	}

	second, err := Sync(Options{
		ProjectRoot: tmp,
		ConfigPath:  filepath.Join(stdai, "config.toml"),
		Version:     "test",
	})
	if err != nil {
		t.Fatalf("second Sync: %v", err)
	}
	if second.Written != 0 {
		t.Errorf("second sync should be stable, wrote %d files", second.Written)
	}
}

func TestValidatePlanCollisions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		plans   []*writer.Plan
		wantErr bool
	}{
		{
			name: "identical shared output",
			plans: []*writer.Plan{
				{Target: "codex", Files: []writer.FileOp{{Path: "AGENTS.md", Content: []byte("same")}}},
				{Target: "factory", Files: []writer.FileOp{{Path: "AGENTS.md", Content: []byte("same")}}},
			},
		},
		{
			name: "divergent shared output",
			plans: []*writer.Plan{
				{Target: "codex", Files: []writer.FileOp{{Path: "AGENTS.md", Content: []byte("one")}}},
				{Target: "factory", Files: []writer.FileOp{{Path: "AGENTS.md", Content: []byte("two")}}},
			},
			wantErr: true,
		},
		{
			name: "different merge semantics",
			plans: []*writer.Plan{
				{Target: "one", Files: []writer.FileOp{{Path: "settings.json", Content: []byte("{}")}}},
				{Target: "two", Files: []writer.FileOp{{Path: "settings.json", Content: []byte("{}"), JSONMerge: true}}},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := validatePlanCollisions(tt.plans)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validatePlanCollisions() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && !strings.Contains(err.Error(), tt.plans[0].Files[0].Path) {
				t.Errorf("error should identify collision path: %v", err)
			}
		})
	}
}

// TestSyncCopilotNativeSkill 验证 copilot skills 走原生 .github/skills/ 包
// （Agent Skills 已 GA，历史的 .instructions.md 降级 + WARN 行为已移除）。
func TestSyncCopilotNativeSkill(t *testing.T) {
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
	mustWrite(t, filepath.Join(stdai, "standards/skills/foo/scripts/x.md"), "aux file")
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
	for _, w := range res.Warnings {
		if strings.Contains(w, "copilot") && strings.Contains(w, "WARN") {
			t.Errorf("copilot native skills should not WARN, got %v", res.Warnings)
		}
	}
	if _, serr := os.Stat(filepath.Join(tmp, ".github/skills/foo/SKILL.md")); serr != nil {
		t.Errorf("expected native .github/skills/foo/SKILL.md: %v", serr)
	}
	if _, serr := os.Stat(filepath.Join(tmp, ".github/skills/foo/scripts/x.md")); serr != nil {
		t.Errorf("expected skill aux file fan-out: %v", serr)
	}
}

// TestSyncCollectsJSONMergeWARN 验证 Apply 阶段产生的 WARN（kilo.jsonc 为 JSONC
// 无法解析时跳过合并）会被 runner 收集，且用户文件不被改动、不被 prune。
func TestSyncCollectsJSONMergeWARN(t *testing.T) {
	tmp := t.TempDir()
	stdai := filepath.Join(tmp, ".stdai")
	mustMkdir(t, filepath.Join(stdai, "standards/rules"))
	mustWrite(t, filepath.Join(stdai, "standards/rules/style.md"), `---
type: rules
name: style
description: Style
---
body
`)
	mustWrite(t, filepath.Join(stdai, "config.toml"), `version = "1.0"
inject = false
backup = false
auto_pull = false

[targets]
kilo-code = { enabled = true, convert = true }
`)
	userConfig := "{\n  // user comment survives\n  \"instructions\": []\n}"
	mustWrite(t, filepath.Join(tmp, "kilo.jsonc"), userConfig)

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
		if strings.Contains(w, "kilo.jsonc") && strings.Contains(w, "WARN") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected kilo.jsonc JSONMerge WARN in res.Warnings, got %v", res.Warnings)
	}
	after, _ := os.ReadFile(filepath.Join(tmp, "kilo.jsonc")) //nolint:gosec
	if string(after) != userConfig {
		t.Errorf("user kilo.jsonc must stay untouched, got:\n%s", after)
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

// TestSyncPrunesLegacyCmdPrefixedSkill：v3 把 command skill 从 cmd-<n>/ 迁到 commands/<n>/，
// 磁盘上未进 state 的旧路径也应被 prune（Codex 会扫描 .agents/skills/ 下全部 SKILL.md）。
func TestSyncPrunesLegacyCmdPrefixedSkill(t *testing.T) {
	tmp := t.TempDir()
	stdai := filepath.Join(tmp, ".stdai")
	mustMkdir(t, filepath.Join(stdai, "standards/commands"))
	mustWrite(t, filepath.Join(stdai, "standards/commands", "ship.md"), `---
type: commands
name: ship
description: Ship release
---
steps
`)
	mustWrite(t, filepath.Join(stdai, "config.toml"), `version = "1.0"
inject = false
backup = false
auto_pull = false

[targets]
codex = { enabled = true, convert = true }
`)
	legacyDir := filepath.Join(tmp, ".agents/skills/cmd-ship")
	mustMkdir(t, legacyDir)
	mustWrite(t, filepath.Join(legacyDir, "SKILL.md"), "## broken legacy skill\n")

	res, err := Sync(Options{
		ProjectRoot: tmp,
		ConfigPath:  filepath.Join(stdai, "config.toml"),
		Version:     "test",
	})
	if err != nil {
		t.Fatalf("Sync: %v", err)
	}
	if res.Pruned != 1 {
		t.Errorf("Pruned = %d, want 1 (legacy cmd-ship), paths=%v", res.Pruned, res.PrunedPaths)
	}
	if _, err := os.Stat(filepath.Join(legacyDir, "SKILL.md")); err == nil {
		t.Error("legacy .agents/skills/cmd-ship/SKILL.md should be removed")
	}
	if _, err := os.Stat(filepath.Join(tmp, ".agents/skills/commands/ship/SKILL.md")); err != nil {
		t.Errorf("new command skill path missing: %v", err)
	}
}

// TestSyncPrunesLegacyCodexMemories：.codex/memories 废弃后（rules inline 到 AGENTS.md），
// 磁盘上未进 state 的旧产物也应被 prune；不带 stdagent marker 的用户文件不动。
func TestSyncPrunesLegacyCodexMemories(t *testing.T) {
	tmp := t.TempDir()
	stdai := filepath.Join(tmp, ".stdai")
	mustMkdir(t, filepath.Join(stdai, "standards/rules"))
	mustWrite(t, filepath.Join(stdai, "standards/rules", "style.md"), `---
type: rules
name: style
description: Code style
---
style body
`)
	mustWrite(t, filepath.Join(stdai, "config.toml"), `version = "1.0"
inject = false
backup = false
auto_pull = false

[targets]
codex = { enabled = true, convert = true }
`)
	legacyDir := filepath.Join(tmp, ".codex/memories")
	mustMkdir(t, filepath.Join(legacyDir, "references"))
	marker := "<!-- Generated by stdagent test. Do not edit by hand. -->\n"
	mustWrite(t, filepath.Join(legacyDir, "style.md"), marker+"old rule body\n")
	mustWrite(t, filepath.Join(legacyDir, "references", "design.md"), marker+"old ref body\n")
	mustWrite(t, filepath.Join(legacyDir, "user-note.md"), "my personal note, no marker\n")

	res, err := Sync(Options{
		ProjectRoot: tmp,
		ConfigPath:  filepath.Join(stdai, "config.toml"),
		Version:     "test",
	})
	if err != nil {
		t.Fatalf("Sync: %v", err)
	}
	if res.Pruned != 2 {
		t.Errorf("Pruned = %d, want 2 (style.md + references/design.md), paths=%v", res.Pruned, res.PrunedPaths)
	}
	for _, gone := range []string{"style.md", "references/design.md"} {
		if _, err := os.Stat(filepath.Join(legacyDir, gone)); err == nil {
			t.Errorf("legacy .codex/memories/%s should be removed", gone)
		}
	}
	if _, err := os.Stat(filepath.Join(legacyDir, "user-note.md")); err != nil {
		t.Errorf("user file without marker must be kept: %v", err)
	}
	agents, err := os.ReadFile(filepath.Join(tmp, "AGENTS.md"))
	if err != nil {
		t.Fatalf("AGENTS.md missing: %v", err)
	}
	if !strings.Contains(string(agents), "style body") {
		t.Errorf("AGENTS.md should inline rule body, got:\n%s", agents)
	}
}

func hasGitignoreLine(content, line string) bool {
	for _, l := range strings.Split(content, "\n") {
		if l == line {
			return true
		}
	}
	return false
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
