# Target: Claude Code

调研日期: 2026-05-07，2026-07-10 复核更新，2026-08-15 复核更新
官方文档: https://code.claude.com/docs/

## 1. 摘要

Claude Code（Anthropic 官方 CLI + Desktop + Agent SDK）是 Tier 1 中配置最丰富的目标。
专有概念包括 `CLAUDE.md` 多层加载、`.claude/` 子目录体系、hooks 事件系统、
sub-agents、skills、slash commands、settings.json 三层合并、`.mcp.json`。

Claude Code **不自动消费** `AGENTS.md`，必须在 `CLAUDE.md` 中通过 `@AGENTS.md`
import 才能融合（官方推荐做法）。

2026-07 复核：skills 与 commands 官方已合并 frontmatter 语义（"support the same
frontmatter"），当前 transformer（`protocol/claude_md.go`）已用同一个
`renderClaudeSkillFrontmatter` helper 渲染两者，command 也能用上 `when_to_use` /
`disable-model-invocation` / `paths` / `context` / `agent` / `effort` 等字段。
来源：https://code.claude.com/docs/en/skills

2026-08-15 复核：官方仍写明 Claude Code 读 `CLAUDE.md` 不自动读 `AGENTS.md`，推荐
`@AGENTS.md` import。`CLAUDE.md` 块级 HTML 注释在注入前被剥离。skill 在
`context: fork` 时新增 `background`（默认 true）。subagent 官方字段
`isolation` / `memory` / `permissionMode` / `maxTurns` / `skills` / `effort`
已由源 schema `isolation` / `memory` / `permission_mode` / `max_turns` /
`preload_skills` / `effort` 映射写出。changelog 见 https://code.claude.com/docs/en/changelog.md

## 2. CLAUDE.md 加载层级

| 优先级 | 路径 | 范围 | 说明 |
|---|---|---|---|
| 1（最高） | `/Library/Application Support/ClaudeCode/managed-settings.json` 等企业路径 | 企业 | 不可被覆盖 |
| 2 | `<project>/CLAUDE.md` | 项目 | git/repo 根，常用主入口 |
| 2b | `<project>/.claude/CLAUDE.md` | 项目 | 与根 `CLAUDE.md` 同级的合法项目 root 位置（2026-07 复核确认） |
| 3 | `<project>/CLAUDE.local.md` | 项目-本地 | 不入仓的本地附加 |
| 4 | `~/.claude/CLAUDE.md` | 用户全局 | 跨项目通用偏好 |
| 5 | 子目录 `<sub>/CLAUDE.md` | 子目录 | 进入子目录时懒加载 |

加载语义：

- 树形向上：从 cwd 向上每级目录扫描，遇到 `.git` 边界停止
- 子目录懒加载：当 Claude 操作的文件位于子目录时，沿路径向下读取；官方推荐 monorepo
  形态，`internal/transformer/protocol/claude_md.go` 的 `buildClaudeNestedRoot` 现实现正确
  （写到 `<NestedPath>/CLAUDE.md`，不带 manifest 不带 glossary）
