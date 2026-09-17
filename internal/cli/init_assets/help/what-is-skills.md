# Skills

`type: skills` is a capability package the AI invokes on demand based on description, suited for reusable workflows and domain operations.

```text
.stdai/standards/skills/code-review/
├── SKILL.md
├── scripts/
├── references/
└── templates/
```

```yaml
---
type: skills
name: code-review
description: Review current changes and report correctness, security, and regression issues
---
```

SKILL.md should state the goal and Done means clearly. Only supporting material needed for execution goes in the package; detailed references load on demand. Targets with native Agent Skills get the native directory; others convert via their defined degradation.

Ongoing hard constraints use rules, explicit user templates use commands, and isolated execution alone uses subagents.
