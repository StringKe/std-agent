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

	files, warns, err := ScanExternal(root, ExternalScanOptions{})
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

	files, warns, err := ScanExternal(root, ExternalScanOptions{})
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

	files, _, err := ScanExternal(root, ExternalScanOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 || files[0].Path != "external/guidelines/real.md" {
		t.Errorf("should only adopt the user file, got %v", files)
	}
}

func TestScanExternalEmpty(t *testing.T) {
	files, warns, err := ScanExternal(t.TempDir(), ExternalScanOptions{})
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

	adopted, _, err := ImportExternal(root, false, ExternalScanOptions{})
	if err != nil {
		t.Fatalf("ImportExternal: %v", err)
	}
	if len(adopted) == 0 {
		t.Fatal("expected adopted files")
	}
	// 落盘文件与内存扫描字节一致（sync 去重可互换的前提）
	mem, _, err := ScanExternal(root, ExternalScanOptions{})
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
	adopted2, _, err := ImportExternal(root, false, ExternalScanOptions{})
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
	adopted, _, err := ImportExternal(root, true, ExternalScanOptions{})
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

// TestScanExternalSkipsTrackedOutputs 上次 sync 写出的文件不得再被采用
// （无 marker 的 skill 辅助文件也一样）。
func TestScanExternalSkipsTrackedOutputs(t *testing.T) {
	root := t.TempDir()
	skillMain := "---\nname: pay\ndescription: Pay\n---\nPay.\n"
	aux := "#!/bin/sh\necho pay\n"
	mkAll(t, filepath.Join(root, ".agents/skills/pay/scripts"))
	mustWrite(t, filepath.Join(root, ".agents/skills/pay/SKILL.md"), skillMain)
	mustWrite(t, filepath.Join(root, ".agents/skills/pay/scripts/check.sh"), aux)
	mkAll(t, filepath.Join(root, ".agents/skills/fresh"))
	mustWrite(t, filepath.Join(root, ".agents/skills/fresh/SKILL.md"), "---\nname: fresh\ndescription: Fresh\n---\nFresh.\n")

	opts := ExternalScanOptions{SkipPaths: map[string]string{
		".agents/skills/pay/SKILL.md":         sha256Hex([]byte(skillMain)),
		".agents/skills/pay/scripts/check.sh": sha256Hex([]byte(aux)),
	}}
	files, warns, err := ScanExternal(root, opts)
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range files {
		if strings.HasPrefix(f.Path, "external/skills/pay/") {
			t.Errorf("tracked output re-adopted: %s", f.Path)
		}
	}
	if len(files) != 1 || files[0].Path != "external/skills/fresh/SKILL.md" {
		t.Errorf("untracked skill must still be adopted, got %v", files)
	}
	joined := strings.Join(warns, "\n")
	if !strings.Contains(joined, "self-generated") {
		t.Errorf("expected self-generated summary warning, got %v", warns)
	}
}

// TestScanExternalReadoptsModifiedTrackedFile 记录 sha 与磁盘不一致时放行，
// 上游改写不被静默丢掉。
func TestScanExternalReadoptsModifiedTrackedFile(t *testing.T) {
	root := t.TempDir()
	mkAll(t, filepath.Join(root, ".agents/skills/pay"))
	mustWrite(t, filepath.Join(root, ".agents/skills/pay/SKILL.md"), "---\nname: pay\ndescription: New\n---\nNew body.\n")

	opts := ExternalScanOptions{SkipPaths: map[string]string{
		// 记录的是旧内容 sha，磁盘已被外部改写
		".agents/skills/pay/SKILL.md": sha256Hex([]byte("old content")),
	}}
	files, warns, err := ScanExternal(root, opts)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 || files[0].Path != "external/skills/pay/SKILL.md" {
		t.Fatalf("modified tracked file must be re-adopted, got %v", files)
	}
	if !strings.Contains(string(files[0].Raw), "New body.") {
		t.Errorf("re-adopted content must be the new body:\n%s", files[0].Raw)
	}
	if !strings.Contains(strings.Join(warns, "\n"), "changed outside stdagent") {
		t.Errorf("expected changed-outside warning, got %v", warns)
	}
}

// TestScanExternalReadoptsWholePackage 上游只改主文件时，sha 未变的附属文件仍随整包采用，
// 否则残包渲染后附属文件被当孤儿删除。
func TestScanExternalReadoptsWholePackage(t *testing.T) {
	root := t.TempDir()
	aux := "# checklist\n"
	mkAll(t, filepath.Join(root, ".agents/skills/pay/reference"))
	mustWrite(t, filepath.Join(root, ".agents/skills/pay/SKILL.md"), "---\nname: pay\ndescription: New\n---\nNew body.\n")
	mustWrite(t, filepath.Join(root, ".agents/skills/pay/reference/checklist.md"), aux)

	opts := ExternalScanOptions{SkipPaths: map[string]string{
		".agents/skills/pay/SKILL.md":               sha256Hex([]byte("old content")),
		".agents/skills/pay/reference/checklist.md": sha256Hex([]byte(aux)),
	}}
	files, _, err := ScanExternal(root, opts)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]string{}
	for _, f := range files {
		got[f.Path] = string(f.Raw)
	}
	for _, want := range []string{"external/skills/pay/SKILL.md", "external/skills/pay/reference/checklist.md"} {
		if _, ok := got[want]; !ok {
			t.Errorf("re-adopted package must keep %s, got %v", want, keys(got))
		}
	}
}

