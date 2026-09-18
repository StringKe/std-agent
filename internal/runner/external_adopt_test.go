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
	if _, _, err := source.ImportExternal(tmp, false, source.ExternalScanOptions{}); err != nil {
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

// TestSyncBoostUpdateScenario boost:update 后的完整复现：同名技能覆盖、
// 自生成辅助文件、paths 规则、index 索引、根 mcp.json 同时出现时，
// 开着自动纳管的 strict sync 必须成功且行为保守。
func TestSyncBoostUpdateScenario(t *testing.T) {
	tmp := t.TempDir()
	stdai := filepath.Join(tmp, ".stdai")
	mustMkdir(t, filepath.Join(stdai, "standards/rules"))
	mustMkdir(t, filepath.Join(stdai, "standards/skills/pay/scripts"))
	mustWrite(t, filepath.Join(stdai, "standards/rules/base.md"), "---\ntype: rules\nname: base\n---\nBase rule.\n")
	mustWrite(t, filepath.Join(stdai, "standards/skills/pay/SKILL.md"), "---\ntype: skills\nname: pay\ndescription: Pay skill.\n---\nPay body.\n")
	mustWrite(t, filepath.Join(stdai, "standards/skills/pay/scripts/check.sh"), "#!/bin/sh\necho pay\n")
	mustWrite(t, filepath.Join(stdai, "config.toml"), `version = "1.0"
inject = false
backup = false
auto_pull = false

[targets]
claude-code = { enabled = true, convert = true }
codex = { enabled = true, convert = true }
`)

	// 首次同步：生成 .agents/skills/pay（含无 marker 辅助文件）等输出
	if _, err := Sync(Options{ProjectRoot: tmp, ConfigPath: filepath.Join(stdai, "config.toml"), Version: "test"}); err != nil {
		t.Fatalf("first Sync: %v", err)
	}

	// 模拟 boost:update：覆盖同名技能、新增辅助脚本、写入规则与 MCP
	if _, err := os.Stat(filepath.Join(tmp, ".agents/skills/pay/SKILL.md")); err != nil {
		t.Fatalf("generated skill missing: %v", err)
	}
	boostSkill := "---\nname: pay\ndescription: Boost pay skill.\n---\nBoost pay body.\n"
	mustWrite(t, filepath.Join(tmp, ".agents/skills/pay/SKILL.md"), boostSkill)
	mustWrite(t, filepath.Join(tmp, ".agents/skills/pay/scripts/boost-helper.sh"), "#!/bin/sh\necho boost\n")
	mustMkdir(t, filepath.Join(tmp, ".ai/rules"))
	mustWrite(t, filepath.Join(tmp, ".ai/rules/app.md"), "---\npaths:\n  - app/**\n---\nApp rule body.\n")
	mustWrite(t, filepath.Join(tmp, ".ai/rules/index.md"), "# Rules index\n")
	mustWrite(t, filepath.Join(tmp, ".mcp.json"), `{"mcpServers": {"laravel-boost": {"command": "php", "args": ["artisan", "boost:mcp"]}}}`)

	// 开着自动纳管的 strict 同步：不得报 output collision
	res, err := Sync(Options{ProjectRoot: tmp, ConfigPath: filepath.Join(stdai, "config.toml"), Version: "test"})
	if err != nil {
		t.Fatalf("boost-update Sync: %v", err)
	}
	if res.ExternalFiles != 1 {
		t.Errorf("only app.md should count as external, got %d", res.ExternalFiles)
	}

	// 1. 自生成文件不回流：check.sh（tracked）与 SKILL.md（同名 shadow）都不采用
	if _, serr := os.Stat(filepath.Join(stdai, "standards/external/skills/pay/scripts/check.sh")); !os.IsNotExist(serr) {
		t.Error("tracked aux file must not be re-adopted")
	}
	if _, serr := os.Stat(filepath.Join(stdai, "standards/external/skills/pay/SKILL.md")); !os.IsNotExist(serr) {
		t.Error("shadowed same-name skill must not be adopted")
	}
	// boost 新增的辅助脚本同属 shadow 包，一并跳过
	if _, serr := os.Stat(filepath.Join(stdai, "standards/external/skills/pay/scripts/boost-helper.sh")); !os.IsNotExist(serr) {
		t.Error("shadowed package aux file must not be adopted")
	}

	// 2. 规则范围保留：app.md 采用且带 applyTo；index.md 不是规则
	appRaw, err := os.ReadFile(filepath.Join(stdai, "standards/external/rules/app.md"))
	if err != nil {
		t.Fatalf("app.md not adopted: %v", err)
	}
	if !strings.Contains(string(appRaw), "applyTo:") || !strings.Contains(string(appRaw), "app/**") {
		t.Errorf("app.md must carry translated applyTo:\n%s", appRaw)
	}
	if _, serr := os.Stat(filepath.Join(stdai, "standards/external/rules/index.md")); !os.IsNotExist(serr) {
		t.Error("index.md must not be adopted as a rule")
	}

	// 3. mcp.json 无空 version
	mcpRaw, err := os.ReadFile(filepath.Join(stdai, "standards/mcp.json"))
	if err != nil {
		t.Fatalf("mcp.json missing: %v", err)
	}
	if strings.Contains(string(mcpRaw), `"version": ""`) {
		t.Errorf("empty version must be omitted:\n%s", mcpRaw)
	}
	if !strings.Contains(string(mcpRaw), "laravel-boost") {
		t.Errorf("boost server must be merged:\n%s", mcpRaw)
	}

	// 4. 用户源获胜：渲染输出仍是用户 skill 内容，而非 boost 覆盖
	for _, p := range []string{
		filepath.Join(tmp, ".claude/skills/pay/SKILL.md"),
		filepath.Join(tmp, ".agents/skills/pay/SKILL.md"),
	} {
		out, rerr := os.ReadFile(p)
		if rerr != nil {
			t.Fatalf("rendered skill missing %s: %v", p, rerr)
		}
		if !strings.Contains(string(out), "Pay body.") {
			t.Errorf("%s must render user source, got:\n%s", p, out)
		}
	}

	// 5. 二次同步稳定
	again, err := Sync(Options{ProjectRoot: tmp, ConfigPath: filepath.Join(stdai, "config.toml"), Version: "test"})
	if err != nil {
		t.Fatalf("repeat Sync: %v", err)
	}
	if again.Written != 0 {
		t.Errorf("repeat sync should be stable, wrote %d", again.Written)
	}
}

// TestSyncExternalUpdateFollowsUpstream 已纳管的外部技能被上游改写后，
// 下次 sync 把改动合进来，而不是静默丢掉或覆盖回旧版本。
func TestSyncExternalUpdateFollowsUpstream(t *testing.T) {
	tmp := t.TempDir()
	stdai := filepath.Join(tmp, ".stdai")
	mustMkdir(t, filepath.Join(stdai, "standards/rules"))
	mustWrite(t, filepath.Join(stdai, "standards/rules/base.md"), "---\ntype: rules\nname: base\n---\nBase rule.\n")
	mustWrite(t, filepath.Join(stdai, "config.toml"), `version = "1.0"
inject = false
backup = false
auto_pull = false

[targets]
codex = { enabled = true, convert = true }
`)

	// 上游技能首次出现：被纳管并渲染
	mustMkdir(t, filepath.Join(tmp, ".agents/skills/helper"))
	mustWrite(t, filepath.Join(tmp, ".agents/skills/helper/SKILL.md"), "---\nname: helper\ndescription: Helper v1\n---\nBody v1.\n")
	if _, err := Sync(Options{ProjectRoot: tmp, ConfigPath: filepath.Join(stdai, "config.toml"), Version: "test"}); err != nil {
		t.Fatalf("first Sync: %v", err)
	}
	src, err := os.ReadFile(filepath.Join(stdai, "standards/external/skills/helper/SKILL.md"))
	if err != nil || !strings.Contains(string(src), "Body v1.") {
		t.Fatalf("helper not adopted: %v\n%s", err, src)
	}

	// 上游改写（模拟 boost:update）：内容变化
	mustWrite(t, filepath.Join(tmp, ".agents/skills/helper/SKILL.md"), "---\nname: helper\ndescription: Helper v2\n---\nBody v2.\n")
	if _, err := Sync(Options{ProjectRoot: tmp, ConfigPath: filepath.Join(stdai, "config.toml"), Version: "test"}); err != nil {
		t.Fatalf("update Sync: %v", err)
	}

	// 源已跟进新版本
	src, err = os.ReadFile(filepath.Join(stdai, "standards/external/skills/helper/SKILL.md"))
	if err != nil || !strings.Contains(string(src), "Body v2.") {
		t.Errorf("adopted source must follow upstream update:\n%s", src)
	}
	// 渲染输出跟进，而非覆盖回旧版本
	out, err := os.ReadFile(filepath.Join(tmp, ".agents/skills/helper/SKILL.md"))
	if err != nil || !strings.Contains(string(out), "Body v2.") {
		t.Errorf("rendered output must follow upstream update:\n%s", out)
	}
	// 三次同步稳定（更新只合入一次）
	again, err := Sync(Options{ProjectRoot: tmp, ConfigPath: filepath.Join(stdai, "config.toml"), Version: "test"})
	if err != nil {
		t.Fatalf("repeat Sync: %v", err)
	}
	if again.Written != 0 {
		t.Errorf("repeat sync should be stable, wrote %d", again.Written)
	}
}
