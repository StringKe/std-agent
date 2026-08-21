# std-agent

![std-agent: eine einzige Quelle der Wahrheit für 25 AI-CLI-Tools](docs/assets/hero.png)

[![Release](https://img.shields.io/github/v/release/StringKe/std-agent?sort=semver)](https://github.com/StringKe/std-agent/releases)
[![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go)](https://go.dev/)
[![Go Report Card](https://goreportcard.com/badge/github.com/StringKe/std-agent)](https://goreportcard.com/report/github.com/StringKe/std-agent)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![CI](https://github.com/StringKe/std-agent/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/StringKe/std-agent/actions/workflows/ci.yml)

[English](README.md) | [简体中文](README.zh-CN.md) | [繁體中文](README.zh-TW.md) | [日本語](README.ja.md) | [한국어](README.ko.md) | [Русский](README.ru.md) | [Français](README.fr.md) | **Deutsch** | [Español](README.es.md) | [Português](README.pt-BR.md)

---

`stdagent` ist ein leichtgewichtiges, reines Go-CLI-Tool, das ein einziges Verzeichnis `.stdai/` als Quelle der Wahrheit für die AI-Konfiguration deines Projekts pflegt und sie dann auf **25 AI-CLI-Tools** verteilt -- mit deren nativen Dateiformaten, Frontmatter-Dialekten und Eigenheiten, die für dich automatisch berücksichtigt werden.

Hör auf, `CLAUDE.md`, `AGENTS.md`, `GEMINI.md`, `.cursor/rules/`, `.windsurf/rules/`, `.clinerules/`, `.github/copilot-instructions.md`, ... händisch zu pflegen. Einmal bearbeiten, überall synchronisieren.

## Warum std-agent?

- **Eine Quelle** -- schreibe `rules` / `skills` / `commands` / `references` / `subagents` einmal in YAML-Frontmatter + Markdown.
- **Funfundzwanzig Ziele** -- Claude Code, Codex, Cursor, GitHub Copilot, Windsurf/Devin, Gemini CLI, Aider, Cline, OpenCode, Crush, Amp, Warp, Factory, Junie, Antigravity, Qwen Code, Pi, Kilo Code, Augment Code, Jules, Grok Build, Kimi Code, Kiro, Goose, Zed.
- **Spezifikationsgenau** -- jeder Ausgabepfad, jeder Frontmatter-Dialekt und jedes Größenlimit wird gegen die offizielle Dokumentation der Tools verifiziert (letztes vollständiges Audit: 2026-07); native Agent-Skills-Verzeichnisse überall, wo sie existieren.
- **Kein Lock-in** -- der Writer greift nur auf eine winzige Whitelist von Pfaden zu; Backups vor jedem Sync; `clean` macht alles rückgängig.
- **Drift-Erkennung** -- `status` zeigt Dateien, die außerhalb von stdagent verändert wurden; `fix` wendet die Quelle erneut an.
- **MCP** -- ein einziges `.stdai/standards/mcp.json` fächert sich auf zu `.mcp.json` / `.cursor/mcp.json` / `.vscode/mcp.json`
- **Monorepo-fähig** -- die Konfigurationssuche läuft von `cwd` aus aufwärts; funktioniert aus jedem Unterverzeichnis.
- **Selbst-aktualisierend** -- `stdagent upgrade` lädt signierte Releases von GitHub mit sha256-Verifizierung und atomarem Ersetzen.

## Unterstützte Tools

### Tier 1 (14)

| Ziel | Primäre Ausgaben |
|---|---|
| Claude Code (Anthropic) | `CLAUDE.md` + `.claude/{rules,skills,commands,agents}/` + `.mcp.json` |
| Codex (OpenAI) | `AGENTS.md` + `.agents/skills/` + `.codex/agents/*.toml` (native Subagenten) |
| Cursor | `.cursor/{rules/*.mdc,skills,commands,agents}/` + `.cursor/mcp.json` |
| GitHub Copilot | `.github/{copilot-instructions,instructions,prompts,agents,skills}/` + `.vscode/mcp.json` |
| Windsurf / Devin (Cognition) | `.windsurf/{rules,skills,workflows}/` + Spiegelung in `.devin/rules/` |
| Gemini CLI (Google) | `GEMINI.md` + `.gemini/skills/` + `.gemini/commands/*.toml` |
| Aider | verwendet `AGENTS.md` wieder (noop) |
| Cline | `.clinerules/` (numerische Präfixe 100/500/900) |
| OpenCode | `.opencode/{skills,commands}/` |
| Junie (JetBrains) | gemeinsames `AGENTS.md` + `.junie/{rules,skills}/` |
| Crush (Charmbracelet) | `CRUSH.md` + `.crush/skills/` + Skills-Registrierung in `crush.json` |
| Amp (Sourcegraph) | `AGENTS.md` (inline) + `.agents/skills/` |
| Warp | `AGENTS.md` (inline + verschachtelt) + `.agents/skills/` |
| Factory (Factory.ai) | `.factory/{rules,skills,commands,droids}/` |

### Tier 2 (11)

| Ziel | Primäre Ausgaben |
|---|---|
| Zed | gemeinsames `AGENTS.md` + `.agents/skills/` |
| Antigravity (Google) | `.agents/{rules,skills,workflows}/` |
| Qwen Code (Alibaba) | `QWEN.md` + `.qwen/{rules,skills,commands}/` |
| Pi | `.pi/skills/` + `.pi/prompts/` |
| Kilo Code (kilo.ai) | `.kilo/{rules,skills,command}/` + Instructions-Registrierung in `kilo.jsonc` |
| Augment Code | `.augment/{rules,skills}/` |
| Jules (Google) | `AGENTS.md` |
| Grok Build (xAI) | `AGENTS.md` + `.grok/skills/` |
| Kimi Code (Moonshot AI) | `AGENTS.md` + `.agents/skills/` |
| Kiro (AWS) | `AGENTS.md` + `.kiro/{steering,skills,agents}/` |
| Goose (AAIF) | `AGENTS.md` + `.agents/skills/` |

Jede Integration ist unter [docs/targets/](docs/targets/) dokumentiert. Dieses Repository aktiviert nur Grok Build; gitignore-Modi und ein kleines Fan-out siehe [examples/](examples/).

## Schnellstart

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

## Ein bestehendes Projekt zu std-agent migrieren

Projekt schon vollgestopft mit `CLAUDE.md` / `AGENTS.md` / `.cursor/rules/` / `.clinerules/` / `.github/copilot-instructions.md`? Füge den folgenden Prompt in Claude Code / Codex / Cursor / Gemini CLI ein, und er wird alles für dich in `.stdai/standards/` neu organisieren.

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

Direkt in eine LLM-CLI pipen geht auch:

```bash
stdagent intro | pbcopy            # macOS: paste into AI chat
stdagent intro --json              # for agent / automation integrations
```

## Befehle

| Befehl | Zweck |
|---|---|
| `stdagent init` | Gerüst für `.stdai/` + `config.toml` + `.stdaiignore` + Beispiel-Standards anlegen |
| `stdagent pull` | Git-basierte Quellen aktualisieren, die in `.stdai/cache/` zwischengespeichert sind |
| `stdagent sync` | Kern: pull -> parse -> convert -> fan out |
| `stdagent fix` | Erneut synchronisieren, um Drift zu beheben (Alias von `sync`) |
| `stdagent status` | Drift pro Ziel + Zeitpunkt des letzten Sync |
| `stdagent clean` | Generierte Dateien entfernen (behält `.stdai/`) |
| `stdagent budget` | Prüfung des LLM-Kontextbudgets (Zeichen + Token-Schätzung) |
| `stdagent which <path>` | Zeigt Rules / References, die für eine Datei gelten (bedarfsgesteuertes Context-Routing für AI) |
| `stdagent explain` | Gibt die 5 std-agent-Typsemantiken (rules/skills/commands/references/subagents) für AI aus |
| `stdagent intro` | Gibt einen Migrations-Prompt aus, mit dem ein LLM deine bestehende Konfiguration konvertiert |
| `stdagent upgrade` | Selbst-Update von GitHub Releases (sha256 + atomarer Ersatz) |
| `stdagent version` | Build-Informationen |

Jeder Befehl unterstützt `--help`. Vollständige Referenz: [docs/commands.md](docs/commands.md).

## Protokollbasierte Architektur

v0.0.4 hat eine dreischichtige Transformer-Architektur eingeführt: `Plan()` jedes Ziels delegiert an eines von 6 Protokollen (AgentsMD / ClaudeMD / Cursor / Clinerules / WindsurfStyle / Copilot), parametrisiert durch ein `protocol.Adapter`-Struct-Literal. Ein neues Tool hinzuzufügen kostet jetzt ~25-35 Zeilen statt 145 (60-70% Code-Deduplizierung).

Graceful Degradation: Wenn ein Ziel einen std-agent-Typ nicht nativ unterstützt (z. B. References überall, Subagents in amp / windsurf), fällt stdagent auf isolierte Unterverzeichnis-Pfade zurück (`<FallbackDir>/references/<name>.md`) mit Frontmatter `std-agent-type: <type>` + HTML-Kommentar-Erklärung, ohne std-agent-private Präfixe.

## Quellformat

Ein vollständiges Schema befindet sich in [docs/spec.md](docs/spec.md) Teil 1. Die minimale Form:

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

MCP-Server (`.stdai/standards/mcp.json`):

```json
{
  "version": "1.0",
  "servers": {
    "github": { "type": "stdio", "command": "gh", "args": ["api"] },
    "linear": { "type": "http", "url": "https://mcp.linear.app/sse" }
  }
}
```

## Konfiguration

`.stdai/config.toml`:

```toml
version = "1.0"
inject = true            # inject "Generated by stdagent" footer in outputs
inject_whatis = true     # add a one-line origin note inside skills
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

Vollständige Referenz: [docs/config-spec.md](docs/config-spec.md).

## Projektstruktur

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

## Monorepo-Unterstützung

Wenn `--config` weggelassen wird, läuft `stdagent` von `cwd` aus aufwärts, um die nächstgelegene `.stdai/config.toml` zu finden. Führe es aus einem beliebigen Unterverzeichnis aus, und es findet die Monorepo-Wurzel automatisch.

## Entwicklung

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

## Dokumentation

- **[docs/spec.md](docs/spec.md)** -- vollständige Spezifikation: std-agent-Standard + Divergenz der 25 Tools + Konvertierungsstrategie
- [docs/prd.md](docs/prd.md) -- Produktanforderungen
- [docs/architecture.md](docs/architecture.md) -- Modulaufbau + Datenfluss
- [docs/commands.md](docs/commands.md) -- CLI-Befehlsreferenz
- [docs/conversion-rules.md](docs/conversion-rules.md) -- Konvertierungsmatrix + Frontmatter-Mapping
- [docs/format-spec.md](docs/format-spec.md) -- Details zum Frontmatter-Schema
- [docs/file-structure.md](docs/file-structure.md) -- Verzeichniskonventionen
- [docs/roadmap.md](docs/roadmap.md) -- Roadmap
- [docs/targets/](docs/targets/) -- Recherchenotizen pro Tool

## Lizenz

MIT -- siehe [LICENSE](LICENSE).
