# Target: Antigravity

调研日期: 2026-05-08，2026-07-10 复核更新
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

2026-07 复核新增（P0 修复）：Skills 已 GA 默认启用，`antigravity.google/docs/skills`
确认 workspace 路径固定为 `.agents/skills/<name>/SKILL.md`，官方 frontmatter
仅 `name`（可选）+ `description`（必填）。旧文档"公开信息不足，降级为
model_decision rule"已过时，transformer 已实现原生落点（见 # 5 / # 6）。

## 2. 配置文件路径

| 类型 | 路径 | 说明 |
|---|---|---|
| 全局 GEMINI.md | `~/.gemini/GEMINI.md` | Antigravity-only 全局规则 |
| 全局 AGENTS.md | `~/.gemini/AGENTS.md` | 跨工具全局规则（v1.20.3+） |
| 全局 MCP 配置 | `~/.gemini/antigravity/mcp_config.json` | MCP 服务器列表 |
| 全局 Workflows | `~/.gemini/antigravity/global_workflows/*.md` | 全局 slash 命令 |
| 全局 Skills | `~/.gemini/antigravity/skills/` | 全局 skill 目录（与 workspace 路径的关系存在文档矛盾，见 # 9） |
| MCP OAuth tokens | `~/.gemini/antigravity/mcp_oauth_tokens.json` | 自动管理，勿编辑 |
| 项目级 GEMINI.md | `<repo>/GEMINI.md` | 项目 Antigravity 专属 |
| 项目级 AGENTS.md | `<repo>/AGENTS.md` | 项目跨工具规则；stdagent 由 codex transformer 写入，antigravity 复用不重复写 |
| 嵌套 AGENTS.md | `<repo>/<subdir>/AGENTS.md` | 子目录限定（**需 Settings 手动启用，默认关闭**；stdagent 侧 `NestedSupported=false`，与该默认关闭状态一致） |
| 工作区 Rules | `<repo>/.agents/rules/*.md` | 当前默认目录 |
| 工作区 Rules（旧） | `<repo>/.agent/rules/*.md` | v1.20.3 前默认，向后兼容 |
| 工作区 Workflows | `<repo>/.agents/workflows/*.md` | 项目 slash 命令 |
| 工作区 Skills | `<repo>/.agents/skills/<name>/SKILL.md` | 项目 skill 目录，官方 workspace 路径已确认（2026-07 修正为原生落点） |

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
| `.agents/workflows/*.md` | Markdown + YAML frontmatter | UNKNOWN（文档未列出 frontmatter 字段，transformer 现按通用 `description` / `argument-hint` / `tools` / `model` 渲染） |
| `.agents/skills/<name>/SKILL.md` | 目录 + Markdown | 官方仅 `name`（可选）+ `description`（必填）；transformer 为与 codex 共用同一 `SkillSupportedFields` 白名单（`name` / `description` / `license` / `compatibility` / `metadata`）额外渲染了 `license` / `compatibility` / `metadata`，官方忽略多余字段无害（2026-07 修正，见 # 5） |
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

## 5. std-agent 五类映射（实际实现，`internal/transformer/antigravity.go`）

| std-agent 类型 | Antigravity 落点 | 加载方式 |
|---|---|---|
| rules | `<repo>/.agents/rules/<name>.md`，frontmatter `trigger: always_on`（核心约束）或 `glob` + `globs:`（`RuleTriggerMode=TriggerTrigger`） | 自动扫描，按 trigger 决定是否注入 |
| skills | `<repo>/.agents/skills/<name>/SKILL.md`（原生 Agent Skills 标准包，2026-07 修正，非降级） | 自动按任务上下文加载 |
| commands | `<repo>/.agents/workflows/<name>.md`，调用方式 `/<name>` | slash 触发 |
| references | rule 正文用 `@filename` 引用；无原生类型走 fallback | rule 加载时按引用解析 |
| subagents | 无原生概念，走 fallback | - |

层级优先级（高 -> 低）：System Rules -> `GEMINI.md` -> `AGENTS.md`
-> `.agents/rules/`。同名规则冲突时高优先级胜出。
(https://antigravity.codes/blog/antigravity-agents-md-guide,
访问日期 2026-05-08)

## 6. 转换器实现要点（对照 `internal/transformer/antigravity.go`）

