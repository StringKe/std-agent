package transformer

import (
	"fmt"
	"path"

	"std-ai/internal/config"
	"std-ai/internal/parser"
	"std-ai/internal/writer"
)

func init() {
	Register(&Cline{})
}

// Cline 是 VS Code Cline 扩展 transformer
type Cline struct{}

// Name 返回 "cline"
func (c *Cline) Name() string { return "cline" }

// Plan 计算输出
func (c *Cline) Plan(docs []*parser.Document, cfg *config.Config) (*writer.Plan, error) {
	plan := &writer.Plan{Target: c.Name()}
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

	for _, d := range rules {
		plan.Files = append(plan.Files, c.buildRule(d, cfg))
	}
	for _, d := range skills {
		plan.Files = append(plan.Files, c.buildWorkflow(d, cfg))
	}
	for _, d := range commands {
		plan.Files = append(plan.Files, c.buildWorkflow(d, cfg))
	}
	return plan, nil
}

// buildRule 输出 .clinerules/<NNN>-<name>.md，NNN 由 priority 决定
func (c *Cline) buildRule(d *parser.Document, cfg *config.Config) writer.FileOp {
	prefix := 500
	switch d.Priority {
	case parser.PriorityHigh:
		prefix = 100
	case parser.PriorityLow:
		prefix = 900
	}
	var fm FmBuilder
	fm.AddList("paths", EffectiveApplyTo(d, c.Name()))
	opts := MakeOpts(cfg, c.Name(), d.Path, false)
	return BuildMarkdownFile(
		path.Join(".clinerules", fmt.Sprintf("%03d-%s.md", prefix, d.Name)),
		fm.String(), d.Body, opts,
	)
}

func (c *Cline) buildWorkflow(d *parser.Document, cfg *config.Config) writer.FileOp {
	opts := MakeOpts(cfg, c.Name(), d.Path, false)
	body := d.Body
	if d.Description != "" {
		body = d.Description + "\n\n" + d.Body
	}
	return BuildMarkdownFile(
		FilePath(".clinerules/workflows", d.Name, ".md"),
		"", body, opts,
	)
}
