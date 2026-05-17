# std-agent

![std-agent：一份真相源，11 个 AI CLI 工具](docs/assets/hero.png)

[![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![CI](https://github.com/StringKe/std-ai/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/StringKe/std-ai/actions/workflows/ci.yml)

[English](README.md) | **简体中文** | [繁體中文](README.zh-TW.md) | [日本語](README.ja.md) | [한국어](README.ko.md) | [Русский](README.ru.md) | [Français](README.fr.md) | [Deutsch](README.de.md) | [Español](README.es.md) | [Português](README.pt-BR.md)

---

`stdagent` 是一个轻量、纯 Go 实现的 CLI 工具。它把 `.stdai/` 作为项目里 AI 配置的唯一真相源，再扩散到 **11 个 AI CLI 工具**——每个工具的原生文件格式、frontmatter 方言、各种坑都已经替你处理好。

不要再手维护 `CLAUDE.md`、`AGENTS.md`、`GEMINI.md`、`.cursor/rules/`、`.windsurf/rules/`、`.clinerules/`、`.github/copilot-instructions.md`...... 写一次，处处生效。

## 为什么是 std-agent？

- **单一来源** — `rules` / `skills` / `commands` / `references` 用 YAML frontmatter + Markdown 写一次。
- **11 个目标** — Claude Code、Codex、Cursor、GitHub Copilot、Windsurf、Gemini CLI、Aider、Cline、OpenCode、Continue.dev、Antigravity。
- **零锁定** — writer 只动白名单内的少数路径；每次 sync 前自动备份；`clean` 一键清回去。
- **drift 检测** — `status` 列出被外部改过的文件，`fix` 重新落盘。
- **MCP** — 单文件 `.stdai/standards/mcp.json` 扩散到 `.mcp.json` / `.cursor/mcp.json` / `.vscode/mcp.json`。
- **monorepo 友好** — 配置查找从 cwd 向上 walk，从任何子目录都能跑。
- **自我升级** — `stdagent upgrade` 从 GitHub Releases 拉签名归档，sha256 校验 + 原子替换。

## 支持的目标工具

### Tier 1（9 个）

| 目标 | 主要落点 |
|---|---|
| Claude Code (Anthropic) | `CLAUDE.md` + `.claude/{rules,skills,commands}/` + `.mcp.json` |
| Codex (OpenAI) | `AGENTS.md` + `.agents/skills/` + `.codex/rules/`（超字节 spillover） |
| Cursor | `.cursor/{rules/*.mdc,skills,commands}/` + `.cursor/mcp.json` |
| GitHub Copilot | `.github/{copilot-instructions,instructions,prompts,agents}/` + `.vscode/mcp.json` |
| Windsurf (Codeium) | `.windsurf/{rules,skills,workflows}/` |
| Gemini CLI (Google) | `GEMINI.md` + `.gemini/commands/*.toml` |
| Aider | 复用 `AGENTS.md`（noop） |
| Cline | `.clinerules/` + `.clinerules/workflows/` |
| OpenCode | `.opencode/{agents,commands}/` |

### Tier 2（2 个）

| 目标 | 主要落点 |
|---|---|
| Continue.dev | `.continue/{rules,prompts}/` |
| Antigravity (Google) | `.agents/{rules,workflows}/` |

每个集成的调研都在 [docs/targets/](docs/targets/) 下。

## 快速开始

```bash
# 安装（macOS / Linux）
curl -fsSL https://raw.githubusercontent.com/StringKe/std-ai/main/install.sh | sh

# 安装（Windows PowerShell）
irm https://raw.githubusercontent.com/StringKe/std-ai/main/install.ps1 | iex

# 在你的项目里初始化
cd your-project
stdagent init

# 编辑 .stdai/standards/rules/example.md，再同步到所有 enabled targets
stdagent sync

# 检查 / 修复 drift
stdagent status
stdagent fix
```

## 从已有项目迁移到 std-ai

项目里已经散落 `CLAUDE.md` / `AGENTS.md` / `.cursor/rules/` / `.clinerules/` / `.github/copilot-instructions.md` 等？把下面这段提示词直接发给 Claude Code / Codex / Cursor / Gemini CLI，它会替你重组成 `.stdai/standards/` 结构。

````text
帮我把当前项目从分散的 AI 配置迁移到 std-agent 统一管理。请按以下步骤执行：

1. 用 Glob / Read 扫项目里所有现存 AI 规则文件：
   - 根目录：CLAUDE.md AGENTS.md GEMINI.md .cursorrules .windsurfrules .clinerules
   - 子目录：.claude/rules/ .claude/skills/ .claude/commands/ .claude/agents/
              .cursor/rules/ .windsurf/rules/ .clinerules/ .continue/rules/
              .github/copilot-instructions.md .github/instructions/
              .rulesync/rules/ .codex/AGENTS.md
   - 同仓库内嵌套 CLAUDE.md（find . -name CLAUDE.md -not -path './.stdai/*'）

2. 给我一份盘点：发现 X 个 rules / Y 个 skills / Z 个 commands / N 个嵌套 CLAUDE.md，
   并指出哪些文件含"项目总览"内容。

3. 提出拆分方案后等我确认，然后写文件：
   - 项目总览（定义、技术栈、铁律、维护流程）-> .stdai/standards/root.md
   - 每条聚焦规则 -> .stdai/standards/rules/<kebab-name>.md
   - skill 能力包 -> .stdai/standards/skills/<name>/SKILL.md（含子目录 scripts/ references/）
   - slash 命令模板 -> .stdai/standards/commands/<name>.md
   - 嵌套 CLAUDE.md -> .stdai/standards/nested/<相对路径>/root.md
   - 每个文件加 frontmatter：type / name / description / priority / applyTo

4. 严禁"重构"原文：保留所有可执行命令、API 端点、错误字符串、文件路径行号；
   允许"优化"：删过渡词、合并重复段落、拆大文件、改写过时工具名。

5. 写完告诉我跑 `stdagent sync`，再删除旧产物（.rulesync/ / .cursorrules 等单文件版）。
   不要删 stdagent 生成的 CLAUDE.md / AGENTS.md / .claude/rules/。

完整规范（含 frontmatter 字段表、root.md 模板、嵌套约定、rulesync 迁移映射）见
`stdagent intro` 命令输出。
````

也可以直接 pipe 给 LLM CLI：

```bash
stdagent intro | pbcopy            # macOS：复制到剪贴板，再粘到 AI 对话
stdagent intro --json              # 给 agent / 自动化集成
```

## 命令清单

| 命令 | 作用 |
|---|---|
| `stdagent init` | 创建 `.stdai/` + `config.toml` + `.stdaiignore` + 示例 standards |
| `stdagent pull` | 拉取 `.stdai/cache/` 中所有 enabled git 源 |
| `stdagent sync` | 核心：pull → parse → convert → 向外扩散 |
| `stdagent fix` | 重新 sync 修复 drift（`sync` 的语义别名） |
| `stdagent status` | 各 target 的 drift 与最后同步时间 |
| `stdagent clean` | 清空生成文件（保留 `.stdai/`） |
| `stdagent budget` | LLM context 预算检查（字符 + token 估算） |
| `stdagent intro` | 输出迁移提示词（喂给 AI 助手把现有配置转 std 格式） |
| `stdagent upgrade` | 从 GitHub Releases 自我升级（sha256 + 原子替换） |
| `stdagent version` | 构建信息 |

每个命令都支持 `--help`。完整参考：[docs/commands.md](docs/commands.md)。

## 源文件格式

完整 schema 见 [docs/spec.md](docs/spec.md) Part 1。最简形式：

```markdown
---
type: rules                       # rules | skills | commands | references
name: coding-style
description: 通用编码风格
priority: high                    # high | normal | low
targets: [claude-code, codex]     # 显式启用（或用 exclude_targets 排除）
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

## 配置示例

`.stdai/config.toml`：

```toml
version = "1.0"
inject = true            # 输出文件注入 "Generated by stdagent" footer
inject_whatis = true     # skills 内嵌一行来源注释
auto_pull = true         # 每次 sync 自动 pull git 源
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
├── .stdai/                    内部管理区（唯一真相源）
│   ├── config.toml            唯一配置文件
│   ├── standards/             编写根目录
│   │   ├── rules/
│   │   ├── skills/
│   │   ├── commands/
│   │   ├── references/
│   │   └── mcp.json           MCP 服务器（可选）
│   ├── cache/                 git 源缓存
│   ├── backups/               每次 sync 前自动快照
│   └── state.json             运行时状态
├── .stdaiignore               gitignore 风格 glob，过滤参与 sync 的源文件
├── CLAUDE.md                  扩散：Claude Code
├── AGENTS.md                  扩散：Codex / Cursor fallback / Copilot agent / OpenCode / Antigravity
├── GEMINI.md                  扩散：Gemini CLI
├── .mcp.json                  Claude 的 MCP
└── .claude/ .codex/ .cursor/ .github/ .windsurf/ .gemini/ .clinerules/ .opencode/ .continue/ .agents/
```

详见 [docs/file-structure.md](docs/file-structure.md)。

## monorepo 支持

不显式 `--config` 时，`stdagent` 从 cwd 向上 walk 找最近的 `.stdai/config.toml`。从任意子目录跑都能自动定位 monorepo 项目根。

## 开发

```bash
# 工具链（mise + go + golangci-lint + gofumpt + git-cliff）
mise install

# 常用任务
mise run fmt        # gofumpt + goimports
mise run lint       # golangci-lint
mise run test       # go test -race -cover
mise run check      # fmt + lint + test 一键
mise run build      # 产出 bin/stdagent
mise run run        # go run ./cmd/stdagent
```

## 文档

- **[docs/spec.md](docs/spec.md)** — 完整 spec：std-ai 标准 + 11 工具差异 + 转换策略
- [docs/prd.md](docs/prd.md) — 产品需求
- [docs/architecture.md](docs/architecture.md) — 模块划分与数据流
- [docs/commands.md](docs/commands.md) — CLI 命令规范
- [docs/conversion-rules.md](docs/conversion-rules.md) — 转换矩阵 + frontmatter 字段映射
- [docs/format-spec.md](docs/format-spec.md) — frontmatter 详细 schema
- [docs/file-structure.md](docs/file-structure.md) — 目录结构原则
- [docs/roadmap.md](docs/roadmap.md) — 路线图
- [docs/targets/](docs/targets/) — 11 个目标工具调研

## License

MIT — 见 [LICENSE](LICENSE)。
