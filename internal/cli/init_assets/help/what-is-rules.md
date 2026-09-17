# Rules

`type: rules` marks coding, architecture, or operational constraints the AI must always follow.

Good fit:

- Constraints whose violation causes real errors, compatibility, or security risks.
- Rules bound to code areas and observable via `applyTo`.
- Short, stable principles worth occupying resident context.

The project entrypoint goes in `.stdai/standards/root.md`; workflows use skills, user templates use commands, long background uses references.

```yaml
---
type: rules
name: exception-handling
description: Go error propagation and boundary conversion
priority: high
applyTo:
  - "**/*.go"
---
```

Output is decided by the target protocol. All `AGENTS.md` consumers share the target-neutral rules content; other targets may use their own rules sidecars. Use `stdagent budget --rendered` to see the actual loaded shape.
