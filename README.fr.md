# std-agent

![std-agent : une seule source de vérité pour 11 outils CLI IA](docs/assets/hero.png)

[![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![CI](https://github.com/StringKe/std-ai/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/StringKe/std-ai/actions/workflows/ci.yml)

[English](README.md) | [简体中文](README.zh-CN.md) | [繁體中文](README.zh-TW.md) | [日本語](README.ja.md) | [한국어](README.ko.md) | [Русский](README.ru.md) | **Français** | [Deutsch](README.de.md) | [Español](README.es.md) | [Português](README.pt-BR.md)

---

`stdagent` est un outil CLI léger, écrit en Go pur. Il conserve un seul répertoire `.stdai/` comme source de vérité pour la configuration IA de votre projet, puis la diffuse vers **11 outils CLI IA**, en prenant en charge pour vous chaque format natif, chaque dialecte de frontmatter et chaque particularité.

Arrêtez de maintenir à la main `CLAUDE.md`, `AGENTS.md`, `GEMINI.md`, `.cursor/rules/`, `.windsurf/rules/`, `.clinerules/`, `.github/copilot-instructions.md`. Écrivez une fois, synchronisez partout.

## Pourquoi std-agent ?

- **Source unique** — écrivez `rules` / `skills` / `commands` / `references` une seule fois en YAML frontmatter + Markdown.
- **Onze cibles** — Claude Code, Codex, Cursor, GitHub Copilot, Windsurf, Gemini CLI, Aider, Cline, OpenCode, Continue.dev, Antigravity.
- **Zéro verrouillage** — le writer ne touche qu'à une liste blanche de chemins ; sauvegarde avant chaque sync ; `clean` annule tout.
- **Détection de drift** — `status` affiche les fichiers modifiés hors de stdagent ; `fix` les réapplique.
- **MCP** — un seul `.stdai/standards/mcp.json` se diffuse vers `.mcp.json` / `.cursor/mcp.json` / `.vscode/mcp.json`.
- **Compatible monorepo** — la recherche de config remonte depuis `cwd` ; fonctionne depuis n'importe quel sous-répertoire.
- **Auto mise à jour** — `stdagent upgrade` télécharge les releases signées depuis GitHub avec vérification sha256 et remplacement atomique.

## Outils supportés

### Tier 1 (9)

| Cible | Sorties principales |
|---|---|
| Claude Code (Anthropic) | `CLAUDE.md` + `.claude/{rules,skills,commands}/` + `.mcp.json` |
| Codex (OpenAI) | `AGENTS.md` + `.agents/skills/` + `.codex/rules/` (spillover en octets) |
| Cursor | `.cursor/{rules/*.mdc,skills,commands}/` + `.cursor/mcp.json` |
| GitHub Copilot | `.github/{copilot-instructions,instructions,prompts,agents}/` + `.vscode/mcp.json` |
| Windsurf (Codeium) | `.windsurf/{rules,skills,workflows}/` |
| Gemini CLI (Google) | `GEMINI.md` + `.gemini/commands/*.toml` |
| Aider | réutilise `AGENTS.md` (noop) |
| Cline | `.clinerules/` + `.clinerules/workflows/` |
| OpenCode | `.opencode/{agents,commands}/` |

### Tier 2 (2)

| Cible | Sorties principales |
|---|---|
| Continue.dev | `.continue/{rules,prompts}/` |
| Antigravity (Google) | `.agents/{rules,workflows}/` |

Chaque intégration est documentée dans [docs/targets/](docs/targets/).

## Démarrage rapide

```bash
# Installation (macOS / Linux)
curl -fsSL https://raw.githubusercontent.com/StringKe/std-ai/main/install.sh | sh

# Installation (Windows PowerShell)
irm https://raw.githubusercontent.com/StringKe/std-ai/main/install.ps1 | iex

# Initialisation dans votre projet
cd your-project
stdagent init

# Éditez .stdai/standards/rules/example.md, puis synchronisez vers toutes les cibles activées
stdagent sync

# Inspection / correction du drift
stdagent status
stdagent fix
```

## Migrer un projet existant vers std-ai

Votre projet est déjà encombré de `CLAUDE.md` / `AGENTS.md` / `.cursor/rules/` / `.clinerules/` / `.github/copilot-instructions.md` ? Collez le prompt ci-dessous dans Claude Code / Codex / Cursor / Gemini CLI et il réorganisera tout dans la structure `.stdai/standards/`.

````text
Aide-moi à migrer ce projet d'une configuration IA dispersée vers std-agent. Procède ainsi :

1. Avec Glob / Read, scanne tous les fichiers de règles IA existants :
   - Racine : CLAUDE.md AGENTS.md GEMINI.md .cursorrules .windsurfrules .clinerules
   - Sous-répertoires : .claude/rules/ .claude/skills/ .claude/commands/ .claude/agents/
              .cursor/rules/ .windsurf/rules/ .clinerules/ .continue/rules/
              .github/copilot-instructions.md .github/instructions/
              .rulesync/rules/ .codex/AGENTS.md
   - CLAUDE.md imbriqués dans le repo (find . -name CLAUDE.md -not -path './.stdai/*')

2. Rends un inventaire : X rules / Y skills / Z commands / N CLAUDE.md imbriqués,
   et signale les fichiers contenant un "aperçu projet".

3. Propose un plan de découpe, attends mon approbation, puis écris les fichiers :
   - Aperçu projet (définition / stack / règles intangibles / flux de maintenance)
     -> .stdai/standards/root.md
   - Chaque règle ciblée -> .stdai/standards/rules/<kebab-name>.md
   - Package skill -> .stdai/standards/skills/<name>/SKILL.md (avec sous-dossiers scripts/ references/)
   - Templates de slash command -> .stdai/standards/commands/<name>.md
   - CLAUDE.md imbriqué -> .stdai/standards/nested/<chemin-relatif>/root.md
   - Chaque fichier reçoit un frontmatter : type / name / description / priority / applyTo

4. Aucune "refactorisation" du contenu original. Conserve toutes les commandes exécutables,
   les API endpoints, les chaînes d'erreur, les chemins de fichiers, les numéros de ligne.
   "Optimisations" autorisées : supprimer les mots de remplissage, fusionner les doublons,
   découper les fichiers trop grands, renommer les outils obsolètes.

5. Une fois terminé, indique-moi de lancer `stdagent sync` et de supprimer les anciens
   artefacts (.rulesync/, .cursorrules version mono-fichier, etc.). NE supprime PAS les
   fichiers générés par stdagent (CLAUDE.md / AGENTS.md / .claude/rules/).

La spec complète (table des champs frontmatter, modèle root.md, layout imbriqué, mapping
de migration rulesync) est dans la sortie de la commande `stdagent intro`.
````

Ou pipez directement dans un CLI LLM :

```bash
stdagent intro | pbcopy            # macOS : copier dans le presse-papier puis coller dans le chat IA
stdagent intro --json              # pour intégrations agent / automatisation
```

## Commandes

| Commande | Rôle |
|---|---|
| `stdagent init` | Crée `.stdai/` + `config.toml` + `.stdaiignore` + standards d'exemple |
| `stdagent pull` | Met à jour les sources git cachées dans `.stdai/cache/` |
| `stdagent sync` | Cœur : pull → parse → convert → diffusion |
| `stdagent fix` | Re-sync pour corriger le drift (alias de `sync`) |
| `stdagent status` | Drift et dernier sync par cible |
| `stdagent clean` | Supprime les fichiers générés (préserve `.stdai/`) |
| `stdagent budget` | Vérification de budget de contexte LLM (caractères + estimation tokens) |
| `stdagent intro` | Imprime un prompt de migration pour qu'un LLM convertisse votre config existante |
| `stdagent upgrade` | Auto-mise à jour depuis GitHub Releases (sha256 + remplacement atomique) |
| `stdagent version` | Infos de build |

Chaque commande supporte `--help`. Référence complète : [docs/commands.md](docs/commands.md).

## Format source

Le schéma complet est dans [docs/spec.md](docs/spec.md) Part 1. Forme minimale :

```markdown
---
type: rules                       # rules | skills | commands | references
name: coding-style
description: Style de code général
priority: high                    # high | normal | low
targets: [claude-code, codex]     # opt-in (ou exclude_targets pour exclure)
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

## Arborescence projet

```
your-project/
├── .stdai/                    Zone de gestion interne (source de vérité)
│   ├── config.toml            Fichier de config unique
│   ├── standards/             Racine d'écriture
│   │   ├── rules/
│   │   ├── skills/
│   │   ├── commands/
│   │   ├── references/
│   │   └── mcp.json           Serveurs MCP (optionnel)
│   ├── cache/                 Cache des sources git
│   ├── backups/               Snapshot auto avant chaque sync
│   └── state.json             État runtime
├── .stdaiignore               glob style gitignore pour exclure des sources
├── CLAUDE.md                  Diffusion : Claude Code
├── AGENTS.md                  Diffusion : Codex / Cursor fallback / Copilot agent / OpenCode / Antigravity
├── GEMINI.md                  Diffusion : Gemini CLI
├── .mcp.json                  MCP pour Claude
└── .claude/ .codex/ .cursor/ .github/ .windsurf/ .gemini/ .clinerules/ .opencode/ .continue/ .agents/
```

Détails : [docs/file-structure.md](docs/file-structure.md).

## Support monorepo

Quand `--config` est omis, `stdagent` remonte depuis `cwd` pour trouver le `.stdai/config.toml` le plus proche. Lancez-le depuis n'importe quel sous-répertoire, la racine du monorepo est localisée automatiquement.

## Développement

```bash
# Toolchain (mise + go + golangci-lint + gofumpt + git-cliff)
mise install

# Tâches courantes
mise run fmt        # gofumpt + goimports
mise run lint       # golangci-lint
mise run test       # go test -race -cover
mise run check      # fmt + lint + test en une commande
mise run build      # produit bin/stdagent
mise run run        # go run ./cmd/stdagent
```

## Documentation

- **[docs/spec.md](docs/spec.md)** — spec complète : standard std-ai + divergences 11 outils + stratégie de conversion
- [docs/prd.md](docs/prd.md) — exigences produit
- [docs/architecture.md](docs/architecture.md) — découpe modules et flux de données
- [docs/commands.md](docs/commands.md) — référence CLI
- [docs/conversion-rules.md](docs/conversion-rules.md) — matrice de conversion + mapping frontmatter
- [docs/format-spec.md](docs/format-spec.md) — schéma détaillé du frontmatter
- [docs/file-structure.md](docs/file-structure.md) — conventions de répertoire
- [docs/roadmap.md](docs/roadmap.md) — feuille de route
- [docs/targets/](docs/targets/) — notes de recherche par outil (11)

## License

MIT — voir [LICENSE](LICENSE).
