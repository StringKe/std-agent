package protocol

import (
	"fmt"
	"path"
	"strings"

	"std-ai/internal/config"
	"std-ai/internal/parser"
	"std-ai/internal/transformer/transformerutil"
	"std-ai/internal/writer"
)

// AgentsMD 是 AGENTS.md 系协议族实现。
//
// 覆盖 target：codex / opencode / antigravity / crush / amp / warp / factory /
// qwen-code / pi / jules / grok-cli。Adapter 通过零值与字段组合控制变体（如
// RootFileName="CRUSH.md" / RuleTriggerMode=TriggerTrigger /
// InjectCommandsToRoot=true / CommandsAsSkillPrefix="cmd-"）。
//
// Cross-cutting 行为：
//   - InjectCommandsToRoot：commands 段 inject 到 root FileOp.Content 的 footer
//     marker 之前（codex 风格），让 aider / 其他读 AGENTS.md 的工具也能看到 slash
//     命令
//   - RuleTriggerMode!=TriggerNone：nonRoot rules 子文件渲染 trigger frontmatter
//     （antigravity 双协议场景）
//   - CommandsAsSkillPrefix 非空：command 降级写为 skill 到 SkillsDir，name 前
//     加前缀（codex 把 commands 写为 .agents/skills/cmd-<n>/SKILL.md）
//   - InjectTypeGlossary=true：root body 头部 prepend 类型速查段
//   - nested root：写到 d.NestedPath/<RootFileName>，无 manifest 无 glossary
//   - graceful degradation：Skills/CommandsDir/ReferencesDir/SubagentsDir 为空
//     则走 BuildDegradedFileOp（rule-equivalent）或 BuildDegradedSkillPackage
//     （Agent Skills 标准）
//   - Disabled=true：直接返回 &writer.Plan{Target: Name}, nil
//   - MaxBytesPerFile>0：rule body 超阈值时写到独立文件（每条独立 rule，
//     不合并），保持单文件可读
type AgentsMD struct{}

