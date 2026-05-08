# Target: OpenCode

调研日期: 2026-05-07
官方文档: https://opencode.ai/docs/

## 1. 摘要

OpenCode 是开源 CLI（sst/opencode），配置主入口是 `opencode.json`，规则用
`AGENTS.md`，命令与代理走 `.opencode/{commands,agents}/` 目录。

OpenCode 把 `AGENTS.md` 当一等公民并兼容 `~/.claude/CLAUDE.md`（向后兼容入口）。
当 AGENTS.md 与 CLAUDE.md 同时存在时**仅 AGENTS.md 生效**。

OpenCode rules **不支持** frontmatter 条件激活（无类似 Cursor `globs` 或
Cline `paths` 的字段）。拆分多文件的官方机制是 `opencode.json` 的
`instructions` 字段（路径数组 + glob 模式），但激活仍是全局的。

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
| `.opencode/commands/<name>.md` | Markdown + frontmatter | `description`、`agent`、`model`、`subtask`（强制以子代理执行）、`template` |
| `.opencode/agents/<name>.md` | Markdown + frontmatter | `mode`（`primary`/`subagent`）、`description`、`tools`、`model`、`permission.{edit,bash,read,glob,grep,list,task,lsp}` 三态 |

变量替换：`{env:VAR}` `{file:path}` `$ARGUMENTS` `$1..$N` `` !`bash` `` `@filename`

## 4. AGENTS.md 加载顺序

```
1. <cwd> 向上递归查找最近的 AGENTS.md
2. fallback: ~/.config/opencode/AGENTS.md
3. fallback: ~/.claude/CLAUDE.md（向后兼容；与 AGENTS.md 同存时被忽略）
```

"第一个匹配文件获胜"，AGENTS.md 与 CLAUDE.md 同时存在时**仅 AGENTS.md 生效**。

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

## 7. std-ai 四类映射

| std-ai 类型 | OpenCode 落点 |
|---|---|
| rules（无 applyTo） | `<repo>/AGENTS.md`（自动消费，由 codex transformer 已写） |
| rules（有 applyTo） | 同上；OpenCode 无条件激活，applyTo 信息会被丢弃。可选额外写入 `opencode.json` 的 `instructions` 数组让多个 rule 文件被同时加载 |
| skills | `.opencode/agents/<name>.md`（mode: subagent，由 primary 或 `@mention` 调度） |
| commands | `.opencode/commands/<name>.md`（自动注册 `/<filename>` 触发） |
| references | `@filename` 内联 / `{file:path}` 变量替换 |

## 8. 转换器实现要点

1. AGENTS.md 已由 codex transformer 生成；OpenCode 直接复用
2. commands 转 `.opencode/commands/<name>.md`：
   - std `description` -> `description`
   - std `model` -> `model`
   - std `allowed_tools` -> （UNKNOWN，OpenCode command frontmatter 是否支持 `tools`）
   - 正文 + footer 作为 prompt template
3. skills 转 `.opencode/agents/<name>.md`：
   - `mode: subagent`
   - `description` 必填（沿用 std description）
   - `permission.*` 由 std `allowed_tools` 推断（v1.0 简化为全 ask）
4. v1.0 不写 `opencode.json`；如需用户启用 `instructions` 字段，在 `init --opencode`
   显式开关时生成
5. `/init` 与 stdagent 协作策略：用户首次跑 `/init`，OpenCode 智能改进 AGENTS.md；
   stdagent 后续 sync 时检测 marker 决定是否覆盖

## 9. 信息来源

- https://opencode.ai/docs/config/
- https://opencode.ai/docs/rules/
- https://opencode.ai/docs/commands/
- https://opencode.ai/docs/agents/

## 10. 已确认与剩余 UNKNOWN

已确认：
- AGENTS.md / rules 不支持 frontmatter 条件激活
- 拆分多文件的机制：`opencode.json` 的 `instructions` 字段（glob 数组），但仍全局激活
- 无 `.opencode/rules/` 约定目录
- agent permission 完整字段（edit/bash/read/glob/grep/list/task/lsp 三态 allow/ask/deny）

剩余 UNKNOWN：
- agent frontmatter 的 `tools` 字段是否存在（permission 已确认存在）
- command frontmatter 是否支持 `argument-hint`
- 项目级 ignore 文件
