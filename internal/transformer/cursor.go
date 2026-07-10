package transformer

import (
	"github.com/StringKe/std-agent/internal/config"
	"github.com/StringKe/std-agent/internal/parser"
	"github.com/StringKe/std-agent/internal/transformer/protocol"
	"github.com/StringKe/std-agent/internal/writer"
)

func init() { Register(&Cursor{}) }

// Cursor 是 Cursor IDE transformer
type Cursor struct{}

// Name 返回 "cursor"
func (c *Cursor) Name() string { return "cursor" }

// Plan 委托 Cursor 协议按 cursorAdapter 计算输出
func (c *Cursor) Plan(docs []*parser.Document, cfg *config.Config) (*writer.Plan, error) {
	return protocol.Cursor{}.Plan(FilterDocs(docs, c.Name()), cursorAdapter, cfg)
}

var cursorAdapter = protocol.Adapter{
	Name:                 "cursor",
	RulesDir:             ".cursor/rules",
	SkillsDir:            ".cursor/skills",
	CommandsDir:          ".cursor/commands",
	FallbackDir:          ".cursor/rules",
	InjectExplainer:      true,
	InjectStdaiTypeField: true,
	GlobsFieldName:       "globs",
	GlobsFieldFormat:     protocol.GlobsCommaString,
	SupportsAlwaysApply:  true,
	SupportsDescription:  true,
	MCPPath:              ".cursor/mcp.json",
	MCPTopKey:            "mcpServers",
	InjectTypeGlossary:   true,
}
