package transformer

import (
	"std-ai/internal/config"
	"std-ai/internal/parser"
	"std-ai/internal/transformer/protocol"
	"std-ai/internal/writer"
)

func init() { Register(&Codex{}) }

// Codex 是 OpenAI Codex CLI transformer
type Codex struct{}

// Name 返回 "codex"
func (c *Codex) Name() string { return "codex" }

// Plan 委托 AgentsMD 协议按 codexAdapter 计算输出
func (c *Codex) Plan(docs []*parser.Document, cfg *config.Config) (*writer.Plan, error) {
	return protocol.AgentsMD{}.Plan(FilterDocs(docs, c.Name()), codexAdapter, cfg)
}

var codexAdapter = protocol.Adapter{
	Name:                  "codex",
	RootFileName:          "AGENTS.md",
	ManifestSection:       "Reference Rules",
	NestedSupported:       true,
	RulesDir:              ".codex/memories",
	SkillsDir:             ".agents/skills",
	FallbackDir:           ".codex/memories",
	InjectExplainer:       true,
	InjectStdaiTypeField:  true,
	SkillSupportedFields:  []string{"name", "description", "license", "compatibility", "metadata"},
	CommandFormat:         protocol.CommandSkillPrefix,
	CommandsAsSkillSubdir: "commands", // v3：子目录隔离（避免 cmd- 私有前缀污染 skill 命名空间）
	InjectCommandsToRoot:  true,
	InjectTypeGlossary:    true,
}
