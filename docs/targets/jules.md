# Target: Jules (Google)

调研日期: 2026-05-17，2026-07-10 复核更新
官方主页: https://jules.google
公司: Google

## 1. 摘要

Jules 是 Google 的异步编码 agent（与 Antigravity / Gemini CLI 同属 Google AI 编码产品矩阵），
2025 年 Google I/O 公开预览，2025 年下半年正式上线。Jules 主要以 cloud-hosted 异步执行
环境运行（GitHub 集成 + VM 内沙箱），同时提供 Jules CLI 本地工具。

Jules 已原生消费根 `AGENTS.md`（与 codex / antigravity / amp 同一事实标准）。无 GitHub
star 数（闭源 SaaS），用户基数走 Google AI Studio 渠道。

协议归属：协议族 A（AgentsMD）。与 codex / amp / warp / crush / factory / pi 等共用
`AgentsMD` Protocol，差异通过 adapter 字段表达。

2026-07 复核结论：本 target **无需改动**（第一轮调研已核实"仍只读根 AGENTS.md"结论一致），
但第二轮调研把 Jules 归入"嵌套 AGENTS.md **不支持**"分组（官方文档只提及根目录，未提嵌套），
比第一轮"推断是"的措辞更谨慎，标 UNKNOWN，详见 # 8。

## 2. 配置文件路径

| 类别 | 路径 | 说明 |
|---|---|---|
| 项目级 AGENTS | 项目根 `AGENTS.md` | 官方文档只明确提及根目录消费；嵌套子目录 AGENTS.md 是否被读取未证实（见 # 8） |
| 用户级 | Google 账户绑定，非文件配置 | 通过 jules.google web console / CLI 登录态管理 |
| 仓库 GitHub 集成 | issue / PR 描述 | Jules 异步消费 issue 描述并产生 PR |

Jules 当前**不约定**独立的 rules / skills / commands / references / subagents 子目录。
所有项目级指令统一进 `AGENTS.md`；扩展能力（CLI 工具调用）由 Jules CLI 内置，非项目级
可编辑。

## 3. 文件格式与 frontmatter

- 文件格式：纯 Markdown（`.md`），UTF-8
- frontmatter：否（与 codex / amp / warp 一致，AGENTS.md 是裸 markdown）
- 字节限制：未公开明确上限；官方无数值文档，`budget.go` 未对 `jules` target 设任何
  Hard 限额（仅套用 `*` 通用 rule/skill/command 软建议），2026-07 复核明确"不设 Hard"
  分组包含 jules（与 warp / crush / pi / factory / amp / qwen-code / kilo-code /
  opencode / continue-dev / roo-code / cline / aider 同组）

## 4. skills / commands / references / subagents 原生支持

| std-agent 类型 | Jules 原生 | std-agent 落点 |
|---|---|---|
| rules | YES（AGENTS.md 全量） | `AGENTS.md`（含 root + 内联 nonRoot） |
| skills | NO | `.jules/rules/skills/<name>/SKILL.md`（含 explainer + std-agent-type） |
| commands | NO | `.jules/rules/commands/<name>.md`（fallback rule） |
| references | NO | `.jules/rules/references/<name>.md`（fallback rule） |
| subagents | NO | `.jules/rules/subagents/<name>.md`（fallback rule） |

所有 fallback 文件 body 头部注入 HTML 注释 explainer，frontmatter 含 `std-agent-type:` 字段
标识原 std-agent 类型，便于 Jules 后端识别（也便于其他读 `.jules/rules/` 目录的工具复用）。

## 5. stdagent 落点（julesAdapter，`internal/transformer/jules.go`）

- RootFileName：`AGENTS.md`
- ManifestSection：`Reference Rules`（与 codex 一致）
- NestedSupported：true（子目录 root 写到 `<sub>/AGENTS.md`）。**注意**：这是
  transformer 的乐观默认行为，第二轮调研把 Jules 归入"嵌套 AGENTS.md 不支持"分组
  （官方文档只说 root，未提嵌套），该字段是否应改 false 留待 Jules 官方文档补充
  嵌套语义后再定（当前未改代码，仅记录该矛盾，见 # 8）
- RulesDir：空（nonRoot rules 全 inline 到 AGENTS.md）
- SkillsAsRule：false（RulesDir 为空时 SkillsAsRule=true 会把 skill 写到仓库根；
  改走 BuildDegradedSkillPackage 把 skill 落到 `.jules/rules/skills/<name>/SKILL.md`）
- CommandsDir / ReferencesDir / SubagentsDir：全空，走 FallbackDir
- FallbackDir：`.jules/rules`
- InjectExplainer：true
- InjectStdaiTypeField：true
- InjectTypeGlossary：true（AGENTS.md 顶部注入 std-agent 类型速查段）

## 6. 与 codex / amp / antigravity 的关系

| 维度 | codex | amp | antigravity | jules |
|---|---|---|---|---|
| 根文件 | AGENTS.md | AGENTS.md | 复用 codex 写的 AGENTS.md | AGENTS.md |
| RulesDir | 空（全 inline，`.codex/memories` 已废弃不再使用） | 空（全 inline） | `.agents/rules` | 空（全 inline） |
| SkillsDir | `.agents/skills` | 空 | `.agents/skills` | 空 |
| frontmatter trigger | 否 | 否 | trigger（windsurf 风格） | 否 |
| FallbackDir | `.agents` | `.amp/rules` | `.agents/rules` | `.jules/rules` |

Jules 与 amp 配置最接近（都是 all-inline AGENTS.md + 私有 fallback 目录）；与 codex
差别在于 Jules 无独立 SkillsDir 约定。

## 7. 信息来源

- https://jules.google（Google 官方主页）
- AGENTS.md 事实标准：https://agentsmd.online（Linux Foundation AAIF 2026 托管）

## 8. 已确认与剩余 UNKNOWN（2026-07-10 复核）

已确认：
- Jules 无需改动，仍只读根 `AGENTS.md`（第一轮 + 第二轮调研一致）
- `budget.go` 明确不为 jules 设任何 Hard 限额

剩余 UNKNOWN（2026-07 复核仍未证实，且比第一轮措辞更谨慎）：
- Jules 是否消费嵌套子目录 `AGENTS.md`：第一轮曾"推断是"（与 codex / amp 类推），
  第二轮把 Jules 归入"不支持"分组但仍标 UNKNOWN（官方文档只说 root，未提嵌套，两种
  可能都没有实证）。当前 `julesAdapter.NestedSupported=true` 是乐观默认，尚未因此
  UNKNOWN 结论改动
- Jules CLI 工具是否预留项目级配置目录（如 `.jules/` 下其他文件）：当前调研仅见
  AGENTS.md 一项；未来若官方公布则需扩展 adapter
- 单 AGENTS.md 字节上限（推断与 Gemini API 一致即 1M token；保守按软指导处理，无 Hard）
