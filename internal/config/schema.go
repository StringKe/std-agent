package config

// Config 是 .stdai/config.toml 的顶层 schema
//
// 重要约束：所有标量字段（version/name/inject/...）必须在 map 类型字段
// （Targets/Sources/Overrides）之前，否则 toml.Marshal 输出后这些标量会被
// 解析进上一个 [section]，导致 Round-trip 失败。
type Config struct {
	Version            string `toml:"version"`
	Name               string `toml:"name"`
	Inject             bool   `toml:"inject"`
	InjectWhatIs       bool   `toml:"inject_whatis"`
	InjectTypeGlossary bool   `toml:"inject_type_glossary"`
	DryRun             bool   `toml:"dry_run"`
	Backup             bool   `toml:"backup"`
	BackupKeep         int    `toml:"backup_keep"`
	AutoPull           bool   `toml:"auto_pull"`
	Verbose            bool   `toml:"verbose"`

	Targets   map[string]TargetConfig `toml:"targets"`
	Sources   map[string]SourceConfig `toml:"sources"`
	Overrides map[string]Override     `toml:"overrides"`

	// MCP 是 runtime 注入的 MCP 配置（来自 .stdai/standards/mcp.json）
	MCP *MCPConfig `toml:"-"`
}

// TargetConfig 是单个 target 的开关
type TargetConfig struct {
	Enabled bool `toml:"enabled"`
	Convert bool `toml:"convert"`
}

// SourceConfig 是单个 git 源的配置
type SourceConfig struct {
	URL      string   `toml:"url"`
	Branch   string   `toml:"branch"`
	Enabled  bool     `toml:"enabled"`
	Paths    []string `toml:"paths"`
	Auth     string   `toml:"auth"`
	TokenEnv string   `toml:"token_env"`
}

// Override 是 per-target 字段覆盖
type Override struct {
	Inject       *bool `toml:"inject"`
	InjectWhatIs *bool `toml:"inject_whatis"`
}

// ValidTargets 是合法 target 名清单（与 plan.md 附录 A Tier 权威表对齐）
var ValidTargets = []string{
	// Tier 1（用户基数大 / 协议主流）
	"claude-code", "codex", "cursor", "copilot", "windsurf",
	"gemini", "aider", "cline", "opencode",
	"roo-code", "crush", "amp", "warp", "factory",
	// Tier 2（新兴 / 小众 / 半实验）
	"continue-dev", "antigravity",
	"qwen-code", "pi", "kilo-code", "augment-code", "jules", "grok-build",
	"kimi-code",
}

// IsValidTarget 检查 target 名合法
func IsValidTarget(name string) bool {
	for _, t := range ValidTargets {
		if t == name {
			return true
		}
	}
	return false
}

// Default 返回 init 时使用的默认配置
func Default() *Config {
	return &Config{
		Version:            "1.0",
		Inject:             true,
		InjectWhatIs:       true,
		InjectTypeGlossary: true,
		Backup:             true,
		BackupKeep:         5,
		AutoPull:           true,
		Targets: map[string]TargetConfig{
			// Tier 1
			"claude-code": {Enabled: true, Convert: true},
			"codex":       {Enabled: true, Convert: true},
			"cursor":      {Enabled: false, Convert: true},
			"copilot":     {Enabled: false, Convert: true},
			"windsurf":    {Enabled: false, Convert: true},
			"gemini":      {Enabled: false, Convert: true},
			"aider":       {Enabled: false, Convert: true},
			"cline":       {Enabled: false, Convert: true},
			"opencode":    {Enabled: false, Convert: true},
			"roo-code":    {Enabled: false, Convert: true},
			"crush":       {Enabled: false, Convert: true},
			"amp":         {Enabled: false, Convert: true},
			"warp":        {Enabled: false, Convert: true},
			"factory":     {Enabled: false, Convert: true},
			// Tier 2
			"continue-dev": {Enabled: false, Convert: true},
			"antigravity":  {Enabled: false, Convert: true},
			"qwen-code":    {Enabled: false, Convert: true},
			"pi":           {Enabled: false, Convert: true},
			"kilo-code":    {Enabled: false, Convert: true},
			"augment-code": {Enabled: false, Convert: true},
			"jules":        {Enabled: false, Convert: true},
			"grok-build":   {Enabled: false, Convert: true},
			"kimi-code":    {Enabled: false, Convert: true},
		},
	}
}
