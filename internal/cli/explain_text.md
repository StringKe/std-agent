# std-agent 类型

`.stdai/standards/` 是 source，`stdagent sync` 将其转换为启用 target 的原生配置。

## rules

持续生效的编码、架构和操作约束。用 `applyTo` 或 target filter 缩小范围；只保留高常驻价值内容。

```yaml
---
type: rules
name: exception-handling
description: Go 错误传播与边界转换
priority: high
applyTo: ["**/*.go"]
---
```

## skills

AI 根据 description 判断是否调用的能力包。主文件为 `skills/<name>/SKILL.md`，可带 scripts、references、templates 或 assets。

```yaml
---
type: skills
name: code-review
description: 审查当前改动并报告回归、安全和正确性问题
---
```

## commands

用户显式输入 `/<name>` 触发的操作模板。它不应承担 session 全程规则。

```yaml
---
type: commands
name: review
description: 审查当前分支
---
```

## references

架构、协议、API 和长篇背景。仅在任务需要时通过 source 路径或 `stdagent which` 查阅，不进入默认常驻规则。

```yaml
---
type: references
name: transformer-design
description: transformer 协议和 adapter 设计
applyTo: ["internal/transformer/**"]
---
```

## subagents

隔离上下文执行的代理定义。用于可独立并行或需要专门上下文的任务；普通流程优先使用 skill。

```yaml
---
type: subagents
name: code-reviewer
description: 在隔离上下文中审查代码
---
```

## 选择标准

| 需求 | 类型 |
|---|---|
| 必须持续遵守 | rules |
| AI 按意图调用工作流 | skills |
| 用户显式触发模板 | commands |
| 需要时查阅背景 | references |
| 隔离执行任务 | subagents |

用 `stdagent which <file>` 查询适用 source，用 `stdagent budget --rendered` 查看实际常驻 root 和 sidecar 体积。
