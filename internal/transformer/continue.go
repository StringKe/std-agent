package transformer

import (
	"std-ai/internal/config"
	"std-ai/internal/parser"
	"std-ai/internal/writer"
)

func init() {
	Register(&ContinueDev{})
}

// ContinueDev 是 Continue.dev VS Code/JetBrains 扩展 transformer
type ContinueDev struct{}

// Name 返回 "continue-dev"（避免与 Go keyword `continue` 视觉冲突）
func (c *ContinueDev) Name() string { return "continue-dev" }

// Plan 计算输出
func (c *ContinueDev) Plan(docs []*parser.Document, cfg *config.Config) (*writer.Plan, error) {
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
		plan.Files = append(plan.Files, c.buildSkillAsRule(d, cfg))
	}
	for _, d := range commands {
		plan.Files = append(plan.Files, c.buildPrompt(d, cfg))
	}
	return plan, nil
}

// buildRule -> .continue/rules/<n>.md，frontmatter name/description/globs/alwaysApply
func (c *ContinueDev) buildRule(d *parser.Document, cfg *config.Config) writer.FileOp {
	var fm FmBuilder
	fm.Add("name", d.Name)
	fm.Add("description", d.Description)
	fm.AddList("globs", d.ApplyTo)
	if d.AlwaysApply {
		fm.AddBool("alwaysApply", true)
	}
	opts := MakeOpts(cfg, c.Name(), d.Path, false)
	return BuildMarkdownFile(
		FilePath(".continue/rules", d.Name, ".md"),
		fm.String(), d.Body, opts,
	)
}

// buildSkillAsRule 把 std skill 降级为 Continue model-decision rule
// 输出到 .continue/rules/skill-<n>.md 避免与同 name 的 rule 冲突
func (c *ContinueDev) buildSkillAsRule(d *parser.Document, cfg *config.Config) writer.FileOp {
	var fm FmBuilder
	fm.Add("name", "Skill: "+d.Name)
	desc := d.Description
	if desc == "" {
		desc = "Skill: " + d.Name
	}
	fm.Add("description", desc)
	fm.AddBool("alwaysApply", false)
	opts := MakeOpts(cfg, c.Name(), d.Path, false)
	return BuildMarkdownFile(
		FilePath(".continue/rules", "skill-"+d.Name, ".md"),
		fm.String(), d.Body, opts,
	)
}

// buildPrompt -> .continue/prompts/<n>.prompt.md，frontmatter
// name/description/version/invokable=true（slash 触发）
func (c *ContinueDev) buildPrompt(d *parser.Document, cfg *config.Config) writer.FileOp {
	var fm FmBuilder
	fm.Add("name", d.Name)
	fm.Add("description", d.Description)
	fm.Add("version", d.Version)
	fm.AddBool("invokable", true)
	opts := MakeOpts(cfg, c.Name(), d.Path, false)
	return BuildMarkdownFile(
		FilePath(".continue/prompts", d.Name, ".prompt.md"),
		fm.String(), d.Body, opts,
	)
}
