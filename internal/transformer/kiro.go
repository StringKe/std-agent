package transformer

import (
	"github.com/StringKe/std-agent/internal/config"
	"github.com/StringKe/std-agent/internal/parser"
	"github.com/StringKe/std-agent/internal/transformer/protocol"
	"github.com/StringKe/std-agent/internal/writer"
)

func init() { Register(&Kiro{}) }

// Kiro 是 AWS Kiro（Amazon Q Developer CLI 继任）transformer。
//
// 官方：https://kiro.dev/docs/steering/ 、https://kiro.dev/docs/skills/ 、
// https://kiro.dev/docs/custom-agents/
// Q CLI 已迁到 Kiro CLI，项目配置以 `.kiro/` 为准。
type Kiro struct{}

// Name 返回 "kiro"
func (k *Kiro) Name() string { return "kiro" }

// Plan 委托 AgentsMD，rules 走 Kiro steering inclusion 方言。
func (k *Kiro) Plan(docs []*parser.Document, cfg *config.Config) (*writer.Plan, error) {
	return protocol.AgentsMD{}.Plan(FilterDocs(docs, k.Name()), kiroAdapter, cfg)
}

// kiroAdapter：
//
//   - 根与嵌套 AGENTS.md 被 Kiro 当 always-on steering（无 inclusion 模式）
//   - nonRoot rules 写 `.kiro/steering/`，frontmatter 用 inclusion
//   - skills 写 `.kiro/skills/`（Agent Skills 标准；同时可 /name slash）
//   - commands 降为 `.kiro/skills/commands/<n>/SKILL.md`
//   - subagents 写 `.kiro/agents/<n>.md`
//   - references 落到 `.kiro/references/`，不进 steering（CLI 会全量加载 steering）
var kiroAdapter = protocol.Adapter{
	Name:                  "kiro",
	RootFileName:          "AGENTS.md",
	ManifestSection:       "Reference Rules",
	NestedSupported:       true,
	RulesDir:              ".kiro/steering",
	SkillsDir:             ".kiro/skills",
	SkillSupportedFields:  []string{"name", "description", "license", "compatibility", "metadata"},
	CommandFormat:         protocol.CommandSkillPrefix,
	CommandsAsSkillSubdir: "commands",
	SubagentsDir:          ".kiro/agents",
	FallbackDir:           ".kiro",
	InjectExplainer:       true,
	InjectStdaiTypeField:  true,
	RuleTriggerMode:       protocol.TriggerInclusion,
	InjectTypeGlossary:    true,
}
