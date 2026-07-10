package transformer

import (
	"github.com/StringKe/std-agent/internal/config"
	"github.com/StringKe/std-agent/internal/parser"
	"github.com/StringKe/std-agent/internal/transformer/protocol"
	"github.com/StringKe/std-agent/internal/writer"
)

func init() {
	Register(&ContinueDev{})
}

// ContinueDev 是 Continue.dev VS Code/JetBrains 扩展 transformer
type ContinueDev struct{}

// Name 返回 "continue-dev"（避免与 Go keyword `continue` 视觉冲突）
func (c *ContinueDev) Name() string { return "continue-dev" }

// Plan 委托 WindsurfStyle 协议族
func (c *ContinueDev) Plan(docs []*parser.Document, cfg *config.Config) (*writer.Plan, error) {
	return protocol.WindsurfStyle{}.Plan(FilterDocs(docs, c.Name()), continueAdapter, cfg)
}

// continueAdapter 配置 WindsurfStyle 协议族的 continue-dev 变体。
//
// 与 windsurf 差异：
//   - 无原生 skills -> SkillsDir 留空，SkillsAsRule=false 走 Agent Skills 标准 fallback
//     落到 .continue/rules/skills/<name>/SKILL.md
//   - commands 走 .continue/prompts/<name>.prompt.md，frontmatter 含
//     name/description/version/invokable=true
//   - 无根文件（continue-dev 主入口就是 .continue/rules/ 多文件）
//
// RuleTriggerMode=TriggerTrigger 与 windsurf 一致（continue 支持 globs + 推断
// alwaysApply / model_decision / manual 语义）。
var continueAdapter = protocol.Adapter{
	Name:                 "continue-dev",
	RulesDir:             ".continue/rules",
	SkillsDir:            "", // 无原生 skills，走 BuildDegradedSkillPackage
	CommandsDir:          ".continue/prompts",
	CommandsFileSuffix:   ".prompt.md",
	CommandFrontmatter:   []string{"name", "description", "version", "invokable"},
	FallbackDir:          ".continue/rules",
	InjectExplainer:      true,
	InjectStdaiTypeField: true,
	SkillsAsRule:         false, // 走 Agent Skills 标准包
	RuleTriggerMode:      protocol.TriggerTrigger,
	InjectTypeGlossary:   true,
}
