# Target: Crush

调研日期: 2026-05-17
官方仓库: https://github.com/charmbracelet/crush
公司: Charmbracelet

## 1. 摘要

Crush 是 Charmbracelet 团队 2025 年发布的终端原生 AI 编码助手，与
stdagent 同为 Go 实现，复用 Charmbracelet 的 Bubble Tea / Lipgloss
TUI 体系，社区采纳速度快，2026-Q1 已突破 5k stars。

定位与 codex / opencode 同属 AGENTS.md 系：项目根读 plain markdown，
无 frontmatter，自顶向下沿 git root 至 cwd merge。Crush 的差异在
**多上下文 / 多 skill 目录共消费**：默认通过 `context_paths` 同时读取
`AGENTS.md` / `CRUSH.md` / `CLAUDE.md` / `GEMINI.md` 任一存在的文件，
并扫描 `.agents/skills/` / `.crush/skills/` / `.claude/skills/` /
`.cursor/skills/` 四个目录加载 Agent Skills 标准包，让单仓库无须重复
书写即可被 crush / codex / claude-code / cursor 同时识别。

## 2. 配置文件路径

| 类型 | 路径 | 说明 |
|---|---|---|
| 全局 CRUSH.md | `~/.config/crush/CRUSH.md` | crush 私有全局规则 |
| 全局 AGENTS.md | `~/.config/AGENTS.md` | 跨工具全局规则 |
| 项目级 CRUSH.md | `<repo>/CRUSH.md` | crush 私有项目规则（优先） |
| 项目级 AGENTS.md | `<repo>/AGENTS.md` | 跨工具项目规则 |
| 项目级 CLAUDE.md | `<repo>/CLAUDE.md` | 来自 claude-code 的共享上下文 |
| 项目级 GEMINI.md | `<repo>/GEMINI.md` | 来自 gemini-cli 的共享上下文 |
| 配置文件 | `<repo>/crush.json` 或 `.crush.json` | model / provider / context_paths 等 |
| 项目 Skills | `<repo>/.crush/skills/<name>/SKILL.md` | crush 私有 skill |
| 共消费 Skills | `.agents/skills/` / `.claude/skills/` / `.cursor/skills/` | crush 也扫描 |
| 全局 Skills | `~/.config/crush/skills/` | 用户级 skill |

`context_paths` 在 `crush.json` 内可配置，默认包含上述 4 个 markdown 文件名。
Skills 加载路径 hardcoded 在 crush 二进制内，不可关闭。

## 3. 文件格式

| 文件 | 格式 | frontmatter |
|---|---|---|
| `CRUSH.md` / `AGENTS.md` | Markdown | 无 frontmatter，纯指令文本 |
| `SKILL.md` | Markdown + YAML frontmatter | Agent Skills 标准字段 |
| `crush.json` | JSON | 不适用 |

字符上限 UNKNOWN：官方文档未列出 rule / skill 字节上限，实测单文件 50k+
仍能加载，但建议遵循通用 8k 软限（与 codex 一致）。

## 4. std-ai 四类映射

| std-ai 类型 | crush 落点 | 加载方式 |
|---|---|---|
| rules | inline 到 `<repo>/CRUSH.md` 主体（crush 无子目录 rules 支持） | 启动时自动读 |
| skills | `<repo>/.crush/skills/<name>/SKILL.md` + 辅助文件 | 按 description 匹配自动加载 |
| commands | fallback `<repo>/.crush/rules/commands/<name>.md`，body 头含 std-ai HTML 注释 explainer | crush 不识别 slash 命令，作为上下文文本读入 |
| references | fallback `<repo>/.crush/rules/references/<name>.md` | 同上 |
| subagents | fallback `<repo>/.crush/rules/subagents/<name>.md` | crush 无 subagent 概念 |

stdagent 与 codex transformer 的关系：codex 写 `<repo>/AGENTS.md`，crush
transformer 写 `<repo>/CRUSH.md`，二者文件名不冲突，且 crush 在 context_paths
配置下两份都读取并 merge，等价于 stdagent 同一份 std 规则被 crush 完整消费。

## 5. 转换器实现要点

1. 复用 `protocol.AgentsMD`，`crushAdapter` 配置 `RootFileName: "CRUSH.md"`
2. `RulesDir` 留空 -> nonRoot rules inline 到 CRUSH.md 正文（与 amp / warp 风格一致，
   而非 codex 的 `.codex/memories/` 子目录）
3. `SkillsDir: ".crush/skills"`，使用 Agent Skills 标准字段白名单
   `name / description / license / compatibility / metadata`
4. commands / references / subagents 走 `protocol.BuildDegradedFileOp`，
   fallback 到 `.crush/rules/{commands,references,subagents}/<name>.md`，
   body 头注入 std-ai HTML 注释 explainer，frontmatter 写 `std-ai-type` 标识
5. `InjectTypeGlossary: true`，CRUSH.md 头部 prepend std-ai 类型速查段
6. `MaxBytesPerFile: 0`，无明确字节限制
7. 不写 `crush.json`：model / provider 由用户在 crush CLI 内配置，
   stdagent 不接管运行时设置
8. 不写 MCP 文件：crush 的 MCP 配置入口在 `crush.json` 的 `mcp_servers` 字段，
   stdagent v0.0.4 MCP 仅 dispatch 给 claude-code / cursor / copilot 三家

## 6. 信息来源

- https://github.com/charmbracelet/crush （访问日期 2026-05-17）
- https://github.com/charmbracelet/crush/blob/main/README.md（访问日期 2026-05-17）
- /tmp/std-ai-protocol-research.md 第 28 行（调研日期 2026-05-17）

## 7. 已确认

- Crush 通过 `context_paths` 同时读 AGENTS.md / CRUSH.md / CLAUDE.md / GEMINI.md
- Skills 扫描 `.agents/skills/` / `.crush/skills/` / `.claude/skills/` / `.cursor/skills/`
  四目录，Agent Skills 标准包格式
- 配置文件 `crush.json` / `.crush.json`，含 `mcp_servers` 字段（v0.0.4 不接管）
- 5k+ stars，2026-Q1 增长快
- 与 codex（AGENTS.md）共消费：codex transformer 写 AGENTS.md，
  crush transformer 写 CRUSH.md，二者并存且 crush 同时读取

## 8. UNKNOWN

- rule / skill / 单文件字节硬上限（官方未列出）
- `context_paths` 是否支持嵌套子目录（与 codex 同样的 nested AGENTS.md 行为）
- `.crush/rules/` 是否是 crush 原生子目录 rules 入口（公开文档未提及，
  本 transformer 仅作为 std-ai 私有 fallback 容器使用）
