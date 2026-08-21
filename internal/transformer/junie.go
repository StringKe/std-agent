package transformer

import (
	"github.com/StringKe/std-agent/internal/config"
	"github.com/StringKe/std-agent/internal/parser"
	"github.com/StringKe/std-agent/internal/transformer/protocol"
	"github.com/StringKe/std-agent/internal/writer"
)

func init() { Register(&Junie{}) }

// Junie 是 JetBrains Junie（IDE + CLI）transformer。
//
// 官方 guidelines 顺序：`.junie/AGENTS.md`，否则根 `AGENTS.md` 加上 `.junie/rules/*.md`。
// stdagent 写共享 `AGENTS.md`，non-root rules 写 `.junie/rules/`，避免再复制一份 `.junie/AGENTS.md`。
// skills 原生 `.junie/skills/<name>/SKILL.md`。
// 来源：https://junie.jetbrains.com/docs/guidelines-and-memory.html
// 与 https://junie.jetbrains.com/docs/agent-skills.html
type Junie struct{}

// Name 返回 "junie"
func (j *Junie) Name() string { return "junie" }

// Plan 委托 AgentsMD。
func (j *Junie) Plan(docs []*parser.Document, cfg *config.Config) (*writer.Plan, error) {
	return protocol.AgentsMD{}.Plan(FilterDocs(docs, j.Name()), junieAdapter, cfg)
}

var junieAdapter = protocol.Adapter{
	Name:                 "junie",
	RootFileName:         "AGENTS.md",
	ManifestSection:      "Reference Rules",
	NestedSupported:      false,
	RulesDir:             ".junie/rules",
	SkillsDir:            ".junie/skills",
	SkillSupportedFields: []string{"name", "description"},
	FallbackDir:          ".junie",
	InjectExplainer:      true,
	InjectStdaiTypeField: true,
	InjectTypeGlossary:   true,
}
