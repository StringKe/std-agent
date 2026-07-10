package transformer

import (
	"github.com/StringKe/std-agent/internal/config"
	"github.com/StringKe/std-agent/internal/parser"
	"github.com/StringKe/std-agent/internal/transformer/protocol"
	"github.com/StringKe/std-agent/internal/writer"
)

func init() { Register(&Crush{}) }

// Crush 是 Charmbracelet Crush transformer（TUI 同源 Go 工具，5k+ stars）。
//
// 协议族：AgentsMD。crush 通过 defaultContextPaths 同时读 AGENTS.md / CRUSH.md /
// CLAUDE.md / GEMINI.md 等（全部命中按序拼接）。本 transformer 写 CRUSH.md
// 作为 crush 私有根，skills 走 .crush/skills/，其余类型走 .crush/rules/ 子目录 fallback。
//
// skills 闭环：crush 默认只扫描 ~/.config/crush/skills/ 与 ~/.config/agents/skills/
// 两个全局路径，项目级 skills 必须在 crush.json 的 options.skills_paths 显式声明
// （https://charmbracelet-crush.mintlify.app/configuration/skills）。因此 Plan 在产出
// skills 时追加 crush.json 的 JSONMerge op 注册 ".crush/skills"。用户已有 crush.json
// 时深合并（数组并集、scalar 保留用户值），解析失败（注释等）跳过并 WARN。
type Crush struct{}

// Name 返回 "crush"
func (c *Crush) Name() string { return "crush" }

// Plan 委托 AgentsMD 协议按 crushAdapter 计算输出，skills 存在时追加 crush.json 注册
func (c *Crush) Plan(docs []*parser.Document, cfg *config.Config) (*writer.Plan, error) {
	filtered := FilterDocs(docs, c.Name())
	plan, err := protocol.AgentsMD{}.Plan(filtered, crushAdapter, cfg)
	if err != nil {
		return nil, err
	}
	for _, d := range filtered {
		if d.Type == parser.TypeSkills {
			plan.Files = append(plan.Files, writer.FileOp{
				Path:      "crush.json",
				Content:   []byte(`{"options":{"skills_paths":[".crush/skills"]}}`),
				JSONMerge: true,
			})
			break
		}
	}
	return plan, nil
}

// crushAdapter 配置 AgentsMD 协议族的 crush 变体
//
// NestedSupported=false：crush 源码只解析相对 working dir 的 context path，
// 无父目录上溯、无子树自动发现，嵌套 CRUSH.md 只有恰好在该子目录启动 crush
// 才被读到（且此时根 CRUSH.md 丢失），写了弊大于利
// （https://github.com/charmbracelet/crush/blob/main/internal/agent/prompt/prompt.go）。
var crushAdapter = protocol.Adapter{
	Name:                 "crush",
	RootFileName:         "CRUSH.md",
	ManifestSection:      "Reference Rules",
	NestedSupported:      false,
	RulesDir:             "", // crush 无子目录 rules，全 inline 到 CRUSH.md
	SkillsDir:            ".crush/skills",
	CommandsDir:          "", // 无原生 commands，fallback
	ReferencesDir:        "", // fallback
	SubagentsDir:         "", // fallback
	FallbackDir:          ".crush/rules",
	InjectExplainer:      true,
	InjectStdaiTypeField: true,
	SkillSupportedFields: []string{"name", "description", "license", "compatibility", "metadata"},
	InjectTypeGlossary:   true,
	MaxBytesPerFile:      0, // 无明确字节限制
}
