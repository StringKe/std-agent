# Target: Muse Code (Meta)

调研日期: 2026-09-17
官方文档: https://dev.meta.ai/docs/muse-code
官方安装: `curl -fsSL https://dev.meta.ai/install.sh | bash`（二进制名 `muse`，模型 Muse Spark）
维护方: Meta MSL

## 1. 摘要

Muse Code 是 Meta 的终端 coding agent（交互式 TUI + `muse exec` headless 模式，
CI 可用），默认模型 `muse-spark-1.2`。配置体系绕 `AGENTS.md` 标准展开：
`muse init` 只写项目根单个 `AGENTS.md`，不建其他文件；项目 skills 走共享
`.agents/skills/<skill-id>/SKILL.md`（与 Codex / Goose / Zed 同命名空间，
spec-kit 的 Muse 集成 likewise 把 skills 装进 `.agents/skills` 并以
`/<command>` 调用）；项目记忆走 `.agents/memory/`（`MEMORY.md` 索引 + 每主题
一 markdown）；项目 hooks 走 `.muse/hooks.json`；MCP 与模型偏好走用户级
`~/.config/muse/settings.json`（`mcp_servers` 块）。

关键信任语义：项目级规则 / skills / hooks 须 trust workspace 后才加载；
已提交的项目记忆即使在 untrusted checkout 也会被读入 context（prompt
注入面，review 非受控 checkout 的记忆文件）。

## 2. 配置文件路径

| 类别 | 路径 | 说明 |
|---|---|---|
| 用户设置 | `~/.config/muse/settings.json`（须含 `"schema_version": 1`） | 模型默认、权限 profile、MCP（`mcp_servers`）、hooks、telemetry |
| 项目指令 | `<repo>/AGENTS.md`（`muse init` 生成，`--dry-run` 预览，`--force` 覆盖） | 项目主入口；另认 `CLAUDE.md`、`.agents/AGENTS.md`、`.claude/CLAUDE.md` |
| 嵌套指令 | `<repo>/<subdir>/AGENTS.md` | 从 workspace root 向上 walk 到最近 `.git` 边界，每层按 AGENTS.md → CLAUDE.md → .agents/AGENTS.md → .claude/CLAUDE.md 取首个存在文件，深层覆盖浅层 |
| 项目 Skills | `<repo>/.agents/skills/<skill-id>/SKILL.md`，另扫描仓库内 `.codex/skills`、`.claude/skills` | slash 调用 + observer 按需加载；`muse skills list/inspect/enable/install/validate` 管理，`muse skills import --from claude\|codex` 迁入 |
| 用户 Skills | `$XDG_CONFIG_HOME/muse/skills`、`~/.agents/skills`（另发现 `~/.claude/skills`、`$CODEX_HOME/skills` 回退 `~/.codex/skills`） | 账号级全项目可用 |
| 项目记忆 | `<repo>/.agents/memory/`（`MEMORY.md` 索引 + 每主题一 `.md`） | session 初注入索引（`MEMORY.md` + 至多 48 个文件路径），正文按需读 |
| 项目 Hooks | `<repo>/.muse/hooks.json` | 绑定 shell 命令到生命周期事件（`SessionStart`、`PreToolUse`、`PostToolUse`、`SubagentStart/Stop` 等），trust 后运行 |
| Subagents | 无文件格式（lead session 经工具派生，可选 worktree 隔离） | `/subagents`、`/tasks` 管理；单树默认并发 8（含 root），`agents.execution_capacity` 可调 1–64 |

冲突优先级：项目规则 > 用户规则；项目文件之间深层 > 浅层。

## 3. 文件格式

| 文件 | 格式 | frontmatter |
|---|---|---|
| `AGENTS.md` | Markdown | 无（纯指令文本） |
| `.agents/skills/<id>/SKILL.md` | Markdown | Agent Skills 标准（`name` / `description`，`name` 须等于目录名） |
| `.agents/memory/*.md` | Markdown | 无（`MEMORY.md` 为索引，每主题一行） |
| `.muse/hooks.json` | JSON | 不适用（事件 → shell 命令绑定） |
| `settings.json` | JSON | 不适用（须 `"schema_version": 1`，`mcp_servers` 下 `stdio` / `streamable_http`） |

