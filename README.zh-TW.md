# std-agent

![std-agent：23 個 AI CLI 工具的唯一事實來源](docs/assets/hero.png)

[![Release](https://img.shields.io/github/v/release/StringKe/std-agent?sort=semver)](https://github.com/StringKe/std-agent/releases)
[![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go)](https://go.dev/)
[![Go Report Card](https://goreportcard.com/badge/github.com/StringKe/std-agent)](https://goreportcard.com/report/github.com/StringKe/std-agent)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![CI](https://github.com/StringKe/std-agent/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/StringKe/std-agent/actions/workflows/ci.yml)

[English](README.md) | [简体中文](README.zh-CN.md) | **繁體中文** | [日本語](README.ja.md) | [한국어](README.ko.md) | [Русский](README.ru.md) | [Français](README.fr.md) | [Deutsch](README.de.md) | [Español](README.es.md) | [Português](README.pt-BR.md)

---

`stdagent` 是一個輕量的純 Go CLI 工具。它把專案的 AI 設定維護在單一的 `.stdai/` 目錄中作為唯一事實來源，再擴散到 **23 個 AI CLI 工具**，各工具的原生檔案格式、frontmatter 方言與各種限制都已經替你處理好。

不要再手動維護 `CLAUDE.md`、`AGENTS.md`、`GEMINI.md`、`.cursor/rules/`、`.windsurf/rules/`、`.clinerules/`、`.github/copilot-instructions.md` 等檔案了。改一次，處處生效。

## 為什麼選 std-agent

- **單一來源**：用 YAML frontmatter + Markdown 一次性寫好 `rules` / `skills` / `commands` / `references` / `subagents`。
- **二十五個目標**：Claude Code、Codex、Cursor、GitHub Copilot、Windsurf/Devin、Gemini CLI、Aider、Cline、OpenCode、Roo Code、Crush、Amp、Warp、Factory、Continue.dev、Antigravity、Qwen Code、Pi、Kilo Code、Augment Code、Jules、Grok Build、Kimi Code、Kiro、Goose。
- **規格精確**：每個輸出路徑、frontmatter 方言、體積上限都對照各工具的官方文件核實過（最近一次全面審查：2026-07）；凡是原生支援 Agent Skills 目錄的工具，都直接落在原生目錄下。
- **零鎖定**：writer 只碰觸一小份路徑白名單；每次 sync 前自動備份；`clean` 一鍵還原所有改動。
- **drift 偵測**：`status` 顯示被外部修改過的檔案，`fix` 重新套用來源檔案。
- **MCP**：單一 `.stdai/standards/mcp.json` 擴散到 `.mcp.json` / `.cursor/mcp.json` / `.vscode/mcp.json`。
- **monorepo 感知**：設定搜尋從 `cwd` 向上尋找，任意子目錄下執行都沒問題。
- **自我升級**：`stdagent upgrade` 從 GitHub Releases 拉取已簽章的發行版，做 sha256 驗證與原子取代。

## 支援的工具

### Tier 1（14 個）

| 目標 | 主要輸出 |
|---|---|
| Claude Code (Anthropic) | `CLAUDE.md` + `.claude/{rules,skills,commands,agents}/` + `.mcp.json` |
| Codex (OpenAI) | `AGENTS.md` + `.agents/skills/` + `.codex/agents/*.toml`（原生 subagents） |
| Cursor | `.cursor/{rules/*.mdc,skills,commands,agents}/` + `.cursor/mcp.json` |
| GitHub Copilot | `.github/{copilot-instructions,instructions,prompts,agents,skills}/` + `.vscode/mcp.json` |
| Windsurf / Devin (Cognition) | `.windsurf/{rules,skills,workflows}/` + `.devin/rules/` 鏡像 |
| Gemini CLI (Google) | `GEMINI.md` + `.gemini/skills/` + `.gemini/commands/*.toml` |
| Aider | 重用 `AGENTS.md`（noop） |
| Cline | `.clinerules/`（100/500/900 數字前綴） |
| OpenCode | `.opencode/{skills,commands}/` |
| Roo Code | `.roo/{rules,skills,commands}/` |
| Crush (Charmbracelet) | `CRUSH.md` + `.crush/skills/` + `crush.json` skills 註冊 |
| Amp (Sourcegraph) | `AGENTS.md`（inline） + `.agents/skills/` |
| Warp | `AGENTS.md`（inline + nested） + `.agents/skills/` |
| Factory (Factory.ai) | `.factory/{rules,skills,commands,droids}/` |

### Tier 2（11 個）

| 目標 | 主要輸出 |
|---|---|
| Continue.dev | `.continue/{rules,skills,prompts}/` + 巢狀 `rules.md` |
| Antigravity (Google) | `.agents/{rules,skills,workflows}/` |
| Qwen Code (Alibaba) | `QWEN.md` + `.qwen/{rules,skills,commands}/` |
| Pi | `.pi/skills/` + `.pi/prompts/` |
| Kilo Code (kilo.ai) | `.kilo/{rules,skills,command}/` + `kilo.jsonc` instructions 註冊 |
| Augment Code | `.augment/{rules,skills}/` |
| Jules (Google) | `AGENTS.md` |
| Grok Build (xAI) | `AGENTS.md` + `.grok/skills/` |
| Kimi Code (Moonshot AI) | `AGENTS.md` + `.agents/skills/` |
| Kiro (AWS) | `AGENTS.md` + `.kiro/{steering,skills,agents}/` |
| Goose (AAIF) | `AGENTS.md` + `.agents/skills/` |

每個整合的詳細說明都在 [docs/targets/](docs/targets/) 下。

## 快速開始

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

## 從既有專案遷移到 std-agent

專案裡已經散落著 `CLAUDE.md` / `AGENTS.md` / `.cursor/rules/` / `.clinerules/` / `.github/copilot-instructions.md`？把下面這段提示詞貼給 Claude Code / Codex / Cursor / Gemini CLI，它會幫你把一切重新整理進 `.stdai/standards/`。

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

也可以直接 pipe 給 LLM CLI：

```bash
stdagent intro | pbcopy            # macOS: paste into AI chat
stdagent intro --json              # for agent / automation integrations
```

## 命令

| 命令 | 用途 |
|---|---|
| `stdagent init` | 產生 `.stdai/` + `config.toml` + `.stdaiignore` + 範例 standards |
| `stdagent pull` | 更新快取在 `.stdai/cache/` 裡的 git 來源 |
| `stdagent sync` | 核心：pull -> parse -> convert -> 擴散 |
| `stdagent fix` | 重新 sync 修復 drift（`sync` 的別名） |
| `stdagent status` | 各 target 的 drift 狀態與上次同步時間 |
| `stdagent clean` | 刪除產生的檔案（保留 `.stdai/`） |
| `stdagent budget` | LLM context 預算檢查（字元數 + token 估算） |
| `stdagent which <path>` | 列出適用於某檔案的 rules / references（供 AI 按需載入上下文） |
| `stdagent explain` | 印出 std-agent 5 種類型（rules/skills/commands/references/subagents）的語意說明，供 AI 參考 |
| `stdagent intro` | 印出遷移提示詞，供 LLM 轉換既有設定 |
| `stdagent upgrade` | 從 GitHub Releases 自我升級（sha256 + 原子取代） |
| `stdagent version` | 建置資訊 |

每個命令都支援 `--help`。完整參考：[docs/commands.md](docs/commands.md)。

## 基於 Protocol 的架構

v0.0.4 引入了三層 transformer 架構：每個 target 的 `Plan()` 都委派給 6 個 protocol（AgentsMD / ClaudeMD / Cursor / Clinerules / WindsurfStyle / Copilot）之一，用一個 `protocol.Adapter` struct literal 參數化。新增一個工具現在只需要約 25-35 行程式碼，而不是 145 行（程式碼去重 60-70%）。

優雅降級：當某個 target 原生不支援 std-agent 的某種類型時（例如 references 在所有工具、subagents 在 amp / windsurf 上），stdagent 會退回到子目錄隔離的路徑（`<FallbackDir>/references/<name>.md`），帶上 frontmatter `std-agent-type: <type>` 加 HTML 註解說明，不使用 std-agent 私有前綴。

## 來源檔案格式

完整 schema 見 [docs/spec.md](docs/spec.md) Part 1。最簡形式：

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

MCP 伺服器（`.stdai/standards/mcp.json`）：

```json
{
  "version": "1.0",
  "servers": {
    "github": { "type": "stdio", "command": "gh", "args": ["api"] },
    "linear": { "type": "http", "url": "https://mcp.linear.app/sse" }
  }
}
```

## 設定

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

完整參考：[docs/config-spec.md](docs/config-spec.md)。

## 專案佈局

```
your-project/
├── .stdai/                    內部管理區（唯一事實來源）
│   ├── config.toml            唯一的設定檔
│   ├── standards/             撰寫根目錄
│   │   ├── rules/
│   │   ├── skills/
│   │   ├── commands/
│   │   ├── references/
│   │   └── mcp.json           MCP 伺服器（選用）
│   ├── cache/                 git 來源快取
│   ├── backups/               每次 sync 前自動快照
│   └── state.json             執行時狀態
├── .stdaiignore               gitignore 風格的 glob，用於排除來源檔案
├── CLAUDE.md                  擴散目標：Claude Code
├── AGENTS.md                  擴散目標：Codex / Cursor fallback / Copilot agent / OpenCode / Antigravity
├── GEMINI.md                  擴散目標：Gemini CLI
├── .mcp.json                  Claude 的 MCP 設定
└── .claude/ .codex/ .cursor/ .github/ .windsurf/ .gemini/ .clinerules/ .opencode/ .continue/ .agents/
```

詳見 [docs/file-structure.md](docs/file-structure.md)。

## Monorepo 支援

不明確指定 `--config` 時，`stdagent` 會從 `cwd` 向上尋找最近的 `.stdai/config.toml`。從任一子目錄執行都能自動定位到 monorepo 根目錄。

## 開發

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

## 文件

- **[docs/spec.md](docs/spec.md)**：完整規格，std-agent 標準 + 23 個工具的差異 + 轉換策略
- [docs/prd.md](docs/prd.md)：產品需求
- [docs/architecture.md](docs/architecture.md)：模組佈局與資料流
- [docs/commands.md](docs/commands.md)：CLI 命令參考
- [docs/conversion-rules.md](docs/conversion-rules.md)：轉換矩陣 + frontmatter 欄位對應
- [docs/format-spec.md](docs/format-spec.md)：frontmatter schema 細節
- [docs/file-structure.md](docs/file-structure.md)：目錄結構慣例
- [docs/roadmap.md](docs/roadmap.md)：路線圖
- [docs/targets/](docs/targets/)：各工具的調研筆記

## License

MIT，詳見 [LICENSE](LICENSE)。
