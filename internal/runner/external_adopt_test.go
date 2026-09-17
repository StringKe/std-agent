package runner

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/StringKe/std-agent/internal/source"
)

func writeExternalConfig(t *testing.T, stdai string) {
	t.Helper()
	mustWrite(t, filepath.Join(stdai, "config.toml"), `version = "1.0"
inject = false
backup = false
auto_pull = false

[targets]
claude-code = { enabled = true, convert = true }
`)
}

// TestSyncAdoptsExternalByDefault 默认 sync 扫描 .ai/* + .agents/skills
func TestSyncAdoptsExternalByDefault(t *testing.T) {
	tmp := t.TempDir()
	stdai := filepath.Join(tmp, ".stdai")
	mustMkdir(t, filepath.Join(stdai, "standards/rules"))
	mustWrite(t, filepath.Join(stdai, "standards/rules/base.md"), "---\ntype: rules\nname: base\n---\nBase rule.\n")
	writeExternalConfig(t, stdai)

	mustMkdir(t, filepath.Join(tmp, ".ai/guidelines"))
	mustWrite(t, filepath.Join(tmp, ".ai/guidelines/ext.md"), "External guideline body.\n")
	mustMkdir(t, filepath.Join(tmp, ".agents/skills/helper"))
	mustWrite(t, filepath.Join(tmp, ".agents/skills/helper/SKILL.md"), "---\nname: helper\ndescription: Helper skill\n---\nHelper body.\n")

	res, err := Sync(Options{ProjectRoot: tmp, ConfigPath: filepath.Join(stdai, "config.toml"), Version: "test"})
	if err != nil {
		t.Fatalf("Sync: %v", err)
	}
	if res.ExternalFiles != 2 {
		t.Errorf("ExternalFiles = %d, want 2", res.ExternalFiles)
	}
	ruleOut, err := os.ReadFile(filepath.Join(tmp, ".claude/rules/ext.md"))
	if err != nil {
		t.Fatalf("adopted rule output: %v", err)
	}
	if !strings.Contains(string(ruleOut), "External guideline body.") {
		t.Errorf(".claude/rules/ext.md missing adopted body:\n%s", ruleOut)
	}
	if _, err := os.Stat(filepath.Join(tmp, ".claude/skills/helper/SKILL.md")); err != nil {
		t.Errorf("adopted skill not emitted: %v", err)
	}
	// sync 默认自动落盘 external/（auto-adopt），二次 sync 幂等
	if _, err := os.Stat(filepath.Join(stdai, "standards/external/guidelines/ext.md")); err != nil {
		t.Errorf("sync should materialize external adopted files: %v", err)
	}
	resAgain, err := Sync(Options{ProjectRoot: tmp, ConfigPath: filepath.Join(stdai, "config.toml"), Version: "test"})
	if err != nil {
		t.Fatalf("repeat Sync: %v", err)
	}
	if resAgain.ExternalFiles != res.ExternalFiles {
		t.Errorf("ExternalFiles unstable: %d -> %d", res.ExternalFiles, resAgain.ExternalFiles)
	}
	if resAgain.Written != 0 {
		t.Errorf("repeat sync should be stable, wrote %d files", resAgain.Written)
	}
}

// TestSyncNoExternalOptsOut --no-external 跳过扫描
func TestSyncNoExternalOptsOut(t *testing.T) {
	tmp := t.TempDir()
	stdai := filepath.Join(tmp, ".stdai")
	mustMkdir(t, filepath.Join(stdai, "standards/rules"))
	mustWrite(t, filepath.Join(stdai, "standards/rules/base.md"), "---\ntype: rules\nname: base\n---\nBase rule.\n")
	writeExternalConfig(t, stdai)

	mustMkdir(t, filepath.Join(tmp, ".ai/guidelines"))
	mustWrite(t, filepath.Join(tmp, ".ai/guidelines/ext.md"), "External guideline body.\n")

	res, err := Sync(Options{ProjectRoot: tmp, ConfigPath: filepath.Join(stdai, "config.toml"), NoExternal: true, Version: "test"})
	if err != nil {
		t.Fatalf("Sync: %v", err)
	}
	if res.ExternalFiles != 0 {
		t.Errorf("ExternalFiles = %d, want 0", res.ExternalFiles)
	}
	claude, _ := os.ReadFile(filepath.Join(tmp, "CLAUDE.md"))
	if strings.Contains(string(claude), "External guideline body.") {
		t.Errorf("NoExternal sync must not include external content:\n%s", claude)
	}
}

