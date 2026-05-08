package transformer

import (
	"encoding/json"
	"path"
	"strings"

	"std-ai/internal/config"
	"std-ai/internal/parser"
	"std-ai/internal/writer"
)

func init() {
	Register(&ClaudeCode{})
}

// ClaudeCode 是 Anthropic Claude Code transformer
type ClaudeCode struct{}

// Name 返回 "claude-code"
func (c *ClaudeCode) Name() string { return "claude-code" }

// Plan 计算输出文件清单
func (c *ClaudeCode) Plan(docs []*parser.Document, cfg *config.Config) (*writer.Plan, error) {
	plan := &writer.Plan{Target: c.Name()}

	if cfg.MCP != nil && len(cfg.MCP.Servers) > 0 {
		plan.Files = append(plan.Files, c.buildMCPJSON(cfg.MCP))
	}

	docs = FilterDocs(docs, c.Name())
	if len(docs) == 0 {
		return plan, nil
	}

	rules := FilterByType(docs, parser.TypeRules)
	skills := FilterByType(docs, parser.TypeSkills)
	commands := FilterByType(docs, parser.TypeCommands)
	subagents := FilterByType(docs, parser.TypeSubagents)
	SortDocs(rules)
	SortDocs(skills)
	SortDocs(commands)
	SortDocs(subagents)

	// 嵌套 root 单独输出到 <NestedPath>/CLAUDE.md（无 manifest）
	topRules, nestedRoots := PartitionNested(rules)
	for _, d := range nestedRoots {
		plan.Files = append(plan.Files, c.buildNestedClaudeMd(d, cfg))
	}

	roots, nonRoot := PartitionRoot(topRules)
	plan.Files = append(plan.Files, c.buildClaudeMd(roots, nonRoot, cfg))
	// root rule 不 fan-out 成 .claude/rules/<root>.md（已是 CLAUDE.md 主体）
	for _, d := range nonRoot {
		plan.Files = append(plan.Files, c.buildRuleFile(d, cfg))
	}
	for _, d := range skills {
		plan.Files = append(plan.Files, c.buildSkillFile(d, cfg)...)
	}
	for _, d := range commands {
		plan.Files = append(plan.Files, c.buildCommandFile(d, cfg))
	}
	for _, d := range subagents {
		plan.Files = append(plan.Files, c.buildSubagentFile(d, cfg))
	}

	return plan, nil
}

// buildSubagentFile 输出 .claude/agents/<name>.md（Claude Code 原生 subagent 定义）。
// frontmatter: name / description / model / tools；body 即 subagent 的系统提示词。
func (c *ClaudeCode) buildSubagentFile(d *parser.Document, cfg *config.Config) writer.FileOp {
	var fm FmBuilder
	fm.Add("name", d.Name)
	fm.Add("description", d.Description)
	fm.Add("model", d.Model)
	fm.AddList("tools", d.AllowedTools)
	opts := MakeOpts(cfg, c.Name(), d.Path, false)
	return BuildMarkdownFile(
		FilePath(".claude/agents", d.Name, ".md"),
		fm.String(), d.Body, opts,
	)
}

func (c *ClaudeCode) buildMCPJSON(mcp *config.MCPConfig) writer.FileOp {
	wrapper := struct {
		MCPServers map[string]config.MCPServer `json:"mcpServers"`
	}{MCPServers: mcp.Servers}
	body, _ := json.MarshalIndent(wrapper, "", "  ")
	body = append(body, '\n')
	return writer.FileOp{Path: ".mcp.json", Content: body}
}

