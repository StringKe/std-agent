# Skills

`type: skills` 是 AI 根据 description 按需调用的能力包，适合可复用工作流和领域操作。

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
description: 审查当前改动并报告正确性、安全和回归问题
---
```

SKILL.md 应明确目标和 Done means。只有执行所需的辅助资料进入 package；详细参考按需读取。target 有原生 Agent Skills 时写入原生目录，否则按其已定义的 degradation 转换。

持续硬约束使用 rule，用户显式模板使用 command，隔离执行才使用 subagent。
