# 跨平台转换规则

本文档汇总 std-ai 的 std 格式（YAML frontmatter + Markdown）到 11 个 target
（9 个 Tier 1 + 2 个 Tier 2）的转换决策矩阵。每个 target 的具体细节见
[targets/](targets/) 子目录。

完整的概念差异、主入口策略、降级处理见 [spec.md](spec.md) Part 3 与 Part 4。

11 列总览矩阵也在 [spec.md](spec.md) 第 3.1 节维护（更宽更全）。下方表格保留
Tier 1 9 列以保证可读性；Tier 2（continue-dev / antigravity）落点：

| 概念 | continue-dev | antigravity |
|---|---|---|
| rules（无 applyTo） | `.continue/rules/<n>.md` | `.agents/rules/<n>.md` (always_on) |
| rules（有 applyTo） | `.continue/rules/<n>.md` (globs) | `.agents/rules/<n>.md` (glob+globs) |
| skills | 降级 model_decision rule -> `.continue/rules/skill-<n>.md` | 降级 model_decision rule -> `.agents/rules/skill-<n>.md` |
| commands | `.continue/prompts/<n>.prompt.md` (invokable: true) | `.agents/workflows/<n>.md` |
| references | rule 内嵌 / `docs:` 索引 | rule 正文 `@<file>` 引用 |
| MCP | 跳过（config.yaml 用户级） | 跳过（用户级 mcp_config.json） |

## 1. 总览矩阵：std type x target

行 = std type，列 = target；单元 = 输出路径或处理策略。

| std type \\ target | claude-code | codex | cursor | copilot | windsurf | gemini | aider | cline | opencode |
|---|---|---|---|---|---|---|---|---|---|
| **rules** (无 applyTo) | `.claude/rules/<n>.md` | 拼接到 `AGENTS.md` | `.cursor/rules/<n>.mdc` (Always/AgentReq) | `.github/copilot-instructions.md` 拼接 | `.windsurf/rules/<n>.md` (always_on) | 拼接到 `GEMINI.md` | `read:` 引用 `AGENTS.md` | `.clinerules/<NNN>-<n>.md` | `AGENTS.md` 复用 |
| **rules** (有 applyTo) | `.claude/rules/<n>.md` (frontmatter applyTo) | `<sub>/AGENTS.md` | `.cursor/rules/<n>.mdc` (globs) | `.github/instructions/<n>.instructions.md` (applyTo) | `.windsurf/rules/<n>.md` (glob+globs) | `<sub>/GEMINI.md` | 同 rules | `.clinerules/<NNN>-<n>.md` (paths) | `AGENTS.md` 复用（applyTo 信息丢弃，OpenCode 无条件激活） |
| **skills** | `.claude/skills/<n>/SKILL.md` | `.agents/skills/<n>/SKILL.md` | `.cursor/skills/<n>/SKILL.md` | `.github/agents/<n>.agent.md` | `.windsurf/skills/<n>/SKILL.md` | 降级为 commands `prompt` | 不支持（降级为 rules） | `.clinerules/workflows/<n>.md` | `.opencode/agents/<n>.md` |
| **commands** | `.claude/commands/<n>.md` | `.agents/skills/cmd-<n>/SKILL.md`（降级 skill） | `.cursor/commands/<n>.md` | `.github/prompts/<n>.prompt.md` | `.windsurf/workflows/<n>.md` | `.gemini/commands/<n>.toml` | `--load` 启动脚本（v1.0 不生成） | `.clinerules/workflows/<n>.md` | `.opencode/commands/<n>.md` |
| **references** | `@<path>` import | 不主动写 | 不主动写 | 不主动写 | 不主动写 | `@<path>` import | `read:` 引用 | `memory-bank/<file>.md`（命中 6 文件之一） | `@<path>` 变量替换 |

## 2. 主入口文件

| 文件 | 写入者 | 内容 | 消费方 |
|---|---|---|---|
| `CLAUDE.md` | claude-code transformer | 固定模板 + 索引 + footer | Claude Code |
| `AGENTS.md` | codex transformer（主） | 拼接 rules + footer | Codex / Cursor (fallback) / Copilot Coding Agent / OpenCode / Aider (经 read:) / Windsurf |
| `GEMINI.md` | gemini transformer | 拼接 rules + 子目录 import | Gemini CLI |
| `.windsurfrules` | 不写（legacy） | - | - |
| `.clinerules/` 目录 | cline transformer | 多文件 | Cline |
| `opencode.json` | 不主动写 | - | - |

