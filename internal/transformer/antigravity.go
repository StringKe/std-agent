package transformer

import (
	"fmt"

	"std-ai/internal/config"
	"std-ai/internal/parser"
	"std-ai/internal/writer"
)

func init() {
	Register(&Antigravity{})
}

// Antigravity 是 Google Antigravity AI IDE transformer
type Antigravity struct{}

// Name 返回 "antigravity"
func (a *Antigravity) Name() string { return "antigravity" }

// antigravityMaxChars 是 Antigravity 单 rule / workflow 文件字符上限
const antigravityMaxChars = 12000

// Plan 计算输出
//
// Antigravity 自 v1.20.3 起原生消费 AGENTS.md，因此 std rules 已由 codex
// transformer 写入根 AGENTS.md 时被自动消费。本 transformer 主要负责
// 细粒度 rules（按 trigger 分类）+ workflows + skills 降级。
func (a *Antigravity) Plan(docs []*parser.Document, cfg *config.Config) (*writer.Plan, error) {
	plan := &writer.Plan{Target: a.Name()}
	docs = FilterDocs(docs, a.Name())
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
		plan.Files = append(plan.Files, a.buildRule(d, cfg))
	}
	for _, d := range skills {
		plan.Files = append(plan.Files, a.buildSkillAsRule(d, cfg))
	}
	for _, d := range commands {
		plan.Files = append(plan.Files, a.buildWorkflow(d, cfg))
	}
	return plan, nil
}

// buildRule -> .agents/rules/<n>.md，trigger 由 frontmatter 推导
//
//   - alwaysApply -> trigger: always_on
//   - applyTo 非空 -> trigger: glob + globs
//   - description 非空 -> trigger: model_decision + description
//   - 其他 -> trigger: manual
func (a *Antigravity) buildRule(d *parser.Document, cfg *config.Config) writer.FileOp {
	var fm FmBuilder
	applyTo := EffectiveApplyTo(d, a.Name())
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
	opts := MakeOpts(cfg, a.Name(), d.Path, false)
	op := BuildMarkdownFile(
		FilePath(".agents/rules", d.Name, ".md"),
		fm.String(), d.Body, opts,
	)
	if len(op.Content) > antigravityMaxChars {
		op.Reason = fmt.Sprintf("WARN: rule exceeds %d chars; consider splitting", antigravityMaxChars)
	}
	return op
}

// buildSkillAsRule 把 std skill 降级为 Antigravity model_decision rule
// （.agents/skills/ schema UNKNOWN，按调研建议走 rule 形态）
func (a *Antigravity) buildSkillAsRule(d *parser.Document, cfg *config.Config) writer.FileOp {
	var fm FmBuilder
	fm.Add("trigger", "model_decision")
	desc := d.Description
	if desc == "" {
		desc = "Skill: " + d.Name
	}
	fm.Add("description", desc)
	opts := MakeOpts(cfg, a.Name(), d.Path, false)
	return BuildMarkdownFile(
		FilePath(".agents/rules", "skill-"+d.Name, ".md"),
		fm.String(), d.Body, opts,
	)
}

// buildWorkflow -> .agents/workflows/<n>.md，文件名即 slash 名
// frontmatter UNKNOWN，v1.1 输出无 frontmatter 的纯 markdown 步骤序列
func (a *Antigravity) buildWorkflow(d *parser.Document, cfg *config.Config) writer.FileOp {
	opts := MakeOpts(cfg, a.Name(), d.Path, false)
	body := d.Body
	if d.Description != "" {
		body = d.Description + "\n\n" + d.Body
	}
	return BuildMarkdownFile(
		FilePath(".agents/workflows", d.Name, ".md"),
		"", body, opts,
	)
}