// TestScanExternalFullyTrackedShadowIsSilent 全包都是自生成物时，
// 即使包名被本地源占用也不报 shadow 噪音。
func TestScanExternalFullyTrackedShadowIsSilent(t *testing.T) {
	root := t.TempDir()
	skillMain := "---\nname: pay\ndescription: Pay\n---\nPay.\n"
	mkAll(t, filepath.Join(root, ".agents/skills/pay"))
	mustWrite(t, filepath.Join(root, ".agents/skills/pay/SKILL.md"), skillMain)

	opts := ExternalScanOptions{
		SkipPaths:  map[string]string{".agents/skills/pay/SKILL.md": sha256Hex([]byte(skillMain))},
		UserSkills: map[string]bool{"pay": true},
	}
	files, warns, err := ScanExternal(root, opts)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 0 {
		t.Errorf("fully tracked package must yield nothing, got %v", files)
	}
	joined := strings.Join(warns, "\n")
	if strings.Contains(joined, "shadows user skill") {
		t.Errorf("fully tracked package must not warn shadow, got %v", warns)
	}
	if !strings.Contains(joined, "self-generated") {
		t.Errorf("expected self-generated summary, got %v", warns)
	}
}

// TestScanExternalSkipsShadowedSkill 同名外部包让位于本地显式源
func TestScanExternalSkipsShadowedSkill(t *testing.T) {
	root := t.TempDir()
	mkAll(t, filepath.Join(root, ".agents/skills/pay"))
	mustWrite(t, filepath.Join(root, ".agents/skills/pay/SKILL.md"), "---\nname: pay\ndescription: Boost pay\n---\nBoost.\n")
	mkAll(t, filepath.Join(root, ".ai/skills"))
	mustWrite(t, filepath.Join(root, ".ai/skills/pay.md"), "---\nname: pay\ndescription: AI pay\n---\nAI.\n")
	mkAll(t, filepath.Join(root, ".agents/skills/fresh"))
	mustWrite(t, filepath.Join(root, ".agents/skills/fresh/SKILL.md"), "---\nname: fresh\ndescription: Fresh\n---\nFresh.\n")

	opts := ExternalScanOptions{UserSkills: map[string]bool{"pay": true}}
	files, warns, err := ScanExternal(root, opts)
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range files {
		if strings.Contains(f.Path, "/pay/") {
			t.Errorf("shadowed skill adopted: %s", f.Path)
		}
	}
	if len(files) != 1 || files[0].Path != "external/skills/fresh/SKILL.md" {
		t.Errorf("unshadowed skill must be adopted, got %v", files)
	}
	joined := strings.Join(warns, "\n")
	if !strings.Contains(joined, "shadows user skill") {
		t.Errorf("expected shadow warning, got %v", warns)
	}
}

// TestLocalSkillNames 收集目录名 + frontmatter name（kebab 归一）
func TestLocalSkillNames(t *testing.T) {
	root := t.TempDir()
	mkAll(t, filepath.Join(root, "skills/pay"))
	mustWrite(t, filepath.Join(root, "skills/pay/SKILL.md"), "---\nname: Pay_Rules\ndescription: x\n---\nBody.\n")
	mkAll(t, filepath.Join(root, "skills/plain"))
	mustWrite(t, filepath.Join(root, "skills/plain/SKILL.md"), "no frontmatter\n")

	got := LocalSkillNames(root)
	for _, want := range []string{"pay", "pay-rules", "plain"} {
		if !got[want] {
			t.Errorf("missing skill name %q (got %v)", want, got)
		}
	}
	if len(LocalSkillNames(filepath.Join(root, "missing"))) != 0 {
		t.Error("missing dir should yield empty set")
	}
}

