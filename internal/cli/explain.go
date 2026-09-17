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

// explainType is a single record of `stdagent explain --json` output
type explainType struct {
	Type      string `json:"type"`
	Semantics string `json:"semantics"`
	WhenToUse string `json:"when_to_use"`
	WhenNot   string `json:"when_not"`
	ExampleFM string `json:"example_frontmatter"`
}

// explainTypes holds the structured semantics of the 5 types for --json output.
// It mirrors explain_text.md; keep them in sync.
var explainTypes = []explainType{
	{
		Type:      "rules",
		Semantics: "Always-on coding, architecture, and operational constraints; auto-loaded by targets or matched by path.",
		WhenToUse: "Stable constraints whose violation causes real risk and that deserve resident context.",
		WhenNot:   "Long background goes in references, on-demand workflows in skills, user templates in commands.",
		ExampleFM: "---\ntype: rules\nname: exception-handling\ndescription: Go error propagation and boundary conversion\npriority: high\napplyTo:\n  - \"**/*.go\"\n---",
	},
	{
		Type:      "skills",
		Semantics: "Capability packages the AI invokes on demand based on description; may carry supporting resources.",
		WhenToUse: "Reusable workflows with a clear outcome and success criteria.",
		WhenNot:   "Ongoing constraints go in rules, explicit user templates in commands.",
		ExampleFM: "---\ntype: skills\nname: code-review\ndescription: Review current changes and report correctness, security, and regression issues\n---",
	},
	{
		Type:      "commands",
		Semantics: "Operation templates the user triggers explicitly via /command-name.",
		WhenToUse: "Fixed operations the user wants to invoke directly.",
		WhenNot:   "AI-judged flows go in skills, ongoing constraints in rules.",
		ExampleFM: "---\ntype: commands\nname: review\ndescription: Review current branch changes and produce a review report\n---",
	},
	{
		Type:      "references",
		Semantics: "Architecture, protocol, API, and long-form background consulted only when needed.",
		WhenToUse: "Domain knowledge worth keeping but not worth occupying default context.",
		WhenNot:   "Ongoing constraints go in rules, executable workflows in skills.",
		ExampleFM: "---\ntype: references\nname: transformer-design\ndescription: Transformer protocol-layer architecture notes\napplyTo:\n  - \"internal/transformer/**\"\n---",
	},
	{
		Type:      "subagents",
		Semantics: "Agent definitions executed in an isolated context.",
		WhenToUse: "Tasks that run independently, need a dedicated context, or parallelize safely.",
		WhenNot:   "In-session flows go in skills, simple templates in commands.",
		ExampleFM: "---\ntype: subagents\nname: code-reviewer\ndescription: Review code in an isolated context and return an issue list\n---",
	},
}

func newExplainCmd() *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "explain [type]",
		Short: "Explain the semantics of the 5 std-agent types (rules/skills/commands/references/subagents)",
		Long: `Print a semantics cheat sheet for the 5 std-agent types: trigger semantics / when to use / when not to / example frontmatter for each type.

Without arguments, print all 5 types. With a type argument, print only that type's section.

Examples:

    stdagent explain                  # all 5 types
    stdagent explain rules            # rules only
    stdagent explain --json           # JSON output (AI integration)
    stdagent explain rules --json     # single rules item as JSON
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
	cmd.Flags().BoolVar(&asJSON, "json", false, "JSON output (for AI / automation integration)")
	return cmd
}

// writeExplainMarkdown prints the markdown cheat sheet. An empty filter prints
// everything, otherwise only the matching type's section.
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

// writeExplainJSON prints []explainType (all) or [explainType] (single-type filter).
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

// extractSection extracts the `## <type>` section from explain_text.md
// (including the header, up to the next `##` header).
//
// explain_text.md is divided by `## rules` / `## skills` ... level-2 headers
// plus a trailing `## Quick reference` summary section. Single-type queries
// return only that type's section, without the summary (to avoid duplication).
func extractSection(text, typeName string) string {
	header := "## " + typeName
	idx := strings.Index(text, header+"\n")
	if idx < 0 {
		return ""
	}
	rest := text[idx:]
	// Find the next `## ` level-2 header (must be at line start)
	next := strings.Index(rest[len(header):], "\n## ")
	if next < 0 {
		return rest
	}
	return rest[:len(header)+next+1] // +1 keeps the newline before next
}
