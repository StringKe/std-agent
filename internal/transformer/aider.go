package transformer

import (
	"std-ai/internal/config"
	"std-ai/internal/parser"
	"std-ai/internal/writer"
)

func init() { Register(&Aider{}) }

// Aider 是 OpenAI Aider transformer。
// aider 通过 .aider.conf.yml 显式 read AGENTS.md，依赖 codex 写出。本 transformer
// 是注册占位，不产 FileOp。
type Aider struct{}

// Name 返回 "aider"。
func (a *Aider) Name() string { return "aider" }

// Plan 返回空 Plan（noop）。
func (a *Aider) Plan(_ []*parser.Document, _ *config.Config) (*writer.Plan, error) {
	return &writer.Plan{Target: a.Name()}, nil
}
