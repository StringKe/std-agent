# Target: Augment Code

调研日期: 2026-05-17（2026-07-10 更新：skills 原生 GA、根文件叠加顺序、嵌套最近邻、
字符上限回填）
官方文档: https://docs.augmentcode.com/
公司: Augment（augmentcode.com）

## 1. 摘要

Augment Code 是企业向 AI 编程平台，闭源 IDE 扩展（VS Code / JetBrains），主打企业代码库
上下文检索与团队共享 Memories。配置以 `.augment/` 子目录为主，rules 用 markdown +
私有 `type` frontmatter（`always_apply` / `agent_requested` / `manual`），老版兼容
`.augment-guidelines` 单文件。同时也读 `AGENTS.md` 与 `CLAUDE.md`（生态兼容）。

Agent Skills **已原生支持**：标准包路径 `.augment/skills/<name>/SKILL.md`
（https://docs.augmentcode.com/using-augment/skills）。transformer 已改为直写原生路径，
不再走 `.augment/rules/skills/` fallback。

根文件消费为**全部叠加**而非三选一：`AGENTS.md` / `CLAUDE.md` 均读取，且 **`CLAUDE.md`
排在 `AGENTS.md` 之前**注入（顺序影响上下文优先级，2026-07-10 确认，见 §6.1）。嵌套子目录
采用**最近邻逐层**策略（子目录规则逐层向上合并，与 `.augment/rules/` 就近文件优先，见 §6.1）。

## 2. 配置文件路径

| 类别 | 路径 | 说明 |
|---|---|---|
| 工作区 rules | `.augment/rules/*.md` | 每文件一条，按 `type` frontmatter 决定触发模式 |
| skills | `.augment/skills/<name>/SKILL.md` | 原生 Agent Skills 标准包（Auto/Manual/Disabled 三态） |
| 老版 rules（兼容） | `.augment-guidelines` | 项目根单文件，所有内容 always-on，无 frontmatter |
| 兼容读取 | `AGENTS.md` / `CLAUDE.md` | 项目根级别，全部叠加读取，CLAUDE.md 排在 AGENTS.md 之前 |
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

## 4. std-agent 四类映射（2026-07-10 更新，与 `internal/transformer/augment_code.go` 一致）

| std-agent 类型 | Augment 落点 | 说明 |
|---|---|---|
| rules（无 applyTo） | `.augment/rules/<name>.md` `trigger: always_on` | 通过 WindsurfStyle 协议族近似映射 |
| rules（有 applyTo） | `.augment/rules/<name>.md` `trigger: glob` `globs:` | Augment 不识别 `trigger`，会忽略未知字段 |
| rules（仅 description） | `.augment/rules/<name>.md` `trigger: model_decision` | 等同 Augment `agent_requested` |
| rules（无任何条件） | `.augment/rules/<name>.md` `trigger: manual` | 等同 Augment `manual` |
| skills | `.augment/skills/<name>/SKILL.md` | 原生 Agent Skills 标准包（`SkillsDir=".augment/skills"`，不再走 fallback） |
| commands | `.augment/rules/workflows/<name>.md` | 子目录名借用 windsurf 风格，纯文档（Augment 无原生 commands） |
| references | `.augment/rules/references/<name>.md` | graceful degradation + `std-agent-type: references` |
| subagents | `.augment/rules/subagents/<name>.md` | graceful degradation + `std-agent-type: subagents` |

## 5. 转换器实现要点

1. 复用 `protocol.WindsurfStyle`：trigger 字段语义近似 Augment 的 type 字段
2. 暂不写严格的 Augment 私有 `type` frontmatter；Augment 容忍未知 YAML 键
3. skills 直写原生路径 `.augment/skills/<name>/SKILL.md`（`SkillsAsRule: false`），
   不再走 `.augment/rules/skills/` fallback
4. 老版 `.augment-guidelines` 单文件 fallback：保留 adapter 字段（SingleFileFallback）
   不主动产出，等用户实测后再决定是否启用
5. augment-code 不写 AGENTS.md 与 CLAUDE.md；前者复用启用 producer 经 runner
   canonicalize 的共享文件，后者由 claude-code transformer 生成。两者会叠加消费
6. MCP：暂未确认 Augment 的项目级 MCP 配置路径，不输出
7. Code Review YAML：脱离 std-agent 四类语义，超出范围

## 6. 已知限制

- WindsurfStyle 协议族近似映射；Augment 私有 `type` 字段（always_apply /
  agent_requested / manual）未严格对齐输出 `trigger`（always_on / model_decision /
  manual）。Augment 容忍未知字段，但若期望严格对齐，需引入独立的
  AugmentStyle 协议（v0.0.5+ 候选）
- glob 触发回退为 always_apply 语义；Augment 自身无 glob 触发
- references / subagents 仍走 fallback，Augment IDE 不会自动识别这些类型，
  仅作为 rule 内容被加载（skills 已转为原生，不再属于此限制）

## 6.1 根文件叠加与嵌套目录（2026-07-10 新增）

- **根文件全部叠加**：Augment 同时读取 `AGENTS.md` 与 `CLAUDE.md`（若均存在），不是
  三选一 fallback；**`CLAUDE.md` 排在 `AGENTS.md` 之前**注入上下文，顺序会影响优先级
  权重。stdagent 的 root.md body 若同时写到两个文件，会被 Augment 重复注入两遍
  （与 cursor / qwen-code / grok-build / crush 同属"多路全读"类，见 spec.md §根文件叠加/
  重复注入问题）
- **嵌套目录最近邻逐层**：子目录 rules 按最近邻策略逐层向上合并（与 amp / factory /
  windsurf 同属"完整支持嵌套"一类），非 codex 式链上全量拼接

## 6.2 字符上限（2026-07-10 新增）

Augment Workspace Guidelines（`.augment-guidelines` / 老版单文件）单项字符上限 24576，
Workspace Guidelines + Rules（`.augment/rules/*.md`）合计上限 49512；超限时按优先级
**淘汰**非截断（manual -> always/auto -> guidelines 顺序应用直至限额，超出部分静默弃置，
不是像 Windsurf 那样截断文本）。`internal/budget/budget.go` 已登记
`{"augment-code", "rules-total", 0, 49512, ...}` 校验合计上限；24576 为 Workspace
Guidelines 单文件上限，budget.go 暂未单独校验此项，标记为跟进项。

## 7. 信息来源

- /tmp/std-agent-protocol-research.md（2026-05-17 调研）
- https://augmentcode.com/
- https://docs.augmentcode.com/
- https://docs.augmentcode.com/using-augment/skills（2026-07-10 新增，skills 原生 GA 确认）
- https://docs.augmentcode.com/setup-augment/rules（2026-07-10 新增，根文件叠加顺序 /
  嵌套最近邻 / 字符上限 24576 与 49512 确认）
- 协议族归类: spec.md §2.4 / plan.md §Phase 5.9

## 8. UNKNOWN

- Workspace Guidelines 单文件 24576 上限超限时的具体截断/淘汰边界细节（合计 49512 的
  淘汰顺序已确认，单文件 24576 上限的独立行为未实证）
- Augment 项目级 MCP 配置路径
