# Target: Kiro (AWS)

调研日期: 2026-08-16
官方文档: https://kiro.dev/docs/
Amazon Q Developer CLI 已迁到 Kiro CLI：https://docs.aws.amazon.com/amazonq/latest/qdeveloper-ug/command-line.html

## 1. 摘要

Kiro 是 AWS 的统一 agent harness（IDE / CLI / Web / Mobile）。项目配置在 `.kiro/`。
Amazon Q CLI 的继任面是 Kiro CLI，stdagent 用 `kiro` 作为 target 名，不另开 amazonq。

Kiro 同时消费仓库根 `AGENTS.md`（always-on，无 inclusion 模式）和 `.kiro/steering/`。

## 2. 配置路径

| 类别 | 路径 | 说明 |
|---|---|---|
| 根指令 | `AGENTS.md` | 始终加载；CLI 还发现嵌套子目录 `AGENTS.md` |
| 全局 AGENTS | `~/.kiro/steering/AGENTS.md` | 用户级 |
| Steering | `.kiro/steering/*.md` | inclusion: always / fileMatch / auto / manual |
| 全局 steering | `~/.kiro/steering/` | 工作区优先 |
| Skills | `.kiro/skills/<name>/SKILL.md` | Agent Skills 标准；可 `/name` |
| 全局 skills | `~/.kiro/skills/` | 工作区同名优先 |
| Custom agents | `.kiro/agents/<name>.md` 或 `.json` | 项目级；全局 `~/.kiro/agents/` |
| MCP / hooks / powers | `.kiro/` 下另有配置 | stdagent 不写 |

## 3. Steering frontmatter

| inclusion | 条件 | 额外字段 |
|---|---|---|
| `always` | 源 `alwaysApply: true` | 无 |
| `fileMatch` | 源 `applyTo` 非空 | `fileMatchPattern` 字符串或数组 |
| `auto` | 仅有 `description` | `name` + `description` |
| `manual` | 以上都不满足 | 无；聊天 `#name` 或 slash |

Kiro CLI 当前不支持 inclusion 模式，会加载 `.kiro/steering/` 下全部文件。因此 references 不得写入该目录。

## 4. Skills

官方字段：`name`（必填，等于目录名，≤64）、`description`（必填，≤1024）、`license`、`compatibility`、`metadata`。

## 5. Custom agents

Markdown 与 JSON 等价。stdagent 写 `.kiro/agents/<name>.md`，frontmatter `name` / `description` / `model` / `tools`，正文作为 prompt。

## 6. std-agent 映射

| std-agent 类型 | Kiro 落点 |
|---|---|
| rules（root） | 共享 `AGENTS.md` |
| rules（nested） | `<path>/AGENTS.md` |
| rules（nonRoot） | `.kiro/steering/<name>.md` + inclusion |
| skills | `.kiro/skills/<name>/SKILL.md` |
| commands | `.kiro/skills/commands/<name>/SKILL.md` |
| subagents | `.kiro/agents/<name>.md` |
| references | `.kiro/references/<name>.md` |

## 7. 来源

- https://kiro.dev/docs/steering/
- https://kiro.dev/docs/skills/
- https://kiro.dev/docs/custom-agents/
- https://docs.aws.amazon.com/amazonq/latest/qdeveloper-ug/command-line.html

## 8. UNKNOWN

- `.amazonq/rules` 在仅使用 Kiro CLI 时是否仍被读取
- custom agent Markdown 的 `prompt` 字段与正文是否互斥
- Web / Mobile 是否加载项目 `.kiro/agents/`
