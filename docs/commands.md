# CLI 命令规范

## 命令总览

```
stdagent <command> [flags]

  init           初始化 .stdai/ 与 config.toml
  pull           更新 .stdai/cache/ 中的 Git 源
  sync           核心同步：pull -> parse -> convert -> 向外扩散
  import         采用外部产物到 .stdai/standards/external/
  fix            重新 sync 修复 drift（sync 的语义别名）
  status         显示 targets 状态与 drift
  clean          清空根目录与平台目录的生成文件
  budget         检查 source 与 rendered target 的上下文体积
  upgrade        自我升级到最新版本
  version        版本与构建信息
  help           帮助
```

## 通用 flag

| flag | 类型 | 默认 | 说明 |
|---|---|---|---|
| `--config` | string | `.stdai/config.toml` | 配置文件路径；未指定时从 cwd 向上 walk |
| `--dry-run` | bool | false | 只输出将做什么，不写盘 |
| `--verbose` `-v` | count | 0 | -v 信息，-vv 调试 |
| `--no-color` | bool | auto | 禁用 ANSI 颜色 |
| `--quiet` `-q` | bool | false | 仅输出错误 |

## `stdagent init`

```
stdagent init [--force] [--minimal] [--source <git-url>]
```

行为：

1. 当前目录创建 `.stdai/`
2. 写 `config.toml`（默认模板）
3. 创建 `.stdai/standards/{rules,skills,commands,references}/` + `.gitkeep`
4. 写示例：`rules/example.md` + `skills/code-review/SKILL.md`
5. 按 `gitignore` 模式（默认 `generated`）写入根 `.gitignore` 的 `# BEGIN stdagent` 块：运行时文件 + 默认启用 target 的可重建产物

| flag | 说明 |
|---|---|
| `--force` | 已存在 `.stdai/` 时强制覆盖（先备份到 `.stdai-backup-<ts>/`） |
| `--minimal` | 不写示例文件，只建空目录 |
| `--source <url>` | 在 `[sources.default]` 写入指定 URL |

退出码：0 成功；1 已存在且无 `--force`；2 IO 失败。

## `stdagent pull`

```
stdagent pull [--source <name>] [--all]
```

对 enabled sources 执行 `git fetch + checkout` 到 `.stdai/cache/<name>/`。
不参与转换与写盘。

| flag | 说明 |
|---|---|
| `--source <name>` | 仅 pull 指定 source |
| `--all` | 即使 `enabled=false` 也 pull（用于预热） |

## `stdagent sync`

```
stdagent sync [--target <name>...] [--dry-run] [--no-pull] [--no-backup] [--no-external] [--strict]
```

行为：

1. 如 `auto_pull=true` 且未传 `--no-pull`，先 `pull`
2. 解析 `.stdai/standards/` + `cache/<source>/<paths>` 合并出最终 source set
3. 默认扫描外部产物（`.ai/guidelines/`、`.ai/rules/`、`.ai/skills/`、`.agents/skills/`、根 `.mcp.json`），按默认映射先落盘到 `.stdai/standards/external/`（幂等：内容一致跳过；`--no-external` 跳过；`[external] enabled=false` 同等关闭），后续本地收集自然纳入。`--dry-run` 时仅内存预览，不写盘
4. 加载 `.stdai/standards/mcp.json`（若存在），外部根 `.mcp.json` 只合并缺失键
5. 对每个 enabled target（或 `--target` 限定的子集）先完成 plan，再统一校验共享路径
6. 若 `gitignore` 不是 `off`，更新根 `.gitignore` 的 `# BEGIN stdagent` 块（dry-run 只报告不写盘）
7. 对每个 plan：与现有文件比对，backup 即将覆盖的旧文件，原子写入
8. 更新 `.stdai/state.json` 的 `last_sync` `outputs[]` `checksums`

`sync` 输出外部采用计数行 `[external] N adopted files (.ai/.agents)`（N 为 0 时不输出）。

| flag | 说明 |
|---|---|
| `--target <name>` | 多次传可只 sync 指定 target；未传 = 所有 enabled |
| `--dry-run` | 不写盘，输出 diff 摘要（外部采用仅内存预览，不落盘） |
| `--no-pull` | 跳过 pull |
| `--no-backup` | 跳过 backup |
| `--no-external` | 跳过外部产物自动采用 |
| `--strict` | 任何 warn 升级为 error |

