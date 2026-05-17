package transformer

import (
	"std-ai/internal/config"
	"std-ai/internal/parser"
	"std-ai/internal/transformer/protocol"
	"std-ai/internal/writer"
)

func init() { Register(&Jules{}) }

// Jules 是 Google Jules transformer。
//
// 协议族：AgentsMD。Jules（jules.google）原生消费根 AGENTS.md，
// 无独立 rules / skills / commands / references / subagents 子目录约定，
// 因此 RulesDir/SkillsDir/CommandsDir/ReferencesDir/SubagentsDir 全部置空，
// 全部走 FallbackDir=.jules/rules 降级路径。
type Jules struct{}

// Name 返回 "jules"
func (j *Jules) Name() string { return "jules" }

// Plan 计算输出
func (j *Jules) Plan(docs []*parser.Document, cfg *config.Config) (*writer.Plan, error) {
	return protocol.AgentsMD{}.Plan(FilterDocs(docs, j.Name()), julesAdapter, cfg)
}

// julesAdapter 配置 AgentsMD 协议族的 jules 变体
var julesAdapter = protocol.Adapter{
	Name:            "jules",
	RootFileName:    "AGENTS.md",
	ManifestSection: "Reference Rules",
	NestedSupported: true,
	RulesDir:        "", // jules 无子目录 rules，nonRoot rules 全 inline 到 AGENTS.md
	// RulesDir 为空时 SkillsAsRule=true 会把 skill 写到仓库根，故置 false
	// 让 skills 走 BuildDegradedSkillPackage -> .jules/rules/skills/<name>/SKILL.md
	SkillsAsRule:         false,
	CommandsDir:          "",
	ReferencesDir:        "",
	SubagentsDir:         "",
	FallbackDir:          ".jules/rules",
	InjectExplainer:      true,
	InjectStdaiTypeField: true,
	InjectTypeGlossary:   true,
}
