package protocol

import (
	"fmt"
	"path"

	"github.com/StringKe/std-agent/internal/config"
	"github.com/StringKe/std-agent/internal/parser"
	"github.com/StringKe/std-agent/internal/transformer/transformerutil"
	"github.com/StringKe/std-agent/internal/writer"
)

// WindsurfStyle 是协议族 E：Windsurf / Continue.dev / Antigravity 共用的协议。
//
// 共同点：
//   - 无根文件（CLAUDE.md / AGENTS.md 不由本协议产出）
//   - rules 落 RulesDir/<name>.md，trigger 字段（always_on / glob / model_decision /
//     manual）由 frontmatter helper 推断
//   - commands 落 CommandsDir/<name><CommandsFileSuffix>（差异：windsurf=`.md` /
//     antigravity=`.md` / continue=`.prompt.md`）
//   - skills 有原生支持（windsurf）走 SkillsDir/<name>/SKILL.md；
//     不支持（continue / antigravity）按 SkillsAsRule 降级为 rule trigger=manual，
//     或走 BuildDegradedSkillPackage（落 <RulesDir>/skills/<name>/SKILL.md）
//   - references / subagents 走 graceful degradation：
//     <RulesDir>/references/<name>.md + frontmatter std-agent-type
//     <RulesDir>/subagents/<name>.md + frontmatter std-agent-type
//   - glossary 落 <RulesDir>/glossary.md（无根文件 target）
type WindsurfStyle struct{}

// Plan 实现 Protocol。
func (p WindsurfStyle) Plan(docs []*parser.Document, adapter Adapter, cfg *config.Config) (*writer.Plan, error) {
	plan := &writer.Plan{Target: adapter.Name}
	if adapter.Disabled {
		return plan, nil
	}
	if len(docs) == 0 {
		return plan, nil
	}

	rules := transformerutil.FilterByType(docs, parser.TypeRules)
	skills := transformerutil.FilterByType(docs, parser.TypeSkills)
	commands := transformerutil.FilterByType(docs, parser.TypeCommands)
	refs := transformerutil.FilterByType(docs, parser.TypeReferences)
	subs := transformerutil.FilterByType(docs, parser.TypeSubagents)
	transformerutil.SortDocs(rules)
	transformerutil.SortDocs(skills)
	transformerutil.SortDocs(commands)
	transformerutil.SortDocs(refs)
	transformerutil.SortDocs(subs)

	for _, d := range rules {
		// 嵌套目录说明文档：continue-dev 只认固定文件名 rules.md 的 colocated
		// rule（continuedev/continue#6048），NestedFileName 配置后写
		// <NestedPath>/<NestedFileName>（纯 body 无 trigger frontmatter）
		if d.NestedPath != "" && adapter.NestedSupported && adapter.NestedFileName != "" {
			plan.Files = append(plan.Files, p.buildNestedRule(d, adapter, cfg))
			continue
		}
		plan.Files = append(plan.Files, p.buildRule(d, adapter, cfg))
	}
	for _, d := range skills {
		plan.Files = append(plan.Files, p.buildSkill(d, adapter, cfg)...)
	}
	for _, d := range commands {
		plan.Files = append(plan.Files, p.buildCommand(d, adapter, cfg))
	}
	for _, d := range refs {
		plan.Files = append(plan.Files, BuildDegradedFileOp(d, adapter, cfg))
	}
	for _, d := range subs {
		if adapter.SubagentsDir != "" {
			plan.Files = append(plan.Files, p.buildSubagent(d, adapter, cfg))
			continue
		}
		plan.Files = append(plan.Files, BuildDegradedFileOp(d, adapter, cfg))
	}

	if g := p.buildGlossary(adapter, cfg); g != nil {
		plan.Files = append(plan.Files, *g)
	}
	return plan, nil
}

// buildNestedRule 把嵌套目录说明文档写到 <NestedPath>/<NestedFileName>
// （纯 body，无 trigger frontmatter 无 manifest，工具按目录 colocation 语义加载）
func (p WindsurfStyle) buildNestedRule(d *parser.Document, adapter Adapter, cfg *config.Config) writer.FileOp {
	opts := transformerutil.MakeOpts(cfg, adapter.Name, d.Path, false)
	return transformerutil.BuildMarkdownFile(
		path.Join(d.NestedPath, adapter.NestedFileName),
		"", d.Body, opts,
	)
}

