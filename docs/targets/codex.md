# Target: Codex (OpenAI)

调研日期: 2026-05-07（2026-06-11 复核 memories / Team Config，废弃 .codex/memories 输出），2026-07-10 复核更新，2026-08-15 复核更新
官方文档: https://developers.openai.com/codex/ 与 https://learn.chatgpt.com/codex/agent-configuration/agents-md

## 1. 摘要

OpenAI Codex CLI 是 OpenAI 官方的命令行/IDE 插件 agent。配置围绕 `~/.codex/`
与项目级 `.codex/` 双层展开。`AGENTS.md` 是项目指令的主入口，子目录可放
`AGENTS.md` 与 `AGENTS.override.md` 实现 per-directory 覆盖。

Skills 走 OpenAI 通用规范 `$HOME/.agents/skills/` 与 `<repo>/.agents/skills/`
（注意是 `.agents` 而非 `.codex`，与官方 agent skills 协议一致）。

Codex **不消费** `CLAUDE.md` 或 `GEMINI.md`，但可通过 `project_doc_fallback_filenames`
显式声明 fallback。

2026-07 复核新增：Codex 官方已支持项目级 subagents `.codex/agents/<name>.toml`
（TOML 格式，https://developers.openai.com/codex/subagents，feature flag
`multi_agent` 默认启用）。transformer 现已实现该落点（见 # 6 / # 7）。

2026-08-15 复核：
- GitHub 上的 Codex Code Review 读取最接近改动的 `AGENTS.md` 中 `## Code Review Rules` 段（https://learn.chatgpt.com/codex/agent-configuration/agents-md）。stdagent 不自动生成该标题，由源正文自行书写。
- 仓库 skills 从 CWD 向上扫描每一级 `.agents/skills`。stdagent 仍写仓库根 `.agents/skills/`。
- 启动技能清单最多约占 context 2%（未知窗口时按 8000 字符），超限先缩短 description。
- `allow_implicit_invocation` 位于可选 `agents/openai.yaml`，不再作为 SKILL.md 核心字段。源 `disable_model_invocation: true` 时写出该 sidecar。
- `/codex/agent-configuration/rules` 是实验性 execpolicy `.rules`（Starlark `prefix_rule`），不是项目 coding rules，stdagent 不生成。

## 2. 配置文件路径

| 类别 | 路径 | 说明 |
|---|---|---|
| 用户 config | `~/.codex/config.toml` | `CODEX_HOME` 可改 |
| 系统 config | `/etc/codex/config.toml` | 可选 |
| 项目 config | `.codex/config.toml` | 沿 cwd 向上至 project root，closest wins，仅 trusted project 生效 |
| 用户 AGENTS | `~/.codex/AGENTS.md` + `AGENTS.override.md` | 单文件 + 可覆盖 |
| 项目 AGENTS | 项目根 + 任意子目录 `AGENTS.md` / `AGENTS.override.md` | 每目录最多取一对；链上拼接 root -> cwd（完整支持嵌套） |
| Prompts（已移除） | `~/.codex/prompts/*.md` | **自 0.117.0 彻底移除（非 deprecated）**，改用 skills |
| Subagents（项目） | `.codex/agents/<name>.toml` | 官方原生支持，字段 `name` / `description` / `developer_instructions`（必填）+ 可选 `model` / `sandbox_mode` 等 |
| Skills (用户) | `$HOME/.agents/skills/<name>/SKILL.md` | 通用 agent skills 协议 |
| Skills (项目) | `<repo>/.agents/skills/<name>/SKILL.md` | 与仓库一同提交；调用策略见同目录 `agents/openai.yaml` |
| Hooks | `~/.codex/hooks.json` 或 `[hooks]` inline；项目 `.codex/hooks.json` | 受 `[features].codex_hooks=true` 控制 |

## 3. 文件格式

