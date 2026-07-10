# Target: Cursor

调研日期: 2026-05-07
官方文档: https://cursor.com/docs/

## 1. 摘要

Cursor 主推 `.cursor/rules/*.mdc` 多文件 + frontmatter 的 MDC 格式。`.cursorrules`
legacy 文件已 deprecated 且在 Agent mode 下静默失效。Commands 与 MCP 各自有独立目录。
`AGENTS.md` 作为 fallback 与 `.cursor/rules` 并行加载。

Cursor 已正式推出独立于 rules 的 Agent Skills 概念，目录为 `.cursor/skills/`
与 `.agents/skills/`，文件名约定 `SKILL.md`。

## 2. 配置文件路径

| 类别 | 路径 | 说明 |
|---|---|---|
| 项目 rules | `.cursor/rules/*.mdc`（也接受 `.md`） | 支持嵌套子目录；任意子文件夹也可有自己的 `.cursor/rules` |
| 项目 skills | `.cursor/skills/<name>/SKILL.md` 或 `.agents/skills/<name>/SKILL.md` | 每个 skill 一个文件夹 |
| 项目 commands | `.cursor/commands/<name>.md` | 文件名即 slash command 名 |
| 项目 MCP | `.cursor/mcp.json` | 项目级 MCP server 列表 |
| 用户 rules | Cursor Settings UI -> Rules（GUI 存储） | Agent/Chat 全局生效，**不影响 Cmd/Ctrl+K Inline Edit** |
| 用户 skills | `~/.cursor/skills/<name>/SKILL.md` 或 `~/.agents/skills/<name>/SKILL.md` | 全局个人 skills |
| 用户 commands | `~/.cursor/commands/<name>.md` | 全局个人 slash |
| 用户 MCP | `~/.cursor/mcp.json` | 全局 MCP |
| Team rules | Cursor Dashboard 管理（Team/Enterprise） | 分发到该 team 所有项目 |
| AGENTS.md fallback | 项目根 + 嵌套子目录 | 简化 rule，无 frontmatter |
| `.cursorrules` legacy | 项目根 | deprecated；与 `.cursor/rules` 共存时 MDC 静默胜出 |

## 3. MDC 格式（rules）

`.mdc` = YAML frontmatter + Markdown 正文。frontmatter 字段：

| 字段 | 类型 | 含义 |
|---|---|---|
| `description` | string | rule 用途；Agent Requested 模式必填 |
| `globs` | string（逗号分隔） / array | 匹配文件路径自动挂载 |
| `alwaysApply` | bool | true 时每个会话恒注入，忽略 globs/description |

## 4. 4 种激活模式（rules）

| 模式 | alwaysApply | globs | description | 触发 |
|---|---|---|---|---|
| Always | true | - | 可选 | 每条会话恒挂载 |
| Auto-Attached | false | 设置 | 可选 | 命中 glob 文件进入上下文时挂载 |
| Agent Requested | false | 空 | 必填 | Agent 据 description 决定是否拉取 |
| Manual | false | 空 | 可选 | 用户 `@rule-name` 手动调用 |

## 5. Skills 格式

每个 skill 是一个文件夹，必须包含 `SKILL.md`。frontmatter 字段：

| 字段 | 必填 | 含义 |
|---|---|---|
| `name` | 是 | 小写 + 连字符，须与文件夹名一致 |
| `description` | 是 | 说明何时使用 |
| `paths` | 否 | glob 限定生效范围 |
| `license` | 否 | 许可证 |
| `compatibility` | 否 | 兼容性声明 |
| `metadata` | 否 | 自定义元数据 |
| `disable-model-invocation` | 否 | bool；true 时仅响应 `/skill-name` slash 调用 |

加载方式：双模式

- 默认自动 - agent 根据 `description` 决定是否拉取整份 SKILL.md 进入上下文
- 显式 - 用户在 Agent Chat 输入 `/skill-name` 触发
- 设置 `disable-model-invocation: true` 后退化为纯 slash 命令

## 6. 加载与优先级

- 加载顺序：**Team Rules > Project Rules > User Rules**；冲突时较早层胜出
- `.mdc` 与 `.cursorrules` 共存时 MDC 胜出，无警告
- 嵌套：`.cursor/rules/<sub>/...` 完整支持

## 7. 字符上限

通用 rules 字符上限：**100,000 字符**（自 2025-04-18 由 20,000 提升至 100,000）。
超限触发截断并提示 `Rule exceeds X characters and may be truncated`。

User Rules 是否独立适用此上限：INSUFFICIENT-EVIDENCE。**std-agent 默认按 100k 字符处理**，
单条超限时 WARN 并截断。

## 8. std-agent 四类映射

| std-agent 类型 | Cursor 落点 |
|---|---|
| rules | `.cursor/rules/<name>.mdc`（项目）；frontmatter 由 std `applyTo` `alwaysApply` `description` 直接 mapped |
| skills | `.cursor/skills/<name>/SKILL.md`（带 frontmatter，与 Claude Code skills 字段对齐） |
| commands | `.cursor/commands/<name>.md`（项目）；正文为 prompt 模板 |
| references | 不主动写入；推荐作为 rules 内段落或独立 docs/ 目录 |

## 9. 转换器实现要点

1. std rule 源 frontmatter -> mdc frontmatter 字段映射：
   - `applyTo: ["**/*.ts"]` -> `globs: "**/*.ts"`
   - `alwaysApply: true` -> `alwaysApply: true`
   - `description: ...` -> `description: ...`
2. 输出文件名：`<name>.mdc`（保留 std 的 kebab-case `name`）
3. 不生成 `.cursorrules`（legacy）
4. `AGENTS.md` 默认不写（已由 codex transformer 生成根目录主入口）
5. commands 输出到 `.cursor/commands/<name>.md`，不带 frontmatter，正文取
   std 文件正文 + footer
6. skills 输出 `.cursor/skills/<name>/SKILL.md`：
   - `name` 必须与文件夹名一致
   - `description` 必填
   - std `applyTo` -> `paths`
   - std `disable_model_invocation`（如有）-> `disable-model-invocation`
7. 单文件超过 100,000 字符 -> WARN，按字节截断或拆分多文件
8. v1.0 不写 `.cursor/mcp.json`，留 v1.1

## 10. 信息来源

- https://cursor.com/docs/rules
- https://cursor.com/docs/skills
- https://cursor.com/docs/cli/reference/slash-commands
- https://cursor.com/changelog/1-6
- https://forum.cursor.com/t/deprecating-notepads-in-cursor/138305
- https://forum.cursor.com/t/how-to-use-agent-skills-in-cursor-ide/149860
- https://forum.cursor.com/t/20k-characters-limit-for-rules-is-too-restrictive/80739

## 11. 已确认与剩余 UNKNOWN

已确认：
- Skills 目录与 frontmatter schema（见 # 5）
- 通用 rules 字符上限 100k（自 2025-04-18 由 20k 提升）
- Skills 双模式加载（自动 + slash），`disable-model-invocation: true` 退化为纯 slash

剩余 UNKNOWN：
- User Rules 是否独立适用 100k 上限（按通用值默认）
- Team Rules 分发的 API / 协议
- Cursor 是否有 hooks 机制
