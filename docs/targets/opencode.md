# Target: OpenCode

调研日期: 2026-05-07（2026-07-10 更新：skills 原生 GA、目录复数/单数兼容、嵌套 AGENTS.md 动态注入回填；2026-08-15 复核路径未变）
官方文档: https://opencode.ai/docs/

## 1. 摘要

OpenCode 是开源 CLI（sst/opencode），配置主入口是 `opencode.json`，规则用
`AGENTS.md`，命令与代理走 `.opencode/{commands,agents}/` 目录。

OpenCode 把 `AGENTS.md` 当一等公民并兼容 `~/.claude/CLAUDE.md`（向后兼容入口）。
当 AGENTS.md 与 CLAUDE.md **同目录**同时存在时**仅 AGENTS.md 生效**（同目录 AGENTS.md
优先于 CLAUDE.md，2026-07-10 确认，见 §4.1）。

OpenCode rules **不支持** frontmatter 条件激活（无类似 Cursor `globs` 或
Cline `paths` 的字段）。拆分多文件的官方机制是 `opencode.json` 的
`instructions` 字段（路径数组 + glob 模式），但激活仍是全局的。

Agent Skills **已原生 GA**：官方标准包路径 `.opencode/skills/<name>/SKILL.md`
（https://open-code.ai/en/docs/skills）。2026-05-07 版本记录的"skill 降级为
`mode: subagent` 单文件 agent"方案已随官方 GA 废弃，transformer 已改为直写
`.opencode/skills/<name>/SKILL.md`（见 §7）。

## 2. 配置文件路径

| 类型 | 路径 | 优先级（数字越大越高） |
|---|---|---|
| 远程组织默认 | `.well-known/opencode` | 1 |
| 用户全局 | `~/.config/opencode/opencode.json` | 2 |
| 自定义路径 | `$OPENCODE_CONFIG` 指向的文件 | 3 |
| 项目级 | `<repo>/opencode.json` | 4 |
| `.opencode/` 目录 | agents / commands / plugins / tools / themes | 4 |
| 内联 | `$OPENCODE_CONFIG_CONTENT` 环境变量 | 6 |
| 系统管理 | 管理员策略 | 7 |
| macOS MDM | 托管偏好 | 8（最高） |

配置**按 key 合并**而非整体替换。

## 3. 文件格式

| 文件 | 格式 | frontmatter |
|---|---|---|
| `opencode.json` / `opencode.jsonc` | JSON | 无 |
| `AGENTS.md` | Markdown | 无 |
| `.opencode/skills/<name>/SKILL.md` | Agent Skills 标准包 | `name` / `description`（官方标准字段，2026-07-10 新增，见 https://open-code.ai/en/docs/skills） |
| `.opencode/commands/<name>.md`（主，复数） | Markdown + frontmatter | `description`、`agent`、`model`、`subtask`（强制以子代理执行）、`template` |
| `.opencode/agents/<name>.md`（主，复数） | Markdown + frontmatter | `mode`（`primary`/`subagent`）、`description`、`tools`、`model`、`permission.{edit,bash,read,glob,grep,list,task,lsp}` 三态 |

2026-07-10 确认：`commands` / `agents` 目录名以**复数为主**，单数 `.opencode/command/` /
`.opencode/agent/` 保留向后兼容（不建议新写入用单数）。

变量替换：`{env:VAR}` `{file:path}` `$ARGUMENTS` `$1..$N` `` !`bash` `` `@filename`

## 4. AGENTS.md 加载顺序

```
1. <cwd> 向上递归查找最近的 AGENTS.md
2. fallback: ~/.config/opencode/AGENTS.md
3. fallback: ~/.claude/CLAUDE.md（向后兼容；与 AGENTS.md 同存时被忽略）
```

"第一个匹配文件获胜"，同目录 AGENTS.md 与 CLAUDE.md 同时存在时**仅 AGENTS.md 生效**。

## 4.1 嵌套 AGENTS.md 动态注入（2026-07-10 新增）

OpenCode 在 `read` 工具读取某子目录下的文件时，会**动态注入**沿途各级目录的 AGENTS.md
（read 触发式，非启动时一次性全量加载）。stdagent 侧 `opencodeAdapter` 设
`NestedSupported=false`：这不代表 OpenCode 不支持嵌套 AGENTS.md，而是嵌套文件本身由
启用的 AGENTS.md producer 计划由 runner canonicalize 后写入 `x/y/AGENTS.md`，
opencode transformer 不重复写，避免同一嵌套文件
被两个 transformer 各写一次。

