package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/spf13/cobra"

	"github.com/StringKe/std-agent/internal/parser"
	"github.com/StringKe/std-agent/internal/source"
)

func newWhichCmd() *cobra.Command {
	var asJSON, pathsOnly, includeGlobal bool
	var typesCSV string
	cmd := &cobra.Command{
		Use:   "which <file-path>",
		Short: "List rules / references / subagents matching a file path (load context on demand)",
		Long: `Given a project-root-relative path, scan the frontmatter applyTo globs of all docs under .stdai/standards/
and return the docs (rules / references / subagents, etc.) that should be loaded for that file.

Purpose: give an AI assistant (or a human) a single-shot query so the AI knows
which rules to read before editing a file, without preloading all context.

Examples:

    stdagent which internal/runner/runner.go
    stdagent which internal/runner/runner.go --json
    stdagent which internal/runner/runner.go --paths
    stdagent which internal/runner/runner.go --type=rules,references
    stdagent which internal/runner/runner.go --include-global   # also list global docs without applyTo
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			_, root := resolveConfigPath()
			target := normalizeQueryPath(root, args[0])

			docs, err := loadAllDocs(root)
			if err != nil {
				return err
			}

			wantTypes := parseTypeFilter(typesCSV)
			matches := matchDocs(docs, target, wantTypes, includeGlobal)

			sort.SliceStable(matches, func(i, j int) bool {
				ri := parser.PriorityRank(matches[i].doc.Priority)
				rj := parser.PriorityRank(matches[j].doc.Priority)
				if ri != rj {
					return ri < rj
				}
				if string(matches[i].doc.Type) != string(matches[j].doc.Type) {
					return string(matches[i].doc.Type) < string(matches[j].doc.Type)
				}
				return matches[i].doc.Name < matches[j].doc.Name
			})

			switch {
			case asJSON:
				return writeWhichJSON(cmd, target, matches)
			case pathsOnly:
				for _, m := range matches {
					cmd.Println(filepath.Join(".stdai/standards", m.doc.Path))
				}
				return nil
			default:
				return writeWhichTable(cmd, target, matches)
			}
		},
	}
	f := cmd.Flags()
	f.BoolVar(&asJSON, "json", false, "JSON output (for AI / automation integration)")
	f.BoolVar(&pathsOnly, "paths", false, "Print source file paths only, convenient for pipes")
	f.BoolVar(&includeGlobal, "include-global", false, "Include global docs without applyTo (matches only by default)")
	f.StringVar(&typesCSV, "type", "", "Filter by type, comma-separated (rules,skills,commands,subagents,references)")
	return cmd
}

type whichMatch struct {
	doc          *parser.Document
	matchedGlobs []string
	global       bool // true means no applyTo / matched via includeGlobal only
}

// normalizeQueryPath normalizes user input to a root-relative forward-slash path.
//
// Accepts absolute paths, ./xxx, and relative paths with ../, converting them
// all to root-relative slash paths. Paths landing outside the root are kept
// as-is (so glob matching naturally fails).
func normalizeQueryPath(root, in string) string {
	abs := in
	if !filepath.IsAbs(in) {
		abs = filepath.Join(root, in)
	}
	clean := filepath.Clean(abs)
	rel, err := filepath.Rel(root, clean)
	if err != nil || strings.HasPrefix(rel, "..") {
		return filepath.ToSlash(filepath.Clean(in))
	}
	return filepath.ToSlash(rel)
}

// loadAllDocs scans all .md files under .stdai/standards/ and returns the
// parsed docs (without running transformers).
//
// Reuses source.NewLocal + parser.Parse; same path as runner.Sync but skips
// git sources / transformers / writers.
func loadAllDocs(root string) ([]*parser.Document, error) {
	localRoot := filepath.Join(root, ".stdai/standards")
	files, err := source.NewLocal(localRoot).Files()
	if err != nil {
		return nil, fmt.Errorf("read .stdai/standards: %w", err)
	}
	docs := make([]*parser.Document, 0, len(files))
	for _, f := range files {
		lower := strings.ToLower(f.Path)
		if !strings.HasSuffix(lower, ".md") && !strings.HasSuffix(lower, ".markdown") {
			continue
		}
		if isSkillSubdir(f.Path) {
			continue
		}
		d, perr := parser.Parse(f.Path, f.Raw)
		if perr != nil {
			continue
		}
		docs = append(docs, d)
	}
	return docs, nil
}

// isSkillSubdir matches the runner's internal rule: skills/<n>/<subdir>/x.md is a SKILL support file
func isSkillSubdir(p string) bool {
	if !strings.HasPrefix(p, "skills/") {
		return false
	}
	return strings.Count(p, "/") >= 3
}

func parseTypeFilter(csv string) map[parser.DocType]struct{} {
	if csv == "" {
		return nil
	}
	out := map[parser.DocType]struct{}{}
	for _, s := range strings.Split(csv, ",") {
		t := strings.TrimSpace(s)
		if t == "" {
			continue
		}
		out[parser.DocType(t)] = struct{}{}
	}
	return out
}

func matchDocs(docs []*parser.Document, target string, wantTypes map[parser.DocType]struct{}, includeGlobal bool) []whichMatch {
	var out []whichMatch
	for _, d := range docs {
		if wantTypes != nil {
			if _, ok := wantTypes[d.Type]; !ok {
				continue
			}
		}
		if d.Root || d.NestedPath != "" {
			continue
		}
		if len(d.ApplyTo) == 0 {
			if includeGlobal {
				out = append(out, whichMatch{doc: d, global: true})
			}
			continue
		}
		var hits []string
		for _, g := range d.ApplyTo {
			ok, _ := doublestar.Match(g, target)
			if ok {
				hits = append(hits, g)
			}
		}
		if len(hits) > 0 {
			out = append(out, whichMatch{doc: d, matchedGlobs: hits})
		}
	}
	return out
}

func writeWhichTable(cmd *cobra.Command, target string, matches []whichMatch) error {
	if len(matches) == 0 {
		cmd.Printf("# %s\n(no matching docs; try --include-global to see globally-applied docs)\n", target)
		return nil
	}
	cmd.Printf("# %s -> %d match(es)\n", target, len(matches))
	const formatStr = "%-12s %-32s %-8s %-50s %s\n"
	cmd.Printf(formatStr, "TYPE", "NAME", "PRIORITY", "SOURCE", "MATCHED")
	for _, m := range matches {
		src := filepath.Join(".stdai/standards", m.doc.Path)
		matchInfo := strings.Join(m.matchedGlobs, ",")
		if m.global {
			matchInfo = "(global, no applyTo)"
		}
		prio := string(m.doc.Priority)
		if prio == "" {
			prio = "normal"
		}
		cmd.Printf(
			formatStr,
			string(m.doc.Type),
			m.doc.Name,
			prio,
			src,
			matchInfo,
		)
		if m.doc.Description != "" {
			cmd.Printf("             %s\n", m.doc.Description)
		}
	}
	return nil
}

type whichPayload struct {
	Target  string         `json:"target"`
	Matches []whichJSONHit `json:"matches"`
}

type whichJSONHit struct {
	Type         string   `json:"type"`
	Name         string   `json:"name"`
	Priority     string   `json:"priority"`
	Source       string   `json:"source"`
	Description  string   `json:"description,omitempty"`
	MatchedGlobs []string `json:"matched_globs,omitempty"`
	Global       bool     `json:"global,omitempty"`
}

func writeWhichJSON(cmd *cobra.Command, target string, matches []whichMatch) error {
	out := whichPayload{Target: target, Matches: make([]whichJSONHit, 0, len(matches))}
	for _, m := range matches {
		prio := string(m.doc.Priority)
		if prio == "" {
			prio = "normal"
		}
		out.Matches = append(out.Matches, whichJSONHit{
			Type:         string(m.doc.Type),
			Name:         m.doc.Name,
			Priority:     prio,
			Source:       filepath.Join(".stdai/standards", m.doc.Path),
			Description:  m.doc.Description,
			MatchedGlobs: m.matchedGlobs,
			Global:       m.global,
		})
	}
	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

// Placeholder so os stays used under some builds: all current paths use it
var _ = os.Stat
