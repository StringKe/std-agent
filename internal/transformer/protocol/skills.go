package protocol

import (
	"slices"

	"github.com/StringKe/std-agent/internal/config"
	"github.com/StringKe/std-agent/internal/parser"
	"github.com/StringKe/std-agent/internal/transformer/transformerutil"
	"github.com/StringKe/std-agent/internal/writer"
)

// BuildNativeSkillPackage 在 adapter.SkillsDir 下输出 Agent Skills 标准包
// （<SkillsDir>/<name>/SKILL.md + SkillFiles 辅助文件）。
//
// 原生落点专用：不注入 explainer / std-agent-type（那是 degraded 语义），
// frontmatter 受 adapter.SkillSupportedFields 白名单约束。
// AgentsMD / Clinerules / Copilot 协议共用。
func BuildNativeSkillPackage(d *parser.Document, a Adapter, cfg *config.Config) []writer.FileOp {
	skillDir := transformerutil.SkillDir(a.SkillsDir, d.Name)
	fm := buildSkillFrontmatter(d, a)
	opts := transformerutil.MakeOpts(cfg, a.Name, d.Path, false)
	skillMd := transformerutil.BuildMarkdownFile(skillDir+"/SKILL.md", fm, d.Body, opts)
	return transformerutil.BuildSkillPackage(skillDir, skillMd, d.SkillFiles)
}

// buildSkillFrontmatter 按 SkillSupportedFields 白名单生成 frontmatter
//
// 空白名单 -> 默认 Agent Skills 标准字段集（name / description / license /
// compatibility / metadata）。白名单非空时仅渲染白名单字段。
// 可选扩展字段（工具声明支持才渲染）：when_to_use / allowed-tools / paths /
// disable-model-invocation / user-invocable。
func buildSkillFrontmatter(d *parser.Document, a Adapter) string {
	allowed := a.SkillSupportedFields
	if len(allowed) == 0 {
		allowed = []string{"name", "description", "license", "compatibility", "metadata"}
	}
	in := func(k string) bool { return slices.Contains(allowed, k) }
	var fm transformerutil.FmBuilder
	if in("name") {
		fm.Add("name", d.Name)
	}
	if in("description") {
		// 不支持 when_to_use 的 target 把 when_to_use 拼到 description 末尾
		desc := d.Description
		if !in("when_to_use") {
			desc = transformerutil.MergeDescription(d.Description, d.WhenToUse)
		}
		fm.Add("description", desc)
	}
	if in("when_to_use") {
		fm.Add("when_to_use", d.WhenToUse)
	}
	if in("allowed-tools") {
		fm.AddList("allowed-tools", d.AllowedTools)
	}
	if in("paths") {
		fm.AddList("paths", transformerutil.EffectiveApplyTo(d, a.Name))
	}
	if in("disable-model-invocation") && d.DisableModelInvocation {
		fm.AddBool("disable-model-invocation", true)
	}
	if in("license") {
		fm.Add("license", d.License)
	}
	if in("compatibility") {
		fm.Add("compatibility", d.Compatibility)
	}
	if in("metadata") {
		fm.AddMap("metadata", d.Metadata)
	}
	return fm.String()
}
