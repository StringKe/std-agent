# Target: Cline

调研日期: 2026-05-07（2026-07-10 更新：skills GA、AGENTS.md 消费行为回填；2026-08-15 迁移到官方推荐 skills 路径）
官方文档: https://docs.cline.bot/

## 1. 摘要

Cline（VS Code 扩展，原 Claude Dev）使用 `.clinerules/` 目录承载多个规则文件，
自动合并为统一 rules。支持 frontmatter `paths:` 进行 glob 条件激活。

Agent Skills 已 GA：官方推荐路径 `.cline/skills/<name>/SKILL.md`，同时保留
`.clinerules/skills/` 与 `.claude/skills/` 作为扫描兼容路径
（https://docs.cline.bot/customization/skills）。transformer 现写推荐路径
`.cline/skills/`。

Cline **消费根 `AGENTS.md`**（项目根一份）与全局 `~/.agents/AGENTS.md`，**不读嵌套子目录**
的 AGENTS.md（仅 workspace root 生效）。旧结论"不消费 AGENTS.md"已过时并纠正。

Memory Bank 是文档约定 + 提示词模式（6 个文件名社区标准），不是代码层强制结构。

## 2. 配置文件路径

| 类型 | 路径 | 说明 |
|---|---|---|
| 项目级 rules | `<repo>/.clinerules/` 目录 | 内含多个 .md / .txt |
| 全局 rules（macOS/Linux） | `~/Documents/Cline/Rules/` | OS 级别 |
| 全局 rules（Windows） | `Documents\Cline\Rules\` | OS 级别 |
| 项目级 workflows | `<repo>/.clinerules/workflows/` | 自定义 slash 命令源 |
| skills（推荐，transformer 落点） | `<repo>/.cline/skills/<name>/SKILL.md` | 官方 Agent Skills 标准包 |
| skills（备用扫描） | `<repo>/.clinerules/skills/`、`<repo>/.claude/skills/` | Cline 仍会发现，stdagent 不再写入 |
| 根文件（消费） | `<repo>/AGENTS.md` | 项目根一份，Cline 自动读取，不读嵌套子目录 |
| 全局根文件（消费） | `~/.agents/AGENTS.md` | 全局默认上下文 |
| Memory Bank | `<repo>/memory-bank/`（约定） | 由 .clinerules 触发 |
| IDE 设置 | VS Code Cline 扩展面板（API key、Custom Instructions 输入框） | 非文件配置 |

## 3. 文件格式与 frontmatter

| 文件 | 扩展名 | frontmatter |
|---|---|---|
| 规则 | `.md` 或 `.txt` | 可选 YAML，关键字段 `paths: ["src/**"]`（条件激活） |
| workflow | `.md` | 无强制；文件名作为命令名（`release.md` -> `/release.md`） |

文档**未提及**单文件 `.clinerules` 的兼容性，目录形式是当前主推。

## 4. 加载机制

- `.clinerules/` 下所有 `.md` `.txt` **自动加载并拼接**为统一 rules
- Rules 面板提供单条开关（toggle）启用 / 禁用每个文件
- 带 `paths:` frontmatter 的规则**仅在当前文件路径匹配 glob 时激活**
- 工作区规则与全局规则冲突时，**工作区优先**
- workflow：不自动执行，需用户输入 `/<filename>` 触发
- 数字前缀（`01-coding.md` `02-testing.md`）用于显式排序

## 5. 内置 slash 命令清单

`/newtask` `/smol`（别名 `/compact`）`/newrule` `/deep-planning`
`/explain-changes`（仅 VS Code）`/reportbug`。

### 5.1 /newrule 行为

`/newrule` 不静默默认到 workspace 或 global，而是在执行时弹出 scope 选择：

| 选项 | 保存位置 |
|---|---|
| Workspace | `<repo>/.clinerules/` |
| Global | `~/Documents/Cline/Rules/`（macOS/Linux）；`Documents\Cline\Rules`（Windows）；备用 `~/Cline/Rules/`（Linux/WSL） |

工程协作场景推荐 Workspace（团队规范、project-specific、可版本控制）。

## 6. Memory Bank 6 文件（社区约定）

```
memory-bank/
├── projectbrief.md
├── productContext.md
├── activeContext.md
├── systemPatterns.md
├── techContext.md
└── progress.md
```

由 `.clinerules` 中指令引导 Cline 读取。Cline 自身不强制此结构。

## 7. std-agent 四类映射

| std-agent 类型 | Cline 落点 |
|---|---|
| rules | `.clinerules/<NN>-<name>.md`（数字前缀控制顺序）；frontmatter `paths:` 来自 std `applyTo` |
| skills | `.cline/skills/<name>/SKILL.md`（官方推荐原生包） |
| commands | `.clinerules/workflows/<name>.md` |
| references | `.cline/references/<name>.md`（不进 `.clinerules/`，避免被当 rule 全量加载） |

## 8. 转换器实现要点

1. std rule 转 `.clinerules/<NNN>-<name>.md`：
   - NNN 由 std `priority` 决定排序（high=100、normal=500、low=900，余按字母）
   - frontmatter `paths:` 来自 std `applyTo`
2. workflows 与 commands 共用目录 `.clinerules/workflows/`
3. skills 走原生 Agent Skills 包，落 `.cline/skills/<name>/SKILL.md`
   （`clineAdapter.SkillsDir`）
4. AGENTS.md / CLAUDE.md 不由本 transformer 写；Cline 可消费启用 producer 经
   runner canonicalize 的共享 AGENTS.md
5. 全局 rules（用户级）目录 v1.0 不主动写入；保留 v1.1
6. references 走私有 `.cline/references/`，不写进 `.clinerules/`

## 9. 信息来源

- https://docs.cline.bot/features/cline-rules
- https://docs.cline.bot/prompting/cline-memory-bank
- https://docs.cline.bot/features/slash-commands/workflows
- https://docs.cline.bot/features/commands-and-shortcuts/overview
- https://docs.cline.bot/customization/skills
- https://docs.cline.bot/customization/cline-rules

## 10. 已确认与剩余 UNKNOWN

已确认：
- `/newrule` 弹出 scope 选择，无静默默认值（见 # 5.1）
- 全局 Rules 路径在 macOS/Linux/Windows/WSL 各 OS 上的具体位置
- 2026-07-10：Agent Skills 已 GA，推荐路径 `.cline/skills/`，`.clinerules/skills/` 为官方保留的
  备用扫描路径
- 2026-08-15：transformer 已迁移到 `.cline/skills/`；规则 `paths` 条件激活、根 `AGENTS.md`
  消费行为与 2026-07 一致
- 2026-07-10：Cline 消费项目根 `AGENTS.md` 与全局 `~/.agents/AGENTS.md`，不读嵌套子目录
  AGENTS.md（原"不消费 AGENTS.md"结论已纠正）

剩余 UNKNOWN：
- 单文件 `.clinerules` 是否仍兼容（社区指南推断"被目录形式取代但向后兼容"）
- 是否有项目级 ignore 文件
