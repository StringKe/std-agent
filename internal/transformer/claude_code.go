package transformer

import (
	"encoding/json"
	"fmt"
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

	if cfg.Hooks != nil && len(cfg.Hooks.Hooks) > 0 {
		plan.Files = append(plan.Files, c.buildHooksJSON(cfg.Hooks))
	}

	docs = FilterDocs(docs, c.Name())
	if len(docs) == 0 {
		return plan, nil
	}

	rules := FilterByType(docs, parser.TypeRules)
	skills := FilterByType(docs, parser.TypeSkills)
	commands := FilterByType(docs, parser.TypeCommands)
	SortDocs(rules)
	SortDocs(skills)
	SortDocs(commands)

	plan.Files = append(plan.Files, c.buildClaudeMd(rules, cfg))
	for _, d := range rules {
		plan.Files = append(plan.Files, c.buildRuleFile(d, cfg))
	}
	for _, d := range skills {
		plan.Files = append(plan.Files, c.buildSkillFile(d, cfg)...)
	}
	for _, d := range commands {
		plan.Files = append(plan.Files, c.buildCommandFile(d, cfg))
	}

	return plan, nil
}

func (c *ClaudeCode) buildMCPJSON(mcp *config.MCPConfig) writer.FileOp {
	wrapper := struct {
		MCPServers map[string]config.MCPServer `json:"mcpServers"`
	}{MCPServers: mcp.Servers}
	body, _ := json.MarshalIndent(wrapper, "", "  ")
	body = append(body, '\n')
	return writer.FileOp{Path: ".mcp.json", Content: body}
}

// buildHooksJSON 输出 .claude/stdagent-hooks.json
//
// Claude Code 实际只读 .claude/settings.json 的 hooks 字段。stdagent 不直接覆盖
// settings.json（避免破坏用户其他配置），改写中间文件 stdagent-hooks.json，
// 由 `stdagent apply-hooks` 命令负责 merge 到 settings.json。
func (c *ClaudeCode) buildHooksJSON(hooks *config.HooksConfig) writer.FileOp {
	wrapper := struct {
		Hooks map[string][]config.HookEntry `json:"hooks"`
	}{Hooks: hooks.Hooks}
	body, _ := json.MarshalIndent(wrapper, "", "  ")
	body = append(body, '\n')
	return writer.FileOp{
		Path:    ".claude/stdagent-hooks.json",
		Content: body,
		Reason:  "INFO: 跑 `stdagent apply-hooks --target claude-code` 把此文件 merge 进 .claude/settings.json",
	}
}

func (c *ClaudeCode) buildClaudeMd(rules []*parser.Document, cfg *config.Config) writer.FileOp {
	opts := MakeOpts(cfg, c.Name(), "", true)
	var body strings.Builder
	body.WriteString("# Project CLAUDE Manifest\n\n")
	if len(rules) == 0 {
		body.WriteString("No rules synced.\n")
	} else {
		for _, d := range rules {
			fmt.Fprintf(&body, "@%s\n", FilePath(".claude/rules", d.Name, ".md"))
		}
	}
	return BuildMarkdownFile("CLAUDE.md", "", body.String(), opts)
}

func (c *ClaudeCode) buildRuleFile(d *parser.Document, cfg *config.Config) writer.FileOp {
	var fm FmBuilder
	fm.AddList("applyTo", d.ApplyTo)
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
	fm.AddList("paths", d.ApplyTo)
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
