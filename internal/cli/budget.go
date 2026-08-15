package cli

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/StringKe/std-agent/internal/budget"
	"github.com/StringKe/std-agent/internal/config"
	"github.com/StringKe/std-agent/internal/parser"
	"github.com/StringKe/std-agent/internal/source"
	"github.com/StringKe/std-agent/internal/transformer"
	"github.com/StringKe/std-agent/internal/writer"
)

func newBudgetCmd() *cobra.Command {
	var (
		asJSON   bool
		rendered bool
		targets  []string
	)
	cmd := &cobra.Command{
		Use:   "budget",
		Short: "估算 source 与实际 target 输出的 LLM 上下文消耗",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runBudget(cmd, asJSON, rendered || len(targets) > 0, targets)
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "结构化 JSON 输出")
	cmd.Flags().BoolVar(&rendered, "rendered", false, "显示启用 target 的实际 root 与 sidecar 体积")
	cmd.Flags().StringSliceVar(&targets, "target", nil, "限定 rendered target，可重复")
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
	Docs            []budgetDocReport            `json:"docs"`
	TotalRulesBytes int                          `json:"total_rules_bytes"`
	TotalTokens     int                          `json:"total_estimated_tokens"`
	TotalWarnings   []string                     `json:"total_warnings,omitempty"`
	SourceLayers    []budgetSourceLayerReport    `json:"source_layers,omitempty"`
	RenderedTargets []budgetRenderedTargetReport `json:"rendered_targets,omitempty"`
}

type budgetSourceLayerReport struct {
	Name   string   `json:"name"`
	Types  []string `json:"types"`
	Bytes  int      `json:"bytes"`
	Tokens int      `json:"estimated_tokens"`
}

type budgetRenderedFileReport struct {
	Path     string   `json:"path"`
	Bytes    int      `json:"bytes"`
	Tokens   int      `json:"estimated_tokens"`
	Warnings []string `json:"warnings,omitempty"`
}

type budgetRenderedTargetReport struct {
	Target        string                     `json:"target"`
	RootFiles     []budgetRenderedFileReport `json:"root_files"`
	RootBytes     int                        `json:"root_bytes"`
	RootTokens    int                        `json:"root_estimated_tokens"`
	SidecarFiles  int                        `json:"sidecar_files"`
	SidecarBytes  int                        `json:"sidecar_bytes"`
	SidecarTokens int                        `json:"sidecar_estimated_tokens"`
}

func runBudget(cmd *cobra.Command, asJSON, rendered bool, targets []string) error {
	cfgPath, root := resolveConfigPath()
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
	collectBudgetSkillPackageFiles(docs, files)

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
	r.TotalWarnings = append(budget.CheckTotalRules(docs), budget.CheckTotalSkills(docs)...)
	sort.Slice(r.Docs, func(i, j int) bool { return r.Docs[i].Bytes > r.Docs[j].Bytes })

	if rendered {
		r.SourceLayers = buildSourceLayerReports(docs)
		cfg, cerr := config.Load(cfgPath)
		if cerr != nil {
			return fmt.Errorf("load config: %w", cerr)
		}
		renderedTargets, rerr := buildRenderedTargetReports(docs, cfg, targets)
		if rerr != nil {
			return rerr
		}
		r.RenderedTargets = renderedTargets
	}

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
	if rendered {
		pl("")
		pl("SOURCE LAYERS")
		for _, layer := range r.SourceLayers {
			pf("%-12s %10d bytes %10d tokens   %s\n",
				layer.Name, layer.Bytes, layer.Tokens, strings.Join(layer.Types, ", "))
		}
		pl("")
		pl("RENDERED TARGETS")
		for _, target := range r.RenderedTargets {
			pf("%-16s root %8d bytes %8d tokens   sidecars %3d / %8d bytes / %8d tokens\n",
				target.Target, target.RootBytes, target.RootTokens,
				target.SidecarFiles, target.SidecarBytes, target.SidecarTokens)
			for _, file := range target.RootFiles {
				pf("    %-46s %8d bytes %8d tokens\n", file.Path, file.Bytes, file.Tokens)
				for _, warning := range file.Warnings {
					pf("        %s\n", warning)
				}
			}
		}
	}
	return nil
}

