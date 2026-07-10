package cli

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/StringKe/std-agent/internal/budget"
	"github.com/StringKe/std-agent/internal/parser"
	"github.com/StringKe/std-agent/internal/source"
)

func newBudgetCmd() *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "budget",
		Short: "估算 std 源文件的 LLM 上下文消耗与限额检查",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runBudget(cmd, asJSON)
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "结构化 JSON 输出")
	return cmd
}

type budgetDocReport struct {
	Path     string   `json:"path"`
	Type     string   `json:"type"`
	Bytes    int      `json:"bytes"`
	Tokens   int      `json:"estimated_tokens"`
	Warnings []string `json:"warnings,omitempty"`
}

type budgetReport struct {
	Docs            []budgetDocReport `json:"docs"`
	TotalRulesBytes int               `json:"total_rules_bytes"`
	TotalTokens     int               `json:"total_estimated_tokens"`
	TotalWarnings   []string          `json:"total_warnings,omitempty"`
}

func runBudget(cmd *cobra.Command, asJSON bool) error {
	_, root := resolveConfigPath()
	standardsRoot := filepath.Join(root, ".stdai/standards")

	files, err := source.NewLocal(standardsRoot).Files()
	if err != nil {
		return fmt.Errorf("read standards: %w", err)
	}
	if len(files) == 0 {
		cmd.PrintErrln("[budget] no standards files found at", standardsRoot)
		return nil
	}

	var docs []*parser.Document
	for _, f := range files {
		if !isMarkdownFile(f.Path) {
			continue
		}
		if isSkillSubdirMarkdownFile(f.Path) {
			continue
		}
		d, perr := parser.Parse(f.Path, f.Raw)
		if perr != nil {
			cmd.PrintErrf("[skip] %s: %v\n", f.Path, perr)
			continue
		}
		docs = append(docs, d)
	}

	r := budgetReport{}
	for _, d := range docs {
		dr := budgetDocReport{
			Path:     d.Path,
			Type:     string(d.Type),
			Bytes:    d.BodyBytes,
			Tokens:   budget.EstimateTokens(d.Body),
			Warnings: budget.CheckDocument(d),
		}
		r.Docs = append(r.Docs, dr)
		r.TotalTokens += dr.Tokens
		if d.Type == parser.TypeRules {
			r.TotalRulesBytes += d.BodyBytes
		}
	}
	r.TotalWarnings = budget.CheckTotalRules(docs)
	sort.Slice(r.Docs, func(i, j int) bool { return r.Docs[i].Bytes > r.Docs[j].Bytes })

	if asJSON {
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(r)
	}

	out := cmd.OutOrStdout()
	pf := func(format string, a ...any) { _, _ = fmt.Fprintf(out, format, a...) }
	pl := func(s string) { _, _ = fmt.Fprintln(out, s) }

	pf("总文档数 %d   总 rules 字节 %d   估算总 tokens %d\n\n",
		len(r.Docs), r.TotalRulesBytes, r.TotalTokens)
	pf("%-50s %-12s %10s %10s\n", "PATH", "TYPE", "BYTES", "~TOKENS")
	pl(strings.Repeat("-", 86))
	for _, dr := range r.Docs {
		pf("%-50s %-12s %10d %10d\n", dr.Path, dr.Type, dr.Bytes, dr.Tokens)
		for _, w := range dr.Warnings {
			pf("    %s\n", w)
		}
	}
	if len(r.TotalWarnings) > 0 {
		pl("")
		for _, w := range r.TotalWarnings {
			pf("[total] %s\n", w)
		}
	}
	return nil
}

// isMarkdownFile 判断 path 后缀是否为 markdown
func isMarkdownFile(p string) bool {
	lower := strings.ToLower(p)
	return strings.HasSuffix(lower, ".md") || strings.HasSuffix(lower, ".markdown")
}

// isSkillSubdirMarkdownFile 判断是否 SKILL package 子目录的辅助 markdown
// （与 runner.isSkillSubdirMarkdown 等价；cli 包独立维护避免跨包导出私有 helper）
func isSkillSubdirMarkdownFile(p string) bool {
	return strings.HasPrefix(p, "skills/") && strings.Count(p, "/") >= 3
}
