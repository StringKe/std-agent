# Target: Gemini CLI (Google)

调研日期: 2026-05-07，2026-07-10 复核更新，2026-08-15 复核更新
官方文档:
- https://github.com/google-gemini/gemini-cli/blob/main/docs/
- https://geminicli.com/docs/

## 1. 摘要

Gemini CLI 是 Google 开源的终端 AI agent。配置围绕 `~/.gemini/` 与项目根
`.gemini/` 双层展开，上下文文件用 `GEMINI.md` 三级层级加载（global -> workspace
祖先链 -> JIT 子目录）。

2026-08-15 复核：`GEMINI.md`、`.gemini/skills/`、`.gemini/commands/*.toml` 路径未变。
官方文档提示 unpaid / Google One 用户的 Gemini CLI 将过渡到 Antigravity CLI；
stdagent 仍同时支持 `gemini` 与 `antigravity` 两个 target。

通过 `settings.json` 的 `context.fileName` 可显式声明同时读 `AGENTS.md`，
是在 Gemini 侧实现 "读 AGENTS.md" 的官方做法；社区 issue #28227 反映该
merge 行为疑似被 revert，截至复核仍未关闭，标 UNKNOWN（见 # 12）。

2026-07 复核新增：Skills 自 v0.38+ **GA 默认启用**，原生落点
`.gemini/skills/<name>/SKILL.md`（旧文档"v1.0 不写，走命令 prompt 封装"已过时，
P0 修复，来源 https://github.com/google-gemini/gemini-cli/blob/main/docs/cli/skills.md）。

## 2. 配置文件路径

| 类别 | 路径 | 说明 |
|---|---|---|
| 全局上下文 | `~/.gemini/GEMINI.md` | 用户级 |
| 项目上下文（祖先链） | 从 cwd 起向上每级目录扫描 `GEMINI.md` | 到 `.git` 边界停止 |
| 子目录 JIT 上下文 | 工具访问文件/目录时按需扫描该目录及其祖先内的 `GEMINI.md` | 懒加载；嵌套文件名须是 `GEMINI.md`，不是 `AGENTS.md` |
| 目录发现上限 | `context.discoveryMaxDirs` 默认 200 | 超过此目录数不再继续扫描（2026-07 新增确认） |
| 项目 skills | `.gemini/skills/<name>/SKILL.md` | 原生 Agent Skills 标准，GA 默认启用 |
| 用户 settings | `~/.gemini/settings.json` | |
| 项目 settings | `<project>/.gemini/settings.json` | |
| 全局自定义命令 | `~/.gemini/commands/*.toml` | 嵌套作命名空间 |
| 项目自定义命令 | `<project>/.gemini/commands/*.toml` | 项目同名覆盖全局 |
| 扩展（全局） | `~/.gemini/extensions/<name>/` | `gemini extensions install` |

## 3. 文件格式与 frontmatter

| 文件 | 扩展名 | 字段/键 |
|---|---|---|
| `GEMINI.md` | `.md` | 无 frontmatter；支持 `@./path.md` 与 `@/abs/path.md` 文件导入，深度 **5 层** |
| `SKILL.md` | `.md` | Agent Skills 标准字段（`name` / `description` / `license` / `compatibility` / `metadata`） |
| 自定义命令 | `.toml` | `prompt`（必填，单/多行）；`description`（可选，单行） |
| `settings.json` | `.json` | `context.fileName`（字符串或数组）、`context.discoveryMaxDirs`、`mcpServers`、其它 UNKNOWN |
| 扩展清单 | `gemini-extension.json` | 完整 schema 见 # 6 |

字节上限：官方无字节上限文档（根文件、skills、commands 均无数值化上限，见 # 12 UNKNOWN）。

## 4. 命令 toml 模板

```toml
description = "..."   # 可选
prompt = """
请帮我...

参数: {{args}}

可注入文件: @{path/to/file.md}

执行 shell: !{git status}
"""
```

注入语法：

| 语法 | 含义 |
|---|---|
| `{{args}}` | shell 块外原样注入；`!{...}` 内自动 shell-escape |
| `!{cmd}` | 执行前提示用户确认 |
| `@{path}` | 文件/目录注入；遵循 `.gitignore`；支持图片/PDF/音视频多模态 |

子目录映射为命名空间（`git/commit.toml` -> `/git:commit`）。

## 5. settings.json 关键字段

```jsonc
{
  "context": { "fileName": ["AGENTS.md", "GEMINI.md"], "discoveryMaxDirs": 200 },
  "mcpServers": { "<name>": { ... } }
}
```

设置 `context.fileName` 可让 Gemini CLI 同时读 `AGENTS.md`（该 merge 行为社区
issue #28227 反映疑似被 revert，未关闭，标 UNKNOWN）。

## 6. gemini-extension.json schema

```jsonc
{
  "name": "...",                   // 必填
  "version": "...",                // 必填
  "description": "...",            // 可选
  "mcpServers": { "<name>": { ... } },  // 与 settings.json mcpServers 同 schema（仅 trust 字段除外）
  "contextFileName": "GEMINI.md",  // 可选；默认 GEMINI.md
  "excludeTools": ["run_shell_command(rm -rf)"],  // 命令级粒度
  "migratedTo": "...",             // 可选；扩展迁移到新仓库时使用
  "plan": { "directory": "..." },  // 可选
  "settings": { ... }              // 可选；扩展可声明配置项
}
```