## `stdagent import`

```
stdagent import [--dry-run]
```

把外部产物显式落盘到 `.stdai/standards/external/`（另合并根 `.mcp.json` 缺失键到 `mcp.json`），供检查与二次编辑后再 `sync`。非 dry-run 的 `sync` 默认已落盘采用外部产物（与 `import` 同一批来源）；`import` 用于在 sync 之外手动固化，或配合 `--dry-run` 预览。重复执行幂等。

映射规则：`.ai/guidelines/` 与 `.ai/rules/` 转 `rules`；单个 `.ai/skills/*.md` 合成为 `skills/<name>/SKILL.md`；成包的 `.ai/skills/<pkg>/` 与 `.agents/skills/<pkg>/` 原样采用（含辅助文件）；provenance 经 `external_source` / `external_path` frontmatter 保留。带 stdagent 生成标记的输出文件会被跳过，避免自循环。

保守过滤（默认启用，`import` 与 `sync` 一致）：

- 不吃未改写的自己生成物：上次 sync 写出的 tracked outputs（`state.json` 记录路径与 sha，含无 marker 的 skill 辅助文件）在磁盘内容一致时不采用，汇总为一条 `[external] skipped N self-generated files` warning；内容被外部改写则放行重新采用并 warning，上游更新不静默丢失。
- skill 包按整包判定：包内全部文件都是未改写的 tracked outputs 才整包跳过；只要主文件或任一辅助文件被改写，整包重新采用并输出一条 `[external] <pkg> changed outside stdagent, re-adopting the whole package`。逐文件判定会在上游只改 `SKILL.md` 时丢掉未变化的辅助文件，残包渲染后辅助文件被当孤儿删除。
- 同名外部 skill 包让位于本地显式源：`skills/` 下已有的同名包使整包剩余文件丢弃并 warning，避免同名双源造成 output collision（Boost 更新覆盖同名技能即属此类）；全包都是自生成物时静默跳过，不重复报 shadow 噪音。
- `.ai/rules/index.md` 是包管理器的发现索引，不是规则，跳过并 warning。
- rules 的 Boost 写法 `paths:`（标量或列表）在无 `applyTo:`/`globs:` 时自动复制为 `applyTo:`，原键保留；`applyTo:` 经 parser 生效为范围 globs。
- 根 `.mcp.json` 只补缺失 servers；旧版本写出的空 `"version": ""` 首次遇到时一次性去掉该键（仅一次，不造成 mtime 抖动）。

## `stdagent fix`

```
stdagent fix [--target <name>...]
```

`sync` 的语义别名，专为 drift auto-fix 场景设计。完全等价：
`stdagent sync` 已能写新文件覆盖 drift，`fix` 让命令更直觉。

## `stdagent status`

```
stdagent status [--target <name>...] [--json]
```

行为：

1. 读 `.stdai/state.json` 与 config
2. 比对当前扩散文件 sha256 vs 上次 sync 的 checksum
3. 输出每个 target 的 enabled / convert / last_sync / files_tracked / drift / missing
4. `--json` 输出机器可读版本

退出码：0 一致；1 有 drift / missing；2 配置错误。

## `stdagent clean`

```
stdagent clean [--target <name>...] [--keep-backups] [-y, --yes]
```

行为：

1. 读 `.stdai/state.json` 的 `outputs[]`
2. 对每个 target 删除曾经生成的文件 + 空父目录
3. 保留 `.stdai/`（包括 backups）
4. 默认交互确认；`--yes` 跳过

| flag | 说明 |
|---|---|
| `--target <name>` | 仅清理指定 target |
| `--keep-backups` | 保留 `.stdai/backups/`（默认即保留） |
| `--yes` `-y` | 跳过确认 |

退出码：0 成功；3 用户拒绝。

## `stdagent budget`

```
stdagent budget [--rendered] [--target <name>...] [--json]
```

默认估算 `.stdai/standards/` source 文档。`--rendered` 额外运行 target plan，
报告 source layers、每个 target 的实际 root 常驻体积和 sidecar 体积。