// Plan 按 Adapter 计算 AGENTS.md 系 target 的输出
func (p AgentsMD) Plan(docs []*parser.Document, a Adapter, cfg *config.Config) (*writer.Plan, error) {
	plan := &writer.Plan{Target: a.Name}
	if a.Disabled {
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

	// nested root rules 单独处理（无 manifest，无 glossary）
	topRules, nestedRoots := transformerutil.PartitionNested(rules)
	for _, d := range nestedRoots {
		plan.Files = append(plan.Files, p.buildNestedRoot(d, a, cfg))
	}

	// rules：拆 root / nonRoot
	roots, nonRoot := transformerutil.PartitionRoot(topRules)

	// RulesDir 为空时，nonRoot rules 默认 inline 到 root body；MaxBytesPerFile>0
	// 时把超阈值条目 split 到 FallbackDir（amp/warp 风格的可选 spill）
	var inlineNonRoot, spilled []*parser.Document
	if a.RulesDir == "" && a.MaxBytesPerFile > 0 {
		inlineNonRoot, spilled = splitBySize(nonRoot, a.MaxBytesPerFile)
	} else {
		inlineNonRoot = nonRoot
	}

	// root 文件（可能含 manifest / glossary / commands inject）
	if a.RootFileName != "" {
		rootOp := p.buildRoot(roots, inlineNonRoot, a, cfg)
		if a.InjectCommandsToRoot && len(commands) > 0 {
			block := BuildCommandsSection(commands)
			rootOp.Content = InjectBeforeFooter(rootOp.Content, block)
		}
		plan.Files = append(plan.Files, rootOp)
	}

	// nonRoot rules fan-out 到 RulesDir（codex 风格：所有 nonRoot 都进 RulesDir）
	if a.RulesDir != "" {
		for _, d := range nonRoot {
			plan.Files = append(plan.Files, p.buildRuleFile(d, a, cfg))
		}
	}
	// spilled rule（仅 RulesDir 为空时产生）写到 FallbackDir/<name>.md
	for _, d := range spilled {
		plan.Files = append(plan.Files, p.buildSpilledRuleFile(d, a, cfg))
	}

	// skills
	for _, d := range skills {
		plan.Files = append(plan.Files, p.planSkill(d, a, cfg)...)
	}

	// commands（若未 inject 到 root）
	if !a.InjectCommandsToRoot {
		for _, d := range commands {
			plan.Files = append(plan.Files, p.planCommand(d, a, cfg)...)
		}
	}

	// references
	for _, d := range refs {
		plan.Files = append(plan.Files, p.planReference(d, a, cfg)...)
	}

	// subagents
	for _, d := range subs {
		plan.Files = append(plan.Files, p.planSubagent(d, a, cfg)...)
	}

	return plan, nil
}

// buildRoot 构造根 AGENTS.md（或 CRUSH.md / QWEN.md 等）。
//
// 内容顺序：glossary（可选）+ root rule body（项目总结）+ inline nonRoot rule body
// （仅 RulesDir 为空时）+ manifest 段（仅 RulesDir 非空时，含 nonRoot rule 索引）
func (p AgentsMD) buildRoot(roots, inlineNonRoot []*parser.Document, a Adapter, cfg *config.Config) writer.FileOp {
	opts := transformerutil.MakeOpts(cfg, a.Name, "", true)
	var body strings.Builder
	if g := RenderGlossaryFor(a); g != "" {
		body.WriteString(g)
		if !strings.HasSuffix(g, "\n") {
			body.WriteString("\n")
		}
		body.WriteString("\n")
	}
	if len(roots) > 0 {
		body.WriteString(transformerutil.RenderRootBody(roots))
	} else if body.Len() == 0 && len(inlineNonRoot) == 0 {
		// 无 root rule 且无 glossary 且无 inline -> 给个最小占位标题
		body.WriteString("# Project AGENTS Manifest\n\n")
	}
	if a.RulesDir != "" && len(inlineNonRoot) > 0 {
		title := a.ManifestSection
		if title == "" {
			title = "Reference Rules"
		}
		rulesDir := a.RulesDir
		body.WriteString(transformerutil.BuildRuleManifestSection(
			title, a.Name, inlineNonRoot,
			func(d *parser.Document) string { return transformerutil.FilePath(rulesDir, d.Name, ".md") },
			false,
		))
	} else if a.RulesDir == "" && len(inlineNonRoot) > 0 {
		// 无 RulesDir -> nonRoot rules 直接 inline 到 root（amp / warp 风格）
		body.WriteString("\n")
		body.WriteString(transformerutil.JoinAGENTSStyle("", inlineNonRoot))
	}
	op := transformerutil.BuildMarkdownFile(a.RootFileName, "", body.String(), opts)
	op.IsRoot = true
	return op
}

// buildSpilledRuleFile 把超 MaxBytesPerFile 阈值的 rule 写到独立文件，
// 用于 RulesDir 为空但有 byte 限制的 adapter（amp / warp 等）。
//
// 路径：<FallbackDir or "rules">/<name>.md（无前缀，无 frontmatter，纯 body）
func (p AgentsMD) buildSpilledRuleFile(d *parser.Document, a Adapter, cfg *config.Config) writer.FileOp {
	opts := transformerutil.MakeOpts(cfg, a.Name, d.Path, false)
	dir := a.FallbackDir
	if dir == "" {
		dir = "rules"
	}
	body := d.Body
	if d.Description != "" {
		body = d.Description + "\n\n" + d.Body
	}
	return transformerutil.BuildMarkdownFile(
		transformerutil.FilePath(dir, d.Name, ".md"),
		"", body, opts,
	)
}

// buildNestedRoot 嵌套子目录 root：写到 <NestedPath>/<RootFileName>，无 manifest，无 glossary
func (p AgentsMD) buildNestedRoot(d *parser.Document, a Adapter, cfg *config.Config) writer.FileOp {
	opts := transformerutil.MakeOpts(cfg, a.Name, d.Path, false)
	rootName := a.RootFileName
	if rootName == "" {
		rootName = "AGENTS.md"
	}
	op := transformerutil.BuildMarkdownFile(
		path.Join(d.NestedPath, rootName),
		"", d.Body, opts,
	)
	op.IsRoot = true
	return op
}

// buildRuleFile fan-out 单条 nonRoot rule 到 RulesDir/<name>.md。
//
// RuleTriggerMode != TriggerNone 时按 trigger 模式渲染 frontmatter；否则纯 body + description 前置。
func (p AgentsMD) buildRuleFile(d *parser.Document, a Adapter, cfg *config.Config) writer.FileOp {
	opts := transformerutil.MakeOpts(cfg, a.Name, d.Path, false)
	body := d.Body

	var frontmatter string
	if a.RuleTriggerMode != TriggerNone {
		// 走 trigger 渲染（antigravity / windsurf 风格）
		trig := RenderTriggerFrontmatter(a.RuleTriggerMode, d)
		if trig != "" {
			frontmatter = "---\n" + trig + "---\n"
		}
	} else if a.GlobsFieldName != "" {
		applyTo := transformerutil.EffectiveApplyTo(d, a.Name)
		if g := RenderGlobs(a.GlobsFieldName, a.GlobsFieldFormat, applyTo); g != "" {
			frontmatter = "---\n" + g + "---\n"
		}
	}

	if frontmatter == "" && d.Description != "" {
		// 无 frontmatter 时把 description 写到 body 头，避免信息丢失
		body = d.Description + "\n\n" + d.Body
	}

	op := transformerutil.BuildMarkdownFile(
		transformerutil.FilePath(a.RulesDir, d.Name, ".md"),
		frontmatter, body, opts,
	)
	if a.MaxBytesPerFile > 0 && len(op.Content) > a.MaxBytesPerFile {
		op.Reason = fmt.Sprintf("WARN: rule exceeds %d bytes; consider splitting", a.MaxBytesPerFile)
	}
	return op
}

// planSkill 输出单条 skill。
//
//   - SkillsAsRule=true：skill 降级为 rule 文件（antigravity 系）
//   - SkillsDir 非空：原生 Agent Skills 标准 <SkillsDir>/<name>/SKILL.md
//   - SkillsDir 为空：graceful degradation -> BuildDegradedSkillPackage
func (p AgentsMD) planSkill(d *parser.Document, a Adapter, cfg *config.Config) []writer.FileOp {
	if a.SkillsAsRule {
		return []writer.FileOp{p.buildSkillAsRule(d, a, cfg)}
	}
	if a.SkillsDir != "" {
		return p.buildSkillPackage(d, a, cfg)
	}
	return BuildDegradedSkillPackage(d, a, cfg)
}

// buildSkillPackage 在 SkillsDir 下输出 Agent Skills 标准包。
// 受 adapter.SkillSupportedFields 白名单约束渲染 frontmatter；为空时取默认字段集。
func (p AgentsMD) buildSkillPackage(d *parser.Document, a Adapter, cfg *config.Config) []writer.FileOp {
	skillDir := transformerutil.SkillDir(a.SkillsDir, d.Name)
	fm := buildSkillFrontmatter(d, a)
	opts := transformerutil.MakeOpts(cfg, a.Name, d.Path, false)
	skillMd := transformerutil.BuildMarkdownFile(skillDir+"/SKILL.md", fm, d.Body, opts)
	return transformerutil.BuildSkillPackage(skillDir, skillMd, d.SkillFiles)
}

// buildSkillFrontmatter 按 SkillSupportedFields 白名单生成 frontmatter
//
// 空白名单 -> 默认 Agent Skills 标准字段集（name / description / license /
// compatibility / metadata）。白名单非空时仅渲染白名单字段。
func buildSkillFrontmatter(d *parser.Document, a Adapter) string {
	allowed := a.SkillSupportedFields
	if len(allowed) == 0 {
		allowed = []string{"name", "description", "license", "compatibility", "metadata"}
	}
	in := func(k string) bool {
		for _, v := range allowed {
			if v == k {
				return true
			}
		}
		return false
	}
	var fm transformerutil.FmBuilder
	if in("name") {
		fm.Add("name", d.Name)
	}
	if in("description") {
		// 不支持 when_to_use 的 target 把 when_to_use 拼到 description 末尾
		desc := d.Description
		if !in("when_to_use") {
			desc = transformerutil.MergeDescription(d.Description, d.WhenToUse)
		}
		fm.Add("description", desc)
	}
	if in("when_to_use") {
		fm.Add("when_to_use", d.WhenToUse)
	}
	if in("license") {
		fm.Add("license", d.License)
	}
	if in("compatibility") {
		fm.Add("compatibility", d.Compatibility)
	}
	if in("metadata") {
		fm.AddMap("metadata", d.Metadata)
	}
	return fm.String()
}

// buildSkillAsRule 把 skill 降级为 model_decision rule（antigravity）
func (p AgentsMD) buildSkillAsRule(d *parser.Document, a Adapter, cfg *config.Config) writer.FileOp {
	var fm transformerutil.FmBuilder
	fm.Add("trigger", "model_decision")
	desc := d.Description
	if desc == "" {
		desc = "Skill: " + d.Name
	}
	fm.Add("description", desc)
	opts := transformerutil.MakeOpts(cfg, a.Name, d.Path, false)
	return transformerutil.BuildMarkdownFile(
		transformerutil.FilePath(a.RulesDir, "skill-"+d.Name, ".md"),
		fm.String(), d.Body, opts,
	)
}

// planCommand 输出单条 command。
//
//   - CommandsAsSkillPrefix 非空：command 降级为 skill 写到 SkillsDir，name 前加前缀
//   - CommandsDir 非空：原生 markdown 写到 <CommandsDir>/<name>.md
//   - 两者都空：graceful degradation -> BuildDegradedFileOp
func (p AgentsMD) planCommand(d *parser.Document, a Adapter, cfg *config.Config) []writer.FileOp {
	if a.CommandsAsSkillPrefix != "" && a.SkillsDir != "" {
		return []writer.FileOp{p.buildCommandAsSkill(d, a, cfg)}
	}
	if a.CommandsDir != "" {
		return []writer.FileOp{p.buildCommandFile(d, a, cfg)}
	}
	return []writer.FileOp{BuildDegradedFileOp(d, a, cfg)}
}

// buildCommandFile 原生 commands：写到 <CommandsDir>/<name>.md
func (p AgentsMD) buildCommandFile(d *parser.Document, a Adapter, cfg *config.Config) writer.FileOp {
	var fm transformerutil.FmBuilder
	fm.Add("description", d.Description)
	fm.Add("argument-hint", d.ArgumentHint)
	fm.AddList("tools", d.AllowedTools)
	fm.Add("model", d.Model)
	opts := transformerutil.MakeOpts(cfg, a.Name, d.Path, false)
	body := d.Body
	if d.Description != "" && fm.String() == "" {
		body = d.Description + "\n\n" + d.Body
	}
	return transformerutil.BuildMarkdownFile(
		transformerutil.FilePath(a.CommandsDir, d.Name, ".md"),
		fm.String(), body, opts,
	)
}

// buildCommandAsSkill 把 command 降级为 skill：写到 <SkillsDir>/<prefix><name>/SKILL.md
// description 加 slash 调用 hint 让模型理解这是 command-flavor skill
func (p AgentsMD) buildCommandAsSkill(d *parser.Document, a Adapter, cfg *config.Config) writer.FileOp {
	skillName := a.CommandsAsSkillPrefix + d.Name
	var fm transformerutil.FmBuilder
	fm.Add("name", skillName)
	desc := d.Description
	if desc == "" {
		desc = "Slash command: " + d.Name
	}
	desc += " (Invoke when user types /" + d.Name + " or asks to run " + d.Name + ")"
	fm.Add("description", desc)
	fm.Add("model", d.Model)
	opts := transformerutil.MakeOpts(cfg, a.Name, d.Path, false)
	return transformerutil.BuildMarkdownFile(
		path.Join(a.SkillsDir, skillName, "SKILL.md"),
		fm.String(), d.Body, opts,
	)
}

// planReference 输出 reference。
//
//   - ReferencesDir 非空：原生 markdown 写到 <ReferencesDir>/<name>.md
//   - SkillsDir 非空（且支持 Agent Skills 的 target）：走 BuildDegradedSkillPackage
//     （references 作为 SKILL.md 子目录形式 fan-out）
//   - 否则 BuildDegradedFileOp（rule-equivalent fallback）
func (p AgentsMD) planReference(d *parser.Document, a Adapter, cfg *config.Config) []writer.FileOp {
	if a.ReferencesDir != "" {
		return []writer.FileOp{p.buildReferenceFile(d, a, cfg)}
	}
	if a.SkillsDir != "" {
		return BuildDegradedSkillPackage(d, a, cfg)
	}
	return []writer.FileOp{BuildDegradedFileOp(d, a, cfg)}
}

// buildReferenceFile 原生 references：写到 <ReferencesDir>/<name>.md
func (p AgentsMD) buildReferenceFile(d *parser.Document, a Adapter, cfg *config.Config) writer.FileOp {
	var fm transformerutil.FmBuilder
	if d.Description != "" {
		fm.Add("description", d.Description)
	}
	opts := transformerutil.MakeOpts(cfg, a.Name, d.Path, false)
	return transformerutil.BuildMarkdownFile(
		transformerutil.FilePath(a.ReferencesDir, d.Name, ".md"),
		fm.String(), d.Body, opts,
	)
}

// planSubagent 输出 subagent。
//
//   - SubagentsDir 非空：原生 markdown 写到 <SubagentsDir>/<name>.md
//   - SubagentInvokeCmd 非空：CLI 调用降级（形态 B），通过 BuildDegradedFileOp
//     渲染 shell 调用 body
//   - 都空：路径降级（形态 A），BuildDegradedFileOp 写到 FallbackDir/subagents/
func (p AgentsMD) planSubagent(d *parser.Document, a Adapter, cfg *config.Config) []writer.FileOp {
	if a.SubagentsDir != "" {
		return []writer.FileOp{p.buildSubagentFile(d, a, cfg)}
	}
	return []writer.FileOp{BuildDegradedFileOp(d, a, cfg)}
}

// buildSubagentFile 原生 subagents：写到 <SubagentsDir>/<name>.md，frontmatter name/description/model/tools
func (p AgentsMD) buildSubagentFile(d *parser.Document, a Adapter, cfg *config.Config) writer.FileOp {
	var fm transformerutil.FmBuilder
	fm.Add("name", d.Name)
	fm.Add("description", d.Description)
	fm.Add("model", d.Model)
	fm.AddList("tools", d.AllowedTools)
	opts := transformerutil.MakeOpts(cfg, a.Name, d.Path, false)
	return transformerutil.BuildMarkdownFile(
		transformerutil.FilePath(a.SubagentsDir, d.Name, ".md"),
		fm.String(), d.Body, opts,
	)
}

// BuildCommandsSection 构造 ## Slash Commands 段落，用于 InjectCommandsToRoot
//
// 抽离自原 codex.buildCommandsSection，方便 Phase 3 codex 迁移复用。
func BuildCommandsSection(commands []*parser.Document) string {
	var b strings.Builder
	b.WriteString("\n## Slash Commands\n\n")
	b.WriteString("以下命令可由用户输入 `/<name>` 触发，或在 LLM 对话中描述意图调用。\n\n")
	for _, d := range commands {
		fmt.Fprintf(&b, "### /%s\n\n", d.Name)
		if d.Description != "" {
			fmt.Fprintf(&b, "%s\n\n", d.Description)
		}
		body := strings.TrimSpace(d.Body)
		if body != "" {
			b.WriteString(body)
			b.WriteString("\n\n")
		}
	}
	return b.String()
}

// InjectBeforeFooter 把 block 插入到 content 尾部，但在 stdagent close marker 之前。
//
// 抽离自原 codex.injectBeforeFooter，cross-cutting helper。无 marker 时直接追加。
func InjectBeforeFooter(content []byte, block string) []byte {
	s := string(content)
	const marker = "<!-- /Generated by stdagent -->"
	idx := strings.LastIndex(s, marker)
	if idx < 0 {
		return []byte(s + block)
	}
	return []byte(s[:idx] + block + s[idx:])
}

// splitBySize 把 docs 按 BodyBytes 拆分：超过 maxBytes 的进 spilled，其余留 keep。
//
// maxBytes <= 0 时不拆分。BodyBytes 由 parser 在解析时填充；零值时按 len(Body) 计算。
func splitBySize(docs []*parser.Document, maxBytes int) (keep, spilled []*parser.Document) {
	if maxBytes <= 0 {
		return docs, nil
	}
	for _, d := range docs {
		size := d.BodyBytes
		if size == 0 {
			size = len(d.Body)
		}
		if size > maxBytes {
			spilled = append(spilled, d)
		} else {
			keep = append(keep, d)
		}
	}
	return keep, spilled
}
