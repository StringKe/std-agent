package parser

import (
	"bytes"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// rawFrontmatter 是直接从 yaml 解析出的原始结构
type rawFrontmatter struct {
	Type           string   `yaml:"type"`
	Name           string   `yaml:"name"`
	Version        string   `yaml:"version"`
	Description    string   `yaml:"description"`
	Targets        []string `yaml:"targets"`
	ExcludeTargets []string `yaml:"exclude_targets"`
	Priority       string   `yaml:"priority"`
	Tags           []string `yaml:"tags"`
	ApplyTo        []string `yaml:"applyTo"`
	// Globs 是 rulesync / Cursor / Cline 业界标准字段名，作为 ApplyTo 别名。
	// 解析时合并到 ApplyTo（applyTo 优先 + globs 兜底，去重）。
	Globs                  []string `yaml:"globs"`
	AlwaysApply            bool     `yaml:"alwaysApply"`
	AllowedTools           []string `yaml:"allowed_tools"`
	ArgumentHint           string   `yaml:"argument_hint"`
	Model                  string   `yaml:"model"`
	DisableModelInvocation bool     `yaml:"disable_model_invocation"`
	UserInvocable          *bool    `yaml:"user_invocable"`
	DisallowedTools        []string `yaml:"disallowed_tools"`
	ReadOnly               bool     `yaml:"readonly"`
	Background             bool     `yaml:"background"`
	Isolation              string   `yaml:"isolation"`
	Memory                 string   `yaml:"memory"`
	PermissionMode         string   `yaml:"permission_mode"`
	MaxTurns               int      `yaml:"max_turns"`
	PreloadSkills          []string `yaml:"preload_skills"`

	// SKILL package 扩展字段
	WhenToUse     string                 `yaml:"when_to_use"`
	Arguments     []string               `yaml:"arguments"`
	Effort        string                 `yaml:"effort"`
	SkillContext  string                 `yaml:"context"`
	Agent         string                 `yaml:"agent"`
	Shell         string                 `yaml:"shell"`
	Hooks         map[string]interface{} `yaml:"hooks"`
	License       string                 `yaml:"license"`
	Compatibility string                 `yaml:"compatibility"`
	Metadata      map[string]interface{} `yaml:"metadata"`
}

var nameRe = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)

// splitFrontmatter 拆分 std 文件的 frontmatter 与正文
// 返回 (frontmatter raw bytes, body bytes, hasFrontmatter)
func splitFrontmatter(raw []byte) ([]byte, []byte, bool) {
	raw = bytes.TrimPrefix(raw, []byte{0xEF, 0xBB, 0xBF})

	var rest []byte
	switch {
	case bytes.HasPrefix(raw, []byte("---\r\n")):
		rest = raw[5:]
	case bytes.HasPrefix(raw, []byte("---\n")):
		rest = raw[4:]
	default:
		return nil, raw, false
	}

	idx := -1
	for i := 0; i < len(rest); i++ {
		if rest[i] != '-' {
			continue
		}
		if i > 0 && rest[i-1] != '\n' {
			continue
		}
		if i+2 >= len(rest) {
			break
		}
		if rest[i+1] != '-' || rest[i+2] != '-' {
			continue
		}
		after := rest[i+3:]
		switch {
		case len(after) == 0:
			idx = i
		case bytes.HasPrefix(after, []byte("\n")):
			idx = i
		case bytes.HasPrefix(after, []byte("\r\n")):
			idx = i
		}
		if idx >= 0 {
			break
		}
	}
	if idx < 0 {
		return nil, raw, false
	}

	front := rest[:idx]
	bodyStart := idx + 3
	if bodyStart < len(rest) && rest[bodyStart] == '\r' {
		bodyStart++
	}
	if bodyStart < len(rest) && rest[bodyStart] == '\n' {
		bodyStart++
	}
	return front, rest[bodyStart:], true
}

