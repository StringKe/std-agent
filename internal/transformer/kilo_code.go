package transformer

import (
	"github.com/StringKe/std-agent/internal/config"
	"github.com/StringKe/std-agent/internal/parser"
	"github.com/StringKe/std-agent/internal/transformer/protocol"
	"github.com/StringKe/std-agent/internal/writer"
)

func init() {
	Register(&KiloCode{})
}

// KiloCode 是 Kilo Code transformer（kilo.ai 新平台，opencode fork，
// github.com/Kilo-Org/kilocode）
//
// 协议族 D（Clinerules）：
//   - `.kilo/rules/*.md` 不会被自动扫描，必须被项目 `kilo.jsonc` 的
//     `instructions[]` 显式引用才加载（https://kilo.ai/docs/customize/custom-instructions）。
//     因此 Plan 额外产出 kilo.jsonc 的 JSONMerge op 注册 `.kilo/rules/*.md` glob
//     （instructions 官方支持 glob）。kilo.jsonc 已存在且含注释（JSONC）时
//     writer 跳过并 WARN，不破坏用户配置。
//   - skills 原生 `.kilo/skills/<n>/SKILL.md`（https://kilo.ai/docs/customize/skills）
//   - commands 原生 `.kilo/commands/<n>.md`（官方文档复数；源码仍兼容单数）
//   - mode-specific `.kilo/rules-{mode}/` 不产出
type KiloCode struct{}

// Name 返回 "kilo-code"
func (k *KiloCode) Name() string { return "kilo-code" }

// Plan 委托给 Clinerules 协议，并在产出 rules 时追加 kilo.jsonc 注册 op
func (k *KiloCode) Plan(docs []*parser.Document, cfg *config.Config) (*writer.Plan, error) {
	filtered := FilterDocs(docs, k.Name())
	plan, err := protocol.Clinerules{}.Plan(filtered, kiloCodeAdapter, cfg)
	if err != nil {
		return nil, err
	}

	// 只要 .kilo/rules/ 下有产物（rules 或 glossary）就注册 instructions glob
	needsRegister := kiloCodeAdapter.InjectTypeGlossary
	for _, d := range filtered {
		if d.Type == parser.TypeRules {
			needsRegister = true
			break
		}
	}
	if needsRegister {
		plan.Files = append(plan.Files, writer.FileOp{
			Path:      "kilo.jsonc",
			Content:   []byte(`{"instructions":[".kilo/rules/*.md"]}`),
			JSONMerge: true,
		})
	}
	return plan, nil
}

// kiloCodeAdapter 注入 Clinerules 协议族
//
// 与 cline / roo-code 的关键差异：
//   - RulesDir `.kilo/rules`（cline `.clinerules` / roo `.roo/rules`）
//   - SkillsDir `.kilo/skills`（原生 Agent Skills）
//   - CommandsDir `.kilo/commands`（官方文档复数）
//   - SingleFileFallback ""（cline `.clinerules` / roo `.roorules`；kilo 无单文件 fallback）
//   - RulePrefix=nil：与 roo 一致，走子目录组织，无 100/500/900 数字前缀
var kiloCodeAdapter = protocol.Adapter{
	Name:                 "kilo-code",
	RulesDir:             ".kilo/rules",
	SkillsDir:            ".kilo/skills",
	CommandsDir:          ".kilo/commands",
	SingleFileFallback:   "", // kilo 无单文件 fallback
	FallbackDir:          ".kilo/rules",
	InjectExplainer:      true,
	InjectStdaiTypeField: true,
	RulePrefix:           nil, // 无数字前缀（与 cline 100/500/900 不同）
	InjectTypeGlossary:   true,
}
