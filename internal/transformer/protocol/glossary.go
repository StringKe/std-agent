package protocol

import (
	_ "embed"

	"github.com/StringKe/std-agent/internal/config"
)

//go:embed glossary.md
var glossaryMarkdown string

// RenderGlossaryFor 返回 std-agent 类型速查 markdown 段。
//
// 协议实现把它 prepend 到根文件 body 之前（CLAUDE.md / AGENTS.md / GEMINI.md /
// .github/copilot-instructions.md 等）。只有 adapter 支持且项目配置显式开启时
// 才返回内容。
func RenderGlossaryFor(adapter Adapter, cfg *config.Config) string {
	if !adapter.InjectTypeGlossary || cfg == nil || !cfg.InjectTypeGlossary {
		return ""
	}
	return glossaryMarkdown
}
