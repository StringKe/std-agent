# Target: Sourcegraph Amp

调研日期: 2026-05-17（2026-07-10 补充调研）
官方主页: https://ampcode.com
官方文档: https://ampcode.com/manual

## 1. 摘要

Sourcegraph Amp 是 Sourcegraph 推出的 AI 编码助手（前身 Cody Agent / Cody
CLI 体系演进），主打 agentic coding，深度集成代码搜索与跨仓库上下文。
2026 年初公开发布，定位与 Cursor / Claude Code 同台竞争。

Amp 的项目配置主入口是仓库根的 `AGENTS.md`（早期版本叫 `AGENT.md`，已迁
移到 `AGENTS.md`，旧文件保留向后兼容）。Amp 与 Cursor 是 AGENTS.md 标准
的核心推动者之一，文档明确读取顶层 AGENTS.md 并支持多文件 + 嵌套子目录
AGENTS.md 自动叠加，作用域按目录深度决定。AGENTS.md 支持 `@path` /
`@glob` 引用其他文件（如 `@doc/*.md`、`@~/some/path`），被引用文件可在自身
frontmatter 写 `globs` 字段限定仅在命中匹配文件时才注入，未写 `globs` 则始终
随引用生效（https://ampcode.com/manual）。

Amp 已于 2026-01-29 官方移除自定义 slash commands 功能，原有能力并入
Agent Skills。Neo（2026-05-06）后不再提供 user-invokable slash；skill 由模型
按 description 决定是否加载。skills 走原生目录 `.agents/skills/`。

Amp 不消费 rules frontmatter，AGENTS.md 是纯 Markdown 指令文本；也没有原生
subagent 文件化定义机制，subagent 是运行时动态生成的 mini-Amp（Task tool），
无法用文件预置。

## 2. 配置文件路径

| 类型 | 路径 | 说明 |
|---|---|---|
| 全局指令 | `~/.config/amp/AGENTS.md`（或 `~/.config/AGENTS.md`） | 用户级全局规则，两个路径都会被读取并叠加 |
| 项目根 | `<repo>/AGENTS.md` | 主入口（原 `AGENT.md` 已迁移） |
| 兼容旧名 | `<repo>/AGENT.md` | 向后兼容，仍读，不建议新建 |
| 嵌套子目录 | `<repo>/<subdir>/AGENTS.md` | 进入子目录时自动叠加上下文 |
| 项目 Skills | `<repo>/.agents/skills/<name>/SKILL.md` | 原生 Agent Skills 标准包；由模型按 description 选用 |

