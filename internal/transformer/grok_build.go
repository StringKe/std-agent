package transformer

import (
	"std-ai/internal/config"
	"std-ai/internal/parser"
	"std-ai/internal/transformer/protocol"
	"std-ai/internal/writer"
)

func init() { Register(&GrokBuild{}) }

// GrokBuild 是 xAI 官方 agentic CLI transformer（二进制名 `grok`）。
//
// 协议族：AgentsMD。读项目根 AGENTS.md（family：AGENTS.md / Agents.md / AGENT.md），
// 从 cwd 向上 walk 到 git root 全部叠加。零配置兼容 Claude Code（自动读 CLAUDE.md
// / .claude/rules/ / .claude/skills/ / .claude/agents/），所以 stdagent 只需要
// 写 AGENTS.md + .grok/skills/<name>/SKILL.md（大写）。
//
// 与 superagent-ai / alphaonedev 等社区 grok-cli fork 完全无关——后者用
// .grok/GROK.md / AGENTS.override.md 等非官方约定，本 adapter 不支持。
//
// 调研：docs/targets/grok-build.md（基于 https://docs.x.ai/build/overview）。
type GrokBuild struct{}

// Name 返回 "grok-build"
func (g *GrokBuild) Name() string { return "grok-build" }

// Plan 委托 AgentsMD 协议按 grokBuildAdapter 计算输出
func (g *GrokBuild) Plan(docs []*parser.Document, cfg *config.Config) (*writer.Plan, error) {
	return protocol.AgentsMD{}.Plan(FilterDocs(docs, g.Name()), grokBuildAdapter, cfg)
}

// grokBuildAdapter 按 xAI 官方文档配置
//
//   - 主指令 AGENTS.md（多层 walk 叠加，无 frontmatter）
//   - SkillsDir=".grok/skills"（Agent Skills 标准，SKILL.md 大写）
//   - CommandsDir 空：项目级无独立 commands，user-invocable skill 自动暴露为 /<name> slash
//   - ReferencesDir 空：无原生 references 概念，走 fallback 子目录
//   - SubagentsDir 空：plugins 内 agents/ 才是 subagents 路径，项目级无独立路径
var grokBuildAdapter = protocol.Adapter{
	Name:                 "grok-build",
	RootFileName:         "AGENTS.md",
	ManifestSection:      "Reference Rules",
	NestedSupported:      true,
	RulesDir:             "",
	SkillsDir:            ".grok/skills",
	SkillSupportedFields: []string{"name", "description", "license", "compatibility", "metadata", "allowed-tools", "disable-model-invocation"},
	CommandsDir:          "",
	ReferencesDir:        "",
	SubagentsDir:         "",
	FallbackDir:          ".grok/rules",
	InjectExplainer:      true,
	InjectStdaiTypeField: true,
	InjectTypeGlossary:   true,
}