func buildSourceLayerReports(docs []*parser.Document) []budgetSourceLayerReport {
	layers := []budgetSourceLayerReport{
		{Name: "rules", Types: []string{"rules"}},
		{Name: "on-demand", Types: []string{"skills", "commands", "references", "subagents"}},
	}
	for _, doc := range docs {
		i := 1
		if doc.Type == parser.TypeRules {
			i = 0
		}
		layers[i].Bytes += doc.BodyBytes
		layers[i].Tokens += budget.EstimateTokens(doc.Body)
	}
	return layers
}

func buildRenderedTargetReports(
	docs []*parser.Document,
	cfg *config.Config,
	requested []string,
) ([]budgetRenderedTargetReport, error) {
	targets := requested
	if len(targets) == 0 {
		for name, targetCfg := range cfg.Targets {
			if targetCfg.Enabled {
				targets = append(targets, name)
			}
		}
	}
	sort.Strings(targets)
	targets = compactStrings(targets)

	plans := make([]*writer.Plan, 0, len(targets))
	for _, name := range targets {
		tr, ok := transformer.Get(name)
		if !ok {
			return nil, fmt.Errorf("unknown target %q", name)
		}
		plan, err := tr.Plan(docs, cfg)
		if err != nil {
			return nil, fmt.Errorf("transform %s: %w", name, err)
		}
		plans = append(plans, plan)
	}
	if err := transformer.CanonicalizeSharedAGENTS(plans, docs, cfg); err != nil {
		return nil, fmt.Errorf("canonicalize shared AGENTS.md: %w", err)
	}

	reports := make([]budgetRenderedTargetReport, 0, len(plans))
	for _, plan := range plans {
		report := budgetRenderedTargetReport{Target: plan.Target}
		for _, op := range plan.Files {
			if op.Skip {
				continue
			}
			size := len(op.Content)
			tokens := budget.EstimateTokens(string(op.Content))
			if op.IsRoot {
				report.RootFiles = append(report.RootFiles, budgetRenderedFileReport{
					Path:     op.Path,
					Bytes:    size,
					Tokens:   tokens,
					Warnings: budget.CheckRootFile(plan.Target, op.Path, size),
				})
				report.RootBytes += size
				report.RootTokens += tokens
				continue
			}
			report.SidecarFiles++
			report.SidecarBytes += size
			report.SidecarTokens += tokens
		}
		sort.Slice(report.RootFiles, func(i, j int) bool {
			return report.RootFiles[i].Path < report.RootFiles[j].Path
		})
		reports = append(reports, report)
	}
	return reports, nil
}

func compactStrings(items []string) []string {
	if len(items) < 2 {
		return items
	}
	out := items[:1]
	for _, item := range items[1:] {
		if item != out[len(out)-1] {
			out = append(out, item)
		}
	}
	return out
}

func collectBudgetSkillPackageFiles(docs []*parser.Document, files []source.File) {
	skillRoots := map[string]*parser.Document{}
	for _, doc := range docs {
		if doc.Type != parser.TypeSkills || filepath.Base(doc.Path) != "SKILL.md" {
			continue
		}
		skillRoots[strings.TrimSuffix(doc.Path, "SKILL.md")] = doc
	}
	for _, file := range files {
		for root, doc := range skillRoots {
			if !strings.HasPrefix(file.Path, root) {
				continue
			}
			rel := strings.TrimPrefix(file.Path, root)
			if rel == "SKILL.md" {
				continue
			}
			doc.SkillFiles = append(doc.SkillFiles, parser.SkillFile{Path: rel, Raw: file.Raw})
			break
		}
	}
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
