package transformer

import (
	"fmt"

	"std-ai/internal/config"
	"std-ai/internal/parser"
	"std-ai/internal/writer"
)

func init() {
	Register(&OpenCode{})
}

// OpenCode 是 sst/opencode CLI transformer
type OpenCode struct{}

// Name 返回 "opencode"
func (o *OpenCode) Name() string { return "opencode" }

// Plan 计算输出
//
// OpenCode 自动消费 codex transformer 写的根 AGENTS.md，rules 不重复落盘。
// rules 的 applyTo 在 OpenCode 上被丢弃（OpenCode 不支持 frontmatter 条件激活）。
func (o *OpenCode) Plan(docs []*parser.Document, cfg *config.Config) (*writer.Plan, error) {
	plan := &writer.Plan{Target: o.Name()}
	docs = FilterDocs(docs, o.Name())
	if len(docs) == 0 {
		return plan, nil
	}
	skills := FilterByType(docs, parser.TypeSkills)
	commands := FilterByType(docs, parser.TypeCommands)
	SortDocs(skills)
	SortDocs(commands)

	for _, d := range skills {
		plan.Files = append(plan.Files, o.buildAgent(d, cfg))
	}
	for _, d := range commands {
		plan.Files = append(plan.Files, o.buildCommand(d, cfg))
	}
	return plan, nil
}

func (o *OpenCode) buildAgent(d *parser.Document, cfg *config.Config) writer.FileOp {
	var fm FmBuilder
	fm.Add("mode", "subagent")
	fm.Add("description", MergeDescription(d.Description, d.WhenToUse))
	fm.Add("model", d.Model)
	// permission v1.0 简化：所有动作走 ask（保守安全），由用户按需放宽
	fm.AddRaw("permission", "{ edit: ask, bash: ask, read: allow, glob: allow, grep: allow, list: allow, task: ask, lsp: allow }")
	opts := MakeOpts(cfg, o.Name(), d.Path, false)
	op := BuildMarkdownFile(
		FilePath(".opencode/agents", d.Name, ".md"),
		fm.String(), d.Body, opts,
	)
	if len(d.SkillFiles) > 0 {
		op.Reason = fmt.Sprintf("WARN: %d SKILL package 辅助文件被忽略，opencode agent 是单文件不支持子目录（参考 docs/spec.md 4.5 降级链）", len(d.SkillFiles))
	}
	return op
}

func (o *OpenCode) buildCommand(d *parser.Document, cfg *config.Config) writer.FileOp {
	var fm FmBuilder
	fm.Add("description", d.Description)
	fm.Add("model", d.Model)
	opts := MakeOpts(cfg, o.Name(), d.Path, false)
	return BuildMarkdownFile(
		FilePath(".opencode/commands", d.Name, ".md"),
		fm.String(), d.Body, opts,
	)
}
