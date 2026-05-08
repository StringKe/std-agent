# Target: Codex (OpenAI)

调研日期: 2026-05-07
官方文档: https://developers.openai.com/codex/

## 1. 摘要

OpenAI Codex CLI 是 OpenAI 官方的命令行/IDE 插件 agent。配置围绕 `~/.codex/`
与项目级 `.codex/` 双层展开。`AGENTS.md` 是项目指令的主入口，子目录可放
`AGENTS.md` 与 `AGENTS.override.md` 实现 per-directory 覆盖。

Skills 走 OpenAI 通用规范 `$HOME/.agents/skills/` 与 `<repo>/.agents/skills/`
（注意是 `.agents` 而非 `.codex`，与官方 agent skills 协议一致）。

Codex **不消费** `CLAUDE.md` 或 `GEMINI.md`，但可通过 `project_doc_fallback_filenames`
显式声明 fallback。

## 2. 配置文件路径

| 类别 | 路径 | 说明 |
|---|---|---|
| 用户 config | `~/.codex/config.toml` | `CODEX_HOME` 可改 |
| 系统 config | `/etc/codex/config.toml` | 可选 |
| 项目 config | `.codex/config.toml` | 沿 cwd 向上至 project root，closest wins，仅 trusted project 生效 |
| 用户 AGENTS | `~/.codex/AGENTS.md` + `AGENTS.override.md` | 单文件 + 可覆盖 |
| 项目 AGENTS | 项目根 + 任意子目录 `AGENTS.md` / `AGENTS.override.md` | 每目录最多取一对 |
| Prompts (deprecated) | `~/.codex/prompts/*.md` | 已 deprecated，改用 skills |
| Skills (用户) | `$HOME/.agents/skills/<name>/SKILL.md` | 通用 agent skills 协议 |
| Skills (项目) | `<repo>/.agents/skills/<name>/SKILL.md` | 与仓库一同提交 |
| Hooks | `~/.codex/hooks.json` 或 `[hooks]` inline；项目 `.codex/hooks.json` | 受 `[features].codex_hooks=true` 控制 |

## 3. 文件格式

| 文件 | 格式 | frontmatter |
|---|---|---|
| `config.toml` | TOML | 无 |
| `AGENTS.md` / `AGENTS.override.md` | Markdown | 无（全文即指令） |
| `SKILL.md` | Markdown | 必填，字段见 OpenAI agents skills 规范（核心: `name`、`description`） |
| 自定义 prompt（弃用） | Markdown | `description`、`argument-hint` 可选 |
| `hooks.json` | JSON | 无 |

`AGENTS.md` 单文件大小受 `project_doc_max_bytes` 限制（默认 32768 字节）。

## 4. config.toml 关键字段

```toml
project_root_markers = [".git"]
project_doc_max_bytes = 32768
project_doc_fallback_filenames = ["AGENTS.md", "CLAUDE.md"]

[features]
codex_hooks = true
memories = true

[mcp_servers.<name>]
command = "..."
args = ["..."]
env = { K = "V" }
enabled = true
enabled_tools = ["*"]
disabled_tools = []
startup_timeout_sec = 10
tool_timeout_sec = 60

[[skills.config]]
path = "/abs/path/to/SKILL.md"
enabled = true

[hooks]
# inline 形式
```

## 5. AGENTS.md 加载顺序

```
全局: ~/.codex/AGENTS.override.md (若存在则替换 ~/.codex/AGENTS.md)
项目: project root -> ... -> cwd
  每层先 AGENTS.override.md 再 AGENTS.md，最后 fallback filenames
```

closest wins：越靠近 cwd 的优先级越高，向 root 拼接。

## 6. std-ai 四类映射

| std-ai 类型 | Codex 落点 |
|---|---|
| rules | 项目 `AGENTS.md`（默认全部 rules 拼接到一个文件）；子目录 rules 写入 `<sub>/AGENTS.md` 或 `<sub>/AGENTS.override.md` |
| skills | `<repo>/.agents/skills/<name>/SKILL.md` + 同目录辅助文件 |
| commands | 内置 slash 不可扩展；自定义 prompt 已 deprecated。v1.1 起降级为 skill 写到 `.agents/skills/cmd-<n>/SKILL.md`，description 含 slash 调用 hint 让模型主动调用 |
| references | 不主动写入；建议作为 rules 段落或独立 docs/ 目录 |

## 7. 转换器实现要点

1. 主输出：项目根 `AGENTS.md`，由 `inject` footer 标识为 stdagent 生成
2. rules 拼接策略：所有 `targets` 含 `codex` 的 rules 文件按 `priority` -> `name`
   排序拼接到 `AGENTS.md` 正文，每段以 `## <name>` 二级标题分隔
3. AGENTS.md 总字节超过 `project_doc_max_bytes`（32768 字节）时：分拆，把
   超出部分写入 `.codex/rules/<name>.md` 并在 AGENTS.md 末尾追加 Rules
   Reference 菜单（"## Rules Reference\n- [coding-style](.codex/rules/coding-style.md)"）
4. skills：写入 `.agents/skills/<name>/SKILL.md` + 辅助文件；frontmatter 至少
   含 `name`、`description`
5. trust 提示：sync 时检测 `.codex/` 已存在却未 trusted，输出 WARN
6. v1.0 不写 hooks.json / config.toml；保留 v1.1 扩展位

## 8. 信息来源

- https://developers.openai.com/codex/concepts/customization
- https://developers.openai.com/codex/guides/agents-md
- https://developers.openai.com/codex/cli/slash-commands
- https://developers.openai.com/codex/cli/features
- https://developers.openai.com/codex/config-basic
- https://developers.openai.com/codex/config-reference
- https://developers.openai.com/codex/mcp
- https://github.com/openai/codex

## 9. 已确认与剩余 UNKNOWN

已确认：
- `[[skills.config]]` schema：TOML 数组表，字段 `path`（string，指向 SKILL.md）
  + `enabled`（bool）。**不自动扫描**，仅显式注册；用途是"临时禁用某个 skill"
  或"注册非默认路径"
- 子代理（subagents）省略 `[[skills.config]]` 时继承父会话配置；修改 `~/.codex/config.toml`
  后必须重启 Codex
- skills 文档正确路径 https://developers.openai.com/codex/skills（之前 `/concepts/skills` 为 404）
- v1.0 不主动写 `[[skills.config]]`，stdagent 仅落 SKILL.md 到 `.agents/skills/`，
  由 Codex 默认机制发现

剩余 UNKNOWN：
- Memories 系统具体落盘格式
- 自定义 prompt 完全废弃后官方推荐 skills 的最小字段集
