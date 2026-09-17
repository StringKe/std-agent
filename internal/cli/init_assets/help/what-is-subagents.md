# Subagents

`type: subagents` defines agents in an isolated context. Use only when a task runs independently, needs a dedicated context, or parallelizes safely.

```yaml
---
type: subagents
name: code-reviewer
description: Review code in an isolated context and return an issue list
readonly: true
---
```

The body is the agent instruction and should include a clear deliverable and success criteria. The model field may be omitted to use the runtime default. Targets with native support convert to their agent format; others degrade per protocol.

Reusable flows within the current session use skills; ongoing rules use rules; user templates use commands.
