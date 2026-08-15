# Subagents

`type: subagents` 定义隔离上下文中的代理。只在任务能独立执行、需要专门上下文或可安全并行时使用。

```yaml
---
type: subagents
name: code-reviewer
description: 在隔离上下文中审查代码并返回问题清单
readonly: true
---
```

Body 是代理指令，应包含明确产出和成功标准。模型字段可省略以使用 runtime 默认值。target 原生支持时转换为其 agent 格式，否则按协议降级。

当前 session 内的复用流程使用 skill；持续规则使用 rule；用户模板使用 command。