| 文件 | 格式 | frontmatter |
|---|---|---|
| `config.toml` | TOML | 无 |
| `AGENTS.md` / `AGENTS.override.md` | Markdown | 无（全文即指令） |
| `SKILL.md` | Markdown | 必填，字段见 OpenAI agents skills 规范（核心: `name`、`description`） |
| `agents/openai.yaml` | YAML | 可选；`policy.allow_implicit_invocation`（默认 true） |
| `.codex/agents/<name>.toml` | TOML | `name` / `description` / `developer_instructions` 必填，`model` 等可选 |
| 自定义 prompt（已移除） | Markdown | 不再适用 |
| `hooks.json` | JSON | 无 |

`AGENTS.md` 单文件大小受 `project_doc_max_bytes` 限制（默认 32768 字节，见 # 10）。

## 4. config.toml 关键字段

```toml
project_root_markers = [".git"]
project_doc_max_bytes = 32768
project_doc_fallback_filenames = ["AGENTS.md", "CLAUDE.md"]

[features]
codex_hooks = true
memories = true
multi_agent = true

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

closest wins：越靠近 cwd 的优先级越高，向 root 拼接（链上完整支持嵌套 AGENTS.md）。

## 6. std-agent 五类映射（实际实现，`internal/transformer/codex.go`）

| std-agent 类型 | Codex 落点 |
|---|---|
| rules | 项目 `AGENTS.md`（RulesDir 留空，所有 nonRoot rules 全文 inline 到一个文件）；子目录 rules 写入 `<sub>/AGENTS.md`（nested root，无 manifest） |
| skills | `<repo>/.agents/skills/<name>/SKILL.md` + 同目录辅助文件；frontmatter 白名单 `name` / `description` / `license` / `compatibility` / `metadata`。`disable_model_invocation: true` 时另写 `agents/openai.yaml` |
| commands | 降级为 skill 写到 `.agents/skills/commands/<n>/SKILL.md`（v3 子目录隔离，无私有前缀），description 含 slash 调用 hint；不进入 shared `AGENTS.md` |
| references | `.agents/references/<n>.md`（降级，AI 按 frontmatter `std-agent-type` 识别） |
| subagents | **双写**：`.agents/subagents/<n>.md`（降级 markdown，人读文档）+ `.codex/agents/<n>.toml`（官方原生格式，`name` / `description` / `developer_instructions` / 可选 `model`） |

**禁止落点 `.codex/` 的例外**：项目级 `.codex/` 是官方 Team Config 配置目录
（`config.toml` / `rules/*.rules` execpolicy 命令权限 / `skills/`），且被沙箱与
`.git` 同级 carveout 保护，历史上 stdagent 不写任何 markdown 类文件进去。
**唯一例外是 `.codex/agents/*.toml`**：官方文档明确这是 subagent 原生格式，
2026-07 起 transformer 主动写入（决策点 C：TOML 是否受 trusted-project 边界限制
官方未明说，见 # 9 剩余 UNKNOWN，写盘本身无害，是否被消费是 Codex 侧行为）。
曾用的 `.codex/memories/` 撞官方 memories 概念
（`~/.codex/memories/` 用户级自动记忆，详见第 9 节），已废弃，runner 的
`legacyCodexMemoriesOrphans` 在 sync 时自动清理带 stdagent marker 的旧产物。

## 7. 转换器实现要点

1. 主输出：项目根 `AGENTS.md`，由 `inject` footer 标识为 stdagent 生成
2. rules 拼接策略：RulesDir 为空，所有 nonRoot rules 按 `name` 排序全文拼接到
   `AGENTS.md` 正文（`transformerutil.JoinAGENTSStyle`）
3. AGENTS.md 总字节接近 / 超过 `project_doc_max_bytes`（32768 字节）时不自动
   分拆，由 `budget.CheckTotalRules` 输出 HARD WARN 提醒精简或对低优先级 rule
   关闭 codex target；语义是**项目链累计和**（root -> cwd 整条链，含 nested AGENTS.md
   累计），超限按链序**文件粒度停止追加（链尾先丢）**，该数值是 `config.toml`
   可调默认值而非硬上限
4. skills：写入 `.agents/skills/<name>/SKILL.md` + 辅助文件；frontmatter 至少
   含 `name`、`description`
5. subagents：`buildCodexAgentTOML`（`internal/transformer/codex.go:47`）渲染
   `.codex/agents/<name>.toml`，`developer_instructions` 用 TOML 三引号多行
   literal string；同时保留 `.agents/subagents/<n>.md` 降级产物过渡
6. trust 提示：sync 时检测 `.codex/` 已存在却未 trusted，输出 WARN
7. v1.0 不写 hooks.json；`.codex/config.toml` 不由 stdagent 生成/合并

## 8. 信息来源

- https://developers.openai.com/codex/concepts/customization
- https://developers.openai.com/codex/guides/agents-md
- https://developers.openai.com/codex/cli/slash-commands
- https://developers.openai.com/codex/cli/features
- https://developers.openai.com/codex/config-basic
- https://developers.openai.com/codex/config-reference
- https://developers.openai.com/codex/mcp
- https://developers.openai.com/codex/subagents
- https://github.com/openai/codex

## 9. 已确认与剩余 UNKNOWN（2026-07-10 复核）

已确认：
- `[[skills.config]]` schema：TOML 数组表，字段 `path`（string，指向 SKILL.md）
  + `enabled`（bool）。**不自动扫描**，仅显式注册；用途是"临时禁用某个 skill"
  或"注册非默认路径"
- 子代理（subagents）省略 `[[skills.config]]` 时继承父会话配置；修改 `~/.codex/config.toml`
  后必须重启 Codex
- skills 文档正确路径 https://developers.openai.com/codex/skills（之前 `/concepts/skills` 为 404）
- v1.0 不主动写 `[[skills.config]]`，stdagent 仅落 SKILL.md 到 `.agents/skills/`，
  由 Codex 默认机制发现
- `~/.codex/prompts` **自 0.117.0 彻底移除**（不是 deprecated，是已删除），commands
  降级为 skills 的方向正确，本次仅更新措辞
- `.codex/agents/<name>.toml` 官方原生 subagent 格式已实现落地写入（见 # 6 / # 7）
- `allow_implicit_invocation` 写在 `<skill>/agents/openai.yaml`，不写进 SKILL.md

2026-06-11 复核确认（memories 落盘格式不再 UNKNOWN）：
- Memories 是**用户级**自动记忆系统：单根 `~/.codex/memories/`
  （`memory_root = codex_home.join("memories")`，read / write crate 同一定义），
  mem v2 迁移（openai/codex PR #11366，2026-02）明确删除 per-cwd 记忆桶。
  内容（`raw_memories.md` / `memory_summary.md` / `rollout_summaries/`）由
  Codex 后台从历史会话提取生成，带 git baseline 与 DB 管理；官方文档明示
  "Treat these files as generated state... don't rely on editing them by hand"。
  feature 默认关闭（`[features] memories = true` 才启用），EEA / UK / 瑞士不可用。
  **不存在项目级 memories 目录**，`<repo>/.codex/memories/` 不被任何 Codex
  机制读取。
- Team Config：项目级 `.codex/` 官方承载 `config.toml`、`rules/`（Starlark
  `.rules` execpolicy 命令权限，仅 trusted 项目加载）、`skills/`；沿 cwd ->
  parent -> repo root -> `~/.codex/` -> `/etc/codex/` 分层，高优先级覆盖低。
  来源：developers.openai.com/codex/memories、/codex/rules、/codex/changelog

剩余 UNKNOWN（2026-07 复核仍未证实）：
- `.codex/agents/*.toml` 的 trusted-project 边界（是否受沙箱 carveout 限制未知）
- Codex 全局 `~/.codex/AGENTS.md` 是否计入 `project_doc_max_bytes` 链累计口径
- SKILL.md 与 `description` 的字节大小上限（官方未给数值，budget.go 未设该 target 的 Hard）
- 自定义 prompt 完全废弃后官方推荐 skills 的最小字段集
