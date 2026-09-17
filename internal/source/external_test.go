package source

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/StringKe/std-agent/internal/config"
	"github.com/StringKe/std-agent/internal/parser"
)

// fixture 覆盖四类外部产物 + 根 .mcp.json
func setupExternalFixture(t *testing.T, root string) {
	t.Helper()
	mkAll(t, filepath.Join(root, ".ai/guidelines"))
	mkAll(t, filepath.Join(root, ".ai/rules"))
	mkAll(t, filepath.Join(root, ".ai/skills/plan-pack/references"))
	mkAll(t, filepath.Join(root, ".agents/skills/agent-pack"))

	// 1. guidelines：无 frontmatter -> 合成 rules
	mustWrite(t, filepath.Join(root, ".ai/guidelines/style.md"), "# Style\n\nUse clear names.\n")
	// 2. rules：自带 globs frontmatter -> 保留
	mustWrite(t, filepath.Join(root, ".ai/rules/api.md"), "---\nglobs:\n  - \"src/**/*.ts\"\n---\nAPI rules.\n")
	// 3. 单文件 skill -> 合成 SKILL.md
	mustWrite(t, filepath.Join(root, ".ai/skills/quick.md"), "# Quick\n\nDo it fast.\n")
	// 4. 整包 skill：SKILL.md 有 frontmatter（注入 provenance），辅助文件原样
	mustWrite(t, filepath.Join(root, ".ai/skills/plan-pack/SKILL.md"), "---\nname: plan-pack\ndescription: Plan skill\n---\nPlan body.\n")
	mustWrite(t, filepath.Join(root, ".ai/skills/plan-pack/references/check.md"), "checklist\n")
	mustWrite(t, filepath.Join(root, ".ai/skills/plan-pack/run.sh"), "#!/bin/sh\necho hi\n")
	// 5. .agents/skills 整包
	mustWrite(t, filepath.Join(root, ".agents/skills/agent-pack/SKILL.md"), "Agent body without frontmatter.\n")
	// 根 .mcp.json（Claude 形状）
	mustWrite(t, filepath.Join(root, ".mcp.json"), `{"mcpServers": {"gh": {"command": "gh-mcp", "args": ["serve"], "env": {"K": "V"}}}}`)
}

func TestScanExternalMappings(t *testing.T) {
	root := t.TempDir()
	setupExternalFixture(t, root)

	files, warns, err := ScanExternal(root)
	if err != nil {
		t.Fatalf("ScanExternal: %v", err)
	}
	if len(warns) != 0 {
		t.Errorf("unexpected warnings: %v", warns)
	}
	got := map[string]string{}
	for _, f := range files {
		got[f.Path] = string(f.Raw)
	}
	for _, want := range []string{
		"external/guidelines/style.md",
		"external/rules/api.md",
		"external/skills/quick/SKILL.md",
		"external/skills/plan-pack/SKILL.md",
		"external/skills/plan-pack/references/check.md",
		"external/skills/plan-pack/run.sh",
		"external/skills/agent-pack/SKILL.md",
	} {
		if _, ok := got[want]; !ok {
			t.Errorf("missing adopted file %s (got %v)", want, keys(got))
		}
	}

	// 合成 frontmatter：guidelines 无 frontmatter -> type rules + provenance
	if s := got["external/guidelines/style.md"]; !strings.Contains(s, "type: rules") ||
		!strings.Contains(s, "external_source: .ai/guidelines") ||
		!strings.Contains(s, "external_path: style.md") ||
		!strings.Contains(s, "Use clear names.") {
		t.Errorf("guidelines adopt wrong:\n%s", s)
	}
	// 原有 frontmatter 逐字节保留（globs 行仍在），仅追加缺失键 + provenance
	if s := got["external/rules/api.md"]; !strings.Contains(s, "globs:") ||
		!strings.Contains(s, "type: rules") ||
		!strings.Contains(s, "external_path: api.md") {
		t.Errorf("rules adopt wrong:\n%s", s)
	}
	// 单文件 skill 合成 SKILL.md
	if s := got["external/skills/quick/SKILL.md"]; !strings.Contains(s, "type: skills") ||
		!strings.Contains(s, "name: quick") {
		t.Errorf("single-file skill adopt wrong:\n%s", s)
	}
	// 整包 SKILL.md 原有 name/description 保留
	if s := got["external/skills/plan-pack/SKILL.md"]; !strings.Contains(s, "name: plan-pack") ||
		!strings.Contains(s, "description: Plan skill") ||
		!strings.Contains(s, "external_source: .ai/skills") {
		t.Errorf("packaged skill adopt wrong:\n%s", s)
	}
	// 非 markdown 辅助文件逐字节原样
	if s := got["external/skills/plan-pack/run.sh"]; s != "#!/bin/sh\necho hi\n" {
		t.Errorf("aux file should be verbatim, got %q", s)
	}

	// 采用产物必须能被 parser 正常解析
	for _, f := range files {
		if !strings.HasSuffix(f.Path, ".md") {
			continue
		}
		d, err := parser.Parse(f.Path, f.Raw)
		if err != nil {
			t.Errorf("Parse(%s): %v", f.Path, err)
			continue
		}
		if strings.HasPrefix(f.Path, "external/skills/") && strings.HasSuffix(f.Path, "SKILL.md") {
			if d.Type != parser.TypeSkills {
				t.Errorf("%s type = %q, want skills", f.Path, d.Type)
			}
		}
	}
	// globs -> ApplyTo 经 parser 生效
	d, err := parser.Parse("external/rules/api.md", []byte(got["external/rules/api.md"]))
	if err != nil {
		t.Fatal(err)
	}
	if len(d.ApplyTo) != 1 || d.ApplyTo[0] != "src/**/*.ts" {
		t.Errorf("globs should merge to ApplyTo, got %v", d.ApplyTo)
	}
}

