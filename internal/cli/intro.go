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
		Short: "Print the AI assistant prompt explaining how to author, migrate, and sync std-agent content",
		Long: `Print an authoritative prompt for AI assistants (Claude / GPT / Gemini, etc.).

Paste the output at the start of an AI conversation and the AI will understand how to:
- Author rules / skills / commands / references under .stdai/standards/
- Migrate from existing CLAUDE.md / .cursor/rules/ / .clinerules/ etc. to std-agent
- Run stdagent commands to finish syncing and maintenance

Suitable for a team AI assistant system prompt, or pipe it directly to an LLM CLI tool.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if asJSON {
				return writeIntroJSON(cmd, copyOnly)
			}
			cmd.Print(introPrompt)
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "JSON output (with version + prompt fields)")
	cmd.Flags().BoolVar(&copyOnly, "copy", false, "In JSON mode, print only the prompt field value (unquoted raw string)")
	return cmd
}

type introPayload struct {
	Version string `json:"version"`
	Prompt  string `json:"prompt"`
}

func writeIntroJSON(cmd *cobra.Command, copyOnly bool) error {
	if copyOnly {
		// Raw output (equivalent to the default text, but keeps the --copy
		// semantics visible so the user's intent is explicit)
		cmd.Print(introPrompt)
		return nil
	}
	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	return enc.Encode(introPayload{Version: versionStr, Prompt: introPrompt})
}

// IntroPrompt returns the embedded AI assistant prompt (for other tools or tests)
func IntroPrompt() string {
	return introPrompt
}
