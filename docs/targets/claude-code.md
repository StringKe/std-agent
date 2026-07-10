# Target: Claude Code

调研日期: 2026-05-07
官方文档: https://code.claude.com/docs/

## 1. 摘要

Claude Code（Anthropic 官方 CLI + Desktop + Agent SDK）是 Tier 1 中配置最丰富的目标。
专有概念包括 `CLAUDE.md` 多层加载、`.claude/` 子目录体系、hooks 事件系统、
sub-agents、skills、slash commands、settings.json 三层合并、`.mcp.json`。

Claude Code **不自动消费** `AGENTS.md`，必须在 `CLAUDE.md` 中通过 `@AGENTS.md`
import 才能融合。

## 2. CLAUDE.md 加载层级

| 优先级 | 路径 | 范围 | 说明 |
|---|---|---|---|
| 1（最高） | `/Library/Application Support/ClaudeCode/managed-settings.json` 等企业路径 | 企业 | 不可被覆盖 |
| 2 | `<project>/CLAUDE.md` | 项目 | git/repo 根，常用主入口 |
| 3 | `<project>/CLAUDE.local.md` | 项目-本地 | 不入仓的本地附加 |
| 4 | `~/.claude/CLAUDE.md` | 用户全局 | 跨项目通用偏好 |
| 5 | 子目录 `<sub>/CLAUDE.md` | 子目录 | 进入子目录时懒加载 |

加载语义：

- 树形向上：从 cwd 向上每级目录扫描，遇到 `.git` 边界停止
- 子目录懒加载：当 Claude 操作的文件位于子目录时，沿路径向下读取
- `@path/to/file` import：递归最多 5 层；可放大段落到独立 markdown

## 3. `.claude/` 项目级目录全景

| 子项 | 用途 | 文件格式 |
|---|---|---|
| `settings.json` | 共享给团队的配置 | JSON |
| `settings.local.json` | 本地不入仓配置（覆盖 settings.json） | JSON |
| `rules/*.md` | 路径条件加载的规则文件 | Markdown，可带 frontmatter |
| `skills/<name>/SKILL.md` | 自定义 slash 触发的能力包 | Markdown + frontmatter |
| `agents/<name>.md` | Sub-agent 定义 | Markdown + frontmatter |
| `commands/<name>.md` | 自定义 slash command | Markdown + frontmatter |
| `output-styles/<name>.md` | 输出风格预设 | Markdown |
| `.mcp.json` | 项目级 MCP server | JSON |

## 4. Frontmatter 字段对照表

| 字段 | skills (`SKILL.md`) | agents | commands |
|---|---|---|---|
| `name` | 必填 | 必填 | 文件名推断 |
| `description` | 必填 | 必填 | 必填 |
| `when_to_use` | 可选 | UNKNOWN | UNKNOWN |
| `paths` | 可选（路径限定） | UNKNOWN | UNKNOWN |
| `disable-model-invocation` | 可选 | UNKNOWN | 不适用 |
| `tools` | 可选（限制可用工具） | 可选（白名单） | UNKNOWN |
| `disallowedTools` | UNKNOWN | 可选（黑名单） | UNKNOWN |
| `model` | 可选 | 可选 | 可选 |
| `allowed-tools` | 不适用 | 不适用 | 可选（限制 slash 内可用工具） |
| `argument-hint` | 不适用 | 不适用 | 可选（提示参数填写） |
| `memory` | 不适用 | 可选 | 不适用 |
| `isolation` | 不适用 | 可选（如 `worktree`） | 不适用 |
| `skills` | 不适用 | 可选（预加载 skills 列表） | 不适用 |
| `maxTurns` | 不适用 | 可选 | 不适用 |

## 5. Slash command 参数变量

| 变量 | 含义 |
|---|---|
| `$1` `$2` ... | 位置参数 |
| `$ARGUMENTS` | 全部参数原样字符串 |

## 6. Hook 事件

settings.json 的 `hooks` 字段，事件名清单：

| 事件 | 触发时机 |
|---|---|
| `PreToolUse` | 工具调用前；exit code 2 拦截 |
| `PostToolUse` | 工具调用后 |
| `UserPromptSubmit` | 用户提交 prompt 前 |
| `Stop` | 主循环停止 |
| `Notification` | 通知投递 |
| `SessionStart` | 会话开始 |
| `SessionEnd` | 会话结束 |
| `SubagentStart` | sub-agent 开始 |
| `SubagentStop` | sub-agent 结束 |
| `PermissionRequest` | 权限请求时 |

