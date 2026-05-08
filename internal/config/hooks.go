package config

// HooksConfig 是 .stdai/standards/hooks.json 的顶层 schema
//
// 只在 runtime 由 runner 加载注入到 Config.Hooks 字段，不参与 toml 持久化。
type HooksConfig struct {
	Version string                 `json:"version"`
	Hooks   map[string][]HookEntry `json:"hooks"`
}

// HookEntry 是一个 hook 触发条件 + 命令。schema 与 Anthropic Claude Code
// settings.json hooks 字段对齐：
//
//	{ "matcher": "Bash", "type": "command", "command": "echo pre-bash" }
//
// 不同 target 的 hooks 转换器按各自 schema 输出（部分字段可能丢弃或重映射）。
type HookEntry struct {
	Matcher string `json:"matcher,omitempty"` // 工具名匹配（PreToolUse/PostToolUse 用）
	Type    string `json:"type,omitempty"`    // command / 等
	Command string `json:"command,omitempty"` // 实际 shell 命令
	Timeout int    `json:"timeout,omitempty"` // 秒；可选
}
