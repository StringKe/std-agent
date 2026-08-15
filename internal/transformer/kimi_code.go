package transformer

import (
	"github.com/StringKe/std-agent/internal/config"
	"github.com/StringKe/std-agent/internal/parser"
	"github.com/StringKe/std-agent/internal/transformer/protocol"
	"github.com/StringKe/std-agent/internal/writer"
)

func init() { Register(&KimiCode{}) }

// KimiCode 是 Moonshot AI Kimi Code CLI transformer。
//
// 协议族：AgentsMD。Kimi Code CLI（github.com/MoonshotAI/kimi-code，npm
// @moonshot-ai/kimi-code）是 kimi-cli 的继任产品，原生消费根 AGENTS.md
// （层级发现：项目根到 cwd 逐级合并，leaf-first 32KiB 预算），原生 Agent
// Skills 扫描 `.kimi-code/skills/` 与共享的 `.agents/skills/`
// （https://moonshotai.github.io/kimi-code/en/customization/skills.html）。
//
// 无独立 rules 目录（Issue #1747 的 rules 提案仍 open 未实现），nonRoot rules
// 全 inline 到 AGENTS.md；无独立 commands 机制（skill 自动注册为
// /skill:<name> slash command），commands 与 codex / amp 相同降级为
// .agents/skills/commands/<n>/SKILL.md。subagents 是 YAML agent spec +
// `--agent-file` CLI 加载，无项目内自动发现目录且字段体系不兼容
// （tools 用 module:ClassName），走 fallback。
type KimiCode struct{}

// Name 返回 "kimi-code"
func (k *KimiCode) Name() string { return "kimi-code" }

// Plan 计算输出
func (k *KimiCode) Plan(docs []*parser.Document, cfg *config.Config) (*writer.Plan, error) {
	return protocol.AgentsMD{}.Plan(FilterDocs(docs, k.Name()), kimiCodeAdapter, cfg)
}

// kimiCodeAdapter 配置 AgentsMD 协议族的 kimi-code 变体
//
// skills / commands 落点与 codex / amp / warp / antigravity 完全一致
// （.agents/skills/ 共享命名空间 + SkillSupportedFields 同集），五个 target
// 产出字节相同、writer 按 unchanged 去重；kimi-code 官方也把 `.agents/skills/`
// 列为 Project 层扫描路径，无需再写私有 `.kimi-code/skills/` 副本。
// 注意 Kimi 的 SKILL.md 无 allowed-tools 字段，字段集里不能加工具字段。
//
// FallbackDir 用私有 `.kimi-code/rules`（degraded 文件的 explainer 含 target
// 名，不能落共享 .agents/ 否则与 codex 互相改写）；`.kimi-code/` 下只有
// AGENTS.md / skills/ / mcp.json / local.toml 被官方读取，rules/ 子目录不会
// 被自动加载，放低频降级物安全。
var kimiCodeAdapter = protocol.Adapter{
	Name:                  "kimi-code",
	RootFileName:          "AGENTS.md",
	ManifestSection:       "Reference Rules",
	NestedSupported:       true,
	RulesDir:              "", // kimi-code 无子目录 rules，全 inline 到 AGENTS.md
	SkillsDir:             ".agents/skills",
	SkillSupportedFields:  []string{"name", "description", "license", "compatibility", "metadata"},
	CommandFormat:         protocol.CommandSkillPrefix,
	CommandsAsSkillSubdir: "commands",
	ReferencesDir:         "",
	SubagentsDir:          ".kimi-code/agents",
	FallbackDir:           ".kimi-code/rules",
	InjectExplainer:       true,
	InjectStdaiTypeField:  true,
	InjectTypeGlossary:    true,
}
