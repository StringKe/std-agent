package protocol

import (
	"encoding/json"
	"path"
	"strings"

	"std-ai/internal/config"
	"std-ai/internal/parser"
	"std-ai/internal/transformer/transformerutil"
	"std-ai/internal/writer"
)

// ClaudeMD 是 Anthropic Claude Code 的协议族实现。
//
// 协议形态（与 ClaudeMDAdapter 配合）：
//   - 根文件：CLAUDE.md（root rule body + manifest 段 + type glossary 头部）
//   - 嵌套根文件：<NestedPath>/CLAUDE.md（纯 body，无 manifest 无 glossary）
//   - rules：.claude/rules/<name>.md，frontmatter 用 Anthropic 私有方言
//     `paths`（由 adapter.GlobsFieldName="paths" 触发，等价于 Copilot 的
//     applyTo / Cursor 的 globs）
//   - skills：.claude/skills/<name>/SKILL.md，含 Agent Skills 规范字段 +
//     Anthropic 私有字段集（when_to_use / argument-hint / tools / effort /
//     paths / hooks / agent / shell / context / arguments / model /
//     disable-model-invocation）。这些私有字段不下放到通用 Adapter，由本
//     协议内部 helper renderClaudeSkillFrontmatter 处理。
//   - commands：.claude/commands/<name>.md
//   - subagents：.claude/agents/<name>.md（Claude Code 原生支持，无需 fallback）
//   - references：原生无 references 类型，统一 fallback 到
//     .claude/skills/<name>/SKILL.md 形式（Agent Skills 标准），frontmatter
//     注入 std-ai-type: references 私有标识便于 AI 识别
//   - MCP：.mcp.json，顶级键 mcpServers（与 Copilot 的 servers 不同）
type ClaudeMD struct{}

// Plan 实现 Protocol 接口
func (ClaudeMD) Plan(docs []*parser.Document, adapter Adapter, cfg *config.Config) (*writer.Plan, error) {
	plan := &writer.Plan{Target: adapter.Name}
	if adapter.Disabled {
		return plan, nil
	}

	// MCP（独立于 docs，先产出确保 cfg.MCP 即使无 docs 也能写入）
	if cfg != nil && cfg.MCP != nil && len(cfg.MCP.Servers) > 0 && adapter.MCPPath != "" {
		plan.Files = append(plan.Files, buildClaudeMCPJSON(adapter, cfg.MCP))
	}

	if len(docs) == 0 {
		return plan, nil
	}

	rules := transformerutil.FilterByType(docs, parser.TypeRules)
	skills := transformerutil.FilterByType(docs, parser.TypeSkills)
	commands := transformerutil.FilterByType(docs, parser.TypeCommands)
	subagents := transformerutil.FilterByType(docs, parser.TypeSubagents)
	references := transformerutil.FilterByType(docs, parser.TypeReferences)
	transformerutil.SortDocs(rules)
	transformerutil.SortDocs(skills)
	transformerutil.SortDocs(commands)
	transformerutil.SortDocs(subagents)
	transformerutil.SortDocs(references)

	// 嵌套 root 单独输出（不进 manifest，不带 glossary）
	topRules, nestedRoots := transformerutil.PartitionNested(rules)
	for _, d := range nestedRoots {
		plan.Files = append(plan.Files, buildClaudeNestedRoot(d, adapter, cfg))
	}

	roots, nonRoot := transformerutil.PartitionRoot(topRules)
	plan.Files = append(plan.Files, buildClaudeRoot(roots, nonRoot, adapter, cfg))
	for _, d := range nonRoot {
		plan.Files = append(plan.Files, buildClaudeRuleFile(d, adapter, cfg))
	}
	for _, d := range skills {
		plan.Files = append(plan.Files, buildClaudeSkillFile(d, adapter, cfg)...)
	}
	for _, d := range commands {
		plan.Files = append(plan.Files, buildClaudeCommandFile(d, adapter, cfg))
	}
	for _, d := range subagents {
		plan.Files = append(plan.Files, buildClaudeSubagentFile(d, adapter, cfg))
	}
	// references 走独立子目录 .claude/references/<name>.md（v3：不借 skills 路径，
	// 否则 AI 会按 skill 触发逻辑加载 references 破坏"按需查阅"语义）
	for _, d := range references {
		plan.Files = append(plan.Files, BuildDegradedFileOp(d, adapter, cfg))
	}

	return plan, nil
}

