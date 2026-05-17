# std-ai 类型速查

stdagent 把 AI 配置分 5 种类型，由 `stdagent sync` 扩散到 22 个 AI CLI 工具的原生格式。源文件统一放 `.stdai/standards/<type>/<name>.md`。

每种类型有不同的**触发语义**（AI 何时加载）与**适用边界**（什么内容该用、什么内容不该用）。

---

## rules

**SEMANTICS**：项目级编码规范 / 操作约束。AI 自动加载（`applyTo` 匹配 / 全局）。session 开始就遵守。

**WHEN TO USE**：写"必须遵守"的硬约束。每条 < 8000 字符，high priority 多条合计 < 16000 字符。

**WHEN NOT**：大段背景知识 -> 用 `references`。一次性任务步骤 -> 用 `commands`。需要主动判断是否调用 -> 用 `skills`。

**Example frontmatter**：

```yaml
---
type: rules
name: exception-handling
description: 异常处理规范
priority: high
applyTo:
  - "**/*.go"
---
```

---

## skills

**SEMANTICS**：按需触发的能力包。AI 看到 `description` 匹配用户意图时主动调用。可以含多文件子目录（`skills/<name>/SKILL.md` + 辅助资源）。

**WHEN TO USE**：写"AI 在 X 场景应当做 Y"的工作流（例如 commit 助手、代码审查流程、调试流程）。description 必须明确触发场景。

**WHEN NOT**：硬规则全程生效 -> 用 `rules`。用户显式调用的固定模板 -> 用 `commands`。

**Example frontmatter**：

```yaml
---
type: skills
name: developer-commit
description: Git 提交助手。自动分析变更并生成 Conventional Commits 信息。触发短语：commit、提交。
---
```

---

## commands

**SEMANTICS**：用户输入 `/command-name` 触发的模板。AI 不主动调用，等用户显式输入 slash command。

**WHEN TO USE**：写固定流程的"操作宏"（例如 `/review`、`/done`），用户每次需要时手动触发。

**WHEN NOT**：AI 自动判断的工作流 -> 用 `skills`。session 全程生效的约束 -> 用 `rules`。

**Example frontmatter**：

```yaml
---
type: commands
name: review
description: 审查当前分支改动并生成 review 报告
---
```

---

## references

**SEMANTICS**：背景参考资料 / 设计文档 / API 速查。AI 仅在需要时查阅，不自动加载到上下文（按需读 / 用 `stdagent which` 查询）。

**WHEN TO USE**：长篇知识库（架构说明、协议规格、外部 API 列表）。容量超过 rules 限制（> 8000 字符）的内容。

**WHEN NOT**：每次都要遵守的约束 -> 用 `rules`。可以主动触发的工作流 -> 用 `skills`。

**Example frontmatter**：

```yaml
---
type: references
name: transformer-design
description: transformer 协议层架构说明
applyTo:
  - "internal/transformer/**"
---
```

---

## subagents

**SEMANTICS**：隔离子代理定义。AI 通过 spawn 子进程或 CLI 调用执行（如 `claude --agent <name>`）。有独立 system prompt 与上下文。

**WHEN TO USE**：需要在干净上下文里完成的任务（代码审查、长 research、并行 dispatch）。每个 subagent 一个独立人格 + 工具白名单。

**WHEN NOT**：当前 session 内的工作流 -> 用 `skills`。简单模板 -> 用 `commands`。

**Example frontmatter**：

```yaml
---
type: subagents
name: code-reviewer
description: 代码审查子代理。读 diff 输出结构化 review 报告。
---
```

---

## 速查表

| 类型 | 触发方式 | 是否自动加载 | 典型长度 |
|---|---|---|---|
| rules | applyTo 匹配 / 全局 | 是 | < 8000 字符 / 条 |
| skills | AI 主动判定 description | 否（按需） | 中（含子目录） |
| commands | 用户输入 `/name` | 否（用户触发） | 短（模板） |
| references | AI 主动查阅 | 否（按需） | 长（知识库） |
| subagents | spawn 子进程 / CLI | 否（隔离） | 中（含 tools 白名单） |

源文件路径：`.stdai/standards/<type>/<name>.md`。用 `stdagent which <file>` 查文件触发的规则集。
