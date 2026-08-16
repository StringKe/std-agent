package transformer

import (
	"github.com/StringKe/std-agent/internal/config"
	"github.com/StringKe/std-agent/internal/parser"
	"github.com/StringKe/std-agent/internal/transformer/protocol"
	"github.com/StringKe/std-agent/internal/writer"
)

func init() { Register(&Goose{}) }

// Goose 是 Linux Foundation AAIF goose transformer（原 Block goose）。
//
// 官方默认上下文文件是 AGENTS.md，其次才是 .goosehints。
// 项目 skills 官方推荐 `.agents/skills/`（兼容 .goose/skills 与 .claude/skills）。
// 来源：https://goose-docs.ai/docs/guides/context-engineering/using-skills/
// 与 https://goose-docs.ai/docs/guides/context-engineering/using-goosehints/
type Goose struct{}

// Name 返回 "goose"
func (g *Goose) Name() string { return "goose" }

// Plan 委托 AgentsMD。
func (g *Goose) Plan(docs []*parser.Document, cfg *config.Config) (*writer.Plan, error) {
	return protocol.AgentsMD{}.Plan(FilterDocs(docs, g.Name()), gooseAdapter, cfg)
}

// gooseAdapter 与 Codex / Amp / Warp / Kimi / Antigravity 共享 `.agents/skills/`，
// SkillSupportedFields 必须同集以保证字节一致。不写 .goosehints，避免与共享
// AGENTS.md 重复注入。FallbackDir 走私有 `.goose/`，explainer 含 target 名。
var gooseAdapter = protocol.Adapter{
	Name:                  "goose",
	RootFileName:          "AGENTS.md",
	ManifestSection:       "Reference Rules",
	NestedSupported:       true,
	RulesDir:              "",
	SkillsDir:             ".agents/skills",
	SkillSupportedFields:  []string{"name", "description", "license", "compatibility", "metadata"},
	CommandFormat:         protocol.CommandSkillPrefix,
	CommandsAsSkillSubdir: "commands",
	ReferencesDir:         "",
	SubagentsDir:          "",
	FallbackDir:           ".goose",
	InjectExplainer:       true,
	InjectStdaiTypeField:  true,
	InjectTypeGlossary:    true,
}
