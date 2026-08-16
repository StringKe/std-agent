# Target: Factory (Factory.ai Droids)

调研日期: 2026-05-17（2026-07-10 补充调研）
官方文档: https://docs.factory.ai
官方主页: https://factory.ai

## 1. 摘要

Factory.ai 是面向企业的 AI 软件交付平台，主推 Droids（自治软件工程
agent）。配置体系绕 `AGENTS.md` 标准展开，是 Linux Foundation AAIF
旗下 60k+ 仓库采纳 AGENTS.md 协议的核心商业方之一。Factory 在企业
付费市场可见度高，定位与 Cursor / Cognition 同维度。

读取路径：项目根 `AGENTS.md`（与全局 `~/.factory/AGENTS.md` 叠加，
nearest wins），项目级 `.factory/rules/*.md`（rules，无 glob 支持），
`.factory/skills/<n>/SKILL.md`（Agent Skills 标准包），
`.factory/droids/<n>.md`（subagent / "droid"），
`.factory/settings.json`（MCP / 模型偏好）。

commands 与 skills 官方是"共存"而非"合并"关系：skills 现在也可通过 slash
命令调用（`.factory/skills/<n>/SKILL.md` 同时支持 `/<n>` 手动触发与模型自动
检索），但 legacy `.factory/commands/*.md` 仍原样继续工作，未被强制迁移
（"Your existing `.factory/commands/` files continue to work unchanged"，
https://docs.factory.ai/cli/configuration/custom-slash-commands）。stdagent
选择零成本迁移路径：commands 类型继续走 `.factory/commands/` 原生目录，不
转写为 skill。

Factory 用 "droid" 词代替业界通用的 "subagent / sub-agent"：
`.factory/droids/<n>.md` 是单文件 markdown，官方 frontmatter 字段为
`name`（必填）/ `description` / `model` / `tools` / `reasoningEffort` /
`mcpServers`（https://docs.factory.ai/cli/configuration/custom-droids）。
`reasoningEffort`（`low`/`medium`/`high`）与 `mcpServers`（限定该 droid 可用
的 MCP server 名单）是 2026-07 复核新确认的字段，stdagent 当前 subagent
frontmatter 仍只输出 `name` / `description` / `model` / `tools`（parser.Document
schema 未承载 reasoningEffort/mcpServers，暂不产出，见 UNKNOWN）。官方文档同时
确认 **droid 之间不能相互调用**（subagent 不可用 `Task` 工具做递归委派）。

## 2. 配置文件路径

| 类别 | 路径 | 说明 |
|---|---|---|
| 全局 AGENTS | `~/.factory/AGENTS.md` | 跨项目共享规则 |
| 项目 AGENTS | `<repo>/AGENTS.md` | 项目入口，nearest wins |
| 嵌套 AGENTS | `<repo>/<subdir>/AGENTS.md` | 子目录自动叠加 |
| 项目 Rules | `<repo>/.factory/rules/*.md` | 细粒度规则，无 glob frontmatter |
| 项目 Skills | `<repo>/.factory/skills/<n>/SKILL.md` | Agent Skills 标准包，同时可 `/<n>` 调用 |
| 项目 Commands（legacy，仍读） | `<repo>/.factory/commands/<n>.md` | slash 命令，官方明确继续兼容 |
| 项目 Droids | `<repo>/.factory/droids/<n>.md` | subagent 单文件 |
| 项目设置 | `<repo>/.factory/settings.json` | MCP / 模型 / 工具白名单 |

## 3. 文件格式

| 文件 | 格式 | frontmatter |
|---|---|---|
| `AGENTS.md` | Markdown | 无（纯指令文本） |
| `.factory/rules/*.md` | Markdown | 可选 `description`，**无 globs / paths / applyTo** |
| `.factory/skills/<n>/SKILL.md` | Markdown | Agent Skills 标准 + `allowed-tools` / `user-invocable` / `disable-model-invocation` |
| `.factory/commands/<n>.md` | Markdown | 可选 frontmatter（legacy 兼容路径） |
| `.factory/droids/<n>.md` | Markdown | `name`（必填）/ `description`（≤500 字符）/ `model` / `tools` / `reasoningEffort` / `mcpServers` |
| `.factory/settings.json` | JSON | 不适用 |

Rules 文件**不支持 glob frontmatter**：Factory 把 rules 视为"always-on
project guidance"，激活由 droid 自行决定（基于 description 内容
匹配），不按文件模式触发。glob-style 选择交由 droid 配置或 skill 的
`compatibility` 字段表达。