| flag | 说明 |
|---|---|
| `--json` | 结构化 JSON 输出（含每文档 path / type / bytes / estimated_tokens / warnings） |
| `--rendered` | 包含启用 target 的实际 plan 体积 |
| `--target <name>` | 限定 rendered target，可重复；隐式开启 `--rendered` |

token 估算用基于字符规则的近似（ASCII ~ 4 chars/token，中文 ~ 1.5 chars/token），
误差约 ±30%。如需精确估算可后续接 `tiktoken-go`。

## `stdagent intro`

```
stdagent intro [--json] [--copy]
```

输出 AI 助手提示词。把输出粘到 AI 对话开头（Claude / GPT / Gemini / Cursor
等），AI 就能理解如何：

- 编写 `.stdai/standards/` 下的 rules / skills / commands / references
- 从已有 `CLAUDE.md` / `.cursor/rules/` / `.clinerules/` / `.github/copilot-instructions.md`
  迁移到 std-agent
- 跑 stdagent 各命令完成同步与维护
- 遵守 frontmatter / 命名 / budget 限制等关键约束

提示词由 `internal/cli/intro_prompt.md` 通过 go embed 内嵌进 binary，
随版本演进同步更新。

| flag | 说明 |
|---|---|
| `--json` | 结构化输出 `{ "version": ..., "prompt": ... }`，适合编程消费 |
| `--copy` | 与 `--json` 联用时输出原始 markdown 不包 JSON（pipe 到 pbcopy 等剪贴板工具）|

典型用法：

```bash
# 直接看
stdagent intro

# 复制到 macOS 剪贴板
stdagent intro | pbcopy

# 喂给本地 LLM CLI
stdagent intro | llm "现在请帮我把项目里的 CLAUDE.md 迁移成 std-agent"

# 团队 AI 助手 system prompt
stdagent intro --json | jq -r .prompt
```

## `stdagent upgrade`

```
stdagent upgrade [--version vX.Y.Z] [--force]
```

安装源识别：先判断当前二进制是否为 Homebrew 管理（可执行路径解析后位于 `Cellar/stdagent/` 下，或 `STDAGENT_INSTALL_METHOD=brew` 显式指定）。brew 安装走 Homebrew 流程，否则走自替换流程。

Homebrew 流程：

1. 调 GitHub Releases API 拿 latest tag（已是最新且无 `--force` 则直接退出）
2. 转调 `brew upgrade StringKe/tap/stdagent`（`--version` 定版不被支持，会直接报错）

自替换流程：

1. 调 GitHub Releases API 拿 latest tag（或 `--version` 指定）
2. 下载平台对应归档：`std-agent_<ver>_<os>_<arch>.<tar.gz|zip>`
3. 下载 `checksums.txt` 校验 sha256
4. 解包提取 `stdagent` binary
5. `inconshreveable/go-update` 原子替换当前 executable

| flag | 说明 |
|---|---|
| `--version vX.Y.Z` | 指定目标 tag（含降级） |
| `--force` | 已是目标版本时仍强制重装 |

环境变量 `GITHUB_TOKEN` 自动用于 GH API（提升 rate limit）。

退出码：0 成功；非 0 各类失败（下载、checksum 不匹配、写盘等）。

## `stdagent version`

```
stdagent version [--json]
```

输出：

```
stdagent X.Y.Z
  commit:  <short-sha>
  built:   <RFC3339-utc>
  go:      go1.26.2
  os/arch: darwin/arm64
```

`--json` 输出结构化版本。

## 退出码约定

| 码 | 含义 |
|---|---|
| 0 | 成功 |
| 1 | 通用错误（IO、参数无效、drift 等） |
| 2 | 配置错误（schema 不合法） |
| 3 | 用户取消（确认拒绝） |

## stdin / stdout 约定

- 默认日志写 stderr，便于 `stdagent sync 2>/dev/null` 静音
- stdout 仅承载明确"数据"输出（`status --json`、`version --json`、`upgrade` 进度）
- 颜色仅在 stderr 是 TTY 时启用

## shell 补全

```
stdagent completion bash > /etc/bash_completion.d/stdagent
stdagent completion zsh  > "${fpath[1]}/_stdagent"
stdagent completion fish > ~/.config/fish/completions/stdagent.fish
```

由 cobra 自动提供。
