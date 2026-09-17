# std-agent types

`.stdai/standards/` is the source; `stdagent sync` converts it into the native config of each enabled target.

## rules

Always-on coding, architecture, and operational constraints. Narrow the scope with `applyTo` or target filters; keep only high resident-value content.

```yaml
---
type: rules
name: exception-handling
description: Go error propagation and boundary conversion
priority: high
applyTo: ["**/*.go"]
---
```

## skills

Capability packages the AI invokes based on description. The main file is `skills/<name>/SKILL.md`, optionally with scripts, references, templates, or assets.

```yaml
---
type: skills
name: code-review
description: Review current changes and report regressions, security, and correctness issues
---
```

## commands

Operation templates triggered by the user explicitly typing `/<name>`. They must not carry session-wide rules.

```yaml
---
type: commands
name: review
description: Review the current branch
---
```

## references

Architecture, protocols, APIs, and long-form background. Consult via the source path or `stdagent which` only when the task needs them; never part of the default resident rules.

```yaml
---
type: references
name: transformer-design
description: Transformer protocol and adapter design
applyTo: ["internal/transformer/**"]
---
```

## subagents

Agent definitions executed in an isolated context. Use for tasks that parallelize independently or need a dedicated context; prefer skills for ordinary flows.

```yaml
---
type: subagents
name: code-reviewer
description: Review code in an isolated context
---
```

## Quick reference

| Need | Type |
|---|---|
| Must always be followed | rules |
| AI-invoked workflow by intent | skills |
| User-explicit trigger template | commands |
| Background consulted as needed | references |
| Isolated execution task | subagents |

Use `stdagent which <file>` to query the applicable sources, and `stdagent budget --rendered` to see the actual resident root and sidecar sizes.
