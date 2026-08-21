# std-agent

![std-agent: uma única fonte de verdade para 25 ferramentas de CLI de IA](docs/assets/hero.png)

[![Release](https://img.shields.io/github/v/release/StringKe/std-agent?sort=semver)](https://github.com/StringKe/std-agent/releases)
[![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go)](https://go.dev/)
[![Go Report Card](https://goreportcard.com/badge/github.com/StringKe/std-agent)](https://goreportcard.com/report/github.com/StringKe/std-agent)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![CI](https://github.com/StringKe/std-agent/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/StringKe/std-agent/actions/workflows/ci.yml)

[English](README.md) | [简体中文](README.zh-CN.md) | [繁體中文](README.zh-TW.md) | [日本語](README.ja.md) | [한국어](README.ko.md) | [Русский](README.ru.md) | [Français](README.fr.md) | [Deutsch](README.de.md) | [Español](README.es.md) | **Português**

---

`stdagent` é uma CLI leve, escrita em Go puro, que mantém um único diretório `.stdai/` como fonte de verdade para a configuração de IA do seu projeto e depois a distribui para **25 ferramentas de CLI de IA**, cuidando dos formatos de arquivo nativos, dialetos de frontmatter e peculiaridades de cada uma.

Pare de manter `CLAUDE.md`, `AGENTS.md`, `GEMINI.md`, `.cursor/rules/`, `.windsurf/rules/`, `.clinerules/`, `.github/copilot-instructions.md`, ... manualmente. Edite uma vez, sincronize em todo lugar.

## Por que o std-agent?

- **Fonte única** -- escreva `rules` / `skills` / `commands` / `references` / `subagents` uma vez em frontmatter YAML + Markdown.
- **Vinte e cinco destinos** -- Claude Code, Codex, Cursor, GitHub Copilot, Windsurf/Devin, Gemini CLI, Aider, Cline, OpenCode, Crush, Amp, Warp, Factory, Junie, Antigravity, Qwen Code, Pi, Kilo Code, Augment Code, Jules, Grok Build, Kimi Code, Kiro, Goose, Zed.
- **Fiel à especificação** -- cada caminho de saída, dialeto de frontmatter e limite de tamanho é verificado com a documentação oficial das ferramentas (última auditoria completa: 2026-07); diretórios nativos de Agent Skills onde existirem.
- **Zero lock-in** -- o writer só toca em uma pequena whitelist de caminhos; backups antes de cada sync; `clean` reverte tudo.
- **Detecção de drift** -- `status` mostra arquivos modificados fora do stdagent; `fix` reaplica a fonte.
- **MCP** -- um único `.stdai/standards/mcp.json` se distribui para `.mcp.json` / `.cursor/mcp.json` / `.vscode/mcp.json`
- **Compatível com monorepo** -- a busca de configuração sobe a partir do `cwd`; funciona a partir de qualquer subdiretório.
- **Autoatualizável** -- `stdagent upgrade` baixa releases assinados do GitHub com verificação sha256 e substituição atômica.

## Ferramentas suportadas

### Tier 1 (14)

| Destino | Saídas principais |
|---|---|
| Claude Code (Anthropic) | `CLAUDE.md` + `.claude/{rules,skills,commands,agents}/` + `.mcp.json` |
| Codex (OpenAI) | `AGENTS.md` + `.agents/skills/` + `.codex/agents/*.toml` (subagentes nativos) |
| Cursor | `.cursor/{rules/*.mdc,skills,commands,agents}/` + `.cursor/mcp.json` |
| GitHub Copilot | `.github/{copilot-instructions,instructions,prompts,agents,skills}/` + `.vscode/mcp.json` |
| Windsurf / Devin (Cognition) | `.windsurf/{rules,skills,workflows}/` + espelho em `.devin/rules/` |
| Gemini CLI (Google) | `GEMINI.md` + `.gemini/skills/` + `.gemini/commands/*.toml` |
| Aider | reutiliza `AGENTS.md` (noop) |
| Cline | `.clinerules/` (prefixos numéricos 100/500/900) |
| OpenCode | `.opencode/{skills,commands}/` |
| Junie (JetBrains) | `AGENTS.md` compartilhado + `.junie/{rules,skills}/` |
| Crush (Charmbracelet) | `CRUSH.md` + `.crush/skills/` + registro de skills em `crush.json` |
| Amp (Sourcegraph) | `AGENTS.md` (inline) + `.agents/skills/` |
| Warp | `AGENTS.md` (inline + aninhado) + `.agents/skills/` |
| Factory (Factory.ai) | `.factory/{rules,skills,commands,droids}/` |

### Tier 2 (11)

| Destino | Saídas principais |
|---|---|
| Zed | `AGENTS.md` compartilhado + `.agents/skills/` |
| Antigravity (Google) | `.agents/{rules,skills,workflows}/` |
| Qwen Code (Alibaba) | `QWEN.md` + `.qwen/{rules,skills,commands}/` |
| Pi | `.pi/skills/` + `.pi/prompts/` |
| Kilo Code (kilo.ai) | `.kilo/{rules,skills,command}/` + registro de instructions em `kilo.jsonc` |
| Augment Code | `.augment/{rules,skills}/` |
| Jules (Google) | `AGENTS.md` |
| Grok Build (xAI) | `AGENTS.md` + `.grok/skills/` |
| Kimi Code (Moonshot AI) | `AGENTS.md` + `.agents/skills/` |
| Kiro (AWS) | `AGENTS.md` + `.kiro/{steering,skills,agents}/` |
| Goose (AAIF) | `AGENTS.md` + `.agents/skills/` |

Cada integração está documentada em [docs/targets/](docs/targets/). Este repositório ativa apenas o Grok Build; os modos de gitignore e um fan-out pequeno estão em [examples/](examples/).

## Início rápido

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

## Migrar um projeto existente para o std-agent

Seu projeto já está cheio de `CLAUDE.md` / `AGENTS.md` / `.cursor/rules/` / `.clinerules/` / `.github/copilot-instructions.md`? Cole o prompt abaixo no Claude Code / Codex / Cursor / Gemini CLI e ele reorganizará tudo em `.stdai/standards/` para você.

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

Envie também direto para uma CLI de LLM:

```bash
stdagent intro | pbcopy            # macOS: paste into AI chat
stdagent intro --json              # for agent / automation integrations
```

## Comandos

| Comando | Finalidade |
|---|---|
| `stdagent init` | Cria a estrutura de `.stdai/` + `config.toml` + `.stdaiignore` + standards de exemplo |
| `stdagent pull` | Atualiza as fontes baseadas em git armazenadas em cache em `.stdai/cache/` |
| `stdagent sync` | Núcleo: pull -> parse -> convert -> fan out |
| `stdagent fix` | Sincroniza novamente para corrigir o drift (alias de `sync`) |
| `stdagent status` | Drift por destino + hora do último sync |
| `stdagent clean` | Remove os arquivos gerados (preserva `.stdai/`) |
| `stdagent budget` | Verificação do orçamento de contexto do LLM (caracteres + estimativa de tokens) |
| `stdagent which <path>` | Lista rules / references aplicáveis a um arquivo (roteamento de contexto sob demanda para IA) |
| `stdagent explain` | Imprime a semântica dos 5 tipos do std-agent (rules/skills/commands/references/subagents) para IA |
| `stdagent intro` | Imprime um prompt de migração para um LLM converter sua configuração existente |
| `stdagent upgrade` | Autoatualização a partir do GitHub Releases (sha256 + substituição atômica) |
| `stdagent version` | Informações do build |

Todos os comandos suportam `--help`. Referência completa: [docs/commands.md](docs/commands.md).

## Arquitetura baseada em protocolos

A v0.0.4 introduziu uma arquitetura de transformer em três camadas: o `Plan()` de cada destino delega para um dos 6 protocolos (AgentsMD / ClaudeMD / Cursor / Clinerules / WindsurfStyle / Copilot), parametrizado por um struct literal `protocol.Adapter`. Adicionar uma nova ferramenta agora custa ~25-35 linhas em vez de 145 (60-70% de deduplicação de código).

Degradação graciosa: quando um destino não suporta nativamente um tipo do std-agent (por exemplo, references em todo lugar, subagents em amp / windsurf), o stdagent recorre a caminhos de subdiretório isolados (`<FallbackDir>/references/<name>.md`) com frontmatter `std-agent-type: <type>` + um comentário HTML explicativo, sem prefixos privados do std-agent.

## Formato de origem

Um esquema completo está em [docs/spec.md](docs/spec.md), Parte 1. A forma mínima:

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

## Configuração

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

Referência completa: [docs/config-spec.md](docs/config-spec.md).

## Estrutura do projeto

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

Detalhes: [docs/file-structure.md](docs/file-structure.md).

## Suporte a monorepo

Quando `--config` é omitido, o `stdagent` sobe a partir do `cwd` para encontrar o `.stdai/config.toml` mais próximo. Execute-o a partir de qualquer subdiretório e ele localizará a raiz do monorepo automaticamente.

## Desenvolvimento

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

## Documentação

- **[docs/spec.md](docs/spec.md)** -- especificação completa: padrão std-agent + divergência das 25 ferramentas + estratégia de conversão
- [docs/prd.md](docs/prd.md) -- requisitos do produto
- [docs/architecture.md](docs/architecture.md) -- estrutura de módulos + fluxo de dados
- [docs/commands.md](docs/commands.md) -- referência de comandos da CLI
- [docs/conversion-rules.md](docs/conversion-rules.md) -- matriz de conversão + mapeamento de frontmatter
- [docs/format-spec.md](docs/format-spec.md) -- detalhes do esquema de frontmatter
- [docs/file-structure.md](docs/file-structure.md) -- convenções de diretórios
- [docs/roadmap.md](docs/roadmap.md) -- roadmap
- [docs/targets/](docs/targets/) -- notas de pesquisa por ferramenta

## Licença

MIT -- veja [LICENSE](LICENSE).
