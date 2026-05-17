# std-agent

![std-agent: 11 개 AI CLI 도구를 위한 단일 진실 공급원](docs/assets/hero.png)

[![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![CI](https://github.com/StringKe/std-ai/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/StringKe/std-ai/actions/workflows/ci.yml)

[English](README.md) | [简体中文](README.zh-CN.md) | [繁體中文](README.zh-TW.md) | [日本語](README.ja.md) | **한국어** | [Русский](README.ru.md) | [Français](README.fr.md)

---

`stdagent` 는 가볍고 순수 Go 로 구현된 CLI 도구입니다. 프로젝트의 AI 설정을 `.stdai/` 디렉터리 하나에 단일 진실 공급원으로 모아두고, **11 개 AI CLI 도구**로 자동 배포합니다. 각 도구의 네이티브 파일 형식, frontmatter 방언, 잡다한 제약은 모두 대신 처리해 줍니다.

`CLAUDE.md`, `AGENTS.md`, `GEMINI.md`, `.cursor/rules/`, `.windsurf/rules/`, `.clinerules/`, `.github/copilot-instructions.md` 같은 파일을 더 이상 손으로 관리하지 마세요. 한 번만 작성하고, 모든 곳에 동기화하세요.

## 왜 std-agent 인가?

- **단일 공급원** — `rules` / `skills` / `commands` / `references` 를 YAML frontmatter + Markdown 으로 한 번만 작성.
- **11 개 타겟** — Claude Code, Codex, Cursor, GitHub Copilot, Windsurf, Gemini CLI, Aider, Cline, OpenCode, Continue.dev, Antigravity.
- **종속성 제로** — writer 는 허용된 경로만 건드리고, sync 전에 자동 백업하며, `clean` 한 번이면 모두 되돌릴 수 있습니다.
- **드리프트 감지** — `status` 가 외부에서 수정된 파일을 표시하고, `fix` 로 다시 동기화.
- **MCP** — 단일 `.stdai/standards/mcp.json` 이 `.mcp.json` / `.cursor/mcp.json` / `.vscode/mcp.json` 으로 배포됩니다.
- **모노레포 친화** — 설정 탐색이 `cwd` 부터 위로 올라가므로, 어떤 하위 디렉터리에서도 실행 가능.
- **자가 업그레이드** — `stdagent upgrade` 가 GitHub Releases 에서 sha256 검증과 원자 교체로 안전하게 갱신.

## 지원 도구

### Tier 1 (9 개)

| 타겟 | 주요 출력 |
|---|---|
| Claude Code (Anthropic) | `CLAUDE.md` + `.claude/{rules,skills,commands}/` + `.mcp.json` |
| Codex (OpenAI) | `AGENTS.md` + `.agents/skills/` + `.codex/rules/`（바이트 예산 분할） |
| Cursor | `.cursor/{rules/*.mdc,skills,commands}/` + `.cursor/mcp.json` |
| GitHub Copilot | `.github/{copilot-instructions,instructions,prompts,agents}/` + `.vscode/mcp.json` |
| Windsurf (Codeium) | `.windsurf/{rules,skills,workflows}/` |
| Gemini CLI (Google) | `GEMINI.md` + `.gemini/commands/*.toml` |
| Aider | `AGENTS.md` 재사용 (noop) |
| Cline | `.clinerules/` + `.clinerules/workflows/` |
| OpenCode | `.opencode/{agents,commands}/` |

### Tier 2 (2 개)

| 타겟 | 주요 출력 |
|---|---|
| Continue.dev | `.continue/{rules,prompts}/` |
| Antigravity (Google) | `.agents/{rules,workflows}/` |

각 통합 조사는 [docs/targets/](docs/targets/) 에 있습니다.

## 빠른 시작

```bash
# 설치 (macOS / Linux)
curl -fsSL https://raw.githubusercontent.com/StringKe/std-ai/main/install.sh | sh

# 설치 (Windows PowerShell)
irm https://raw.githubusercontent.com/StringKe/std-ai/main/install.ps1 | iex

# 프로젝트에서 초기화
cd your-project
stdagent init

# .stdai/standards/rules/example.md 를 편집한 뒤 모든 활성 타겟으로 동기화
stdagent sync

# 드리프트 확인 / 복구
stdagent status
stdagent fix
```

## 기존 프로젝트를 std-ai 로 마이그레이션

프로젝트에 `CLAUDE.md` / `AGENTS.md` / `.cursor/rules/` / `.clinerules/` / `.github/copilot-instructions.md` 등이 흩어져 있나요? 아래 프롬프트를 Claude Code / Codex / Cursor / Gemini CLI 에 그대로 붙여 넣으면 `.stdai/standards/` 구조로 재구성해 줍니다.

````text
이 프로젝트의 분산된 AI 설정을 std-agent 로 통합 관리하도록 마이그레이션해 주세요. 다음 단계를 따르세요:

1. Glob / Read 로 기존 AI 규칙 파일을 모두 스캔:
   - 루트: CLAUDE.md AGENTS.md GEMINI.md .cursorrules .windsurfrules .clinerules
   - 하위: .claude/rules/ .claude/skills/ .claude/commands/ .claude/agents/
              .cursor/rules/ .windsurf/rules/ .clinerules/ .continue/rules/
              .github/copilot-instructions.md .github/instructions/
              .rulesync/rules/ .codex/AGENTS.md
   - 같은 repo 내 중첩 CLAUDE.md (find . -name CLAUDE.md -not -path './.stdai/*')

2. 인벤토리 보고: rules X 개 / skills Y 개 / commands Z 개 / 중첩 CLAUDE.md N 개,
   그리고 "프로젝트 개요" 내용을 담은 파일을 표시.

3. 분할 계획을 제시한 뒤 내 승인을 기다린 후 파일 작성:
   - 프로젝트 개요 (정의 / 기술 스택 / 철칙 / 유지보수 흐름) -> .stdai/standards/root.md
   - 각 집중된 규칙 -> .stdai/standards/rules/<kebab-name>.md
   - skill 패키지 -> .stdai/standards/skills/<name>/SKILL.md (scripts/ references/ 하위 포함)
   - slash 명령 템플릿 -> .stdai/standards/commands/<name>.md
   - 중첩 CLAUDE.md -> .stdai/standards/nested/<상대경로>/root.md
   - 모든 파일에 frontmatter 추가: type / name / description / priority / applyTo

4. 원문 "리팩토링" 금지: 실행 가능한 명령, API 엔드포인트, 에러 문자열, 파일 경로,
   라인 번호는 모두 그대로 보존. 허용되는 "최적화": 잉여 연결어 삭제, 중복 단락 병합,
   거대한 파일 분할, 구식 도구명 교체.

5. 작업 완료 후 `stdagent sync` 실행을 안내하고, 구 산출물 (.rulesync/ /
   .cursorrules 단일 파일 버전 등) 을 삭제. stdagent 가 생성한 CLAUDE.md /
   AGENTS.md / .claude/rules/ 는 삭제하지 말 것.

전체 사양 (frontmatter 필드표, root.md 템플릿, 중첩 규약, rulesync 마이그레이션 매핑)
은 `stdagent intro` 명령 출력 참조.
````

LLM CLI 에 직접 파이프하는 것도 가능:

```bash
stdagent intro | pbcopy            # macOS: 클립보드에 복사하여 AI 대화에 붙여 넣기
stdagent intro --json              # 에이전트 / 자동화 통합용
```

## 명령 목록

| 명령 | 용도 |
|---|---|
| `stdagent init` | `.stdai/` + `config.toml` + `.stdaiignore` + 샘플 standards 생성 |
| `stdagent pull` | `.stdai/cache/` 에 캐시된 모든 활성 git 소스 갱신 |
| `stdagent sync` | 핵심: pull → parse → convert → 배포 |
| `stdagent fix` | 드리프트 복구를 위한 재동기화 (`sync` 의 별칭) |
| `stdagent status` | 타겟별 드리프트 + 마지막 동기화 시각 |
| `stdagent clean` | 생성 파일 제거 (`.stdai/` 는 보존) |
| `stdagent budget` | LLM 컨텍스트 예산 검사 (문자 + 토큰 추정) |
| `stdagent which <path>` | 해당 파일에 적용되는 rules / references 나열 (AI 의 온디맨드 컨텍스트 로딩) |
| `stdagent intro` | 기존 설정을 변환하기 위한 마이그레이션 프롬프트 출력 |
| `stdagent upgrade` | GitHub Releases 에서 자가 업그레이드 (sha256 + 원자 교체) |
| `stdagent version` | 빌드 정보 |

모든 명령은 `--help` 를 지원합니다. 전체 참조: [docs/commands.md](docs/commands.md).

## 소스 파일 형식

전체 schema 는 [docs/spec.md](docs/spec.md) Part 1 참조. 최소 형태:

```markdown
---
type: rules                       # rules | skills | commands | references
name: coding-style
description: 일반 코딩 스타일
priority: high                    # high | normal | low
targets: [claude-code, codex]     # 명시적 활성화 (또는 exclude_targets 로 제외)
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
inject = true            # 출력 파일에 "Generated by stdagent" footer 삽입
inject_whatis = true     # skills 내부에 출처 주석 한 줄 추가
auto_pull = true         # sync 마다 git 소스 자동 pull
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
├── .stdai/                    내부 관리 영역 (단일 진실 공급원)
│   ├── config.toml            유일한 설정 파일
│   ├── standards/             작성 루트
│   │   ├── rules/
│   │   ├── skills/
│   │   ├── commands/
│   │   ├── references/
│   │   └── mcp.json           MCP 서버 (선택)
│   ├── cache/                 git 소스 캐시
│   ├── backups/               sync 전 자동 스냅샷
│   └── state.json             런타임 상태
├── .stdaiignore               gitignore 스타일 glob, 소스 파일 제외
├── CLAUDE.md                  배포: Claude Code
├── AGENTS.md                  배포: Codex / Cursor fallback / Copilot agent / OpenCode / Antigravity
├── GEMINI.md                  배포: Gemini CLI
├── .mcp.json                  Claude 의 MCP
└── .claude/ .codex/ .cursor/ .github/ .windsurf/ .gemini/ .clinerules/ .opencode/ .continue/ .agents/
```

자세한 내용: [docs/file-structure.md](docs/file-structure.md).

## 모노레포 지원

`--config` 를 명시하지 않으면 `stdagent` 가 `cwd` 부터 위로 올라가 가장 가까운 `.stdai/config.toml` 을 찾습니다. 어떤 하위 디렉터리에서 실행해도 모노레포 루트를 자동으로 찾습니다.

## 개발

```bash
# 도구 체인 (mise + go + golangci-lint + gofumpt + git-cliff)
mise install

# 자주 쓰는 작업
mise run fmt        # gofumpt + goimports
mise run lint       # golangci-lint
mise run test       # go test -race -cover
mise run check      # fmt + lint + test 한 번에
mise run build      # bin/stdagent 빌드
mise run run        # go run ./cmd/stdagent
```

## 문서

- **[docs/spec.md](docs/spec.md)** — 전체 사양: std-ai 표준 + 11 도구 차이 + 변환 전략
- [docs/prd.md](docs/prd.md) — 제품 요구 사항
- [docs/architecture.md](docs/architecture.md) — 모듈 구성과 데이터 흐름
- [docs/commands.md](docs/commands.md) — CLI 명령 사양
- [docs/conversion-rules.md](docs/conversion-rules.md) — 변환 매트릭스 + frontmatter 필드 매핑
- [docs/format-spec.md](docs/format-spec.md) — frontmatter 상세 schema
- [docs/file-structure.md](docs/file-structure.md) — 디렉터리 구조 원칙
- [docs/roadmap.md](docs/roadmap.md) — 로드맵
- [docs/targets/](docs/targets/) — 11 개 타겟 도구 조사

## License

MIT — [LICENSE](LICENSE) 참조.