注意：

- commands 不在 manifest 内嵌，由扩展目录下 `commands/` 子目录中的 toml 提供
- 扩展中可使用 `${extensionPath}` 引用扩展目录内的资源
- mcpServers 与 settings.json mcpServers 完全同 schema（仅 `trust` 字段不允许）

## 7. 加载机制

- `GEMINI.md` 三级合并：global -> workspace 祖先链 -> JIT 子目录；嵌套文件名固定 `GEMINI.md`
- 完整支持嵌套：逐层 + JIT（按需触发子目录扫描），`context.discoveryMaxDirs` 默认 200 目录为扫描上限
- footer 显示已加载文件计数；`/memory show` 查看合并后内容
- `/memory reload` 是 CLI 主线规范名；ACP 子系统中规范名为 `/memory refresh`，
  CLI 端 `memory reload` 与 ACP 端 `memory refresh` 互为别名
- 命令：项目同名覆盖全局；`/commands list`、`/commands reload` 管理
- 扩展通过 `gemini extensions install <git-url>` 安装到全局

## 8. std-agent 五类映射（实际实现，`internal/transformer/gemini.go`）

| std-agent 类型 | Gemini CLI 落点 |
|---|---|
| rules（root） | 项目根 `GEMINI.md`（type glossary + root body） |
| rules（nonRoot） | RulesDir 留空，全部 inline 拼接到 `GEMINI.md` 正文 |
| rules（nested） | `<NestedPath>/GEMINI.md`（无 manifest，与 root 同文件名，符合官方嵌套文件名须为 `GEMINI.md` 的要求） |
| skills | `.gemini/skills/<name>/SKILL.md`（原生 Agent Skills 标准包） |
| commands | `.gemini/commands/<name>.toml`（独立 TOML 渲染，不走 AgentsMD 协议） |
| references / subagents | `.gemini/rules/{references,subagents}/<name>.md`（`FallbackDir=".gemini/rules"` 降级） |

## 9. 转换器实现要点（对照 `internal/transformer/gemini.go`）

1. 主输出：项目根 `GEMINI.md`（`AgentsMD` 协议渲染 rules / skills / references / subagents）
2. commands 在 `Plan` 内提前从 `docs` 中剥离，单独走 `buildGeminiCommandTOML` 渲染
   `.gemini/commands/<name>.toml`：
   - std `description` -> `description`
   - std 正文 -> `prompt`（TOML 三引号多行 literal string）
3. skills 原生：`.gemini/skills/<name>/SKILL.md` + 辅助文件
4. references / subagents 无原生目录，走 `FallbackDir=".gemini/rules"` 降级
5. v1.0 不写 `settings.json`；不提示用户手动加 `context.fileName`（该 merge 行为社区反馈不稳定，见 # 12）
6. extensions 不在 v1.0 范围

## 10. 信息来源

- https://github.com/google-gemini/gemini-cli/blob/main/docs/cli/gemini-md.md
- https://github.com/google-gemini/gemini-cli/blob/main/docs/cli/skills.md
- https://github.com/google-gemini/gemini-cli/blob/main/docs/cli/custom-commands.md
- https://github.com/google-gemini/gemini-cli/blob/main/docs/cli/cli-reference.md
- https://github.com/google-gemini/gemini-cli/blob/main/docs/extensions/index.md
- https://github.com/google-gemini/gemini-cli/blob/main/docs/extensions/reference.md
- https://github.com/google-gemini/gemini-cli/blob/main/docs/extensions/writing-extensions.md
- https://github.com/google-gemini/gemini-cli/blob/main/docs/reference/configuration.md
- https://github.com/google-gemini/gemini-cli/blob/main/packages/cli/src/acp/commands/memory.ts
- https://geminicli.com/docs/cli/gemini-md/

## 11. budget.go 限额（2026-07 回填）

| kind | Soft | Hard | 语义 |
|---|---|---|---|
| root-file（GEMINI.md） | 8000 字符 | **0（无 Hard）** | 官方无字节上限文档；旧代码曾写 Hard 32000 无依据，已归 0，只留软指导 |

## 12. 已确认与剩余 UNKNOWN（2026-07-10 复核）

已确认：
- `gemini-extension.json` 完整 schema（见 # 6）
- mcpServers 与 settings.json 同 schema（除 `trust`）
- commands 不在 extension manifest 内嵌，由扩展目录 `commands/*.toml` 提供
- `/memory reload` 是 CLI 主线规范名；`/memory refresh` 仅在 ACP 子系统中是规范名
- skills 已 GA 默认启用，原生落点 `.gemini/skills/`（P0 修复）
- `@import` 深度 5 层；`context.discoveryMaxDirs` 默认 200 目录

剩余 UNKNOWN（2026-07 复核仍未证实）：
- gemini-cli AGENTS.md 默认加载 merge 后疑似 revert（issue #28227 未关闭）
- 根文件（GEMINI.md）与 skills / commands 的字节上限及截断行为（官方均无数值文档）
- hooks 完整 schema（仓库 docs 列出 hooks/ 目录但完整 schema 未公开）
- `settings.json` 全部字段
- trusted root 的精确边界规则
