package transformer

import (
	"std-ai/internal/config"
	"std-ai/internal/parser"
	"std-ai/internal/transformer/protocol"
	"std-ai/internal/writer"
)

func init() { Register(&GrokCLI{}) }

// GrokCLI 是 xAI Grok CLI transformer。
//
// 协议族：AgentsMD。grok-cli 3 个并行 fork（superagent-ai / alphaonedev / baba20o）
// 主流未定，v0.0.4 对齐 superagent-ai fork（最早实现）。读项目根 AGENTS.md
// （merge），AGENTS.override.md per-dir 覆盖。.grok/settings.json 存 MCP / model，
// .grok/GROK.md 旧路径已被 AGENTS.md 取代。
//
// 与 amp / warp 同属"AGENTS.md 全 inline"风格：无原生子目录 rules / skills /
// commands / references / subagents -> 全部走 fallback 到 .grok/rules/<subdir>/。
type GrokCLI struct{}

// Name 返回 "grok-cli"
func (g *GrokCLI) Name() string { return "grok-cli" }

// Plan 委托 AgentsMD 协议按 grokCLIAdapter 计算输出
func (g *GrokCLI) Plan(docs []*parser.Document, cfg *config.Config) (*writer.Plan, error) {
	return protocol.AgentsMD{}.Plan(FilterDocs(docs, g.Name()), grokCLIAdapter, cfg)
}

// grokCLIAdapter 配置 AgentsMD 协议族的 grok-cli 变体。
//
// PerDirOverrideFileName 字段在 adapter.go 已定义但 v0.0.4 protocol 未完整使用，
// 保留意图标识。grok 的 AGENTS.override.md per-dir 覆盖行为完整实现待 v0.0.5。
//
// SkillsAsRule 设为 false：RulesDir 为空时若 SkillsAsRule=true 会把 skill 写到
// 仓库根 skill-<name>.md，与 amp 风格保持一致，改走 BuildDegradedSkillPackage
// 落到 .grok/rules/skills/<name>/SKILL.md。
var grokCLIAdapter = protocol.Adapter{
	Name:                   "grok-cli",
	RootFileName:           "AGENTS.md",
	ManifestSection:        "Reference Rules",
	NestedSupported:        true,
	PerDirOverrideFileName: "AGENTS.override.md", // grok 私有：per-dir 覆盖文件（v0.0.5 完整实现）
	RulesDir:               "",                   // 无子目录 rules，全 inline 到 AGENTS.md
	SkillsAsRule:           false,                // 走 BuildDegradedSkillPackage -> .grok/rules/skills/<name>/SKILL.md
	CommandsDir:            "",
	ReferencesDir:          "",
	SubagentsDir:           "",
	FallbackDir:            ".grok/rules",
	InjectExplainer:        true,
	InjectStdaiTypeField:   true,
	InjectTypeGlossary:     true,
}
