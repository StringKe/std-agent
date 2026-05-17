package transformer

import (
	"encoding/json"
	"strings"

	"std-ai/internal/config"
	"std-ai/internal/parser"
	"std-ai/internal/transformer/transformerutil"
	"std-ai/internal/writer"
)

func init() {
	Register(&Cursor{})
}

// Cursor 是 Cursor IDE transformer
type Cursor struct{}

// Name 返回 "cursor"
func (c *Cursor) Name() string { return "cursor" }

// Plan 计算输出
func (c *Cursor) Plan(docs []*parser.Document, cfg *config.Config) (*writer.Plan, error) {
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

	for _, d := range rules {
		plan.Files = append(plan.Files, c.buildRule(d, cfg))
	}
	for _, d := range skills {
		plan.Files = append(plan.Files, c.buildSkill(d, cfg)...)
	}
	for _, d := range commands {
		plan.Files = append(plan.Files, c.buildCommand(d, cfg))
	}
	return plan, nil
}

func (c *Cursor) buildMCPJSON(mcp *config.MCPConfig) writer.FileOp {
	wrapper := struct {
		MCPServers map[string]config.MCPServer `json:"mcpServers"`
	}{MCPServers: mcp.Servers}
	body, _ := json.MarshalIndent(wrapper, "", "  ")
	body = append(body, '\n')
	return writer.FileOp{Path: ".cursor/mcp.json", Content: body}
}

func (c *Cursor) buildRule(d *parser.Document, cfg *config.Config) writer.FileOp {
	var fm transformerutil.FmBuilder
	if d.AlwaysApply {
		fm.AddBool("alwaysApply", true)
	}
	applyTo := transformerutil.EffectiveApplyTo(d, c.Name())
	if len(applyTo) > 0 {
		// Cursor globs 接受逗号分隔字符串
		fm.Add("globs", strings.Join(applyTo, ","))
	}
	fm.Add("description", d.Description)
	opts := transformerutil.MakeOpts(cfg, c.Name(), d.Path, false)
	return transformerutil.BuildMarkdownFile(
		transformerutil.FilePath(".cursor/rules", d.Name, ".mdc"),
		fm.String(), d.Body, opts,
	)
}

func (c *Cursor) buildSkill(d *parser.Document, cfg *config.Config) []writer.FileOp {
	skillDir := transformerutil.SkillDir(".cursor/skills", d.Name)
	var fm transformerutil.FmBuilder
	fm.Add("name", d.Name)
	fm.Add("description", transformerutil.MergeDescription(d.Description, d.WhenToUse))
	fm.AddList("paths", transformerutil.EffectiveApplyTo(d, c.Name()))
	if d.DisableModelInvocation {
		fm.AddBool("disable-model-invocation", true)
	}
	fm.Add("license", d.License)
	fm.Add("compatibility", d.Compatibility)
	fm.AddMap("metadata", d.Metadata)
	opts := transformerutil.MakeOpts(cfg, c.Name(), d.Path, false)
	skillMd := transformerutil.BuildMarkdownFile(skillDir+"/SKILL.md", fm.String(), d.Body, opts)
	return transformerutil.BuildSkillPackage(skillDir, skillMd, d.SkillFiles)
}

func (c *Cursor) buildCommand(d *parser.Document, cfg *config.Config) writer.FileOp {
	opts := transformerutil.MakeOpts(cfg, c.Name(), d.Path, false)
	body := d.Body
	if d.Description != "" {
		body = d.Description + "\n\n" + d.Body
	}
	return transformerutil.BuildMarkdownFile(
		transformerutil.FilePath(".cursor/commands", d.Name, ".md"),
		"", body, opts,
	)
}
