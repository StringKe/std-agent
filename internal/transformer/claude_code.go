package transformer

import (
	"github.com/StringKe/std-agent/internal/config"
	"github.com/StringKe/std-agent/internal/parser"
	"github.com/StringKe/std-agent/internal/transformer/protocol"
	"github.com/StringKe/std-agent/internal/writer"
)

func init() { Register(&ClaudeCode{}) }

// ClaudeCode 是 Anthropic Claude Code transformer
type ClaudeCode struct{}

// Name 返回 "claude-code"
func (c *ClaudeCode) Name() string { return "claude-code" }

// Plan 委托 ClaudeMD 协议按 claudeCodeAdapter 计算输出
func (c *ClaudeCode) Plan(docs []*parser.Document, cfg *config.Config) (*writer.Plan, error) {
	return protocol.ClaudeMD{}.Plan(FilterDocs(docs, c.Name()), claudeCodeAdapter, cfg)
}

// claudeCodeAdapter 配置 Anthropic Claude Code 协议方言。
//
// 私有字段（when_to_use / argument-hint / tools / effort / paths / hooks /
// agent / shell / context / arguments / model / disable-model-invocation）已在
// ClaudeMDProtocol 内部 helper renderClaudeSkillFrontmatter 处理，
// 不下放到 Adapter，让其他 protocol 不必感知 Anthropic 方言。
var claudeCodeAdapter = protocol.Adapter{
	Name:                 "claude-code",
	RootFileName:         "CLAUDE.md",
	ManifestSection:      "Imported Rules",
	NestedSupported:      true,
	RulesDir:             ".claude/rules",
	SkillsDir:            ".claude/skills",
	CommandsDir:          ".claude/commands",
	SubagentsDir:         ".claude/agents",
	ReferencesDir:        "",        // 无原生 references
	FallbackDir:          ".claude", // 不能用 .claude/rules：该目录会被全量自动加载
	InjectExplainer:      true,
	InjectStdaiTypeField: true,
	GlobsFieldName:       "paths", // Anthropic 私有方言
	GlobsFieldFormat:     protocol.GlobsList,
	SupportsAlwaysApply:  false,
	SupportsDescription:  true,
	MCPPath:              ".mcp.json",
	MCPTopKey:            "mcpServers",
	InjectTypeGlossary:   true,
}
