# Target: Kimi Code (Moonshot AI)

调研日期: 2026-07-17
官方仓库: https://github.com/MoonshotAI/kimi-code
官方文档: https://moonshotai.github.io/kimi-code/

## 1. 摘要

Kimi Code CLI 是月之暗面（Moonshot AI）的 AI 编码 CLI，npm 包
`@moonshot-ai/kimi-code`（TypeScript/Node），是 Python 版 kimi-cli
（https://github.com/MoonshotAI/kimi-cli）的继任产品。kimi-cli README 明确
"Kimi CLI is evolving into Kimi Code CLI"，旧产品逐步下线但仍在发版
（2026-07-16 最新 1.49.0）；kimi-code 当前版本 0.26.0（2026-07-16），
40+ 贡献者，活跃维护。安装 kimi-code 会自动迁移 kimi-cli 的配置与 session。

std-agent 以新架构 kimi-code（`.kimi-code/` 目录体系）为落点依据，
legacy `~/.kimi/` 布局仅作兼容参考，target 命名用 `kimi-code`。

## 2. 配置文件路径

| 类别 | 路径 | 说明 |
|---|---|---|
| 全局配置目录 | `$KIMI_CODE_HOME`，默认 `~/.kimi-code/` | config.toml / tui.toml / AGENTS.md / mcp.json / skills/ / plugins/ |
| 全局指令 | `~/.kimi-code/AGENTS.md` | 用户级；另读跨工具通用层 `~/.agents/AGENTS.md` |
| 项目指令 | `<repo>/AGENTS.md` | 同目录内 `.kimi-code/AGENTS.md` 优先于 `AGENTS.md`；大写胜小写 `agents.md` |
| 嵌套指令 | `<repo>/<subdir>/AGENTS.md` | 层级发现：从项目根（向上找最近 `.git`）到 cwd 逐级加载合并，leaf-first 分配 32KiB 预算 |
| 项目 skills | `.kimi-code/skills/`、`.agents/skills/` | 两个 Project 层扫描路径都原生支持 |
| 用户 skills | `~/.kimi-code/skills/`、`~/.agents/skills/` | User 层；另有 config.toml `extra_skill_dirs` |
| 项目 MCP | `.kimi-code/mcp.json` | `mcpServers` 顶层 key，同名覆盖用户级 `~/.kimi-code/mcp.json` |
| machine-local | `.kimi-code/local.toml` | 官方建议 gitignore |

`KIMI.md` 已废弃（迁移到 AGENTS.md）。

## 3. 五类概念支持度

| std-agent 类型 | Kimi Code 原生概念 | 结论 |
|---|---|---|
| rules | 无独立 rules 目录。社区提案 Issue #1747（Three-tier Rules System，含 paths/priority frontmatter）截至 2026-07-17 仍 open 未实现 | 无条件加载能力，nonRoot rules 只能 inline 进 AGENTS.md |
| skills | 原生，遵循 Agent Skills 开放标准，目录形式 `<skillsDir>/<name>/SKILL.md`（也支持 flat 单 .md）。frontmatter：`name`（必填，1-64 字符小写字母数字连字符）/ `description`（必填，1-1024 字符）/ `type`（prompt 默认 / inline / flow）/ `whenToUse` / `disableModelInvocation` / `arguments` / `license` / `compatibility` / `metadata`。**无 `allowed-tools` 字段**（与 Claude Code 不同） | 原生支持 |
| commands | 无独立自定义 commands 机制。Skill 自动注册为 `/skill:<name>`（不冲突时 `/<name>`）；`type: flow` 的 skill 走 `/flow:<name>` | 降级为 skill |
| references | 无原生概念 | 降级 |
| subagents | 内置 coder / explore / plan 三个。自定义走 YAML agent spec 文件 + `kimi --agent-file <path>` 或 `--agents-dir <dir>`，**非项目内自动发现目录**；字段体系（`tools: module:ClassName`、`system_prompt_path`、`extend`）与 markdown frontmatter 完全不兼容 | 降级（机械转换成本高，不做 YAML 转换器） |

## 4. std-agent 五类映射（实际实现，`internal/transformer/kimi_code.go`）

