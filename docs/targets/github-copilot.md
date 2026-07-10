# Target: GitHub Copilot

调研日期: 2026-05-07
官方文档:
- https://docs.github.com/en/copilot/customizing-copilot/
- https://code.visualstudio.com/docs/copilot/customization/

## 1. 摘要

GitHub Copilot 是多客户端 + 多服务端（Chat、Coding Agent、Code Review、CLI、
Code Completion）产品族。自定义在仓库、用户、组织三级叠加。仓库级配置位于
`.github/` 子树，VS Code 端有额外的 prompts / agents / chat modes。

Coding Agent 兼容 `AGENTS.md` / `CLAUDE.md` / `GEMINI.md`（最近邻 wins 语义）；
其他客户端不消费这些文件。

## 2. 配置文件路径

| 类别 | 路径 | 说明 |
|---|---|---|
| 仓库通用指令 | `.github/copilot-instructions.md` | 单文件，仓库根 |
| 仓库路径限定指令 | `.github/instructions/<name>.instructions.md` | 多文件，可嵌套子目录 |
| Coding Agent 兼容 | 仓库内任意 `AGENTS.md`（最近邻 wins）；仓库根的 `CLAUDE.md` / `GEMINI.md`（单文件） | 仅 Coding Agent 消费 |
| 仓库 prompts | `.github/prompts/<name>.prompt.md` | VS Code 自定义 prompt |
| 仓库 agents | `.github/agents/<name>.agent.md`（旧 `.chatmode.md`） | VS Code custom agents（前称 chat modes） |
| 用户 prompts/agents | profile 目录（如 macOS `~/.copilot/agents`） | 由 `chat.promptFilesLocations` `chat.agentFilesLocations` 控制 |
| MCP（VS Code 工作区） | `.vscode/mcp.json` | 顶层键 `servers` |
| MCP（Coding Agent） | GitHub.com Settings -> Copilot -> Cloud agent | 顶层键 `mcpServers` |

## 3. 文件格式与 frontmatter

| 文件 | 扩展名 | frontmatter 字段 |
|---|---|---|
| `copilot-instructions.md` | `.md` | 无 |
| `*.instructions.md` | `.instructions.md` | `applyTo`（必填，glob，逗号分隔多模式）；`excludeAgent`（可选，值 `code-review` 或 `cloud-agent`） |
| `AGENTS.md` / `CLAUDE.md` / `GEMINI.md` | `.md` | 无 |
| `*.prompt.md` | `.prompt.md` | `description`、`name`、`argument-hint`、`agent`（`ask`/`agent`/`plan`/自定义）、`model`、`tools` |
| `*.agent.md` | `.agent.md` | `description`、`tools`、`model`、`name`、`handoffs` |

## 4. 客户端支持矩阵

| 文件 | Chat | VS Code | Coding Agent | Code Review |
|---|---|---|---|---|
| `copilot-instructions.md` | 是 | 是 | 是 | 是 |
| `*.instructions.md` | 部分 | 是 | 是 | 是 |
| `AGENTS.md` / `CLAUDE.md` / `GEMINI.md` | 否 | 否 | 是 | 否 |
| `*.prompt.md` | 否 | 是 | 否 | 否 |
| `*.agent.md` | 否 | 是 | 否 | 否 |

## 5. 加载机制与优先级

优先级（高 -> 低）：Personal -> Path-specific repo -> Repository-wide -> Agent instructions -> Organization。

- `.github/copilot-instructions.md`：自动注入到所有 Chat 请求
- `.instructions.md`：按 `applyTo` glob 匹配自动启用；多文件合并 "no specific order guaranteed"
- `AGENTS.md`：最近邻 wins（仓库目录树就近匹配）
- Code Review 仅读前 4000 字符
- PR 上下文读 base branch 的 instructions，不读 feature branch

## 6. VS Code settings 关键字段

