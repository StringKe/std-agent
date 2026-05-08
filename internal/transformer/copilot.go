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
	Register(&Copilot{})
}

// Copilot 是 GitHub Copilot transformer
type Copilot struct{}

// Name 返回 "copilot"
func (c *Copilot) Name() string { return "copilot" }

// Plan 计算输出
func (c *Copilot) Plan(docs []*parser.Document, cfg *config.Config) (*writer.Plan, error) {
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
	SortDocs(rules)
	SortDocs(skills)
	SortDocs(commands)

	// rules 拆分：无 applyTo -> 拼到 copilot-instructions.md；有 applyTo -> 单文件
	var general, pathSpecific []*parser.Document
	for _, d := range rules {
		if len(d.ApplyTo) > 0 {
			pathSpecific = append(pathSpecific, d)
		} else {
			general = append(general, d)
		}
	}

	if len(general) > 0 {
		plan.Files = append(plan.Files, c.buildInstructions(general, cfg))
	}
	for _, d := range pathSpecific {
		plan.Files = append(plan.Files, c.buildPathInstruction(d, cfg))
	}
	for _, d := range skills {
		plan.Files = append(plan.Files, c.buildAgent(d, cfg))
	}
	for _, d := range commands {
		plan.Files = append(plan.Files, c.buildPrompt(d, cfg))
	}
	return plan, nil
}

// buildMCPJSON 写 .vscode/mcp.json（注意：VS Code 工作区 MCP 顶级键是 "servers"，
// 与 Claude Code 的 .mcp.json 顶级键 "mcpServers" 不同）
func (c *Copilot) buildMCPJSON(mcp *config.MCPConfig) writer.FileOp {
	wrapper := struct {
		Servers map[string]config.MCPServer `json:"servers"`
	}{Servers: mcp.Servers}
	body, _ := json.MarshalIndent(wrapper, "", "  ")
	body = append(body, '\n')
	return writer.FileOp{Path: ".vscode/mcp.json", Content: body}
}

func (c *Copilot) buildInstructions(general []*parser.Document, cfg *config.Config) writer.FileOp {
	opts := MakeOpts(cfg, c.Name(), "", true)
	body := JoinAGENTSStyle("Copilot Repository Instructions", general)
	return BuildMarkdownFile(".github/copilot-instructions.md", "", body, opts)
}

func (c *Copilot) buildPathInstruction(d *parser.Document, cfg *config.Config) writer.FileOp {
	var fm FmBuilder
	// applyTo 接受逗号分隔的多 glob 字符串
	fm.Add("applyTo", strings.Join(d.ApplyTo, ","))
	opts := MakeOpts(cfg, c.Name(), d.Path, false)
	body := d.Body
	if d.Description != "" {
		body = d.Description + "\n\n" + d.Body
	}
	return BuildMarkdownFile(
		FilePath(".github/instructions", d.Name, ".instructions.md"),
		fm.String(), body, opts,
	)
}

func (c *Copilot) buildAgent(d *parser.Document, cfg *config.Config) writer.FileOp {
	var fm FmBuilder
	fm.Add("description", MergeDescription(d.Description, d.WhenToUse))
	fm.AddList("tools", d.AllowedTools)
	fm.Add("model", d.Model)
	opts := MakeOpts(cfg, c.Name(), d.Path, false)
	op := BuildMarkdownFile(
		FilePath(".github/agents", d.Name, ".agent.md"),
		fm.String(), d.Body, opts,
	)
	if len(d.SkillFiles) > 0 {
		op.Reason = fmt.Sprintf("WARN: %d SKILL package 辅助文件被忽略，copilot .agent.md 是单文件不支持子目录（参考 docs/spec.md 4.5 降级链）", len(d.SkillFiles))
	}
	return op
}

func (c *Copilot) buildPrompt(d *parser.Document, cfg *config.Config) writer.FileOp {
	var fm FmBuilder
	fm.Add("description", d.Description)
	fm.Add("argument-hint", d.ArgumentHint)
	fm.AddList("tools", d.AllowedTools)
	fm.Add("model", d.Model)
	opts := MakeOpts(cfg, c.Name(), d.Path, false)
	return BuildMarkdownFile(
		FilePath(".github/prompts", d.Name, ".prompt.md"),
		fm.String(), d.Body, opts,
	)
}