// TestScanExternalSkipsIndex .ai/rules/index.md 是发现索引，不是规则
func TestScanExternalSkipsIndex(t *testing.T) {
	root := t.TempDir()
	mkAll(t, filepath.Join(root, ".ai/rules"))
	mustWrite(t, filepath.Join(root, ".ai/rules/app.md"), "---\npaths:\n  - app/**\n---\nApp.\n")
	mustWrite(t, filepath.Join(root, ".ai/rules/index.md"), "# index\n")

	files, warns, err := ScanExternal(root, ExternalScanOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 || files[0].Path != "external/rules/app.md" {
		t.Fatalf("want only app.md adopted, got %v", files)
	}
	joined := strings.Join(warns, "\n")
	if !strings.Contains(joined, "discovery index") {
		t.Errorf("expected index warning, got %v", warns)
	}
	// paths: 转译为 applyTo，经 parser 生效
	d, err := parser.Parse(files[0].Path, files[0].Raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(d.ApplyTo) != 1 || d.ApplyTo[0] != "app/**" {
		t.Errorf("paths should translate to ApplyTo, got %v", d.ApplyTo)
	}
	if !strings.Contains(string(files[0].Raw), "paths:") {
		t.Error("original paths: key must be preserved")
	}
}

// TestTranslateRulesPaths 转译表
func TestTranslateRulesPaths(t *testing.T) {
	cases := []struct {
		name     string
		front    string
		contains []string
		absent   []string
	}{
		{"list", "paths:\n  - app/**\n  - web/**\n", []string{"applyTo:\n  - app/**\n  - web/**\n"}, nil},
		{"scalar", "title: x\npaths: app/**\n", []string{"applyTo: app/**"}, nil},
		{"has applyTo", "paths:\n  - a/**\napplyTo:\n  - b/**\n", []string{"- b/**"}, []string{"applyTo:\n  - a/**"}},
		{"has globs", "paths:\n  - a/**\nglobs:\n  - b/**\n", nil, []string{"applyTo:"}},
		{"no paths", "title: x\n", nil, []string{"applyTo:"}},
		{"block scalar", "paths: |\n  multi\n  line\n", nil, []string{"applyTo:"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := string(translateRulesPaths([]byte(c.front)))
			for _, want := range c.contains {
				if !strings.Contains(got, want) {
					t.Errorf("missing %q in:\n%s", want, got)
				}
			}
			for _, no := range c.absent {
				if strings.Contains(got, no) {
					t.Errorf("must not contain %q in:\n%s", no, got)
				}
			}
		})
	}
}

// TestImportExternalOmitsEmptyMCPVersion 空 version 不得序列化
func TestImportExternalOmitsEmptyMCPVersion(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, ".mcp.json"), `{"mcpServers": {"b": {"command": "b-bin"}}}`)

	if _, _, err := ImportExternal(root, false, ExternalScanOptions{}); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(root, ".stdai", "standards", "mcp.json"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), `"version"`) {
		t.Errorf("empty version must be omitted:\n%s", raw)
	}
	// 预设版本不受合并影响
	mkAll(t, filepath.Join(root, ".stdai", "standards"))
	mustWrite(t, filepath.Join(root, ".stdai", "standards", "mcp.json"), `{"version": "1.0", "servers": {}}`)
	if _, _, err := ImportExternal(root, false, ExternalScanOptions{}); err != nil {
		t.Fatal(err)
	}
	raw, _ = os.ReadFile(filepath.Join(root, ".stdai", "standards", "mcp.json"))
	if !strings.Contains(string(raw), `"version": "1.0"`) {
		t.Errorf("preset version must survive:\n%s", raw)
	}
	// 旧版空 version 一次性规范化：去掉该键，server 保留，二次 import 稳定
	mustWrite(t, filepath.Join(root, ".stdai", "standards", "mcp.json"), `{"version": "", "servers": {"b": {"command": "b-bin"}}}`)
	if _, _, err := ImportExternal(root, false, ExternalScanOptions{}); err != nil {
		t.Fatal(err)
	}
	raw, _ = os.ReadFile(filepath.Join(root, ".stdai", "standards", "mcp.json"))
	if strings.Contains(string(raw), `"version"`) {
		t.Errorf("stale empty version must be normalized away:\n%s", raw)
	}
	if !strings.Contains(string(raw), `"b-bin"`) {
		t.Errorf("servers must survive normalization:\n%s", raw)
	}
	if _, _, err := ImportExternal(root, false, ExternalScanOptions{}); err != nil {
		t.Fatal(err)
	}
	raw2, _ := os.ReadFile(filepath.Join(root, ".stdai", "standards", "mcp.json"))
	if string(raw2) != string(raw) {
		t.Error("normalization must be one-shot: content changed again")
	}
}
