# Rules

`type: rules` 表示 AI 必须持续遵守的编码、架构或操作约束。

适合：

- 违反会造成真实错误、兼容性或安全风险的约束。
- 与代码区域绑定、可用 `applyTo` 观察的规则。
- 短小、稳定、值得占用常驻上下文的原则。

项目入口写入 `.stdai/standards/root.md`；工作流用 skill；用户模板用 command；长背景用 reference。

```yaml
---
type: rules
name: exception-handling
description: Go 错误传播与边界转换
priority: high
applyTo:
  - "**/*.go"
---
```

输出由 target protocol 决定。所有 `AGENTS.md` consumer 共享 target-neutral rules 内容，其他 target 可使用自己的 rules sidecar。以 `stdagent budget --rendered` 查看实际加载形态。
