package transformer

import (
	"std-ai/internal/config"
	"std-ai/internal/parser"
	"std-ai/internal/writer"
)

// Transformer 是所有 target 的统一接口
type Transformer interface {
	Name() string
	Plan(docs []*parser.Document, cfg *config.Config) (*writer.Plan, error)
}

// Registry 全局注册的所有 target transformer
var Registry = map[string]Transformer{}

// Register 在包 init 时调用注入 transformer
func Register(t Transformer) {
	Registry[t.Name()] = t
}

// Get 按 target 名取出 transformer
func Get(name string) (Transformer, bool) {
	t, ok := Registry[name]
	return t, ok
}

// Names 列出已注册 transformer 名（用于诊断）
func Names() []string {
	out := make([]string, 0, len(Registry))
	for n := range Registry {
		out = append(out, n)
	}
	return out
}
