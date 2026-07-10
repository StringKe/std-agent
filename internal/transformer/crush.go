package transformer

import (
	"github.com/StringKe/std-agent/internal/config"
	"github.com/StringKe/std-agent/internal/parser"
	"github.com/StringKe/std-agent/internal/transformer/protocol"
	"github.com/StringKe/std-agent/internal/writer"
)

func init() { Register(&Crush{}) }

// Crush 是 Charmbracelet Crush transformer（TUI 同源 Go 工具，5k+ stars）。
//
// 协议族：AgentsMD。crush 通过 context_paths 同时读 AGENTS.md / CRUSH.md /
// CLAUDE.md / GEMINI.md，并扫描 .agents/skills/ / .crush/skills/ /
// .claude/skills/ / .cursor/skills/ 多目录。本 transformer 写 CRUSH.md
// 作为 crush 私有根（避免与 codex 共消费时互相覆盖），skills 走 .crush/skills/，
// 其余类型走 .crush/rules/ 下子目录 fallback。
type Crush struct{}

// Name 返回 "crush"
func (c *Crush) Name() string { return "crush" }

// Plan 委托 AgentsMD 协议按 crushAdapter 计算输出
func (c *Crush) Plan(docs []*parser.Document, cfg *config.Config) (*writer.Plan, error) {
	return protocol.AgentsMD{}.Plan(FilterDocs(docs, c.Name()), crushAdapter, cfg)
}

// crushAdapter 配置 AgentsMD 协议族的 crush 变体
var crushAdapter = protocol.Adapter{
	Name:                 "crush",
	RootFileName:         "CRUSH.md",
	ManifestSection:      "Reference Rules",
	NestedSupported:      true,
	RulesDir:             "", // crush 无子目录 rules，全 inline 到 CRUSH.md
	SkillsDir:            ".crush/skills",
	CommandsDir:          "", // 无原生 commands，fallback
	ReferencesDir:        "", // fallback
	SubagentsDir:         "", // fallback
	FallbackDir:          ".crush/rules",
	InjectExplainer:      true,
	InjectStdaiTypeField: true,
	SkillSupportedFields: []string{"name", "description", "license", "compatibility", "metadata"},
	InjectTypeGlossary:   true,
	MaxBytesPerFile:      0, // 无明确字节限制
}