- `@path/to/file` import：递归最多 **4 层**（旧文档写 5 层已过时，2026-07 复核修正）；可放大段落到独立 markdown

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
| `when_to_use` | 可选 | 官方存在但 transformer 暂无对应输出 | 可选（复用 skill frontmatter） |
| `paths` | 可选（路径限定） | **不支持**（2026-07 复核确认，subagent 无 paths 字段） | 可选（复用 skill frontmatter） |
| `disable-model-invocation` | 可选 | **不支持**（2026-07 复核确认） | 可选（复用 skill frontmatter） |
| `allowed-tools` | 可选（限制可用工具；旧版曾误写 `tools:` 被静默忽略，已修复见 # 12） | 不适用 | 可选（复用 skill frontmatter） |
| `disallowed-tools` | 可选 | 可选（agents 用字段名 `disallowedTools`，驼峰） | 可选（复用 skill frontmatter） |
| `user-invocable` | 可选，默认 true，仅显式 false 才渲染 | 不适用 | 不适用（commands 本身即用户触发） |
| `model` | 可选 | 可选 | 可选 |
| `tools`（agents 专属字段名） | 不适用 | 可选（白名单，对应 transformer 内部的 `AllowedTools`） | 不适用 |
| `argument-hint` | 不适用 | 不适用 | 可选（提示参数填写） |
| `memory` | 不适用 | 源 `memory` -> `memory`（user / project / local） | 不适用 |
| `isolation` | 不适用 | 源 `isolation` -> `isolation`（如 `worktree`） | 不适用 |
| `skills` | 不适用 | 源 `preload_skills` -> `skills` | 不适用 |
| `permissionMode` | 不适用 | 源 `permission_mode` -> `permissionMode` | 不适用 |
| `background` | `context: fork` 时可选 | 可选，bool | 不适用 |
| `maxTurns` | 不适用 | 源 `max_turns` -> `maxTurns` | 不适用 |

agents 一行的 `isolation` / `skills` / `memory` / `permissionMode` / `maxTurns` /
`effort` 已在 2026-08-15 接入源 schema 与 `buildClaudeSubagentFile`。`mcpServers` /
`hooks` / `color` / `initialPrompt` 仍无对应 std 字段，保持不写出。

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

## 8. std-agent 五类映射（实际实现，claude_code.go / protocol/claude_md.go）

| std-agent 类型 | Claude Code 落点 |
|---|---|
| rules（root） | `CLAUDE.md`（含 type glossary 头部 + root body + Imported Rules manifest 段） |
| rules（nonRoot） | `.claude/rules/<name>.md`；frontmatter 用 `paths`（Anthropic 私有方言，等价 Copilot 的 `applyTo`） |
| rules（nested，`NestedPath` 非空） | `<NestedPath>/CLAUDE.md`，纯 body，无 manifest 无 glossary |
| skills | `.claude/skills/<name>/SKILL.md` + 同目录 `SkillFiles` 辅助文件 |
| commands | `.claude/commands/<name>.md`，frontmatter 复用 skill 全字段集 |
| subagents | `.claude/agents/<name>.md`（Claude Code 原生支持，非 v1.0 遗留的"不生成"，见 # 12 修正） |
| references | **原生无 references 类型**；fallback 到 `.claude/rules/references/<name>.md`（rule-equivalent 形式，非 SKILL.md；frontmatter 注入 `std-agent-type: references`） |

## 9. 转换器实现要点（对照 `internal/transformer/claude_code.go` + `internal/transformer/protocol/claude_md.go`）

1. `CLAUDE.md` 主入口：type glossary（`InjectTypeGlossary=true`）+ root rule body + `Imported Rules` manifest（nonRoot rule 索引）+ stdagent header/footer marker
2. `.claude/rules/<name>.md`：frontmatter 仅 `paths`（GlobsList 格式）+ `description`；Claude Code 官方仅认 `paths`，不支持 `alwaysApply` / `applyTo` / `globs`
3. `.claude/skills/<name>/SKILL.md`：`renderClaudeSkillFrontmatter` 渲染 Agent Skills 标准字段（`name` / `description` / `license` / `compatibility` / `metadata`）+ Anthropic 私有字段集（`when_to_use` / `model` / `argument-hint` / `arguments` / `effort` / `context` / `agent` / `shell` / `allowed-tools` / `disallowed-tools` / `paths` / `disable-model-invocation` / `user-invocable` / `hooks`）
4. `.claude/commands/<name>.md`：直接复用第 3 点同一渲染函数（官方 commands/skills frontmatter 已合并），不再是旧版的 `argument_hint -> argument-hint` 单字段映射
5. `.claude/agents/<name>.md`：**已实现**（`buildClaudeSubagentFile`），frontmatter `name` / `description` / `model` / `tools`（来自 `AllowedTools`）/ `disallowedTools` / `background`；旧文档"v1.0 不直接生成"已过时
6. references：无原生类型，统一走 `BuildDegradedFileOp` fallback 到 `.claude/rules/references/<name>.md`
7. `settings.json` / `.mcp.json`：`.mcp.json` 已实现（`buildClaudeMCPJSON`，顶级键 `mcpServers`）；`settings.json`（hooks/权限等）仍未实现，保留扩展位
8. 备份：sync 前对 `CLAUDE.md` `.claude/rules/` `.claude/skills/` `.claude/commands/` `.claude/agents/` 做快照

