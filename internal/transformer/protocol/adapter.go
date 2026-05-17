package protocol

import "std-ai/internal/parser"

// Adapter 是每个 target 注入协议族的配置 struct。字段全部 zero-value 安全，
// 协议实现根据零值决定行为，避免在调用侧写大量 if-else。
//
// # 零值语义（不可改写，是协议实现假设的契约）
//
//   - Name="" -> 非法（协议返回 error）
//   - Disabled=true -> Plan 直接返回空 Plan{Target: Name}，不产任何 FileOp（aider 用）
//   - RootFileName="" -> 不写根文件（极少见，amp / warp 等纯 inline target 可能为空）
//   - 5 个 Dir 字段为空 -> 该 type 走 graceful degradation
//     (RulesDir / SkillsDir / CommandsDir / ReferencesDir / SubagentsDir)
//   - GlobsFieldName="" -> 不渲染 globs 字段（无论 GlobsFieldFormat 怎么设）
//   - SubagentInvokeCmd="" 且 SubagentsDir="" -> subagent 走形态 A（路径降级到 FallbackDir/subagents/）
//   - SubagentInvokeCmd!="" -> subagent 走形态 B（CLI 调用降级，body 含 shell 调用指引）
//   - FallbackDir="" -> 用 RulesDir 兜底
//   - FallbackSubdir=nil 或 key 缺失 -> 按 defaultFallbackSubdir 自动 "skills"/"commands"/"references"/"subagents"
//   - InjectExplainer=false -> fallback 文件 body 头部不注入 HTML 注释 explainer
//   - InjectExplainerOverride[t]=false -> 即使 InjectExplainer=true，该 type 也不注入
//   - InjectStdaiTypeField=false -> fallback 文件 frontmatter 不写 std-ai-type 字段
//   - RulePrefix=nil -> rule 文件无数字前缀（cline / roo / kilo 才设置）
//   - RuleTriggerMode=TriggerNone -> rule frontmatter 不写 trigger / alwaysApply / applyTo
//   - MaxBytesPerFile=0 -> 不做 HARD limit 检查
//   - SoftBytes=0 -> 不做 SOFT WARN 检查
//   - InjectTypeGlossary=false -> 根文件不注入 std-ai 类型速查段
//   - CommandsFileSuffix="" -> 默认 ".md"
//   - SkillsAsSubagent=false -> skill 走标准 Agent Skills 包 / fallback；true 时
//     输出扁平 <SkillsDir>/<name>.md 含 mode: subagent frontmatter（opencode 风格）
//   - SkillsSubagentPermission="" -> SkillsAsSubagent=true 时不渲染 permission 字段
type Adapter struct {
	Name string

	// Transformer 隔离
	Disabled bool

	// 根文件
	RootFileName     string
	RootBodyTemplate string
	ManifestSection  string
	NestedSupported  bool
	NestedFileName   string

	// 5 种 type 原生落点（空字符串 = fallback；和 Disabled 字段配合判定"完全不输出"）
	RulesDir      string
	SkillsDir     string
	CommandsDir   string
	ReferencesDir string
	SubagentsDir  string

	// Fallback（v3：子目录隔离 + frontmatter 标识，无前缀）
	FallbackDir             string
	FallbackSubdir          map[parser.DocType]string
	InjectExplainer         bool
	InjectExplainerOverride map[parser.DocType]bool
	InjectStdaiTypeField    bool

	// Rules frontmatter 方言
	GlobsFieldName      string
	GlobsFieldFormat    GlobsFormat
	SupportsAlwaysApply bool
	SupportsDescription bool
	RulePrefix          func(d *parser.Document) string
	RuleTriggerMode     TriggerMode

	// Skills
	SkillSupportedFields []string
	SkillsAsRule         bool
	SkillsAsCmd          bool
	// SkillsAsSubagent=true：skill 输出为 <SkillsDir>/<name>.md 扁平文件，
	// frontmatter 含 mode: subagent + description + model + permission（opencode 风格）。
	// 检测到 d.SkillFiles 非空时 op.Reason 加 WARN（单文件不支持 SKILL package 子目录）。
	SkillsAsSubagent bool
	// SkillsSubagentPermission 是 SkillsAsSubagent=true 时写入 frontmatter 的 permission YAML inline。
	// 为空时不渲染 permission 字段。opencode v1.0 简化为全 ask（保守安全）。
	SkillsSubagentPermission string

	// Commands
	CommandFormat         CommandFormat
	CommandFrontmatter    []string
	InjectCommandsToRoot  bool
	CommandsAsSkillPrefix string
	// CommandsFileSuffix 是 command 输出文件扩展（含前导点），如 ".md" / ".prompt.md"。
	// 空值默认 ".md"。windsurf / antigravity workflows 用 ".md"，continue-dev prompts 用 ".prompt.md"。
	CommandsFileSuffix string

	// Subagents
	SubagentInvokeCmd string

	// 嵌套 + per-dir override（grok-cli）
	PerDirOverrideFileName string

	// 多文件单文件 fallback（cline / roo / augment）
	SingleFileFallback string

	// MCP
	MCPPath   string
	MCPTopKey string

	// 限制
	MaxBytesPerFile int
	SoftBytes       int

	// glossary 注入（N3）
	InjectTypeGlossary bool
}

// GlobsFormat 枚举：rule frontmatter 中 globs 字段的渲染方式
type GlobsFormat int

// GlobsFormat 取值
const (
	// GlobsList 渲染为 YAML list：
	//   globs:
	//     - "**/*.go"
	//     - "**/*.md"
	GlobsList GlobsFormat = iota
	// GlobsCommaString 渲染为逗号分隔字符串：
	//   globs: "**/*.go,**/*.md"
	GlobsCommaString
)

// CommandFormat 枚举：commands 输出文件格式
type CommandFormat int

// CommandFormat 取值
const (
	// CommandMarkdown 直接写 markdown 文件 + frontmatter
	CommandMarkdown CommandFormat = iota
	// CommandTOML 写 TOML 文件（gemini 用）
	CommandTOML
	// CommandSkillPrefix 把 command 转写为 skill，文件名加 CommandsAsSkillPrefix 前缀
	CommandSkillPrefix
)

// TriggerMode 枚举：rule frontmatter 中"何时触发"字段的方言
type TriggerMode int

// TriggerMode 取值
const (
	// TriggerNone 不写任何 trigger 字段
	TriggerNone TriggerMode = iota
	// TriggerAlwaysApply 写 alwaysApply: true / false（cursor / continue-dev）
	TriggerAlwaysApply
	// TriggerTrigger 写 trigger: always_on / glob / model_decision / manual（windsurf 系）
	TriggerTrigger
	// TriggerApplyTo 写 applyTo: globs（copilot 系）
	TriggerApplyTo
)
