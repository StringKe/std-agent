package transformer

import (
	"github.com/StringKe/std-agent/internal/config"
	"github.com/StringKe/std-agent/internal/parser"
	"github.com/StringKe/std-agent/internal/transformer/protocol"
	"github.com/StringKe/std-agent/internal/writer"
)

func init() { Register(&GrokBuild{}) }

// GrokBuild 是 xAI 官方 agentic CLI transformer（二进制名 `grok`）。
//
// 协议族：AgentsMD。读项目根 AGENTS.md（family：AGENTS.md / Agents.md / AGENT.md），
// 从 cwd 向上 walk 到 git root 全部叠加。零配置兼容 Claude Code（自动读 CLAUDE.md
// / .claude/rules/ / .claude/skills/ / .claude/agents/），所以 stdagent 只需要
// 写 AGENTS.md + .grok/skills/<name>/SKILL.md（大写）。
//
// 与 superagent-ai / alphaonedev 等社区 grok-cli fork 完全无关——后者用
// .grok/GROK.md / AGENTS.override.md 等非官方约定，本 adapter 不支持。
//
// 调研：docs/targets/grok-build.md（基于 https://docs.x.ai/build/overview）。
type GrokBuild struct{}

// Name 返回 "grok-build"
func (g *GrokBuild) Name() string { return "grok-build" }

// Plan 委托 AgentsMD 协议按 grokBuildAdapter 计算输出
func (g *GrokBuild) Plan(docs []*parser.Document, cfg *config.Config) (*writer.Plan, error) {
	return protocol.AgentsMD{}.Plan(FilterDocs(docs, g.Name()), grokBuildAdapter, cfg)
}

// grokBuildAdapter 按 xAI 官方文档配置
//
//   - 主指令 AGENTS.md（多层 walk 叠加，无 frontmatter）
//   - SkillsDir=".grok/skills"（Agent Skills 标准，SKILL.md 大写）
//   - CommandsDir=".grok/commands"（官方 user-guide 的扁平 slash markdown）
//   - SubagentsDir=".grok/agents"（官方 agent 定义目录）
//   - FallbackDir=".grok/docs"（不能用 .grok/rules——那是 Grok 每 session 全量
//     加载的原生 rules 目录，把 references 降级物放进去等于把低频
//     参考资料永久注入 context）
var grokBuildAdapter = protocol.Adapter{
	Name:                 "grok-build",
	RootFileName:         "AGENTS.md",
	ManifestSection:      "Reference Rules",
	NestedSupported:      true,
	RulesDir:             "",
	SkillsDir:            ".grok/skills",
	SkillSupportedFields: []string{"name", "description", "license", "compatibility", "metadata", "allowed-tools", "disable-model-invocation"},
	CommandsDir:          ".grok/commands",
	SubagentsDir:         ".grok/agents",
	ReferencesDir:        "",
	FallbackDir:          ".grok/docs",
	InjectExplainer:      true,
	InjectStdaiTypeField: true,
	InjectTypeGlossary:   true,
}
