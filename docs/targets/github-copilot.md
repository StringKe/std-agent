# Target: GitHub Copilot

调研日期: 2026-05-07，2026-07-10 复核更新
官方文档:
- https://docs.github.com/en/copilot/customizing-copilot/
- https://code.visualstudio.com/docs/copilot/customization/
- https://docs.github.com/en/copilot/concepts/agents/about-agent-skills

## 1. 摘要

GitHub Copilot 是多客户端 + 多服务端（Chat、Coding Agent、Code Review、CLI、
Code Completion）产品族。自定义在仓库、用户、组织三级叠加。仓库级配置位于
`.github/` 子树，VS Code 端有额外的 prompts / agents / chat modes。

Coding Agent 兼容 `AGENTS.md` / `CLAUDE.md` / `GEMINI.md`（最近邻 wins 语义）；
2026-07 复核新增：VS Code Copilot Chat 也已消费根 `AGENTS.md`
（`chat.useAgentsMdFile` 默认 true，仍标 experimental，是否转正 UNKNOWN 见 # 10）。

Agent Skills 已 GA：cloud agent / Code Review / CLI / VS Code 客户端均支持
`.github/skills/<name>/SKILL.md`，旧文档把 skills 标"不支持"已过时（P0 修复，见 # 10）。

## 2. 配置文件路径

| 类别 | 路径 | 说明 |
|---|---|---|
| 仓库通用指令 | `.github/copilot-instructions.md` | 单文件，仓库根 |
| 仓库路径限定指令 | `.github/instructions/<name>.instructions.md` | 多文件，可嵌套子目录 |
| 仓库 skills | `.github/skills/<name>/SKILL.md` | Agent Skills 标准，cloud agent / Code Review / CLI / VS Code 均 GA |
| Coding Agent 兼容 | 仓库内任意 `AGENTS.md`（最近邻 wins）；仓库根的 `CLAUDE.md` / `GEMINI.md`（单文件） | Coding Agent 消费；VS Code Chat 亦消费 `AGENTS.md`（experimental） |
| 仓库 prompts | `.github/prompts/<name>.prompt.md` | VS Code 自定义 prompt |
| 仓库 agents | `.github/agents/<name>.agent.md`（旧 `.chatmode.md`，**已确认废弃**） | 原生 subagents，VS Code custom agents（前称 chat modes） |
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
| `SKILL.md` | `.md` | `name`（必填，须等于目录名，≤64）、`description`（必填，≤1024）、`license`（可选）、`compatibility`（可选，≤500）、`metadata`（可选）、`allowed-tools`（可选，experimental，空格分隔） |
| `*.agent.md` | `.agent.md` | `description`、`tools`、`model`、`disable-model-invocation`（取代已废弃的 `infer`）、`user-invocable`（默认 true）；官方另有 `target` / `agents` / `mcp-servers` / `handoffs` 字段，transformer 暂未渲染（见 # 10 P1） |

## 4. 客户端支持矩阵

| 文件 | Chat | VS Code | Coding Agent | Code Review |
|---|---|---|---|---|
| `copilot-instructions.md` | 是 | 是 | 是 | 是 |
| `*.instructions.md` | 部分 | 是 | 是 | 是 |
| `AGENTS.md` / `CLAUDE.md` / `GEMINI.md` | 是（experimental，`chat.useAgentsMdFile`） | 否 | 是 | 否 |
| `*.prompt.md` | 否 | 是 | 否 | 否 |
| `*.agent.md` | 否 | 是 | 否 | 否 |
| `SKILL.md` | 是 | 是 | 是 | 是（Agent Skills 全客户端 GA） |

## 5. 加载机制与优先级

优先级（高 -> 低）：Personal -> Path-specific repo -> Repository-wide -> Agent instructions -> Organization。

