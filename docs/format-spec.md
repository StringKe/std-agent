# Source format

`.stdai/standards/` 是 stdagent 的单一真相源。普通文档使用 YAML frontmatter + Markdown；`root.md` 和无 frontmatter Markdown 按 rules 解析。

## Minimal document

```markdown
---
type: rules
name: coding-style
description: Go 命名和错误处理约束
priority: high
applyTo:
  - "**/*.go"
---

正文。
```

`name` 使用 `^[a-z0-9][a-z0-9-]*$`。省略时从文件名推导。`type` 省略时默认为 `rules`。

## Types

| type | 消费语义 | 标准路径 |
|---|---|---|
| `rules` | 持续约束或路径匹配规则 | `rules/<name>.md` |
| `skills` | AI 按 description 调用的能力包 | `skills/<name>/SKILL.md` |
| `commands` | 用户显式触发的模板 | `commands/<name>.md` |
| `references` | 按需查阅的背景资料 | `references/<name>.md` |
| `subagents` | 隔离上下文代理定义 | `subagents/<name>.md` |

`root.md` 是项目入口，不属于子目录，也不会 fan out 为普通 rule。`nested/<relative-path>/root.md` 是同一仓库的目录级入口。

## Common fields

| field | type | semantics |
|---|---|---|
| `type` | enum | 五种 source type |
| `name` | string | kebab-case 标识 |
| `description` | string | 触发或索引所需的简短说明 |
| `priority` | `high/normal/low` | 稳定排序 |
| `targets` | string[] | target 白名单 |
| `exclude_targets` | string[] | target 黑名单，与 `targets` 互斥 |
| `applyTo` / `globs` | string[] | 路径匹配；两者合并去重 |
| `alwaysApply` | bool | target 支持时映射为 always-on |

Target-specific path override 可使用 `claudecode.paths`、`codexcli.paths`、`cursor.paths`、`copilot.paths`、`windsurf.paths`、`gemini.paths`、`aider.paths`、`cline.paths` 或 `opencode.paths`。

## Capability fields

- commands：`argument_hint`、`allowed_tools`、`model`。
- skills：`when_to_use`、`arguments`、`effort`、`context`、`agent`、`shell`、`hooks`、`license`、`compatibility`、`metadata`、`disable_model_invocation`、`user_invocable`、`disallowed_tools`。
- subagents：`model`、`allowed_tools`、`disallowed_tools`、`readonly`、`background`、`isolation`、`memory`、`permission_mode`、`max_turns`、`preload_skills`。

省略 `model` 表示使用 runtime 默认模型。具体字段是否可消费由 target protocol 决定；无法原生表达的字段不得伪装为已支持。

## Skill packages

`skills/<name>/SKILL.md` 是 package 主文件。同目录的 `scripts/`、`references/`、`templates/`、`assets/` 等辅助文件随原生 skill 输出复制。

```text
skills/code-review/
├── SKILL.md
├── scripts/check.sh
└── references/checklist.md
```

## Filtering and ordering

- `targets` 非空时只应用于列出的 target。
- `exclude_targets` 非空时应用于其余 target。
- 两者都为空时应用于所有启用 target。
- 输出按 `priority` 后 `name` 稳定排序。

## Validation

Parser 拒绝非法 YAML、未知 type、非法 name、非法 priority，以及同时设置 `targets` 与 `exclude_targets`。无 frontmatter Markdown按 rules 兼容解析。

转换路径和字段方言见 [conversion-rules.md](conversion-rules.md)，target 官方证据见 [targets/](targets/)。