// buildRule -> RulesDir/<name>.md，trigger frontmatter 由
// RenderTriggerFrontmatter 推断。MaxBytesPerFile 非零时检测超限。
func (p WindsurfStyle) buildRule(d *parser.Document, adapter Adapter, cfg *config.Config) writer.FileOp {
	// 让 EffectiveApplyTo 处理 target 专属覆盖
	applyTo := transformerutil.EffectiveApplyTo(d, adapter.Name)
	// RenderTriggerFrontmatter 直接读 doc.ApplyTo，这里构造一个浅拷贝把
	// EffectiveApplyTo 结果灌回 ApplyTo 后调用，避免修改原 doc
	dCopy := *d
	dCopy.ApplyTo = applyTo

	body := RenderTriggerFrontmatter(adapter.RuleTriggerMode, &dCopy)
	if adapter.RuleTriggerMode == TriggerNone {
		// 协议族 E 默认 TriggerTrigger；若 caller 误传 TriggerNone，
		// 仍 fallback 到 TriggerTrigger 避免 rule 文件无 frontmatter
		body = RenderTriggerFrontmatter(TriggerTrigger, &dCopy)
	}
	fm := wrapFrontmatter(body)

	opts := transformerutil.MakeOpts(cfg, adapter.Name, d.Path, false)
	op := transformerutil.BuildMarkdownFile(
		transformerutil.FilePath(adapter.RulesDir, d.Name, ".md"),
		fm, d.Body, opts,
	)
	if adapter.MaxBytesPerFile > 0 && len(op.Content) > adapter.MaxBytesPerFile {
		op.Reason = fmt.Sprintf("WARN: rule exceeds %d chars; consider splitting", adapter.MaxBytesPerFile)
	}
	return op
}

// buildSkill 按 adapter 三种模式落 skill：
//   - SkillsDir 非空 -> 原生 SKILL.md 包（windsurf）
//   - SkillsDir 空 + SkillsAsRule=true -> 降级为 rule（continue / antigravity 当前
//     行为：写到 RulesDir/skill-<name>.md trigger=model_decision，与现有
//     transformer 兼容）
//   - SkillsDir 空 + SkillsAsRule=false -> 走 Agent Skills 标准 fallback 包
func (p WindsurfStyle) buildSkill(d *parser.Document, adapter Adapter, cfg *config.Config) []writer.FileOp {
	if adapter.SkillsDir != "" {
		return p.buildNativeSkill(d, adapter, cfg)
	}
	if adapter.SkillsAsRule {
		return []writer.FileOp{p.buildSkillAsRule(d, adapter, cfg)}
	}
	return BuildDegradedSkillPackage(d, adapter, cfg)
}

// buildNativeSkill -> SkillsDir/<name>/SKILL.md，遵循 Agent Skills 规范。
// 字段集由 adapter.SkillSupportedFields 控制：未列出的字段不输出，避免风险。
// Windsurf 文档仅承认 name + description，其他字段空时自动 skip。
func (p WindsurfStyle) buildNativeSkill(d *parser.Document, adapter Adapter, cfg *config.Config) []writer.FileOp {
	skillDir := transformerutil.SkillDir(adapter.SkillsDir, d.Name)

	var fm transformerutil.FmBuilder
	allowed := skillFieldAllowedSet(adapter.SkillSupportedFields)
	addIfAllowed := func(key, val string) {
		if allowed == nil || allowed[key] {
			fm.Add(key, val)
		}
	}
	addIfAllowed("name", d.Name)
	addIfAllowed("description", transformerutil.MergeDescription(d.Description, d.WhenToUse))
	addIfAllowed("license", d.License)
	addIfAllowed("compatibility", d.Compatibility)
	if allowed == nil || allowed["metadata"] {
		fm.AddMap("metadata", d.Metadata)
	}

	opts := transformerutil.MakeOpts(cfg, adapter.Name, d.Path, false)
	skillMd := transformerutil.BuildMarkdownFile(skillDir+"/SKILL.md", fm.String(), d.Body, opts)
	return transformerutil.BuildSkillPackage(skillDir, skillMd, d.SkillFiles)
}

// buildSkillAsRule 把 skill 降级为 rule trigger=model_decision，保持与现有
// continue / antigravity transformer 一致：
//   - continue: .continue/rules/skill-<name>.md（name + description + alwaysApply=false）
//   - antigravity: .agents/rules/skill-<name>.md（trigger=model_decision + description）
func (p WindsurfStyle) buildSkillAsRule(d *parser.Document, adapter Adapter, cfg *config.Config) writer.FileOp {
	var fm transformerutil.FmBuilder
	// 推断走 trigger=model_decision（skill 本来就是按 description 触发）
	desc := d.Description
	if desc == "" {
		desc = "Skill: " + d.Name
	}
	switch adapter.RuleTriggerMode {
	case TriggerTrigger:
		fm.Add("trigger", "model_decision")
		fm.Add("description", desc)
	case TriggerAlwaysApply:
		fm.Add("name", "Skill: "+d.Name)
		fm.Add("description", desc)
		fm.AddBool("alwaysApply", false)
	default:
		fm.Add("description", desc)
	}
	opts := transformerutil.MakeOpts(cfg, adapter.Name, d.Path, false)
	return transformerutil.BuildMarkdownFile(
		transformerutil.FilePath(adapter.RulesDir, "skill-"+d.Name, ".md"),
		fm.String(), d.Body, opts,
	)
}