// buildClaudeRoot 构造 CLAUDE.md 顶级根文件。
//
// 结构（自上而下）：
//  1. stdagent header marker（由 BuildMarkdownFile 注入）
//  2. type glossary 段（adapter.InjectTypeGlossary=true 时）
//  3. root rule body 或占位标题
//  4. Imported Rules manifest 段（含 nonRoot rule 索引）
//  5. stdagent footer marker
func buildClaudeRoot(roots, nonRoot []*parser.Document, adapter Adapter, cfg *config.Config) writer.FileOp {
	opts := transformerutil.MakeOpts(cfg, adapter.Name, "", true)
	var body strings.Builder
	if g := RenderGlossaryFor(adapter); g != "" {
		body.WriteString(g)
		if !strings.HasSuffix(g, "\n") {
			body.WriteString("\n")
		}
		body.WriteString("\n")
	}
	if len(roots) > 0 {
		body.WriteString(transformerutil.RenderRootBody(roots))
	} else {
		body.WriteString("# Project CLAUDE Manifest\n\n")
	}
	if len(nonRoot) > 0 {
		body.WriteString(transformerutil.BuildRuleManifestSection(
			manifestTitle(adapter),
			adapter.Name,
			nonRoot,
			func(d *parser.Document) string {
				return transformerutil.FilePath(claudeRulesDir(adapter), d.Name, ".md")
			},
			true,
		))
	} else if len(roots) == 0 {
		body.WriteString("No rules synced.\n")
	}
	rootName := adapter.RootFileName
	if rootName == "" {
		rootName = "CLAUDE.md"
	}
	op := transformerutil.BuildMarkdownFile(rootName, "", body.String(), opts)
	op.IsRoot = true
	return op
}

// buildClaudeNestedRoot 把嵌套 root doc 输出到 <NestedPath>/CLAUDE.md。
// 不含 manifest 段也不含 type glossary。
func buildClaudeNestedRoot(d *parser.Document, adapter Adapter, cfg *config.Config) writer.FileOp {
	opts := transformerutil.MakeOpts(cfg, adapter.Name, d.Path, false)
	nestedName := adapter.NestedFileName
	if nestedName == "" {
		nestedName = adapter.RootFileName
	}
	if nestedName == "" {
		nestedName = "CLAUDE.md"
	}
	op := transformerutil.BuildMarkdownFile(
		path.Join(d.NestedPath, nestedName),
		"", d.Body, opts,
	)
	op.IsRoot = true
	return op
}

// buildClaudeRuleFile 输出 .claude/rules/<name>.md
//
// frontmatter 行为：
//   - paths（Anthropic 私有方言，由 adapter.GlobsFieldName="paths" 触发）
//   - description（人类可读 metadata）
//
// Claude Code 官方仅认 paths 字段，不支持 alwaysApply / applyTo / globs。
func buildClaudeRuleFile(d *parser.Document, adapter Adapter, cfg *config.Config) writer.FileOp {
	var fm transformerutil.FmBuilder
	field := adapter.GlobsFieldName
	if field == "" {
		field = "paths"
	}
	fm.AddList(field, transformerutil.EffectiveApplyTo(d, adapter.Name))
	if adapter.SupportsDescription || d.Description != "" {
		fm.Add("description", d.Description)
	}
	opts := transformerutil.MakeOpts(cfg, adapter.Name, d.Path, false)
	return transformerutil.BuildMarkdownFile(
		transformerutil.FilePath(claudeRulesDir(adapter), d.Name, ".md"),
		fm.String(), d.Body, opts,
	)
}

// buildClaudeSkillFile 输出 .claude/skills/<name>/SKILL.md + SkillFiles。
//
// frontmatter 渲染走 renderClaudeSkillFrontmatter（含 Anthropic 私有字段集）。
func buildClaudeSkillFile(d *parser.Document, adapter Adapter, cfg *config.Config) []writer.FileOp {
	skillDir := transformerutil.SkillDir(claudeSkillsDir(adapter), d.Name)
	fm := renderClaudeSkillFrontmatter(d, adapter)
	opts := transformerutil.MakeOpts(cfg, adapter.Name, d.Path, false)
	skillMd := transformerutil.BuildMarkdownFile(skillDir+"/SKILL.md", fm, d.Body, opts)
	return transformerutil.BuildSkillPackage(skillDir, skillMd, d.SkillFiles)
}