1. `RootFileName=""`：不重复写根 `AGENTS.md`（由 codex transformer 写入，antigravity 复用，官方自 v1.20.3 起原生消费）
2. 细粒度 rules -> `<repo>/.agents/rules/<name>.md`，frontmatter `trigger` 按
   std-agent 元数据决定：
   - 始终激活 -> `trigger: always_on`
   - 文件 glob -> `trigger: glob` + `globs:`
   - 描述驱动 -> `trigger: model_decision` + `description:`
   - 用户显式调用 -> `trigger: manual`
3. commands -> `<repo>/.agents/workflows/<name>.md`，文件名即 slash
   命令，frontmatter `description` / `argument-hint` / `tools` / `model`
4. skills **已实现原生落点** `<repo>/.agents/skills/<name>/SKILL.md`（`SkillsDir=".agents/skills"`），
   旧文档"降级为 model_decision rule"已过时（P0 修复）
5. 不修改 `~/.gemini/GEMINI.md` 全局文件，避免与 Gemini CLI 冲突
6. 不写 `mcp_config.json`：MCP 不在五类映射范围
7. `NestedSupported=false`：与官方"嵌套 AGENTS.md 需 Settings 手动启用，默认关闭"一致，不主动写嵌套根文件
8. 兼容性：保留对 `.agent/rules/`（旧路径）的读取兼容认知；stdagent 默认只写新路径 `.agents/rules/`
9. `MaxBytesPerFile=12000` / `SoftBytes=8000`：仅对 rule 文件生效（`buildRuleFile` 内检查），workflow 文件的 12000 上限由 `budget.go` 的通用 `command` kind 检查兜底（见 # 8）

## 7. 信息来源

- https://antigravity.google/docs/rules-workflows
  （访问日期 2026-05-08）
- https://antigravity.google/docs/mcp （访问日期 2026-05-08）
- https://antigravity.google/docs/skills （2026-07 新增）
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

## 8. budget.go 限额（2026-07 回填）

| kind | Soft | Hard | 语义 |
|---|---|---|---|
| rule（antigravity 专属） | - | 12000 字符 | 官方文档单文件上限 |
| command（antigravity 专属，2026-07 新增） | - | 12000 字符 | 官方文档："Workflow files are limited to 12,000 characters each" |

## 9. 已确认与剩余 UNKNOWN（2026-07-10 复核）

已确认：
- Antigravity v1.20.3（2026-03-05）起原生读取 `AGENTS.md`，与 Cursor /
  Windsurf 形成跨工具兼容线
- 优先级：System Rules（不可改）> `GEMINI.md` > `AGENTS.md` >
  `.agents/rules/`
- 工作区 rules 默认目录已从 `.agent/rules/` 迁移到 `.agents/rules/`，
  旧路径保持向后兼容
- Rules frontmatter 必带 `trigger`，四种激活模式：`always_on` /
  `manual` / `model_decision` / `glob`
- Rules / Workflows 文件大小上限 12000 字符（Workflows 上限 2026-07 补入 budget.go）
- MCP 配置在 `~/.gemini/antigravity/mcp_config.json`，根键 `mcpServers`，
  与 Claude Desktop canonical 格式一致；支持 stdio（`command` + `args`）
  与 streamable HTTP（`serverUrl`）两种 transport
- Antigravity 与 Gemini CLI 共享 `~/.gemini/GEMINI.md`，存在配置串扰，
  推荐用 `~/.gemini/AGENTS.md` 隔离跨工具内容
- 不要使用 `.vscode/mcp.json`：Antigravity 有独立 MCP 配置体系
- 嵌套 AGENTS.md 需 Settings 手动启用，默认关闭（P0，与 `NestedSupported=false` 实现一致）
- workspace skills 路径 `.agents/skills/<name>/SKILL.md` 已确认，官方 frontmatter 仅 `name`（可选）+ `description`（必填）

剩余 UNKNOWN（2026-07 复核仍未证实）：
- Workflows 文件完整 frontmatter 字段（公开文档仅说明文件本质是 markdown
  + 12000 字符上限，未列具体字段）
- 全局 skills 路径三态文档自相矛盾（`~/.gemini/antigravity/skills/` 与
  workspace `.agents/skills/` 的关系未理清；workspace 路径本身可信）
- Antigravity skills 的字节上限与超限行为（官方未给数值）
- Antigravity CLI 命令（如部分文章提到 `antigravity mcp --init` 等）
  在官方文档中未确认，可能是第三方工具的扩展，转换器不依赖
- Hub / 远端 blocks 体系（类似 Continue Hub）暂未在 Antigravity 公开
  文档中出现，假设无此机制
