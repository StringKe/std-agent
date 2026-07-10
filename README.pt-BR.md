# std-agent

![std-agent: uma única fonte da verdade para 11 ferramentas CLI de IA](docs/assets/hero.png)

[![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![CI](https://github.com/StringKe/std-agent/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/StringKe/std-agent/actions/workflows/ci.yml)

[English](README.md) | [简体中文](README.zh-CN.md) | [繁體中文](README.zh-TW.md) | [日本語](README.ja.md) | [한국어](README.ko.md) | [Русский](README.ru.md) | [Français](README.fr.md) | [Deutsch](README.de.md) | [Español](README.es.md) | **Português**

---

`stdagent` é uma ferramenta CLI leve, escrita em Go puro. Mantém um único diretório `.stdai/` como fonte da verdade para a configuração de IA do seu projeto e a distribui para **11 ferramentas CLI de IA**, cuidando dos formatos nativos, dialetos de frontmatter e particularidades de cada ferramenta para você.

Pare de manter `CLAUDE.md`, `AGENTS.md`, `GEMINI.md`, `.cursor/rules/`, `.windsurf/rules/`, `.clinerules/`, `.github/copilot-instructions.md` na mão. Escreva uma vez, sincronize em todo lugar.

## Por que std-agent?

- **Fonte única** — escreva `rules` / `skills` / `commands` / `references` apenas uma vez em YAML frontmatter + Markdown.
- **Onze destinos** — Claude Code, Codex, Cursor, GitHub Copilot, Windsurf, Gemini CLI, Aider, Cline, OpenCode, Continue.dev, Antigravity.
- **Zero lock-in** — o writer só toca em uma whitelist pequena de caminhos; backup antes de cada sync; `clean` reverte tudo.
- **Detecção de drift** — `status` mostra arquivos modificados fora do stdagent; `fix` reaplica.
- **MCP** — um único `.stdai/standards/mcp.json` se distribui para `.mcp.json` / `.cursor/mcp.json` / `.vscode/mcp.json`.
- **Pronto para monorepo** — busca de config sobe a partir de `cwd`; funciona de qualquer subdiretório.
- **Auto-upgrade** — `stdagent upgrade` baixa releases assinados do GitHub com verificação sha256 e substituição atômica.

## Ferramentas suportadas

### Tier 1 (9)

| Destino | Saídas principais |
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

| Destino | Saídas principais |
|---|---|
| Continue.dev | `.continue/{rules,prompts}/` |
| Antigravity (Google) | `.agents/{rules,workflows}/` |

Cada integração está documentada em [docs/targets/](docs/targets/).

## Início rápido

```bash
# Instalação (macOS / Linux)
curl -fsSL https://raw.githubusercontent.com/StringKe/std-agent/main/install.sh | sh

# Instalação (Windows PowerShell)
irm https://raw.githubusercontent.com/StringKe/std-agent/main/install.ps1 | iex

# Inicialize no seu projeto
cd your-project
stdagent init

# Edite .stdai/standards/rules/example.md e sincronize para todos os targets ativos
stdagent sync

# Inspecionar / corrigir drift
stdagent status
stdagent fix
```

## Migrar um projeto existente para std-agent

Projeto já cheio de `CLAUDE.md` / `AGENTS.md` / `.cursor/rules/` / `.clinerules/` / `.github/copilot-instructions.md`? Cole o prompt abaixo no Claude Code / Codex / Cursor / Gemini CLI e ele vai reorganizar tudo na estrutura `.stdai/standards/`.

````text
Me ajuda a migrar este projeto de uma configuração de IA dispersa para o std-agent. Faça:

1. Use Glob / Read para escanear todos os arquivos de regras IA existentes:
   - Raiz: CLAUDE.md AGENTS.md GEMINI.md .cursorrules .windsurfrules .clinerules
   - Subdiretórios: .claude/rules/ .claude/skills/ .claude/commands/ .claude/agents/
              .cursor/rules/ .windsurf/rules/ .clinerules/ .continue/rules/
              .github/copilot-instructions.md .github/instructions/
              .rulesync/rules/ .codex/AGENTS.md
   - CLAUDE.md aninhado dentro do repo (find . -name CLAUDE.md -not -path './.stdai/*')

2. Reporte um inventário: X rules / Y skills / Z commands / N CLAUDE.md aninhados,
   e indique quais arquivos contêm "visão geral do projeto".

3. Proponha um plano de divisão, aguarde minha aprovação e escreva os arquivos:
   - Visão geral do projeto (definição / stack / regras de ouro / fluxo de manutenção)
     -> .stdai/standards/root.md
   - Cada regra focada -> .stdai/standards/rules/<kebab-name>.md
   - Pacote skill -> .stdai/standards/skills/<name>/SKILL.md (com subdiretórios scripts/ references/)
   - Templates de slash command -> .stdai/standards/commands/<name>.md
   - CLAUDE.md aninhado -> .stdai/standards/nested/<caminho-relativo>/root.md
   - Cada arquivo recebe frontmatter: type / name / description / priority / applyTo

4. Sem "refatoração" do conteúdo original. Preserve todos os comandos executáveis,
   endpoints de API, strings de erro, caminhos de arquivos, números de linha.
   "Otimizações" permitidas: remover palavras de transição, juntar duplicatas,
   dividir arquivos grandes, renomear ferramentas obsoletas.

5. Ao terminar, me avise para rodar `stdagent sync` e remover artefatos antigos
   (.rulesync/, .cursorrules em arquivo único etc.). NÃO apague os arquivos que o
   próprio stdagent produz (CLAUDE.md / AGENTS.md / .claude/rules/).

A spec completa (tabela de campos frontmatter, template root.md, layout aninhado,
mapeamento de migração rulesync) está na saída do comando `stdagent intro`.
````

Também dá para piping direto num LLM CLI:

```bash
stdagent intro | pbcopy            # macOS: copia para clipboard e cola no chat IA
stdagent intro --json              # para integrações agente / automação
```

## Comandos

| Comando | Função |
|---|---|
| `stdagent init` | Cria `.stdai/` + `config.toml` + `.stdaiignore` + standards de exemplo |
| `stdagent pull` | Atualiza as fontes git em cache `.stdai/cache/` |
| `stdagent sync` | Núcleo: pull → parse → convert → distribuição |
| `stdagent fix` | Re-sync para corrigir drift (alias de `sync`) |
| `stdagent status` | Drift e último sync por destino |
| `stdagent clean` | Remove arquivos gerados (preserva `.stdai/`) |
| `stdagent budget` | Verificação de orçamento de contexto LLM (caracteres + estimativa de tokens) |
| `stdagent which <path>` | Lista as rules / references aplicáveis a um arquivo (contexto sob demanda para IA) |
| `stdagent intro` | Imprime um prompt de migração para um LLM converter sua config existente |
| `stdagent upgrade` | Auto-upgrade do GitHub Releases (sha256 + substituição atômica) |
| `stdagent version` | Info de build |

Todo comando suporta `--help`. Referência completa: [docs/commands.md](docs/commands.md).

## Formato dos arquivos-fonte

Schema completo em [docs/spec.md](docs/spec.md) Part 1. Forma mínima:

```markdown
---
type: rules                       # rules | skills | commands | references
name: coding-style
description: Estilo de código geral
priority: high                    # high | normal | low
targets: [claude-code, codex]     # opt-in (ou exclude_targets para excluir)
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
inject = true            # insere footer "Generated by stdagent" nas saídas
inject_whatis = true     # acrescenta nota de origem em uma linha dentro de skills
auto_pull = true         # faz pull das fontes git a cada sync
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

Referência completa: [docs/config-spec.md](docs/config-spec.md).

## Layout do projeto

```
your-project/
├── .stdai/                    Área de gestão interna (fonte da verdade)
│   ├── config.toml            Único arquivo de configuração
│   ├── standards/             Raiz de autoria
│   │   ├── rules/
│   │   ├── skills/
│   │   ├── commands/
│   │   ├── references/
│   │   └── mcp.json           Servidores MCP (opcional)
│   ├── cache/                 Cache das fontes git
│   ├── backups/               Snapshot automático antes de cada sync
│   └── state.json             Estado de runtime
├── .stdaiignore               globs no estilo gitignore para excluir fontes
├── CLAUDE.md                  Distribuição: Claude Code
├── AGENTS.md                  Distribuição: Codex / Cursor fallback / Copilot agent / OpenCode / Antigravity
├── GEMINI.md                  Distribuição: Gemini CLI
├── .mcp.json                  MCP para Claude
└── .claude/ .codex/ .cursor/ .github/ .windsurf/ .gemini/ .clinerules/ .opencode/ .continue/ .agents/
```

Detalhes: [docs/file-structure.md](docs/file-structure.md).

## Suporte a monorepo

Quando `--config` é omitido, `stdagent` sobe a partir de `cwd` até achar o `.stdai/config.toml` mais próximo. Rode de qualquer subdiretório, a raiz do monorepo é localizada automaticamente.

## Desenvolvimento

```bash
# Toolchain (mise + go + golangci-lint + gofumpt + git-cliff)
mise install

# Tarefas comuns
mise run fmt        # gofumpt + goimports
mise run lint       # golangci-lint
mise run test       # go test -race -cover
mise run check      # fmt + lint + test em um comando só
mise run build      # produz bin/stdagent
mise run run        # go run ./cmd/stdagent
```

## Documentação

- **[docs/spec.md](docs/spec.md)** — spec completa: padrão std-agent + divergência das 11 ferramentas + estratégia de conversão
- [docs/prd.md](docs/prd.md) — requisitos de produto
- [docs/architecture.md](docs/architecture.md) — módulos e fluxo de dados
- [docs/commands.md](docs/commands.md) — referência CLI
- [docs/conversion-rules.md](docs/conversion-rules.md) — matriz de conversão + mapping de frontmatter
- [docs/format-spec.md](docs/format-spec.md) — schema detalhado do frontmatter
- [docs/file-structure.md](docs/file-structure.md) — convenções de diretórios
- [docs/roadmap.md](docs/roadmap.md) — roadmap
- [docs/targets/](docs/targets/) — notas de pesquisa por ferramenta (11)

## License

MIT — ver [LICENSE](LICENSE).
