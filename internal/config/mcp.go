package config

// MCPConfig 是 .stdai/standards/mcp.json 的顶层 schema
//
// 只在 runtime 由 runner 加载注入到 Config.MCP 字段，不参与 toml 持久化。
type MCPConfig struct {
	Version string               `json:"version"`
	Servers map[string]MCPServer `json:"servers"`
}

// MCPServer 描述单个 MCP server，覆盖 stdio 与 http/sse 两类
type MCPServer struct {
	Type    string            `json:"type,omitempty"`    // stdio | http | sse
	Command string            `json:"command,omitempty"` // stdio
	Args    []string          `json:"args,omitempty"`    // stdio
	URL     string            `json:"url,omitempty"`     // http/sse
	Env     map[string]string `json:"env,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}
