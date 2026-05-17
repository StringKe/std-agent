# Target: Grok CLI

调研日期: 2026-05-17
官方仓库: https://github.com/superagent-ai/grok-cli (fork chosen)
其他 fork: https://github.com/alphaonedev/grok-cli / https://github.com/baba20o/grok-cli
公司: xAI（模型）/ superagent-ai（CLI 实现）

## 1. 摘要

Grok CLI 是 xAI Grok 模型的命令行 agent。**当前生态有 3 个并行 fork**，
主流未定：

- `superagent-ai/grok-cli`（最早实现，v0.0.4 对齐目标）
- `alphaonedev/grok-cli`
- `baba20o/grok-cli`

v0.0.4 对齐 `superagent-ai` fork，docs 与 transformer 行为以该 fork 为准；
未来若另一 fork 取得社区主导地位，v0.0.5 可能 pivot 切换。

协议归属：AGENTS.md 系。Grok CLI 读项目根 `AGENTS.md` 并自顶向下沿
git root 至 cwd merge，子目录可通过 `AGENTS.override.md` 实现 per-dir
覆盖（与 codex 的同名机制语义相同）。`.grok/settings.json` 存项目级
MCP / model 配置，`.grok/GROK.md` 是早期版本的项目级 markdown 入口，
已被 `AGENTS.md` 取代但仍向后兼容。用户级配置在 `~/.grok/user-settings.json`。

## 2. 配置文件路径

| 类型 | 路径 | 说明 |
|---|---|---|
| 项目 AGENTS | `<repo>/AGENTS.md` | 主入口，与 codex / amp / crush 共享 |
| 嵌套 AGENTS | `<repo>/<subdir>/AGENTS.md` | per-dir 追加规则 |
| Per-dir 覆盖 | `<repo>/<subdir>/AGENTS.override.md` | 子目录强制覆盖父层 rules |
| 项目设置 | `<repo>/.grok/settings.json` | MCP servers / model / provider |
| 项目级 GROK.md（旧） | `<repo>/.grok/GROK.md` | 早期入口，已被 AGENTS.md 取代 |
| 用户设置 | `~/.grok/user-settings.json` | 用户级偏好（API key / model） |

`.grok/settings.json` 内含 `mcpServers` 字段（与 Claude Desktop canonical
格式兼容），v0.0.4 不主动写入：std-ai MCP 仅 dispatch 给 claude-code /
cursor / copilot 三家。

## 3. 文件格式

| 文件 | 格式 | frontmatter |
|---|---|---|
| `AGENTS.md` / `AGENTS.override.md` | Markdown | 无 frontmatter，纯指令文本 |
| `.grok/settings.json` | JSON | 不适用 |
| `~/.grok/user-settings.json` | JSON | 不适用 |

字符上限 UNKNOWN：3 个 fork 均未在公开文档明确单文件字节上限，实测
单 AGENTS.md 50k+ 仍能加载。

## 4. std-ai 四类映射

| std-ai 类型 | grok-cli 落点 | 加载方式 |
|---|---|---|
| rules | inline 到 `<repo>/AGENTS.md` 主体（grok 无子目录 rules 入口） | 启动时自动 merge |
| skills | fallback `<repo>/.grok/rules/skills/<name>/SKILL.md`（Agent Skills 标准包形式） | grok 不原生扫描 skills，作为 std-ai 私有 fallback 容器，由 AI 通过 frontmatter `std-ai-type: skills` + body explainer 识别 |
| commands | fallback `<repo>/.grok/rules/commands/<name>.md` | grok 不识别 slash 命令，作为上下文文本读入 |
| references | fallback `<repo>/.grok/rules/references/<name>.md` | 同上 |
| subagents | fallback `<repo>/.grok/rules/subagents/<name>.md` | grok 无 subagent 概念，路径降级 |

## 5. 转换器实现要点

1. 复用 `protocol.AgentsMD`，`grokCLIAdapter` 配置 `RootFileName: "AGENTS.md"`
2. `RulesDir: ""` -> nonRoot rules inline 到 AGENTS.md 正文（与 amp / warp /
   crush 风格一致，而非 codex 的 `.codex/memories/` 子目录）
3. `SkillsAsRule: false`：RulesDir 为空时 SkillsAsRule=true 会把 skill 写到
   仓库根 `skill-<name>.md`，与 amp 行为对齐，改走 BuildDegradedSkillPackage
   落到 `.grok/rules/skills/<name>/SKILL.md`
4. `PerDirOverrideFileName: "AGENTS.override.md"`：标识 grok 的 per-dir
   override 文件名意图。**v0.0.4 暂未完整支持** AGENTS.override.md 的 per-dir
   合并行为（protocol 实现待 v0.0.5），当前 adapter 保留该字段以供后续协议
   层读取
5. commands / references / subagents 走 `protocol.BuildDegradedFileOp`，
   fallback 到 `.grok/rules/{commands,references,subagents}/<name>.md`，
   body 头注入 std-ai HTML 注释 explainer，frontmatter 写 `std-ai-type` 标识
6. `InjectTypeGlossary: true`：AGENTS.md 头部 prepend std-ai 类型速查段
7. 不写 `.grok/settings.json`：MCP 与 model 由用户在 grok CLI 内配置，
   std-ai 不接管运行时设置
8. 不写 `~/.grok/user-settings.json`：用户级配置不在 std-ai 写入范围
9. 不写 `.grok/GROK.md`：已被 AGENTS.md 取代，避免双源冲突

## 6. 信息来源

- https://github.com/superagent-ai/grok-cli （访问日期 2026-05-17）
- https://github.com/alphaonedev/grok-cli （访问日期 2026-05-17）
- /tmp/std-ai-protocol-research.md 第 36 行（调研日期 2026-05-17）

## 7. 已确认

- 3 个并行 fork（superagent-ai / alphaonedev / baba20o）主流未定
- **Fork chosen: superagent-ai**（最早实现），v0.0.5 可能 pivot 到其他 fork
- 读 `AGENTS.md`，支持 `AGENTS.override.md` per-dir 覆盖（与 codex 同名机制
  语义相同）
- `.grok/settings.json` 存项目级 MCP / model；`~/.grok/user-settings.json`
  存用户级偏好
- `.grok/GROK.md` 是早期入口，已被 AGENTS.md 取代

## 8. UNKNOWN

- 3 个 fork 中究竟哪个会成为主流（社区分裂状态）
- AGENTS.override.md 在 grok-cli 内的合并算法细节（是否完全替换父层 vs
  字段级覆盖），3 个 fork 实现可能不一致
- AGENTS.md / settings.json 单文件字节硬上限（公开文档未列）
- v0.0.4 暂未完整支持 AGENTS.override.md per-dir 行为（adapter
  PerDirOverrideFileName 字段保留意图，protocol 实现待 v0.0.5）
