package transformer

import (
	"github.com/StringKe/std-agent/internal/config"
	"github.com/StringKe/std-agent/internal/parser"
	"github.com/StringKe/std-agent/internal/transformer/protocol"
	"github.com/StringKe/std-agent/internal/writer"
)

func init() { Register(&Factory{}) }

// Factory 是 Factory.ai Droids transformer。
//
// 协议族：AgentsMD。读 AGENTS.md（含 ~/.factory/AGENTS.md 全局），
// `.factory/rules/*.md`（无 glob 支持，frontmatter 跳过 applyTo），
// `.factory/skills/<name>/SKILL.md`，`.factory/droids/<name>.md`（subagent），
// `.factory/settings.json`。企业付费市场显眼，AGENTS.md 标准核心采纳方之一。
type Factory struct{}

// Name 返回 "factory"
func (f *Factory) Name() string { return "factory" }

// Plan 委托 AgentsMD 协议按 factoryAdapter 计算输出
func (f *Factory) Plan(docs []*parser.Document, cfg *config.Config) (*writer.Plan, error) {
	return protocol.AgentsMD{}.Plan(FilterDocs(docs, f.Name()), factoryAdapter, cfg)
}

// factoryAdapter 配置 AgentsMD 协议族的 factory 变体。
//
//   - SubagentsDir = `.factory/droids`（factory 把 subagents 叫 droids）
//   - GlobsFieldName = ""（factory rules 无 glob 支持，frontmatter 跳过 applyTo）
//   - CommandsDir = `.factory/commands`（与 skills 共存的一等目录；新工作流
//     更推荐 skills，但 commands 仍被消费，https://docs.factory.ai/harness/custom-slash-commands）
//   - ReferencesDir 留空 -> 走 graceful degradation 到 FallbackDir
var factoryAdapter = protocol.Adapter{
	Name:                 "factory",
	RootFileName:         "AGENTS.md",
	ManifestSection:      "Reference Rules",
	NestedSupported:      true,
	RulesDir:             ".factory/rules",
	SkillsDir:            ".factory/skills",
	SubagentsDir:         ".factory/droids",
	CommandsDir:          ".factory/commands",
	ReferencesDir:        "",
	FallbackDir:          ".factory",
	InjectExplainer:      true,
	InjectStdaiTypeField: true,
	GlobsFieldName:       "",
	SkillSupportedFields: []string{"name", "description", "license", "compatibility", "metadata", "allowed-tools", "user-invocable", "disable-model-invocation"},
	InjectTypeGlossary:   true,
}
