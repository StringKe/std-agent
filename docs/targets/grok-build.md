# Target: Grok Build (xAI 官方 CLI)

调研日期: 2026-05-17，2026-07-10 复核更新

## 一句话定义

xAI 官方 agentic CLI 编码代理（terminal-native coding agent），2026-05-14 早期 beta，仅 SuperGrok Heavy（$299/月）订户可用。底层 `grok-code-fast-1`（SWE-Bench Verified 70.8）和 `grok-4.3`。CLI 二进制名 `grok`，`grok-build` 是项目代号。

## 仓库与文档

- 官方产品页：https://x.ai/cli
- 发布公告：https://x.ai/news/grok-build-cli
- 官方文档：https://docs.x.ai/build/overview
- Skills/Plugins/Marketplaces：https://docs.x.ai/build/features/skills-plugins-marketplaces
- Project Rules：https://docs.x.ai/build/features/project-rules
- 安装：`curl -fsSL https://x.ai/cli/install.sh | bash`
- 公开 GitHub 仓库：UNKNOWN

## 与三方 fork 的关系

**完全独立**于以下社区 fork（v0.0.4 误把它们当作 grok-build）：

- `superagent-ai/grok-cli` - 用 `.grok/GROK.md` + `AGENTS.override.md`
- `alphaonedev/grok-cli` - JSON 配置 + sub-agents 列表

xAI 官方文档无 GROK.md / AGENTS.override.md 任何提及。v0.0.5 重做 grok-build adapter 对齐 xAI 官方。

## 项目级配置完整路径

项目根（cwd 向上 walk 到 git root 叠加）：

- `AGENTS.md` / `Agents.md` / `AGENT.md` - 主指令，多层 walk，**无 frontmatter**
- `CLAUDE.md` / `Claude.md` / `CLAUDE.local.md` - **官方明确零配置兼容 Claude Code**
- `.claude/rules/` - Claude Code 规则目录自动读取
- `.grok/skills/<name>/SKILL.md` - 项目级 skills（SKILL.md 必须大写）
- `.grok/plugins/` / `.grok/hooks/` / `.grok/config.toml`

用户级：

- `~/.grok/config.toml` - 主配置 TOML
- `~/.grok/skills/` / `~/.grok/plugins/` / `~/.grok/hooks/`

跨工具共享（agents.md 生态）：

- `~/.agents/skills/` - 多客户端共享 skills
- `~/.agents/commands/` - 多客户端共享 commands

2026-07 复核新增确认：**根文件叠加/重复注入**。grok-build 会同时读
`AGENTS.md` + `CLAUDE.md` + `.claude/rules/` + `.cursor/rules/` 多路，
若仓库内其他 target（claude-code / cursor）也各自写了自己的规则文件，
grok-build 会把它们全部叠加进 context，不做去重。stdagent 目前不主动
缓解该重复注入问题（仅在本文档记录，供用户评估是否精简某些 target 的 rules）。

## frontmatter / 文件格式

- **AGENTS.md**：纯 markdown，无 frontmatter（沿用 OpenAI/Codex 推广的通用约定）
- **SKILL.md**：YAML frontmatter，agentskills.io 标准字段 + Claude Code 扩展字段被兼容
- 字节限制：**官方明示 "no size cap"**（2026-07 复核确认，`docs.x.ai/build/features/project-rules`），旧文档"UNKNOWN"已解决

## 5 种 std-agent type 原生支持

| std-agent type | grok-build | 落点 |
|---|---|---|
| rules | 原生（AGENTS.md / .claude/rules/） | `AGENTS.md`（推荐，全部 nonRoot rule inline，无独立 RulesDir） |
| skills | 原生 | `.grok/skills/<name>/SKILL.md`（大写）|
| commands | 原生 `.grok/commands/<name>.md` | 官方 user-guide 扁平 slash markdown |
| subagents | plugins 内 `agents/` | 项目级无独立路径；降级到 `.grok/docs/subagents/<n>.md`（2026-07 修正落点，见下） |
| references | 无原生概念 | `.grok/docs/references/<n>.md`（2026-07 修正落点，见下） |

## 嵌套行为

- AGENTS.md 多层自动叠加：官方明确 "walked from cwd to the repo root"，属于完整支持嵌套（子树作用域）
- skills 向上 walk：`.grok/skills/` 也 walked up
- per-dir override：UNKNOWN（xAI 官方未提，三方 fork 的 `AGENTS.override.md` 不在官方约定）

## stdagent adapter（已实现，`internal/transformer/grok_build.go`）

```go
var grokBuildAdapter = protocol.Adapter{
    Name:                  "grok-build",
    RootFileName:          "AGENTS.md",
    ManifestSection:       "Reference Rules",
    NestedSupported:       true,
    RulesDir:              "",
    SkillsDir:             ".grok/skills",
    SkillSupportedFields:  []string{"name", "description", "license",
                                    "compatibility", "metadata",
                                    "allowed-tools", "disable-model-invocation"},
    CommandFormat:         protocol.CommandSkillPrefix,
    CommandsAsSkillSubdir: "commands",
    ReferencesDir:         "",
    SubagentsDir:          "",
    FallbackDir:           ".grok/docs",
    InjectExplainer:       true,
    InjectStdaiTypeField:  true,
    InjectTypeGlossary:    true,
}
```

**2026-07 关键修正（第二轮 P0）**：`FallbackDir` 从早期版本的 `.grok/rules`
改为 `.grok/docs`。原因：xAI 官方文档明确 `.grok/rules/` 是**每 session 全量
加载**的原生 rules 目录（与 Claude Code 的按需子目录不同），若把 references /
subagents 这类低频参考资料降级堆进 `.grok/rules/` 子目录，等于把它们
**永久注入 context**，违背 references"按需查阅"的语义。改到独立的
`.grok/docs/` 后，这些降级产物不会被 grok-build 的原生 rules 全量加载机制
捡到，只在需要时由 AI 主动读取。来源：https://docs.x.ai/build/features/project-rules

## v0.0.4 -> v0.0.5 修订（历史记录）

1. transformer rename `grok-cli` -> `grok-build`
2. adapter 按 xAI 官方文档重写：`SkillsDir=".grok/skills"`，删除 `PerDirOverrideFileName`（三方 fork 概念）
3. `ValidTargets` / `Default().Targets` / `.stdai/config.toml` 同步

## 2026-07 修订（本轮）

1. `FallbackDir` 由 `.grok/rules` 改为 `.grok/docs`，避免 references / subagents 降级物污染每 session 全量加载的原生 rules 目录（见上）
2. 确认字节限制：官方明示 "no size cap"，UNKNOWN 解除
3. 记录根文件叠加/重复注入问题：`AGENTS.md` + `CLAUDE.md` + `.claude/rules/` + `.cursor/rules/` 多路全读，无去重

## UNKNOWN（2026-07 复核）

- grok-build 公开 GitHub 仓库
- 项目级独立 subagents 路径（仅 plugins 内 agents/ 文档化）
- references 是否有原生概念
- per-dir override 文件名是否官方支持
- 用户基数 / stars
- MCP 是否支持项目级 `.grok/config.toml`
