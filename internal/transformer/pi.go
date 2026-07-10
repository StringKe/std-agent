package transformer

import (
	"github.com/StringKe/std-agent/internal/config"
	"github.com/StringKe/std-agent/internal/parser"
	"github.com/StringKe/std-agent/internal/transformer/protocol"
	"github.com/StringKe/std-agent/internal/writer"
)

func init() { Register(&Pi{}) }

// Pi 是 earendil-works pi (github.com/earendil-works/pi) transformer
//
// 协议族：AgentsMD。pi 是高度可扩展的 agent runner，读 AGENTS.md
// （`~/.pi/agent/` + parent + cwd 三级加载），原生支持 Agent Skills
// 标准（.pi/skills/ 与 .agents/skills/），prompts 走 .pi/prompts/。
// 严格执行者：不容忍非法 skill frontmatter。
type Pi struct{}

// Name 返回 "pi"
func (p *Pi) Name() string { return "pi" }

// Plan 委托 AgentsMD 协议按 piAdapter 计算输出
func (p *Pi) Plan(docs []*parser.Document, cfg *config.Config) (*writer.Plan, error) {
	return protocol.AgentsMD{}.Plan(FilterDocs(docs, p.Name()), piAdapter, cfg)
}

// piAdapter 配置 AgentsMD 协议族的 pi 变体
//
//   - 无 RulesDir：所有 nonRoot rules inline 到 AGENTS.md（pi 无子目录 rules）
//   - SkillsDir=.pi/skills：原生 Agent Skills 标准
//   - CommandsDir=.pi/prompts：pi 把 slash 模板放在 prompts/
//   - FallbackDir=.pi/rules：references / subagents fallback 落点
var piAdapter = protocol.Adapter{
	Name:                 "pi",
	RootFileName:         "AGENTS.md",
	ManifestSection:      "Reference Rules",
	NestedSupported:      true,
	RulesDir:             "", // pi 无子目录 rules，全 inline
	SkillsDir:            ".pi/skills",
	CommandsDir:          ".pi/prompts",
	ReferencesDir:        "",
	SubagentsDir:         "",
	FallbackDir:          ".pi/rules",
	InjectExplainer:      true,
	InjectStdaiTypeField: true,
	SkillSupportedFields: []string{"name", "description", "license", "compatibility", "metadata"},
	InjectTypeGlossary:   true,
}