// TestSyncExternalConfigOptOut [external] enabled=false 关闭默认扫描（老配置缺表时仍启用）
func TestSyncExternalConfigOptOut(t *testing.T) {
	tmp := t.TempDir()
	stdai := filepath.Join(tmp, ".stdai")
	mustMkdir(t, filepath.Join(stdai, "standards/rules"))
	mustWrite(t, filepath.Join(stdai, "standards/rules/base.md"), "---\ntype: rules\nname: base\n---\nBase rule.\n")
	mustWrite(t, filepath.Join(stdai, "config.toml"), `version = "1.0"
inject = false
backup = false
auto_pull = false

[targets]
claude-code = { enabled = true, convert = true }

[external]
enabled = false
`)
	mustMkdir(t, filepath.Join(tmp, ".ai/guidelines"))
	mustWrite(t, filepath.Join(tmp, ".ai/guidelines/ext.md"), "External guideline body.\n")

	res, err := Sync(Options{ProjectRoot: tmp, ConfigPath: filepath.Join(stdai, "config.toml"), Version: "test"})
	if err != nil {
		t.Fatalf("Sync: %v", err)
	}
	if res.ExternalFiles != 0 {
		t.Errorf("ExternalFiles = %d, want 0 with [external] disabled", res.ExternalFiles)
	}
}

// TestSyncExternalDeduplicatesImported import 落盘后 sync 不重复计数
func TestSyncExternalDeduplicatesImported(t *testing.T) {
	tmp := t.TempDir()
	stdai := filepath.Join(tmp, ".stdai")
	mustMkdir(t, filepath.Join(stdai, "standards/rules"))
	mustWrite(t, filepath.Join(stdai, "standards/rules/base.md"), "---\ntype: rules\nname: base\n---\nBase rule.\n")
	writeExternalConfig(t, stdai)

	mustMkdir(t, filepath.Join(tmp, ".ai/guidelines"))
	mustWrite(t, filepath.Join(tmp, ".ai/guidelines/ext.md"), "External guideline body.\n")

	first, err := Sync(Options{ProjectRoot: tmp, ConfigPath: filepath.Join(stdai, "config.toml"), Version: "test"})
	if err != nil {
		t.Fatalf("first Sync: %v", err)
	}
	if first.ExternalFiles != 1 {
		t.Fatalf("ExternalFiles = %d, want 1", first.ExternalFiles)
	}
	firstClaude, _ := os.ReadFile(filepath.Join(tmp, "CLAUDE.md"))

	// import 落盘（与内存扫描字节一致），再 sync：输出字节必须一致且不重复
	if _, _, err := source.ImportExternal(tmp, false); err != nil {
		t.Fatalf("import: %v", err)
	}
	second, err := Sync(Options{ProjectRoot: tmp, ConfigPath: filepath.Join(stdai, "config.toml"), Version: "test"})
	if err != nil {
		t.Fatalf("second Sync: %v", err)
	}
	if second.Docs != first.Docs {
		t.Errorf("Docs changed after import: %d -> %d (duplicate?)", first.Docs, second.Docs)
	}
	secondClaude, _ := os.ReadFile(filepath.Join(tmp, "CLAUDE.md"))
	if string(firstClaude) != string(secondClaude) {
		t.Errorf("shared output changed after import materialization:\n--- before ---\n%s\n--- after ---\n%s", firstClaude, secondClaude)
	}
}

// TestSyncAdoptsRootMCP 根 .mcp.json 仅补缺失 servers
func TestSyncAdoptsRootMCP(t *testing.T) {
	tmp := t.TempDir()
	stdai := filepath.Join(tmp, ".stdai")
	mustMkdir(t, filepath.Join(stdai, "standards/rules"))
	mustWrite(t, filepath.Join(stdai, "standards/rules/base.md"), "---\ntype: rules\nname: base\n---\nBase rule.\n")
	mustWrite(t, filepath.Join(stdai, "standards/mcp.json"), `{"version": "1.0", "servers": {"keep": {"type": "stdio", "command": "keep-bin"}}}`)
	writeExternalConfig(t, stdai)
	mustWrite(t, filepath.Join(tmp, ".mcp.json"), `{"mcpServers": {"keep": {"command": "evil-bin"}, "fresh": {"command": "fresh-bin"}}}`)

	if _, err := Sync(Options{ProjectRoot: tmp, ConfigPath: filepath.Join(stdai, "config.toml"), Version: "test"}); err != nil {
		t.Fatalf("Sync: %v", err)
	}
	// sync 自动把根 .mcp.json 缺失项合并进 standards/mcp.json，既有项不覆盖
	mcpRaw, _ := os.ReadFile(filepath.Join(stdai, "standards/mcp.json"))
	if !strings.Contains(string(mcpRaw), "fresh") {
		t.Errorf("sync should merge missing servers to disk mcp.json, got:\n%s", mcpRaw)
	}
	if strings.Contains(string(mcpRaw), "evil-bin") {
		t.Errorf("existing server must not be overwritten:\n%s", mcpRaw)
	}
}