Amp 无独立 `.amp/` 或 `.sourcegraph/` 私有目录用于 rule fan-out；rules 仍集中
写在 AGENTS.md，但 skills（含原 commands 语义）已有原生目录承载。
(https://ampcode.com/manual；https://agents.md；https://ampcode.com/manual/agent-skills.md)

## 3. 文件格式

| 文件 | 格式 | frontmatter |
|---|---|---|
| `AGENTS.md` / `AGENT.md` | Markdown | 无，纯指令文本 |
| 嵌套 `AGENTS.md` | Markdown | 无 |
| `~/.config/amp/AGENTS.md` | Markdown | 无 |
| `.agents/skills/<name>/SKILL.md` | Markdown | Agent Skills 标准字段：`name` / `description` / `license` / `compatibility` / `metadata` |

AGENTS.md 无 trigger / applyTo 等条件加载机制，是 always-on，嵌套版本按工作
目录自动追加；仅 `@path` / `@glob` 引用文件可携带 `globs` frontmatter 做条件注入。

## 4. std-agent 五类映射

| std-agent 类型 | Amp 落点 | 加载方式 |
|---|---|---|
| rules | 内联到顶层 `AGENTS.md`（root rule body + nonRoot rule 段落） | 总是注入 |
| skills | `.agents/skills/<name>/SKILL.md`（原生 Agent Skills 标准包） | 模型按 description 自动检索 + `/<name>` 手动调用 |
| commands | `.agents/skills/commands/<name>/SKILL.md`（与 codex 相同的降级形态：commands 转写为 skill，落 `commands` 子目录） | 同 skills，`/<name>` 手动调用 |
| references | `.amp/rules/references/<name>.md`（fallback） | 不自动加载 |
| subagents | `.amp/rules/subagents/<name>.md`（fallback，Amp 无文件化 subagent） | 不自动加载 |

skills 与 commands 落点跟 codex transformer 完全一致（同一 `.agents/skills/`
命名空间、同一 `SkillSupportedFields` 集合），两个 target 产出字节相同，writer
按 unchanged 去重。references / subagents 仍无原生机制，走 `.amp/rules/`
子目录降级，frontmatter 写 `std-agent-type` 私有字段，body 头部插 HTML 注释
explainer 说明该类型语义。

## 5. 转换器实现要点

1. 主输出：项目根 `AGENTS.md`（标准 stdagent generated marker），含
   glossary 头部 + root rule body + nonRoot rule inline 段落
2. `RulesDir=""`：amp 没有官方 rule 子目录，nonRoot rules 全部 inline 到
   `AGENTS.md` 而非 fan-out 到 `.amp/rules/`
3. `SkillsDir=".agents/skills"`：原生 Agent Skills 标准目录，`SkillSupportedFields`
   为 `name / description / license / compatibility / metadata`
4. `CommandFormat=CommandSkillPrefix` + `CommandsAsSkillSubdir="commands"`：
   commands 降级写为 skill，落 `.agents/skills/commands/<name>/SKILL.md`
   （对齐官方 2026-01-29 commands 并入 skills 的公告）
5. `ReferencesDir=""` `SubagentsDir=""`：references / subagents 无原生落点，
   走 `FallbackDir=".amp/rules"` 子目录 graceful degradation
6. `InjectExplainer=true` / `InjectStdaiTypeField=true`：fallback 文件 body
   头部加 HTML 注释说明该类型语义，frontmatter 写 `std-agent-type`
7. `InjectTypeGlossary=true`：AGENTS.md 头部注入 std-agent 五类速查段
8. `NestedSupported=true`：子目录 `<subdir>/AGENTS.md` 由 root rule
   `nested_path` 字段触发，复用 codex 风格嵌套 root（无 manifest，无
   glossary）
9. MCP：Amp 暂无公开的 MCP 客户端配置文档，本 transformer 不写 MCP

## 6. 信息来源

- https://ampcode.com （访问日期 2026-05-17）
- https://ampcode.com/manual （访问日期 2026-05-17 / 2026-07-10 复核）
- https://agents.md （访问日期 2026-05-17）
- https://ampcode.com/manual/agent-skills.md （原生 Agent Skills 标准，2026-07-10）
- https://ampcode.com/news/slashing-custom-commands （commands 移除并入 skills，2026-01-29 公告，2026-07-10 复核）

## 7. 已确认

- Amp 读取顶层 `AGENTS.md`，旧文件名 `AGENT.md` 已迁移并向后兼容
- 全局指令位于 `~/.config/amp/AGENTS.md`（`~/.config/AGENTS.md` 同样被读取；
  旧调研记录的 `~/.config/AGENT.md` 路径过时，已修正）
- 嵌套子目录 `AGENTS.md` 自动叠加，按目录作用域
- AGENTS.md 无 frontmatter / 无 trigger / 无 globs / 无 applyTo；但支持
  `@path` / `@glob` 动态引用其他文件，被引用文件可用 `globs` frontmatter 做
  条件注入
- 原生 Agent Skills 标准，目录 `.agents/skills/<name>/SKILL.md`
- 自定义 slash commands 已于 2026-01-29 官方移除，能力并入 skills，
  `/<name>` 直接调用同名 skill
- 无原生 subagent 文件化机制（subagent 是运行时动态 Task，非预置文件）
- Sourcegraph 与 Cursor 并列推动 AGENTS.md 跨工具标准

## 8. UNKNOWN

- Amp 是否计划支持 rules frontmatter 条件加载（trigger / globs）
- Amp 的 MCP 客户端配置路径与格式（公开文档未列出）
- 是否存在企业版 site-level config 与项目 AGENTS.md 的冲突解析规则
