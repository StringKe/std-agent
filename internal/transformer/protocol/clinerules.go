package protocol

import (
	"path"

	"github.com/StringKe/std-agent/internal/config"
	"github.com/StringKe/std-agent/internal/parser"
	"github.com/StringKe/std-agent/internal/transformer/transformerutil"
	"github.com/StringKe/std-agent/internal/writer"
)

// Clinerules 是 cline / roo-code / kilo-code 共用协议族。
//
// 三个 target 共同特征：
//   - 无根文件（不写 CLAUDE.md / AGENTS.md）。AI 直接读取 RulesDir 下所有 .md
//   - rules 通过文件名数字前缀（如 100- / 500- / 900-）暗示加载优先级
//     （cline / kilo 用此约定；roo-code 走 .roo/rules/ 子目录无前缀）
//   - workflows / commands 落到 RulesDir/workflows/<name>.md
//   - skills / references / subagents 走 graceful degradation（target 原生不支持）
//
// 关键 adapter 字段：
//   - RulesDir 必填（如 ".clinerules" / ".roo/rules" / ".kilo/rules"）
//   - RulePrefix（func(*parser.Document) string）：rule 文件名数字前缀生成器；
//     nil 时无前缀（roo / kilo 走 .roo/rules/ 不加数字前缀的情况）
//   - GlobsFieldName="paths"（cline 方言）；通过 RenderGlobs 渲染
//   - SingleFileFallback：v0.0.4 默认不触发（保留字段供未来兼容老 cline 项目使用）
//
// graceful degradation 落点（spec §2.3）：
//   - skills:     <RulesDir>/skills/<n>/SKILL.md（Agent Skills 标准形式）
//   - references: <RulesDir>/references/<n>.md
//   - subagents:  <RulesDir>/subagents/<n>.md
//   - glossary:   <RulesDir>/glossary.md（无前缀，子目录隔离 + frontmatter std-agent-type: glossary）
type Clinerules struct{}

// Plan 实现 Protocol.Plan
func (p Clinerules) Plan(docs []*parser.Document, adapter Adapter, cfg *config.Config) (*writer.Plan, error) {
	plan := &writer.Plan{Target: adapter.Name}
	if adapter.Disabled {
		return plan, nil
	}

	rules := transformerutil.FilterByType(docs, parser.TypeRules)
	skills := transformerutil.FilterByType(docs, parser.TypeSkills)
	commands := transformerutil.FilterByType(docs, parser.TypeCommands)
	references := transformerutil.FilterByType(docs, parser.TypeReferences)
	subagents := transformerutil.FilterByType(docs, parser.TypeSubagents)
	transformerutil.SortDocs(rules)
	transformerutil.SortDocs(skills)
	transformerutil.SortDocs(commands)
	transformerutil.SortDocs(references)
	transformerutil.SortDocs(subagents)

	// glossary 落到 <RulesDir>/glossary.md（无根文件，子目录隔离方案）
	if adapter.InjectTypeGlossary && cfg != nil && cfg.InjectTypeGlossary && adapter.RulesDir != "" {
		plan.Files = append(plan.Files, p.buildGlossary(adapter, cfg))
	}

	for _, d := range rules {
		plan.Files = append(plan.Files, p.buildRule(d, adapter, cfg))
	}
	for _, d := range commands {
		plan.Files = append(plan.Files, p.buildWorkflow(d, adapter, cfg))
	}

	// skills：SkillsDir 非空走原生 Agent Skills 包（cline `.cline/skills/`、
	// roo `.roo/skills/`、kilo `.kilo/skills/`）；为空走 fallback
	// （`.clinerules/skills/` 仍是 Cline 官方备用扫描路径）
	for _, d := range skills {
		if adapter.SkillsDir != "" {
			plan.Files = append(plan.Files, BuildNativeSkillPackage(d, adapter, cfg)...)
		} else {
			plan.Files = append(plan.Files, BuildDegradedSkillPackage(d, adapter, cfg)...)
		}
	}
	// references fallback -> <RulesDir>/references/<n>.md
	for _, d := range references {
		plan.Files = append(plan.Files, BuildDegradedFileOp(d, adapter, cfg))
	}
	// subagents fallback -> <RulesDir>/subagents/<n>.md
	for _, d := range subagents {
		plan.Files = append(plan.Files, BuildDegradedFileOp(d, adapter, cfg))
	}

	return plan, nil
}

