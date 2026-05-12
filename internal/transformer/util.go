package transformer

import (
	"bytes"
	"fmt"
	"path"
	"sort"
	"strings"

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

// YAMLScalar 简单 YAML 字符串转义；含 plain scalar 危险字符则双引号
//
// YAML 1.2 plain scalar 在以下情况必须 quote：
//   - 含 indicators: : # \n " \\
//   - 含 flow / glob 字符: * ? [ ] { } | > , (在 plain 中可能被误解析)
//   - 首字符为 reserved indicator: - ? : , [ ] { } # & * ! | > ' " % @ `
//   - 首字符为空格 / Tab
//
// 全角中文标点不需要 quote。
func YAMLScalar(s string) string {
	if s == "" {
		return `""`
	}
	if needsYAMLQuote(s) {
		s = strings.ReplaceAll(s, `\`, `\\`)
		s = strings.ReplaceAll(s, `"`, `\"`)
		s = strings.ReplaceAll(s, "\n", `\n`)
		return `"` + s + `"`
	}
	return s
}

func needsYAMLQuote(s string) bool {
	// 含明确危险字符
	if strings.ContainsAny(s, ":#\n\"\\*?[]{}|>,&!%`") {
		return true
	}
	// 首字符 reserved indicator 或空白
	first := s[0]
	switch first {
	case '-', '@', '\'', ' ', '\t':
		return true
	}
	// 全为数字 / true/false / null 这类会被解析为非 string 的也应 quote，
	// 但 frontmatter 字段语义都是 string 不影响（YAML loader 会按 schema 解析），
	// 这里只关注 plain scalar 解析层面是否合法。
	return false
}

// rulesyncTargetAliases 把 stdagent target 名映射回 rulesync 风格的字段名，
// 让 frontmatter 里写的 claudecode.paths / cursor.paths 能命中对应 target。
//
//nolint:gochecknoglobals // 只读静态映射
var rulesyncTargetAliases = map[string]string{
	"claude-code":  "claudecode",
	"codex":        "codexcli",
	"cursor":       "cursor",
	"copilot":      "copilot",
	"windsurf":     "windsurf",
	"gemini":       "gemini",
	"aider":        "aider",
	"cline":        "cline",
	"opencode":     "opencode",
	"continue-dev": "continue",
	"antigravity":  "antigravity",
}

// EffectiveApplyTo 返回某 target 实际生效的 applyTo glob：
//   - 如果 doc 含 target-specific paths（如 frontmatter 里写 cursor.paths），优先用之
//   - 否则用全局 ApplyTo
//
// transformer 应当用这个 helper 替代直接读 d.ApplyTo，让 target 专属覆盖生效。
func EffectiveApplyTo(d *parser.Document, target string) []string {
	if d == nil || len(d.TargetPaths) == 0 {
		if d != nil {
			return d.ApplyTo
		}
		return nil
	}
	alias := rulesyncTargetAliases[target]
	if alias != "" {
		if v, ok := d.TargetPaths[alias]; ok && len(v) > 0 {
			return v
		}
	}
	if v, ok := d.TargetPaths[target]; ok && len(v) > 0 {
		return v
	}
	return d.ApplyTo
}

// PartitionNested 把 rules 拆为 (顶级 rules, 嵌套 root docs)。
// 嵌套 root docs（NestedPath 非空）在调用方独立处理，不进顶级 manifest 也不 fan-out 到子目录。
func PartitionNested(rules []*parser.Document) (top, nested []*parser.Document) {
	for _, d := range rules {
		if d.NestedPath != "" {
			nested = append(nested, d)
		} else {
			top = append(top, d)
		}
	}
	return top, nested
}

// PartitionRoot 把 rules 拆为 (root rules, non-root rules)。
// 保持 caller 已 sort 过的顺序，不重排。
//
// root rule 的 body 由 transformer 用作根文件（CLAUDE.md / AGENTS.md 等）的项目总结
// 头部；stdagent 仍然在尾部自动追加 manifest 段（含 nonRoot rule 索引）。
// root rule 本身不会 fan-out 成独立 .claude/rules/<n>.md。
func PartitionRoot(rules []*parser.Document) (root, rest []*parser.Document) {
	for _, d := range rules {
		if d.Root {
			root = append(root, d)
		} else {
			rest = append(rest, d)
		}
	}
	return root, rest
}

// RenderRootBody 把多个 root rule 的 body 拼接成根文件主体内容。
// 各 rule body 之间用空行分隔，不加额外标题（root rule 自己负责标题）。
func RenderRootBody(roots []*parser.Document) string {
	var b strings.Builder
	for i, d := range roots {
		body := strings.TrimSpace(d.Body)
		if body == "" {
			continue
		}
		if i > 0 {
			b.WriteString("\n\n")
		}
		b.WriteString(body)
	}
	if b.Len() > 0 {
		b.WriteString("\n")
	}
	return b.String()
}

// BuildRuleManifestSection 生成"自描述规则清单"段。每条记录 path + description +
// applyTo，让不支持 @import 的 AI 工具也能理解每个文件的语义。
//
// useImport=true 用 Claude Code 的 @<path> import 语法（工具读到 @ 自动加载文件）；
// false 用纯 list 形式（codex / aider 等读 markdown 时按需 Read 文件）。
//
// title 形如 "Imported Rules" / "Reference Rules"；空 docs 返回空串。
//
// stdagent 概念解释（什么是 rules / skills / 等）由 .stdai/help/*.md 提供，
// root.md 通过 @<path> 引用，本 manifest 段不再嵌入概念速读。
func BuildRuleManifestSection(title, target string, docs []*parser.Document, pathFn func(*parser.Document) string, useImport bool) string {
	if len(docs) == 0 {
		return ""
	}
	var b strings.Builder
	fmt.Fprintf(&b, "\n## %s\n\n", title)
	dir := filepathPrefix(pathFn(docs[0]))
	if useImport {
		fmt.Fprintf(&b, "下列条目由 stdagent 同步生成到 `%s/`，AI 工具读到 `@<path>` 自动加载，AI 也可通过 description 提示判断是否相关：\n\n", dir)
		for _, d := range docs {
			line := "@" + pathFn(d)
			if d.Description != "" {
				line += " -- " + d.Description
			}
			if applyTo := EffectiveApplyTo(d, target); len(applyTo) > 0 {
				line += "  (applyTo: `" + strings.Join(applyTo, ", ") + "`)"
			}
			b.WriteString("- " + line + "\n")
		}
	} else {
		fmt.Fprintf(&b, "下列条目由 stdagent 同步到 `%s/`，AI 按 description / applyTo 判断是否相关后用 Read 工具读取具体内容：\n\n", dir)
		for _, d := range docs {
			fmt.Fprintf(&b, "- `%s`", pathFn(d))
			if d.Description != "" {
				fmt.Fprintf(&b, " -- %s", d.Description)
			}
			if applyTo := EffectiveApplyTo(d, target); len(applyTo) > 0 {
				fmt.Fprintf(&b, "  (applyTo: `%s`)", strings.Join(applyTo, ", "))
			}
			b.WriteString("\n")
		}
	}
	return b.String()
}

// filepathPrefix 取 "a/b/c.md" 的目录前缀 "a/b"。空或单段返回 "."
func filepathPrefix(p string) string {
	idx := strings.LastIndex(p, "/")
	if idx <= 0 {
		return "."
	}
	return p[:idx]
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
