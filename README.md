# std-agent

![std-agent: one source of truth for 25 AI CLI tools](docs/assets/hero.png)

[![Release](https://img.shields.io/github/v/release/StringKe/std-agent?sort=semver)](https://github.com/StringKe/std-agent/releases)
[![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go)](https://go.dev/)
[![Go Report Card](https://goreportcard.com/badge/github.com/StringKe/std-agent)](https://goreportcard.com/report/github.com/StringKe/std-agent)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![CI](https://github.com/StringKe/std-agent/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/StringKe/std-agent/actions/workflows/ci.yml)

**English** | [简体中文](README.zh-CN.md) | [繁體中文](README.zh-TW.md) | [日本語](README.ja.md) | [한국어](README.ko.md) | [Русский](README.ru.md) | [Français](README.fr.md) | [Deutsch](README.de.md) | [Español](README.es.md) | [Português](README.pt-BR.md)

---

`stdagent` is a lightweight, pure Go CLI that keeps a single `.stdai/` directory as the source of truth for your project's AI configuration, then fans it out to **25 AI CLI tools** with their native file formats, frontmatter dialects, and quirks handled for you.

Stop maintaining `CLAUDE.md`, `AGENTS.md`, `GEMINI.md`, `.cursor/rules/`, `.windsurf/rules/`, `.clinerules/`, `.github/copilot-instructions.md`, ... by hand. Edit once, sync everywhere.

## Why std-agent?

- **Single source** — write `rules` / `skills` / `commands` / `references` / `subagents` once in YAML frontmatter + Markdown.
- **Twenty-five targets** — Claude Code, Codex, Cursor, GitHub Copilot, Windsurf/Devin, Gemini CLI, Aider, Cline, OpenCode, Crush, Amp, Warp, Factory, Junie, Antigravity, Qwen Code, Pi, Kilo Code, Augment Code, Jules, Grok Build, Kimi Code, Kiro, Goose, Zed.
- **Spec-accurate** — every output path, frontmatter dialect, and size limit is verified against the tools' official docs (last full audit: 2026-08); native Agent Skills directories everywhere they exist.
- **Zero lock-in** — the writer only touches a tiny whitelist of paths; backups before every sync; `clean` reverses everything.
- **Drift detection** — `status` shows files modified outside stdagent; `fix` reapplies the source.
- **MCP** — single `.stdai/standards/mcp.json` fans out to `.mcp.json` / `.cursor/mcp.json` / `.vscode/mcp.json`
- **Monorepo aware** — config lookup walks up from `cwd`; works from any subdirectory.
- **Self-upgrading** — `stdagent upgrade` pulls signed releases from GitHub with sha256 verification and atomic replace.

## Supported tools

### Tier 1 (14)

| Target | Primary outputs |
|---|---|
| Claude Code (Anthropic) | `CLAUDE.md` + `.claude/{rules,skills,commands,agents}/` + `.mcp.json` |
| Codex (OpenAI) | `AGENTS.md` + `.agents/skills/` + `.codex/agents/*.toml` (native subagents) |
| Cursor | `.cursor/{rules/*.mdc,skills,commands,agents}/` + `.cursor/mcp.json` |
| GitHub Copilot | `.github/{copilot-instructions,instructions,prompts,agents,skills}/` + `.vscode/mcp.json` |
| Windsurf / Devin (Cognition) | `.windsurf/{rules,skills,workflows}/` + `.devin/rules/` mirror |
| Gemini CLI (Google) | `GEMINI.md` + `.gemini/skills/` + `.gemini/commands/*.toml` |
| Aider | reuses `AGENTS.md` (noop) |
| Cline | `.clinerules/` (100/500/900 numeric prefixes) |
| OpenCode | `.opencode/{skills,commands}/` |
| Junie (JetBrains) | shared `AGENTS.md` + `.junie/{rules,skills}/` |
| Crush (Charmbracelet) | `CRUSH.md` + `.crush/skills/` + `crush.json` skills registration |
| Amp (Sourcegraph) | `AGENTS.md` (inline) + `.agents/skills/` |
| Warp | `AGENTS.md` (inline + nested) + `.agents/skills/` |
| Factory (Factory.ai) | shared `AGENTS.md` + `.factory/{rules,skills,commands,droids}/` |

### Tier 2 (11)

| Target | Primary outputs |
|---|---|
| Zed | shared `AGENTS.md` + `.agents/skills/` |
| Antigravity (Google) | `.agents/{rules,skills,workflows}/` |
| Qwen Code (Alibaba) | `QWEN.md` + `.qwen/{rules,skills,commands}/` |
| Pi | shared `AGENTS.md` + `.pi/skills/` + `.pi/prompts/` |
| Kilo Code (kilo.ai) | `.kilo/{rules,skills,command}/` + `kilo.jsonc` instructions registration |
| Augment Code | `.augment/{rules,skills}/` |
| Jules (Google) | `AGENTS.md` |
| Grok Build (xAI) | `AGENTS.md` + `.grok/skills/` |
| Kimi Code (Moonshot AI) | `AGENTS.md` + `.agents/skills/` |
| Kiro (AWS) | `AGENTS.md` + `.kiro/{steering,skills,agents}/` |
| Goose (AAIF) | `AGENTS.md` + `.agents/skills/` |

Each integration is documented under [docs/targets/](docs/targets/). This repository enables Grok Build only; see [examples/](examples/) for gitignore modes and a small two-target fan-out.

## Quick start

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

## Migrate an existing project to std-agent

Project already littered with `CLAUDE.md` / `AGENTS.md` / `.cursor/rules/` / `.clinerules/` / `.github/copilot-instructions.md`? Paste the prompt below into Claude Code / Codex / Cursor / Gemini CLI and it will reorganize everything into `.stdai/standards/` for you.

````text
Migrate this repository to std-agent with `.stdai/standards/` as the single source of truth.

Done means:
- All existing root, nested, hidden, skill, command, agent, and target-specific AI configuration has been read and inventoried.
- Project facts, safety constraints, executable commands, protocol fields, paths, endpoints, and error strings are preserved.
- Repeated role text, process narration, redundant rules, and inferable guidance are removed.
- Content is classified by consumption semantics into root, rules, skills, commands, references, subagents, and nested roots.
- `stdagent sync --strict`, `stdagent status`, and `stdagent budget --rendered` pass.

Before writing, report source ownership, conflicts, unique information, and the proposed boundaries for any high-impact split. Keep root.md limited to the project entry point and high-value global constraints. Use focused source files for details. Do not edit generated outputs directly. Clean legacy artifacts only after proving the new source covers them and the deletion is recoverable.

Use `stdagent intro` for the complete source schema and migration contract.
````

Pipe straight into an LLM CLI as well:

```bash
stdagent intro | pbcopy            # macOS: paste into AI chat
stdagent intro --json              # for agent / automation integrations
```

## Commands

| Command | Purpose |
|---|---|
| `stdagent init` | Scaffold `.stdai/` + `config.toml` + `.stdaiignore` + sample standards |
| `stdagent pull` | Update git-backed sources cached in `.stdai/cache/` |
| `stdagent sync` | Core: pull → parse → convert → fan out |
| `stdagent fix` | Re-sync to repair drift (alias of `sync`) |
| `stdagent status` | Per-target drift + last sync time |
| `stdagent clean` | Remove generated files (preserves `.stdai/`) |
| `stdagent budget --rendered` | Source plus exact per-target root and sidecar context estimate |
| `stdagent which <path>` | List rules / references applicable to a file (on-demand context routing for AI) |
| `stdagent explain` | Print std-agent 5 type semantics (rules/skills/commands/references/subagents) for AI |
| `stdagent intro` | Print a migration prompt for an LLM to convert your existing config |
| `stdagent upgrade` | Self-upgrade from GitHub Releases (sha256 + atomic replace) |
| `stdagent version` | Build info |

Every command supports `--help`. Full reference: [docs/commands.md](docs/commands.md).

## Protocol-based architecture

v0.0.4 introduced a three-layer transformer architecture: each target's `Plan()` delegates to one of 6 protocols (AgentsMD / ClaudeMD / Cursor / Clinerules / WindsurfStyle / Copilot), parametrized by a `protocol.Adapter` struct literal. Adding a new tool now costs ~25-35 lines instead of 145 (60-70% code dedup).

Graceful degradation: when a target doesn't natively support a std-agent type (e.g. references everywhere, subagents in amp / windsurf), stdagent falls back to subdirectory-isolated paths (`<FallbackDir>/references/<name>.md`) with frontmatter `std-agent-type: <type>` + HTML comment explainer, no std-agent-private prefixes.

Targets that write `AGENTS.md` share one canonical rules document. Target-specific commands, skills, references, and subagents stay in native sidecars; incompatible same-path outputs fail before any write.

## Source format

A complete schema lives in [docs/spec.md](docs/spec.md) Part 1. The minimal shape:

```markdown
---
type: rules                       # rules | skills | commands | references | subagents
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

MCP servers (`.stdai/standards/mcp.json`):

```json
{
  "version": "1.0",
  "servers": {
    "github": { "type": "stdio", "command": "gh", "args": ["api"] },
    "linear": { "type": "http", "url": "https://mcp.linear.app/sse" }
  }
}
```

## Configuration

`.stdai/config.toml`:

```toml
version = "1.0"
inject = true            # inject "Generated by stdagent" footer in outputs
inject_whatis = true     # add a one-line origin note inside skills
inject_type_glossary = false # opt-in type glossary in rendered rules
auto_pull = true         # pull git sources on every sync
backup = true
backup_keep = 5
gitignore = "generated"  # off | generated | portable; empty means generated

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
junie        = { enabled = false, convert = true }
antigravity  = { enabled = false, convert = true }
zed          = { enabled = false, convert = true }

[sources.default]
url     = "https://github.com/your-org/ai-standards.git"
branch  = "main"
enabled = true
paths   = ["standards/"]
```

Full reference: [docs/config-spec.md](docs/config-spec.md).

## Project layout

```
your-project/
├── .stdai/                    Internal management area (single source of truth)
│   ├── config.toml            One config file
│   ├── standards/             Authoring root
│   │   ├── rules/
│   │   ├── skills/
│   │   ├── commands/
│   │   ├── references/
│   │   └── mcp.json           MCP servers (optional)
│   ├── cache/                 Git source cache
│   ├── backups/               Auto-snapshot before each sync
│   └── state.json             Runtime state
├── .stdaiignore               gitignore-style globs to exclude source files
├── CLAUDE.md                  Fan-out: Claude Code
├── AGENTS.md                  Fan-out: Codex / Cursor fallback / Copilot agent / OpenCode / Antigravity
├── GEMINI.md                  Fan-out: Gemini CLI
├── .mcp.json                  MCP for Claude
└── .claude/ .codex/ .cursor/ .github/ .windsurf/ .gemini/ .clinerules/ .opencode/ .continue/ .agents/
```

Details: [docs/file-structure.md](docs/file-structure.md).

## Monorepo support

When `--config` is omitted, `stdagent` walks up from `cwd` to find the nearest `.stdai/config.toml`. Run it from any subdirectory and it will locate the monorepo root automatically.

## Development

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

## Documentation

- **[docs/spec.md](docs/spec.md)** — full spec: std-agent standard + 23-tool divergence + conversion strategy
- [docs/prd.md](docs/prd.md) — product requirements
- [docs/architecture.md](docs/architecture.md) — module layout + data flow
- [docs/commands.md](docs/commands.md) — CLI command reference
- [docs/conversion-rules.md](docs/conversion-rules.md) — conversion matrix + frontmatter mapping
- [docs/format-spec.md](docs/format-spec.md) — frontmatter schema details
- [docs/file-structure.md](docs/file-structure.md) — directory conventions
- [docs/roadmap.md](docs/roadmap.md) — roadmap
- [docs/targets/](docs/targets/) — per-tool research notes

## License

MIT — see [LICENSE](LICENSE).
