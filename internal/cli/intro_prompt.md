# stdagent AI 配置迁移与维护

目标：以 `.stdai/standards/` 为项目 AI 配置的单一真相源，让 `stdagent sync` 为启用的 target 生成可直接消费、无冲突、可验证的原生配置。

## Done means

- 所有现有 AI 配置已完整盘点，项目事实、硬约束和可执行信息没有丢失。
- 内容按 rules、skills、commands、references、subagents 正确归类，重复和冲突已消解。
- `.stdai/standards/root.md` 只保留项目入口和跨领域硬约束，详细内容按需加载。
- `stdagent sync --strict`、`stdagent status` 和 `stdagent budget --rendered` 通过。
- 仅在确认生成物可恢复且已被新源覆盖后处理旧产物。

## 工作模型

AI 维护 source，stdagent 维护 rendered output：

```text
.stdai/standards/ -> parse -> target protocols -> CLAUDE.md / AGENTS.md / target sidecars
```

不要直接编辑带 stdagent marker 的根文件或 target 目录。修改源后重新 sync。

### 五种类型

| 类型 | 用途 | 触发 |
|---|---|---|
| `rules` | 必须持续遵守的编码、架构和操作约束 | target 自动加载或按路径匹配 |
| `skills` | AI 根据任务意图调用的能力包 | description 匹配时按需加载 |
| `commands` | 用户显式触发的操作模板 | `/<name>` |
| `references` | 规格、架构和长篇背景 | 需要时查阅 |
| `subagents` | 适合隔离上下文的代理定义 | runtime spawn |

分类以消费语义为准，不以原文件名为准。target 缺少原生能力时，由 transformer 使用已定义的 graceful degradation。

## 首次迁移

适用条件：项目已有分散配置，但 `.stdai/standards/` 尚未成为真相源。

### 1. 完整盘点

读取仓库规则、根级说明、target 配置目录、skills、commands、agents、嵌套指令和相关文档。包括隐藏目录与同仓库嵌套根文件。区分：

- 项目事实与全局硬约束。
- 路径或领域专属规则。
- 可复用工作流与用户命令。
- 长篇参考资料。
- 生成物、手写源和过时工具残留。

先报告文件与类型清单、重复或冲突、可能的 source owner，以及任何 `UNKNOWN`。

### 2. 设计 source 边界

- `root.md`：项目定义、结构、关键技术栈、少量硬约束和完成门槛。
- `rules/<name>.md`：一个可观察主题一份规则，使用 `applyTo` 限定范围。
- `skills/<name>/SKILL.md`：自包含能力包，辅助资料放同 package。
- `commands/<name>.md`：用户显式触发的固定操作。
- `references/<name>.md`：不应常驻上下文的背景。
- `subagents/<name>.md`：只有隔离执行确有价值时使用。
- `nested/<relative-path>/root.md`：同一 git 仓库的目录级附加说明。

高影响拆分或语义冲突先确认边界，再写入。

### 3. 优化内容

以 outcome 和成功标准为中心：

- 删除角色化开场、思考过程、重复说明、显然可从上下文推断的内容和冗长示例。
- 将大量具体规则压缩为可观察原则，但保留安全、合规、破坏性操作和机器解析格式。
- 保留真实命令、协议字段、路径、端点、错误字符串及其适用条件。
- 过时工具流程改为当前 stdagent 语义；无法确认的事实标记 `UNKNOWN`。
- 默认关闭 `inject_type_glossary`。只有下游确实需要内嵌类型说明时才开启。

目标是减少 50%-80% 的无效上下文，而不是机械缩短仍有独立价值的领域信息。

### 4. 写入与验证

最小 rule frontmatter：

```yaml
---
type: rules
name: coding-style
description: 适用范围和核心约束
priority: high
applyTo:
  - "**/*.go"
---
```

`name` 使用 kebab-case。`targets` 与 `exclude_targets` 互斥；省略表示适用于所有启用 target。完整字段以 `docs/format-spec.md` 为准。

验证：

```bash
stdagent sync --strict
stdagent status
stdagent budget --rendered
```

rendered budget 用于确认每个 target 的实际 root 常驻体积和 sidecar 体积。共享 `AGENTS.md` 只含对至少一个启用 AGENTS consumer 生效的 rules；commands、skills、references 和 subagents 留在 target sidecar。

清理旧文件前确认它是可恢复的旧生成物，不是仍有唯一信息的手写源。优先使用 `stdagent clean` 或可恢复删除方式。

## 日常维护

项目已有 `.stdai/standards/` 时：

1. 读取相关 source、代码和 target 证据。
2. 只修改 `.stdai/standards/` 或 transformer 实现。
3. 运行 sync、status 和与风险匹配的测试。
4. 不手改生成物，不把临时状态或 TODO 写入最终规则。

删除 source 后，`sync` 默认按 state prune 旧生成物。跨 git submodule 的配置由各 submodule 独立维护，父仓库 sync 不应写入其边界。

## 根文件与嵌套

- `root.md` 是项目入口，不手写自动 manifest。
- 同一路径被多个 target 消费时必须有单一 owner 或字节一致。
- 所有 `AGENTS.md` producer 共享 canonical rules 内容，消除 target 顺序导致的覆盖。
- `nested/<path>/root.md` 输出到 `<path>` 的受支持根文件，不携带顶级 manifest。
- root 只放常驻价值高的内容；细节放 rule、skill 或 reference。

## 命令

| 命令 | 结果 |
|---|---|
| `stdagent init` | 初始化 `.stdai/` |
| `stdagent sync` | 解析并生成启用 target 的配置 |
| `stdagent status` | 检查 rendered output drift |
| `stdagent fix` | 重新 sync 修复 drift |
| `stdagent which <path>` | 查询某文件适用的 source |
| `stdagent explain [type]` | 查看类型语义 |
| `stdagent budget --rendered` | 比较 source 与实际 target 上下文体积 |
| `stdagent clean` | 清理 state 管理的生成物，保留 source |

完成回复应列出修改后的 source、验证结果及 `UNKNOWN`，不要求读者了解迁移过程。