// mergeGlobs 把 applyTo 与 globs 合并去重，applyTo 优先（保持原顺序），
// 同名 glob 不重复加入。让 rulesync / Cursor / Cline 用户用 globs 字段也能生效。
func mergeGlobs(applyTo, globs []string) []string {
	if len(globs) == 0 {
		return applyTo
	}
	seen := make(map[string]struct{}, len(applyTo)+len(globs))
	out := make([]string, 0, len(applyTo)+len(globs))
	for _, g := range applyTo {
		if _, ok := seen[g]; ok {
			continue
		}
		seen[g] = struct{}{}
		out = append(out, g)
	}
	for _, g := range globs {
		if _, ok := seen[g]; ok {
			continue
		}
		seen[g] = struct{}{}
		out = append(out, g)
	}
	return out
}

// extractTargetPaths 扫顶层 frontmatter map，捕获形如
//
//	claudecode:
//	  paths: [...]
//	cursor:
//	  paths: [...]
//
// 的 nested map，返回 map[<rulesync-target-name>]=paths。
//
// 已知的 rulesync target 名（claudecode / codexcli / cursor / copilot / windsurf /
// gemini / aider / cline / opencode）会在 transformer 用 mapping 转 stdagent target
// 名（claude-code / codex / 等），其他未知 key 忽略。
func extractTargetPaths(m map[string]interface{}) map[string][]string {
	if len(m) == 0 {
		return nil
	}
	known := map[string]bool{
		"claudecode": true, "codexcli": true, "cursor": true, "copilot": true,
		"windsurf": true, "gemini": true, "aider": true, "cline": true,
		"opencode": true, "agentsmd": true,
	}
	out := map[string][]string{}
	for k, v := range m {
		if !known[k] {
			continue
		}
		nested, ok := v.(map[string]interface{})
		if !ok {
			continue
		}
		paths, ok := nested["paths"].([]interface{})
		if !ok {
			continue
		}
		strs := make([]string, 0, len(paths))
		for _, p := range paths {
			if s, ok := p.(string); ok && s != "" {
				strs = append(strs, s)
			}
		}
		if len(strs) > 0 {
			out[k] = strs
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// Parse 解析单个 std 源文件，path 是相对源根的路径
func Parse(path string, raw []byte) (*Document, error) {
	front, body, hasFront := splitFrontmatter(raw)
	doc := &Document{
		Path:      path,
		Body:      string(body),
		BodyBytes: len(body),
	}

	if !hasFront {
		doc.Type = TypeRules
		doc.Name = filenameToName(path)
		if isRootPath(path) {
			doc.Root = true
		}
		if np, ok := nestedRootPath(path); ok {
			doc.Root = true
			doc.NestedPath = np
		}
		return doc, nil
	}

	var fm rawFrontmatter
	if err := yaml.Unmarshal(front, &fm); err != nil {
		return nil, fmt.Errorf("%s: invalid YAML frontmatter: %w", path, err)
	}

	// 二次解析为通用 map 抽 target-specific 字段（如 claudecode.paths / cursor.paths）
	var loose map[string]interface{}
	_ = yaml.Unmarshal(front, &loose)
	doc.TargetPaths = extractTargetPaths(loose)

	switch {
	case fm.Type == "":
		doc.Type = TypeRules
	case !IsValidType(fm.Type):
		return nil, fmt.Errorf("%s: invalid type %q (expected rules/skills/commands/references)", path, fm.Type)
	default:
		doc.Type = DocType(fm.Type)
	}

	if fm.Name == "" {
		doc.Name = filenameToName(path)
	} else {
		if !isValidName(fm.Name) {
			return nil, fmt.Errorf("%s: invalid name %q (expected kebab-case ^[a-z0-9][a-z0-9-]*$)", path, fm.Name)
		}
		doc.Name = fm.Name
	}

	if len(fm.Targets) > 0 && len(fm.ExcludeTargets) > 0 {
		return nil, fmt.Errorf("%s: targets and exclude_targets are mutually exclusive", path)
	}

	if !IsValidPriority(fm.Priority) {
		return nil, fmt.Errorf("%s: invalid priority %q", path, fm.Priority)
	}

	doc.Version = fm.Version
	doc.Description = fm.Description
	doc.Targets = fm.Targets
	doc.ExcludeTargets = fm.ExcludeTargets
	doc.Priority = Priority(fm.Priority)
	doc.Tags = fm.Tags
	doc.ApplyTo = mergeGlobs(fm.ApplyTo, fm.Globs)
	doc.AlwaysApply = fm.AlwaysApply
	// Root 由路径决定：standards 顶层 root.md 或 nested/<sub>/root.md 自动识别。
	if isRootPath(path) {
		doc.Root = true
	}
	if np, ok := nestedRootPath(path); ok {
		doc.Root = true
		doc.NestedPath = np
	}
	doc.AllowedTools = fm.AllowedTools
	doc.ArgumentHint = fm.ArgumentHint
	doc.Model = fm.Model
	doc.DisableModelInvocation = fm.DisableModelInvocation
	doc.UserInvocable = fm.UserInvocable
	doc.DisallowedTools = fm.DisallowedTools
	doc.ReadOnly = fm.ReadOnly
	doc.Background = fm.Background
	doc.Isolation = fm.Isolation
	doc.Memory = fm.Memory
	doc.PermissionMode = fm.PermissionMode
	doc.MaxTurns = fm.MaxTurns
	doc.PreloadSkills = fm.PreloadSkills

	// SKILL package 扩展字段
	doc.WhenToUse = fm.WhenToUse
	doc.Arguments = fm.Arguments
	doc.Effort = fm.Effort
	doc.SkillContext = fm.SkillContext
	doc.Agent = fm.Agent
	doc.Shell = fm.Shell
	doc.Hooks = fm.Hooks
	doc.License = fm.License
	doc.Compatibility = fm.Compatibility
	doc.Metadata = fm.Metadata

	return doc, nil
}

// isRootPath 判断 path 是否为 standards 顶层 root.md（不在 rules/ skills/ 等子目录）。
// path 是相对 .stdai/standards/ 的路径，所以顶层 root.md 直接是 "root.md"。
// 不区分大小写支持 ROOT.md / Root.md。
func isRootPath(path string) bool {
	lower := strings.ToLower(path)
	return lower == "root.md" || lower == "root.markdown"
}

// nestedRootPath 检查 path 是否为嵌套 root，形如 "nested/<sub-path>/root.md"。
// 返回 nestedSubPath（即 <sub-path>，相对项目根的目录），第二个返回值表示是否匹配。
//
// 例：
//
//	"nested/igx-modules/auth/root.md"        -> "igx-modules/auth", true
//	"nested/src/api/v1/root.md"              -> "src/api/v1", true
//	"root.md"                                -> "", false
//	"nested/root.md"                         -> "", false  (路径深度不足，nested/ 直接挂 root 没有子路径)
func nestedRootPath(p string) (string, bool) {
	const prefix = "nested/"
	const suffix = "/root.md"
	lower := strings.ToLower(p)
	if !strings.HasPrefix(lower, prefix) {
		return "", false
	}
	if !strings.HasSuffix(lower, suffix) && !strings.HasSuffix(lower, "/root.markdown") {
		return "", false
	}
	suf := suffix
	if strings.HasSuffix(lower, "/root.markdown") {
		suf = "/root.markdown"
	}
	mid := p[len(prefix) : len(p)-len(suf)]
	if mid == "" {
		return "", false
	}
	return mid, true
}

// isValidName 检查 kebab-case 格式 ^[a-z0-9][a-z0-9-]*$
func isValidName(s string) bool {
	return nameRe.MatchString(s)
}

// filenameToName 从 path 推断 kebab-case name
func filenameToName(path string) string {
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	name = strings.ReplaceAll(name, "_", "-")
	name = strings.ReplaceAll(name, ".", "-")
	return strings.ToLower(name)
}
