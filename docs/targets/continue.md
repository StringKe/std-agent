# Target: Continue.dev

调研日期: 2026-05-08
官方文档: https://docs.continue.dev
GitHub: https://github.com/continuedev/continue

## 1. 摘要

Continue.dev 是开源 AI 编码助手，VS Code 与 JetBrains 双扩展形态，
配置以 `config.yaml` 为核心，新格式（YAML）取代旧 `config.json`
（仍可读但已 deprecated）。配置存在三层：用户级 `~/.continue/`、
项目级 `.continue/` 与 `.continuerc.json` 工作区覆写文件，外加 Continue Hub
通过 `uses:` 指令引入远端 blocks。

Continue 把扩展点拆分成细粒度 blocks：rules / prompts / models /
mcpServers / docs / data / context，每类既可在 `config.yaml` 内联，也可
作为单独文件（如 `.continue/rules/*.md`、`.continue/prompts/*.prompt.md`）
放进文件夹自动加载，或通过 Hub slug `owner/package` 远程引用。

Rules 支持 `globs` / `regex` / `alwaysApply` 三种激活模式，和 Cursor
rules 系统语义接近。Slash commands 通过 prompts blocks 注册，文件名
即命令名。MCP 直接走 `mcpServers` 字段，仅在 agent 模式可用。
(https://docs.continue.dev/customize/overview, 访问日期 2026-05-08)

## 2. 配置文件路径

| 类型 | 路径 | 说明 |
|---|---|---|
| 用户全局（YAML） | `~/.continue/config.yaml` | 主配置，跨项目生效 |
| 用户全局（JSON, 旧） | `~/.continue/config.json` | deprecated，仍兼容读取 |
| 用户全局规则目录 | `~/.continue/rules/*.md` | 全局 rules，按字典序加载 |
| 用户全局 prompts | `~/.continue/prompts/*.prompt.md` | 全局 slash prompts |
| 用户全局 assistants | `~/.continue/assistants/*.yaml` | 多 assistant 定义 |
| 项目级 rules 目录 | `<repo>/.continue/rules/*.md` | 项目 rules，自动检测 |
| 项目级 prompts 目录 | `<repo>/.continue/prompts/*.prompt.md` | 项目 slash 命令 |
| 工作区覆写 | `<repo>/.continuerc.json` | merge / overwrite 行为可控 |
| 高级 TS 配置 | `~/.continue/config.ts` 或项目同名 | 程序化扩展 |

加载顺序：用户 `config.yaml` -> 工作区 `.continuerc.json` 按
`mergeBehavior` 合并（默认 `merge`，可选 `overwrite`），rules / prompts
目录中的文件按字典序追加。Hub blocks 通过 `uses:` 在任意位置插入。
(https://docs.continue.dev/customize/deep-dives/configuration,
访问日期 2026-05-08;
https://docs.continue.dev/customize/deep-dives/rules,
访问日期 2026-05-08)

## 3. 文件格式

| 文件 | 格式 | frontmatter |
|---|---|---|
| `config.yaml` | YAML | 无 frontmatter，顶层字段为 schema/name/version 等 |
| `.continuerc.json` | JSON | 同 config.json，多一个 `mergeBehavior` |
| `.continue/rules/*.md` | Markdown + YAML frontmatter | `name` / `globs` / `regex` / `description` / `alwaysApply` |
| `.continue/prompts/*.prompt.md` | Markdown + YAML frontmatter | `name` / `description` / `version` / `invokable` |
| `config.ts` | TypeScript | 不适用 |

(https://docs.continue.dev/customize/deep-dives/rules,
访问日期 2026-05-08;
https://docs.continue.dev/customize/deep-dives/prompts,
访问日期 2026-05-08)

## 4. config.yaml 关键字段

```yaml
name: my-assistant
version: 1.0.0
schema: v1

models:
  - name: claude
    provider: anthropic
    model: claude-sonnet-4-6

rules:
  - "Always use TypeScript interfaces for object shapes"
  - uses: owner/typescript-rules     # Hub slug 引用

prompts:
  - name: review
    description: Code review checklist
    prompt: "Review for security and perf"
  - uses: supabase/create-functions  # Hub prompt

mcpServers:
  - name: SQLite MCP
    type: stdio
    command: npx
    args: ["-y", "mcp-sqlite", "/path/to/db"]
    env:
      KEY: ${{ secrets.KEY }}

docs:
  - name: React Docs
    startUrl: https://react.dev/reference

context:
  - provider: codebase
  - provider: docs
```

Rules 文件示例（`.continue/rules/typescript.md`）：

```markdown
---
name: TypeScript Best Practices
globs: ["**/*.ts", "**/*.tsx"]
alwaysApply: false
description: Standards for TypeScript development
---

- Always use TypeScript interfaces for object shapes
- Use strict null checks
```

`alwaysApply` 取值语义：`true` 始终注入；`false` 仅当 `globs` 匹配或
模型基于 `description` 决定时注入；`undefined`（默认）当未声明 `globs`
时始终注入，否则按 globs 匹配。
(https://docs.continue.dev/customize/deep-dives/rules,
访问日期 2026-05-08;
https://docs.continue.dev/customize/deep-dives/mcp,
访问日期 2026-05-08)

## 5. std-ai 四类映射

| std-ai 类型 | Continue 落点 | 加载方式 |
|---|---|---|
| rules | `.continue/rules/*.md` 或 `config.yaml` 顶层 `rules:` 数组 | 自动扫描目录，按字典序追加；frontmatter 控制 globs / alwaysApply |
| skills | 无原生 skill 概念。降级为带 `description` 的 model-decision rule（`alwaysApply: false`，仅描述触发） | 由模型读 description 决定；不是真正运行时切换 |
| commands | `.continue/prompts/*.prompt.md`，frontmatter 设 `invokable: true`，文件名即 `/<name>` slash 命令 | 输入 `/` 触发选择面板 |
| references | `docs:` 字段（远程 doc 索引）或内联到 rule 正文。`@<file>` 在 chat 中按需引入 | docs 字段构建嵌入索引；rule 内嵌为静态文本 |

注：Continue 没有 Cursor / Antigravity 那种独立 skill 文件夹概念。若
std-ai 输入有 skill，建议作为 `.continue/rules/<skill-name>.md` 写入，
通过 `description` + `alwaysApply: false` 实现 model-decision 触发。
(https://docs.continue.dev/customize/deep-dives/rules,
访问日期 2026-05-08;
https://docs.continue.dev/customize/deep-dives/prompts,
访问日期 2026-05-08)

## 6. 转换器实现要点

1. v1.0 默认产出 `<repo>/.continue/rules/<name>.md`：每条 std-ai rule
   一个文件，frontmatter 写 `name` / `description` / `globs` /
   `alwaysApply`。globs 缺省时只写 `description`，由 model 决定
2. std-ai commands -> `<repo>/.continue/prompts/<name>.prompt.md`，
   frontmatter 必带 `invokable: true`，正文为命令模板
3. 不主动写 `<repo>/config.yaml`：用户多用 Hub assistant 或全局
   `~/.continue/config.yaml`，避免冲突。仅在 `init --continue`
   显式开关时写工作区 `config.yaml` 并附 marker
4. references -> 在 rule 正文用 `@docs/...` 引用或写到 `docs:` 字段
   （需用户许可，因为 docs 索引会触发远程抓取）
5. MCP 不在四类映射范围；保留旁路：若用户自带 `mcpServers:`，
   不修改

## 7. 信息来源

- https://docs.continue.dev/customize/overview （访问日期 2026-05-08）
- https://docs.continue.dev/customize/deep-dives/configuration
  （访问日期 2026-05-08）
- https://docs.continue.dev/customize/deep-dives/rules
  （访问日期 2026-05-08）
- https://docs.continue.dev/customize/deep-dives/prompts
  （访问日期 2026-05-08）
- https://docs.continue.dev/customize/deep-dives/mcp
  （访问日期 2026-05-08）
- https://docs.continue.dev/reference （访问日期 2026-05-08）
- https://github.com/continuedev/continue （访问日期 2026-05-08）

## 8. 已确认

- `config.yaml` 是 v1 主格式，`config.json` 已 deprecated 但仍兼容
- Rules 文件扩展名为 `.md`，YAML frontmatter；旧 YAML 直写格式仍支持
  但已 deprecated
- Prompts 文件扩展名为 `.prompt.md`，`invokable: true` 才能作为 slash
- `mcpServers` 仅在 agent 模式生效，支持 `stdio` / `sse` / `streamable-http`
- Hub blocks 通过 `uses: owner/package` 引用，对 rules / prompts /
  models / mcpServers / docs / data 均适用
- `.continuerc.json` 通过 `mergeBehavior: merge | overwrite` 控制
  工作区覆写策略

## 9. UNKNOWN

- `assistants/` 目录是否仍是 v1 推荐结构（文档在重组中，部分页面已
  转向以 Hub assistant 为主）。v1.0 转换器不写 `assistants/`，仅写
  `rules/` 与 `prompts/`，避免踩坑
- `config.ts` 与 `config.yaml` 共存时的合并优先级未在公开文档明确
  描述
- Hub `uses:` 是否支持版本固定语法（如 `owner/package@1.0`）需进一步
  核实，标 UNKNOWN，转换器输出统一不带版本
