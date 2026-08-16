# std-agent

![std-agent: единый источник истины для 23 AI CLI инструментов](docs/assets/hero.png)

[![Release](https://img.shields.io/github/v/release/StringKe/std-agent?sort=semver)](https://github.com/StringKe/std-agent/releases)
[![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go)](https://go.dev/)
[![Go Report Card](https://goreportcard.com/badge/github.com/StringKe/std-agent)](https://goreportcard.com/report/github.com/StringKe/std-agent)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![CI](https://github.com/StringKe/std-agent/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/StringKe/std-agent/actions/workflows/ci.yml)

[English](README.md) | [简体中文](README.zh-CN.md) | [繁體中文](README.zh-TW.md) | [日本語](README.ja.md) | [한국어](README.ko.md) | **Русский** | [Français](README.fr.md) | [Deutsch](README.de.md) | [Español](README.es.md) | [Português](README.pt-BR.md)

---

`stdagent` -- это лёгкий CLI-инструмент на чистом Go. Он хранит одну директорию `.stdai/` как единый источник истины для AI-конфигурации проекта, а затем раскладывает её по **23 AI CLI инструментам**, беря на себя все нативные форматы файлов, диалекты frontmatter и особенности каждого инструмента.

Перестаньте вручную поддерживать `CLAUDE.md`, `AGENTS.md`, `GEMINI.md`, `.cursor/rules/`, `.windsurf/rules/`, `.clinerules/`, `.github/copilot-instructions.md` и так далее. Пишите один раз -- синхронизируется везде.

## Почему std-agent?

- **Единый источник** -- пишите `rules` / `skills` / `commands` / `references` / `subagents` один раз в YAML frontmatter + Markdown.
- **Двадцать пять целей** -- Claude Code, Codex, Cursor, GitHub Copilot, Windsurf/Devin, Gemini CLI, Aider, Cline, OpenCode, Roo Code, Crush, Amp, Warp, Factory, Continue.dev, Antigravity, Qwen Code, Pi, Kilo Code, Augment Code, Jules, Grok Build, Kimi Code, Kiro, Goose.
- **Точное соответствие спецификациям** -- каждый выходной путь, диалект frontmatter и лимит размера проверяются по официальной документации инструментов (последний полный аудит: 2026-07); нативные директории Agent Skills используются везде, где они существуют.
- **Без вендор-лока** -- writer трогает только небольшой белый список путей; перед каждой синхронизацией делается бэкап; `clean` откатывает всё.
- **Обнаружение дрифта** -- `status` показывает файлы, изменённые вне stdagent; `fix` восстанавливает их из источника.
- **MCP** -- один `.stdai/standards/mcp.json` раскладывается в `.mcp.json` / `.cursor/mcp.json` / `.vscode/mcp.json`.
- **Поддержка monorepo** -- поиск конфига идёт вверх от `cwd`; работает из любой поддиректории.
- **Самообновление** -- `stdagent upgrade` тянет подписанные релизы с GitHub с проверкой sha256 и атомарной заменой.

## Поддерживаемые инструменты

### Tier 1 (14)

| Цель | Основные выходы |
|---|---|
| Claude Code (Anthropic) | `CLAUDE.md` + `.claude/{rules,skills,commands,agents}/` + `.mcp.json` |
| Codex (OpenAI) | `AGENTS.md` + `.agents/skills/` + `.codex/agents/*.toml` (нативные subagents) |
| Cursor | `.cursor/{rules/*.mdc,skills,commands,agents}/` + `.cursor/mcp.json` |
| GitHub Copilot | `.github/{copilot-instructions,instructions,prompts,agents,skills}/` + `.vscode/mcp.json` |
| Windsurf / Devin (Cognition) | `.windsurf/{rules,skills,workflows}/` + зеркало `.devin/rules/` |
| Gemini CLI (Google) | `GEMINI.md` + `.gemini/skills/` + `.gemini/commands/*.toml` |
| Aider | переиспользует `AGENTS.md` (noop) |
| Cline | `.clinerules/` (числовые префиксы 100/500/900) |
| OpenCode | `.opencode/{skills,commands}/` |
| Roo Code | `.roo/{rules,skills,commands}/` |
| Crush (Charmbracelet) | `CRUSH.md` + `.crush/skills/` + регистрация skills в `crush.json` |
| Amp (Sourcegraph) | `AGENTS.md` (инлайн) + `.agents/skills/` |
| Warp | `AGENTS.md` (инлайн + вложенный) + `.agents/skills/` |
| Factory (Factory.ai) | `.factory/{rules,skills,commands,droids}/` |

### Tier 2 (11)

| Цель | Основные выходы |
|---|---|
| Continue.dev | `.continue/{rules,skills,prompts}/` + вложенный `rules.md` |
| Antigravity (Google) | `.agents/{rules,skills,workflows}/` |
| Qwen Code (Alibaba) | `QWEN.md` + `.qwen/{rules,skills,commands}/` |
| Pi | `.pi/skills/` + `.pi/prompts/` |
| Kilo Code (kilo.ai) | `.kilo/{rules,skills,command}/` + регистрация instructions в `kilo.jsonc` |
| Augment Code | `.augment/{rules,skills}/` |
| Jules (Google) | `AGENTS.md` |
| Grok Build (xAI) | `AGENTS.md` + `.grok/skills/` |
| Kimi Code (Moonshot AI) | `AGENTS.md` + `.agents/skills/` |
| Kiro (AWS) | `AGENTS.md` + `.kiro/{steering,skills,agents}/` |
| Goose (AAIF) | `AGENTS.md` + `.agents/skills/` |

Каждая интеграция описана в [docs/targets/](docs/targets/).

## Быстрый старт

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

## Миграция существующего проекта на std-agent

В проекте уже разбросаны `CLAUDE.md` / `AGENTS.md` / `.cursor/rules/` / `.clinerules/` / `.github/copilot-instructions.md`? Вставьте промпт ниже в Claude Code / Codex / Cursor / Gemini CLI, и он переорганизует всё в структуру `.stdai/standards/`.

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

Также можно направить вывод сразу в LLM CLI:

```bash
stdagent intro | pbcopy            # macOS: paste into AI chat
stdagent intro --json              # for agent / automation integrations
```

## Команды

| Команда | Назначение |
|---|---|
| `stdagent init` | Создать `.stdai/` + `config.toml` + `.stdaiignore` + примеры standards |
| `stdagent pull` | Обновить git-источники, кэшируемые в `.stdai/cache/` |
| `stdagent sync` | Основное: pull -> parse -> convert -> раскладка |
| `stdagent fix` | Пересинхронизация для исправления дрифта (алиас `sync`) |
| `stdagent status` | Дрифт и время последней синхронизации по каждой цели |
| `stdagent clean` | Удалить сгенерированные файлы (сохранив `.stdai/`) |
| `stdagent budget` | Проверка бюджета LLM-контекста (символы + оценка токенов) |
| `stdagent which <path>` | Список rules / references, применимых к файлу (контекст по требованию для AI) |
| `stdagent explain` | Выводит для AI смысл 5 типов std-agent (rules/skills/commands/references/subagents) |
| `stdagent intro` | Печатает промпт для LLM-миграции существующей конфигурации |
| `stdagent upgrade` | Самообновление с GitHub Releases (sha256 + атомарная замена) |
| `stdagent version` | Информация о сборке |

Каждая команда поддерживает `--help`. Полная справка: [docs/commands.md](docs/commands.md).

## Архитектура на основе протоколов

В v0.0.4 представлена трёхслойная архитектура transformer: `Plan()` каждой цели делегирует одному из 6 протоколов (AgentsMD / ClaudeMD / Cursor / Clinerules / WindsurfStyle / Copilot), параметризованному литералом структуры `protocol.Adapter`. Добавление нового инструмента теперь стоит ~25-35 строк вместо 145 (дедупликация кода на 60-70%).

Плавная деградация (graceful degradation): если цель не поддерживает тип std-agent нативно (например, references повсеместно, subagents в amp / windsurf), stdagent переключается на изолированные пути в подкаталогах (`<FallbackDir>/references/<name>.md`) с frontmatter `std-agent-type: <type>` и пояснением в HTML-комментарии, без приватных префиксов std-agent.

## Формат исходных файлов

Полная схема -- в [docs/spec.md](docs/spec.md) Part 1. Минимальный вид:

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

MCP-серверы (`.stdai/standards/mcp.json`):

```json
{
  "version": "1.0",
  "servers": {
    "github": { "type": "stdio", "command": "gh", "args": ["api"] },
    "linear": { "type": "http", "url": "https://mcp.linear.app/sse" }
  }
}
```

## Конфигурация

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

Полная справка: [docs/config-spec.md](docs/config-spec.md).

## Структура проекта

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

Подробности: [docs/file-structure.md](docs/file-structure.md).

## Поддержка monorepo

Если `--config` не указан, `stdagent` идёт вверх от `cwd` к ближайшему `.stdai/config.toml`. Запускайте из любой поддиректории -- корень монорепо будет найден автоматически.

## Разработка

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

## Документация

- **[docs/spec.md](docs/spec.md)** -- полная спецификация: стандарт std-agent + различия 23 инструментов + стратегия конвертации
- [docs/prd.md](docs/prd.md) -- требования к продукту
- [docs/architecture.md](docs/architecture.md) -- модульная структура и потоки данных
- [docs/commands.md](docs/commands.md) -- справочник CLI
- [docs/conversion-rules.md](docs/conversion-rules.md) -- матрица конвертаций + маппинг frontmatter
- [docs/format-spec.md](docs/format-spec.md) -- детальная схема frontmatter
- [docs/file-structure.md](docs/file-structure.md) -- соглашения о структуре директорий
- [docs/roadmap.md](docs/roadmap.md) -- дорожная карта
- [docs/targets/](docs/targets/) -- заметки исследования по каждому инструменту

## License

MIT -- см. [LICENSE](LICENSE).
