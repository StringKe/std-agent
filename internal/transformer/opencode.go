package transformer

import (
	"std-ai/internal/config"
	"std-ai/internal/parser"
	"std-ai/internal/transformer/protocol"
	"std-ai/internal/writer"
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
//   - skills 走 opencode 单文件 agent 形态（SkillsAsSubagent=true + permission 全 ask 保守安全）
//   - commands 原生 markdown 到 .opencode/commands/
var opencodeAdapter = protocol.Adapter{
	Name:                     "opencode",
	RootFileName:             "",
	NestedSupported:          false,
	RulesDir:                 "",
	SkillsDir:                ".opencode/agents",
	CommandsDir:              ".opencode/commands",
	FallbackDir:              ".opencode", // references / subagents fallback 子目录隔离
	InjectExplainer:          true,
	InjectStdaiTypeField:     true,
	SkillsAsSubagent:         true,
	SkillsSubagentPermission: "{ edit: ask, bash: ask, read: allow, glob: allow, grep: allow, list: allow, task: ask, lsp: allow }",
}
