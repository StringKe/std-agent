# std-agent

![std-agent: 22 個の AI CLI ツールのための唯一の信頼できる情報源](docs/assets/hero.png)

[![Release](https://img.shields.io/github/v/release/StringKe/std-agent?sort=semver)](https://github.com/StringKe/std-agent/releases)
[![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go)](https://go.dev/)
[![Go Report Card](https://goreportcard.com/badge/github.com/StringKe/std-agent)](https://goreportcard.com/report/github.com/StringKe/std-agent)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![CI](https://github.com/StringKe/std-agent/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/StringKe/std-agent/actions/workflows/ci.yml)

[English](README.md) | [简体中文](README.zh-CN.md) | [繁體中文](README.zh-TW.md) | **日本語** | [한국어](README.ko.md) | [Русский](README.ru.md) | [Français](README.fr.md) | [Deutsch](README.de.md) | [Español](README.es.md) | [Português](README.pt-BR.md)

---

`stdagent` は軽量な純 Go 製 CLI ツールです。プロジェクトの AI 設定を単一の `.stdai/` ディレクトリに唯一の信頼できる情報源として集約し、**22 個の AI CLI ツール**へ展開します。各ツールのネイティブなファイル形式、frontmatter 方言、固有の制約はすべて代わりに処理します。

`CLAUDE.md`、`AGENTS.md`、`GEMINI.md`、`.cursor/rules/`、`.windsurf/rules/`、`.clinerules/`、`.github/copilot-instructions.md` などを手作業で維持するのはもうやめましょう。一度書けば、どこでも反映されます。

## なぜ std-agent なのか

- **単一の情報源**：`rules` / `skills` / `commands` / `references` / `subagents` を YAML frontmatter + Markdown で一度だけ記述します。
- **22 個のターゲット**：Claude Code、Codex、Cursor、GitHub Copilot、Windsurf/Devin、Gemini CLI、Aider、Cline、OpenCode、Roo Code、Crush、Amp、Warp、Factory、Continue.dev、Antigravity、Qwen Code、Pi、Kilo Code、Augment Code、Jules、Grok Build。
- **仕様に忠実**：すべての出力パス、frontmatter 方言、サイズ上限は各ツールの公式ドキュメントと照合済みです（直近の全面調査：2026-07）。ネイティブの Agent Skills ディレクトリが存在する場合は、そちらを使用します。
- **ロックインなし**：writer はごく小さなパスのホワイトリストにしか触れません。sync ごとに自動バックアップ、`clean` で全て元に戻せます。
- **drift 検出**：`status` が外部から変更されたファイルを表示し、`fix` でソースを再適用します。
- **MCP**：単一の `.stdai/standards/mcp.json` を `.mcp.json` / `.cursor/mcp.json` / `.vscode/mcp.json` に展開します。
- **monorepo 対応**：設定の探索は `cwd` から上方向に行われるため、どのサブディレクトリからでも実行できます。
- **自己アップグレード**：`stdagent upgrade` が GitHub Releases から署名済みリリースを取得し、sha256 検証と原子的な置き換えを行います。

## 対応ツール

### Tier 1（14 個）

| ターゲット | 主な出力 |
|---|---|
| Claude Code (Anthropic) | `CLAUDE.md` + `.claude/{rules,skills,commands,agents}/` + `.mcp.json` |
| Codex (OpenAI) | `AGENTS.md` + `.agents/skills/` + `.codex/agents/*.toml`（ネイティブ subagents） |
| Cursor | `.cursor/{rules/*.mdc,skills,commands,agents}/` + `.cursor/mcp.json` |
| GitHub Copilot | `.github/{copilot-instructions,instructions,prompts,agents,skills}/` + `.vscode/mcp.json` |
| Windsurf / Devin (Cognition) | `.windsurf/{rules,skills,workflows}/` + `.devin/rules/` ミラー |
| Gemini CLI (Google) | `GEMINI.md` + `.gemini/skills/` + `.gemini/commands/*.toml` |
| Aider | `AGENTS.md` を再利用（noop） |
| Cline | `.clinerules/`（100/500/900 の数値プレフィックス） |
| OpenCode | `.opencode/{skills,commands}/` |
| Roo Code | `.roo/{rules,skills,commands}/` |
| Crush (Charmbracelet) | `CRUSH.md` + `.crush/skills/` + `crush.json` skills 登録 |
| Amp (Sourcegraph) | `AGENTS.md`（inline） + `.agents/skills/` |
| Warp | `AGENTS.md`（inline + nested） + `.agents/skills/` |
| Factory (Factory.ai) | `.factory/{rules,skills,commands,droids}/` |

### Tier 2（8 個）

| ターゲット | 主な出力 |
|---|---|
| Continue.dev | `.continue/{rules,skills,prompts}/` + ネストされた `rules.md` |
| Antigravity (Google) | `.agents/{rules,skills,workflows}/` |
| Qwen Code (Alibaba) | `QWEN.md` + `.qwen/{rules,skills,commands}/` |
| Pi | `.pi/skills/` + `.pi/prompts/` |
| Kilo Code (kilo.ai) | `.kilo/{rules,skills,command}/` + `kilo.jsonc` instructions 登録 |
| Augment Code | `.augment/{rules,skills}/` |
| Jules (Google) | `AGENTS.md` |
| Grok Build (xAI) | `AGENTS.md` + `.grok/skills/` |

各統合の詳細は [docs/targets/](docs/targets/) にあります。

## クイックスタート

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

## 既存プロジェクトを std-agent に移行する

`CLAUDE.md` / `AGENTS.md` / `.cursor/rules/` / `.clinerules/` / `.github/copilot-instructions.md` などが散在していませんか。以下のプロンプトを Claude Code / Codex / Cursor / Gemini CLI に貼り付ければ、すべてを `.stdai/standards/` に再構成してくれます。

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

LLM CLI に直接パイプすることもできます。

```bash
stdagent intro | pbcopy            # macOS: paste into AI chat
stdagent intro --json              # for agent / automation integrations
```

## コマンド

| コマンド | 用途 |
|---|---|
| `stdagent init` | `.stdai/` + `config.toml` + `.stdaiignore` + サンプル standards を生成 |
| `stdagent pull` | `.stdai/cache/` にキャッシュされた git ソースを更新 |
| `stdagent sync` | 中核処理：pull -> parse -> convert -> 展開 |
| `stdagent fix` | drift を修復するための再 sync（`sync` の別名） |
| `stdagent status` | 各 target の drift 状況と最終 sync 時刻 |
| `stdagent clean` | 生成されたファイルを削除（`.stdai/` は保持） |
| `stdagent budget` | LLM コンテキスト予算チェック（文字数 + token 推定） |
| `stdagent which <path>` | 指定ファイルに適用される rules / references を列挙（AI のオンデマンドなコンテキスト読込用） |
| `stdagent explain` | std-agent の 5 つの型（rules/skills/commands/references/subagents）の意味を AI 向けに出力 |
| `stdagent intro` | 既存設定を変換させるための移行プロンプトを出力 |
| `stdagent upgrade` | GitHub Releases から自己アップグレード（sha256 + 原子的な置き換え） |
| `stdagent version` | ビルド情報 |

すべてのコマンドは `--help` に対応しています。完全なリファレンス：[docs/commands.md](docs/commands.md)。

## Protocol ベースのアーキテクチャ

v0.0.4 で三層構造の transformer アーキテクチャが導入されました。各 target の `Plan()` は 6 個の protocol（AgentsMD / ClaudeMD / Cursor / Clinerules / WindsurfStyle / Copilot）のいずれかに処理を委譲し、`protocol.Adapter` の struct literal でパラメータ化されます。新しいツールの追加コストは、145 行ではなく約 25-35 行になりました（コード重複が 60-70% 削減）。

段階的縮退：target が std-agent のある型をネイティブにサポートしない場合（例：references はすべてのツールで、subagents は amp / windsurf で）、stdagent はサブディレクトリで隔離されたパス（`<FallbackDir>/references/<name>.md`）にフォールバックします。frontmatter に `std-agent-type: <type>` を付け、HTML コメントで説明を添えますが、std-agent 専用のプレフィックスは使用しません。

## ソースファイル形式

完全な schema は [docs/spec.md](docs/spec.md) の Part 1 を参照してください。最小構成は次の通りです。

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

MCP サーバー（`.stdai/standards/mcp.json`）：

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

完全なリファレンス：[docs/config-spec.md](docs/config-spec.md)。

## プロジェクト構成

```
your-project/
├── .stdai/                    内部管理領域（唯一の信頼できる情報源）
│   ├── config.toml            唯一の設定ファイル
│   ├── standards/             執筆用ルート
│   │   ├── rules/
│   │   ├── skills/
│   │   ├── commands/
│   │   ├── references/
│   │   └── mcp.json           MCP サーバー（任意）
│   ├── cache/                 git ソースキャッシュ
│   ├── backups/               sync ごとの自動スナップショット
│   └── state.json             ランタイム状態
├── .stdaiignore               ソースファイルを除外する gitignore 形式の glob
├── CLAUDE.md                  展開先：Claude Code
├── AGENTS.md                  展開先：Codex / Cursor fallback / Copilot agent / OpenCode / Antigravity
├── GEMINI.md                  展開先：Gemini CLI
├── .mcp.json                  Claude 用の MCP 設定
└── .claude/ .codex/ .cursor/ .github/ .windsurf/ .gemini/ .clinerules/ .opencode/ .continue/ .agents/
```

詳細：[docs/file-structure.md](docs/file-structure.md)。

## Monorepo サポート

`--config` を指定しない場合、`stdagent` は `cwd` から上方向に最も近い `.stdai/config.toml` を探索します。どのサブディレクトリから実行しても monorepo のルートを自動的に特定します。

## 開発

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

## ドキュメント

- **[docs/spec.md](docs/spec.md)**：完全な仕様、std-agent 標準 + 22 ツールの差異 + 変換戦略
- [docs/prd.md](docs/prd.md)：製品要件
- [docs/architecture.md](docs/architecture.md)：モジュール構成とデータフロー
- [docs/commands.md](docs/commands.md)：CLI コマンドリファレンス
- [docs/conversion-rules.md](docs/conversion-rules.md)：変換マトリクス + frontmatter フィールド対応
- [docs/format-spec.md](docs/format-spec.md)：frontmatter schema の詳細
- [docs/file-structure.md](docs/file-structure.md)：ディレクトリ構成の慣例
- [docs/roadmap.md](docs/roadmap.md)：ロードマップ
- [docs/targets/](docs/targets/)：各ツールの調査資料

## License

MIT。詳細は [LICENSE](LICENSE) を参照してください。
