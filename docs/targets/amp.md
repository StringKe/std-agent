# Target: Sourcegraph Amp

调研日期: 2026-05-17
官方主页: https://ampcode.com
官方文档: https://ampcode.com/manual

## 1. 摘要

Sourcegraph Amp 是 Sourcegraph 推出的 AI 编码助手（前身 Cody Agent / Cody
CLI 体系演进），主打 agentic coding，深度集成代码搜索与跨仓库上下文。
2026 年初公开发布，定位与 Cursor / Claude Code 同台竞争。

Amp 的项目配置主入口是仓库根的 `AGENTS.md`（早期版本叫 `AGENT.md`，已迁
移到 `AGENTS.md`，旧文件保留向后兼容）。Amp 与 Cursor 是 AGENTS.md 标准
的核心推动者之一，文档明确读取顶层 AGENTS.md 并支持多文件 + 嵌套子目录
AGENTS.md 自动叠加，作用域按目录深度决定。

Amp 不消费 frontmatter，AGENTS.md 是纯 Markdown 指令文本；也没有原生
skill / command / subagent 扩展机制，所有结构化内容须以章节形式 inline
进 AGENTS.md。
(https://ampcode.com/manual, 访问日期 2026-05-17)

## 2. 配置文件路径

| 类型 | 路径 | 说明 |
|---|---|---|
| 全局指令 | `~/.config/AGENT.md` | 用户级全局规则（Amp 自带读取） |
| 项目根 | `<repo>/AGENTS.md` | 主入口（原 `AGENT.md` 已迁移） |
| 兼容旧名 | `<repo>/AGENT.md` | 向后兼容，仍读，不建议新建 |
| 嵌套子目录 | `<repo>/<subdir>/AGENTS.md` | 进入子目录时自动叠加上下文 |

Amp 无 `.amp/` 或 `.sourcegraph/` 私有目录用于 rule fan-out；所有项目级
约束都集中在 AGENTS.md。
(https://ampcode.com/manual, 访问日期 2026-05-17;
https://agents.md, 访问日期 2026-05-17)

## 3. 文件格式

| 文件 | 格式 | frontmatter |
|---|---|---|
| `AGENTS.md` / `AGENT.md` | Markdown | 无，纯指令文本 |
| 嵌套 `AGENTS.md` | Markdown | 无 |
| `~/.config/AGENT.md` | Markdown | 无 |

无 trigger / globs / applyTo 等条件加载机制。AGENTS.md 是 always-on，
嵌套版本按工作目录自动追加。

## 4. std-ai 五类映射

| std-ai 类型 | Amp 落点 | 加载方式 |
|---|---|---|
| rules | 内联到顶层 `AGENTS.md`（root rule body + nonRoot rule 段落） | 总是注入 |
| skills | `.amp/rules/skills/<name>/SKILL.md`（std-ai 私有降级目录，Amp 不读取，仅供其他工具或 AI 在显式提示时检索） | 不自动加载，AI 通过 explainer 注释理解 |
| commands | `.amp/rules/commands/<name>.md` | 不自动加载 |
| references | `.amp/rules/references/<name>.md` | 不自动加载 |
| subagents | `.amp/rules/subagents/<name>.md` | 不自动加载 |

Amp 只原生消费 AGENTS.md；后四类 std-ai 类型在 amp transformer 下全
部走 `.amp/rules/` 子目录降级，frontmatter 写 `std-ai-type` 私有字段，
body 头部插 HTML 注释 explainer 说明该类型语义与"为何不在 AGENTS.md 主
体"。

## 5. 转换器实现要点

1. 主输出：项目根 `AGENTS.md`（标准 stdagent generated marker），含
   glossary 头部 + root rule body + nonRoot rule inline 段落
2. `RulesDir=""`：amp 没有官方 rule 子目录，nonRoot rules 全部 inline 到
   `AGENTS.md` 而非 fan-out 到 `.amp/rules/`
3. `SkillsAsRule=false`：留空时 skills 降级走 Agent Skills 标准包形态写到
   `.amp/rules/skills/<name>/SKILL.md`，避免污染仓库根
4. `FallbackDir=".amp/rules"`：commands / references / subagents 进
   `.amp/rules/<type>/<name>.md`，frontmatter 写 `std-ai-type`
5. `InjectExplainer=true`：fallback 文件 body 头部加 HTML 注释说明该类型
   语义，方便 AI / 人类读懂
6. `InjectTypeGlossary=true`：AGENTS.md 头部注入 std-ai 五类速查段
7. `NestedSupported=true`：子目录 `<subdir>/AGENTS.md` 由 root rule
   `nested_path` 字段触发，复用 codex 风格嵌套 root（无 manifest，无
   glossary）
8. MCP：Amp 暂无公开的 MCP 客户端配置文档，本 transformer 不写 MCP

## 6. 信息来源

- https://ampcode.com （访问日期 2026-05-17）
- https://ampcode.com/manual （访问日期 2026-05-17）
- https://agents.md （访问日期 2026-05-17）

## 7. 已确认

- Amp 读取顶层 `AGENTS.md`，旧文件名 `AGENT.md` 已迁移并向后兼容
- 全局指令位于 `~/.config/AGENT.md`
- 嵌套子目录 `AGENTS.md` 自动叠加，按目录作用域
- 无 frontmatter / 无 trigger / 无 globs / 无 applyTo
- 无原生 skills / commands / subagents 机制
- Sourcegraph 与 Cursor 并列推动 AGENTS.md 跨工具标准

## 8. UNKNOWN

- Amp 是否计划支持 frontmatter 条件加载（trigger / globs）
- Amp 的 MCP 客户端配置路径与格式（公开文档未列出）
- 是否存在企业版 site-level config 与项目 AGENTS.md 的冲突解析规则
