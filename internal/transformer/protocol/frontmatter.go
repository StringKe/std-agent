package protocol

import (
	"strings"

	"github.com/StringKe/std-agent/internal/parser"
	"github.com/StringKe/std-agent/internal/transformer/transformerutil"
)

// RenderGlobs 按 adapter 配置渲染 globs 类 frontmatter 字段。
//
// field 是字段名（如 "globs" / "paths" / "applyTo"）；为 "" 返回空串（跳过）。
// format 决定渲染形式：GlobsList -> YAML list / GlobsCommaString -> 逗号分隔字符串。
// vals 为空也返回空串。
//
// 返回的字符串以 "\n" 结尾（与 FmBuilder 输出一致），调用方可直接拼到
// frontmatter 文本中；若返回为空表示"未渲染任何内容"。
func RenderGlobs(field string, format GlobsFormat, vals []string) string {
	if field == "" || len(vals) == 0 {
		return ""
	}
	var fm transformerutil.FmBuilder
	switch format {
	case GlobsCommaString:
		fm.Add(field, strings.Join(vals, ","))
	case GlobsList:
		fm.AddList(field, vals)
	default:
		fm.AddList(field, vals)
	}
	return fmBodyOnly(&fm)
}

// RenderTriggerFrontmatter 按 TriggerMode 渲染 doc 对应的 trigger 类字段。
//
// 各模式语义：
//   - TriggerNone：返回 ""（不写任何字段）
//   - TriggerAlwaysApply：写 alwaysApply: true 当 d.AlwaysApply=true，否则空
//   - TriggerTrigger（windsurf 系）：按 doc 推断 always_on / glob / model_decision / manual
//   - TriggerApplyTo（copilot 系）：写 applyTo: <逗号分隔 globs>
//   - TriggerInclusion（Kiro steering）：写 inclusion: always / fileMatch / auto / manual
//
// 返回的字符串以 "\n" 结尾（FmBuilder 风格），可直接拼到 frontmatter 文本中；
// 空串表示该模式下未渲染任何内容。
func RenderTriggerFrontmatter(mode TriggerMode, doc *parser.Document) string {
	if doc == nil || mode == TriggerNone {
		return ""
	}
	var fm transformerutil.FmBuilder
	applyTo := doc.ApplyTo
	switch mode {
	case TriggerAlwaysApply:
		if doc.AlwaysApply {
			fm.AddBool("alwaysApply", true)
		}
	case TriggerTrigger:
		switch {
		case doc.AlwaysApply:
			fm.Add("trigger", "always_on")
		case len(applyTo) > 0:
			fm.Add("trigger", "glob")
			fm.AddList("globs", applyTo)
		case doc.Description != "":
			fm.Add("trigger", "model_decision")
			fm.Add("description", doc.Description)
		default:
			fm.Add("trigger", "manual")
		}
	case TriggerApplyTo:
		if len(applyTo) > 0 {
			fm.Add("applyTo", strings.Join(applyTo, ","))
		}
	case TriggerInclusion:
		switch {
		case doc.AlwaysApply:
			fm.Add("inclusion", "always")
		case len(applyTo) > 0:
			fm.Add("inclusion", "fileMatch")
			if len(applyTo) == 1 {
				fm.Add("fileMatchPattern", applyTo[0])
			} else {
				fm.AddList("fileMatchPattern", applyTo)
			}
		case doc.Description != "":
			fm.Add("inclusion", "auto")
			fm.Add("name", doc.Name)
			fm.Add("description", doc.Description)
		default:
			fm.Add("inclusion", "manual")
		}
	}
	return fmBodyOnly(&fm)
}

// fmBodyOnly 从 FmBuilder.String() 中剥掉外层 "---\n" ... "---\n" 包裹，
// 返回纯字段行（含末尾 \n）。空 FmBuilder 返回 ""。
//
// 这个 helper 让 RenderGlobs / RenderTriggerFrontmatter 输出可被外层 Builder
// 进一步拼装；最终包裹 "---" 由调用方一次完成。
func fmBodyOnly(fm *transformerutil.FmBuilder) string {
	s := fm.String()
	if s == "" {
		return ""
	}
	s = strings.TrimPrefix(s, "---\n")
	s = strings.TrimSuffix(s, "---\n")
	return s
}
