# Target: Crush

调研日期: 2026-05-17（2026-07-10 补充调研）
官方仓库: https://github.com/charmbracelet/crush
官方文档: https://charmbracelet-crush.mintlify.app
公司: Charmbracelet

## 1. 摘要

Crush 是 Charmbracelet 团队 2025 年发布的终端原生 AI 编码助手，与
stdagent 同为 Go 实现，复用 Charmbracelet 的 Bubble Tea / Lipgloss
TUI 体系，社区采纳速度快，2026-Q1 已突破 5k stars。

定位与 codex / opencode 同属 AGENTS.md 系：项目根读 plain markdown，
无 frontmatter。差异在**多上下文文件共消费**：默认通过 `context_paths`
同时读取 `AGENTS.md` / `CRUSH.md` / `CLAUDE.md` / `GEMINI.md` 等任一存在
的文件，全部命中按序拼接，让单仓库无须重复书写即可被 crush / codex /
claude-code / gemini 同时识别。

**Skills 闭环需显式声明**：crush 二进制默认只扫描两个全局路径
`~/.config/crush/skills/` 与 `~/.config/agents/skills/`；项目级 skills 目录
（如 `.crush/skills/`）**不在默认扫描范围**，必须在项目 `crush.json` 的
`options.skills_paths` 数组里显式声明才会被加载
（https://charmbracelet-crush.mintlify.app/configuration/skills）。这与此前
调研认为"四目录 hardcoded 自动扫描"不符，已修正。

**无父目录 / 子目录发现**：crush 源码按 cwd 解析 `context_paths`，不做父目录
上溯、不做子树自动发现（https://github.com/charmbracelet/crush/blob/main/internal/agent/prompt/prompt.go）。
子目录 `x/y/CRUSH.md` 只有恰好在该子目录启动 crush 时才被读到，且此时根
`CRUSH.md` 会丢失，写入嵌套文件弊大于利。

## 2. 配置文件路径

| 类型 | 路径 | 说明 |
|---|---|---|
| 全局 CRUSH.md | `~/.config/crush/CRUSH.md` | crush 私有全局规则 |
| 全局 AGENTS.md | `~/.config/AGENTS.md` | 跨工具全局规则 |
| 项目级 CRUSH.md | `<repo>/CRUSH.md` | crush 私有项目规则 |
| 项目级 AGENTS.md | `<repo>/AGENTS.md` | 跨工具项目规则 |
| 项目级 CLAUDE.md | `<repo>/CLAUDE.md` | 来自 claude-code 的共享上下文 |
| 项目级 GEMINI.md | `<repo>/GEMINI.md` | 来自 gemini-cli 的共享上下文 |
| 配置文件 | `<repo>/crush.json` 或 `.crush.json` | model / provider / `context_paths` /
  `options.skills_paths` 等 |
| 全局 Skills（默认扫描） | `~/.config/crush/skills/`、`~/.config/agents/skills/` | 唯二默认生效的全局路径 |
| 项目 Skills（需声明） | `<repo>/.crush/skills/<name>/SKILL.md` | 必须在 `crush.json` 的
  `options.skills_paths` 里显式列出该路径才会被扫描 |

`context_paths` 在 `crush.json` 内可配置，默认包含上述多个 markdown 文件名，
全部命中即按顺序拼接注入，不做单选。

## 3. 文件格式

| 文件 | 格式 | frontmatter |
|---|---|---|
| `CRUSH.md` / `AGENTS.md` | Markdown | 无 frontmatter，纯指令文本 |
| `SKILL.md` | Markdown + YAML frontmatter | Agent Skills 标准字段 |
| `crush.json` | JSON | 不适用 |

字符上限 UNKNOWN：官方文档未列出 rule / skill 字节上限，实测单文件 50k+
仍能加载，但建议遵循通用 8k 软限（与 codex 一致）。

## 4. std-agent 四类映射

| std-agent 类型 | crush 落点 | 加载方式 |
|---|---|---|
| rules | inline 到 `<repo>/CRUSH.md` 主体（crush 无子目录 rules 支持） | 启动时自动读 |
| skills | `<repo>/.crush/skills/<name>/SKILL.md` + 辅助文件，同时在 `crush.json` 注册
  `options.skills_paths` | 按 description 匹配自动加载（前提是路径已在 crush.json 声明） |
| commands | fallback `<repo>/.crush/rules/commands/<name>.md`，body 头含 std-agent HTML 注释 explainer | crush 不识别 slash 命令，作为上下文文本读入 |
| references | fallback `<repo>/.crush/rules/references/<name>.md` | 同上 |
| subagents | fallback `<repo>/.crush/rules/subagents/<name>.md` | crush 无 subagent 概念 |

