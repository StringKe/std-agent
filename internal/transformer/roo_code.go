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

// RooCode 是 Roo Code transformer（Cline fork；官方扩展已于 2026-05-15 归档）。
//
// 协议族 D（Clinerules）：
//   - 主目录 `.roo/rules/` 递归加载所有 .md
//   - 单文件 `.roorules` 作为目录为空时的 fallback
//   - 与 cline 的差异：roo-code 默认不用 100/500/900 优先级协议；
//     官方示例可用 `01-` / `02-` 做字母序，cline 用数字前缀决定加载顺序
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
//   - RulesDir `.roo/rules` 而非 `.clinerules`。目录内部递归扫描；
//     `enableSubfolderRules`（默认 false）管的是另一套子目录 `.roo/rules/`，
//     不是本目录内部递归。skills / commands 走各自原生目录
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
	FallbackDir:          ".roo",      // 不进 .roo/rules/，避免 references 被递归当 rule 加载
	GlobsFieldName:       "paths",
	GlobsFieldFormat:     protocol.GlobsList,
	InjectExplainer:      true,
	InjectStdaiTypeField: true,
	RulePrefix:           nil, // roo 用子目录组织，无数字前缀（与 cline 100/500/900 不同）
	InjectTypeGlossary:   true,
}
