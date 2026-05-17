package parser

import (
	"os"
	"path/filepath"
	"testing"
)

// TestReferencesParseSmoke 验证 references 类源文件 parser 行为（v0.0.4 Phase 0.3 验收）
//
// 现有 parser 应已支持 type: references（document.go IsValidType 含 TypeReferences），
// 通用字段 description / applyTo 走 rules 同一路径。本 test 用本仓自举源文件做端到端验证。
func TestReferencesParseSmoke(t *testing.T) {
	// 找到仓库根（向上 walk 直到看到 .stdai/）
	wd, _ := os.Getwd()
	root := wd
	for i := 0; i < 5; i++ {
		if _, err := os.Stat(filepath.Join(root, ".stdai/standards")); err == nil {
			break
		}
		root = filepath.Dir(root)
	}
	src := filepath.Join(root, ".stdai/standards/references/transformer-design.md")
	raw, err := os.ReadFile(src) //nolint:gosec
	if err != nil {
		t.Skipf("fixture not found (likely running outside repo): %v", err)
	}
	d, err := Parse("references/transformer-design.md", raw)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if d.Type != TypeReferences {
		t.Errorf("Type = %q, want references", d.Type)
	}
	if d.Name != "transformer-design" {
		t.Errorf("Name = %q", d.Name)
	}
	if d.Description == "" {
		t.Error("Description empty")
	}
	if len(d.ApplyTo) != 2 {
		t.Errorf("ApplyTo len = %d, want 2 (internal/transformer/**/*.go + internal/runner/runner.go)", len(d.ApplyTo))
	}
	if d.BodyBytes == 0 {
		t.Error("BodyBytes 0, body should have content")
	}
}
