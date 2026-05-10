package writer

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// Writer 是文件落盘抽象
type Writer struct {
	ProjectRoot string
	DryRun      bool
}

// NewWriter 创建 writer
func NewWriter(root string, dryRun bool) *Writer {
	return &Writer{ProjectRoot: root, DryRun: dryRun}
}

// Apply 把 Plan 写入磁盘
func (w *Writer) Apply(plan *Plan) (written, skipped int, err error) {
	for i := range plan.Files {
		op := &plan.Files[i]
		full := filepath.Join(w.ProjectRoot, op.Path)
		if op.Skip {
			skipped++
			continue
		}
		exist, _ := os.ReadFile(full) //nolint:gosec
		if len(exist) > 0 && bytes.Equal(exist, op.Content) {
			op.Skip = true
			op.Reason = "unchanged"
			skipped++
			continue
		}
		// marker 含每次 sync 不同的 timestamp（footer.HeaderComment 注入 GeneratedAt），
		// 归一化 marker 后再比一次：源文档未变则不重写，避免污染 git diff / mtime。
		if len(exist) > 0 && bytes.Equal(stripGeneratedMarker(exist), stripGeneratedMarker(op.Content)) {
			op.Skip = true
			op.Reason = "unchanged"
			skipped++
			continue
		}
		if w.DryRun {
			continue
		}
		if mkErr := os.MkdirAll(filepath.Dir(full), 0o750); mkErr != nil {
			return written, skipped, fmt.Errorf("mkdir %s: %w", filepath.Dir(full), mkErr)
		}
		if wErr := atomicWrite(full, op.Content); wErr != nil {
			return written, skipped, wErr
		}
		written++
	}
	return written, skipped, nil
}

// atomicWrite 写临时文件然后 rename，避免半写状态
func atomicWrite(path string, data []byte) error {
	tmp := path + ".stdagent.tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return fmt.Errorf("write tmp %s: %w", tmp, err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("rename %s -> %s: %w", tmp, path, err)
	}
	return nil
}

// Checksum 返回内容的 SHA256 hex（state.json 用）
func Checksum(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

// FileExists 检查路径是否存在（不区分是否文件）
func FileExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	return !errors.Is(err, fs.ErrNotExist)
}
