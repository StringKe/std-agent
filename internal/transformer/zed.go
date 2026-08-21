package transformer

import (
	"github.com/StringKe/std-agent/internal/config"
	"github.com/StringKe/std-agent/internal/parser"
	"github.com/StringKe/std-agent/internal/transformer/protocol"
	"github.com/StringKe/std-agent/internal/writer"
)

func init() { Register(&Zed{}) }

// Zed 是 Zed 编辑器 Agent transformer。
//
// 官方项目指令优先读 AGENTS.md（兼容 CLAUDE.md 等，但 stdagent 只写共享 AGENTS.md）。
// 项目 skills 仅扫描 `.agents/skills/<name>/SKILL.md`，且必须扁平、不可分子目录。
// 来源：https://zed.dev/docs/ai/instructions 与 https://zed.dev/docs/ai/skills
type Zed struct{}

// Name 返回 "zed"
func (z *Zed) Name() string { return "zed" }

// Plan 委托 AgentsMD。
func (z *Zed) Plan(docs []*parser.Document, cfg *config.Config) (*writer.Plan, error) {
	return protocol.AgentsMD{}.Plan(FilterDocs(docs, z.Name()), zedAdapter, cfg)
}

// zedAdapter 与 Codex / Goose 等共享 `.agents/skills/`，SkillSupportedFields 必须同集。
// 官方 skills 禁止嵌套目录，因此 commands 不走 CommandsAsSkillSubdir，改走私有 `.zed/`。
var zedAdapter = protocol.Adapter{
	Name:                 "zed",
	RootFileName:         "AGENTS.md",
	ManifestSection:      "Reference Rules",
	NestedSupported:      true,
	SkillsDir:            ".agents/skills",
	SkillSupportedFields: []string{"name", "description", "license", "compatibility", "metadata"},
	FallbackDir:          ".zed",
	InjectExplainer:      true,
	InjectStdaiTypeField: true,
	InjectTypeGlossary:   true,
}
