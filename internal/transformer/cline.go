package transformer

import (
	"github.com/StringKe/std-agent/internal/config"
	"github.com/StringKe/std-agent/internal/parser"
	"github.com/StringKe/std-agent/internal/transformer/protocol"
	"github.com/StringKe/std-agent/internal/writer"
)

func init() {
	Register(&Cline{})
}

// Cline 是 VS Code Cline 扩展 transformer，走 Clinerules 协议族
type Cline struct{}

// Name 返回 "cline"
func (c *Cline) Name() string { return "cline" }

// Plan 委托给 Clinerules 协议
func (c *Cline) Plan(docs []*parser.Document, cfg *config.Config) (*writer.Plan, error) {
	return protocol.Clinerules{}.Plan(FilterDocs(docs, c.Name()), clineAdapter, cfg)
}

// clinePriorityPrefix 把 priority 映射成文件名数字前缀
//   - PriorityHigh   -> "100-"
//   - PriorityNormal -> "500-"
//   - PriorityLow    -> "900-"
func clinePriorityPrefix(d *parser.Document) string {
	switch d.Priority {
	case parser.PriorityHigh:
		return "100-"
	case parser.PriorityLow:
		return "900-"
	}
	return "500-"
}

// clineAdapter 注入 Clinerules 协议族。
//
// 官方 2026-08 仍推荐项目 skills 写到 `.cline/skills/`，同时保留
// `.clinerules/skills/` 与 `.claude/skills/` 作为扫描兼容路径。
var clineAdapter = protocol.Adapter{
	Name:                 "cline",
	RulesDir:             ".clinerules",
	SkillsDir:            ".cline/skills",
	SkillSupportedFields: []string{"name", "description", "license", "compatibility", "metadata"},
	CommandsDir:          ".clinerules/workflows",
	SingleFileFallback:   ".clinerules", // 向后兼容（v0.0.4 默认走子目录）
	FallbackDir:          ".cline",      // 不进 .clinerules/，避免 references 被当 rule 全量加载
	GlobsFieldName:       "paths",
	GlobsFieldFormat:     protocol.GlobsList,
	InjectExplainer:      true,
	InjectStdaiTypeField: true,
	RulePrefix:           clinePriorityPrefix,
	InjectTypeGlossary:   true,
}