```jsonc
{
  "github.copilot.chat.codeGeneration.useInstructionFiles": true,
  "github.copilot.chat.commitMessageGeneration.instructions": [...],
  "github.copilot.chat.reviewSelection.instructions": [...],
  "github.copilot.chat.pullRequestDescriptionGeneration.instructions": [...],
  "chat.instructionsFilesLocations": [...],
  "chat.promptFilesLocations": [...],
  "chat.agentFilesLocations": [...]
}
```

注：`github.copilot.chat.codeGeneration.instructions` 与 `github.copilot.chat.testGeneration.instructions`
**自 VS Code 1.102 起 deprecated**，迁移到 file-based instructions
（`.github/copilot-instructions.md` 或 `*.instructions.md`）。
剩余三个 `commitMessageGeneration` / `reviewSelection` / `pullRequestDescriptionGeneration`
字段仍生效。

## 7. std-agent 四类映射

| std-agent 类型 | Copilot 落点 |
|---|---|
| rules（无 applyTo） | `.github/copilot-instructions.md`（拼接所有无路径限定的 rules） |
| rules（有 applyTo） | `.github/instructions/<name>.instructions.md`（每条 rule 一个文件，frontmatter `applyTo` 来自 std `applyTo`） |
| skills | `.github/agents/<name>.agent.md`（仓库共享）；frontmatter 含 `description` `tools` `model` |
| commands | `.github/prompts/<name>.prompt.md`（frontmatter `description` `agent` `model` `tools`） |
| references | 不主动写入；建议拼入 `copilot-instructions.md` 或独立 docs/ |

## 8. 转换器实现要点

1. rules 拆分：std rule 有 `applyTo` 走 `.github/instructions/<name>.instructions.md`；无 `applyTo` 拼到 `.github/copilot-instructions.md`
2. `.instructions.md` frontmatter：
   - `applyTo: ["**/*.ts", "**/*.tsx"]` -> `applyTo: '**/*.ts,**/*.tsx'`（逗号分隔字符串）
   - 由 std `exclude_targets` 中含 `coding-agent` 推断出 `excludeAgent: cloud-agent`
3. `.prompt.md` frontmatter：`argument_hint` -> `argument-hint`，`allowed_tools` -> `tools`
4. `.agent.md`：v1.0 不主动生成（无对应 std type）；保留扩展位
5. `AGENTS.md` 已由 codex transformer 写入；Copilot 复用
6. v1.0 不写 `.vscode/mcp.json`

## 9. 信息来源

- https://docs.github.com/en/copilot/customizing-copilot/adding-repository-custom-instructions-for-github-copilot
- https://docs.github.com/en/copilot/customizing-copilot/about-customizing-github-copilot-chat-responses
- https://docs.github.com/en/copilot/how-tos/configure-custom-instructions/add-repository-instructions
- https://docs.github.com/copilot/how-tos/agents/copilot-coding-agent/extending-copilot-coding-agent-with-mcp
- https://code.visualstudio.com/docs/copilot/customization/custom-instructions
- https://code.visualstudio.com/docs/copilot/customization/prompt-files
- https://code.visualstudio.com/docs/copilot/customization/custom-chat-modes
- https://code.visualstudio.com/docs/copilot/customization/mcp-servers

## 10. 已确认与剩余 UNKNOWN

已确认：
- `codeGeneration.instructions` / `testGeneration.instructions` settings 自 VS Code 1.102 起
  deprecated，改用 file-based instructions
- VS Code chat modes 已改名 custom agents，文件后缀 `.chatmode.md` -> `.agent.md`，
  settings 改名 `chat.agentFilesLocations`
- Coding Agent MCP 顶层键 `mcpServers`，VS Code 工作区 `.vscode/mcp.json` 顶层键
  `servers`，schema 不一致，stdagent 写两边时需双向转换

剩余 UNKNOWN：
- `.agent.md` 的 `handoffs` schema
- Coding Agent MCP `mcpServers` schema 全部字段（`type` 已知 `local`/`stdio`/`http`/`sse`）