官方未文档化任何字节硬限（root-file / rule / skill 均无）；沿用通用软建议。

## 4. std-agent 五类映射

| std-agent 类型 | Muse 落点 | 加载方式 |
|---|---|---|
| rules | 根 `AGENTS.md`（所有 nonRoot rules inline 拼接；嵌套 root 写 `<path>/AGENTS.md`） | 启动 walk 叠加（trust 后） |
| skills | `.agents/skills/<name>/SKILL.md` + 同目录辅助文件（字段集与 Codex / Goose / Zed 同集，共享路径字节一致） | slash 调用 + observer 按需加载 |
| commands | `.agents/skills/commands/<name>/SKILL.md`（无原生 commands 目录，降级为 skill 子目录；slash 即可调用） | 同 skills |
| references | `.agents/memory/<name>.md`（官方项目记忆主题文件位） | 记忆索引列出后按需读 |
| subagents | `.muse/subagents/<name>.md`（无原生 markdown 格式，fallback 供人阅读；不被 Muse 加载） | 不加载（人工查阅） |

## 5. 转换器实现要点（对照 `internal/transformer/muse.go`）

1. 主输出 `AGENTS.md`：root rules + inline nonRoot（`RulesDir=""`，`muse init` 单文件语义）
2. `NestedSupported: true`：带 `NestedPath` 的 root rule 产出 `<NestedPath>/AGENTS.md`，对齐按层 walk 语义
3. skills 走共享 `.agents/skills/`，`SkillSupportedFields` 与 codex / goose / zed 同集
4. commands 以 `CommandSkillPrefix` 降级到 `.agents/skills/commands/`（goose 同风格）
5. references 原生写 `.agents/memory/<name>.md`；`MEMORY.md` 索引不自动生成，需用户手工加一行（见 UNKNOWN）
6. subagents 无 `SubagentsDir`，走 `FallbackDir=".muse"` 子目录隔离（`.muse/` 为官方项目 hooks 目录）
7. 不写 `settings.json` / `.mcp.json`：MCP（`mcp_servers`）与模型偏好是用户级配置，不在项目同步范围
8. 备份：sync 前对 `AGENTS.md` `.agents/skills/` `.agents/memory/` `.muse/` 做快照（以 runner 实际快照为准）

## 6. 信息来源

- https://dev.meta.ai/docs/muse-code （总览：TUI + `muse exec`，Muse Spark）
- https://dev.meta.ai/docs/muse-code/configuration.md （settings 路径与 schema_version、AGENTS.md 四路径加载顺序与 trust 语义、`.agents/memory/` 布局与 48 文件索引、untrusted 记忆加载警告）
- https://dev.meta.ai/docs/muse-code/extending.md （skills 四来源与管理命令、内建 `/plan` `/grill` `/taste` `/threejs` 与 `/migrate`、`.muse/hooks.json` 与事件清单、`mcp_servers` 声明、subagent worktree 隔离与 8 并发）
- https://github.com/github/spec-kit/commit/4dd8afa （Muse 集成：skills 装进 `.agents/skills` 并以 `/speckit-<command>` 调用）
- https://github.com/ulises-jeremias/agentic-harness/blob/HEAD/docs/quickstarts/MUSE_CODE.md （`~/.config/muse/skills/` + `.agents/skills`，`muse skills import --from claude`）

## 7. UNKNOWN

- `.agents/skills/commands/` 子目录 skill 是否被 Muse 递归发现（Zed 明确扁平才发现；Muse 官方只写 `<skill-id>/SKILL.md` 一层，未说明子目录；若实测不发现，commands 应改走 `.muse/commands/` fallback）
- `MEMORY.md` 索引缺失时新主题文件是否仍被索引列出（官方写"MEMORY.md + 其他文件路径列表"，推测列出，但未明确无索引行时的行为；sync 不写索引是保守选择）
- `.muse/` 下除 `hooks.json` 外的 fallback 文件是否被读取（官方只文档化 hooks.json；fallback 产物按人读文档处理）
