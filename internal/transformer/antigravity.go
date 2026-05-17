package transformer

import (
	"std-ai/internal/config"
	"std-ai/internal/parser"
	"std-ai/internal/transformer/protocol"
	"std-ai/internal/writer"
)

func init() {
	Register(&Antigravity{})
}

// Antigravity 是 Google Antigravity AI IDE transformer
//
// 协议族：AgentsMD + RuleTriggerMode=TriggerTrigger（windsurf 风格 trigger frontmatter）。
// antigravity 自 v1.20.3 起原生消费根 AGENTS.md（由 codex transformer 写入），
// 因此本 transformer 仅输出 .agents/rules / .agents/workflows / skills 降级 rule，
// 不重复写根 AGENTS.md（adapter.RootFileName 留空）。
type Antigravity struct{}

// Name 返回 "antigravity"
func (a *Antigravity) Name() string { return "antigravity" }

// Plan 计算输出
func (a *Antigravity) Plan(docs []*parser.Document, cfg *config.Config) (*writer.Plan, error) {
	return protocol.AgentsMD{}.Plan(FilterDocs(docs, a.Name()), antigravityAdapter, cfg)
}

// antigravityAdapter 配置 AgentsMD 协议族的 antigravity 变体
var antigravityAdapter = protocol.Adapter{
	Name:            "antigravity",
	RootFileName:    "", // 复用 codex 写的根 AGENTS.md，不重复写
	NestedSupported: false,
	RulesDir:        ".agents/rules",
	SkillsAsRule:    true, // 无原生 skills，降级为 model_decision rule
	CommandsDir:     ".agents/workflows",
	FallbackDir:     ".agents/rules",
	RuleTriggerMode: protocol.TriggerTrigger, // trigger frontmatter（always_on / glob / model_decision / manual）
	MaxBytesPerFile: 12000,                   // antigravity 单 rule / workflow 字符上限（实测）
	SoftBytes:       8000,
}
