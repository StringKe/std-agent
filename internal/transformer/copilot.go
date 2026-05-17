package transformer

import (
	"std-ai/internal/config"
	"std-ai/internal/parser"
	"std-ai/internal/transformer/protocol"
	"std-ai/internal/writer"
)

func init() { Register(&Copilot{}) }

// Copilot 是 GitHub Copilot transformer
type Copilot struct{}

// Name 返回 "copilot"
func (c *Copilot) Name() string { return "copilot" }

// Plan 委托 CopilotProtocol 按 copilotAdapter 计算输出
func (c *Copilot) Plan(docs []*parser.Document, cfg *config.Config) (*writer.Plan, error) {
	return protocol.Copilot{}.Plan(FilterDocs(docs, c.Name()), copilotAdapter, cfg)
}

// copilotAdapter 配置 CopilotProtocol 协议族的 copilot 变体
//
// 落点：
//   - .github/copilot-instructions.md（根文件）
//   - .github/instructions/<n>.instructions.md（path-specific rules）
//   - .github/prompts/<n>.prompt.md（slash commands）
//   - .github/agents/<n>.agent.md（subagents 原生）
//   - .vscode/mcp.json（顶级键 servers，与 Claude 的 mcpServers 不同）
//
// instructions / prompts / agents 的特殊文件后缀由 CopilotProtocol 内部处理。
var copilotAdapter = protocol.Adapter{
	Name:                 "copilot",
	RootFileName:         ".github/copilot-instructions.md",
	ManifestSection:      "Reference Rules",
	RulesDir:             ".github/instructions",
	CommandsDir:          ".github/prompts",
	SubagentsDir:         ".github/agents",
	SkillsDir:            "", // fallback 到 instructions/skills/
	ReferencesDir:        "", // fallback 到 instructions/references/
	FallbackDir:          ".github/instructions",
	InjectExplainer:      true,
	InjectStdaiTypeField: true,
	GlobsFieldName:       "applyTo",
	GlobsFieldFormat:     protocol.GlobsCommaString,
	SupportsDescription:  true,
	MCPPath:              ".vscode/mcp.json",
	MCPTopKey:            "servers",
	MaxBytesPerFile:      8000,
	SoftBytes:            4000,
	InjectTypeGlossary:   true,
}
