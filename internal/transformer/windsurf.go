package transformer

import (
	"fmt"

	"std-ai/internal/config"
	"std-ai/internal/parser"
	"std-ai/internal/writer"
)

func init() {
	Register(&Windsurf{})
}

// Windsurf 是 Codeium Windsurf transformer
type Windsurf struct{}

// Name 返回 "windsurf"
func (w *Windsurf) Name() string { return "windsurf" }

// windsurfRuleMaxChars 是 Windsurf 单 rule 文件字符上限
const windsurfRuleMaxChars = 12000

// Plan 计算输出
func (w *Windsurf) Plan(docs []*parser.Document, cfg *config.Config) (*writer.Plan, error) {
	plan := &writer.Plan{Target: w.Name()}
	docs = FilterDocs(docs, w.Name())
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
		plan.Files = append(plan.Files, w.buildRule(d, cfg))
	}
	for _, d := range skills {
		plan.Files = append(plan.Files, w.buildSkill(d, cfg)...)
	}
	for _, d := range commands {
		plan.Files = append(plan.Files, w.buildWorkflow(d, cfg))
	}
	return plan, nil
}

func (w *Windsurf) buildRule(d *parser.Document, cfg *config.Config) writer.FileOp {
	var fm FmBuilder
	switch {
	case d.AlwaysApply:
		fm.Add("trigger", "always_on")
	case len(d.ApplyTo) > 0:
		fm.Add("trigger", "glob")
		fm.AddList("globs", d.ApplyTo)
	case d.Description != "":
		fm.Add("trigger", "model_decision")
		fm.Add("description", d.Description)
	default:
		fm.Add("trigger", "manual")
	}
	opts := MakeOpts(cfg, w.Name(), d.Path, false)
	op := BuildMarkdownFile(
		FilePath(".windsurf/rules", d.Name, ".md"),
		fm.String(), d.Body, opts,
	)
	if len(op.Content) > windsurfRuleMaxChars {
		op.Reason = fmt.Sprintf("WARN: rule exceeds %d chars; consider splitting", windsurfRuleMaxChars)
	}
	return op
}

func (w *Windsurf) buildSkill(d *parser.Document, cfg *config.Config) []writer.FileOp {
	skillDir := SkillDir(".windsurf/skills", d.Name)
	var fm FmBuilder
	fm.Add("name", d.Name)
	// Windsurf 文档仅承认 name + description 必填；其他字段未确认背书，不输出避免风险
	fm.Add("description", MergeDescription(d.Description, d.WhenToUse))
	opts := MakeOpts(cfg, w.Name(), d.Path, false)
	skillMd := BuildMarkdownFile(skillDir+"/SKILL.md", fm.String(), d.Body, opts)
	return BuildSkillPackage(skillDir, skillMd, d.SkillFiles)
}

func (w *Windsurf) buildWorkflow(d *parser.Document, cfg *config.Config) writer.FileOp {
	opts := MakeOpts(cfg, w.Name(), d.Path, false)
	body := d.Body
	if d.Description != "" {
		body = d.Description + "\n\n" + d.Body
	}
	return BuildMarkdownFile(
		FilePath(".windsurf/workflows", d.Name, ".md"),
		"", body, opts,
	)
}