// renderClaudeSkillFrontmatter 构造 Claude Code skill 的完整 frontmatter，
// 含 Anthropic 私有字段集（when_to_use / argument-hint / tools / effort /
// paths / hooks / agent / shell / context / arguments / model /
// disable-model-invocation）+ Agent Skills 规范字段（name / description /
// license / compatibility / metadata）。
//
// 这些私有字段不下放到通用 Adapter.SkillSupportedFields，统一在本协议内部
// 处理，让其他 protocol 不必感知 Anthropic 方言。
func renderClaudeSkillFrontmatter(d *parser.Document, adapter Adapter) string {
	var fm transformerutil.FmBuilder
	fm.Add("name", d.Name)
	fm.Add("description", d.Description)
	// Claude Code 私有字段
	fm.Add("when_to_use", d.WhenToUse)
	fm.Add("model", d.Model)
	fm.Add("argument-hint", d.ArgumentHint)
	fm.AddList("arguments", d.Arguments)
	fm.Add("effort", d.Effort)
	fm.Add("context", d.SkillContext)
	fm.Add("agent", d.Agent)
	fm.Add("shell", d.Shell)
	fm.AddList("tools", d.AllowedTools)
	fm.AddList("paths", transformerutil.EffectiveApplyTo(d, adapter.Name))
	if d.DisableModelInvocation {
		fm.AddBool("disable-model-invocation", true)
	}
	// Agent Skills 标准字段
	fm.Add("license", d.License)
	fm.Add("compatibility", d.Compatibility)
	fm.AddMap("metadata", d.Metadata)
	// Claude Code 专属 hooks（嵌套 map）
	fm.AddMap("hooks", d.Hooks)
	return fm.String()
}

// buildClaudeCommandFile 输出 .claude/commands/<name>.md
//
// frontmatter：description / argument-hint / allowed-tools / model。
func buildClaudeCommandFile(d *parser.Document, adapter Adapter, cfg *config.Config) writer.FileOp {
	var fm transformerutil.FmBuilder
	fm.Add("description", d.Description)
	fm.Add("argument-hint", d.ArgumentHint)
	fm.AddList("allowed-tools", d.AllowedTools)
	fm.Add("model", d.Model)
	opts := transformerutil.MakeOpts(cfg, adapter.Name, d.Path, false)
	return transformerutil.BuildMarkdownFile(
		transformerutil.FilePath(claudeCommandsDir(adapter), d.Name, ".md"),
		fm.String(), d.Body, opts,
	)
}

// buildClaudeSubagentFile 输出 .claude/agents/<name>.md
//
// Claude Code 原生支持 subagent 定义，frontmatter：name / description /
// model / tools，body 是 subagent 的系统提示词。
func buildClaudeSubagentFile(d *parser.Document, adapter Adapter, cfg *config.Config) writer.FileOp {
	var fm transformerutil.FmBuilder
	fm.Add("name", d.Name)
	fm.Add("description", d.Description)
	fm.Add("model", d.Model)
	fm.AddList("tools", d.AllowedTools)
	opts := transformerutil.MakeOpts(cfg, adapter.Name, d.Path, false)
	return transformerutil.BuildMarkdownFile(
		transformerutil.FilePath(claudeSubagentsDir(adapter), d.Name, ".md"),
		fm.String(), d.Body, opts,
	)
}

// buildClaudeMCPJSON 输出 .mcp.json，顶级键 mcpServers（≠ Copilot 的 servers）
func buildClaudeMCPJSON(adapter Adapter, mcp *config.MCPConfig) writer.FileOp {
	topKey := adapter.MCPTopKey
	if topKey == "" {
		topKey = "mcpServers"
	}
	wrapper := map[string]map[string]config.MCPServer{topKey: mcp.Servers}
	body, _ := json.MarshalIndent(wrapper, "", "  ")
	body = append(body, '\n')
	return writer.FileOp{Path: adapter.MCPPath, Content: body}
}

// claudeRulesDir 返回 rules 输出目录（默认 .claude/rules）
func claudeRulesDir(adapter Adapter) string {
	if adapter.RulesDir != "" {
		return adapter.RulesDir
	}
	return ".claude/rules"
}

// claudeSkillsDir 返回 skills 输出目录（默认 .claude/skills）
func claudeSkillsDir(adapter Adapter) string {
	if adapter.SkillsDir != "" {
		return adapter.SkillsDir
	}
	return ".claude/skills"
}

// claudeCommandsDir 返回 commands 输出目录（默认 .claude/commands）
func claudeCommandsDir(adapter Adapter) string {
	if adapter.CommandsDir != "" {
		return adapter.CommandsDir
	}
	return ".claude/commands"
}

// claudeSubagentsDir 返回 subagents 输出目录（默认 .claude/agents）
func claudeSubagentsDir(adapter Adapter) string {
	if adapter.SubagentsDir != "" {
		return adapter.SubagentsDir
	}
	return ".claude/agents"
}

// manifestTitle 返回 root 文件 manifest 段标题（默认 "Imported Rules"）
func manifestTitle(adapter Adapter) string {
	if adapter.ManifestSection != "" {
		return adapter.ManifestSection
	}
	return "Imported Rules"
}