## 5. agent permission 三态

```yaml
permission:
  edit: ask     # allow / ask / deny
  bash: deny
  read: allow
  glob: allow
  grep: allow
  list: allow
  task: ask
  lsp: allow
```

## 6. opencode.json instructions 字段

OpenCode rules 不支持 frontmatter 条件激活。拆分多文件的官方机制是
`opencode.json` 的 `instructions` 字段，接受路径数组 + glob 模式：

```json
{
  "instructions": [
    "AGENTS.md",
    "CONTRIBUTING.md",
    "docs/guidelines.md",
    ".cursor/rules/*.md",
    "packages/*/AGENTS.md"
  ]
}
```

激活仍是**全局**的，没有按文件路径条件激活的能力。Monorepo 推荐用
`packages/*/AGENTS.md` 这类 glob，比手工在 AGENTS.md 内引用更可维护。

## 7. std-agent 四类映射（2026-07-10 更新，与 `internal/transformer/opencode.go` 一致）

| std-agent 类型 | OpenCode 落点 |
|---|---|
| rules（无 applyTo） | 不落盘（`opencodeAdapter.RulesDir` 为空，由 shared `AGENTS.md` 承担） |
| rules（有 applyTo） | 同上；OpenCode 无条件激活，applyTo 信息会被丢弃。可选额外写入 `opencode.json` 的 `instructions` 数组让多个 rule 文件被同时加载（v1.0 未实现，见 §8） |
| skills | `.opencode/skills/<name>/SKILL.md`（原生 Agent Skills 标准包；旧的 `mode: subagent` 降级方案已废弃） |
| commands | `.opencode/commands/<name>.md`（自动注册 `/<filename>` 触发） |
| references | `.opencode/references/<name>.md`（fallback，`FallbackDir=".opencode"`） |

## 8. 转换器实现要点（2026-07-10 更新）

1. OpenCode transformer 不写 AGENTS.md；复用其他启用 producer 经 runner canonicalize 的共享文件
2. commands 转 `.opencode/commands/<name>.md`：
   - std `description` -> `description`
   - std `model` -> `model`
   - std `allowed_tools` -> （UNKNOWN，OpenCode command frontmatter 是否支持 `tools`）
   - 正文 + footer 作为 prompt template
3. skills 直写 `.opencode/skills/<name>/SKILL.md`（原生 Agent Skills 标准包，`BuildNativeSkillPackage`），
   废弃旧版"降级为 `mode: subagent` 单文件 agent"整套逻辑
4. 不写 `opencode.json`；如需用户启用 `instructions` 字段，留待后续版本显式开关时生成
5. `/init` 与 stdagent 协作策略：用户首次跑 `/init`，OpenCode 智能改进 AGENTS.md；
   stdagent 后续 sync 时检测 marker 决定是否覆盖

## 9. 信息来源

- https://opencode.ai/docs/config/
- https://opencode.ai/docs/rules/
- https://opencode.ai/docs/commands/
- https://opencode.ai/docs/agents/
- https://open-code.ai/en/docs/skills（2026-07-10 新增，skills 原生 GA 确认）

## 10. 已确认与剩余 UNKNOWN

已确认：
- AGENTS.md / rules 不支持 frontmatter 条件激活
- 拆分多文件的机制：`opencode.json` 的 `instructions` 字段（glob 数组），但仍全局激活
- 无 `.opencode/rules/` 约定目录
- agent permission 完整字段（edit/bash/read/glob/grep/list/task/lsp 三态 allow/ask/deny）
- 2026-07-10：skills 原生 GA，`.opencode/skills/<name>/SKILL.md`；旧降级方案（`mode: subagent`）已废弃
- 2026-07-10：commands / agents 目录复数为主、单数向后兼容
- 2026-07-10：嵌套 AGENTS.md 由 read 工具动态注入触发，非启动时全量加载；同目录 AGENTS.md
  优先于 CLAUDE.md

剩余 UNKNOWN：
- agent frontmatter 的 `tools` 字段是否存在（permission 已确认存在）
- command frontmatter 是否支持 `argument-hint`
- 项目级 ignore 文件
