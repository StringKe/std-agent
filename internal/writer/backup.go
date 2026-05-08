package writer

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// Backup 把指定文件快照到 .stdai/backups/<RFC3339-utc>/
type Backup struct {
	BackupRoot string
	Keep       int
}

// NewBackup 创建 Backup
func NewBackup(root string, keep int) *Backup {
	if keep < 1 {
		keep = 1
	}
	return &Backup{BackupRoot: root, Keep: keep}
}

// Snapshot 把项目根下的指定文件复制到带 timestamp 的子目录
// projectRoot 是项目根，files 是相对项目根的路径
// 返回 backup dir（仅在实际备份了至少一个文件时非空）
func (b *Backup) Snapshot(projectRoot string, files []string) (string, error) {
	if len(files) == 0 {
		return "", nil
	}
	ts := time.Now().UTC().Format("2006-01-02T15-04-05Z")
	dir := filepath.Join(b.BackupRoot, ts)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return "", err
	}
	wrote := false
	for _, f := range files {
		src := filepath.Join(projectRoot, f)
		if !fileExists(src) {
			continue
		}
		dst := filepath.Join(dir, f)
		if err := copyFile(src, dst); err != nil {
			return "", err
		}
		wrote = true
	}
	if !wrote {
		_ = os.Remove(dir)
		return "", nil
	}
	if err := b.prune(); err != nil {
		return dir, err
	}
	return dir, nil
}

// prune 删除超过 Keep 的最老 backup 子目录
func (b *Backup) prune() error {
	entries, err := os.ReadDir(b.BackupRoot)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return err
	}
	dirs := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			dirs = append(dirs, e.Name())
		}
	}
	if len(dirs) <= b.Keep {
		return nil
	}
	sort.Strings(dirs)
	for _, d := range dirs[:len(dirs)-b.Keep] {
		_ = os.RemoveAll(filepath.Join(b.BackupRoot, d))
	}
	return nil
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o750); err != nil {
		return err
	}
	in, err := os.Open(src) //nolint:gosec
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()
	out, err := os.Create(dst) //nolint:gosec
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()
	if _, err := io.Copy(out, in); err != nil {
		return fmt.Errorf("copy %s -> %s: %w", src, dst, err)
	}
	return nil
}