// buildClaudeMd 拼 CLAUDE.md = root body（项目总结）+ stdagent 自动 manifest 清单。
//
// 设计契约：
//   - root rule body 写项目总结（项目说明、技术栈、关键铁律、子模块入口）
//   - stdagent 始终在 root body 尾部追加 ## Imported Rules 段（含 nonRoot rule 索引）
//   - 用户**不应**在 root body 里手写 rule 清单（stdagent 自动管，避免重复）
//   - 无 root rule 时，body = "# Project CLAUDE Manifest" 占位标题 + manifest
func (c *ClaudeCode) buildClaudeMd(roots, nonRoot []*parser.Document, cfg *config.Config) writer.FileOp {
	opts := MakeOpts(cfg, c.Name(), "", true)
	var body strings.Builder
	if len(roots) > 0 {
		body.WriteString(RenderRootBody(roots))
	} else {
		body.WriteString("# Project CLAUDE Manifest\n\n")
	}
	if len(nonRoot) > 0 {
		body.WriteString(BuildRuleManifestSection(
			"Imported Rules",
			c.Name(),
			nonRoot,
			func(d *parser.Document) string { return FilePath(".claude/rules", d.Name, ".md") },
			true,
		))
	} else if len(roots) == 0 {
		body.WriteString("No rules synced.\n")
	}
	op := BuildMarkdownFile("CLAUDE.md", "", body.String(), opts)
	op.IsRoot = true
	return op
}

// buildNestedClaudeMd 把嵌套 root doc 输出到 <NestedPath>/CLAUDE.md，纯 body + marker，无 manifest。
// AI 在该子目录工作时，Claude Code 自动叠加加载顶级 + 嵌套 CLAUDE.md。
func (c *ClaudeCode) buildNestedClaudeMd(d *parser.Document, cfg *config.Config) writer.FileOp {
	opts := MakeOpts(cfg, c.Name(), d.Path, false)
	op := BuildMarkdownFile(
		path.Join(d.NestedPath, "CLAUDE.md"),
		"", d.Body, opts,
	)
	op.IsRoot = true
	return op
}

func (c *ClaudeCode) buildRuleFile(d *parser.Document, cfg *config.Config) writer.FileOp {
	var fm FmBuilder
	fm.AddList("applyTo", EffectiveApplyTo(d, c.Name()))
	if d.AlwaysApply {
		fm.AddBool("alwaysApply", true)
	}
	fm.Add("description", d.Description)
	opts := MakeOpts(cfg, c.Name(), d.Path, false)
	return BuildMarkdownFile(
		FilePath(".claude/rules", d.Name, ".md"),
		fm.String(), d.Body, opts,
	)
}

func (c *ClaudeCode) buildSkillFile(d *parser.Document, cfg *config.Config) []writer.FileOp {
	skillDir := SkillDir(".claude/skills", d.Name)
	var fm FmBuilder
	fm.Add("name", d.Name)
	fm.Add("description", d.Description)
	// Claude Code 原生支持 when_to_use；其他字段按需透传
	fm.Add("when_to_use", d.WhenToUse)
	fm.Add("model", d.Model)
	fm.Add("argument-hint", d.ArgumentHint)
	fm.AddList("arguments", d.Arguments)
	fm.Add("effort", d.Effort)
	fm.Add("context", d.SkillContext)
	fm.Add("agent", d.Agent)
	fm.Add("shell", d.Shell)
	fm.AddList("tools", d.AllowedTools)
	fm.AddList("paths", EffectiveApplyTo(d, c.Name()))
	if d.DisableModelInvocation {
		fm.AddBool("disable-model-invocation", true)
	}
	// agentskills 标准字段
	fm.Add("license", d.License)
	fm.Add("compatibility", d.Compatibility)
	fm.AddMap("metadata", d.Metadata)
	// Claude Code 专属 hooks（嵌套 map）
	fm.AddMap("hooks", d.Hooks)

	opts := MakeOpts(cfg, c.Name(), d.Path, false)
	skillMd := BuildMarkdownFile(skillDir+"/SKILL.md", fm.String(), d.Body, opts)
	return BuildSkillPackage(skillDir, skillMd, d.SkillFiles)
}

func (c *ClaudeCode) buildCommandFile(d *parser.Document, cfg *config.Config) writer.FileOp {
	var fm FmBuilder
	fm.Add("description", d.Description)
	fm.Add("argument-hint", d.ArgumentHint)
	fm.AddList("allowed-tools", d.AllowedTools)
	fm.Add("model", d.Model)
	opts := MakeOpts(cfg, c.Name(), d.Path, false)
	return BuildMarkdownFile(
		FilePath(".claude/commands", d.Name, ".md"),
		fm.String(), d.Body, opts,
	)
}
