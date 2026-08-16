# Conversion rules

stdagent 将统一 source 转换为 25 个 AI CLI target 的原生配置。本文只定义跨 target 的稳定契约；具体路径、字段和限制以 [targets/](targets/) 的官方证据及对应 adapter 为准。

## Core contract

1. Filter source by `targets` / `exclude_targets`.
2. Sort by `priority` -> `name`.
3. Map each type through the target protocol.
4. Use native capabilities first, then explicit graceful degradation.
5. Canonicalize shared paths and reject incompatible collisions before writing.
6. Atomically write, update state and prune prior managed orphans.

## Protocol families

| protocol | representative targets | primary shape |
|---|---|---|
| `AgentsMD` | codex, amp, warp, factory, jules, grok-build, kimi-code, kiro, goose | `AGENTS.md` plus target sidecars |
| `ClaudeMD` | claude-code, gemini, qwen-code | root document plus imports or native directories |
| `Cursor` | cursor | `.cursor/rules/*.mdc` and native packages |
| `Clinerules` | cline, roo-code, kilo-code | rules directory with target dialect |
| `WindsurfStyle` | windsurf, continue-dev, antigravity | rules, workflows and skills directories |
| `Copilot` | GitHub Copilot | `.github` instructions, prompts, agents and skills |

Targets may add narrowly scoped native output around a protocol plan.

## Shared AGENTS.md

All enabled target plans that produce `AGENTS.md` receive one canonical, target-neutral document:

- Include each rule that applies to at least one enabled `AGENTS.md` consumer.
- Include root and nested rules only.
- Keep commands, skills, references and subagents in target-specific sidecars.
- Every consumer records the same shared content hash.

Any remaining same-path output must be byte-identical and use the same merge semantics. Otherwise sync fails before the first write.

## Root and sidecar behavior

- `root.md` supplies the project entry body.
- Protocols may inline rules into a root or emit rule sidecars with a manifest.
- `nested/<path>/root.md` produces a supported root at `<path>` without the top-level manifest.
- `inject_type_glossary` is opt-in and defaults to false.
- `inject` controls generated markers; `inject_whatis` controls optional origin notes.

Use `stdagent budget --rendered` to inspect exact root and sidecar bytes for enabled targets.

## Frontmatter mapping

Mappings preserve intent rather than field spelling:

- `applyTo` / `globs` -> target path matcher such as `paths`, `globs`, `applyTo` or trigger mode.
- `alwaysApply` -> target always-on representation when supported.
- command fields -> native prompt or command frontmatter.
- skill metadata -> Agent Skills fields supported by the target.
- subagent permissions and runtime fields -> native agent schema when supported.

Unsupported fields are dropped or represented by an explicit degradation file; they are not silently advertised as native support.

## Graceful degradation

When a target lacks a native type, the protocol writes to its configured fallback namespace. Types use isolated subdirectories such as `skills/`, `commands/`, `references/` and `subagents/`; generated metadata identifies the original std-agent type where the target format permits it.

Degradation is valid only if the target can discover or the agent can deliberately read the output. Target research documents any `UNKNOWN`.

## MCP

`.stdai/standards/mcp.json` currently fans out to:

| target | path | top-level key |
|---|---|---|
| claude-code | `.mcp.json` | `mcpServers` |
| cursor | `.cursor/mcp.json` | `mcpServers` |
| copilot | `.vscode/mcp.json` | `servers` |

Other targets keep user-level or tool-owned MCP configuration unchanged.

## Safety and state

- Writes are planned globally before mutation.
- Git submodule boundaries are not crossed.
- Existing output may be backed up according to config.
- Unchanged content is skipped.
- Prune only considers paths tracked for that target plus narrowly defined legacy generated files.
- JSON merge outputs preserve existing scalar values and skip invalid JSON rather than overwriting it.

Implementation authority: `internal/transformer/`, `internal/runner/`, `internal/writer/`, and [targets/](targets/).
