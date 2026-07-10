package transformer

import (
	"github.com/StringKe/std-agent/internal/config"
	"github.com/StringKe/std-agent/internal/parser"
	"github.com/StringKe/std-agent/internal/transformer/protocol"
	"github.com/StringKe/std-agent/internal/writer"
)

func init() { Register(&OpenCode{}) }

// OpenCode 是 sst/opencode CLI transformer
//
// OpenCode 自动消费 codex transformer 写的根 AGENTS.md，rules 不重复落盘。
// rules 的 applyTo 在 OpenCode 上被丢弃（OpenCode 不支持 frontmatter 条件激活）。
type OpenCode struct{}

// Name 返回 "opencode"
func (o *OpenCode) Name() string { return "opencode" }

// Plan 委托 AgentsMD 协议按 opencodeAdapter 计算输出
func (o *OpenCode) Plan(docs []*parser.Document, cfg *config.Config) (*writer.Plan, error) {
	return protocol.AgentsMD{}.Plan(FilterDocs(docs, o.Name()), opencodeAdapter, cfg)
}

// opencodeAdapter 配置 opencode 的协议族行为：
//   - 不写根 AGENTS.md（RootFileName 为空，复用 codex 的根文件）
//   - rules 不落盘（RulesDir 为空 + RootFileName 为空 -> rules 静默丢弃，由 codex AGENTS.md 承担）
//   - skills 原生 Agent Skills 标准包 .opencode/skills/<n>/SKILL.md（官方已 GA，
//     旧的"skill 降级为 mode: subagent 单文件 agent"方案随之废弃）
//   - commands 原生 markdown 到 .opencode/commands/（官方复数为主，单数向后兼容）
//   - NestedSupported=false：opencode 会在 read 子目录文件时动态注入沿途
//     AGENTS.md，但嵌套文件本身由 codex target 写入，这里不重复写
var opencodeAdapter = protocol.Adapter{
	Name:                 "opencode",
	RootFileName:         "",
	NestedSupported:      false,
	RulesDir:             "",
	SkillsDir:            ".opencode/skills",
	CommandsDir:          ".opencode/commands",
	FallbackDir:          ".opencode", // references / subagents fallback 子目录隔离
	InjectExplainer:      true,
	InjectStdaiTypeField: true,
}