func TestScanExternalAgentsConflictSkipped(t *testing.T) {
	root := t.TempDir()
	mkAll(t, filepath.Join(root, ".ai/skills/dup"))
	mkAll(t, filepath.Join(root, ".agents/skills/dup"))
	mustWrite(t, filepath.Join(root, ".ai/skills/dup/SKILL.md"), "---\nname: dup\ndescription: ai wins\n---\nai\n")
	mustWrite(t, filepath.Join(root, ".agents/skills/dup/SKILL.md"), "---\nname: dup\ndescription: agents loses\n---\nagents\n")

	files, warns, err := ScanExternal(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 {
		t.Fatalf("want 1 file (.ai wins), got %d", len(files))
	}
	if !strings.Contains(string(files[0].Raw), "ai\n") {
		t.Errorf("expected .ai content to win, got:\n%s", files[0].Raw)
	}
	if len(warns) != 1 {
		t.Errorf("expected 1 conflict warning, got %v", warns)
	}
}

// TestScanExternalSkipsGenerated 上次 sync 的生成物（带 marker）不得再被采用
func TestScanExternalSkipsGenerated(t *testing.T) {
	root := t.TempDir()
	mkAll(t, filepath.Join(root, ".agents/skills/code-review"))
	mustWrite(t, filepath.Join(root, ".agents/skills/code-review/SKILL.md"),
		"---\nname: code-review\ndescription: X\n---\nbody\n<!-- Generated by stdagent test -->\n")
	mkAll(t, filepath.Join(root, ".ai/guidelines"))
	mustWrite(t, filepath.Join(root, ".ai/guidelines/real.md"), "real user content\n")

	files, _, err := ScanExternal(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 || files[0].Path != "external/guidelines/real.md" {
		t.Errorf("should only adopt the user file, got %v", files)
	}
}

func TestScanExternalEmpty(t *testing.T) {
	files, warns, err := ScanExternal(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 0 || len(warns) != 0 {
		t.Errorf("empty root should yield nothing, got %d files %v", len(files), warns)
	}
}

func TestAdoptRootMCP(t *testing.T) {
	root := t.TempDir()
	setupExternalFixture(t, root)

	mcp := &config.MCPConfig{Servers: map[string]config.MCPServer{
		"gh": {Type: "stdio", Command: "other"},
	}}
	added, warns := AdoptRootMCP(root, mcp)
	if len(warns) != 0 {
		t.Errorf("warns: %v", warns)
	}
	if added != 0 {
		t.Errorf("existing server must not be overwritten, added=%d", added)
	}
	if mcp.Servers["gh"].Command != "other" {
		t.Errorf("existing server overwritten: %+v", mcp.Servers["gh"])
	}

	fresh := &config.MCPConfig{}
	added, _ = AdoptRootMCP(root, fresh)
	if added != 1 {
		t.Fatalf("added = %d, want 1", added)
	}
	srv := fresh.Servers["gh"]
	if srv.Type != "stdio" || srv.Command != "gh-mcp" || len(srv.Args) != 1 || srv.Env["K"] != "V" {
		t.Errorf("converted server wrong: %+v", srv)
	}
}

func TestImportExternalRoundTrip(t *testing.T) {
	root := t.TempDir()
	setupExternalFixture(t, root)

	adopted, _, err := ImportExternal(root, false)
	if err != nil {
		t.Fatalf("ImportExternal: %v", err)
	}
	if len(adopted) == 0 {
		t.Fatal("expected adopted files")
	}
	// 落盘文件与内存扫描字节一致（sync 去重可互换的前提）
	mem, _, err := ScanExternal(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range mem {
		disk, rerr := os.ReadFile(filepath.Join(root, ".stdai", "standards", filepath.FromSlash(f.Path)))
		if rerr != nil {
			t.Errorf("imported file missing on disk: %s", f.Path)
			continue
		}
		if string(disk) != string(f.Raw) {
			t.Errorf("%s disk/memory mismatch", f.Path)
		}
	}
	// mcp.json 合并落盘
	mcpRaw, rerr := os.ReadFile(filepath.Join(root, ".stdai", "standards", "mcp.json"))
	if rerr != nil {
		t.Fatalf("mcp.json not written: %v", rerr)
	}
	if !strings.Contains(string(mcpRaw), `"gh"`) {
		t.Errorf("mcp.json missing gh server:\n%s", mcpRaw)
	}
	// 二次 import 幂等：无新写入
	adopted2, _, err := ImportExternal(root, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(adopted2) != 0 {
		t.Errorf("second import should be no-op, got %v", adopted2)
	}
}

func TestImportExternalDryRun(t *testing.T) {
	root := t.TempDir()
	setupExternalFixture(t, root)
	adopted, _, err := ImportExternal(root, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(adopted) == 0 {
		t.Error("dry-run should still report would-adopt files")
	}
	if _, err := os.Stat(filepath.Join(root, ".stdai")); !os.IsNotExist(err) {
		t.Error("dry-run must not write .stdai")
	}
}

func keys(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