## 10. budget.go 限额（2026-07 回填，来源 docs/sdlc/2026-07-10/spec-refresh/spec.md）

| kind | Soft | Hard | 语义 |
|---|---|---|---|
| root-file（CLAUDE.md） | 8000 字符 | **0（无 Hard）** | 官方明确 CLAUDE.md 每次 session 全量加载且不截断；旧文档/代码曾写 Hard 32000 无官方依据，已归 0，只留软指导；官方软指导 CLAUDE.md under 200 lines |
| rule（`*` 通用） | 8000 字符 | 0 | 跨工具通用 context 友好建议 |
| skill（`*` 通用） | 20000 字符 | 0 | Agent Skills 规范建议正文 < 500 行 |
| skill-name（`*` 通用） | - | 64 字符 | Agent Skills 规范硬拒载：name 须等于目录名 |
| skill-description（`*` 通用） | - | 1024 字符 | Agent Skills 规范硬拒载 |
| skill-listing（claude-code 专属） | - | 1536 字符 | `description + when_to_use` 合计，超限 skill listing 中被截断（调用时仍读全文），触发词要前置 |

三种超限语义要分清：**拒载**（skill-name / skill-description，超限整个 skill 不被索引）、
**截断**（skill-listing，只影响 listing 摘要不影响全文调用）、**无 Hard 只 Soft**（root-file / rule / skill，纯 context 友好提醒）。

## 11. 信息来源

- https://code.claude.com/docs/en/memory
- https://code.claude.com/docs/en/skills
- https://code.claude.com/docs/en/sub-agents
- https://code.claude.com/docs/en/slash-commands
- https://code.claude.com/docs/en/hooks
- https://code.claude.com/docs/en/settings
- https://code.claude.com/docs/en/mcp

## 12. output-styles 详细 schema

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

## 13. 已确认与剩余 UNKNOWN（2026-07-10 复核）

已确认（本轮新增）：
- P0 修复：SKILL.md frontmatter 官方字段是 `allowed-tools`，不是 `tools:`（`tools` 是 subagent 字段名）；`protocol/claude_md.go:211-213` 已改为 `allowed-tools`，`claude_md_test.go` 已改为红后绿复现该 bug；来源 https://code.claude.com/docs/en/skills
- rules `paths` frontmatter 字段已官方确认（不是猜测）
- subagent **不支持** `when_to_use` / `paths` / `disable-model-invocation`（2026-07 复核确认，非 UNKNOWN）
- `@import` 深度为 **4 hops**（旧文档写 5 层已过时）
- `.claude/CLAUDE.md` 也是合法项目 root 位置，与根 `CLAUDE.md` 同级
- Claude Code 不读 `AGENTS.md`（无变化，再次确认）
- output-styles schema（见 # 12）；v1.0 不主动生成 output-styles 文件（无对应 std type）

剩余 UNKNOWN（2026-07 复核仍未证实）：
- Claude Code 对未知 frontmatter 字段是严格报错还是静默忽略（`license` / `compatibility` / `metadata` 三字段命运未证实）
- `permissions.ask` 与 `deny` 的合并语义在多层 settings 中的具体细节
- `mcpServers` / `hooks` / `color` / `initialPrompt` 是否纳入 std schema（2026-08 仍未承接）
