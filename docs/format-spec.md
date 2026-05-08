# 源文件格式规范

## 概览

`.stdai/standards/` 下所有文件必须为 **Markdown + YAML Frontmatter**。
Frontmatter 是文件开头被 `---` 包裹的一段 YAML，用于声明元数据。
正文是标准 CommonMark Markdown，将被各 target 转换器消费。

## 文件头模板

```markdown
---
type: rules                # 必填：rules | skills | commands | references | subagents
name: coding-style         # 必填：唯一标识，kebab-case
version: 1.2               # 可选：semver 风格
description: |             # 可选：单行或多行
  General coding style and naming conventions.
targets:                   # 可选：限定生效平台；省略 = 所有 enabled
  - claude-code
  - codex
  - cursor
exclude_targets: []        # 可选：黑名单，与 targets 二选一
priority: high             # 可选：high | normal | low；用于排序与冲突仲裁
tags: [style, security]    # 可选：自由标签
applyTo:                   # 可选：glob 列表；映射到 Cursor globs / Copilot applyTo
  - "**/*.ts"
  - "**/*.tsx"
globs: []                  # 可选：applyTo 别名（rulesync / Cursor / Cline 风格），合并去重
claudecode:                # 可选：target 专属 paths 覆盖（rulesync 风格嵌套字段）
  paths: ["**/*Service.java"]
cursor:
  paths: ["src/**/*.ts"]
alwaysApply: false         # 可选：Cursor 的 always-on 模式
allowed_tools:             # 可选：仅 commands 类有效，映射 Claude allowed-tools
  - Read
  - Bash(git status:*)
argument_hint: ""          # 可选：仅 commands 类，映射 Claude argument-hint
model: sonnet              # 可选：仅 commands/skills/agents 类有效
---

正文 Markdown 从这里开始。
```

## 必填字段

| 字段 | 类型 | 校验 | 说明 |
|---|---|---|---|
| `type` | string | enum: rules/skills/commands/references/subagents | 决定走哪类转换 |
| `name` | string | kebab-case，` ^[a-z0-9][a-z0-9-]*$ ` | 在 type 内全局唯一 |

## 可选元数据

| 字段 | 类型 | 默认 | 说明 |
|---|---|---|---|
| `version` | string | "1.0.0" | semver |
| `description` | string | "" | 1-3 行 |
| `targets` | string[] | [] (= 所有 enabled) | target 白名单 |
| `exclude_targets` | string[] | [] | target 黑名单（与 targets 互斥） |
| `priority` | enum | normal | high/normal/low |
| `tags` | string[] | [] | 任意 |
| `applyTo` | string[] | [] | gitignore 风格 glob，匹配的代码文件触发该规则 |
| `globs` | string[] | [] | `applyTo` 别名（rulesync / Cursor / Cline 业界字段名），合并去重 |
| `<rulesync-target>.paths` | string[] | [] | target 专属 paths 覆盖。key 为 rulesync 风格 target 名（`claudecode` / `codexcli` / `cursor` / `copilot` / `windsurf` / `gemini` / `aider` / `cline` / `opencode`），value 为该 target 用的 glob 列表，覆盖全局 `applyTo`/`globs`。同一文件给不同 target 不同生效路径用 |
| `alwaysApply` | bool | false | Cursor always-on |
| `allowed_tools` | string[] | [] | 仅 commands |
| `argument_hint` | string | "" | 仅 commands |
| `model` | string | "" | 仅 commands/skills/agents |

## 四种 type 的语义

### rules

通用规则文件。会被转换并写入到各 target 的 rules/instructions 位置：

- Claude Code: `.claude/rules/<name>.md`（或合并到 CLAUDE.md，根据策略）
- Codex: `.codex/memories/<name>.md`（codex CLI 自动加载该目录的项目级 memory），AGENTS.md 末尾追加自描述清单
- Cursor: `.cursor/rules/<name>.mdc`
- Copilot: `.github/instructions/<name>.instructions.md`（带 applyTo）
- Windsurf: `.windsurfrules` 单文件追加段
- Cline: `.clinerules/<name>.md`

### skills

可复用能力包。每个 skill 是一个目录，目录内有 `SKILL.md` + 可选辅助文件：

```
.stdai/standards/skills/
├── code-review/
│   ├── SKILL.md            必须有 frontmatter，type=skills
│   ├── checklist.md        辅助参考
│   └── examples/
└── ...
```

- Claude Code: `.claude/skills/<name>/SKILL.md` + 同目录辅助文件
- 其他工具大多无原生 skills 概念，转换为 rules 或 commands fallback

