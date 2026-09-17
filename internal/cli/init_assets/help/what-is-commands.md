# Commands

`type: commands` is an operation template the user triggers by explicitly typing `/<name>`.

```yaml
---
type: commands
name: release
description: Verify and release the current version
argument_hint: "[version]"
---
```

A command should describe the outcome, preconditions, success criteria, and any required recovery strategy. Use the runtime default model unless the target's protocol or the user explicitly requires a pinned model.

Workflows the AI judges on its own use skills; ongoing constraints use rules. Output paths are decided by the target protocol; targets without native commands use the defined degradation.
