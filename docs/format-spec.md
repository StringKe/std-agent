# 源文件格式规范

## 概览

`.stdai/standards/` 下所有文件必须为 **Markdown + YAML Frontmatter**。
Frontmatter 是文件开头被 `---` 包裹的一段 YAML，用于声明元数据。
正文是标准 CommonMark Markdown，将被各 target 转换器消费。

## 文件头模板

```markdown
---
type: rules                # 必填：rules | skills | commands | references
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
| `type` | string | enum: rules/skills/commands/references | 决定走哪类转换 |
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
| `applyTo` | string[] | [] | gitignore 风格 glob |
| `alwaysApply` | bool | false | Cursor always-on |
| `allowed_tools` | string[] | [] | 仅 commands |
| `argument_hint` | string | "" | 仅 commands |
| `model` | string | "" | 仅 commands/skills/agents |

## 四种 type 的语义

### rules

通用规则文件。会被转换并写入到各 target 的 rules/instructions 位置：

- Claude Code: `.claude/rules/<name>.md`（或合并到 CLAUDE.md，根据策略）
- Codex: `.codex/rules/<name>.md`（或拼接到 AGENTS.md 末尾）
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

### references

参考文档/上下文资料。一般不直接被 AI 加载，作为人类可读的支撑材料；
也可通过 Claude Code 的 `@<path>` import 机制被显式引用。

- 所有 target 默认输出到 `.stdai/standards/references/` 镜像位置或 `docs/`
  下一个独立子目录（路径由 config 控制）
- 不进入 `CLAUDE.md` / `AGENTS.md` 等入口文件

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
