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
//
// v3 修订：SkillsAsRule=false。原 SkillsAsRule=true 会产 `.agents/rules/skill-<name>.md`
// 含 std-ai 私有 `skill-` 前缀，违反 v3 "不造私有前缀，用子目录隔离" 原则。
// 改走 BuildDegradedSkillPackage 标准 Agent Skills fallback：`.agents/rules/skills/<name>/SKILL.md`。
var antigravityAdapter = protocol.Adapter{
	Name:                 "antigravity",
	RootFileName:         "", // 复用 codex 写的根 AGENTS.md，不重复写
	NestedSupported:      false,
	RulesDir:             ".agents/rules",
	SkillsAsRule:         false, // 走 fallback（v3 子目录隔离）
	CommandsDir:          ".agents/workflows",
	FallbackDir:          ".agents/rules",
	InjectExplainer:      true,
	InjectStdaiTypeField: true,
	RuleTriggerMode:      protocol.TriggerTrigger, // trigger frontmatter
	MaxBytesPerFile:      12000,                   // antigravity 单 rule / workflow 字符上限（实测）
	SoftBytes:            8000,
}
