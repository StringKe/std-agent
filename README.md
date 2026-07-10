# std-agent

![std-agent: one source of truth for 22 AI CLI tools](docs/assets/hero.png)

[![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![CI](https://github.com/StringKe/std-agent/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/StringKe/std-agent/actions/workflows/ci.yml)

**English** | [简体中文](README.zh-CN.md) | [繁體中文](README.zh-TW.md) | [日本語](README.ja.md) | [한국어](README.ko.md) | [Русский](README.ru.md) | [Français](README.fr.md) | [Deutsch](README.de.md) | [Español](README.es.md) | [Português](README.pt-BR.md)

---

`stdagent` is a lightweight, pure Go CLI that keeps a single `.stdai/` directory as the source of truth for your project's AI configuration, then fans it out to **22 AI CLI tools** with their native file formats, frontmatter dialects, and quirks handled for you.

Stop maintaining `CLAUDE.md`, `AGENTS.md`, `GEMINI.md`, `.cursor/rules/`, `.windsurf/rules/`, `.clinerules/`, `.github/copilot-instructions.md`, ... by hand. Edit once, sync everywhere.

## Why std-agent?

- **Single source** — write `rules` / `skills` / `commands` / `references` / `subagents` once in YAML frontmatter + Markdown.
- **Twenty-two targets** — Claude Code, Codex, Cursor, GitHub Copilot, Windsurf, Gemini CLI, Aider, Cline, OpenCode, Roo Code, Crush, Amp, Warp, Factory, Continue.dev, Antigravity, Qwen Code, Pi, Kilo Code, Augment Code, Jules, Grok CLI.
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
| Codex (OpenAI) | `AGENTS.md` + `.codex/memories/` + `.agents/skills/` |
| Cursor | `.cursor/{rules/*.mdc,skills,commands}/` + `.cursor/mcp.json` |
| GitHub Copilot | `.github/{copilot-instructions,instructions,prompts,agents}/` + `.vscode/mcp.json` |
| Windsurf (Codeium) | `.windsurf/{rules,skills,workflows}/` |
| Gemini CLI (Google) | `GEMINI.md` + `.gemini/commands/*.toml` |
| Aider | reuses `AGENTS.md` (noop) |
| Cline | `.clinerules/` (100/500/900 numeric prefixes) |
| OpenCode | `.opencode/{agents,commands}/` |
| Roo Code | `.roo/rules/` (Cline fork, 18k stars) |
| Crush (Charmbracelet) | `CRUSH.md` + `.crush/skills/` |
| Amp (Sourcegraph) | `AGENTS.md` (inline) |
| Warp | `AGENTS.md` (inline + nested) |
| Factory (Factory.ai) | `.factory/{rules,skills,droids}/` |

### Tier 2 (8)

| Target | Primary outputs |
|---|---|
| Continue.dev | `.continue/{rules,prompts}/` |
| Antigravity (Google) | `.agents/{rules,workflows}/` |
| Qwen Code (Alibaba) | `QWEN.md` + `.qwen/commands/` |
| Pi | `.pi/skills/` + `.pi/prompts/` |
| Kilo Code | `.kilo/rules/` (Cline second-fork) |
| Augment Code | `.augment/rules/` |
| Jules (Google) | `AGENTS.md` |
| Grok CLI (xAI) | `AGENTS.md` + `AGENTS.override.md` |

Each integration is documented under [docs/targets/](docs/targets/).

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
| `stdagent budget` | LLM context budget check (chars + token estimate) |
| `stdagent which <path>` | List rules / references applicable to a file (on-demand context routing for AI) |
| `stdagent explain` | Print std-agent 5 type semantics (rules/skills/commands/references/subagents) for AI |
| `stdagent intro` | Print a migration prompt for an LLM to convert your existing config |
| `stdagent upgrade` | Self-upgrade from GitHub Releases (sha256 + atomic replace) |
| `stdagent version` | Build info |

Every command supports `--help`. Full reference: [docs/commands.md](docs/commands.md).

## Protocol-based architecture

v0.0.4 introduced a three-layer transformer architecture: each target's `Plan()` delegates to one of 6 protocols (AgentsMD / ClaudeMD / Cursor / Clinerules / WindsurfStyle / Copilot), parametrized by a `protocol.Adapter` struct literal. Adding a new tool now costs ~25-35 lines instead of 145 (60-70% code dedup).

Graceful degradation: when a target doesn't natively support a std-agent type (e.g. skills in codex / references everywhere), stdagent falls back to subdirectory-isolated paths (`<RulesDir>/skills/<name>/SKILL.md`) with frontmatter `std-agent-type: <type>` + HTML comment explainer, no std-agent-private prefixes.

## Source format

A complete schema lives in [docs/spec.md](docs/spec.md) Part 1. The minimal shape:

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

- **[docs/spec.md](docs/spec.md)** — full spec: std-agent standard + 22-tool divergence + conversion strategy
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