stdagent 与 codex transformer 的关系：codex 写 `<repo>/AGENTS.md`，crush
transformer 写 `<repo>/CRUSH.md`，二者文件名不冲突，且 crush 在 `context_paths`
默认配置下两份都读取并 merge，等价于 stdagent 同一份 std 规则被 crush 完整消费。

**不支持嵌套**：与 codex / amp / windsurf 不同，crush 不会自动发现子目录下的
`CRUSH.md` / `AGENTS.md`，stdagent 也不为 crush 产出 NestedPath 落点。

## 5. 转换器实现要点

1. 复用 `protocol.AgentsMD`，`crushAdapter` 配置 `RootFileName: "CRUSH.md"`
2. `RulesDir` 留空 -> nonRoot rules inline 到 CRUSH.md 正文（与 amp / warp 风格一致，
   而非 codex 的 `.codex/memories/` 子目录）
3. `NestedSupported: false`：crush 源码只解析相对 cwd 的 context path，无父目录
   上溯、无子树自动发现，写嵌套 `CRUSH.md` 弊大于利，故不产出
4. `SkillsDir: ".crush/skills"`，使用 Agent Skills 标准字段白名单
   `name / description / license / compatibility / metadata`
5. Plan 在产出 skills 类型文档时，额外追加一个 `writer.FileOp{Path: "crush.json",
   JSONMerge: true}`，内容为 `{"options":{"skills_paths":[".crush/skills"]}}`，
   闭环声明项目级 skills 路径（对应决策点 A 的闭环方案）。用户已有 `crush.json`
   时按 JSON 深合并（数组并集、scalar 保留用户值），解析失败（如含注释）则跳过
   并 WARN，不覆盖用户文件（`internal/writer/writer.go` 的 JSONMerge 逻辑，
   测试见 `internal/writer/jsonmerge_test.go`）
6. commands / references / subagents 走 `protocol.BuildDegradedFileOp`，
   fallback 到 `.crush/rules/{commands,references,subagents}/<name>.md`，
   body 头注入 std-agent HTML 注释 explainer，frontmatter 写 `std-agent-type` 标识
7. `InjectTypeGlossary: true`，CRUSH.md 头部 prepend std-agent 类型速查段
8. `MaxBytesPerFile: 0`，无明确字节限制
9. 不写 model / provider 配置：`crush.json` 只由 Plan 追加 `options.skills_paths`
   一个字段，其余 model / provider / MCP 设置由用户在 crush CLI 内自行配置

## 6. 信息来源

- https://github.com/charmbracelet/crush （访问日期 2026-05-17）
- https://github.com/charmbracelet/crush/blob/main/README.md（访问日期 2026-05-17）
- https://charmbracelet-crush.mintlify.app/configuration/skills（项目级 skills 需
  `options.skills_paths` 显式声明，2026-07-10 新增）
- https://github.com/charmbracelet/crush/blob/main/internal/agent/prompt/prompt.go（
  无父目录上溯 / 无子树发现，2026-07-10 新增）

## 7. 已确认

- Crush 通过 `context_paths` 同时读 AGENTS.md / CRUSH.md / CLAUDE.md / GEMINI.md
  等，全部命中按序拼接
- 默认只扫描两个全局 skills 路径 `~/.config/crush/skills/` 与
  `~/.config/agents/skills/`；项目级 skills 目录必须在 `crush.json` 的
  `options.skills_paths` 显式声明才被扫描，stdagent 已实现自动 JSONMerge 注册
- crush 不做父目录上溯、不做子树自动发现，嵌套 CRUSH.md 无法被正常消费，
  transformer `NestedSupported` 已设为 `false`
- 配置文件 `crush.json` / `.crush.json`，含 `mcp_servers` 字段（v0.0.4 起本
  transformer 不接管）
- 5k+ stars，2026-Q1 增长快
- 与 codex（AGENTS.md）共消费：codex transformer 写 AGENTS.md，
  crush transformer 写 CRUSH.md，二者并存且 crush 同时读取

## 8. UNKNOWN

- rule / skill / 单文件字节硬上限（官方未列出）
- `.crush/rules/` 是否是 crush 原生子目录 rules 入口（公开文档未提及，
  本 transformer 仅作为 std-agent 私有 fallback 容器使用）
- `crush.json` 与用户既有配置文件的合并策略在含注释 / 非标准 JSON 场景下的
  详细行为边界（当前实现遇到解析失败即跳过并 WARN，未做局部修复）
