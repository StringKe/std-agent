package transformer

import (
	"std-ai/internal/config"
	"std-ai/internal/parser"
	"std-ai/internal/transformer/protocol"
	"std-ai/internal/writer"
)

func init() { Register(&Amp{}) }

// Amp 是 Sourcegraph Amp transformer。
//
// 协议族：AgentsMD + 全 inline（无原生 rules / skills / commands / subagents 子目录）。
// Amp 读 AGENTS.md（自 AGENT.md 迁移），支持多文件 + 嵌套，~/.config/AGENT.md 全局。
// 无 frontmatter / 无原生 skills/commands/subagents -> 全部走 fallback。
type Amp struct{}

// Name 返回 "amp"
func (a *Amp) Name() string { return "amp" }

// Plan 计算输出
func (a *Amp) Plan(docs []*parser.Document, cfg *config.Config) (*writer.Plan, error) {
	return protocol.AgentsMD{}.Plan(FilterDocs(docs, a.Name()), ampAdapter, cfg)
}

// ampAdapter 配置 AgentsMD 协议族的 amp 变体
var ampAdapter = protocol.Adapter{
	Name:                 "amp",
	RootFileName:         "AGENTS.md",
	ManifestSection:      "Reference Rules",
	NestedSupported:      true,
	RulesDir:             "",    // amp 无子目录 rules，全 inline 到 AGENTS.md
	SkillsAsRule:         false, // RulesDir 为空时 SkillsAsRule 会把 skill 写到仓库根，改走 BuildDegradedSkillPackage -> .amp/rules/skills/<name>/SKILL.md
	CommandsDir:          "",
	ReferencesDir:        "",
	SubagentsDir:         "",
	FallbackDir:          ".amp/rules",
	InjectExplainer:      true,
	InjectStdaiTypeField: true,
	InjectTypeGlossary:   true,
}
