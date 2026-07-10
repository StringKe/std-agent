# Target: Cursor

调研日期: 2026-05-07，2026-07-10 复核更新
官方文档: https://cursor.com/docs/

## 1. 摘要

Cursor 主推 `.cursor/rules/*.mdc` 多文件 + frontmatter 的 MDC 格式。`.cursorrules`
legacy 文件已 deprecated 且在 Agent mode 下静默失效。**`.md` 后缀在 `.cursor/rules/`
下被静默忽略，只认 `.mdc`**（2026-07 复核修正：旧文档"也接受 `.md`"过时；stdagent 输出
本身一直是 `.mdc` 后缀，不受影响）。Commands 与 MCP 各自有独立目录。
`AGENTS.md` 作为 fallback 与 `.cursor/rules` 并行加载。

Cursor 已正式推出独立于 rules 的 Agent Skills 概念，目录为 `.cursor/skills/`
与 `.agents/skills/`，文件名约定 `SKILL.md`。

2026-07 复核新增：Cursor 已原生支持 subagents `.cursor/agents/<name>.md`
（https://cursor.com/docs/subagents.md），transformer 已从降级到 `.cursor/rules/subagents/`
迁移为原生落点（见 # 8）。

## 2. 配置文件路径

| 类别 | 路径 | 说明 |
|---|---|---|
| 项目 rules | `.cursor/rules/*.mdc`（**仅 `.mdc`，`.md` 被忽略**） | 支持嵌套子目录；任意子文件夹也可有自己的 `.cursor/rules` |
| 项目 skills | `.cursor/skills/<name>/SKILL.md` 或 `.agents/skills/<name>/SKILL.md` | 每个 skill 一个文件夹 |
| 项目 commands | `.cursor/commands/<name>.md` | 文件名即 slash command 名 |
| 项目 subagents | `.cursor/agents/<name>.md` | 原生支持，frontmatter `name` / `description` / `model` / `readonly` / `is_background` |
| 项目 MCP | `.cursor/mcp.json` | 项目级 MCP server 列表 |
| 用户 rules | Cursor Settings UI -> Rules（GUI 存储） | Agent/Chat 全局生效，**不影响 Cmd/Ctrl+K Inline Edit** |
| 用户 skills | `~/.cursor/skills/<name>/SKILL.md` 或 `~/.agents/skills/<name>/SKILL.md` | 全局个人 skills |
| 用户 commands | `~/.cursor/commands/<name>.md` | 全局个人 slash |
| 用户 MCP | `~/.cursor/mcp.json` | 全局 MCP |
| Team rules | Cursor Dashboard 管理（Team/Enterprise） | 分发到该 team 所有项目 |
| AGENTS.md fallback | 项目根 + 嵌套子目录 | 简化 rule，无 frontmatter；**Cursor 官方对嵌套 AGENTS.md 完整支持（隐式 glob）**，与 `CLAUDE.md` 同时读并叠加 |
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

## 6. Subagents 格式（2026-07 新增，原生）

frontmatter 字段：

| 字段 | 必填 | 含义 |
|---|---|---|
| `name` | 是 | subagent 名 |
| `description` | 是 | 何时委派此 subagent |
| `model` | 否 | 覆盖默认模型 |
| `readonly` | 否 | bool，是否只读（不允许写操作） |
| `is_background` | 否 | bool，是否后台运行 |

来源：https://cursor.com/docs/subagents.md

## 7. 加载与优先级

- 加载顺序：**Team Rules > Project Rules > User Rules**；冲突时较早层胜出
- `.mdc` 与 `.cursorrules` 共存时 MDC 胜出，无警告
- 嵌套：`.cursor/rules/<sub>/...` 完整支持（官方文档不再明确宣传此细节，2026-07 复核标 UNKNOWN，按历史行为保留）

## 8. 字符上限

通用 rules 字符上限：**100,000 字符**（自 2025-04-18 由 20,000 提升至 100,000）。
超限触发截断并提示 `Rule exceeds X characters and may be truncated`。

Cursor 对 `AGENTS.md` / `CLAUDE.md` 同样按单 rule 100,000 字符上限处理（服务端下发，
可能变动）。这两个根文件不是由 cursor transformer 自己写的（复用 codex / claude-code
transformer 的产出），但 Cursor 消费时套用同一截断逻辑，因此 `budget.go` 单独登记了
`cursor root-file soft 80000 hard 100000`。

User Rules 是否独立适用此上限：INSUFFICIENT-EVIDENCE。**std-agent 默认按 100k 字符处理**，
超限只发 WARN（见 # 11 转换器实现要点第 7 条），不做自动截断/拆分。

