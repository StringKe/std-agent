package parser

// DocType 是 std 文件 frontmatter type 字段枚举
type DocType string

// 四种 std type 枚举
const (
	TypeRules      DocType = "rules"
	TypeSkills     DocType = "skills"
	TypeCommands   DocType = "commands"
	TypeReferences DocType = "references"
)

// Priority 控制 rule 在拼接输出中的相对位置
type Priority string

// 三种 priority 枚举
const (
	PriorityHigh   Priority = "high"
	PriorityNormal Priority = "normal"
	PriorityLow    Priority = "low"
)

// SkillFile 是 SKILL package 内的辅助文件（scripts/ references/ assets/ etc）
//
// Path 是相对 skill 目录的相对路径，如 "scripts/lint.sh"、"references/checklist.md"。
// Raw 是原始 bytes，由 source / runner 在 parse 后填充。
type SkillFile struct {
	Path string
	Raw  []byte
}

// Document 是解析后的 std 源文件抽象表达
type Document struct {
	Path                   string
	Type                   DocType
	Name                   string
	Version                string
	Description            string
	Targets                []string
	ExcludeTargets         []string
	Priority               Priority
	Tags                   []string
	ApplyTo                []string
	AlwaysApply            bool
	AllowedTools           []string
	ArgumentHint           string
	Model                  string
	DisableModelInvocation bool
	Body                   string

	// SKILL package 扩展字段（agentskills.io 标准 + Claude Code 私有扩展）
	WhenToUse     string                 // Claude Code: 触发匹配补充
	Arguments     []string               // Claude Code: 命名位置参数
	Effort        string                 // Claude Code: low/medium/high/xhigh/max
	SkillContext  string                 // Claude Code: "fork" 表示子代理隔离上下文
	Agent         string                 // Claude Code: context=fork 时指定子代理类型
	Shell         string                 // Claude Code: bash/powershell
	Hooks         map[string]interface{} // Claude Code: skill 生命周期 hook
	License       string                 // agentskills 标准
	Compatibility string                 // agentskills 标准
	Metadata      map[string]interface{} // agentskills 自由 map（author / version / tags 等）

	// SKILL package 辅助文件（仅 type=skills 时非空）
	SkillFiles []SkillFile

	// BodyBytes 是 Body 的字节数，由 Parse 在解析时填充
	// 用于 budget 包做上下文消耗检查
	BodyBytes int
}

// IsValidType 检查 type 在四种枚举内
func IsValidType(t string) bool {
	switch DocType(t) {
	case TypeRules, TypeSkills, TypeCommands, TypeReferences:
		return true
	}
	return false
}

// IsValidPriority 检查 priority 合法
func IsValidPriority(p string) bool {
	switch Priority(p) {
	case PriorityHigh, PriorityNormal, PriorityLow, "":
		return true
	}
	return false
}

// PriorityRank 把 priority 翻成排序整数（越小越靠前）
func PriorityRank(p Priority) int {
	switch p {
	case PriorityHigh:
		return 100
	case PriorityNormal, "":
		return 500
	case PriorityLow:
		return 900
	}
	return 500
}