## 4. std-agent 四类映射

| std-agent 类型 | Factory 落点 | 加载方式 |
|---|---|---|
| rules | `.factory/rules/<n>.md`（frontmatter 仅 `description`），多条 nonRoot 由 AGENTS.md manifest 段 `## Reference Rules` 索引 | Factory 启动时全量扫描 |
| skills | `.factory/skills/<n>/SKILL.md` + 同目录辅助文件 | Agent Skills 标准发现 |
| commands | `.factory/commands/<n>.md`（legacy 原生目录，官方持续兼容） | Factory 启动时扫描，`/<n>` 调用 |
| references | `.factory/references/<n>.md`（不进 `.factory/rules/`） | 按需查阅 |
| subagents | `.factory/droids/<n>.md`，frontmatter `name` / `description` / `model` / `tools` | Factory 显式调度，droid 之间不能互相调用 |

## 5. 转换器实现要点

1. 主输出 `AGENTS.md`：root rules + nonRoot manifest 段 `## Reference Rules`，
   inject footer 由 stdagent 注入
2. nonRoot rules fan-out 到 `.factory/rules/<n>.md`，frontmatter
   **不**渲染 `globs` / `paths` / `applyTo`（`adapter.GlobsFieldName=""`）；
   description 写入 frontmatter
3. skills 走 Agent Skills 标准 `<.factory/skills/<n>/SKILL.md>`，
   字段含 `name` / `description` / `license` / `compatibility` / `metadata` /
   `allowed-tools` / `user-invocable` / `disable-model-invocation`
4. subagents 走 `.factory/droids/<n>.md`（`SubagentsDir=".factory/droids"`），
   frontmatter 目前仅渲染 `name` / `description` / `model` / `tools`
   （协议层 subagent frontmatter 字段集固定，未扩展 `reasoningEffort` /
   `mcpServers`，见 UNKNOWN）
5. commands 走 `.factory/commands/<n>.md`（`CommandsDir=".factory/commands"`），
   对齐官方 legacy 目录持续兼容的结论，不再降级到 skills 或 rules fallback
6. references 留空 `ReferencesDir`，由 AgentsMD 协议自动走 graceful
   degradation 到 `FallbackDir`（`.factory/rules`）
7. v0.0.4 不写 `.factory/settings.json`：MCP 走 stdagent 现有
   `.mcp.json` 体系，settings.json 是后续 phase 扩展位

## 6. 信息来源

- https://docs.factory.ai/cli/configuration/overview （访问日期 2026-05-17）
- https://docs.factory.ai/droids/overview （访问日期 2026-05-17）
- https://docs.factory.ai/agents-md （访问日期 2026-05-17）
- https://docs.factory.ai/cli/configuration/custom-slash-commands （commands 与
  skills 共存关系，legacy 目录继续兼容，2026-07-10 复核）
- https://docs.factory.ai/cli/configuration/custom-droids （droid frontmatter 全字段：
  `name` / `description` / `model` / `tools` / `reasoningEffort` / `mcpServers`，
  droid 不可互相调用，2026-07-10 新增）

## 7. 已确认

- AGENTS.md 是 Factory 项目入口；全局 `~/.factory/AGENTS.md` 与项目
  根叠加，nearest wins，与 codex / amp 行为一致
- `.factory/rules/*.md` 无 glob frontmatter 支持；rule 激活由 droid
  在运行时基于内容匹配，非文件模式触发
- `.factory/skills/<n>/SKILL.md` 遵循 Agent Skills 标准包结构，同时可作 slash
  命令调用
- `.factory/commands/<n>.md` legacy 目录官方持续兼容，未被 skills 取代
- `.factory/droids/<n>.md` 是 subagent 落点，完整官方字段为 `name` / `description`
  / `model` / `tools` / `reasoningEffort` / `mcpServers`
- droid 之间不能相互调用（无递归委派）
- Factory 在企业付费市场可见度高，AGENTS.md 标准核心采纳方之一
- 初始 AGENTS 类指南上限 80,000 字符；动态 Read 路径发现 40,000 字符

## 8. UNKNOWN

- `.factory/settings.json` 完整 schema（已知含 MCP / 模型配置，字段
  细节未在公开文档列出）
- stdagent 是否应扩展 subagent frontmatter 支持 `reasoningEffort` / `mcpServers`：
  需要 parser.Document schema 新增对应字段并评估其余 target（如 claude-code
  subagent 的 `model`）复用可能性，本轮未实现，留待后续 plan
- `.factory/rules/` 是否支持优先级数字前缀（公开文档未提及）
