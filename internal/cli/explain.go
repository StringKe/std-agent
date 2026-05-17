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
		Semantics: "项目级编码规范 / 操作约束。AI 自动加载（applyTo 匹配 / 全局）。session 开始就遵守。",
		WhenToUse: "写\"必须遵守\"的硬约束。每条 < 8000 字符，high priority 多条合计 < 16000 字符。",
		WhenNot:   "大段背景知识 -> 用 references。一次性任务步骤 -> 用 commands。需要主动判断是否调用 -> 用 skills。",
		ExampleFM: "---\ntype: rules\nname: exception-handling\ndescription: 异常处理规范\npriority: high\napplyTo:\n  - \"**/*.go\"\n---",
	},
	{
		Type:      "skills",
		Semantics: "按需触发的能力包。AI 看到 description 匹配用户意图时主动调用。可以含多文件子目录（skills/<name>/SKILL.md + 辅助资源）。",
		WhenToUse: "写\"AI 在 X 场景应当做 Y\"的工作流（例如 commit 助手、代码审查流程、调试流程）。description 必须明确触发场景。",
		WhenNot:   "硬规则全程生效 -> 用 rules。用户显式调用的固定模板 -> 用 commands。",
		ExampleFM: "---\ntype: skills\nname: developer-commit\ndescription: Git 提交助手。自动分析变更并生成 Conventional Commits 信息。触发短语：commit、提交。\n---",
	},
	{
		Type:      "commands",
		Semantics: "用户输入 /command-name 触发的模板。AI 不主动调用，等用户显式输入 slash command。",
		WhenToUse: "写固定流程的\"操作宏\"（例如 /review、/done），用户每次需要时手动触发。",
		WhenNot:   "AI 自动判断的工作流 -> 用 skills。session 全程生效的约束 -> 用 rules。",
		ExampleFM: "---\ntype: commands\nname: review\ndescription: 审查当前分支改动并生成 review 报告\n---",
	},
	{
		Type:      "references",
		Semantics: "背景参考资料 / 设计文档 / API 速查。AI 仅在需要时查阅，不自动加载到上下文（按需读 / 用 stdagent which 查询）。",
		WhenToUse: "长篇知识库（架构说明、协议规格、外部 API 列表）。容量超过 rules 限制（> 8000 字符）的内容。",
		WhenNot:   "每次都要遵守的约束 -> 用 rules。可以主动触发的工作流 -> 用 skills。",
		ExampleFM: "---\ntype: references\nname: transformer-design\ndescription: transformer 协议层架构说明\napplyTo:\n  - \"internal/transformer/**\"\n---",
	},
	{
		Type:      "subagents",
		Semantics: "隔离子代理定义。AI 通过 spawn 子进程或 CLI 调用执行（如 claude --agent <name>）。有独立 system prompt 与上下文。",
		WhenToUse: "需要在干净上下文里完成的任务（代码审查、长 research、并行 dispatch）。每个 subagent 一个独立人格 + 工具白名单。",
		WhenNot:   "当前 session 内的工作流 -> 用 skills。简单模板 -> 用 commands。",
		ExampleFM: "---\ntype: subagents\nname: code-reviewer\ndescription: 代码审查子代理。读 diff 输出结构化 review 报告。\n---",
	},
}

func newExplainCmd() *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "explain [type]",
		Short: "解释 std-ai 5 种类型（rules/skills/commands/references/subagents）的语义",
		Long: `输出 std-ai 5 种 type 的语义速查：每种类型的触发语义 / 何时使用 / 何时不用 / 示例 frontmatter。

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
