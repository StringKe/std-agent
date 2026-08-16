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
// skills：当前源码默认扫描 .crush/skills（以及 .agents/.claude/.cursor/skills）。
// crush.json options.skills_paths 仍可用于额外路径；继续 JSONMerge 注册
// ".crush/skills" 对旧版 Crush 无害，新版则是冗余声明。
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
