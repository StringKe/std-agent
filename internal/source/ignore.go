package source

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

// Ignore 持有从 .stdaiignore 加载的 glob 模式。
//
// 行为类比 .gitignore：
//   - 每行一个 glob（doublestar 风格，支持 `**`）
//   - 空行与 `#` 起始的注释行被忽略
//   - 路径用 forward slash（与 source.File.Path 一致）
//   - Match 命中即视为忽略
type Ignore struct {
	patterns []string
}

// LoadIgnoreFile 从 path 加载 .stdaiignore；文件不存在返回空 Ignore（无报错）
func LoadIgnoreFile(path string) (*Ignore, error) {
	f, err := os.Open(path) //nolint:gosec
	if err != nil {
		if os.IsNotExist(err) {
			return &Ignore{}, nil
		}
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer func() { _ = f.Close() }()
	return parseIgnore(f)
}

// parseIgnore 把 reader 内容按行解析成 patterns
func parseIgnore(r io.Reader) (*Ignore, error) {
	ig := &Ignore{}
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		ig.patterns = append(ig.patterns, line)
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("read ignore: %w", err)
	}
	return ig, nil
}

// Match 返回 true 表示 rel 应被忽略（匹配任一 pattern）
//
// rel 必须是 forward-slash 形式的相对路径（与 source.File.Path 一致）。
func (ig *Ignore) Match(rel string) bool {
	if ig == nil || len(ig.patterns) == 0 {
		return false
	}
	for _, p := range ig.patterns {
		ok, _ := doublestar.Match(p, rel)
		if ok {
			return true
		}
	}
	return false
}

// Patterns 返回 patterns 副本（用于诊断与测试）
func (ig *Ignore) Patterns() []string {
	if ig == nil {
		return nil
	}
	out := make([]string, len(ig.patterns))
	copy(out, ig.patterns)
	return out
}
