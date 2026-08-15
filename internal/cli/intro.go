package cli

import (
	_ "embed"
	"encoding/json"

	"github.com/spf13/cobra"
)

//go:embed intro_prompt.md
var introPrompt string

func newIntroCmd() *cobra.Command {
	var asJSON bool
	var copyOnly bool
	cmd := &cobra.Command{
		Use:   "intro",
		Short: "输出 AI 助手提示词，告诉 AI 如何编写、迁移、同步 std-agent 内容",
		Long: `输出一段权威提示词给 AI 助手（Claude / GPT / Gemini 等）。

把输出粘到 AI 对话开头，AI 就能理解如何：
- 编写 .stdai/standards/ 下的 rules / skills / commands / references
- 从已有 CLAUDE.md / .cursor/rules/ / .clinerules/ 等迁移到 std-agent
- 跑 stdagent 命令完成同步与维护

适合放在团队 AI 助手 system prompt 里，或者直接 pipe 给 LLM CLI 工具。`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if asJSON {
				return writeIntroJSON(cmd, copyOnly)
			}
			cmd.Print(introPrompt)
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "JSON 输出（含 version + prompt 字段）")
	cmd.Flags().BoolVar(&copyOnly, "copy", false, "JSON 模式下仅输出 prompt 字段值（无引号 json 字符串）")
	return cmd
}

type introPayload struct {
	Version string `json:"version"`
	Prompt  string `json:"prompt"`
}

func writeIntroJSON(cmd *cobra.Command, copyOnly bool) error {
	if copyOnly {
		// raw 输出（与默认 text 等价，但保留 --copy 语义占位让用户清晰意图）
		cmd.Print(introPrompt)
		return nil
	}
	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	return enc.Encode(introPayload{Version: versionStr, Prompt: introPrompt})
}

// IntroPrompt 返回内嵌的 AI 助手提示词内容（供其他工具或测试使用）
func IntroPrompt() string {
	return introPrompt
}
