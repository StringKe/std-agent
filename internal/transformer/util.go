package transformer

import (
	"bytes"
	"fmt"
	"path"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"std-ai/internal/config"
	"std-ai/internal/parser"
	"std-ai/internal/writer"
)

// transformerVersion 由 cli 通过 SetVersion 注入；fallback "dev"
var transformerVersion = "dev"

// SetVersion 注入版本字符串供 footer marker 使用
func SetVersion(v string) {
	if v != "" {
		transformerVersion = v
	}
}

// FilterDocs 按 target 名筛选适用的 docs，返回新 slice
func FilterDocs(docs []*parser.Document, target string) []*parser.Document {
	out := make([]*parser.Document, 0, len(docs))
	for _, d := range docs {
		if targetApplies(d, target) {
			out = append(out, d)
		}
	}
	return out
}

func targetApplies(d *parser.Document, target string) bool {
	if len(d.ExcludeTargets) > 0 {
		for _, t := range d.ExcludeTargets {
			if t == target {
				return false
			}
		}
		return true
	}
	if len(d.Targets) == 0 {
		return true
	}
	for _, t := range d.Targets {
		if t == target {
			return true
		}
	}
	return false
}

// FilterByType 按 type 过滤
func FilterByType(docs []*parser.Document, t parser.DocType) []*parser.Document {
	out := make([]*parser.Document, 0, len(docs))
	for _, d := range docs {
		if d.Type == t {
			out = append(out, d)
		}
	}
	return out
}

// SortDocs 按 priority -> name 稳定排序
func SortDocs(docs []*parser.Document) {
	sort.SliceStable(docs, func(i, j int) bool {
		ri, rj := parser.PriorityRank(docs[i].Priority), parser.PriorityRank(docs[j].Priority)
		if ri != rj {
			return ri < rj
		}
		return docs[i].Name < docs[j].Name
	})
}

// MakeOpts 构造 writer.FooterOptions
func MakeOpts(cfg *config.Config, target, source string, withWhatIs bool) writer.FooterOptions {
	return writer.FooterOptions{
		Inject:       cfg.Inject,
		InjectWhatIs: cfg.InjectWhatIs && withWhatIs,
		Version:      transformerVersion,
		SourcePath:   source,
		TargetName:   target,
		GeneratedAt:  time.Now(),
	}
}

// BuildMarkdownFile 把 frontmatter + body + footer 拼成完整文件
func BuildMarkdownFile(path, frontmatter, body string, opts writer.FooterOptions) writer.FileOp {
	var c bytes.Buffer
	c.WriteString(writer.HeaderComment(opts))
	if frontmatter != "" {
		c.WriteString(frontmatter)
		if !strings.HasSuffix(frontmatter, "\n") {
			c.WriteString("\n")
		}
	}
	body = strings.TrimSpace(body)
	if body != "" {
		c.WriteString(body)
		c.WriteString("\n")
	}
	c.WriteString(writer.FooterMarker(opts))
	return writer.FileOp{Path: path, Content: c.Bytes(), Marker: opts.Inject}
}

// FmBuilder 是简易 YAML frontmatter 构造器
type FmBuilder struct {
	b strings.Builder
	n int
}

// Add 添加 key: val（val 为空则跳过）
func (f *FmBuilder) Add(key, val string) {
	if val == "" {
		return
	}
	fmt.Fprintf(&f.b, "%s: %s\n", key, YAMLScalar(val))
	f.n++
}

// AddList 添加 key: [items]
func (f *FmBuilder) AddList(key string, vals []string) {
	if len(vals) == 0 {
		return
	}
	fmt.Fprintf(&f.b, "%s:\n", key)
	for _, v := range vals {
		fmt.Fprintf(&f.b, "  - %s\n", YAMLScalar(v))
	}
	f.n++
}

// AddBool 强制写入 key: true/false
func (f *FmBuilder) AddBool(key string, v bool) {
	fmt.Fprintf(&f.b, "%s: %t\n", key, v)
	f.n++
}

// AddRaw 写入 key: rawValue（不 escape，调用者保证安全）
func (f *FmBuilder) AddRaw(key, val string) {
	if val == "" {
		return
	}
	fmt.Fprintf(&f.b, "%s: %s\n", key, val)
	f.n++
}

