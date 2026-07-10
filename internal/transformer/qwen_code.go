package transformer

import (
	"github.com/StringKe/std-agent/internal/config"
	"github.com/StringKe/std-agent/internal/parser"
	"github.com/StringKe/std-agent/internal/transformer/protocol"
	"github.com/StringKe/std-agent/internal/writer"
)

func init() { Register(&QwenCode{}) }

// QwenCode 是 Alibaba Qwen Code transformer
//
// 协议族：AgentsMD。Qwen Code 是 Gemini CLI 的 fork（github.com/QwenLM/qwen-code，
// 22k stars），优先读 `QWEN.md`，fallback 读 `AGENTS.md`，commands 走
// `.qwen/commands/`。与 gemini-cli 的关键差异：commands 是 markdown 而非 TOML，
// 因此可以完全复用 AgentsMD 协议族的原生 commands 渲染（CommandsDir）。
//
// 设计：
//   - 根文件 QWEN.md（用户自带 fallback AGENTS.md，无需在 transformer 侧重复写）
//   - 无子目录 rules（与 gemini-cli 同源），nonRoot rules 全部 inline 到 QWEN.md
//   - 中国大陆开发者覆盖（qwen 在国内可用性更稳定）
type QwenCode struct{}

// Name 返回 "qwen-code"
func (q *QwenCode) Name() string { return "qwen-code" }

// Plan 委托 AgentsMD 协议按 qwenCodeAdapter 计算输出
func (q *QwenCode) Plan(docs []*parser.Document, cfg *config.Config) (*writer.Plan, error) {
	return protocol.AgentsMD{}.Plan(FilterDocs(docs, q.Name()), qwenCodeAdapter, cfg)
}

// qwenCodeAdapter 描述 Qwen Code 的根文件协议族行为。
//
// 与 gemini-cli adapter 区别：commands 走原生 markdown（不是 TOML），
// 因此 CommandsDir 直接生效，无需 transformer 自定义 buildCommandTOML。
// skills / references / subagents 全部走 fallback 到 .qwen/rules/ 子目录。
var qwenCodeAdapter = protocol.Adapter{
	Name:                 "qwen-code",
	RootFileName:         "QWEN.md",
	ManifestSection:      "Reference Rules",
	NestedSupported:      true,
	RulesDir:             "", // 与 gemini-cli 同源，无子目录 rules，全 inline 到 QWEN.md
	CommandsDir:          ".qwen/commands",
	SkillsAsRule:         false, // RulesDir 为空时 SkillsAsRule 会把 skill 写到仓库根（与 amp 同理），改走 BuildDegradedSkillPackage -> .qwen/rules/skills/<name>/SKILL.md
	ReferencesDir:        "",
	SubagentsDir:         "",
	FallbackDir:          ".qwen/rules",
	InjectExplainer:      true,
	InjectStdaiTypeField: true,
	InjectTypeGlossary:   true,
}
