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

// KiloCode 是 Kilo Code transformer（Cline fork，github.com/Kilo-Org/kilocode）
//
// 协议族 D（Clinerules）：
//   - 主目录 `.kilo/rules/` 自动加载所有 .md（kilo.jsonc 的 instructions[] 引用）
//   - 向后兼容 `.kilocode/rules/`（旧路径，v0.0.4 默认只写 `.kilo/rules/`）
//   - mode-specific `.kilo/rules-{mode}/`
//   - 与 cline 的差异：kilo 走子目录组织（不需要数字前缀）；无单文件 fallback
//
// v0.0.4 备注：spec §2.4 提到的 kilo.jsonc 写入（AdditionalConfigWriter hook）
// 暂未实现，留待 v0.0.5。当前只输出 `.kilo/rules/` 子目录，用户需手动在
// kilo.jsonc 的 instructions[] 中引用对应文件。
type KiloCode struct{}

// Name 返回 "kilo-code"
func (k *KiloCode) Name() string { return "kilo-code" }

// Plan 委托给 Clinerules 协议
func (k *KiloCode) Plan(docs []*parser.Document, cfg *config.Config) (*writer.Plan, error) {
	return protocol.Clinerules{}.Plan(FilterDocs(docs, k.Name()), kiloCodeAdapter, cfg)
}

// kiloCodeAdapter 注入 Clinerules 协议族
//
// 与 cline / roo-code 的关键差异：
//   - RulesDir `.kilo/rules`（cline `.clinerules` / roo `.roo/rules`）
//   - SingleFileFallback ""（cline `.clinerules` / roo `.roorules`；kilo 无单文件 fallback）
//   - RulePrefix=nil：与 roo 一致，走子目录组织，无 100/500/900 数字前缀
var kiloCodeAdapter = protocol.Adapter{
	Name:                 "kilo-code",
	RulesDir:             ".kilo/rules",
	CommandsDir:          ".kilo/rules/workflows",
	SingleFileFallback:   "", // kilo 无单文件 fallback
	FallbackDir:          ".kilo/rules",
	InjectExplainer:      true,
	InjectStdaiTypeField: true,
	RulePrefix:           nil, // 无数字前缀（与 cline 100/500/900 不同）
	InjectTypeGlossary:   true,
}
