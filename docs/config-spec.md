# `.stdai/config.toml` 配置规范

## 设计目标

- 极简：核心字段不超过 20 个
- 默认即可用：MVP 默认 enable Claude Code + Codex
- 显式优于隐式：每个 target 单独 enabled 与 convert 开关
- 可读：toml 格式 + 注释友好

## 完整示例

```toml
version = "1.0"
name = "my-ai-standards"

# 注入控制
inject = true              # 是否在生成文件末尾追加 stdagent footer
inject_whatis = true       # footer 是否包含详细说明（来源、命令、文件作用）

# 全局开关
dry_run = false            # true 时只 diff 不写盘
backup = true              # sync 前备份将被覆盖的扩散文件
backup_keep = 5            # 备份目录保留数量
auto_pull = true           # sync 自动 pull 远端
verbose = false            # 详细日志

# 目标平台开关。enabled=false 完全跳过；convert=false 走 raw copy
[targets]
claude-code = { enabled = true,  convert = true }
codex       = { enabled = true,  convert = true }
cursor      = { enabled = false, convert = true }
copilot     = { enabled = false, convert = true }
windsurf    = { enabled = false, convert = true }
gemini      = { enabled = false, convert = true }
aider       = { enabled = false, convert = true }
cline       = { enabled = false, convert = true }
opencode    = { enabled = false, convert = true }

# 远端源（多源合并优先级：内置 < 默认 source < 后续 source；同名文件后者覆盖前者）
[sources.default]
url = "https://github.com/yourname/ai-standards.git"
branch = "main"
enabled = true
paths = ["standards/"]
auth = "ssh"               # ssh | https-token | none；none 仅 public

# 可选额外源
# [sources.team]
# url = "git@github.com:org/team-rules.git"
# branch = "main"
# enabled = true
# paths = ["rules/", "skills/"]

# 字段级别覆盖（per-target 高级覆盖，可选）
# [overrides.claude-code]
# inject = false
```

**重要约束**：toml 文件中所有顶层标量字段（`version` `name` `inject` `dry_run` 等）
必须放在第一个 `[section]`（如 `[targets]`）之前。一旦进入某个 `[section]`，
后续标量赋值会被解析为该 section 的子字段。
`Save()` 序列化时由 schema 字段顺序保证此约束。

## 字段语义

### 顶层

| 字段 | 类型 | 默认 | 说明 |
|---|---|---|---|
| `version` | string | "1.0" | 配置 schema 版本，stdagent 用于兼容判断 |
| `name` | string | 项目目录名 | 标识用，写入备份/日志 |
| `inject` | bool | true | 全局 footer 注入开关 |
| `inject_whatis` | bool | true | footer 包含详细说明 |
| `dry_run` | bool | false | 全局 dry-run |
| `backup` | bool | true | sync 前备份 |
| `backup_keep` | int | 5 | 备份保留份数 |
| `auto_pull` | bool | true | sync 自动 pull |
| `verbose` | bool | false | 详细日志 |

### `[targets]`

每个 target 是一张 inline table：`{ enabled = bool, convert = bool }`。

| 字段 | 类型 | 默认 | 说明 |
|---|---|---|---|
| `enabled` | bool | varies | true 时参与 sync |
| `convert` | bool | true | true 走平台特定转换；false 走 raw copy |

合法 target 名（与 frontmatter `targets` 字段对齐）:

```
claude-code  codex  cursor  copilot  windsurf  gemini  aider  cline  opencode
```

### `[sources.<name>]`

| 字段 | 类型 | 默认 | 说明 |
|---|---|---|---|
| `url` | string | 必填 | git URL（https / ssh） |
| `branch` | string | "main" | 分支名 |
| `enabled` | bool | true | 关闭后该源跳过 |
| `paths` | string[] | ["standards/"] | 仓库内要 sync 的子路径 |
| `auth` | string | "none" | ssh / https-token / none |
| `token_env` | string | "" | https-token 模式下读取的环境变量名 |

多源合并：按字典序遍历 sources，后者覆盖前者；本地手写 `.stdai/standards/`
始终最高优先级（不会被远端覆盖）。

### `[overrides.<target>]`

per-target 字段覆盖（可选高级用法）。允许的 keys 与顶层全局开关同名子集
（如 `inject` / `inject_whatis`）。

## 默认 enable 策略

`stdagent init` 生成的默认配置中：

- `claude-code` `codex` 默认 `enabled = true`（覆盖 P0 场景）
- 其他 Tier 1 target 默认 `enabled = false`，需用户显式开启
- 全部 target 默认 `convert = true`

## 校验规则

stdagent 启动时执行：

1. `version` 必须存在且 stdagent 已知（否则报 schema 升级提示）
2. 每个 target 必须是合法 target 名
3. `inject_whatis` 仅当 `inject = true` 时才生效，否则忽略并 warn
4. `[sources.X]` 的 `url` 必填，`paths` 至少一项
5. `auth = "https-token"` 时 `token_env` 必填且对应环境变量必须存在
6. `backup_keep ≥ 1`

## 与环境变量的交互

| 环境变量 | 作用 |
|---|---|
| `STDAI_CONFIG` | 覆盖 `.stdai/config.toml` 路径 |
| `STDAI_DRY_RUN=1` | 等价 `dry_run = true` |
| `STDAI_VERBOSE=1` | 等价 `verbose = true` |
| `STDAI_NO_PULL=1` | 等价 `auto_pull = false` |
| `STDAI_<TOKEN_ENV_NAME>` | 由 `token_env` 字段引用 |

CLI flag > 环境变量 > config.toml。

## per-subdir 配置（v1.1）

monorepo 场景下，子目录可有自己的 `.stdai/config.toml`，仅影响子目录扩散。
v1.0 不支持，v1.1 引入。
