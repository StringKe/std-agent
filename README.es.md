# std-agent

![std-agent: una única fuente de verdad para 25 herramientas de CLI de IA](docs/assets/hero.png)

[![Release](https://img.shields.io/github/v/release/StringKe/std-agent?sort=semver)](https://github.com/StringKe/std-agent/releases)
[![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go)](https://go.dev/)
[![Go Report Card](https://goreportcard.com/badge/github.com/StringKe/std-agent)](https://goreportcard.com/report/github.com/StringKe/std-agent)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![CI](https://github.com/StringKe/std-agent/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/StringKe/std-agent/actions/workflows/ci.yml)

[English](README.md) | [简体中文](README.zh-CN.md) | [繁體中文](README.zh-TW.md) | [日本語](README.ja.md) | [한국어](README.ko.md) | [Русский](README.ru.md) | [Français](README.fr.md) | [Deutsch](README.de.md) | **Español** | [Português](README.pt-BR.md)

---

`stdagent` es una CLI ligera y escrita en Go puro que mantiene un único directorio `.stdai/` como fuente de verdad para la configuración de IA de tu proyecto, y luego la distribuye a **25 herramientas de CLI de IA** gestionando por ti sus formatos de archivo nativos, dialectos de frontmatter y particularidades.

Deja de mantener `CLAUDE.md`, `AGENTS.md`, `GEMINI.md`, `.cursor/rules/`, `.windsurf/rules/`, `.clinerules/`, `.github/copilot-instructions.md`, ... a mano. Edita una vez, sincroniza en todas partes.

## Por qué std-agent

- **Fuente única** -- escribe `rules` / `skills` / `commands` / `references` / `subagents` una sola vez en frontmatter YAML + Markdown.
- **Veinticinco destinos** -- Claude Code, Codex, Cursor, GitHub Copilot, Windsurf/Devin, Gemini CLI, Aider, Cline, OpenCode, Crush, Amp, Warp, Factory, Junie, Antigravity, Qwen Code, Pi, Kilo Code, Augment Code, Jules, Grok Build, Kimi Code, Kiro, Goose, Zed.
- **Precisión de especificación** -- cada ruta de salida, dialecto de frontmatter y límite de tamaño se verifica contra la documentación oficial de las herramientas (última auditoría completa: 2026-07); directorios nativos de Agent Skills donde existen.
- **Sin dependencia** -- el writer solo toca una pequeña whitelist de rutas; hace backup antes de cada sync; `clean` revierte todo.
- **Detección de drift** -- `status` muestra archivos modificados fuera de stdagent; `fix` vuelve a aplicar la fuente.
- **MCP** -- un único `.stdai/standards/mcp.json` se distribuye a `.mcp.json` / `.cursor/mcp.json` / `.vscode/mcp.json`
- **Compatible con monorepo** -- la búsqueda de configuración sube desde `cwd`; funciona desde cualquier subdirectorio.
- **Autoactualizable** -- `stdagent upgrade` obtiene releases firmados desde GitHub con verificación sha256 y reemplazo atómico.

## Herramientas admitidas

### Tier 1 (14)

| Destino | Salidas principales |
|---|---|
| Claude Code (Anthropic) | `CLAUDE.md` + `.claude/{rules,skills,commands,agents}/` + `.mcp.json` |
| Codex (OpenAI) | `AGENTS.md` + `.agents/skills/` + `.codex/agents/*.toml` (subagentes nativos) |
| Cursor | `.cursor/{rules/*.mdc,skills,commands,agents}/` + `.cursor/mcp.json` |
| GitHub Copilot | `.github/{copilot-instructions,instructions,prompts,agents,skills}/` + `.vscode/mcp.json` |
| Windsurf / Devin (Cognition) | `.windsurf/{rules,skills,workflows}/` + espejo en `.devin/rules/` |
| Gemini CLI (Google) | `GEMINI.md` + `.gemini/skills/` + `.gemini/commands/*.toml` |
| Aider | reutiliza `AGENTS.md` (noop) |
| Cline | `.clinerules/` (prefijos numéricos 100/500/900) |
| OpenCode | `.opencode/{skills,commands}/` |
| Junie (JetBrains) | `AGENTS.md` compartido + `.junie/{rules,skills}/` |
| Crush (Charmbracelet) | `CRUSH.md` + `.crush/skills/` + registro de skills en `crush.json` |
| Amp (Sourcegraph) | `AGENTS.md` (inline) + `.agents/skills/` |
| Warp | `AGENTS.md` (inline + anidado) + `.agents/skills/` |
| Factory (Factory.ai) | `.factory/{rules,skills,commands,droids}/` |

### Tier 2 (11)

| Destino | Salidas principales |
|---|---|
| Zed | `AGENTS.md` compartido + `.agents/skills/` |
| Antigravity (Google) | `.agents/{rules,skills,workflows}/` |
| Qwen Code (Alibaba) | `QWEN.md` + `.qwen/{rules,skills,commands}/` |
| Pi | `.pi/skills/` + `.pi/prompts/` |
| Kilo Code (kilo.ai) | `.kilo/{rules,skills,command}/` + registro de instructions en `kilo.jsonc` |
| Augment Code | `.augment/{rules,skills}/` |
| Jules (Google) | `AGENTS.md` |
| Grok Build (xAI) | `AGENTS.md` + `.grok/skills/` |
| Kimi Code (Moonshot AI) | `AGENTS.md` + `.agents/skills/` |
| Kiro (AWS) | `AGENTS.md` + `.kiro/{steering,skills,agents}/` |
| Goose (AAIF) | `AGENTS.md` + `.agents/skills/` |

Cada integración está documentada en [docs/targets/](docs/targets/). Este repositorio solo activa Grok Build; los modos de gitignore y un fan-out pequeño están en [examples/](examples/).

## Inicio rápido

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

## Migrar un proyecto existente a std-agent

¿Tu proyecto ya está lleno de `CLAUDE.md` / `AGENTS.md` / `.cursor/rules/` / `.clinerules/` / `.github/copilot-instructions.md`? Pega el siguiente prompt en Claude Code / Codex / Cursor / Gemini CLI y reorganizará todo en `.stdai/standards/` por ti.

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

Envíalo también directamente a una CLI de LLM:

```bash
stdagent intro | pbcopy            # macOS: paste into AI chat
stdagent intro --json              # for agent / automation integrations
```

## Comandos

| Comando | Propósito |
|---|---|
| `stdagent init` | Genera el andamiaje de `.stdai/` + `config.toml` + `.stdaiignore` + standards de ejemplo |
| `stdagent pull` | Actualiza las fuentes basadas en git cacheadas en `.stdai/cache/` |
| `stdagent sync` | Núcleo: pull -> parse -> convert -> fan out |
| `stdagent fix` | Vuelve a sincronizar para reparar el drift (alias de `sync`) |
| `stdagent status` | Drift por destino + hora del último sync |
| `stdagent clean` | Elimina los archivos generados (conserva `.stdai/`) |
| `stdagent budget` | Verifica el presupuesto de contexto del LLM (caracteres + estimación de tokens) |
| `stdagent which <path>` | Lista las rules / references aplicables a un archivo (enrutamiento de contexto a demanda para IA) |
| `stdagent explain` | Imprime la semántica de los 5 tipos de std-agent (rules/skills/commands/references/subagents) para IA |
| `stdagent intro` | Imprime un prompt de migración para que un LLM convierta tu configuración existente |
| `stdagent upgrade` | Autoactualización desde GitHub Releases (sha256 + reemplazo atómico) |
| `stdagent version` | Información del build |

Todos los comandos admiten `--help`. Referencia completa: [docs/commands.md](docs/commands.md).

## Arquitectura basada en protocolos

v0.0.4 introdujo una arquitectura de transformer de tres capas: el `Plan()` de cada destino delega en uno de 6 protocolos (AgentsMD / ClaudeMD / Cursor / Clinerules / WindsurfStyle / Copilot), parametrizado por un struct literal `protocol.Adapter`. Agregar una nueva herramienta ahora cuesta ~25-35 líneas en lugar de 145 (60-70% de deduplicación de código).

Degradación elegante: cuando un destino no admite nativamente un tipo de std-agent (por ejemplo, references en todas partes, subagents en amp / windsurf), stdagent recurre a rutas de subdirectorio aisladas (`<FallbackDir>/references/<name>.md`) con frontmatter `std-agent-type: <type>` + un comentario HTML explicativo, sin prefijos privados de std-agent.

## Formato de origen

Un esquema completo está en [docs/spec.md](docs/spec.md), Parte 1. La forma mínima:

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

Servidores MCP (`.stdai/standards/mcp.json`):

```json
{
  "version": "1.0",
  "servers": {
    "github": { "type": "stdio", "command": "gh", "args": ["api"] },
    "linear": { "type": "http", "url": "https://mcp.linear.app/sse" }
  }
}
```

## Configuración

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

Referencia completa: [docs/config-spec.md](docs/config-spec.md).

## Estructura del proyecto

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

Detalles: [docs/file-structure.md](docs/file-structure.md).

## Soporte para monorepos

Cuando se omite `--config`, `stdagent` sube desde `cwd` para encontrar el `.stdai/config.toml` más cercano. Ejecútalo desde cualquier subdirectorio y localizará automáticamente la raíz del monorepo.

## Desarrollo

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

## Documentación

- **[docs/spec.md](docs/spec.md)** -- especificación completa: estándar std-agent + divergencia de las 25 herramientas + estrategia de conversión
- [docs/prd.md](docs/prd.md) -- requisitos del producto
- [docs/architecture.md](docs/architecture.md) -- estructura de módulos + flujo de datos
- [docs/commands.md](docs/commands.md) -- referencia de comandos de la CLI
- [docs/conversion-rules.md](docs/conversion-rules.md) -- matriz de conversión + mapeo de frontmatter
- [docs/format-spec.md](docs/format-spec.md) -- detalles del esquema de frontmatter
- [docs/file-structure.md](docs/file-structure.md) -- convenciones de directorios
- [docs/roadmap.md](docs/roadmap.md) -- roadmap
- [docs/targets/](docs/targets/) -- notas de investigación por herramienta

## Licencia

MIT -- ver [LICENSE](LICENSE).
