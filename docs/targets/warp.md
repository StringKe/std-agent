# Target: Warp

调研日期: 2026-05-17（2026-07-10 补充调研）
官方主页: https://warp.dev
官方文档: https://docs.warp.dev

## 1. 摘要

Warp 由 Warp Inc 开发，是 macOS / Linux / Windows 上的智能终端（Rust 实现），
内置 AI agent、阻塞式 prompt、跨终端 Workflows / Notebook / Drive 共享。
全球 Top-3 智能终端，闭源。

Warp 自 2026-01 起把项目级默认配置文件从 `WARP.md` 切换为 `AGENTS.md`
（与 Codex / Cursor / Antigravity 等对齐 Linux Foundation AAIF 标准），
旧的 `WARP.md` 仍向后兼容读取，且**两者同目录共存时 `WARP.md` 优先**
（https://docs.warp.dev/agent-platform/capabilities/rules/）。全局 Rules
通过 Warp Drive UI 管理（不在文件系统里），项目级 Rules 走根 `AGENTS.md` /
`WARP.md`。嵌套子目录下的 `AGENTS.md` 被 Warp Agent 自动叠加，与 Codex CLI
行为一致。两个文件名都**必须全大写**才会被识别，`agents.md` / `warp.md`
不生效。

Warp 已原生支持 Agent Skills 标准，推荐目录 `.agents/skills/<name>/SKILL.md`
（同时兼容读取 `.warp/skills/`、`.claude/skills/`、`.cursor/skills/` 等多种
命名空间，https://docs.warp.dev/agent-platform/capabilities/skills/），此前
调研认为"无原生 skills"已过时。commands 与 skills 是否完全合并未见官方明文
声明，仍无独立 commands 文件目录。

Warp 不支持任何形式的 frontmatter，AGENTS.md 是纯 Markdown 文本；references /
subagents 同样无独立文件目录。

## 2. 配置文件路径

| 类型 | 路径 | 说明 |
|---|---|---|
| 项目级 AGENTS.md | `<repo>/AGENTS.md` | 默认配置文件（2026-01+），文件名须全大写 |
| 项目级 WARP.md（旧） | `<repo>/WARP.md` | 向后兼容仍读，与 AGENTS.md 同存时优先生效 |
| 嵌套 AGENTS.md | `<repo>/<subdir>/AGENTS.md` | 进入子目录时自动叠加 |
| 项目 Skills | `<repo>/.agents/skills/<name>/SKILL.md` | 原生 Agent Skills 标准包（推荐路径） |
| 全局 Rules | Warp Drive UI（云端） | 不在文件系统，跨设备同步 |

Warp Drive 全局 Rules 通过 GUI 创建管理，stdagent 不写入。

## 3. 文件格式

| 文件 | 格式 | frontmatter |
|---|---|---|
| `AGENTS.md` / `WARP.md` | Markdown | 无 |
| 嵌套 `<subdir>/AGENTS.md` | Markdown | 无 |
| `.agents/skills/<name>/SKILL.md` | Markdown | Agent Skills 标准字段：`name` / `description` / `license` / `compatibility` / `metadata` |

Warp 不解析 AGENTS.md / WARP.md 的任何 frontmatter 字段，全部走纯文本指令；
skills 遵循标准 Agent Skills frontmatter。

## 4. std-agent 五类映射

| std-agent 类型 | Warp 落点 | 加载方式 |
|---|---|---|
| rules | 项目根 `AGENTS.md`（inline）；嵌套 root rule 写 `<NestedPath>/AGENTS.md` | 进入目录自动加载 |
| skills | `.agents/skills/<name>/SKILL.md`（原生标准包） | Warp Agent 按 description 自动发现 |
| commands | `.warp/rules/commands/<name>.md`（fallback，含 explainer + `std-agent-type: commands`） | 模型按 explainer 提示理解 |
| references | `.warp/rules/references/<name>.md`（同上） | 同上 |
| subagents | `.warp/rules/subagents/<name>.md`（同上） | 同上 |

`.warp/rules/` 是 stdagent 的私有 fallback 目录，Warp 不主动扫描；
fallback 文件靠根 `AGENTS.md` 中 manifest 段引用让模型可见。

## 5. 转换器实现要点

1. 协议族：`AgentsMD`，`RootFileName="AGENTS.md"`，`RulesDir=""` 强制
   nonRoot rules inline 进根 `AGENTS.md`
2. `NestedSupported=true`：嵌套 root rule（含 `NestedPath`）写到子目录 `AGENTS.md`，
   不带 manifest / glossary 头
3. `SkillsDir=".agents/skills"`：Warp 原生 Skills 推荐路径，`SkillSupportedFields`
   为 `name / description / license / compatibility / metadata`，与 codex / amp
   共享落点，字节一致由 writer 去重
4. `InjectTypeGlossary=true`：根 `AGENTS.md` 头部注入 std-agent 类型速查段，
   方便 Warp Agent 理解 `std-agent-type` 字段语义
5. `FallbackDir=".warp/rules"`：commands / references / subagents 走
   `BuildDegradedFileOp` 落到该目录下子目录，frontmatter 含
   `std-agent-type` + body 头含 explainer HTML 注释
6. 不写入 `WARP.md`（旧路径），统一走 `AGENTS.md`（与 codex / cursor / antigravity
   共用根文件）；注意 Warp 侧 `WARP.md` 若存在会优先于 `AGENTS.md` 生效，
   若用户仓库同时保留旧 `WARP.md`，stdagent 产出的 `AGENTS.md` 会被忽略，
   需要用户自行清理旧文件（stdagent 不主动删除非托管文件）
7. 不写入 Warp Drive 全局 Rules（GUI-only，无文件系统接口）

## 6. 信息来源

- https://warp.dev （主页）
- https://docs.warp.dev/agent-platform/capabilities/rules/（AGENTS.md 默认切换、
  WARP.md 优先级、文件名大小写要求，2026-07-10 复核）
- https://docs.warp.dev/agent-platform/capabilities/skills/（原生 Skills 目录，
  2026-07-10 新增）

## 7. 已确认

- Warp 自 2026-01 起把默认项目配置由 `WARP.md` 切换为 `AGENTS.md`
- 旧的 `WARP.md` 仍读取（向后兼容），且与 `AGENTS.md` 同目录共存时 `WARP.md` 优先
- `AGENTS.md` / `WARP.md` 文件名必须全大写才被识别
- 嵌套子目录 `AGENTS.md` 自动叠加（与 Codex CLI 行为一致）
- 全局 Rules 在 Warp Drive UI 管理，不在文件系统
- 全球 Top-3 智能终端，闭源
- 不支持 frontmatter，纯 Markdown
- 原生支持 Agent Skills 标准，推荐目录 `.agents/skills/<name>/SKILL.md`，同时兼容
  读取 `.warp/skills/` / `.claude/skills/` / `.cursor/skills/` 等命名空间
- 无独立 commands / references / subagents 文件目录

## 8. UNKNOWN

- Warp 对单文件 `AGENTS.md` 字节上限（未公开；保守不设 `MaxBytesPerFile`）
- 嵌套 `AGENTS.md` 的优先级合并策略（推测后加载者覆盖，未官方确认）
- Warp Drive 全局 Rules 与项目级 `AGENTS.md` 的优先级关系
- commands 是否会像 Amp 一样正式并入 skills（当前只有 skills 目录官方文档化，
  未见 commands 合并的明文声明）