| std-agent 类型 | Kimi Code 落点 | 加载方式 |
|---|---|---|
| rules（root） | 项目根 `AGENTS.md` | 层级发现自动加载 |
| rules（nonRoot） | inline 到 `AGENTS.md`（无 RulesDir） | 随根文件加载 |
| skills | `.agents/skills/<name>/SKILL.md`（与 codex / amp / warp / antigravity 共享落点） | 官方 Project 层扫描路径，AI 按 description 触发或 `/skill:<name>` |
| commands | `.agents/skills/commands/<name>/SKILL.md`（降级 skill，与 codex / amp 同模式） | `/skill:<name>` 触发 |
| references | fallback `.kimi-code/rules/references/<name>.md` | std-agent 降级 |
| subagents | `.kimi-code/agents/<name>.md` | 官方自动发现目录 |

嵌套 root（源文档带 `NestedPath`）写到对应子目录的 `AGENTS.md`（kimi-code
层级发现原生消费），不带 manifest，与 codex / amp / warp 一致。

## 5. 转换器实现要点

1. **`.agents/skills/` 共享字节一致铁律扩到五家**：kimi-code 官方把
   `.agents/skills/` 列为 Project 层扫描路径，与 codex / amp / warp /
   antigravity 共享落点。`SkillSupportedFields` 必须同集
   （`name / description / license / compatibility / metadata`），改任一家
   五家同步。防回归：`TestKimiCodeCodexSkillByteIdentical`。
2. 不写私有 `.kimi-code/skills/` 副本：两个扫描路径都原生，写一份共享的即可
   （writer unchanged 去重）；kimi 特有 skill 字段（`type` / `whenToUse`
   frontmatter）std 源格式没有对应输入，无字段损失。
3. `FallbackDir=".kimi-code/rules"` 必须私有：degraded 文件的 explainer HTML
   注释含 target 名，落共享 `.agents/` 会与 codex 的降级产物互相改写
   （flip-flop）。`.kimi-code/` 下官方只读 AGENTS.md / skills/ / mcp.json /
   local.toml，`rules/` 子目录不会被自动加载，放低频降级物安全。
4. 根 `AGENTS.md` 渲染与 amp / warp / jules 同风格（无 commands inject、无
   manifest），字节一致不加重多 target 共写 AGENTS.md 的已知 flip-flop wart。
5. subagents 不做 YAML agent spec 转换：`--agent-file` 无自动发现目录，
   字段体系不兼容，机械转换收益低；走形态 A 路径降级。
6. MCP 暂不接：目标文件 `.kimi-code/mcp.json`（顶层 key `mcpServers`，字段
   与 Claude Code `.mcp.json` 高度相似，可选字段另有 `bearerTokenEnvVar` /
   `enabledTools` / `startupTimeoutMs` 等），未来接入可复用现有 JSONMerge。

## 6. 信息来源

- https://github.com/MoonshotAI/kimi-code （访问日期 2026-07-17）
- https://github.com/MoonshotAI/kimi-cli （legacy，README 声明代际更替）
- https://www.npmjs.com/package/@moonshot-ai/kimi-code
- https://moonshotai.github.io/kimi-code/en/customization/agents
- https://moonshotai.github.io/kimi-code/en/customization/skills.html
- https://moonshotai.github.io/kimi-code/en/customization/mcp
- https://moonshotai.github.io/kimi-code/en/configuration/data-locations.html
- https://www.kimi.com/code/docs/en/kimi-code-cli/reference/slash-commands.html
- https://github.com/MoonshotAI/kimi-cli/issues/1747 （rules 提案，open）
- https://github.com/MoonshotAI/kimi-code/pull/523 （--agents-dir）

## 7. 已确认（2026-07-17）

- kimi-code 是当前主推产品，kimi-cli 逐步下线，配置自动迁移
- 根文件 AGENTS.md，`.kimi-code/AGENTS.md` 同目录优先；KIMI.md 已废弃
- 层级发现从 `.git` 项目根到 cwd 逐级合并，leaf-first 32KiB 预算
- skills 原生扫描 `.kimi-code/skills/` 与 `.agents/skills/` 两个 Project 层路径
- SKILL.md frontmatter 无 `allowed-tools` 字段
- 无独立 commands 机制（skill 自动注册 slash command）、无 rules 目录
- 自定义 subagents 走 YAML agent spec + CLI flag，无项目内自动发现目录

## 8. UNKNOWN

- AGENTS.md 是否在 session 启动时全量注入系统提示存在社区反馈与官方文档出入
  （Issue https://github.com/MoonshotAI/kimi-cli/issues/850，维护者回复有分歧），
  以最新版本实际行为为准
- 32KiB AGENTS.md 预算超限时的截断顺序细节（leaf-first 分配的具体算法）
- `~/.kimi-code/config.toml` 完整字段集（仅确认 `extra_skill_dirs` 等零散字段）
