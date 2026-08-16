# std-agent

![std-agent : une seule source de vérité pour 23 outils CLI IA](docs/assets/hero.png)

[![Release](https://img.shields.io/github/v/release/StringKe/std-agent?sort=semver)](https://github.com/StringKe/std-agent/releases)
[![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go)](https://go.dev/)
[![Go Report Card](https://goreportcard.com/badge/github.com/StringKe/std-agent)](https://goreportcard.com/report/github.com/StringKe/std-agent)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![CI](https://github.com/StringKe/std-agent/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/StringKe/std-agent/actions/workflows/ci.yml)

[English](README.md) | [简体中文](README.zh-CN.md) | [繁體中文](README.zh-TW.md) | [日本語](README.ja.md) | [한국어](README.ko.md) | [Русский](README.ru.md) | **Français** | [Deutsch](README.de.md) | [Español](README.es.md) | [Português](README.pt-BR.md)

---

`stdagent` est un outil CLI léger, écrit en Go pur, qui conserve un seul répertoire `.stdai/` comme source de vérité pour la configuration IA de votre projet, puis la diffuse vers **23 outils CLI IA** en prenant en charge pour vous chaque format de fichier natif, chaque dialecte de frontmatter et chaque particularité.

Arrêtez de maintenir à la main `CLAUDE.md`, `AGENTS.md`, `GEMINI.md`, `.cursor/rules/`, `.windsurf/rules/`, `.clinerules/`, `.github/copilot-instructions.md`, etc. Écrivez une fois, synchronisez partout.

## Pourquoi std-agent ?

- **Source unique** -- écrivez `rules` / `skills` / `commands` / `references` / `subagents` une seule fois en YAML frontmatter + Markdown.
- **Vingt-cinq cibles** -- Claude Code, Codex, Cursor, GitHub Copilot, Windsurf/Devin, Gemini CLI, Aider, Cline, OpenCode, Roo Code, Crush, Amp, Warp, Factory, Continue.dev, Antigravity, Qwen Code, Pi, Kilo Code, Augment Code, Jules, Grok Build, Kimi Code, Kiro, Goose.
- **Conformité aux spécifications** -- chaque chemin de sortie, chaque dialecte de frontmatter et chaque limite de taille est vérifié par rapport à la documentation officielle des outils (dernier audit complet : 2026-07) ; les répertoires natifs Agent Skills sont utilisés partout où ils existent.
- **Zéro verrouillage** -- le writer ne touche qu'à une petite liste blanche de chemins ; sauvegarde avant chaque sync ; `clean` annule tout.
- **Détection de drift** -- `status` affiche les fichiers modifiés hors de stdagent ; `fix` les réapplique depuis la source.
- **MCP** -- un seul `.stdai/standards/mcp.json` se diffuse vers `.mcp.json` / `.cursor/mcp.json` / `.vscode/mcp.json`.
- **Compatible monorepo** -- la recherche de config remonte depuis `cwd` ; fonctionne depuis n'importe quel sous-répertoire.
- **Auto mise à jour** -- `stdagent upgrade` télécharge les releases signées depuis GitHub avec vérification sha256 et remplacement atomique.

## Outils supportés

### Tier 1 (14)

| Cible | Sorties principales |
|---|---|
| Claude Code (Anthropic) | `CLAUDE.md` + `.claude/{rules,skills,commands,agents}/` + `.mcp.json` |
| Codex (OpenAI) | `AGENTS.md` + `.agents/skills/` + `.codex/agents/*.toml` (subagents natifs) |
| Cursor | `.cursor/{rules/*.mdc,skills,commands,agents}/` + `.cursor/mcp.json` |
| GitHub Copilot | `.github/{copilot-instructions,instructions,prompts,agents,skills}/` + `.vscode/mcp.json` |
| Windsurf / Devin (Cognition) | `.windsurf/{rules,skills,workflows}/` + miroir `.devin/rules/` |
| Gemini CLI (Google) | `GEMINI.md` + `.gemini/skills/` + `.gemini/commands/*.toml` |
| Aider | réutilise `AGENTS.md` (noop) |
| Cline | `.clinerules/` (préfixes numériques 100/500/900) |
| OpenCode | `.opencode/{skills,commands}/` |
| Roo Code | `.roo/{rules,skills,commands}/` |
| Crush (Charmbracelet) | `CRUSH.md` + `.crush/skills/` + enregistrement des skills dans `crush.json` |
| Amp (Sourcegraph) | `AGENTS.md` (inline) + `.agents/skills/` |
| Warp | `AGENTS.md` (inline + imbriqué) + `.agents/skills/` |
| Factory (Factory.ai) | `.factory/{rules,skills,commands,droids}/` |

### Tier 2 (11)

| Cible | Sorties principales |
|---|---|
| Continue.dev | `.continue/{rules,skills,prompts}/` + `rules.md` imbriqué |
| Antigravity (Google) | `.agents/{rules,skills,workflows}/` |
| Qwen Code (Alibaba) | `QWEN.md` + `.qwen/{rules,skills,commands}/` |
| Pi | `.pi/skills/` + `.pi/prompts/` |
| Kilo Code (kilo.ai) | `.kilo/{rules,skills,command}/` + enregistrement des instructions dans `kilo.jsonc` |
| Augment Code | `.augment/{rules,skills}/` |
| Jules (Google) | `AGENTS.md` |
| Grok Build (xAI) | `AGENTS.md` + `.grok/skills/` |
| Kimi Code (Moonshot AI) | `AGENTS.md` + `.agents/skills/` |
| Kiro (AWS) | `AGENTS.md` + `.kiro/{steering,skills,agents}/` |
| Goose (AAIF) | `AGENTS.md` + `.agents/skills/` |

Chaque intégration est documentée dans [docs/targets/](docs/targets/).

## Démarrage rapide

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

## Migrer un projet existant vers std-agent

Votre projet est déjà encombré de `CLAUDE.md` / `AGENTS.md` / `.cursor/rules/` / `.clinerules/` / `.github/copilot-instructions.md` ? Collez le prompt ci-dessous dans Claude Code / Codex / Cursor / Gemini CLI et il réorganisera tout dans la structure `.stdai/standards/`.

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

Ou envoyez directement le résultat dans un CLI LLM :

```bash
stdagent intro | pbcopy            # macOS: paste into AI chat
stdagent intro --json              # for agent / automation integrations
```

## Commandes

| Commande | Rôle |
|---|---|
| `stdagent init` | Crée `.stdai/` + `config.toml` + `.stdaiignore` + standards d'exemple |
| `stdagent pull` | Met à jour les sources git cachées dans `.stdai/cache/` |
| `stdagent sync` | Cœur : pull -> parse -> convert -> diffusion |
| `stdagent fix` | Re-sync pour corriger le drift (alias de `sync`) |
| `stdagent status` | Drift et dernier sync par cible |
| `stdagent clean` | Supprime les fichiers générés (préserve `.stdai/`) |
| `stdagent budget` | Vérification de budget de contexte LLM (caractères + estimation de tokens) |
| `stdagent which <path>` | Liste les rules / references applicables à un fichier (chargement contextuel à la demande pour l'IA) |
| `stdagent explain` | Affiche pour l'IA le sens des 5 types std-agent (rules/skills/commands/references/subagents) |
| `stdagent intro` | Imprime un prompt de migration pour qu'un LLM convertisse votre config existante |
| `stdagent upgrade` | Auto-mise à jour depuis GitHub Releases (sha256 + remplacement atomique) |
| `stdagent version` | Infos de build |

Chaque commande supporte `--help`. Référence complète : [docs/commands.md](docs/commands.md).

## Architecture à base de protocoles

La v0.0.4 a introduit une architecture transformer à trois couches : le `Plan()` de chaque cible délègue à l'un des 6 protocoles (AgentsMD / ClaudeMD / Cursor / Clinerules / WindsurfStyle / Copilot), paramétré par un littéral de struct `protocol.Adapter`. Ajouter un nouvel outil coûte désormais ~25-35 lignes au lieu de 145 (déduplication de code de 60-70%).

Dégradation progressive (graceful degradation) : lorsqu'une cible ne supporte pas nativement un type std-agent (par exemple les references partout, les subagents dans amp / windsurf), stdagent bascule vers des chemins isolés dans un sous-répertoire (`<FallbackDir>/references/<name>.md`) avec un frontmatter `std-agent-type: <type>` et une explication en commentaire HTML, sans préfixes privés propres à std-agent.

## Format source

Le schéma complet est dans [docs/spec.md](docs/spec.md) Part 1. Forme minimale :

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

Serveurs MCP (`.stdai/standards/mcp.json`) :

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

`.stdai/config.toml` :

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

Référence complète : [docs/config-spec.md](docs/config-spec.md).

## Arborescence du projet

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

Détails : [docs/file-structure.md](docs/file-structure.md).

## Support monorepo

Quand `--config` est omis, `stdagent` remonte depuis `cwd` pour trouver le `.stdai/config.toml` le plus proche. Lancez-le depuis n'importe quel sous-répertoire : la racine du monorepo est localisée automatiquement.

## Développement

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

- **[docs/spec.md](docs/spec.md)** -- spec complète : standard std-agent + divergences des 23 outils + stratégie de conversion
- [docs/prd.md](docs/prd.md) -- exigences produit
- [docs/architecture.md](docs/architecture.md) -- découpe des modules et flux de données
- [docs/commands.md](docs/commands.md) -- référence CLI
- [docs/conversion-rules.md](docs/conversion-rules.md) -- matrice de conversion + mapping frontmatter
- [docs/format-spec.md](docs/format-spec.md) -- schéma détaillé du frontmatter
- [docs/file-structure.md](docs/file-structure.md) -- conventions de répertoire
- [docs/roadmap.md](docs/roadmap.md) -- feuille de route
- [docs/targets/](docs/targets/) -- notes de recherche par outil

## License

MIT -- voir [LICENSE](LICENSE).
