# std-agent

![std-agent：23 个 AI CLI 工具的唯一事实来源](docs/assets/hero.png)

[![Release](https://img.shields.io/github/v/release/StringKe/std-agent?sort=semver)](https://github.com/StringKe/std-agent/releases)
[![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go)](https://go.dev/)
[![Go Report Card](https://goreportcard.com/badge/github.com/StringKe/std-agent)](https://goreportcard.com/report/github.com/StringKe/std-agent)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![CI](https://github.com/StringKe/std-agent/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/StringKe/std-agent/actions/workflows/ci.yml)

[English](README.md) | **简体中文** | [繁體中文](README.zh-TW.md) | [日本語](README.ja.md) | [한국어](README.ko.md) | [Русский](README.ru.md) | [Français](README.fr.md) | [Deutsch](README.de.md) | [Español](README.es.md) | [Português](README.pt-BR.md)

---

`stdagent` 是一个轻量的纯 Go CLI 工具。它把项目的 AI 配置维护在单一的 `.stdai/` 目录中作为唯一事实来源，再扩散到 **23 个 AI CLI 工具**，各工具的原生文件格式、frontmatter 方言和各种限制都已经替你处理好。

不要再手动维护 `CLAUDE.md`、`AGENTS.md`、`GEMINI.md`、`.cursor/rules/`、`.windsurf/rules/`、`.clinerules/`、`.github/copilot-instructions.md` 等文件了。改一次，处处生效。

## 为什么选 std-agent

- **单一来源**：用 YAML frontmatter + Markdown 一次性写好 `rules` / `skills` / `commands` / `references` / `subagents`。
- **二十三个目标**：Claude Code、Codex、Cursor、GitHub Copilot、Windsurf/Devin、Gemini CLI、Aider、Cline、OpenCode、Roo Code、Crush、Amp、Warp、Factory、Continue.dev、Antigravity、Qwen Code、Pi、Kilo Code、Augment Code、Jules、Grok Build、Kimi Code。
- **规范精确**：每个输出路径、frontmatter 方言、体积上限都对照各工具的官方文档核实过（最近一次全面审查：2026-07）；凡是原生支持 Agent Skills 目录的工具，都直接落在原生目录下。
- **零锁定**：writer 只触碰一小份路径白名单；每次 sync 前自动备份；`clean` 一键还原全部改动。
- **drift 检测**：`status` 显示被外部修改过的文件，`fix` 重新应用源文件。
- **MCP**：单一 `.stdai/standards/mcp.json` 扩散到 `.mcp.json` / `.cursor/mcp.json` / `.vscode/mcp.json`。
- **monorepo 感知**：配置查找从 `cwd` 向上遍历，任意子目录下执行都没问题。
- **自我升级**：`stdagent upgrade` 从 GitHub Releases 拉取已签名的发行版，做 sha256 校验和原子替换。

## 支持的工具

### Tier 1（14 个）

| 目标 | 主要输出 |
|---|---|
| Claude Code (Anthropic) | `CLAUDE.md` + `.claude/{rules,skills,commands,agents}/` + `.mcp.json` |
| Codex (OpenAI) | `AGENTS.md` + `.agents/skills/` + `.codex/agents/*.toml`（原生 subagents） |
| Cursor | `.cursor/{rules/*.mdc,skills,commands,agents}/` + `.cursor/mcp.json` |
| GitHub Copilot | `.github/{copilot-instructions,instructions,prompts,agents,skills}/` + `.vscode/mcp.json` |
| Windsurf / Devin (Cognition) | `.windsurf/{rules,skills,workflows}/` + `.devin/rules/` 镜像 |
| Gemini CLI (Google) | `GEMINI.md` + `.gemini/skills/` + `.gemini/commands/*.toml` |
| Aider | 复用 `AGENTS.md`（noop） |
| Cline | `.clinerules/`（100/500/900 数字前缀） |
| OpenCode | `.opencode/{skills,commands}/` |
| Roo Code | `.roo/{rules,skills,commands}/` |
| Crush (Charmbracelet) | `CRUSH.md` + `.crush/skills/` + `crush.json` skills 注册 |
| Amp (Sourcegraph) | `AGENTS.md`（inline） + `.agents/skills/` |
| Warp | `AGENTS.md`（inline + nested） + `.agents/skills/` |
| Factory (Factory.ai) | `.factory/{rules,skills,commands,droids}/` |

### Tier 2（9 个）

| 目标 | 主要输出 |
|---|---|
| Continue.dev | `.continue/{rules,skills,prompts}/` + 嵌套 `rules.md` |
| Antigravity (Google) | `.agents/{rules,skills,workflows}/` |
| Qwen Code (Alibaba) | `QWEN.md` + `.qwen/{rules,skills,commands}/` |
| Pi | `.pi/skills/` + `.pi/prompts/` |
| Kilo Code (kilo.ai) | `.kilo/{rules,skills,command}/` + `kilo.jsonc` instructions 注册 |
| Augment Code | `.augment/{rules,skills}/` |
| Jules (Google) | `AGENTS.md` |
| Grok Build (xAI) | `AGENTS.md` + `.grok/skills/` |
| Kimi Code (Moonshot AI) | `AGENTS.md` + `.agents/skills/` |

每个集成的详细说明都在 [docs/targets/](docs/targets/) 下。

## 快速开始

```bash
# Install (macOS / Linux)
curl -fsSL https://raw.githubusercontent.com/StringKe/std-agent/main/install.sh | sh

# Install (Windows PowerShell)
irm https://raw.githubusercontent.com/StringKe/std-agent/main/install.ps1 | iex

# Initialize in your project
cd your-project
stdagent init

# Edit .stdai/standards/rules/example.md, then sync to all enabled targets
stdagent sync

# Inspect / fix drift
stdagent status
stdagent fix
```

## 从已有项目迁移到 std-agent

项目里已经散落着 `CLAUDE.md` / `AGENTS.md` / `.cursor/rules/` / `.clinerules/` / `.github/copilot-instructions.md`？把下面这段提示词粘贴给 Claude Code / Codex / Cursor / Gemini CLI，它会帮你把一切都重新整理进 `.stdai/standards/`。

````text
Help me migrate this project from scattered AI configuration to std-agent. Please do:

1. Use Glob / Read to scan every existing AI rule file:
   - Root: CLAUDE.md AGENTS.md GEMINI.md .cursorrules .windsurfrules .clinerules
   - Subdirs: .claude/rules/ .claude/skills/ .claude/commands/ .claude/agents/
              .cursor/rules/ .windsurf/rules/ .clinerules/ .continue/rules/
              .github/copilot-instructions.md .github/instructions/
              .rulesync/rules/ .codex/AGENTS.md
   - In-repo nested CLAUDE.md (find . -name CLAUDE.md -not -path './.stdai/*')

2. Report an inventory: X rules / Y skills / Z commands / N nested CLAUDE.md,
   and flag which files contain "project overview" content.

3. Propose a split plan, wait for my approval, then write files:
   - Project overview (definition / stack / iron rules / maintenance flow)
     -> .stdai/standards/root.md
   - Each focused rule -> .stdai/standards/rules/<kebab-name>.md
   - Skill package -> .stdai/standards/skills/<name>/SKILL.md (with scripts/ references/ subdirs)
   - Slash commands -> .stdai/standards/commands/<name>.md
   - Nested CLAUDE.md -> .stdai/standards/nested/<relative-path>/root.md
   - Every file gets a frontmatter: type / name / description / priority / applyTo

4. No "refactoring" of original content. Keep every executable command, API endpoint,
   error string, file path, line number. Allowed "optimizations": drop filler words,
   merge duplicates, split oversized files, rename outdated tool names.

5. When done, tell me to run `stdagent sync` and remove legacy artifacts
   (.rulesync/, .cursorrules single-file, etc.). DO NOT delete the files stdagent
   itself produces (CLAUDE.md / AGENTS.md / .claude/rules/).

Full spec (frontmatter field table, root.md template, nested layout, rulesync mapping)
is in the `stdagent intro` command output.
````

也可以直接 pipe 给 LLM CLI：

```bash
stdagent intro | pbcopy            # macOS: paste into AI chat
stdagent intro --json              # for agent / automation integrations
```

## 命令

| 命令 | 用途 |
|---|---|
| `stdagent init` | 生成 `.stdai/` + `config.toml` + `.stdaiignore` + 示例 standards |
| `stdagent pull` | 更新缓存在 `.stdai/cache/` 里的 git 源 |
| `stdagent sync` | 核心：pull -> parse -> convert -> 扩散 |
| `stdagent fix` | 重新 sync 修复 drift（`sync` 的别名） |
| `stdagent status` | 各 target 的 drift 状态与上次 sync 时间 |
| `stdagent clean` | 删除生成的文件（保留 `.stdai/`） |
| `stdagent budget` | LLM context 预算检查（字符数 + token 估算） |
| `stdagent which <path>` | 列出适用于某文件的 rules / references（供 AI 按需加载上下文） |
| `stdagent explain` | 打印 std-agent 5 种类型（rules/skills/commands/references/subagents）的语义说明，供 AI 参考 |
| `stdagent intro` | 打印迁移提示词，供 LLM 转换现有配置 |
| `stdagent upgrade` | 从 GitHub Releases 自我升级（sha256 + 原子替换） |
| `stdagent version` | 构建信息 |

每个命令都支持 `--help`。完整参考：[docs/commands.md](docs/commands.md)。

## 基于 Protocol 的架构

v0.0.4 引入了三层 transformer 架构：每个 target 的 `Plan()` 委派给 6 个 protocol（AgentsMD / ClaudeMD / Cursor / Clinerules / WindsurfStyle / Copilot）之一，用一个 `protocol.Adapter` struct literal 参数化。新增一个工具现在只需要约 25-35 行代码，而不是 145 行（代码去重 60-70%）。

优雅降级：当某个 target 原生不支持 std-agent 的某种类型时（例如 references 在所有工具、subagents 在 amp / windsurf 上），stdagent 会退回到子目录隔离的路径（`<FallbackDir>/references/<name>.md`），带上 frontmatter `std-agent-type: <type>` 加 HTML 注释说明，不使用 std-agent 私有前缀。

## 源文件格式

完整 schema 见 [docs/spec.md](docs/spec.md) Part 1。最简形式：

```markdown
---
type: rules                       # rules | skills | commands | references
name: coding-style
description: General coding style
priority: high                    # high | normal | low
targets: [claude-code, codex]     # opt-in (or use exclude_targets to opt-out)
applyTo: ["**/*.go"]
alwaysApply: false
---

# Coding Style

Always use meaningful variable names...
```

MCP 服务器（`.stdai/standards/mcp.json`）：

```json
{
  "version": "1.0",
  "servers": {
    "github": { "type": "stdio", "command": "gh", "args": ["api"] },
    "linear": { "type": "http", "url": "https://mcp.linear.app/sse" }
  }
}
```

## 配置

`.stdai/config.toml`：

```toml
version = "1.0"
inject = true            # inject "Generated by stdagent" footer in outputs
inject_whatis = true     # add a one-line origin note inside skills
auto_pull = true         # pull git sources on every sync
backup = true
backup_keep = 5

[targets]
claude-code  = { enabled = true,  convert = true }
codex        = { enabled = true,  convert = true }
cursor       = { enabled = false, convert = true }
copilot      = { enabled = false, convert = true }
windsurf     = { enabled = false, convert = true }
gemini       = { enabled = false, convert = true }
aider        = { enabled = false, convert = true }
cline        = { enabled = false, convert = true }
opencode     = { enabled = false, convert = true }
continue-dev = { enabled = false, convert = true }
antigravity  = { enabled = false, convert = true }

[sources.default]
url     = "https://github.com/your-org/ai-standards.git"
branch  = "main"
enabled = true
paths   = ["standards/"]
```

完整参考：[docs/config-spec.md](docs/config-spec.md)。

## 项目布局

```
your-project/
├── .stdai/                    内部管理区（唯一事实来源）
│   ├── config.toml            唯一的配置文件
│   ├── standards/             编写根目录
│   │   ├── rules/
│   │   ├── skills/
│   │   ├── commands/
│   │   ├── references/
│   │   └── mcp.json           MCP 服务器（可选）
│   ├── cache/                 git 源缓存
│   ├── backups/               每次 sync 前自动快照
│   └── state.json             运行时状态
├── .stdaiignore               gitignore 风格的 glob，用于排除源文件
├── CLAUDE.md                  扩散目标：Claude Code
├── AGENTS.md                  扩散目标：Codex / Cursor fallback / Copilot agent / OpenCode / Antigravity
├── GEMINI.md                  扩散目标：Gemini CLI
├── .mcp.json                  Claude 的 MCP 配置
└── .claude/ .codex/ .cursor/ .github/ .windsurf/ .gemini/ .clinerules/ .opencode/ .continue/ .agents/
```

详见 [docs/file-structure.md](docs/file-structure.md)。

## Monorepo 支持

不显式指定 `--config` 时，`stdagent` 会从 `cwd` 向上遍历，找到最近的 `.stdai/config.toml`。从任意子目录运行都能自动定位到 monorepo 根目录。

## 开发

```bash
# Toolchain (mise + go + golangci-lint + gofumpt + git-cliff)
mise install

# Common tasks
mise run fmt        # gofumpt + goimports
mise run lint       # golangci-lint
mise run test       # go test -race -cover
mise run check      # fmt + lint + test in one go
mise run build      # produces bin/stdagent
mise run run        # go run ./cmd/stdagent
```

## 文档

- **[docs/spec.md](docs/spec.md)**：完整规范，std-agent 标准 + 23 个工具的差异 + 转换策略
- [docs/prd.md](docs/prd.md)：产品需求
- [docs/architecture.md](docs/architecture.md)：模块布局与数据流
- [docs/commands.md](docs/commands.md)：CLI 命令参考
- [docs/conversion-rules.md](docs/conversion-rules.md)：转换矩阵 + frontmatter 字段映射
- [docs/format-spec.md](docs/format-spec.md)：frontmatter schema 细节
- [docs/file-structure.md](docs/file-structure.md)：目录结构约定
- [docs/roadmap.md](docs/roadmap.md)：路线图
- [docs/targets/](docs/targets/)：各工具的调研笔记

## License

MIT，详见 [LICENSE](LICENSE)。