### commands

可触发的 slash command 或自定义 prompt：

- Claude Code: `.claude/commands/<name>.md`（frontmatter 含 description / allowed-tools / argument-hint / model）
- Codex CLI: 暂无项目级自定义命令的官方机制，UNKNOWN，可降级为 rules 提示
- Cursor: 无原生 slash 自定义，降级为 rules
- Copilot: `.github/prompts/<name>.prompt.md`
- Gemini CLI: `.gemini/commands/<name>.toml`（待 targets/gemini-cli.md 核实）

### subagents

Claude Code 原生支持的 subagent（spawnable 子代理，独立 context）。与 SKILL（按需触发的能力包，
main session 内联使用）区别：subagent 是隔离 context 的子进程，由 main session 调 Task 工具触发。

- Claude Code: `.claude/agents/<name>.md`，frontmatter 字段：
  - `name`：kebab-case 标识
  - `description`：模型理解何时调起
  - `model`：可选，指定 subagent 用的模型（如 `claude-sonnet-4-5`）
  - `allowed_tools` -> 输出为 `tools`：subagent 可用的工具白名单
- 其他 target：UNKNOWN / 不输出（codex / cursor / copilot 等无对应原生概念）

源 frontmatter 示例：

```yaml
---
type: subagents
name: code-reviewer
description: Reviews code changes for safety and clarity
model: claude-sonnet-4-5
allowed_tools: [Read, Grep, Bash]
---

You are a strict code reviewer ...
```

### references

参考文档/上下文资料。一般不直接被 AI 加载，作为人类可读的支撑材料；
也可通过 Claude Code 的 `@<path>` import 机制被显式引用。

- 所有 target 默认输出到 `.stdai/standards/references/` 镜像位置或 `docs/`
  下一个独立子目录（路径由 config 控制）
- 不进入 `CLAUDE.md` / `AGENTS.md` 等入口文件

## 根文件主体（root.md 约定）

`.stdai/standards/root.md`（顶层，**不**在 rules/ skills/ 等子目录里）是**项目总览**文件。stdagent sync 时把它的 body 直接写到所有 target 根文件（CLAUDE.md / AGENTS.md / GEMINI.md / .github/copilot-instructions.md）的头部，再在尾部追加自动生成的 rule manifest 段。

约定：

- 文件名不区分大小写（`root.md` / `Root.md` / `ROOT.md` 都识别）
- 不需要 frontmatter（路径就是约定），但写也无害
- root.md 不会 fan-out 成 `.claude/rules/root.md` 等子目录文件（它已是根文件主体）
- root.md 通常含：项目定义 + 模块结构 + 全局铁律 + AI 配置维护流程；**不**应该在里面手写 rule 文件清单（stdagent 自动追加 manifest 段）
- 项目可以**不**写 root.md：此时 stdagent 用 `# Project XXX Manifest` 占位标题作根文件头部

详见 stdagent intro 提示词（`stdagent intro` 命令）。

## targets 字段语义

```yaml
# 写法 1：白名单
targets:
  - claude-code
  - codex

# 写法 2：黑名单
exclude_targets:
  - cursor
```

当一个文件 `targets` 非空，`exclude_targets` 必须为空，反之亦然。
两者都为空表示对所有 `config.toml` 中 `enabled = true` 的 target 都生效。

合法 target 名（与 `config.toml` 一致）:

```
claude-code  codex  cursor  copilot  windsurf  gemini  aider  cline  opencode
```

## 校验规则

`stdagent` 在 parse 阶段必须执行：

1. Frontmatter 必须存在且合法 YAML
2. `type` 在四种枚举内
3. `name` 满足正则 `^[a-z0-9][a-z0-9-]*$`
4. 同 type 下 `name` 唯一
5. `targets` 与 `exclude_targets` 不同时非空
6. `targets` / `exclude_targets` 内的值必须是合法 target 名
7. `applyTo` 内每条必须是合法 glob（与 doublestar 兼容）
8. `priority` 在 enum 内

校验失败：明确指出文件路径、字段名、违规原因，整个 sync 中止。

## 降级处理

当源文件没有 frontmatter（纯 Markdown）时：

- `type` 推断为 `rules`
- `name` 由文件名（去扩展名）推断
- 其他字段全为默认
- 输出 `WARN: <path> 缺少 frontmatter，按 rules 处理` 的提示

## 与各 target frontmatter 的映射

详见 [conversion-rules.md](conversion-rules.md)。
