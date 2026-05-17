// Package protocol 抽象协议族（AgentsMD / ClaudeMD / Cursor / Clinerules /
// WindsurfStyle / Copilot）。每个协议实现 Protocol 接口，transformer 层通过
// Adapter struct literal 注入 target 专属配置后委托给协议实现，避免 11+ 个
// transformer 重复实现相同的 fanout 模板。
//
// 协议层只看 Adapter / parser.Document / config.Config / writer.Plan，不感知
// 具体 target 名，便于增量加入新 target。
package protocol

import (
	"std-ai/internal/config"
	"std-ai/internal/parser"
	"std-ai/internal/writer"
)

// Protocol 是协议族层抽象。每个协议族（AgentsMD / ClaudeMD / Cursor /
// Clinerules / WindsurfStyle / Copilot）实现一个。
//
// Plan 接受筛选后的 docs（已通过 transformer.FilterDocs 过滤 target）+
// adapter 配置 + cfg，返回完整 *writer.Plan。内部按 type 分桶并处理
// cross-cutting 逻辑（如 InjectCommandsToRoot：commands 段合并到 root
// FileOp.Content 的 footer marker 之前）。
//
// 实现者约束：
//   - adapter.Disabled=true 时直接返回 &writer.Plan{Target: adapter.Name}, nil
//   - 不修改入参 docs / adapter / cfg
//   - 返回的 Plan.Files 顺序稳定（依赖 transformerutil.SortDocs 排序）
type Protocol interface {
	Plan(docs []*parser.Document, adapter Adapter, cfg *config.Config) (*writer.Plan, error)
}