// AddMap 把 map 序列化为 YAML 嵌套块，附到 frontmatter
//
// 用于 metadata / hooks 等自由形式字段。空 map 跳过。
// yaml.Marshal 保证缩进与类型正确（string / bool / int / list / nested map）。
func (f *FmBuilder) AddMap(key string, m map[string]interface{}) {
	if len(m) == 0 {
		return
	}
	data, err := yaml.Marshal(map[string]interface{}{key: m})
	if err != nil {
		return
	}
	f.b.Write(data)
	if !bytes.HasSuffix(data, []byte("\n")) {
		f.b.WriteString("\n")
	}
	f.n++
}

// String 返回 ---\nbody\n--- 包裹形式；空 frontmatter 返回 ""
func (f *FmBuilder) String() string {
	if f.n == 0 {
		return ""
	}
	return "---\n" + f.b.String() + "---\n"
}

// YAMLScalar 简单 YAML 字符串转义；含特殊字符则双引号
func YAMLScalar(s string) string {
	if strings.ContainsAny(s, ":#\n\"\\") || strings.HasPrefix(s, "-") || strings.HasPrefix(s, " ") {
		s = strings.ReplaceAll(s, `\`, `\\`)
		s = strings.ReplaceAll(s, `"`, `\"`)
		s = strings.ReplaceAll(s, "\n", `\n`)
		return `"` + s + `"`
	}
	return s
}

// JoinAGENTSStyle 把 docs 拼成 # title + ## name + body 的 markdown
func JoinAGENTSStyle(title string, docs []*parser.Document) string {
	var b strings.Builder
	if title != "" {
		fmt.Fprintf(&b, "# %s\n\n", title)
	}
	for i, d := range docs {
		if i > 0 {
			b.WriteString("\n")
		}
		fmt.Fprintf(&b, "## %s\n\n", d.Name)
		body := strings.TrimSpace(d.Body)
		if d.Description != "" {
			fmt.Fprintf(&b, "%s\n\n", d.Description)
		}
		if body != "" {
			b.WriteString(body)
			b.WriteString("\n")
		}
	}
	return b.String()
}

// CommaJoin 把 string slice 用逗号拼接（用于 Copilot applyTo）
func CommaJoin(items []string) string {
	return strings.Join(items, ",")
}

// MergeDescription 把 std description 与 Claude Code 风格的 when_to_use 合并成单串
// 用于不支持 when_to_use 字段的 target，把触发线索拼入 description 末尾
func MergeDescription(desc, whenToUse string) string {
	if whenToUse == "" {
		return desc
	}
	if desc == "" {
		return whenToUse
	}
	return desc + " " + whenToUse
}

// BuildSkillPackage 把 SKILL.md FileOp + 同目录辅助文件转为多 FileOp
//
// skillDir 是相对项目根的 skill 目录（如 ".claude/skills/code-review"）。
// skillMd 是 BuildMarkdownFile 已构造好的 SKILL.md FileOp。
// files 是 parser 收集到的 SKILL package 辅助文件，path 相对 skill 目录。
//
// 用 path.Join 统一拼接（logical forward-slash path），跨 OS 一致；
// 写盘时由 writer 通过 filepath.Join(projectRoot, FileOp.Path) 转 OS 形式。
func BuildSkillPackage(skillDir string, skillMd writer.FileOp, files []parser.SkillFile) []writer.FileOp {
	out := make([]writer.FileOp, 0, len(files)+1)
	out = append(out, skillMd)
	for _, f := range files {
		out = append(out, writer.FileOp{
			Path:    path.Join(skillDir, f.Path),
			Content: f.Raw,
		})
	}
	return out
}

// SkillDir 拼 SKILL package 目录（logical path），如 SkillDir(".claude/skills", "review")
// 返回 ".claude/skills/review"
func SkillDir(base, name string) string {
	return path.Join(base, name)
}

// FilePath 拼 logical 文件路径，等价 path.Join + 扩展名
// e.g. FilePath(".claude/rules", "style", ".md") -> ".claude/rules/style.md"
func FilePath(dir, name, ext string) string {
	return path.Join(dir, name+ext)
}
