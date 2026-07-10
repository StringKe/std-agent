package transformer

import (
	"github.com/StringKe/std-agent/internal/config"
	"github.com/StringKe/std-agent/internal/parser"
	"github.com/StringKe/std-agent/internal/transformer/protocol"
	"github.com/StringKe/std-agent/internal/writer"
)

func init() { Register(&Amp{}) }

// Amp 是 Sourcegraph Amp transformer。
//
// 协议族：AgentsMD。Amp 读 AGENTS.md（多文件 + 嵌套，~/.config/amp/AGENTS.md 全局），
// 原生 Agent Skills 于 .agents/skills/（https://ampcode.com/manual/agent-skills.md）。
// 自定义 commands 已被官方移除并入 skills（2026-01-29
// https://ampcode.com/news/slashing-custom-commands），/name 直接调用同名 skill，
// 因此 commands 与 codex 相同降级为 .agents/skills/commands/<n>/SKILL.md。
// subagents 仍是运行时动态 mini-Amp（Task tool），无文件化定义，走 fallback。
type Amp struct{}

// Name 返回 "amp"
func (a *Amp) Name() string { return "amp" }

// Plan 计算输出
func (a *Amp) Plan(docs []*parser.Document, cfg *config.Config) (*writer.Plan, error) {
	return protocol.AgentsMD{}.Plan(FilterDocs(docs, a.Name()), ampAdapter, cfg)
}

// ampAdapter 配置 AgentsMD 协议族的 amp 变体
//
// skills / commands 落点与 codex 完全一致（.agents/skills/ 共享命名空间 +
// SkillSupportedFields 同集），两个 target 产出字节相同、writer 按 unchanged 去重。
var ampAdapter = protocol.Adapter{
	Name:                  "amp",
	RootFileName:          "AGENTS.md",
	ManifestSection:       "Reference Rules",
	NestedSupported:       true,
	RulesDir:              "", // amp 无子目录 rules，全 inline 到 AGENTS.md
	SkillsDir:             ".agents/skills",
	SkillSupportedFields:  []string{"name", "description", "license", "compatibility", "metadata"},
	CommandFormat:         protocol.CommandSkillPrefix,
	CommandsAsSkillSubdir: "commands",
	ReferencesDir:         "",
	SubagentsDir:          "",
	FallbackDir:           ".amp/rules",
	InjectExplainer:       true,
	InjectStdaiTypeField:  true,
	InjectTypeGlossary:    true,
}