// buildCommand -> CommandsDir/<name><CommandsFileSuffix>
//
// 若 CommandFrontmatter 非空，按 whitelist 渲染字段；否则不写 frontmatter。
// description 仍可注入到 body 头部（与现有 windsurf / antigravity 行为一致）。
func (p WindsurfStyle) buildCommand(d *parser.Document, adapter Adapter, cfg *config.Config) writer.FileOp {
	suffix := adapter.CommandsFileSuffix
	if suffix == "" {
		suffix = ".md"
	}

	var fm transformerutil.FmBuilder
	if len(adapter.CommandFrontmatter) > 0 {
		allowed := skillFieldAllowedSet(adapter.CommandFrontmatter)
		if allowed["name"] {
			fm.Add("name", d.Name)
		}
		if allowed["description"] {
			fm.Add("description", d.Description)
		}
		if allowed["version"] {
			fm.Add("version", d.Version)
		}
		if allowed["invokable"] {
			fm.AddBool("invokable", true)
		}
	}

	body := d.Body
	if len(adapter.CommandFrontmatter) == 0 && d.Description != "" {
		body = d.Description + "\n\n" + d.Body
	}

	opts := transformerutil.MakeOpts(cfg, adapter.Name, d.Path, false)
	return transformerutil.BuildMarkdownFile(
		transformerutil.FilePath(adapter.CommandsDir, d.Name, suffix),
		fm.String(), body, opts,
	)
}

func (p WindsurfStyle) buildSubagent(d *parser.Document, adapter Adapter, cfg *config.Config) writer.FileOp {
	var fm transformerutil.FmBuilder
	fm.Add("name", d.Name)
	fm.Add("description", d.Description)
	fm.Add("model", d.Model)
	fm.Add("mode", adapter.SubagentMode)
	fm.AddList("tools", d.AllowedTools)
	fm.AddList("disallowedTools", d.DisallowedTools)
	opts := transformerutil.MakeOpts(cfg, adapter.Name, d.Path, false)
	return transformerutil.BuildMarkdownFile(
		transformerutil.FilePath(adapter.SubagentsDir, d.Name, ".md"),
		fm.String(), d.Body, opts,
	)
}

// buildGlossary 在无根文件 target 下，把 std-agent 类型速查段落到
// <RulesDir>/glossary.md。adapter.InjectTypeGlossary && cfg.InjectTypeGlossary
// 均开启时才输出。RulesDir 为空时跳过。
func (p WindsurfStyle) buildGlossary(adapter Adapter, cfg *config.Config) *writer.FileOp {
	if !adapter.InjectTypeGlossary || cfg == nil || !cfg.InjectTypeGlossary {
		return nil
	}
	if adapter.RulesDir == "" {
		return nil
	}
	body := RenderGlossaryFor(adapter, cfg)
	if body == "" {
		return nil
	}
	var fm transformerutil.FmBuilder
	if adapter.InjectStdaiTypeField {
		fm.Add("std-agent-type", "glossary")
	}
	opts := transformerutil.MakeOpts(cfg, adapter.Name, "", false)
	op := transformerutil.BuildMarkdownFile(
		path.Join(adapter.RulesDir, "glossary.md"),
		fm.String(), body, opts,
	)
	return &op
}

// skillFieldAllowedSet 把白名单 slice 转 set。nil / empty 返回 nil（表示
// "全部允许"，由调用方判断）。
func skillFieldAllowedSet(fields []string) map[string]bool {
	if len(fields) == 0 {
		return nil
	}
	m := make(map[string]bool, len(fields))
	for _, f := range fields {
		m[f] = true
	}
	return m
}

// wrapFrontmatter 把 RenderTriggerFrontmatter 输出（无 "---" 包裹）补成
// 完整 frontmatter 块。空串原样返回（BuildMarkdownFile 跳过 frontmatter 段）。
func wrapFrontmatter(body string) string {
	if body == "" {
		return ""
	}
	return "---\n" + body + "---\n"
}
