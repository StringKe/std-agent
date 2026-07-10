package writer

// FileOp 描述单个文件的写盘动作
type FileOp struct {
	Path    string // 相对项目根
	Content []byte
	Marker  bool   // 是否注入 stdagent marker
	IsRoot  bool   // 是否为 target 根文件（CLAUDE.md / AGENTS.md / GEMINI.md / .github/copilot-instructions.md），runner 触发 budget root-file 检查
	Skip    bool   // 与现有文件一致时跳过
	Reason  string // 诊断（dry-run 输出）
	// JSONMerge=true 时 Content 是 JSON 片段：目标文件不存在则直接写入片段；
	// 存在则深合并（object 递归、array 去重并集、scalar 保留已有值）后写回。
	// 目标文件不是合法 JSON（如带注释的 JSONC）时跳过并在 Reason 写 WARN，
	// 绝不破坏用户手写配置。crush.json / kilo.jsonc 注册项用。
	JSONMerge bool
}

// Plan 是 target transformer 计算出的写入计划
type Plan struct {
	Target string
	Files  []FileOp
}
