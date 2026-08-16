package writer

import (
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/StringKe/std-agent/internal/config"
)

const (
	gitignoreBegin = "# BEGIN stdagent"
	gitignoreEnd   = "# END stdagent"
	gitignoreNote  = "# 由 stdagent sync/init 维护。块外条目不会改动。"
)

// alwaysIgnore 是任何 generated/portable 模式下都忽略的运行时与本机文件。
var alwaysIgnore = []string{
	".stdai/cache/",
	".stdai/backups/",
	".stdai/logs/",
	".stdai/state.json",
	"CLAUDE.local.md",
	"AGENTS.local.md",
	"AGENTS.override.md",
	".claude/settings.local.json",
	".kimi-code/local.toml",
	".qwen/QWEN.local.md",
}

// targetIgnorePrefixes 是各 target 可重建产物的 gitignore 前缀。
// crush.json / kilo.jsonc 是用户配置合并目标，不列入。
var targetIgnorePrefixes = map[string][]string{
	"claude-code":  {"CLAUDE.md", ".claude/", ".mcp.json"},
	"codex":        {"AGENTS.md", ".agents/", ".codex/"},
	"cursor":       {".cursor/"},
	"copilot":      {".github/copilot-instructions.md", ".github/instructions/", ".github/prompts/", ".github/agents/", ".github/skills/", ".vscode/mcp.json"},
	"windsurf":     {".windsurf/", ".devin/"},
	"gemini":       {"GEMINI.md", ".gemini/"},
	"aider":        {"AGENTS.md"},
	"cline":        {".clinerules/", ".cline/"},
	"opencode":     {".opencode/"},
	"roo-code":     {".roo/"},
	"crush":        {"CRUSH.md", ".crush/"},
	"amp":          {"AGENTS.md", ".agents/", ".amp/"},
	"warp":         {"AGENTS.md", ".agents/", ".warp/"},
	"factory":      {"AGENTS.md", ".factory/"},
	"continue-dev": {".continue/"},
	"antigravity":  {"AGENTS.md", ".agents/"},
	"qwen-code":    {"QWEN.md", ".qwen/"},
	"pi":           {"AGENTS.md", ".pi/"},
	"kilo-code":    {".kilo/"},
	"augment-code": {".augment/"},
	"jules":        {"AGENTS.md", ".jules/"},
	"grok-build":   {"AGENTS.md", ".grok/"},
	"kimi-code":    {"AGENTS.md", ".agents/", ".kimi-code/"},
	"kiro":         {"AGENTS.md", ".kiro/"},
	"goose":        {"AGENTS.md", ".agents/", ".goose/"},
}

// GitignoreEntries 按模式和启用 target（再并入本次 plan 路径）生成忽略列表。
func GitignoreEntries(mode string, enabled []string, plans []*Plan) []string {
	mode = config.NormalizeGitignore(mode)
	if mode == config.GitignoreOff {
		return nil
	}
	seen := map[string]struct{}{}
	var out []string
	add := func(p string) {
		if p == "" {
			return
		}
		if mode == config.GitignorePortable && keepPortable(p) {
			return
		}
		if _, ok := seen[p]; ok {
			return
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}
	for _, p := range alwaysIgnore {
		add(p)
	}
	for _, name := range enabled {
		for _, p := range targetIgnorePrefixes[name] {
			add(p)
		}
	}
	for _, plan := range plans {
		for _, f := range plan.Files {
			if f.JSONMerge {
				continue
			}
			add(collapseIgnorePath(f.Path))
		}
	}
	sort.Strings(out)
	return out
}

func keepPortable(p string) bool {
	p = strings.TrimPrefix(p, "/")
	if p == "AGENTS.md" {
		return true
	}
	return p == ".agents" || p == ".agents/" || strings.HasPrefix(p, ".agents/")
}

func collapseIgnorePath(p string) string {
	p = strings.TrimPrefix(p, "./")
	p = strings.ReplaceAll(p, "\\", "/")
	parts := strings.Split(p, "/")
	if len(parts) == 1 {
		return parts[0]
	}
	switch parts[0] {
	case ".github":
		if parts[1] == "copilot-instructions.md" {
			return ".github/copilot-instructions.md"
		}
		return ".github/" + parts[1] + "/"
	case ".vscode":
		return path.Join(parts...)
	}
	if strings.HasPrefix(parts[0], ".") {
		return parts[0] + "/"
	}
	return path.Base(p)
}

// UpsertGitignore 更新根 .gitignore 中的 stdagent 块。dryRun 只比较不写盘。
func UpsertGitignore(root string, entries []string, dryRun bool) (bool, error) {
	if len(entries) == 0 {
		return false, nil
	}
	gi := filepath.Join(root, ".gitignore")
	existing, err := os.ReadFile(gi) //nolint:gosec
	if err != nil && !os.IsNotExist(err) {
		return false, err
	}
	next := replaceManagedBlock(string(existing), entries)
	if next == string(existing) {
		return false, nil
	}
	if dryRun {
		return true, nil
	}
	if !strings.HasSuffix(next, "\n") {
		next += "\n"
	}
	return true, os.WriteFile(gi, []byte(next), 0o600) //nolint:gosec // 路径固定为项目根 .gitignore
}

func replaceManagedBlock(content string, entries []string) string {
	block := buildManagedBlock(entries)
	start := strings.Index(content, gitignoreBegin)
	end := strings.Index(content, gitignoreEnd)
	if start >= 0 && end > start {
		end += len(gitignoreEnd)
		if end < len(content) && content[end] == '\n' {
			end++
		}
		before := strings.TrimRight(content[:start], "\n")
		after := strings.TrimLeft(content[end:], "\n")
		var b strings.Builder
		if before != "" {
			b.WriteString(before)
			b.WriteString("\n\n")
		}
		b.WriteString(block)
		b.WriteByte('\n')
		if after != "" {
			b.WriteString(after)
			if !strings.HasSuffix(after, "\n") {
				b.WriteByte('\n')
			}
		}
		return b.String()
	}
	cleaned := stripDuplicateLines(content, entries)
	if cleaned != "" && !strings.HasSuffix(cleaned, "\n") {
		cleaned += "\n"
	}
	if cleaned == "" {
		return block + "\n"
	}
	return cleaned + "\n" + block + "\n"
}

func buildManagedBlock(entries []string) string {
	var b strings.Builder
	b.WriteString(gitignoreBegin)
	b.WriteByte('\n')
	b.WriteString(gitignoreNote)
	b.WriteByte('\n')
	for _, e := range entries {
		b.WriteString(e)
		b.WriteByte('\n')
	}
	b.WriteString(gitignoreEnd)
	return b.String()
}

func stripDuplicateLines(content string, entries []string) string {
	if content == "" {
		return ""
	}
	drop := map[string]struct{}{}
	for _, e := range entries {
		drop[e] = struct{}{}
	}
	var kept []string
	for _, line := range strings.Split(content, "\n") {
		trim := strings.TrimSpace(line)
		if _, ok := drop[trim]; ok {
			continue
		}
		kept = append(kept, line)
	}
	return strings.TrimRight(strings.Join(kept, "\n"), "\n")
}
