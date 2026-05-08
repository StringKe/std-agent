# Target: Antigravity

调研日期: 2026-05-08
官方文档: https://antigravity.google/docs
官方主页: https://antigravity.google

## 1. 摘要

Antigravity 是 Google 2025 年底发布的 AI-first IDE，基于 VS Code，
深度集成 Gemini，被定位为 Google 在 Cursor / Windsurf 之后的代际
答卷。配置体系绑定 Gemini 工具链：全局目录 `~/.gemini/`，IDE 私有
配置在 `~/.gemini/antigravity/`，与 Gemini CLI 共享 `~/.gemini/GEMINI.md`
（已知冲突，参见已确认章节）。

Antigravity 自 v1.20.3（2026-03-05）起原生消费 `AGENTS.md` 与
`GEMINI.md` 两份配置，叠加层级为：System Rules（不可改）-> `GEMINI.md`
（最高优先用户层）-> `AGENTS.md`（跨工具基础层）-> `.agents/rules/`
（工作区补充）。这是 Antigravity 与 Cursor / Windsurf 的关键差异：
原生支持 AGENTS.md，无需额外适配。

工具组件分三类：Rules（约束）/ Workflows（slash 命令式步骤序列）/
Skills（可复用知识包）。MCP 通过 `~/.gemini/antigravity/mcp_config.json`
全局配置，仅 stdio / streamable HTTP 两种 transport，与 Claude Desktop
canonical 格式兼容。
(https://antigravity.google/docs/rules-workflows, 访问日期 2026-05-08;
https://antigravity.google/docs/mcp, 访问日期 2026-05-08)

## 2. 配置文件路径

| 类型 | 路径 | 说明 |
|---|---|---|
| 全局 GEMINI.md | `~/.gemini/GEMINI.md` | Antigravity-only 全局规则 |
| 全局 AGENTS.md | `~/.gemini/AGENTS.md` | 跨工具全局规则（v1.20.3+） |
| 全局 MCP 配置 | `~/.gemini/antigravity/mcp_config.json` | MCP 服务器列表 |
| 全局 Workflows | `~/.gemini/antigravity/global_workflows/*.md` | 全局 slash 命令 |
| 全局 Skills | `~/.gemini/antigravity/skills/` | 全局 skill 目录 |
| MCP OAuth tokens | `~/.gemini/antigravity/mcp_oauth_tokens.json` | 自动管理，勿编辑 |
| 项目级 GEMINI.md | `<repo>/GEMINI.md` | 项目 Antigravity 专属 |
| 项目级 AGENTS.md | `<repo>/AGENTS.md` | 项目跨工具规则 |
| 嵌套 AGENTS.md | `<repo>/<subdir>/AGENTS.md` | 子目录限定（需 Settings 启用） |
| 工作区 Rules | `<repo>/.agents/rules/*.md` | 当前默认目录 |
| 工作区 Rules（旧） | `<repo>/.agent/rules/*.md` | v1.20.3 前默认，向后兼容 |
| 工作区 Workflows | `<repo>/.agents/workflows/*.md` | 项目 slash 命令 |
| 工作区 Skills | `<repo>/.agents/skills/` | 项目 skill 目录 |

Windows 路径将 `~/` 替换为 `%USERPROFILE%\`。
(https://antigravity.google/docs/rules-workflows, 访问日期 2026-05-08;
https://getmcp.es/guides/antigravity, 访问日期 2026-05-08;
https://antigravitylab.net/en/articles/tips/agents-md-guide,
访问日期 2026-05-08)

## 3. 文件格式

| 文件 | 格式 | frontmatter |
|---|---|---|
| `GEMINI.md` / `AGENTS.md` | Markdown | 无 frontmatter，纯指令文本 |
| `.agents/rules/*.md` | Markdown + YAML frontmatter | `trigger` / `description` / `globs` |
| `.agents/workflows/*.md` | Markdown + YAML frontmatter | UNKNOWN（文档未列出 frontmatter 字段） |
| `.agents/skills/<name>/` | 目录 + Markdown | UNKNOWN（公开文档信息有限） |
| `mcp_config.json` | JSON | 不适用 |

Rules 文件大小上限 12000 字符，Workflows 同上限。
(https://antigravity.google/docs/rules-workflows, 访问日期 2026-05-08)

## 4. Rules frontmatter 字段对照

```markdown
---
trigger: always_on
description: Core Project Instructions
---

# Rule body
```

| 字段 | 必填 | 取值 | 含义 |
|---|---|---|---|
| `trigger` | 是 | `always_on` / `manual` / `model_decision` / `glob` | 激活模式 |
| `description` | `model_decision` 时必填 | 字符串 | 模型基于此决定是否注入 |
| `globs` | `trigger: glob` 时必填 | 例如 `{ts,tsx}` 或 `src/**/*.py` | 文件 glob 限定 |

激活语义：

- `always_on`：每次会话注入
- `manual`：仅当用户在输入框 `@<rule-name>` 时注入
- `model_decision`：模型读 `description` 决定是否注入
- `glob`：当前活动文件匹配 `globs` 时注入

Rules 内可用 `@filename` 或 `@/abs/path/file.md` 引用其他文件作为
上下文。
(https://antigravity.google/docs/rules-workflows, 访问日期 2026-05-08)

## 5. std-ai 四类映射

| std-ai 类型 | Antigravity 落点 | 加载方式 |
|---|---|---|
| rules | `<repo>/.agents/rules/<name>.md`，frontmatter `trigger: always_on`（核心约束）或 `glob` + `globs:` | 自动扫描，按 trigger 决定是否注入 |
| skills | `<repo>/.agents/skills/<name>/`（项目）或 `~/.gemini/antigravity/skills/<name>/`（全局）。结构细节 UNKNOWN，最稳妥的回退是写为 `model_decision` rule | 自动按任务上下文加载 |
| commands | `<repo>/.agents/workflows/<name>.md`，调用方式 `/<name>` | slash 触发 |
| references | rule 正文用 `@filename` 引用；或写入 `AGENTS.md` 段落 | rule 加载时按引用解析 |

层级优先级（高 -> 低）：System Rules -> `GEMINI.md` -> `AGENTS.md`
-> `.agents/rules/`。同名规则冲突时高优先级胜出。
(https://antigravity.codes/blog/antigravity-agents-md-guide,
访问日期 2026-05-08)

## 6. 转换器实现要点

1. std-ai 默认主输出走项目根 `AGENTS.md`（与 codex / cursor / claude-code
   transformer 共用），Antigravity 自动消费，无需额外动作
2. 细粒度 rules（globs 限定 / model_decision）-> 写入
   `<repo>/.agents/rules/<name>.md`，frontmatter `trigger` 按
   std-ai 元数据决定：
   - 始终激活 -> `trigger: always_on`
   - 文件 glob -> `trigger: glob` + `globs:`
   - 描述驱动 -> `trigger: model_decision` + `description:`
   - 用户显式调用 -> `trigger: manual`
3. commands -> `<repo>/.agents/workflows/<name>.md`，文件名即 slash
   命令。frontmatter 字段 UNKNOWN，转换器先输出无 frontmatter 的纯
   markdown 步骤序列
4. skills 暂不主动适配独立目录结构（`.agents/skills/` 公开文档信息
   不足），降级为 `model_decision` rule 写入 `.agents/rules/`
5. 不修改 `~/.gemini/GEMINI.md` 全局文件，避免与 Gemini CLI 冲突
6. 不写 `mcp_config.json`：MCP 不在四类映射范围
7. 兼容性：保留对 `.agent/rules/`（旧路径）的输出选项；v1.0 默认
   写新路径 `.agents/rules/`

## 7. 信息来源

- https://antigravity.google/docs/rules-workflows
  （访问日期 2026-05-08）
- https://antigravity.google/docs/mcp （访问日期 2026-05-08）
- https://antigravity.google/ （访问日期 2026-05-08）
- https://antigravitylab.net/en/articles/tips/agents-md-guide
  （访问日期 2026-05-08）
- https://antigravity.codes/blog/antigravity-agents-md-guide
  （访问日期 2026-05-08）
- https://antigravity.codes/blog/user-rules
  （访问日期 2026-05-08）
- https://getmcp.es/guides/antigravity （访问日期 2026-05-08）
- https://github.com/omar-haris/smart-coding-mcp/blob/main/docs/ide-setup/antigravity.md
  （访问日期 2026-05-08）
- https://petronellatech.com/blog/google-antigravity-ide-setup-guide-2026/
  （访问日期 2026-05-08）

## 8. 已确认

- Antigravity v1.20.3（2026-03-05）起原生读取 `AGENTS.md`，与 Cursor /
  Windsurf 形成跨工具兼容线
- 优先级：System Rules（不可改）> `GEMINI.md` > `AGENTS.md` >
  `.agents/rules/`
- 工作区 rules 默认目录已从 `.agent/rules/` 迁移到 `.agents/rules/`，
  旧路径保持向后兼容
- Rules frontmatter 必带 `trigger`，四种激活模式：`always_on` /
  `manual` / `model_decision` / `glob`
- Rules / Workflows 文件大小上限 12000 字符
- MCP 配置在 `~/.gemini/antigravity/mcp_config.json`，根键 `mcpServers`，
  与 Claude Desktop canonical 格式一致；支持 stdio（`command` + `args`）
  与 streamable HTTP（`serverUrl`）两种 transport
- Antigravity 与 Gemini CLI 共享 `~/.gemini/GEMINI.md`，存在配置串扰，
  推荐用 `~/.gemini/AGENTS.md` 隔离跨工具内容
- 不要使用 `.vscode/mcp.json`：Antigravity 有独立 MCP 配置体系

## 9. UNKNOWN

- Workflows 文件 frontmatter 字段（公开文档仅说明文件本质是 markdown
  + 12000 字符上限，未列具体字段）
- `.agents/skills/` 与全局 `~/.gemini/antigravity/skills/` 的目录内
  结构（是否类似 Claude Code skill 目录、是否有 `SKILL.md` 入口、
  frontmatter 形式）。第三方文章提到 skill 概念但未给出官方 schema
- Antigravity CLI 命令（如部分文章提到 `antigravity mcp --init` 等）
  在官方文档中未确认，可能是第三方工具的扩展，转换器不依赖
- Hub / 远端 blocks 体系（类似 Continue Hub）暂未在 Antigravity 公开
  文档中出现，假设无此机制