// buildRule 写 rule 文件。
// 路径：<RulesDir>/<prefix><name>.md，prefix 由 adapter.RulePrefix 决定（nil=无前缀）。
// frontmatter 写 paths 字段（cline 方言；adapter.GlobsFieldName 控制）。
func (p Clinerules) buildRule(d *parser.Document, adapter Adapter, cfg *config.Config) writer.FileOp {
	prefix := ""
	if adapter.RulePrefix != nil {
		prefix = adapter.RulePrefix(d)
	}
	fileName := prefix + d.Name + ".md"
	fullPath := path.Join(adapter.RulesDir, fileName)

	fm := p.buildRuleFrontmatter(d, adapter)
	opts := transformerutil.MakeOpts(cfg, adapter.Name, d.Path, false)
	return transformerutil.BuildMarkdownFile(fullPath, fm, d.Body, opts)
}

// buildRuleFrontmatter 拼装 rule frontmatter（含 globs / trigger 字段）。
// 返回完整 "---\n...\n---\n" 段；无字段时返回 ""。
func (p Clinerules) buildRuleFrontmatter(d *parser.Document, adapter Adapter) string {
	var fm transformerutil.FmBuilder
	if adapter.SupportsDescription && d.Description != "" {
		fm.Add("description", d.Description)
	}
	globsField := adapter.GlobsFieldName
	if globsField != "" {
		vals := transformerutil.EffectiveApplyTo(d, adapter.Name)
		if len(vals) > 0 {
			switch adapter.GlobsFieldFormat {
			case GlobsCommaString:
				fm.Add(globsField, transformerutil.CommaJoin(vals))
			default:
				fm.AddList(globsField, vals)
			}
		}
	}
	if adapter.RuleTriggerMode == TriggerAlwaysApply && adapter.SupportsAlwaysApply && d.AlwaysApply {
		fm.AddBool("alwaysApply", true)
	}
	return fm.String()
}

// buildWorkflow 写 workflow / command 文件
//
//   - CommandsDir 非空（roo `.roo/commands/` / kilo `.kilo/commands/` 原生 slash
//     commands）：<CommandsDir>/<name>.md，frontmatter 写 description / argument-hint
//   - CommandsDir 为空（cline 无原生 commands）：<RulesDir>/workflows/<name>.md，
//     description 非空时 prepend 到 body 头
func (p Clinerules) buildWorkflow(d *parser.Document, adapter Adapter, cfg *config.Config) writer.FileOp {
	opts := transformerutil.MakeOpts(cfg, adapter.Name, d.Path, false)
	if adapter.CommandsDir != "" {
		var fm transformerutil.FmBuilder
		fm.Add("description", d.Description)
		fm.Add("argument-hint", d.ArgumentHint)
		fullPath := transformerutil.FilePath(adapter.CommandsDir, d.Name, ".md")
		return transformerutil.BuildMarkdownFile(fullPath, fm.String(), d.Body, opts)
	}
	body := d.Body
	if d.Description != "" {
		body = d.Description + "\n\n" + d.Body
	}
	fullPath := transformerutil.FilePath(path.Join(adapter.RulesDir, "workflows"), d.Name, ".md")
	return transformerutil.BuildMarkdownFile(fullPath, "", body, opts)
}

// buildGlossary 把 std-agent 类型速查写到 <RulesDir>/glossary.md
// frontmatter 加 std-agent-type: glossary 让 AI 识别
func (p Clinerules) buildGlossary(adapter Adapter, cfg *config.Config) writer.FileOp {
	var fm transformerutil.FmBuilder
	fm.Add("std-agent-type", "glossary")
	body := glossaryMarkdown
	opts := transformerutil.MakeOpts(cfg, adapter.Name, "", false)
	return transformerutil.BuildMarkdownFile(
		path.Join(adapter.RulesDir, "glossary.md"),
		fm.String(), body, opts,
	)
}
