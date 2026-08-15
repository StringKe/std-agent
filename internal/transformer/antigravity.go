package transformer

import (
	"github.com/StringKe/std-agent/internal/config"
	"github.com/StringKe/std-agent/internal/parser"
	"github.com/StringKe/std-agent/internal/transformer/protocol"
	"github.com/StringKe/std-agent/internal/writer"
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
// skills 原生 `.agents/skills/<n>/SKILL.md`（antigravity.google/docs/skills 确认
// workspace 路径固定）。与 codex 共享同一落点，SkillSupportedFields 保持与 codex
// 一致让两个 target 产出字节相同、writer 按 unchanged 去重（antigravity 官方
// frontmatter 仅 name/description，多余的 Agent Skills 标准字段被忽略无害）。
var antigravityAdapter = protocol.Adapter{
	Name:                 "antigravity",
	RootFileName:         "", // 复用 codex 写的根 AGENTS.md，不重复写
	NestedSupported:      false,
	RulesDir:             ".agents/rules",
	SkillsDir:            ".agents/skills",
	SkillSupportedFields: []string{"name", "description", "license", "compatibility", "metadata"},
	CommandsDir:          ".agents/workflows",
	SubagentsDir:         ".agents/agents",
	FallbackDir:          ".agents/rules",
	InjectExplainer:      true,
	InjectStdaiTypeField: true,
	RuleTriggerMode:      protocol.TriggerTrigger, // trigger frontmatter
	MaxBytesPerFile:      12000,                   // antigravity 单 rule / workflow 字符上限（实测）
	SoftBytes:            8000,
}