- `.github/copilot-instructions.md`：自动注入到所有 Chat 请求
- `.instructions.md`：按 `applyTo` glob 匹配自动启用；多文件合并 "no specific order guaranteed"
- `AGENTS.md`：最近邻 wins（仓库目录树就近匹配）
- **Code Review 截前 4000 字符的规则已于 2026-06-12 官方移除**（旧文档"仅读前 4000 字符"过时，来源 https://github.blog/changelog/2026-06-12-copilot-code-review-new-configurations-and-controls/）
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
  "chat.agentFilesLocations": [...],
  "chat.useAgentsMdFile": true,
  "chat.useNestedAgentsMdFiles": false
}
```

注：`github.copilot.chat.codeGeneration.instructions` 与 `github.copilot.chat.testGeneration.instructions`
**自 VS Code 1.102 起 deprecated**，迁移到 file-based instructions
（`.github/copilot-instructions.md` 或 `*.instructions.md`）。
剩余三个 `commitMessageGeneration` / `reviewSelection` / `pullRequestDescriptionGeneration`
字段仍生效。`chat.useNestedAgentsMdFiles` **默认 false**：VS Code 端默认不扫描嵌套子目录
`AGENTS.md`（与 Coding Agent 的最近邻语义不同，Coding Agent 侧无此开关限制）。

## 7. std-agent 五类映射（实际实现，`internal/transformer/copilot.go` / `protocol/copilot.go`）

| std-agent 类型 | Copilot 落点 |
|---|---|
| rules（root 或无 applyTo） | `.github/copilot-instructions.md`（拼接所有无路径限定的 rules） |
| rules（有 applyTo） | `.github/instructions/<name>.instructions.md`（每条 rule 一个文件，`applyTo` 逗号分隔字符串） |
| skills | `.github/skills/<name>/SKILL.md`（原生 Agent Skills 标准包，frontmatter 白名单见 # 3） |
| commands | `.github/prompts/<name>.prompt.md`（frontmatter `description` `argument-hint` `tools` `model`） |
| subagents | `.github/agents/<name>.agent.md`（原生落点）；**但 body 内容走 CLI 委派降级**（见下） |
| references | `.github/instructions/references/<n>.instructions.md`（子目录隔离降级，`applyTo: ""` 显式空字符串表示不限路径，`std-agent-type: references`） |

**subagents 的落点/内容分离**：`copilotAdapter.SubagentInvokeCmd = "claude --agent {name}"`
非空，`buildCopilotSubagent`（`protocol/copilot.go:214`）虽然写到官方原生路径
`.github/agents/<n>.agent.md`（frontmatter 用官方字段），但 **body 正文被替换为
"通过 shell 调用 `claude --agent <name>` 委派"的指引文本**，不是 subagent 原始 prompt
直出。这是因为 Copilot 侧没有等价的隔离子代理执行环境，stdagent 选择让 Copilot 的 AI
把工作转发给 Claude Code CLI 执行。`SkillFiles`（skill 包辅助文件）在 `.agent.md`
单文件场景下会被忽略并产生 WARN（`.instructions.md` 降级场景同理）。

## 8. 转换器实现要点

1. rules 拆分：std rule 有 `applyTo` 走 `.github/instructions/<name>.instructions.md`；无 `applyTo` 或 `Root=true` 拼到 `.github/copilot-instructions.md`
2. `.instructions.md` frontmatter：`applyTo` 用逗号分隔字符串（`GlobsCommaString`）
3. `.prompt.md` frontmatter：`argument-hint` 直传，`AllowedTools` -> `tools`
4. `.agent.md`：**已实现**（原生落点 + CLI 委派 body，见 # 7），旧文档"v1.0 不主动生成"过时
5. skills：**已实现原生落点** `.github/skills/<name>/SKILL.md`，旧文档"降级为 `.github/agents/<name>.agent.md`"过时（P0 修复）
6. `AGENTS.md`：copilot transformer 自身不写，复用 codex / claude-code 等 transformer 的产出
7. `.vscode/mcp.json` **已实现**（`buildCopilotMCP`，顶层键 `servers`），旧文档"v1.0 不写"过时

## 9. 信息来源

- https://docs.github.com/en/copilot/customizing-copilot/adding-repository-custom-instructions-for-github-copilot
- https://docs.github.com/en/copilot/customizing-copilot/about-customizing-github-copilot-chat-responses
- https://docs.github.com/en/copilot/how-tos/configure-custom-instructions/add-repository-instructions
- https://docs.github.com/copilot/how-tos/agents/copilot-coding-agent/extending-copilot-coding-agent-with-mcp
- https://docs.github.com/en/copilot/concepts/agents/about-agent-skills
- https://code.visualstudio.com/docs/copilot/customization/custom-instructions
- https://code.visualstudio.com/docs/copilot/customization/prompt-files
- https://code.visualstudio.com/docs/copilot/customization/custom-chat-modes
- https://code.visualstudio.com/docs/copilot/customization/mcp-servers
- https://github.blog/changelog/2026-06-12-copilot-code-review-new-configurations-and-controls/

## 10. budget.go 限额（2026-07 回填）

| kind | Soft | Hard | 语义 |
|---|---|---|---|
| root-file（copilot-instructions.md） | 8000 字符 | **0（无 Hard）** | 历史截断规则已于 2026-06-12 移除，官方软指导不超过约 2 页 |
| subagent（.agent.md 正文） | - | 30000 字符 | GitHub.com custom agents 参考文档硬上限 |
| skill-name（`*` 通用） | - | 64 字符 | Agent Skills 规范硬拒载 |
| skill-description（`*` 通用） | - | 1024 字符 | Agent Skills 规范硬拒载 |
| command（`*` 通用，旧结论已废） | 4000 字符 | 0 | 旧 Code Review 4000 字符规则已移除，现仅剩纯经验软建议 |

## 11. 已确认与剩余 UNKNOWN（2026-07-10 复核）

已确认：
- `codeGeneration.instructions` / `testGeneration.instructions` settings 自 VS Code 1.102 起
  deprecated，改用 file-based instructions
- VS Code chat modes 已改名 custom agents，文件后缀 `.chatmode.md` -> `.agent.md`，
  settings 改名 `chat.agentFilesLocations`；`.chatmode.md` 确认废弃
- Coding Agent MCP 顶层键 `mcpServers`，VS Code 工作区 `.vscode/mcp.json` 顶层键
  `servers`，schema 不一致，stdagent 写两边时需双向转换
- Agent Skills 已 GA（P0 修复）：cloud agent / Code Review / CLI / VS Code 全客户端支持
  `.github/skills/<name>/SKILL.md`
- Code Review 4000 字符截断规则已于 2026-06-12 官方移除
- `.agent.md` frontmatter 扩展字段 `disable-model-invocation` / `user-invocable` 已实现；
  `target` / `agents` / `mcp-servers` / `handoffs` 官方确认存在但 transformer 暂未渲染

剩余 UNKNOWN（2026-07 复核仍未证实）：
- `chat.useAgentsMdFile` 是否已从 experimental 转正
- `.instructions.md` 的 on-demand 加载模式是否扩散到 GitHub.com 侧（目前仅 VS Code 确认）
- `.agent.md` 的 `handoffs` / `agents` / `mcp-servers` 完整 schema
- Coding Agent MCP `mcpServers` schema 全部字段（`type` 已知 `local`/`stdio`/`http`/`sse`）
