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
	Type                   string   `yaml:"type"`
	Name                   string   `yaml:"name"`
	Version                string   `yaml:"version"`
	Description            string   `yaml:"description"`
	Targets                []string `yaml:"targets"`
	ExcludeTargets         []string `yaml:"exclude_targets"`
	Priority               string   `yaml:"priority"`
	Tags                   []string `yaml:"tags"`
	ApplyTo                []string `yaml:"applyTo"`
	AlwaysApply            bool     `yaml:"alwaysApply"`
	AllowedTools           []string `yaml:"allowed_tools"`
	ArgumentHint           string   `yaml:"argument_hint"`
	Model                  string   `yaml:"model"`
	DisableModelInvocation bool     `yaml:"disable_model_invocation"`

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
		return doc, nil
	}

	var fm rawFrontmatter
	if err := yaml.Unmarshal(front, &fm); err != nil {
		return nil, fmt.Errorf("%s: invalid YAML frontmatter: %w", path, err)
	}

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
	doc.ApplyTo = fm.ApplyTo
	doc.AlwaysApply = fm.AlwaysApply
	doc.AllowedTools = fm.AllowedTools
	doc.ArgumentHint = fm.ArgumentHint
	doc.Model = fm.Model
	doc.DisableModelInvocation = fm.DisableModelInvocation

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
