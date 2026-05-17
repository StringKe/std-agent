# Target: Jules (Google)

调研日期: 2026-05-17
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

## 2. 配置文件路径

| 类别 | 路径 | 说明 |
|---|---|---|
| 项目级 AGENTS | 项目根 `AGENTS.md` | Jules 与子目录嵌套 AGENTS.md 合并加载（沿用 AGENTS.md 事实标准） |
| 用户级 | Google 账户绑定，非文件配置 | 通过 jules.google web console / CLI 登录态管理 |
| 仓库 GitHub 集成 | issue / PR 描述 | Jules 异步消费 issue 描述并产生 PR |

Jules 当前**不约定**独立的 rules / skills / commands / references / subagents 子目录。
所有项目级指令统一进 `AGENTS.md`；扩展能力（CLI 工具调用）由 Jules CLI 内置，非项目级
可编辑。

## 3. 文件格式与 frontmatter

- 文件格式：纯 Markdown（`.md`），UTF-8
- frontmatter：否（与 codex / amp / warp 一致，AGENTS.md 是裸 markdown）
- 字节限制：未公开明确上限。Jules 后端基于 Gemini，单次 prompt 上下文 1M token，
  AGENTS.md 实测 32KB 内安全（与 codex `project_doc_max_bytes` 默认值同档）。
  stdagent v0.0.4 不为 jules 强制字节限制

## 4. skills / commands / references / subagents 原生支持

| std-ai 类型 | Jules 原生 | std-ai 落点 |
|---|---|---|
| rules | YES（AGENTS.md 全量） | `AGENTS.md`（含 root + 内联 nonRoot） |
| skills | NO | `.jules/rules/skills/<name>/SKILL.md`（含 explainer + std-ai-type） |
| commands | NO | `.jules/rules/commands/<name>.md`（fallback rule） |
| references | NO | `.jules/rules/references/<name>.md`（fallback rule） |
| subagents | NO | `.jules/rules/subagents/<name>.md`（fallback rule） |

所有 fallback 文件 body 头部注入 HTML 注释 explainer，frontmatter 含 `std-ai-type:` 字段
标识原 std-ai 类型，便于 Jules 后端识别（也便于其他读 `.jules/rules/` 目录的工具复用）。

## 5. stdagent 落点（julesAdapter）

- RootFileName：`AGENTS.md`
- ManifestSection：`Reference Rules`（与 codex 一致）
- NestedSupported：true（子目录 root 写到 `<sub>/AGENTS.md`）
- RulesDir：空（nonRoot rules 全 inline 到 AGENTS.md）
- SkillsAsRule：false（RulesDir 为空时 SkillsAsRule=true 会把 skill 写到仓库根；
  改走 BuildDegradedSkillPackage 把 skill 落到 `.jules/rules/skills/<name>/SKILL.md`）
- CommandsDir / ReferencesDir / SubagentsDir：全空，走 FallbackDir
- FallbackDir：`.jules/rules`
- InjectExplainer：true
- InjectStdaiTypeField：true
- InjectTypeGlossary：true（AGENTS.md 顶部注入 std-ai 类型速查段）

## 6. 与 codex / amp / antigravity 的关系

| 维度 | codex | amp | antigravity | jules |
|---|---|---|---|---|
| 根文件 | AGENTS.md | AGENTS.md | 复用 codex 写的 AGENTS.md | AGENTS.md |
| RulesDir | `.codex/memories` | 空（全 inline） | `.agents/rules` | 空（全 inline） |
| SkillsDir | `.agents/skills` | 空 | 无（SkillsAsRule） | 空 |
| frontmatter trigger | 否 | 否 | trigger（windsurf 风格） | 否 |
| FallbackDir | `.codex/memories` | `.amp/rules` | `.agents/rules` | `.jules/rules` |

Jules 与 amp 配置最接近（都是 all-inline AGENTS.md + 私有 fallback 目录）；与 codex
差别在于 Jules 无独立 SkillsDir / `.codex/memories` 等私有目录约定。

## 7. 信息来源

- https://jules.google（Google 官方主页）
- /tmp/std-ai-protocol-research.md §1 row 20 + §2.A
- AGENTS.md 事实标准：https://agentsmd.online（Linux Foundation AAIF 2026 托管）

## 8. UNKNOWN

- Jules 是否消费嵌套子目录 AGENTS.md（推断"是"，与 codex / amp 一致，未在官方 doc 明示）
- Jules CLI 工具是否预留项目级配置目录（如 `.jules/` 下其他文件）— 当前调研仅见
  AGENTS.md 一项；未来若官方公布则需扩展 adapter
- 单 AGENTS.md 字节上限（推断与 Gemini API 一致即 1M token；保守按 32KB 处理）
