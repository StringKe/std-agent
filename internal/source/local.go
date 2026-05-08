package source

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Local 是本地 standards/ 目录源
type Local struct {
	Root string
}

// NewLocal 创建本地源
func NewLocal(root string) *Local {
	return &Local{Root: root}
}

// Name 返回 "local"
func (l *Local) Name() string { return "local" }

// Files 扫描 root 下所有 .md / .markdown 文件，并额外收集 skills/<name>/
// 目录下所有非 .md 辅助文件（scripts/ references/ assets/ 等），用于
// SKILL package 子目录复制
func (l *Local) Files() ([]File, error) {
	var out []File
	err := filepath.WalkDir(l.Root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(l.Root, path)
		rel = filepath.ToSlash(rel)
		if !shouldInclude(rel, d.Name()) {
			return nil
		}
		raw, rerr := os.ReadFile(path) //nolint:gosec
		if rerr != nil {
			return fmt.Errorf("read %s: %w", path, rerr)
		}
		out = append(out, File{Path: rel, Raw: raw})
		return nil
	})
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	return out, nil
}

// shouldInclude 决定是否纳入 sync
//
// 规则：
//  1. .md / .markdown 文件总是收
//  2. skills/<name>/ 子树（任何深度）下的非 markdown 文件也收（SKILL package 辅助文件）
//
// rel 是相对源根的 slash 风格路径，name 是 basename
func shouldInclude(rel, name string) bool {
	if isMarkdown(name) {
		return true
	}
	// skills/<name>/<...> 子树下的非 markdown 文件
	return strings.HasPrefix(rel, "skills/")
}

// isMarkdown 判断扩展名是否为 .md / .markdown
func isMarkdown(name string) bool {
	lower := strings.ToLower(name)
	return strings.HasSuffix(lower, ".md") || strings.HasSuffix(lower, ".markdown")
}
