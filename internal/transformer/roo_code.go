package transformer

import (
	"github.com/StringKe/std-agent/internal/config"
	"github.com/StringKe/std-agent/internal/parser"
	"github.com/StringKe/std-agent/internal/transformer/protocol"
	"github.com/StringKe/std-agent/internal/writer"
)

func init() {
	Register(&RooCode{})
}

// RooCode 是 Roo Code transformer（Cline fork，18k stars，github.com/RooCodeInc/Roo-Code）
//
// 协议族 D（Clinerules）：
//   - 主目录 `.roo/rules/` 自动加载所有 .md
//   - 单文件 `.roorules` 作为 v0.0.4 之前老项目的向后兼容 fallback
//   - 与 cline 的差异：roo-code 用子目录组织（不需要数字前缀）；
//     cline 用 100- / 500- / 900- 前缀决定加载顺序
type RooCode struct{}

// Name 返回 "roo-code"
func (r *RooCode) Name() string { return "roo-code" }

// Plan 委托给 Clinerules 协议
func (r *RooCode) Plan(docs []*parser.Document, cfg *config.Config) (*writer.Plan, error) {
	return protocol.Clinerules{}.Plan(FilterDocs(docs, r.Name()), rooCodeAdapter, cfg)
}

// rooCodeAdapter 注入 Clinerules 协议族（spec v3 §2.4）
//
// 与 cline 的差异：
//   - RulesDir `.roo/rules` 而非 `.clinerules`。注意 roo 不递归扫描 rules 子目录
//     （只读顶层 .md），skills / commands 必须走各自原生目录
//   - SkillsDir `.roo/skills`（2026-05-15 GA，https://docs.roocode.com/features/skills）
//   - CommandsDir `.roo/commands`（原生 slash commands，frontmatter
//     description / argument-hint，https://roocodeinc.github.io/Roo-Code/features/slash-commands/）
//   - SingleFileFallback `.roorules`（cline 是 `.clinerules`）
//   - RulePrefix=nil：roo-code 走子目录组织，无 100/500/900 数字前缀
var rooCodeAdapter = protocol.Adapter{
	Name:                 "roo-code",
	RulesDir:             ".roo/rules",
	SkillsDir:            ".roo/skills",
	CommandsDir:          ".roo/commands",
	SingleFileFallback:   ".roorules", // 向后兼容老 roo 项目
	FallbackDir:          ".roo/rules",
	GlobsFieldName:       "paths",
	GlobsFieldFormat:     protocol.GlobsList,
	InjectExplainer:      true,
	InjectStdaiTypeField: true,
	RulePrefix:           nil, // roo 用子目录组织，无数字前缀（与 cline 100/500/900 不同）
	InjectTypeGlossary:   true,
}
