# Std-Agent 文档

stdagent 是一个轻量级、纯 Go 实现的 AI CLI 配置同步工具。以 `.stdai/` 为
内部单一真相源，把 YAML frontmatter + Markdown 同步到 11 个 AI CLI 工具的
扩散文件，外加 MCP 服务器配置。

## 文档索引

### 必读

- **[spec.md](spec.md) — 完整 spec：std-ai 标准 + 11 工具差异 + 转换实现策略**
- [../README.md](../README.md) 项目入口与快速开始
- [prd.md](prd.md) 产品需求
- [roadmap.md](roadmap.md) 路线图

### 规范

- [format-spec.md](format-spec.md) frontmatter 详细 schema
- [config-spec.md](config-spec.md) `.stdai/config.toml` 字段
- [commands.md](commands.md) CLI 命令规范
- [conversion-rules.md](conversion-rules.md) 转换矩阵 + frontmatter 字段映射
- [file-structure.md](file-structure.md) 目录结构与扩散原则

### 实现

- [architecture.md](architecture.md) 模块划分与数据流

### 11 个目标工具调研

Tier 1（9 个）：

- [targets/claude-code.md](targets/claude-code.md)
- [targets/codex.md](targets/codex.md)
- [targets/cursor.md](targets/cursor.md)
- [targets/github-copilot.md](targets/github-copilot.md)
- [targets/windsurf.md](targets/windsurf.md)
- [targets/gemini-cli.md](targets/gemini-cli.md)
- [targets/aider.md](targets/aider.md)
- [targets/cline.md](targets/cline.md)
- [targets/opencode.md](targets/opencode.md)

Tier 2（2 个）：

- [targets/continue.md](targets/continue.md)
- [targets/antigravity.md](targets/antigravity.md)

## 阅读顺序建议

1. 项目根 `README.md` 看产品全貌
2. `docs/spec.md` 是权威参考，理解概念差异 + 我们 spec
3. `docs/architecture.md` 看实现模块
4. `docs/commands.md` 看每个 CLI 命令
5. `docs/targets/<X>.md` 按需查阅特定工具

## 字符规范

所有 markdown / 代码 / commit message 遵守 ASCII + 中文 + 全角中文标点 + 数学符号
（× ÷ ± ≥ ≤）字符白名单。禁止 emoji、装饰星号（★）、装饰勾叉、em/en dash、smart quotes。
优先级用 P0/P1/P2，状态用 PASS/FAIL/SKIP，进度用"已落地/进行中/未开始"。
