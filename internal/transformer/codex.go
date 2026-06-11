package transformer

import (
	"std-ai/internal/config"
	"std-ai/internal/parser"
	"std-ai/internal/transformer/protocol"
	"std-ai/internal/writer"
)

func init() { Register(&Codex{}) }

// Codex 是 OpenAI Codex CLI transformer
type Codex struct{}

// Name 返回 "codex"
func (c *Codex) Name() string { return "codex" }

// Plan 委托 AgentsMD 协议按 codexAdapter 计算输出
func (c *Codex) Plan(docs []*parser.Document, cfg *config.Config) (*writer.Plan, error) {
	return protocol.AgentsMD{}.Plan(FilterDocs(docs, c.Name()), codexAdapter, cfg)
}

// codexAdapter 配置 AgentsMD 协议族的 codex 变体。
//
// RulesDir 为空：Codex 项目级唯一指令机制是 AGENTS.md（project_doc_max_bytes
// 默认 32768 字节），nonRoot rules 全文 inline（amp / warp 同风格）。曾经的
// RulesDir=".codex/memories" 与官方语义冲突，已废弃：Codex 的 memories 是
// ~/.codex/memories/ 用户级自动记忆系统（无项目级目录，mem v2 起单根），而
// 项目级 .codex/ 是 Team Config 配置目录（config.toml / rules/*.rules
// execpolicy / skills/），且被沙箱与 .git 同级 carveout 保护，不适合放
// markdown 指令。旧产物由 runner 的 legacyCodexMemoriesOrphans 自动清理。
//
// FallbackDir=".agents"：references / subagents 降级落 .agents/<subdir>/，
// 与官方 .agents/skills/（Agent Skills 协议）同命名空间，不与 antigravity
// 的 .agents/rules/ 冲突。
var codexAdapter = protocol.Adapter{
	Name:                  "codex",
	RootFileName:          "AGENTS.md",
	NestedSupported:       true,
	RulesDir:              "",
	SkillsDir:             ".agents/skills",
	FallbackDir:           ".agents",
	InjectExplainer:       true,
	InjectStdaiTypeField:  true,
	SkillSupportedFields:  []string{"name", "description", "license", "compatibility", "metadata"},
	CommandFormat:         protocol.CommandSkillPrefix,
	CommandsAsSkillSubdir: "commands", // v3：子目录隔离（避免 cmd- 私有前缀污染 skill 命名空间）
	InjectCommandsToRoot:  true,
	InjectTypeGlossary:    true,
}