## 7. settings.json 关键字段

```json
{
  "model": "claude-sonnet-4-6",
  "permissions": {
    "allow": ["Read", "Bash(git status:*)"],
    "deny": [],
    "ask": []
  },
  "env": { "FOO": "bar" },
  "hooks": { "PreToolUse": [...], "PostToolUse": [...] },
  "mcpServers": { "<name>": { ... } },
  "statusLine": { "command": "/path/to/script" },
  "includeCoAuthoredBy": false
}
```

三层合并优先级：managed（企业，最高）> project（`.claude/settings.json`）>
user（`~/.claude/settings.json`）。数组合并不覆盖。

## 8. std-agent 四类映射

| std-agent 类型 | Claude Code 落点 |
|---|---|
| rules | `.claude/rules/<name>.md`；可选合并到 `CLAUDE.md` 的尾部段（取决于 conversion 策略） |
| skills | `.claude/skills/<name>/SKILL.md` + 同目录辅助文件 |
| commands | `.claude/commands/<name>.md`（带 frontmatter） |
| references | 不主动写入；通过 `CLAUDE.md` 的 `@<path>` 显式引用 |

## 9. 转换器实现要点

1. `CLAUDE.md` 主入口：固定模板 + footer 注入；rules 索引可在尾部以 `@.claude/rules/<name>.md` 列表方式 import
2. `.claude/rules/<name>.md`：保留源文件 frontmatter（`applyTo` 等通用字段保留）
3. `.claude/skills/<name>/SKILL.md`：frontmatter 必有 `name` `description`，从源文件 mapped
4. `.claude/commands/<name>.md`：源 frontmatter 的 `argument_hint` -> `argument-hint`，`allowed_tools` -> `allowed-tools`，`model` 直传
5. `.claude/agents/<name>.md`：v1.0 不直接生成（无 std type），保留扩展位
6. `settings.json` / `.mcp.json` 的写入由 v1.1 处理（涉及 hooks/MCP）；v1.0 仅写 markdown 类
7. 备份：sync 前对 `CLAUDE.md` `.claude/rules/` `.claude/skills/` `.claude/commands/` 做快照

## 10. 信息来源

- https://code.claude.com/docs/en/memory
- https://code.claude.com/docs/en/skills
- https://code.claude.com/docs/en/sub-agents
- https://code.claude.com/docs/en/slash-commands
- https://code.claude.com/docs/en/hooks
- https://code.claude.com/docs/en/settings
- https://code.claude.com/docs/en/mcp

## 11. output-styles 详细 schema

文件名约定 `<name>.md` 单文件，**不是** `<name>/STYLE.md` 子目录形式。

保存路径三级：

- 用户级 `~/.claude/output-styles/`
- 项目级 `.claude/output-styles/`
- Managed policy `.claude/output-styles/`（企业 managed settings 目录）
- Plugins 可在 `output-styles/` 目录中 ship

frontmatter 字段：

| 字段 | 必填 | 默认 | 说明 |
|---|---|---|---|
| `name` | 否（缺省取文件名） | 文件名 | 显示名 |
| `description` | 否 | - | `/config` 选择器中显示 |
| `keep-coding-instructions` | 否 | false | 是否保留 Claude Code 默认编码相关 system prompt |
| `force-for-plugin` | 否 | false | 仅 plugin 用，启用插件即自动应用，覆盖用户 outputStyle |

加载方式：**自动应用型**（不是 slash 触发）。运行 `/config` -> Output style 选择，
或直接编辑 settings 文件中 `outputStyle` 字段（写入 `settings.local.json`）。
session 启动时生效，运行中切换需开新会话。

正文格式：Markdown，会被追加到 system prompt 末尾。

来源：https://code.claude.com/docs/en/output-styles

## 12. 已确认与剩余 UNKNOWN

已确认：
- output-styles schema（见 # 11）
- v1.0 不主动生成 output-styles 文件（无对应 std type）

剩余 UNKNOWN（落代码前再次核实）：
- `agents` frontmatter 的 `when_to_use` `paths` `disable-model-invocation` 在最新版是否仍存在
- `permissions.ask` 与 `deny` 的合并语义在多层 settings 中的具体细节
