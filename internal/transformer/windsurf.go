package transformer

import (
	"fmt"

	"std-ai/internal/config"
	"std-ai/internal/parser"
	"std-ai/internal/transformer/transformerutil"
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
	rules := transformerutil.FilterByType(docs, parser.TypeRules)
	skills := transformerutil.FilterByType(docs, parser.TypeSkills)
	commands := transformerutil.FilterByType(docs, parser.TypeCommands)
	transformerutil.SortDocs(rules)
	transformerutil.SortDocs(skills)
	transformerutil.SortDocs(commands)

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
	var fm transformerutil.FmBuilder
	applyTo := transformerutil.EffectiveApplyTo(d, w.Name())
	switch {
	case d.AlwaysApply:
		fm.Add("trigger", "always_on")
	case len(applyTo) > 0:
		fm.Add("trigger", "glob")
		fm.AddList("globs", applyTo)
	case d.Description != "":
		fm.Add("trigger", "model_decision")
		fm.Add("description", d.Description)
	default:
		fm.Add("trigger", "manual")
	}
	opts := transformerutil.MakeOpts(cfg, w.Name(), d.Path, false)
	op := transformerutil.BuildMarkdownFile(
		transformerutil.FilePath(".windsurf/rules", d.Name, ".md"),
		fm.String(), d.Body, opts,
	)
	if len(op.Content) > windsurfRuleMaxChars {
		op.Reason = fmt.Sprintf("WARN: rule exceeds %d chars; consider splitting", windsurfRuleMaxChars)
	}
	return op
}

func (w *Windsurf) buildSkill(d *parser.Document, cfg *config.Config) []writer.FileOp {
	skillDir := transformerutil.SkillDir(".windsurf/skills", d.Name)
	var fm transformerutil.FmBuilder
	fm.Add("name", d.Name)
	// Windsurf 文档仅承认 name + description 必填；其他字段未确认背书，不输出避免风险
	fm.Add("description", transformerutil.MergeDescription(d.Description, d.WhenToUse))
	opts := transformerutil.MakeOpts(cfg, w.Name(), d.Path, false)
	skillMd := transformerutil.BuildMarkdownFile(skillDir+"/SKILL.md", fm.String(), d.Body, opts)
	return transformerutil.BuildSkillPackage(skillDir, skillMd, d.SkillFiles)
}

func (w *Windsurf) buildWorkflow(d *parser.Document, cfg *config.Config) writer.FileOp {
	opts := transformerutil.MakeOpts(cfg, w.Name(), d.Path, false)
	body := d.Body
	if d.Description != "" {
		body = d.Description + "\n\n" + d.Body
	}
	return transformerutil.BuildMarkdownFile(
		transformerutil.FilePath(".windsurf/workflows", d.Name, ".md"),
		"", body, opts,
	)
}
