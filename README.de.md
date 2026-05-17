# std-agent

![std-agent: eine Quelle der Wahrheit für 11 AI-CLI-Tools](docs/assets/hero.png)

[![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![CI](https://github.com/StringKe/std-ai/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/StringKe/std-ai/actions/workflows/ci.yml)

[English](README.md) | [简体中文](README.zh-CN.md) | [繁體中文](README.zh-TW.md) | [日本語](README.ja.md) | [한국어](README.ko.md) | [Русский](README.ru.md) | [Français](README.fr.md) | **Deutsch** | [Español](README.es.md) | [Português](README.pt-BR.md)

---

`stdagent` ist ein leichtgewichtiges, reines Go-CLI-Tool. Es hält ein einziges `.stdai/`-Verzeichnis als Single Source of Truth für die AI-Konfiguration deines Projekts und verteilt sie dann an **11 AI-CLI-Tools**. Die nativen Dateiformate, frontmatter-Dialekte und Eigenheiten jedes Tools werden für dich erledigt.

Hör auf, `CLAUDE.md`, `AGENTS.md`, `GEMINI.md`, `.cursor/rules/`, `.windsurf/rules/`, `.clinerules/`, `.github/copilot-instructions.md` von Hand zu pflegen. Einmal schreiben, überall synchronisieren.

## Warum std-agent?

- **Einzige Quelle** — schreibe `rules` / `skills` / `commands` / `references` einmal in YAML frontmatter + Markdown.
- **Elf Ziele** — Claude Code, Codex, Cursor, GitHub Copilot, Windsurf, Gemini CLI, Aider, Cline, OpenCode, Continue.dev, Antigravity.
- **Kein Lock-in** — der Writer fasst nur eine kleine Whitelist von Pfaden an; Backup vor jedem Sync; `clean` macht alles rückgängig.
- **Drift-Erkennung** — `status` zeigt Dateien, die außerhalb von stdagent geändert wurden; `fix` wendet die Quelle neu an.
- **MCP** — eine einzelne `.stdai/standards/mcp.json` wird zu `.mcp.json` / `.cursor/mcp.json` / `.vscode/mcp.json` verteilt.
- **Monorepo-fähig** — die Config-Suche läuft von `cwd` aufwärts; funktioniert aus jedem Unterverzeichnis.
- **Selbst-Upgrade** — `stdagent upgrade` zieht signierte Releases von GitHub mit sha256-Prüfung und atomarem Austausch.

## Unterstützte Tools

### Tier 1 (9)

| Ziel | Hauptausgaben |
|---|---|
| Claude Code (Anthropic) | `CLAUDE.md` + `.claude/{rules,skills,commands}/` + `.mcp.json` |
| Codex (OpenAI) | `AGENTS.md` + `.agents/skills/` + `.codex/rules/` (Byte-Budget-Spillover) |
| Cursor | `.cursor/{rules/*.mdc,skills,commands}/` + `.cursor/mcp.json` |
| GitHub Copilot | `.github/{copilot-instructions,instructions,prompts,agents}/` + `.vscode/mcp.json` |
| Windsurf (Codeium) | `.windsurf/{rules,skills,workflows}/` |
| Gemini CLI (Google) | `GEMINI.md` + `.gemini/commands/*.toml` |
| Aider | nutzt `AGENTS.md` wieder (noop) |
| Cline | `.clinerules/` + `.clinerules/workflows/` |
| OpenCode | `.opencode/{agents,commands}/` |

### Tier 2 (2)

| Ziel | Hauptausgaben |
|---|---|
| Continue.dev | `.continue/{rules,prompts}/` |
| Antigravity (Google) | `.agents/{rules,workflows}/` |

Jede Integration ist in [docs/targets/](docs/targets/) dokumentiert.

## Schnellstart

```bash
# Installation (macOS / Linux)
curl -fsSL https://raw.githubusercontent.com/StringKe/std-ai/main/install.sh | sh

# Installation (Windows PowerShell)
irm https://raw.githubusercontent.com/StringKe/std-ai/main/install.ps1 | iex

# In deinem Projekt initialisieren
cd your-project
stdagent init

# .stdai/standards/rules/example.md bearbeiten, dann an alle aktivierten Ziele syncen
stdagent sync

# Drift prüfen / beheben
stdagent status
stdagent fix
```

## Bestehendes Projekt zu std-ai migrieren

Dein Projekt ist bereits voll mit `CLAUDE.md` / `AGENTS.md` / `.cursor/rules/` / `.clinerules/` / `.github/copilot-instructions.md`? Füge den folgenden Prompt in Claude Code / Codex / Cursor / Gemini CLI ein und alles wird in die `.stdai/standards/`-Struktur reorganisiert.

````text
Hilf mir, dieses Projekt von verstreuter AI-Konfiguration zu std-agent zu migrieren. Bitte mache:

1. Scanne mit Glob / Read jede vorhandene AI-Regeldatei:
   - Wurzel: CLAUDE.md AGENTS.md GEMINI.md .cursorrules .windsurfrules .clinerules
   - Unterverzeichnisse: .claude/rules/ .claude/skills/ .claude/commands/ .claude/agents/
              .cursor/rules/ .windsurf/rules/ .clinerules/ .continue/rules/
              .github/copilot-instructions.md .github/instructions/
              .rulesync/rules/ .codex/AGENTS.md
   - Verschachtelte CLAUDE.md im Repo (find . -name CLAUDE.md -not -path './.stdai/*')

2. Berichte ein Inventar: X rules / Y skills / Z commands / N verschachtelte CLAUDE.md,
   und markiere, welche Dateien "Projektüberblick"-Inhalt enthalten.

3. Schlage einen Aufteilungsplan vor, warte auf meine Zustimmung, dann schreibe Dateien:
   - Projektüberblick (Definition / Stack / eiserne Regeln / Wartungsablauf)
     -> .stdai/standards/root.md
   - Jede fokussierte Regel -> .stdai/standards/rules/<kebab-name>.md
   - Skill-Paket -> .stdai/standards/skills/<name>/SKILL.md (mit Unterordnern scripts/ references/)
   - Slash-Commands -> .stdai/standards/commands/<name>.md
   - Verschachtelte CLAUDE.md -> .stdai/standards/nested/<relativer-pfad>/root.md
   - Jede Datei erhält frontmatter: type / name / description / priority / applyTo

4. Kein "Refactoring" des Originals. Behalte jeden ausführbaren Befehl, API-Endpunkt,
   Fehlerstring, Dateipfad, jede Zeilennummer. Erlaubte "Optimierungen": Füllwörter
   entfernen, Duplikate zusammenführen, übergroße Dateien aufteilen, veraltete Toolnamen
   umbenennen.

5. Wenn fertig, sag mir, `stdagent sync` auszuführen und Alt-Artefakte (.rulesync/,
   .cursorrules Einzeldatei usw.) zu löschen. Lösche NICHT die von stdagent erzeugten
   Dateien (CLAUDE.md / AGENTS.md / .claude/rules/).

Vollständige Spezifikation (Tabelle der frontmatter-Felder, root.md-Vorlage, verschachteltes
Layout, rulesync-Migrations-Mapping) in der Ausgabe von `stdagent intro`.
````

Auch direkt in einen LLM-CLI pipen:

```bash
stdagent intro | pbcopy            # macOS: in Zwischenablage kopieren und in AI-Chat einfügen
stdagent intro --json              # für Agenten- / Automatisierungs-Integrationen
```

## Befehle

| Befehl | Zweck |
|---|---|
| `stdagent init` | Erstellt `.stdai/` + `config.toml` + `.stdaiignore` + Beispiel-Standards |
| `stdagent pull` | Aktualisiert git-Quellen im Cache `.stdai/cache/` |
| `stdagent sync` | Kern: pull → parse → convert → Verteilung |
| `stdagent fix` | Re-Sync zur Drift-Reparatur (Alias für `sync`) |
| `stdagent status` | Drift und letzter Sync je Ziel |
| `stdagent clean` | Entfernt generierte Dateien (behält `.stdai/`) |
| `stdagent budget` | LLM-Context-Budget-Check (Zeichen + Token-Schätzung) |
| `stdagent which <path>` | Listet auf eine Datei anwendbare rules / references (On-Demand-Context für die KI) |
| `stdagent intro` | Druckt einen Migrations-Prompt für ein LLM zur Konvertierung bestehender Configs |
| `stdagent upgrade` | Selbst-Upgrade von GitHub Releases (sha256 + atomarer Austausch) |
| `stdagent version` | Build-Infos |

Jeder Befehl unterstützt `--help`. Vollständige Referenz: [docs/commands.md](docs/commands.md).

## Quelldateiformat

Vollständiges Schema in [docs/spec.md](docs/spec.md) Part 1. Minimalform:

```markdown
---
type: rules                       # rules | skills | commands | references
name: coding-style
description: Allgemeiner Code-Stil
priority: high                    # high | normal | low
targets: [claude-code, codex]     # explizit aktivieren (oder exclude_targets zum Ausschluss)
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
inject = true            # "Generated by stdagent" footer in Ausgaben einfuegen
inject_whatis = true     # einzeilige Herkunfts-Notiz in skills einfuegen
auto_pull = true         # git-Quellen bei jedem Sync pullen
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

Vollständige Referenz: [docs/config-spec.md](docs/config-spec.md).

## Projektlayout

```
your-project/
├── .stdai/                    Interner Verwaltungsbereich (Source of Truth)
│   ├── config.toml            Einzige Config-Datei
│   ├── standards/             Autoring-Wurzel
│   │   ├── rules/
│   │   ├── skills/
│   │   ├── commands/
│   │   ├── references/
│   │   └── mcp.json           MCP-Server (optional)
│   ├── cache/                 Cache der git-Quellen
│   ├── backups/               Auto-Snapshot vor jedem Sync
│   └── state.json             Runtime-State
├── .stdaiignore               gitignore-style globs zum Ausschluss
├── CLAUDE.md                  Verteilung: Claude Code
├── AGENTS.md                  Verteilung: Codex / Cursor fallback / Copilot agent / OpenCode / Antigravity
├── GEMINI.md                  Verteilung: Gemini CLI
├── .mcp.json                  MCP für Claude
└── .claude/ .codex/ .cursor/ .github/ .windsurf/ .gemini/ .clinerules/ .opencode/ .continue/ .agents/
```

Details: [docs/file-structure.md](docs/file-structure.md).

## Monorepo-Unterstützung

Ohne `--config` läuft `stdagent` von `cwd` aufwärts zum nächstgelegenen `.stdai/config.toml`. Starte ihn aus jedem Unterverzeichnis, die Monorepo-Wurzel wird automatisch gefunden.

## Entwicklung

```bash
# Toolchain (mise + go + golangci-lint + gofumpt + git-cliff)
mise install

# Häufige Aufgaben
mise run fmt        # gofumpt + goimports
mise run lint       # golangci-lint
mise run test       # go test -race -cover
mise run check      # fmt + lint + test in einem Schritt
mise run build      # baut bin/stdagent
mise run run        # go run ./cmd/stdagent
```

## Dokumentation

- **[docs/spec.md](docs/spec.md)** — vollständige Spezifikation: std-ai-Standard + 11-Tool-Divergenz + Konvertierungsstrategie
- [docs/prd.md](docs/prd.md) — Produktanforderungen
- [docs/architecture.md](docs/architecture.md) — Modulaufteilung und Datenfluss
- [docs/commands.md](docs/commands.md) — CLI-Befehlsreferenz
- [docs/conversion-rules.md](docs/conversion-rules.md) — Konvertierungsmatrix + frontmatter-Mapping
- [docs/format-spec.md](docs/format-spec.md) — detailliertes frontmatter-Schema
- [docs/file-structure.md](docs/file-structure.md) — Verzeichniskonventionen
- [docs/roadmap.md](docs/roadmap.md) — Roadmap
- [docs/targets/](docs/targets/) — Recherche-Notizen je Tool (11)

## License

MIT — siehe [LICENSE](LICENSE).
