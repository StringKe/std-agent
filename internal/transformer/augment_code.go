package transformer

import (
	"github.com/StringKe/std-agent/internal/config"
	"github.com/StringKe/std-agent/internal/parser"
	"github.com/StringKe/std-agent/internal/transformer/protocol"
	"github.com/StringKe/std-agent/internal/writer"
)

func init() {
	Register(&AugmentCode{})
}

// AugmentCode 是 Augment（augmentcode.com）transformer
//
// 协议族：WindsurfStyle（trigger frontmatter 语义近似 augment 的 type 字段：
// always_apply / agent_requested / manual <-> always_on / model_decision / manual）。
// Augment 私有方言 type 字段在 v0.0.4 未严格对齐，仍输出 trigger，由 Augment 容忍
// 多余 frontmatter 字段（实测 Augment 不会因未知键报错）。未来如需严格映射，
// 再独立 AugmentStyle protocol（v0.0.5+）。
//
// 落点：
//   - rules -> .augment/rules/<name>.md（trigger frontmatter）
//   - skills / references / subagents -> .augment/rules/{skills,references,subagents}/<name>.md
//     （走 graceful degradation）
//   - commands -> .augment/rules/workflows/<name>.md
//
// 老版兼容：SingleFileFallback 配 ".augment-guidelines"（v0.0.4 不主动产出，
// 字段保留供未来 fallback 逻辑使用）。
type AugmentCode struct{}

// Name 返回 "augment-code"
func (a *AugmentCode) Name() string { return "augment-code" }

// Plan 计算输出
func (a *AugmentCode) Plan(docs []*parser.Document, cfg *config.Config) (*writer.Plan, error) {
	return protocol.WindsurfStyle{}.Plan(FilterDocs(docs, a.Name()), augmentCodeAdapter, cfg)
}

// augmentCodeAdapter 配置 WindsurfStyle 协议族的 augment-code 变体
//
// skills 原生 .augment/skills/<n>/SKILL.md
// （https://docs.augmentcode.com/using-augment/skills，Auto/Manual/Disabled 三态）。
var augmentCodeAdapter = protocol.Adapter{
	Name:                 "augment-code",
	RulesDir:             ".augment/rules",
	SkillsDir:            ".augment/skills",
	CommandsDir:          ".augment/rules/workflows",
	CommandsFileSuffix:   ".md",
	SingleFileFallback:   ".augment-guidelines", // 老版兼容（v0.0.4 不主动产出）
	ReferencesDir:        "",
	SubagentsDir:         "",
	FallbackDir:          ".augment/rules",
	InjectExplainer:      true,
	InjectStdaiTypeField: true,
	RuleTriggerMode:      protocol.TriggerTrigger,
	SkillsAsRule:         false,
	InjectTypeGlossary:   true,
}
