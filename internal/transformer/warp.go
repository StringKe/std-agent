package transformer

import (
	"github.com/StringKe/std-agent/internal/config"
	"github.com/StringKe/std-agent/internal/parser"
	"github.com/StringKe/std-agent/internal/transformer/protocol"
	"github.com/StringKe/std-agent/internal/writer"
)

func init() { Register(&Warp{}) }

// Warp 是 Warp Inc 智能终端 transformer
//
// Warp 自 2026-01 起把默认配置文件由 WARP.md 切换为 AGENTS.md，旧的 WARP.md 仍读。
// 全局 Rules 走 Warp Drive；项目级走根 AGENTS.md + 嵌套子目录 AGENTS.md 自动叠加。
//
// 协议族：AgentsMD（RulesDir 空 -> nonRoot rules inline 进根文件）。
// 无原生 skills / commands / references / subagents 子目录，全部 fallback 到 .warp/rules/。
type Warp struct{}

// Name 返回 "warp"
func (w *Warp) Name() string { return "warp" }

// Plan 计算输出
func (w *Warp) Plan(docs []*parser.Document, cfg *config.Config) (*writer.Plan, error) {
	return protocol.AgentsMD{}.Plan(FilterDocs(docs, w.Name()), warpAdapter, cfg)
}

// warpAdapter 配置 AgentsMD 协议族的 warp 变体
//
// 关键字段：
//   - RootFileName=AGENTS.md（warp 2026-01 起默认）
//   - NestedSupported=true（嵌套子目录 AGENTS.md 自动叠加）
//   - RulesDir 空 -> nonRoot rules inline 进根 AGENTS.md（amp / warp 共享风格）
//   - SkillsDir=".agents/skills"：Warp 原生 Skills 推荐路径
//     （https://docs.warp.dev/agent-platform/capabilities/skills/），与 codex / amp
//     共享落点，SkillSupportedFields 同集保证字节一致由 writer 去重
//   - 其他 type 全部 fallback 到 .warp/rules/<subdir>/<name>.md
var warpAdapter = protocol.Adapter{
	Name:                 "warp",
	RootFileName:         "AGENTS.md",
	ManifestSection:      "Reference Rules",
	NestedSupported:      true,
	RulesDir:             "",
	SkillsDir:            ".agents/skills",
	SkillSupportedFields: []string{"name", "description", "license", "compatibility", "metadata"},
	CommandsDir:          "",
	ReferencesDir:        "",
	SubagentsDir:         "",
	FallbackDir:          ".warp/rules",
	InjectExplainer:      true,
	InjectStdaiTypeField: true,
	InjectTypeGlossary:   true,
}