## 3. Frontmatter 字段映射

std frontmatter -> target 文件 frontmatter（按 target 列出）：

### Claude Code skill (`.claude/skills/<n>/SKILL.md`)

| std | target |
|---|---|
| `name` | `name` |
| `description` | `description` |
| `applyTo` | `paths` |
| `model` | `model` |
| `allowed_tools` | `tools` |

### Claude Code command (`.claude/commands/<n>.md`)

| std | target |
|---|---|
| `description` | `description` |
| `argument_hint` | `argument-hint` |
| `allowed_tools` | `allowed-tools` |
| `model` | `model` |

### Cursor MDC (`.cursor/rules/<n>.mdc`)

| std | target |
|---|---|
| `description` | `description` |
| `applyTo: ["**/*.ts"]` | `globs: "**/*.ts"`（逗号分隔） |
| `alwaysApply` | `alwaysApply` |

激活模式由 `(alwaysApply, applyTo, description)` 组合推导：

| 输入 | mode | target frontmatter |
|---|---|---|
| `alwaysApply: true` | Always | `alwaysApply: true` |
| `applyTo` 非空 | Auto-Attached | `globs: ...` |
| `applyTo` 空, `description` 非空 | Agent Requested | `description: ...` |
| 其他 | Manual | （空 frontmatter） |

### Cursor Skill (`.cursor/skills/<n>/SKILL.md`)

| std | target |
|---|---|
| `name` | `name`（须与文件夹名一致） |
| `description` | `description` |
| `applyTo` | `paths` |
| `disable_model_invocation` | `disable-model-invocation` |

### Copilot instructions (`.github/instructions/<n>.instructions.md`)

| std | target |
|---|---|
| `applyTo: ["**/*.ts", "**/*.tsx"]` | `applyTo: '**/*.ts,**/*.tsx'`（逗号分隔字符串） |
| `exclude_targets` 含 `coding-agent` | `excludeAgent: cloud-agent` |

### Copilot prompt (`.github/prompts/<n>.prompt.md`)

| std | target |
|---|---|
| `description` | `description` |
| `argument_hint` | `argument-hint` |
| `allowed_tools` | `tools` |
| `model` | `model` |

### Windsurf rule (`.windsurf/rules/<n>.md`)

由 `(alwaysApply, applyTo, description)` 推导 trigger：

| 输入 | trigger | 附加字段 |
|---|---|---|
| `alwaysApply: true` | `always_on` | - |
| `applyTo` 非空 | `glob` | `globs: <list>` |
| `applyTo` 空, `description` 非空 | `model_decision` | `description: ...` |
| 其他 | `manual` | - |

### Cline rule (`.clinerules/<NNN>-<n>.md`)

| std | target |
|---|---|
| `applyTo` | `paths` |

NNN 排序前缀由 `priority` 决定：`high=100`、`normal=500`、`low=900`，
余按字母排序。

### OpenCode rules（无条件激活）

OpenCode AGENTS.md / rules 不支持 frontmatter 条件激活。std `applyTo` 字段在
OpenCode 转换时被丢弃；如需多文件 rules 同时加载，writer 可选生成 `opencode.json`
的 `instructions` 字段（v1.0 不主动写，需 `init --opencode` 显式开启）：

```json
{ "instructions": ["AGENTS.md", "docs/*.md", "packages/*/AGENTS.md"] }
```

### OpenCode command / agent (`.opencode/commands/<n>.md` / `.opencode/agents/<n>.md`)

| std | target (command) | target (agent) |
|---|---|---|
| `description` | `description` | `description` |
| `model` | `model` | `model` |
| - | - | `mode: subagent` |
| `allowed_tools` | UNKNOWN | 推导 `permission.*` 三态（v1.0 默认 ask） |

## 4. 主入口文件拼接策略

### `AGENTS.md`（主）

