# stdagent AI config migration and maintenance

Goal: use `.stdai/standards/` as the single source of truth for the project's AI config, so `stdagent sync` generates directly consumable, conflict-free, verifiable native config for every enabled target.

## Done means

- All existing AI configs have been fully inventoried; project facts, hard constraints, and actionable information are preserved.
- Content is correctly classified as rules, skills, commands, references, or subagents; duplicates and conflicts are resolved.
- `.stdai/standards/root.md` keeps only the project entrypoint and cross-domain hard constraints; details load on demand.
- `stdagent sync --strict`, `stdagent status`, and `stdagent budget --rendered` pass.
- Old artifacts are handled only after confirming they are recoverable and fully covered by the new sources.

## Working model

The AI maintains sources; stdagent maintains rendered output:

```text
.stdai/standards/ -> parse -> target protocols -> CLAUDE.md / AGENTS.md / target sidecars
```

Never edit root files carrying a stdagent marker or target directories directly. Re-sync after changing sources.

### The five types

| Type | Purpose | Trigger |
|---|---|---|
| `rules` | Coding, architecture, and operational constraints that must always be followed | Auto-loaded by targets or matched by path |
| `skills` | Capability packages the AI invokes based on task intent | Loaded on demand when the description matches |
| `commands` | Operation templates the user triggers explicitly | `/<name>` |
| `references` | Specs, architecture, and long-form background | Consulted when needed |
| `subagents` | Agent definitions suited for isolated contexts | runtime spawn |

Classify by consumption semantics, not by the original filename. When a target lacks a native capability, the transformer applies the defined graceful degradation.

## First-time migration

Applies when the project already has scattered configs but `.stdai/standards/` is not yet the source of truth.

### 1. Full inventory

Read repo rules, root-level notes, target config dirs, skills, commands, agents, nested instructions, and related docs. Include hidden dirs and nested root files in the same repo. Distinguish:

- Project facts and global hard constraints.
- Path- or domain-specific rules.
- Reusable workflows and user commands.
- Long-form reference material.
- Artifacts, hand-written sources, and stale tool leftovers.

First report the file and type list, duplicates or conflicts, likely source owners, and anything `UNKNOWN`.

### 2. Design source boundaries

- `root.md`: project definition, structure, key stack, a few hard constraints, and the done bar.
- `rules/<name>.md`: one rule per observable topic, scoped with `applyTo`.
- `skills/<name>/SKILL.md`: self-contained capability packages, with supporting material in the same package.
- `commands/<name>.md`: fixed operations the user triggers explicitly.
- `references/<name>.md`: background that should not stay resident in context.
- `subagents/<name>.md`: only when isolated execution is genuinely valuable.
- `nested/<relative-path>/root.md`: directory-level notes for the same git repo.

Confirm boundaries before writing when a split is high-impact or semantics conflict.

### 3. Optimize content

Center on outcomes and success criteria:

- Delete role-play intros, thinking traces, duplicated explanations, content obvious from context, and verbose examples.
- Compress piles of specific rules into observable principles, but keep safety, compliance, destructive-operation, and machine-parsed formats.
- Keep real commands, protocol fields, paths, endpoints, error strings, and their applicability conditions.
- Rewrite stale tool flows in current stdagent semantics; mark unconfirmable facts `UNKNOWN`.
- Keep `inject_type_glossary` off by default. Enable it only when downstream consumers truly need the embedded type glossary.

Aim to cut 50%-80% of ineffective context, not to mechanically shorten domain information that still has standalone value.

### 4. Write and verify

Minimal rule frontmatter:

```yaml
---
type: rules
name: coding-style
description: Scope and core constraints
priority: high
applyTo:
  - "**/*.go"
---
```

Use kebab-case for `name`. `targets` and `exclude_targets` are mutually exclusive; omitting both means all enabled targets. See `docs/format-spec.md` for the full field reference.

Verify:

```bash
stdagent sync --strict
stdagent status
stdagent budget --rendered
```

The rendered budget confirms each target's actual root resident size and sidecar size. The shared `AGENTS.md` only contains rules effective for at least one enabled AGENTS consumer; commands, skills, references, and subagents stay in target sidecars.

Before deleting old files, confirm each is a recoverable old artifact, not a hand-written source with unique information. Prefer `stdagent clean` or another recoverable deletion.

## Daily maintenance

When the project already has `.stdai/standards/`:

1. Read the relevant sources, code, and target evidence.
2. Change only `.stdai/standards/` or the transformer implementation.
3. Run sync, status, and the tests matching the risk.
4. Never hand-edit artifacts, and never write transient state or TODOs into final rules.

After deleting a source, `sync` prunes the old artifacts via state by default. Configs across git submodules are maintained independently by each submodule; the parent repo sync must not write across their boundaries.

## Root files and nesting

- `root.md` is the project entrypoint; never hand-write the automatic manifest.
- When one path is consumed by multiple targets, it must have a single owner or byte-identical content.
- All `AGENTS.md` producers share the canonical rules content, eliminating target-order overwrites.
- `nested/<path>/root.md` renders into the supported root file under `<path>`, without the top-level manifest.
- The root keeps only high resident-value content; details go in rules, skills, or references.

## Commands

| Command | Result |
|---|---|
| `stdagent init` | Initialize `.stdai/` |
| `stdagent sync` | Parse and generate config for enabled targets |
| `stdagent status` | Check rendered output drift |
| `stdagent fix` | Re-sync to fix drift |
| `stdagent which <path>` | Query the applicable sources for a file |
| `stdagent explain [type]` | Show type semantics |
| `stdagent budget --rendered` | Compare source vs. actual target context sizes |
| `stdagent clean` | Clean state-managed artifacts, keep sources |

The completion reply should list the changed sources, verification results, and `UNKNOWN` items; the reader should not need to know the migration process.
