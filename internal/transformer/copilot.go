package transformer

import (
	"encoding/json"
	"fmt"
	"strings"

	"std-ai/internal/config"
	"std-ai/internal/parser"
	"std-ai/internal/transformer/transformerutil"
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
	rules := transformerutil.FilterByType(docs, parser.TypeRules)
	skills := transformerutil.FilterByType(docs, parser.TypeSkills)
	commands := transformerutil.FilterByType(docs, parser.TypeCommands)
	transformerutil.SortDocs(rules)
	transformerutil.SortDocs(skills)
	transformerutil.SortDocs(commands)

	// rules 拆分：root 或 无 applyTo -> 拼到 copilot-instructions.md；其他 -> 单文件
	var general, pathSpecific []*parser.Document
	for _, d := range rules {
		if d.Root || len(transformerutil.EffectiveApplyTo(d, c.Name())) == 0 {
			general = append(general, d)
		} else {
			pathSpecific = append(pathSpecific, d)
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
	opts := transformerutil.MakeOpts(cfg, c.Name(), "", true)
	roots, nonRoot := transformerutil.PartitionRoot(general)
	var body strings.Builder
	if len(roots) > 0 {
		body.WriteString(transformerutil.RenderRootBody(roots))
		if len(nonRoot) > 0 {
			body.WriteString("\n")
			body.WriteString(transformerutil.JoinAGENTSStyle("", nonRoot))
		}
	} else {
		body.WriteString(transformerutil.JoinAGENTSStyle("Copilot Repository Instructions", nonRoot))
	}
	op := transformerutil.BuildMarkdownFile(".github/copilot-instructions.md", "", body.String(), opts)
	op.IsRoot = true
	return op
}

func (c *Copilot) buildPathInstruction(d *parser.Document, cfg *config.Config) writer.FileOp {
	var fm transformerutil.FmBuilder
	// applyTo 接受逗号分隔的多 glob 字符串
	fm.Add("applyTo", strings.Join(transformerutil.EffectiveApplyTo(d, c.Name()), ","))
	opts := transformerutil.MakeOpts(cfg, c.Name(), d.Path, false)
	body := d.Body
	if d.Description != "" {
		body = d.Description + "\n\n" + d.Body
	}
	return transformerutil.BuildMarkdownFile(
		transformerutil.FilePath(".github/instructions", d.Name, ".instructions.md"),
		fm.String(), body, opts,
	)
}

func (c *Copilot) buildAgent(d *parser.Document, cfg *config.Config) writer.FileOp {
	var fm transformerutil.FmBuilder
	fm.Add("description", transformerutil.MergeDescription(d.Description, d.WhenToUse))
	fm.AddList("tools", d.AllowedTools)
	fm.Add("model", d.Model)
	opts := transformerutil.MakeOpts(cfg, c.Name(), d.Path, false)
	op := transformerutil.BuildMarkdownFile(
		transformerutil.FilePath(".github/agents", d.Name, ".agent.md"),
		fm.String(), d.Body, opts,
	)
	if len(d.SkillFiles) > 0 {
		op.Reason = fmt.Sprintf("WARN: %d SKILL package 辅助文件被忽略，copilot .agent.md 是单文件不支持子目录（参考 docs/spec.md 4.5 降级链）", len(d.SkillFiles))
	}
	return op
}

func (c *Copilot) buildPrompt(d *parser.Document, cfg *config.Config) writer.FileOp {
	var fm transformerutil.FmBuilder
	fm.Add("description", d.Description)
	fm.Add("argument-hint", d.ArgumentHint)
	fm.AddList("tools", d.AllowedTools)
	fm.Add("model", d.Model)
	opts := transformerutil.MakeOpts(cfg, c.Name(), d.Path, false)
	return transformerutil.BuildMarkdownFile(
		transformerutil.FilePath(".github/prompts", d.Name, ".prompt.md"),
		fm.String(), d.Body, opts,
	)
}
