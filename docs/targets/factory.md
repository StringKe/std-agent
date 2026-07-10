# Target: Factory (Factory.ai Droids)

调研日期: 2026-05-17
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

Factory 用 "droid" 词代替业界通用的 "subagent / sub-agent"：
`.factory/droids/<n>.md` 是单文件 markdown，frontmatter `name` /
`description` / `model` / `tools`，body 是 system prompt，与 Claude
Code `.claude/agents/` 形态一致。

## 2. 配置文件路径

| 类别 | 路径 | 说明 |
|---|---|---|
| 全局 AGENTS | `~/.factory/AGENTS.md` | 跨项目共享规则 |
| 项目 AGENTS | `<repo>/AGENTS.md` | 项目入口，nearest wins |
| 嵌套 AGENTS | `<repo>/<subdir>/AGENTS.md` | 子目录自动叠加 |
| 项目 Rules | `<repo>/.factory/rules/*.md` | 细粒度规则，无 glob frontmatter |
| 项目 Skills | `<repo>/.factory/skills/<n>/SKILL.md` | Agent Skills 标准包 |
| 项目 Droids | `<repo>/.factory/droids/<n>.md` | subagent 单文件 |
| 项目设置 | `<repo>/.factory/settings.json` | MCP / 模型 / 工具白名单 |

## 3. 文件格式

| 文件 | 格式 | frontmatter |
|---|---|---|
| `AGENTS.md` | Markdown | 无（纯指令文本） |
| `.factory/rules/*.md` | Markdown | 可选 `description`，**无 globs / paths / applyTo** |
| `.factory/skills/<n>/SKILL.md` | Markdown | Agent Skills 标准：`name` / `description` / `license` / `compatibility` / `metadata` |
| `.factory/droids/<n>.md` | Markdown | `name` / `description` / `model` / `tools` |
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
| commands | 无原生 commands 落点，graceful degradation 到 `.factory/rules/cmd-<n>.md` | 走 fallback rule |
| references | 无原生 references 落点，degradation 到 SkillsDir（Agent Skills 标准） | 同 skills 加载 |
| subagents | `.factory/droids/<n>.md`，frontmatter `name` / `description` / `model` / `tools` | Factory 显式调度 |

## 5. 转换器实现要点

1. 主输出 `AGENTS.md`：root rules + nonRoot manifest 段 `## Reference Rules`，
   inject footer 由 stdagent 注入
2. nonRoot rules fan-out 到 `.factory/rules/<n>.md`，frontmatter
   **不**渲染 `globs` / `paths` / `applyTo`（`adapter.GlobsFieldName=""`）；
   description 写入 frontmatter
3. skills 走 Agent Skills 标准 `<.factory/skills/<n>/SKILL.md>`，
   `SkillSupportedFields = name / description / license / compatibility / metadata`
4. subagents 走 `.factory/droids/<n>.md`（`SubagentsDir=".factory/droids"`），
   frontmatter `name` / `description` / `model` / `tools`
5. commands / references 留空 `CommandsDir` / `ReferencesDir`，由
   AgentsMD 协议自动走 graceful degradation 到 `FallbackDir`
   （`.factory/rules`）
6. v0.0.4 不写 `.factory/settings.json`：MCP 走 stdagent 现有
   `.mcp.json` 体系，settings.json 是后续 phase 扩展位

## 6. 信息来源

- https://docs.factory.ai/cli/configuration/overview （访问日期 2026-05-17）
- https://docs.factory.ai/droids/overview （访问日期 2026-05-17）
- https://docs.factory.ai/agents-md （访问日期 2026-05-17）
- /tmp/std-agent-protocol-research.md（行 31，调研日期 2026-05-17）

## 7. 已确认

- AGENTS.md 是 Factory 项目入口；全局 `~/.factory/AGENTS.md` 与项目
  根叠加，nearest wins，与 codex / amp 行为一致
- `.factory/rules/*.md` 无 glob frontmatter 支持；rule 激活由 droid
  在运行时基于内容匹配，非文件模式触发
- `.factory/skills/<n>/SKILL.md` 遵循 Agent Skills 标准包结构
- `.factory/droids/<n>.md` 是 subagent 落点，单文件 markdown +
  frontmatter，形态与 Claude Code `.claude/agents/` 一致
- Factory 在企业付费市场可见度高，AGENTS.md 标准核心采纳方之一

## 8. UNKNOWN

- `.factory/settings.json` 完整 schema（已知含 MCP / 模型配置，字段
  细节未在公开文档列出）
- droids 之间互相调用的语义（是否支持 droid 嵌套调用 droid）
- `.factory/rules/` 是否支持优先级数字前缀（公开文档未提及）
