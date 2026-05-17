# Target: Warp

调研日期: 2026-05-17
官方主页: https://warp.dev
官方文档: https://docs.warp.dev

## 1. 摘要

Warp 由 Warp Inc 开发，是 macOS / Linux / Windows 上的智能终端（Rust 实现），
内置 AI agent、阻塞式 prompt、跨终端 Workflows / Notebook / Drive 共享。
全球 Top-3 智能终端，闭源。

Warp 自 2026-01 起把项目级默认配置文件从 `WARP.md` 切换为 `AGENTS.md`
（与 Codex / Cursor / Antigravity 等对齐 Linux Foundation AAIF 标准），
旧的 `WARP.md` 仍向后兼容读取。全局 Rules 通过 Warp Drive UI 管理（不在
文件系统里），项目级 Rules 走根 `AGENTS.md`。嵌套子目录下的 `AGENTS.md`
被 Warp Agent 自动叠加，与 Codex CLI 行为一致。

Warp 不支持任何形式的 frontmatter，AGENTS.md 是纯 Markdown 文本；不支持
独立 skills / commands / references / subagents 文件目录。

## 2. 配置文件路径

| 类型 | 路径 | 说明 |
|---|---|---|
| 项目级 AGENTS.md | `<repo>/AGENTS.md` | 默认配置文件（2026-01+） |
| 项目级 WARP.md（旧） | `<repo>/WARP.md` | 向后兼容，仍读取 |
| 嵌套 AGENTS.md | `<repo>/<subdir>/AGENTS.md` | 进入子目录时自动叠加 |
| 全局 Rules | Warp Drive UI（云端） | 不在文件系统，跨设备同步 |

Warp Drive 全局 Rules 通过 GUI 创建管理，stdagent 不写入。

## 3. 文件格式

| 文件 | 格式 | frontmatter |
|---|---|---|
| `AGENTS.md` / `WARP.md` | Markdown | 无 |
| 嵌套 `<subdir>/AGENTS.md` | Markdown | 无 |

Warp 不解析任何 frontmatter 字段，全部走纯文本指令。

## 4. std-ai 四类映射

| std-ai 类型 | Warp 落点 | 加载方式 |
|---|---|---|
| rules | 项目根 `AGENTS.md`（inline）；嵌套 root rule 写 `<NestedPath>/AGENTS.md` | 进入目录自动加载 |
| skills | 降级为 `model_decision` rule（`SkillsAsRule=true`）；Warp 无原生 skill 概念 | 同上 |
| commands | `.warp/rules/commands/<name>.md`（fallback，含 explainer + `std-ai-type: commands`） | 模型按 explainer 提示理解 |
| references | `.warp/rules/references/<name>.md`（同上） | 同上 |
| subagents | `.warp/rules/subagents/<name>.md`（同上） | 同上 |

`.warp/rules/` 是 stdagent 的私有 fallback 目录，Warp 不主动扫描；
fallback 文件靠根 `AGENTS.md` 中 manifest 段引用让模型可见。

## 5. 转换器实现要点

1. 协议族：`AgentsMD`，`RulesDir=""` 强制 nonRoot rules inline 进根 `AGENTS.md`
2. `NestedSupported=true`：嵌套 root rule（含 `NestedPath`）写到子目录 `AGENTS.md`，
   不带 manifest / glossary 头
3. `SkillsAsRule=true`：skill 降级为 `trigger: model_decision` rule
4. `InjectTypeGlossary=true`：根 `AGENTS.md` 头部注入 std-ai 类型速查段，
   方便 Warp Agent 理解 `std-ai-type` 字段语义
5. `FallbackDir=".warp/rules"`：其他 type 走 `BuildDegradedFileOp` 落到
   该目录下子目录（commands / references / subagents），frontmatter 含
   `std-ai-type` + body 头含 explainer HTML 注释
6. 不写入 `WARP.md`（旧路径），统一走 `AGENTS.md`（与 codex / cursor / antigravity
   共用根文件）
7. 不写入 Warp Drive 全局 Rules（GUI-only，无文件系统接口）

## 6. 信息来源

- /tmp/std-ai-protocol-research.md 第 27 行（调研日期 2026-05-17）
- https://warp.dev （主页）
- https://docs.warp.dev/agent-mode/configuration （AGENTS.md 默认切换公告）

## 7. 已确认

- Warp 自 2026-01 起把默认项目配置由 `WARP.md` 切换为 `AGENTS.md`
- 旧的 `WARP.md` 仍读取（向后兼容）
- 嵌套子目录 `AGENTS.md` 自动叠加（与 Codex CLI 行为一致）
- 全局 Rules 在 Warp Drive UI 管理，不在文件系统
- 全球 Top-3 智能终端，闭源
- 不支持 frontmatter，纯 Markdown
- 无原生 skills / commands / references / subagents 独立目录

## 8. UNKNOWN

- Warp 对单文件 `AGENTS.md` 字节上限（未公开；保守不设 `MaxBytesPerFile`）
- 嵌套 `AGENTS.md` 的优先级合并策略（推测后加载者覆盖，未官方确认）
- Warp Drive 全局 Rules 与项目级 `AGENTS.md` 的优先级关系
