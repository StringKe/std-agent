# Commands

`type: commands` 是用户显式输入 `/<name>` 触发的操作模板。

```yaml
---
type: commands
name: release
description: 验证并发布当前版本
argument_hint: "[version]"
---
```

Command 应描述结果、前置条件、成功标准和必要恢复策略。使用 runtime 默认模型，除非该 target 的协议或用户明确要求固定模型。

AI 自动判断的工作流使用 skill；持续约束使用 rule。输出路径由 target protocol 决定，缺少原生 command 的 target 使用已定义的 degradation。
