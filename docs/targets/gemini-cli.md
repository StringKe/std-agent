# Target: Gemini CLI (Google)

调研日期: 2026-05-07
官方文档:
- https://github.com/google-gemini/gemini-cli/blob/main/docs/
- https://geminicli.com/docs/

## 1. 摘要

Gemini CLI 是 Google 开源的终端 AI agent。配置围绕 `~/.gemini/` 与项目根
`.gemini/` 双层展开，上下文文件用 `GEMINI.md` 三级层级加载（global -> workspace
祖先链 -> JIT 子目录）。

通过 `settings.json` 的 `context.fileName` 可显式声明同时读 `AGENTS.md`，
是在 Gemini 侧实现 "读 AGENTS.md" 的官方做法。

## 2. 配置文件路径

| 类别 | 路径 | 说明 |
|---|---|---|
| 全局上下文 | `~/.gemini/GEMINI.md` | 用户级 |
| 项目上下文（祖先链） | 从 cwd 起向上每级目录扫描 `GEMINI.md` | 到 `.git` 边界停止 |
| 子目录 JIT 上下文 | 工具访问文件/目录时按需扫描该目录及其祖先内的 `GEMINI.md` | 懒加载 |
| 用户 settings | `~/.gemini/settings.json` | |
| 项目 settings | `<project>/.gemini/settings.json` | |
| 全局自定义命令 | `~/.gemini/commands/*.toml` | 嵌套作命名空间 |
| 项目自定义命令 | `<project>/.gemini/commands/*.toml` | 项目同名覆盖全局 |
| 扩展（全局） | `~/.gemini/extensions/<name>/` | `gemini extensions install` |

## 3. 文件格式与 frontmatter

| 文件 | 扩展名 | 字段/键 |
|---|---|---|
| `GEMINI.md` | `.md` | 无 frontmatter；支持 `@./path.md` 与 `@/abs/path.md` 文件导入 |
| 自定义命令 | `.toml` | `prompt`（必填，单/多行）；`description`（可选，单行） |
| `settings.json` | `.json` | `context.fileName`（字符串或数组）、`mcpServers`、其它 UNKNOWN |
| 扩展清单 | `gemini-extension.json` | 完整 schema 见 # 6 |

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
  "context": { "fileName": ["AGENTS.md", "GEMINI.md"] },
  "mcpServers": { "<name>": { ... } }
}
```

设置 `context.fileName` 可让 Gemini CLI 同时读 `AGENTS.md`。

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

- `GEMINI.md` 三级合并：global -> workspace 祖先链 -> JIT 子目录
- footer 显示已加载文件计数；`/memory show` 查看合并后内容
- `/memory reload` 是 CLI 主线规范名；ACP 子系统中规范名为 `/memory refresh`，
  CLI 端 `memory reload` 与 ACP 端 `memory refresh` 互为别名
- 命令：项目同名覆盖全局；`/commands list`、`/commands reload` 管理
- 扩展通过 `gemini extensions install <git-url>` 安装到全局

## 8. std-ai 四类映射

| std-ai 类型 | Gemini CLI 落点 |
|---|---|
| rules | 项目根 `GEMINI.md` 主文件 + 子目录 `GEMINI.md`（按 std `applyTo` 路径决定写到哪个目录） |
| skills | 通过命令 `prompt` 字段封装；复杂封装走 extensions（v1.0 不写） |
| commands | `.gemini/commands/<name>.toml`，子目录命名空间映射 std rule 的标签 |
| references | `GEMINI.md` 内用 `@file.md` 模块化导入 |

## 9. 转换器实现要点

1. 主输出：项目根 `GEMINI.md`（拼接无 `applyTo` 限定的 rules）
2. 有 `applyTo` 的 rule 写入对应目录的 `GEMINI.md`（取 glob 公共前缀目录；
   多 glob 取最大公共前缀）
3. commands 输出 `.gemini/commands/<name>.toml`：
   - std `description` -> `description`
   - std 正文 + footer -> `prompt`
4. v1.0 不写 `settings.json`；提示用户手动加 `context.fileName: ["AGENTS.md", "GEMINI.md"]`
   即可同步读 AGENTS.md
5. extensions 不在 v1.0 范围

## 10. 信息来源

- https://github.com/google-gemini/gemini-cli/blob/main/docs/cli/gemini-md.md
- https://github.com/google-gemini/gemini-cli/blob/main/docs/cli/custom-commands.md
- https://github.com/google-gemini/gemini-cli/blob/main/docs/cli/cli-reference.md
- https://github.com/google-gemini/gemini-cli/blob/main/docs/extensions/index.md
- https://github.com/google-gemini/gemini-cli/blob/main/docs/extensions/reference.md
- https://github.com/google-gemini/gemini-cli/blob/main/docs/extensions/writing-extensions.md
- https://github.com/google-gemini/gemini-cli/blob/main/docs/reference/configuration.md
- https://github.com/google-gemini/gemini-cli/blob/main/packages/cli/src/acp/commands/memory.ts
- https://geminicli.com/docs/cli/gemini-md/

## 11. 已确认与剩余 UNKNOWN

已确认：
- `gemini-extension.json` 完整 schema（见 # 6）
- mcpServers 与 settings.json 同 schema（除 `trust`）
- commands 不在 extension manifest 内嵌，由扩展目录 `commands/*.toml` 提供
- `/memory reload` 是 CLI 主线规范名；`/memory refresh` 仅在 ACP 子系统中是规范名

剩余 UNKNOWN：
- hooks 完整 schema（仓库 docs 列出 hooks/ 目录但完整 schema 未公开）
- `settings.json` 全部字段
- trusted root 的精确边界规则
