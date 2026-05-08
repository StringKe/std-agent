package transformer

import (
	"std-ai/internal/config"
	"std-ai/internal/parser"
	"std-ai/internal/writer"
)

func init() {
	Register(&Aider{})
}

// Aider transformer：v1.0 默认 noop。
//
// Aider 不支持任何 skill / agent / persona / 自定义 chat mode 扩展，
// 也不会自动消费 AGENTS.md（必须 read: 显式声明）。stdagent 在 v1.0
// 依赖 codex transformer 写的根 AGENTS.md，由用户手动在 .aider.conf.yml
// 加 `read: [AGENTS.md, CONVENTIONS.md]`；后续 init --aider 显式开关
// 时再写 .aider.conf.yml。
type Aider struct{}

// Name 返回 "aider"
func (a *Aider) Name() string { return "aider" }

// Plan 返回空 Plan（noop）
func (a *Aider) Plan(_ []*parser.Document, _ *config.Config) (*writer.Plan, error) {
	return &writer.Plan{Target: a.Name()}, nil
}
