# Target: Grok Build (xAI 官方 CLI)

调研日期: 2026-05-17

## 一句话定义

xAI 官方 agentic CLI 编码代理（terminal-native coding agent），2026-05-14 早期 beta，仅 SuperGrok Heavy（$299/月）订户可用。底层 `grok-code-fast-1`（SWE-Bench Verified 70.8）和 `grok-4.3`。CLI 二进制名 `grok`，`grok-build` 是项目代号。

## 仓库与文档

- 官方产品页：https://x.ai/cli
- 发布公告：https://x.ai/news/grok-build-cli
- 官方文档：https://docs.x.ai/build/overview
- Skills/Plugins/Marketplaces：https://docs.x.ai/build/features/skills-plugins-marketplaces
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

## frontmatter / 文件格式

- **AGENTS.md**：纯 markdown，无 frontmatter（沿用 OpenAI/Codex 推广的通用约定）
- **SKILL.md**：YAML frontmatter，agentskills.io 标准字段 + Claude Code 扩展字段被兼容
- 字节限制：UNKNOWN

## 5 种 std-ai type 原生支持

| std-ai type | grok-build | 落点 |
|---|---|---|
| rules | 原生（AGENTS.md / .claude/rules/） | `AGENTS.md`（推荐）或 `.claude/rules/*.md` |
| skills | 原生 | `.grok/skills/<name>/SKILL.md`（大写）|
| commands | 部分原生 | skills + `user-invocable: true` 自动暴露为 `/<name>` slash |
| subagents | plugins 内 `agents/` | 项目级无独立路径；走 Claude Code 兼容 `.claude/agents/` |
| references | 无原生概念 | `.grok/rules/references/<name>.md` 子目录 fallback |

## 嵌套行为

- AGENTS.md 多层自动叠加：官方明确 "walked from cwd to the repo root"
- skills 向上 walk：`.grok/skills/` 也 walked up
- per-dir override：UNKNOWN（xAI 官方未提，三方 fork 的 `AGENTS.override.md` 不在官方约定）

## stdagent adapter（已实现）

```go
var grokBuildAdapter = protocol.Adapter{
    Name:                 "grok-build",
    RootFileName:         "AGENTS.md",
    ManifestSection:      "Reference Rules",
    NestedSupported:      true,
    RulesDir:             "",
    SkillsDir:            ".grok/skills",
    SkillSupportedFields: []string{"name", "description", "license",
                                   "compatibility", "metadata",
                                   "allowed-tools", "disable-model-invocation"},
    CommandsDir:          "",
    ReferencesDir:        "",
    SubagentsDir:         "",
    FallbackDir:          ".grok/rules",
    InjectExplainer:      true,
    InjectStdaiTypeField: true,
    InjectTypeGlossary:   true,
}
```

## v0.0.4 -> v0.0.5 修订

1. transformer rename `grok-cli` → `grok-build`
2. adapter 按 xAI 官方文档重写：`SkillsDir=".grok/skills"`，删除 `PerDirOverrideFileName`（三方 fork 概念）
3. `ValidTargets` / `Default().Targets` / `.stdai/config.toml` 同步

## UNKNOWN

- grok-build 公开 GitHub 仓库
- AGENTS.md / SKILL.md 字节硬上限
- 项目级独立 subagents 路径（仅 plugins 内 agents/ 文档化）
- references 是否有原生概念
- per-dir override 文件名是否官方支持
- 用户基数 / stars
- MCP 是否支持项目级 `.grok/config.toml`