## 9. std-agent 五类映射（实际实现，`internal/transformer/cursor.go` / `protocol/cursor.go`）

| std-agent 类型 | Cursor 落点 |
|---|---|
| rules | `.cursor/rules/<name>.mdc`（含 root rule；Cursor 无单一根文件概念，root doc 与 nonRoot doc 走同一渠道生成 `.mdc`）；frontmatter 由 std `applyTo` `alwaysApply` `description` 直接 mapped |
| skills | `.cursor/skills/<name>/SKILL.md`（Agent Skills 标准，frontmatter `name` / `description`(合并 `when_to_use`) / `paths` / `disable-model-invocation` / `license` / `compatibility` / `metadata`） |
| commands | `.cursor/commands/<name>.md`（项目）；正文为 prompt 模板，无 frontmatter |
| subagents | `.cursor/agents/<name>.md`（原生，`SubagentsDir` 非空即走此分支） |
| references | `.cursor/rules/references/<name>.md`（子目录隔离降级，frontmatter `std-agent-type: references`） |

**当前实现限制**：`protocol/cursor.go` 的 `Plan` 方法未调用 `PartitionRoot` /
`PartitionNested`，`NestedPath` 非空的源文档目前会被当作普通 rule 写入
`.cursor/rules/<name>.mdc`（顶层），**不会**落到 `<NestedPath>/` 子目录。Cursor
官方对嵌套 AGENTS.md 的隐式 glob 支持是工具侧行为，与 stdagent 是否输出嵌套 `.mdc`
是两件事；这是本轮调研发现的实现细节，非 bug，但文档需明确标注（后续若要对齐
"嵌套子目录写专属 rule" 语义需要补代码）。

## 10. 转换器实现要点

1. std rule 源 frontmatter -> mdc frontmatter 字段映射：
   - `applyTo: ["**/*.ts"]` -> `globs: "**/*.ts"`（`GlobsCommaString`，逗号分隔字符串）
   - `alwaysApply: true` -> `alwaysApply: true`
   - `description: ...` -> `description: ...`
2. 输出文件名：`<name>.mdc`（保留 std 的 kebab-case `name`）
3. 不生成 `.cursorrules`（legacy）
4. `AGENTS.md` 不写（复用 codex / claude-code transformer 的产出）
5. commands 输出到 `.cursor/commands/<name>.md`，不带 frontmatter，正文取
   std `description` + `Body`
6. skills 输出 `.cursor/skills/<name>/SKILL.md`：
   - `name` 必须与文件夹名一致
   - `description` 合并 `when_to_use`
   - std `applyTo` -> `paths`
   - `disable-model-invocation`（如有）直传
7. 单文件超过 100,000 字符：仅由 `budget.go` 发 SOFT/HARD WARN 提示，**不做自动截断或拆分**（旧文档"截断或拆分多文件"过时）
8. `.cursor/mcp.json` **已实现**（`buildMCPJSON`，顶层键 `mcpServers`），旧文档"v1.0 不写，留 v1.1"过时
9. subagents 已实现原生落点 `.cursor/agents/<name>.md`

## 11. 信息来源

- https://cursor.com/docs/rules
- https://cursor.com/docs/skills
- https://cursor.com/docs/subagents.md
- https://cursor.com/docs/cli/reference/slash-commands
- https://cursor.com/changelog/1-6
- https://forum.cursor.com/t/deprecating-notepads-in-cursor/138305
- https://forum.cursor.com/t/how-to-use-agent-skills-in-cursor-ide/149860
- https://forum.cursor.com/t/20k-characters-limit-for-rules-is-too-restrictive/80739

## 12. 已确认与剩余 UNKNOWN（2026-07-10 复核）

已确认：
- Skills 目录与 frontmatter schema（见 # 5）
- 通用 rules 字符上限 100k（自 2025-04-18 由 20k 提升）
- Skills 双模式加载（自动 + slash），`disable-model-invocation: true` 退化为纯 slash
- Subagents 原生 schema（见 # 6），已落地实现
- MCP（`.cursor/mcp.json`）已实现，非 v1.1 待办
- `.md` 后缀在 `.cursor/rules/` 下被忽略，仅认 `.mdc`

剩余 UNKNOWN（2026-07 复核仍未证实）：
- User Rules 是否独立适用 100k 上限（按通用值默认）
- Team Rules 分发的 API / 协议
- Cursor 子目录 `.cursor/rules/<sub>/` 的官方现状（文档不再明确宣传，按历史行为保留）
- 嵌套 `CLAUDE.md` 是否被 Cursor 读取（仅确认嵌套 `AGENTS.md` 官方支持）
- Cursor 是否有 hooks 机制（记录为未来候选，非本轮实现范围）
