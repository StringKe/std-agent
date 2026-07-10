# std-agent

![std-agent: una fuente única de verdad para 11 herramientas CLI de IA](docs/assets/hero.png)

[![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![CI](https://github.com/StringKe/std-agent/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/StringKe/std-agent/actions/workflows/ci.yml)

[English](README.md) | [简体中文](README.zh-CN.md) | [繁體中文](README.zh-TW.md) | [日本語](README.ja.md) | [한국어](README.ko.md) | [Русский](README.ru.md) | [Français](README.fr.md) | [Deutsch](README.de.md) | **Español** | [Português](README.pt-BR.md)

---

`stdagent` es una herramienta CLI ligera, escrita en Go puro. Mantiene un único directorio `.stdai/` como fuente de verdad para la configuración de IA de tu proyecto y luego la distribuye a **11 herramientas CLI de IA**, encargándose de los formatos nativos, los dialectos de frontmatter y todas las peculiaridades por ti.

Deja de mantener a mano `CLAUDE.md`, `AGENTS.md`, `GEMINI.md`, `.cursor/rules/`, `.windsurf/rules/`, `.clinerules/`, `.github/copilot-instructions.md`. Escribe una vez, sincroniza en todas partes.

## ¿Por qué std-agent?

- **Fuente única** — escribe `rules` / `skills` / `commands` / `references` una sola vez en YAML frontmatter + Markdown.
- **Once destinos** — Claude Code, Codex, Cursor, GitHub Copilot, Windsurf, Gemini CLI, Aider, Cline, OpenCode, Continue.dev, Antigravity.
- **Sin lock-in** — el writer solo toca una lista blanca de rutas; backup antes de cada sync; `clean` revierte todo.
- **Detección de drift** — `status` muestra archivos modificados fuera de stdagent; `fix` los reaplicará.
- **MCP** — un único `.stdai/standards/mcp.json` se distribuye a `.mcp.json` / `.cursor/mcp.json` / `.vscode/mcp.json`.
- **Apto para monorepo** — la búsqueda de config sube desde `cwd`; funciona desde cualquier subdirectorio.
- **Auto-actualización** — `stdagent upgrade` descarga releases firmados de GitHub con verificación sha256 y reemplazo atómico.

## Herramientas soportadas

### Tier 1 (9)

| Destino | Salidas principales |
|---|---|
| Claude Code (Anthropic) | `CLAUDE.md` + `.claude/{rules,skills,commands}/` + `.mcp.json` |
| Codex (OpenAI) | `AGENTS.md` + `.agents/skills/` + `.codex/rules/` (spillover por bytes) |
| Cursor | `.cursor/{rules/*.mdc,skills,commands}/` + `.cursor/mcp.json` |
| GitHub Copilot | `.github/{copilot-instructions,instructions,prompts,agents}/` + `.vscode/mcp.json` |
| Windsurf (Codeium) | `.windsurf/{rules,skills,workflows}/` |
| Gemini CLI (Google) | `GEMINI.md` + `.gemini/commands/*.toml` |
| Aider | reutiliza `AGENTS.md` (noop) |
| Cline | `.clinerules/` + `.clinerules/workflows/` |
| OpenCode | `.opencode/{agents,commands}/` |

### Tier 2 (2)

| Destino | Salidas principales |
|---|---|
| Continue.dev | `.continue/{rules,prompts}/` |
| Antigravity (Google) | `.agents/{rules,workflows}/` |

Cada integración está documentada en [docs/targets/](docs/targets/).

## Inicio rápido

```bash
# Instalación (macOS / Linux)
curl -fsSL https://raw.githubusercontent.com/StringKe/std-agent/main/install.sh | sh

# Instalación (Windows PowerShell)
irm https://raw.githubusercontent.com/StringKe/std-agent/main/install.ps1 | iex

# Inicializa en tu proyecto
cd your-project
stdagent init

# Edita .stdai/standards/rules/example.md y sincroniza a todos los targets habilitados
stdagent sync

# Inspecciona / corrige drift
stdagent status
stdagent fix
```

## Migrar un proyecto existente a std-agent

¿Tu proyecto ya está repleto de `CLAUDE.md` / `AGENTS.md` / `.cursor/rules/` / `.clinerules/` / `.github/copilot-instructions.md`? Pega el prompt de abajo en Claude Code / Codex / Cursor / Gemini CLI y reorganizará todo en la estructura `.stdai/standards/`.

````text
Ayúdame a migrar este proyecto desde una configuración de IA dispersa a std-agent. Por favor:

1. Con Glob / Read escanea todos los archivos de reglas IA existentes:
   - Raíz: CLAUDE.md AGENTS.md GEMINI.md .cursorrules .windsurfrules .clinerules
   - Subdirectorios: .claude/rules/ .claude/skills/ .claude/commands/ .claude/agents/
              .cursor/rules/ .windsurf/rules/ .clinerules/ .continue/rules/
              .github/copilot-instructions.md .github/instructions/
              .rulesync/rules/ .codex/AGENTS.md
   - CLAUDE.md anidados dentro del repo (find . -name CLAUDE.md -not -path './.stdai/*')

2. Reporta un inventario: X rules / Y skills / Z commands / N CLAUDE.md anidados,
   e indica qué archivos contienen "resumen del proyecto".

3. Propón un plan de división, espera mi aprobación y luego escribe archivos:
   - Resumen del proyecto (definición / stack / reglas de hierro / flujo de mantenimiento)
     -> .stdai/standards/root.md
   - Cada regla enfocada -> .stdai/standards/rules/<kebab-name>.md
   - Paquete skill -> .stdai/standards/skills/<name>/SKILL.md (con subdirectorios scripts/ references/)
   - Plantillas de slash command -> .stdai/standards/commands/<name>.md
   - CLAUDE.md anidado -> .stdai/standards/nested/<ruta-relativa>/root.md
   - Cada archivo recibe frontmatter: type / name / description / priority / applyTo

4. Prohibido "refactorizar" el contenido original. Conserva todos los comandos ejecutables,
   endpoints de API, strings de error, rutas de archivos, números de línea. "Optimizaciones"
   permitidas: eliminar palabras de relleno, fusionar duplicados, dividir archivos enormes,
   renombrar herramientas obsoletas.

5. Al terminar, indícame ejecutar `stdagent sync` y eliminar artefactos antiguos
   (.rulesync/, .cursorrules monoarchivo, etc.). NO borres los archivos que stdagent
   produce (CLAUDE.md / AGENTS.md / .claude/rules/).

La spec completa (tabla de campos frontmatter, plantilla root.md, layout anidado, mapping
de migración rulesync) está en la salida del comando `stdagent intro`.
````

También se puede tubear directamente a un LLM CLI:

```bash
stdagent intro | pbcopy            # macOS: copia al portapapeles y pega en el chat IA
stdagent intro --json              # para integraciones agente / automatización
```

## Comandos

| Comando | Propósito |
|---|---|
| `stdagent init` | Crea `.stdai/` + `config.toml` + `.stdaiignore` + standards de ejemplo |
| `stdagent pull` | Actualiza las fuentes git cacheadas en `.stdai/cache/` |
| `stdagent sync` | Núcleo: pull → parse → convert → distribución |
| `stdagent fix` | Re-sync para reparar drift (alias de `sync`) |
| `stdagent status` | Drift y último sync por destino |
| `stdagent clean` | Elimina los archivos generados (preserva `.stdai/`) |
| `stdagent budget` | Chequeo de presupuesto de contexto LLM (chars + estimación de tokens) |
| `stdagent which <path>` | Lista las rules / references aplicables a un archivo (contexto bajo demanda para IA) |
| `stdagent intro` | Imprime un prompt de migración para que un LLM convierta tu config existente |
| `stdagent upgrade` | Auto-actualización desde GitHub Releases (sha256 + reemplazo atómico) |
| `stdagent version` | Info de build |

Cada comando soporta `--help`. Referencia completa: [docs/commands.md](docs/commands.md).

## Formato de los archivos fuente

Schema completo en [docs/spec.md](docs/spec.md) Part 1. Forma mínima:

```markdown
---
type: rules                       # rules | skills | commands | references
name: coding-style
description: Estilo de código general
priority: high                    # high | normal | low
targets: [claude-code, codex]     # opt-in (o exclude_targets para excluir)
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
inject = true            # inserta footer "Generated by stdagent" en las salidas
inject_whatis = true     # añade nota de una línea sobre el origen dentro de skills
auto_pull = true         # hace pull de fuentes git en cada sync
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

Referencia completa: [docs/config-spec.md](docs/config-spec.md).

## Estructura del proyecto

```
your-project/
├── .stdai/                    Área de gestión interna (fuente de verdad)
│   ├── config.toml            Único archivo de configuración
│   ├── standards/             Raíz de edición
│   │   ├── rules/
│   │   ├── skills/
│   │   ├── commands/
│   │   ├── references/
│   │   └── mcp.json           Servidores MCP (opcional)
│   ├── cache/                 Cache de fuentes git
│   ├── backups/               Snapshot automático antes de cada sync
│   └── state.json             Estado de runtime
├── .stdaiignore               globs estilo gitignore para excluir fuentes
├── CLAUDE.md                  Distribución: Claude Code
├── AGENTS.md                  Distribución: Codex / Cursor fallback / Copilot agent / OpenCode / Antigravity
├── GEMINI.md                  Distribución: Gemini CLI
├── .mcp.json                  MCP para Claude
└── .claude/ .codex/ .cursor/ .github/ .windsurf/ .gemini/ .clinerules/ .opencode/ .continue/ .agents/
```

Detalles: [docs/file-structure.md](docs/file-structure.md).

## Soporte de monorepo

Si se omite `--config`, `stdagent` sube desde `cwd` para encontrar el `.stdai/config.toml` más cercano. Ejecútalo desde cualquier subdirectorio y localizará la raíz del monorepo automáticamente.

## Desarrollo

```bash
# Toolchain (mise + go + golangci-lint + gofumpt + git-cliff)
mise install

# Tareas habituales
mise run fmt        # gofumpt + goimports
mise run lint       # golangci-lint
mise run test       # go test -race -cover
mise run check      # fmt + lint + test en un comando
mise run build      # produce bin/stdagent
mise run run        # go run ./cmd/stdagent
```

## Documentación

- **[docs/spec.md](docs/spec.md)** — spec completa: estándar std-agent + divergencia 11 herramientas + estrategia de conversión
- [docs/prd.md](docs/prd.md) — requerimientos de producto
- [docs/architecture.md](docs/architecture.md) — módulos y flujo de datos
- [docs/commands.md](docs/commands.md) — referencia CLI
- [docs/conversion-rules.md](docs/conversion-rules.md) — matriz de conversión + mapping de frontmatter
- [docs/format-spec.md](docs/format-spec.md) — schema detallado del frontmatter
- [docs/file-structure.md](docs/file-structure.md) — convenciones de directorios
- [docs/roadmap.md](docs/roadmap.md) — roadmap
- [docs/targets/](docs/targets/) — notas de investigación de cada herramienta (11)

## License

MIT — ver [LICENSE](LICENSE).
