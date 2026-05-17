package protocol

import _ "embed"

//go:embed glossary.md
var glossaryMarkdown string

// RenderGlossaryFor 返回 std-ai 类型速查 markdown 段。
//
// 协议实现把它 prepend 到根文件 body 之前（CLAUDE.md / AGENTS.md / GEMINI.md /
// .github/copilot-instructions.md 等）。adapter.InjectTypeGlossary=false 时
// 返回空串，调用方按空串处理即可（不需要 if 分支）。
func RenderGlossaryFor(adapter Adapter) string {
	if !adapter.InjectTypeGlossary {
		return ""
	}
	return glossaryMarkdown
}
