package transformer

import (
	"github.com/StringKe/std-agent/internal/config"
	"github.com/StringKe/std-agent/internal/parser"
	"github.com/StringKe/std-agent/internal/transformer/protocol"
	"github.com/StringKe/std-agent/internal/writer"
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
//   - .github/skills/<n>/SKILL.md（Agent Skills，cloud agent / code review /
//     CLI / VS Code 均已 GA，https://docs.github.com/en/copilot/concepts/agents/about-agent-skills）
//   - .vscode/mcp.json（顶级键 servers，与 Claude 的 mcpServers 不同）
//
// instructions / prompts / agents 的特殊文件后缀由 CopilotProtocol 内部处理。
var copilotAdapter = protocol.Adapter{
	Name:            "copilot",
	RootFileName:    ".github/copilot-instructions.md",
	ManifestSection: "Reference Rules",
	RulesDir:        ".github/instructions",
	CommandsDir:     ".github/prompts",
	SubagentsDir:    ".github/agents",
	SkillsDir:       ".github/skills",
	// 官方 SKILL.md 字段：name（必填=目录名，≤64）/ description（必填，≤1024）/
	// license / compatibility（≤500）/ metadata / allowed-tools（experimental）
	SkillSupportedFields: []string{"name", "description", "license", "compatibility", "metadata", "allowed-tools"},
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
	// Phase 4.2: subagent CLI 调用降级形态 B。copilot 原生支持 .github/agents/<name>.agent.md，
	// 当 SubagentInvokeCmd 非空时 subagent body 加"通过 shell 调用 claude --agent {name}"指引，
	// 让 copilot 的 AI 把 std subagent 委派给 Claude Code 执行（保留 isolated context 语义）。
	SubagentInvokeCmd: "claude --agent {name}",
}
