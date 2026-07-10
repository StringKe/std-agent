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
		Short: "列出与给定文件路径匹配的 rules / references / subagents（按需加载上下文）",
		Long: `根据传入的相对路径（相对项目根），扫 .stdai/standards/ 下所有 docs 的 frontmatter applyTo glob，
返回应当为该文件加载的 docs（rules / references / subagents 等）。

设计目的：给 AI 助手（或人）一个 single-shot 查询，让 AI 在编辑某文件前知道
该读哪些规则，避免预加载全部上下文。

示例：

    stdagent which internal/runner/runner.go
    stdagent which internal/runner/runner.go --json
    stdagent which internal/runner/runner.go --paths
    stdagent which internal/runner/runner.go --type=rules,references
    stdagent which internal/runner/runner.go --include-global   # 也列出无 applyTo 的全局 docs
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
	f.BoolVar(&asJSON, "json", false, "JSON 输出（给 AI / 自动化集成）")
	f.BoolVar(&pathsOnly, "paths", false, "只输出源文件路径，方便 pipe")
	f.BoolVar(&includeGlobal, "include-global", false, "包含无 applyTo 的全局 docs（默认只列匹配的）")
	f.StringVar(&typesCSV, "type", "", "按 type 过滤，逗号分隔（rules,skills,commands,subagents,references）")
	return cmd
}

type whichMatch struct {
	doc          *parser.Document
	matchedGlobs []string
	global       bool // true 表示无 applyTo / 仅靠 includeGlobal 命中
}

// normalizeQueryPath 把用户输入的路径归一化为相对项目根的 forward-slash 路径
//
// 接受绝对路径、./xxx、含 ../ 的相对路径，统一转成 root 相对的 slash 路径。
// 落到 root 外部的路径保持原样（让 glob 匹配自然失败）。
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

// loadAllDocs 扫 .stdai/standards/ 下所有 .md，parse 后返回 docs（不跑 transformer）。
//
// 复用 source.NewLocal + parser.Parse；与 runner.Sync 路径一致但跳过 git 源 / transformer / writer。
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

// isSkillSubdir 与 runner 内部判定一致：skills/<n>/<subdir>/x.md 是 SKILL 辅助文件
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

// 占位避免某些 build 下 os 未用：当前实现路径都用到了
var _ = os.Stat