```markdown
<!-- Generated by stdagent vX.Y.Z. Do not edit by hand. Source: .stdai/standards/ -->

# Project AGENTS Manifest

## Coding Style
<rule body>

## Git Commits
<rule body>

## Rules Reference

- Coding Style: rules/coding-style.md
- Git Commits: rules/git-commits.md
- ...

<!-- /Generated by stdagent -->
```

字节超 32768 时拆出超出部分到 `.codex/rules/<n>.md`，AGENTS.md 末尾追加
"Rules Reference" 链接菜单。

### `CLAUDE.md`（主）

```markdown
<!-- Generated by stdagent vX.Y.Z. Do not edit by hand. -->

# Project CLAUDE Manifest

@.claude/rules/coding-style.md
@.claude/rules/git-commits.md

<!-- /Generated by stdagent -->
```

只放 import 不放正文，rules 实际内容在 `.claude/rules/` 子文件中。

### `GEMINI.md`（主）

类似 AGENTS.md 拼接策略；子目录 rules 落到对应子目录的 `GEMINI.md`。

### `.github/copilot-instructions.md`（主）

```markdown
<!-- Generated by stdagent vX.Y.Z. -->

<rule body>

<rule body>

<!-- /Generated by stdagent -->
```

无 frontmatter；按 `priority` -> `name` 顺序拼接所有无 `applyTo` 的 rules。

### `.windsurf/rules/<n>.md`

每条 rule 一个独立文件，trigger 由 std frontmatter 推导。

## 5. Footer 注入

当 `inject = true`：

```html
<!--
Generated by stdagent vX.Y.Z at 2026-05-07T17:30:00Z.
Source: .stdai/standards/<source-relative-path>
Run `stdagent sync` to regenerate.
Do not edit this file by hand.
-->
```

当 `inject_whatis = true`，附加：

```markdown

---

## About This File

This file is part of `stdagent` synchronized AI standards.

- **Edit source**: `.stdai/standards/<source-relative-path>`
- **Regenerate**: `stdagent sync`
- **Disable target**: set `[targets].<name>.enabled = false` in `.stdai/config.toml`
- **Documentation**: https://github.com/StringKe/std-ai
```

## 6. 冲突与备份

每次 sync 前对将被覆盖的扩散文件做快照到 `.stdai/backups/<RFC3339>/`。
保留份数由 `backup_keep` 控制。

冲突检测策略（用户手写 vs stdagent 生成）：

1. 检测目标文件首尾的 stdagent marker（如 `<!-- Generated by stdagent`）
2. 若不存在 marker（说明用户手写）：备份后写入，并 stderr 输出 WARN
3. 若存在 marker：直接覆盖，不输出 WARN

## 7. 跳过策略

以下情况跳过写入：

- target `enabled = false`
- std frontmatter `targets` 不含该 target
- std frontmatter `exclude_targets` 含该 target
- std type 在该 target 上无对应映射（如 commands -> codex）
- 输出与现有文件 checksum 一致（dry-run 也不输出 diff）

## 7.5 MCP 转换

`.stdai/standards/mcp.json` 单 JSON 源 -> 三个 target 的项目级 MCP 配置：

| target | 输出路径 | 顶级键 |
|---|---|---|
| claude-code | `.mcp.json` | `mcpServers` |
| cursor | `.cursor/mcp.json` | `mcpServers` |
| copilot | `.vscode/mcp.json` | `servers`（注意与上两者不同） |

其他 target 不分发 MCP（用户级配置 / settings.json 风险 / 不支持）。

实现：runner 在 parse 之后加载 `.stdai/standards/mcp.json` 到 `cfg.MCP`，
3 个 transformer Plan 函数检测 `cfg.MCP != nil` 时 append 对应 mcp.json
FileOp 到 plan。MCP 分发块前置在 docs early return 之前，确保仅 mcp 而无
其他 standards 时也能输出。

## 8. raw copy 模式（convert = false）

当 `[targets].<name>.convert = false` 时：

- 跳过 frontmatter 字段映射
- 直接复制 std 文件正文（保留原 frontmatter）到 target 的 rules 默认目录
- 仍执行 footer 注入（除非 `inject = false`）
- 适用：用户希望"跨工具尽量原样保留"的场景
