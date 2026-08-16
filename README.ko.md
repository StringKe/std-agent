# std-agent

![std-agent: 23 개 AI CLI 도구를 위한 단일 진실 공급원](docs/assets/hero.png)

[![Release](https://img.shields.io/github/v/release/StringKe/std-agent?sort=semver)](https://github.com/StringKe/std-agent/releases)
[![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go)](https://go.dev/)
[![Go Report Card](https://goreportcard.com/badge/github.com/StringKe/std-agent)](https://goreportcard.com/report/github.com/StringKe/std-agent)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![CI](https://github.com/StringKe/std-agent/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/StringKe/std-agent/actions/workflows/ci.yml)

[English](README.md) | [简体中文](README.zh-CN.md) | [繁體中文](README.zh-TW.md) | [日本語](README.ja.md) | **한국어** | [Русский](README.ru.md) | [Français](README.fr.md) | [Deutsch](README.de.md) | [Español](README.es.md) | [Português](README.pt-BR.md)

---

`stdagent` 는 가볍고 순수 Go 로 구현된 CLI 도구입니다. 프로젝트의 AI 설정을 위한 단일 진실 공급원으로 `.stdai/` 디렉터리 하나만 유지하고, 이를 **23 개 AI CLI 도구**로 확산시키면서 각 도구의 네이티브 파일 형식, frontmatter 방언, 세부적인 특이 사항까지 대신 처리해 줍니다.

`CLAUDE.md`, `AGENTS.md`, `GEMINI.md`, `.cursor/rules/`, `.windsurf/rules/`, `.clinerules/`, `.github/copilot-instructions.md` 같은 파일을 더 이상 손으로 관리하지 마세요. 한 번만 작성하고, 모든 곳에 동기화하세요.

## 왜 std-agent 인가?

- **단일 공급원** -- `rules` / `skills` / `commands` / `references` / `subagents` 를 YAML frontmatter + Markdown 으로 한 번만 작성.
- **25 개 타겟** -- Claude Code, Codex, Cursor, GitHub Copilot, Windsurf/Devin, Gemini CLI, Aider, Cline, OpenCode, Roo Code, Crush, Amp, Warp, Factory, Continue.dev, Antigravity, Qwen Code, Pi, Kilo Code, Augment Code, Jules, Grok Build, Kimi Code, Kiro, Goose.
- **스펙 정확성** -- 모든 출력 경로, frontmatter 방언, 크기 제한을 각 도구의 공식 문서와 대조해 검증합니다 (최근 전체 감사: 2026-07). 네이티브 Agent Skills 디렉터리가 존재하는 곳에는 그대로 사용합니다.
- **종속성 제로** -- writer 는 화이트리스트에 등록된 소수의 경로만 건드립니다. 매 sync 전에 백업하며, `clean` 으로 모든 변경을 되돌릴 수 있습니다.
- **드리프트 감지** -- `status` 가 stdagent 밖에서 수정된 파일을 표시하고, `fix` 로 소스를 다시 적용합니다.
- **MCP** -- 단일 `.stdai/standards/mcp.json` 이 `.mcp.json` / `.cursor/mcp.json` / `.vscode/mcp.json` 으로 확산됩니다.
- **모노레포 친화** -- 설정 탐색이 `cwd` 부터 위로 올라가므로, 어떤 하위 디렉터리에서도 실행 가능.
- **자가 업그레이드** -- `stdagent upgrade` 가 GitHub Releases 에서 sha256 검증과 원자 교체로 안전하게 갱신.

## 지원 도구

### Tier 1 (14 개)

| 타겟 | 주요 출력 |
|---|---|
| Claude Code (Anthropic) | `CLAUDE.md` + `.claude/{rules,skills,commands,agents}/` + `.mcp.json` |
| Codex (OpenAI) | `AGENTS.md` + `.agents/skills/` + `.codex/agents/*.toml` (네이티브 subagents) |
| Cursor | `.cursor/{rules/*.mdc,skills,commands,agents}/` + `.cursor/mcp.json` |
| GitHub Copilot | `.github/{copilot-instructions,instructions,prompts,agents,skills}/` + `.vscode/mcp.json` |
| Windsurf / Devin (Cognition) | `.windsurf/{rules,skills,workflows}/` + `.devin/rules/` 미러 |
| Gemini CLI (Google) | `GEMINI.md` + `.gemini/skills/` + `.gemini/commands/*.toml` |
| Aider | `AGENTS.md` 재사용 (noop) |
| Cline | `.clinerules/` (100/500/900 숫자 프리픽스) |
| OpenCode | `.opencode/{skills,commands}/` |
| Roo Code | `.roo/{rules,skills,commands}/` |
| Crush (Charmbracelet) | `CRUSH.md` + `.crush/skills/` + `crush.json` skills 등록 |
| Amp (Sourcegraph) | `AGENTS.md` (인라인) + `.agents/skills/` |
| Warp | `AGENTS.md` (인라인 + 중첩) + `.agents/skills/` |
| Factory (Factory.ai) | `.factory/{rules,skills,commands,droids}/` |

### Tier 2 (11 개)

| 타겟 | 주요 출력 |
|---|---|
| Continue.dev | `.continue/{rules,skills,prompts}/` + 중첩 `rules.md` |
| Antigravity (Google) | `.agents/{rules,skills,workflows}/` |
| Qwen Code (Alibaba) | `QWEN.md` + `.qwen/{rules,skills,commands}/` |
| Pi | `.pi/skills/` + `.pi/prompts/` |
| Kilo Code (kilo.ai) | `.kilo/{rules,skills,command}/` + `kilo.jsonc` instructions 등록 |
| Augment Code | `.augment/{rules,skills}/` |
| Jules (Google) | `AGENTS.md` |
| Grok Build (xAI) | `AGENTS.md` + `.grok/skills/` |
| Kimi Code (Moonshot AI) | `AGENTS.md` + `.agents/skills/` |
| Kiro (AWS) | `AGENTS.md` + `.kiro/{steering,skills,agents}/` |
| Goose (AAIF) | `AGENTS.md` + `.agents/skills/` |

각 통합 조사는 [docs/targets/](docs/targets/) 에 있습니다.

## 빠른 시작

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

## 기존 프로젝트를 std-agent 로 마이그레이션

프로젝트에 `CLAUDE.md` / `AGENTS.md` / `.cursor/rules/` / `.clinerules/` / `.github/copilot-instructions.md` 등이 흩어져 있나요? 아래 프롬프트를 Claude Code / Codex / Cursor / Gemini CLI 에 그대로 붙여 넣으면 `.stdai/standards/` 구조로 재구성해 줍니다.

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

LLM CLI 에 직접 파이프하는 것도 가능:

```bash
stdagent intro | pbcopy            # macOS: paste into AI chat
stdagent intro --json              # for agent / automation integrations
```

## 명령 목록

| 명령 | 용도 |
|---|---|
| `stdagent init` | `.stdai/` + `config.toml` + `.stdaiignore` + 샘플 standards 생성 |
| `stdagent pull` | `.stdai/cache/` 에 캐시된 git 소스 갱신 |
| `stdagent sync` | 핵심: pull -> parse -> convert -> 확산 |
| `stdagent fix` | 드리프트 복구를 위한 재동기화 (`sync` 의 별칭) |
| `stdagent status` | 타겟별 드리프트 + 마지막 동기화 시각 |
| `stdagent clean` | 생성 파일 제거 (`.stdai/` 는 보존) |
| `stdagent budget` | LLM 컨텍스트 예산 검사 (문자 + 토큰 추정) |
| `stdagent which <path>` | 해당 파일에 적용되는 rules / references 나열 (AI 의 온디맨드 컨텍스트 로딩) |
| `stdagent explain` | AI 를 위한 std-agent 5 가지 타입 의미 (rules/skills/commands/references/subagents) 출력 |
| `stdagent intro` | 기존 설정을 변환하기 위한 마이그레이션 프롬프트 출력 |
| `stdagent upgrade` | GitHub Releases 에서 자가 업그레이드 (sha256 + 원자 교체) |
| `stdagent version` | 빌드 정보 |

모든 명령은 `--help` 를 지원합니다. 전체 참조: [docs/commands.md](docs/commands.md).

## 프로토콜 기반 아키텍처

v0.0.4 에서 3 계층 transformer 아키텍처가 도입되었습니다: 각 타겟의 `Plan()` 은 6 개 프로토콜 (AgentsMD / ClaudeMD / Cursor / Clinerules / WindsurfStyle / Copilot) 중 하나에 위임되며, `protocol.Adapter` struct 리터럴로 파라미터화됩니다. 새 도구를 추가하는 비용은 이제 145 줄이 아니라 약 25-35 줄입니다 (코드 중복 60-70% 제거).

우아한 저하 (graceful degradation): 타겟이 std-agent 타입을 네이티브로 지원하지 않는 경우 (예: 모든 곳의 references, amp / windsurf 의 subagents), stdagent 는 하위 디렉터리로 격리된 경로 (`<FallbackDir>/references/<name>.md`) 로 폴백하며, frontmatter `std-agent-type: <type>` 와 HTML 주석 설명을 추가합니다. std-agent 전용 프리픽스는 사용하지 않습니다.

## 소스 파일 형식

전체 schema 는 [docs/spec.md](docs/spec.md) Part 1 참조. 최소 형태:

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

MCP 서버 (`.stdai/standards/mcp.json`):

```json
{
  "version": "1.0",
  "servers": {
    "github": { "type": "stdio", "command": "gh", "args": ["api"] },
    "linear": { "type": "http", "url": "https://mcp.linear.app/sse" }
  }
}
```

## 설정 예시

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

전체 참조: [docs/config-spec.md](docs/config-spec.md).

## 프로젝트 레이아웃

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

자세한 내용: [docs/file-structure.md](docs/file-structure.md).

## 모노레포 지원

`--config` 를 명시하지 않으면 `stdagent` 가 `cwd` 부터 위로 올라가 가장 가까운 `.stdai/config.toml` 을 찾습니다. 어떤 하위 디렉터리에서 실행해도 모노레포 루트를 자동으로 찾습니다.

## 개발

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

## 문서

- **[docs/spec.md](docs/spec.md)** -- 전체 사양: std-agent 표준 + 23 개 도구 차이 + 변환 전략
- [docs/prd.md](docs/prd.md) -- 제품 요구 사항
- [docs/architecture.md](docs/architecture.md) -- 모듈 구성과 데이터 흐름
- [docs/commands.md](docs/commands.md) -- CLI 명령 사양
- [docs/conversion-rules.md](docs/conversion-rules.md) -- 변환 매트릭스 + frontmatter 필드 매핑
- [docs/format-spec.md](docs/format-spec.md) -- frontmatter 상세 schema
- [docs/file-structure.md](docs/file-structure.md) -- 디렉터리 구조 원칙
- [docs/roadmap.md](docs/roadmap.md) -- 로드맵
- [docs/targets/](docs/targets/) -- 도구별 조사 노트

## License

MIT -- [LICENSE](LICENSE) 참조.
