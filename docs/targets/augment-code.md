# Target: Augment Code

调研日期: 2026-05-17
官方文档: https://docs.augmentcode.com/
公司: Augment（augmentcode.com）

## 1. 摘要

Augment Code 是企业向 AI 编程平台，闭源 IDE 扩展（VS Code / JetBrains），主打企业代码库
上下文检索与团队共享 Memories。配置以 `.augment/` 子目录为主，rules 用 markdown +
私有 `type` frontmatter（`always_apply` / `agent_requested` / `manual`），老版兼容
`.augment-guidelines` 单文件。同时也读 `AGENTS.md` 与 `CLAUDE.md`（生态兼容）。

## 2. 配置文件路径

| 类别 | 路径 | 说明 |
|---|---|---|
| 工作区 rules | `.augment/rules/*.md` | 每文件一条，按 `type` frontmatter 决定触发模式 |
| 老版 rules（兼容） | `.augment-guidelines` | 项目根单文件，所有内容 always-on，无 frontmatter |
| 兼容读取 | `AGENTS.md` / `CLAUDE.md` | 项目根级别，作为 always-on 上下文 |
| Code Review | `.augment/code_review_guidelines.yaml` | Augment Code Review 功能专用，独立格式 |

## 3. Rules frontmatter（私有方言）

| 字段 | 类型 | 取值 |
|---|---|---|
| `type` | enum | `always_apply` / `agent_requested` / `manual` |
| `description` | string | `agent_requested` 必填，作为模型决策触发的依据 |

触发语义对照表（与 Windsurf 几乎一一对应）：

| Augment `type` | Windsurf `trigger` | 行为 |
|---|---|---|
| `always_apply` | `always_on` | 全文进 system prompt |
| `agent_requested` | `model_decision` | description 进 system prompt，按需读全文 |
| `manual` | `manual` | 仅 @mention 触发 |

Augment 没有等同 Windsurf `glob` 的 `applyTo` 触发；带 glob 的 std rule 在 Augment 中
回退为 `always_apply`（保守生效）。

## 4. std-ai 四类映射（v0.0.4）

| std-ai 类型 | Augment 落点 | 说明 |
|---|---|---|
| rules（无 applyTo） | `.augment/rules/<name>.md` `trigger: always_on` | 通过 WindsurfStyle 协议族近似映射 |
| rules（有 applyTo） | `.augment/rules/<name>.md` `trigger: glob` `globs:` | Augment 不识别 `trigger`，会忽略未知字段 |
| rules（仅 description） | `.augment/rules/<name>.md` `trigger: model_decision` | 等同 Augment `agent_requested` |
| rules（无任何条件） | `.augment/rules/<name>.md` `trigger: manual` | 等同 Augment `manual` |
| skills | `.augment/rules/skills/<name>/SKILL.md` | Agent Skills 标准包 fallback（Augment 无原生 skills） |
| commands | `.augment/rules/workflows/<name>.md` | 子目录名借用 windsurf 风格，纯文档 |
| references | `.augment/rules/references/<name>.md` | graceful degradation + `std-ai-type: references` |
| subagents | `.augment/rules/subagents/<name>.md` | graceful degradation + `std-ai-type: subagents` |

## 5. 转换器实现要点

1. 复用 `protocol.WindsurfStyle`：trigger 字段语义近似 Augment 的 type 字段
2. v0.0.4 暂不写严格的 Augment 私有 `type` frontmatter；Augment 容忍未知 YAML 键
3. 老版 `.augment-guidelines` 单文件 fallback：v0.0.4 保留 adapter 字段（SingleFileFallback）
   不主动产出，等用户实测后再决定是否启用
4. AGENTS.md 与 CLAUDE.md 由 codex / claude-code transformer 写根目录，augment-code
   自动消费，不重复写
5. MCP：暂未确认 Augment 的项目级 MCP 配置路径，v0.0.4 不输出
6. Code Review YAML：脱离 std-ai 四类语义，超出 v0.0.4 范围

## 6. 已知限制

- v0.0.4 用 WindsurfStyle 协议族近似映射；Augment 私有 `type` 字段（always_apply /
  agent_requested / manual）未严格对齐输出 `trigger`（always_on / model_decision /
  manual）。Augment 容忍未知字段，但若期望严格对齐，需在 v0.0.5+ 引入独立的
  AugmentStyle 协议
- glob 触发回退为 always_apply 语义；Augment 自身无 glob 触发
- skills / references / subagents 均走 fallback，Augment IDE 不会自动识别这些类型，
  仅作为 rule 内容被加载

## 7. 信息来源

- /tmp/std-ai-protocol-research.md（2026-05-17 调研）
- https://augmentcode.com/
- https://docs.augmentcode.com/
- 协议族归类: spec.md §2.4 / plan.md §Phase 5.9
