package transformer

import (
	"github.com/StringKe/std-agent/internal/config"
	"github.com/StringKe/std-agent/internal/parser"
	"github.com/StringKe/std-agent/internal/transformer/protocol"
	"github.com/StringKe/std-agent/internal/writer"
)

func init() { Register(&Muse{}) }

// Muse 是 Meta Muse Code transformer（二进制名 `muse`，模型 Muse Spark）。
//
// 协议族：AgentsMD。`muse init` 只写项目根单个 AGENTS.md，不建其他文件；
// 加载时从 workspace root 向上 walk 到最近 .git 边界，每层按
// AGENTS.md、CLAUDE.md、.agents/AGENTS.md、.claude/CLAUDE.md 顺序取首个
// 存在文件，深层覆盖浅层；项目级规则须 trust workspace 后才加载。
// 项目 skills 在 `<repo>/.agents/skills/<skill-id>/SKILL.md`（与 Codex /
// Goose / Zed 共享命名空间），另扫描仓库内 .codex/skills 与 .claude/skills；
// 用户级 skills 在 `$XDG_CONFIG_HOME/muse/skills` 与 `~/.agents/skills`。
// 项目记忆（references 对应物）在 `<repo>/.agents/memory/`（MEMORY.md 索引 +
// 每主题一 markdown）；项目 hooks 在 `.muse/hooks.json`。
// 来源：https://dev.meta.ai/docs/muse-code/configuration.md
// 与 https://dev.meta.ai/docs/muse-code/extending.md（skills / hooks / MCP 章节）。
//
// 调研：docs/targets/muse.md。
type Muse struct{}

// Name 返回 "muse"
func (m *Muse) Name() string { return "muse" }

// Plan 委托 AgentsMD 协议按 museAdapter 计算输出
func (m *Muse) Plan(docs []*parser.Document, cfg *config.Config) (*writer.Plan, error) {
	return protocol.AgentsMD{}.Plan(FilterDocs(docs, m.Name()), museAdapter, cfg)
}

// museAdapter 按 Muse Code 官方文档配置
//
//   - RulesDir 留空：官方无项目级 rules 目录，`muse init` 只写单个 AGENTS.md，
//     nonRoot rules 全文 inline（codex / goose 同风格）
//   - SkillsDir=".agents/skills"：官方项目 skills 路径；SkillSupportedFields 与
//     Codex / Goose / Zed 同集，保证共享 SKILL.md 字节一致
//   - CommandsAsSkillSubdir="commands"：官方无 commands 原生目录，slash 调用即
//     skill 调用（spec-kit 集成把 skills 装进 .agents/skills 并以 /<command> 调用），
//     故 commands 降级为 skill 子目录（goose 同风格）
//   - ReferencesDir=".agents/memory"：官方项目记忆目录（MEMORY.md 索引 + 主题文件）；
//     协议只写主题文件，MEMORY.md 索引需用户手工维护（见 docs/targets/muse.md）
//   - SubagentsDir 留空：subagent 经 lead session 工具派生（可选 worktree 隔离），
//     无 markdown 文件格式，走 FallbackDir=".muse"（官方项目 hooks 目录所在）
//   - MCP / settings.json（~/.config/muse/settings.json，mcp_servers 块）是用户级
//     配置，stdagent 不写
var museAdapter = protocol.Adapter{
	Name:                  "muse",
	RootFileName:          "AGENTS.md",
	ManifestSection:       "Reference Rules",
	NestedSupported:       true,
	RulesDir:              "",
	SkillsDir:             ".agents/skills",
	SkillSupportedFields:  []string{"name", "description", "license", "compatibility", "metadata"},
	CommandFormat:         protocol.CommandSkillPrefix,
	CommandsAsSkillSubdir: "commands",
	ReferencesDir:         ".agents/memory",
	SubagentsDir:          "",
	FallbackDir:           ".muse",
	InjectExplainer:       true,
	InjectStdaiTypeField:  true,
	InjectTypeGlossary:    true,
}
