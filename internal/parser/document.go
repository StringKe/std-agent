package parser

// DocType 是 std 文件 frontmatter type 字段枚举
type DocType string

// 四种 std type 枚举
const (
	TypeRules      DocType = "rules"
	TypeSkills     DocType = "skills"
	TypeCommands   DocType = "commands"
	TypeReferences DocType = "references"
	// TypeSubagents 是 Claude Code 原生的 subagent 定义（spawnable 子代理，独立 context）。
	// 输出到 .claude/agents/<name>.md，frontmatter 含 name / description / model / tools。
	// 与 SKILL（按需触发能力包，main session 内联使用）区别：subagent 是隔离 context 的子进程。
	TypeSubagents DocType = "subagents"
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
	Path           string
	Type           DocType
	Name           string
	Version        string
	Description    string
	Targets        []string
	ExcludeTargets []string
	Priority       Priority
	Tags           []string
	ApplyTo        []string
	AlwaysApply    bool
	// Root 标识此 rule 是根文件（CLAUDE.md / AGENTS.md / GEMINI.md / .github/copilot-instructions.md）的项目总结主体。
	// transformer 把 Body 写到根文件头部，再由 stdagent 自动追加 rule manifest 段（指向其他 nonRoot rule）。
	// root rule 本身不再 fan-out 到 .claude/rules/<n>.md（它已是根文件主体）。
	// 用户写 root rule 时**不应**在 body 里手写 rule 索引清单（stdagent 自动追加）。
	Root bool
	// NestedPath 非空时，本 rule 是嵌套子目录的说明文档（.stdai/standards/nested/<path>/root.md），
	// transformer 把 Body 写到 <NestedPath>/CLAUDE.md 与 <NestedPath>/AGENTS.md，**不**追加 manifest（嵌套位置只是说明文档）。
	// AI 在该子目录工作时 Claude Code 自动叠加加载顶级 + 嵌套 CLAUDE.md。
	NestedPath string
	// TargetPaths 是 target 专属 paths 覆盖（rulesync 风格嵌套字段）。
	// key 是 rulesync target 名（claudecode / codexcli / cursor / copilot / 等），
	// value 是该 target 用的 glob 列表，覆盖全局 ApplyTo。
	// transformer 通过 RulesyncTargetName 映射决定优先级。
	TargetPaths            map[string][]string
	AllowedTools           []string
	ArgumentHint           string
	Model                  string
	DisableModelInvocation bool
	// UserInvocable 控制 skill / command 是否进 slash 菜单（Claude Code / Copilot /
	// Factory / Grok 等 2026 起支持）。nil = 未设置（工具默认 true），指针区分显式 false。
	UserInvocable *bool
	// DisallowedTools 在 skill 活跃期间从工具池移除的工具（Claude Code skills），
	// 或 subagent 的禁用工具（Claude Code subagent disallowedTools）。
	DisallowedTools []string
	// ReadOnly 限制 subagent 写权限（Cursor .cursor/agents readonly 字段）
	ReadOnly bool
	// Background 让 subagent 后台运行（Cursor is_background / Claude Code background）
	Background bool
	Body       string

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
	case TypeRules, TypeSkills, TypeCommands, TypeReferences, TypeSubagents:
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
