package cli

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

//go:embed explain_text.md
var explainText string

// explainType 是 stdagent explain --json 输出的单条记录
type explainType struct {
	Type      string `json:"type"`
	Semantics string `json:"semantics"`
	WhenToUse string `json:"when_to_use"`
	WhenNot   string `json:"when_not"`
	ExampleFM string `json:"example_frontmatter"`
}

// explainTypes 是 5 种 type 的结构化语义，供 --json 输出。
// 与 explain_text.md 内容对应，保持同步。
var explainTypes = []explainType{
	{
		Type:      "rules",
		Semantics: "持续生效的编码、架构和操作约束；由 target 自动加载或按路径匹配。",
		WhenToUse: "违反会造成真实风险、且值得占用常驻上下文的稳定约束。",
		WhenNot:   "长背景用 references，按需工作流用 skills，用户模板用 commands。",
		ExampleFM: "---\ntype: rules\nname: exception-handling\ndescription: Go 错误传播与边界转换\npriority: high\napplyTo:\n  - \"**/*.go\"\n---",
	},
	{
		Type:      "skills",
		Semantics: "AI 根据 description 按需调用的能力包，可携带辅助资源。",
		WhenToUse: "可复用、有明确结果和成功标准的工作流。",
		WhenNot:   "持续约束用 rules，用户显式模板用 commands。",
		ExampleFM: "---\ntype: skills\nname: code-review\ndescription: 审查当前改动并报告正确性、安全和回归问题\n---",
	},
	{
		Type:      "commands",
		Semantics: "用户输入 /command-name 显式触发的操作模板。",
		WhenToUse: "用户需要主动调用的固定操作。",
		WhenNot:   "AI 自动判断的流程用 skills，持续约束用 rules。",
		ExampleFM: "---\ntype: commands\nname: review\ndescription: 审查当前分支改动并生成 review 报告\n---",
	},
	{
		Type:      "references",
		Semantics: "仅在需要时查阅的架构、协议、API 和长篇背景。",
		WhenToUse: "不应占用默认上下文但需要保留的领域知识。",
		WhenNot:   "持续约束用 rules，可执行工作流用 skills。",
		ExampleFM: "---\ntype: references\nname: transformer-design\ndescription: transformer 协议层架构说明\napplyTo:\n  - \"internal/transformer/**\"\n---",
	},
	{
		Type:      "subagents",
		Semantics: "在隔离上下文中执行的代理定义。",
		WhenToUse: "可独立执行、需要专门上下文或可安全并行的任务。",
		WhenNot:   "当前 session 内的流程用 skills，简单模板用 commands。",
		ExampleFM: "---\ntype: subagents\nname: code-reviewer\ndescription: 在隔离上下文中审查代码并返回问题清单\n---",
	},
}

func newExplainCmd() *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "explain [type]",
		Short: "解释 std-agent 5 种类型（rules/skills/commands/references/subagents）的语义",
		Long: `输出 std-agent 5 种 type 的语义速查：每种类型的触发语义 / 何时使用 / 何时不用 / 示例 frontmatter。

不带参数时输出全部 5 种。带 type 参数时只输出该 type 一段。

示例：

    stdagent explain                  # 全部 5 种
    stdagent explain rules            # 只看 rules
    stdagent explain --json           # JSON 输出（AI 集成）
    stdagent explain rules --json     # rules 单项 JSON
`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var filter string
			if len(args) > 0 {
				filter = strings.ToLower(strings.TrimSpace(args[0]))
			}
			if asJSON {
				return writeExplainJSON(cmd, filter)
			}
			return writeExplainMarkdown(cmd, filter)
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "JSON 输出（给 AI / 自动化集成）")
	return cmd
}

// writeExplainMarkdown 输出 markdown 速查。filter 为空输出全部，否则只输出对应 type 段。
func writeExplainMarkdown(cmd *cobra.Command, filter string) error {
	if filter == "" {
		cmd.Print(explainText)
		return nil
	}
	if !isKnownType(filter) {
		return fmt.Errorf("unknown type %q (valid: rules, skills, commands, references, subagents)", filter)
	}
	section := extractSection(explainText, filter)
	if section == "" {
		return fmt.Errorf("section for type %q not found in explain_text.md", filter)
	}
	cmd.Print(section)
	return nil
}

// writeExplainJSON 输出 []explainType（全部）或 [explainType]（单项过滤）。
func writeExplainJSON(cmd *cobra.Command, filter string) error {
	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	if filter == "" {
		return enc.Encode(explainTypes)
	}
	if !isKnownType(filter) {
		return fmt.Errorf("unknown type %q (valid: rules, skills, commands, references, subagents)", filter)
	}
	for _, t := range explainTypes {
		if t.Type == filter {
			return enc.Encode([]explainType{t})
		}
	}
	return fmt.Errorf("type %q not found in explainTypes", filter)
}

func isKnownType(t string) bool {
	for _, e := range explainTypes {
		if e.Type == t {
			return true
		}
	}
	return false
}

// extractSection 从 explain_text.md 抽出 ## <type> 标题对应的段落（含标题，到下一个 ## 之前）。
//
// explain_text.md 用 `## rules` / `## skills` 等二级标题分段，最后有一个 `## 速查表` 总表段。
// 单项查询时只返回该 type 的段，不包含速查表（避免重复信息）。
func extractSection(text, typeName string) string {
	header := "## " + typeName
	idx := strings.Index(text, header+"\n")
	if idx < 0 {
		return ""
	}
	rest := text[idx:]
	// 找下一个 `## ` 二级标题（注意要在行首）
	next := strings.Index(rest[len(header):], "\n## ")
	if next < 0 {
		return rest
	}
	return rest[:len(header)+next+1] // +1 包含 next 之前的换行
}
