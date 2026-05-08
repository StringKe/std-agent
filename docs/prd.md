# Std-Agent PRD

**版本** v1.0（2026-05-07）
**产品名** Std-Agent
**CLI 命令** `stdagent`
**一句话定义** 轻量级、纯 Go 实现的 AI CLI 能力标准化同步工具，以 `.stdai/`
为内部单一真相源，通过 YAML Frontmatter + Markdown 格式，实现 Claude Code、
Codex、Cursor 等多平台 `skills` / `rules` / `commands` / `references` 的智能
同步与差异化转换，严格遵守"内部管理 + 向外扩散"的目录原则。

## 1. 产品背景与核心问题

当前 AI 开发环境中，主流工具的配置高度碎片化：

- Claude Code（Anthropic）使用 `CLAUDE.md` + `.claude/rules/` + `skills/` +
  `hooks` + `policy` + `memories` + MCP 等专有分层机制
- Codex（OpenAI）使用 `AGENTS.md` + `.codex/rules/` 等开放标准
- Cursor、GitHub Copilot、Windsurf、Gemini CLI 等各有自己的格式

关键痛点：

1. Claude Code 的 rules 与 Codex 的 rules **不是同一个东西**。格式、加载机制、
   支持特性完全不同，必须显式区分处理。
2. 手动维护多份文件容易漂移、重复劳动、团队不一致。
3. 用户明确要求：**不允许工具在项目根目录随意"拉屎"**，但最终生效的规则
   文件（`CLAUDE.md`、`AGENTS.md` 等）必须生成在根目录及平台目录中。
4. 需要支持远端 Git 单一源头 + 本地缓存 + 默认内置标准。

Std-Agent 的使命：建立 `.stdai/` 内部管理区 + 自动向外扩散机制，让开发者
一次维护、多平台自动生效，且工具自身不污染项目其他文件。

## 2. 产品目标与成功标准

主要目标：

- 多 AI CLI 能力标准化同步（skills、rules、commands、references 四类）
- 智能处理平台差异（默认全转换，可单独关闭转换走 raw copy）
- 目录严格分离：`.stdai/` 仅内部管理，生成文件扩散到根目录
- 配置极简、开箱即用、易于团队 Git 共享

成功标准：

- `stdagent sync` 后根目录文件与 `.stdai/standards/` 保持一致
- 至少 8 个主流 target 在 MVP 内可用（见第 3 节 Tier 1）
- 漂移检测 < 1 秒
- 单 binary，无运行时外部依赖，跨平台（macOS/Linux/Windows）

## 3. 范围

### 3.1 包含（MVP/v1.0）

- YAML Frontmatter + Markdown 统一源格式
- 4 种内容类型：skills / rules / commands / references
- Targets 极简开关 + per-target convert 开关（默认全转换）
- Git 远端源 + 内置默认标准
- 注入 Footer 控制（`inject` + `inject_whatis`）
- 命令：`init` / `pull` / `sync` / `status` / `clean` / `version`
- Claude Code 与 Codex 差异化转换（hooks、尾部菜单）

### 3.2 不包含（v1.0）

- Web Dashboard（v1.1 考虑）
- MCP 自动转换（v1.1）
- VSCode/Cursor 扩展（v1.2）
- 复杂冲突交互合并（先 backup + 提示）

### 3.3 MVP 优先级（Tier 划分）

#### Tier 1 必须支持（MVP 核心，覆盖 85%+ 用户场景）

| 优先级 | 工具 | 类型 | 配置/规则机制 | 备注 |
|---|---|---|---|---|
| P0 | Claude Code | CLI + Desktop + Agent | `CLAUDE.md` + `.claude/{rules,skills,commands,agents}/` + hooks/policy/MCP | 专有格式最丰富，差异最大 |
| P0 | Codex (OpenAI) | CLI + IDE + ChatGPT | `AGENTS.md` + `.codex/rules/` | AGENTS.md 准标准 |
| P0 | Cursor | AI-native IDE | `.cursor/rules/*.mdc`（YAML frontmatter）+ AGENTS.md fallback | 社区最大 |
| P1 | GitHub Copilot | IDE + CLI | `copilot-instructions.md` + `.github/instructions/*.md` + AGENTS.md | 企业市场最大 |
| P1 | Windsurf | AI IDE | `.windsurfrules` + Cascade Flows | 增长迅猛 |
| P1 | Gemini CLI | CLI | `GEMINI.md` + `.gemini/` extensions | 长上下文 |
| P1 | Aider | Git-native CLI | `AGENTS.md` + `.aider.conf.yml` | 终端老玩家 |
| P1 | Cline | VS Code 扩展 + CLI | `.clinerules/` 目录 | 开源、多模型 |
| P1 | OpenCode | 开源 CLI | `AGENTS.md` + `opencode.json` | BYOK 用户 |

#### Tier 2 强烈建议早期支持

Continue.dev、Antigravity（Google）。

#### Tier 3 后期扩展

JetBrains AI/Junie、Amazon Q Developer、Tabnine、Replit Agent、Roo Code、
Kiro、Zencoder、Augment、Lovable/Bolt.new/v0、CodeGPT、Qodo、Snyk Code、Warp 等。

## 4. 目录结构规范

```
my-project/
├── .stdai/                          内部管理区，工具专属，不污染其他
│   ├── config.toml                  唯一配置文件
│   ├── standards/                   单一真相源（Git pull 或内置）
│   │   ├── rules/
│   │   ├── skills/
│   │   ├── commands/
│   │   └── references/
│   ├── cache/                       Git 远端缓存
│   └── backups/                     每次 sync 前备份根目录文件
├── CLAUDE.md                        stdagent 生成/覆盖
├── AGENTS.md                        stdagent 生成/覆盖
├── .claude/                         Claude Code 专用（向外扩散）
│   ├── rules/
│   ├── skills/
│   └── commands/
├── .codex/                          Codex 专用（向外扩散）
├── .cursor/                         Cursor 专用
└── ...                              项目其他文件
```

核心原则：`.stdai/` 只存放工具内部数据；所有最终生效文件向外扩散到根目录
及平台目录。详见 [file-structure.md](file-structure.md)。

## 5. 源文件格式规范

所有 `.stdai/standards/` 下的文件必须为 Markdown + YAML Frontmatter：

```markdown
---
type: rules
name: coding-style
version: 1.2
targets:
  - claude-code
  - codex
priority: high
tags: [style, security]
---

# 正文内容
- Always use meaningful variable names
- ...
```

详见 [format-spec.md](format-spec.md)。

## 6. 配置 `.stdai/config.toml`（极简）

```toml
version = "1.0"
name = "my-ai-standards"

# 全局开关：标量字段必须在 [targets]/[sources] 之前，否则被 toml 解析进上一个 section
inject = true
inject_whatis = true
dry_run = false
backup = true
backup_keep = 5
auto_pull = true

[targets]
claude-code = { enabled = true, convert = true }
codex       = { enabled = true, convert = true }
cursor      = { enabled = false, convert = true }
copilot     = { enabled = false, convert = true }
windsurf    = { enabled = false, convert = true }
gemini      = { enabled = false, convert = true }
aider       = { enabled = false, convert = true }
cline       = { enabled = false, convert = true }
opencode    = { enabled = false, convert = true }

[sources.default]
url = "https://github.com/yourname/ai-standards.git"
branch = "main"
enabled = true
paths = ["standards/"]
```

详见 [config-spec.md](config-spec.md)。

## 7. 主要命令清单

| 命令 | 描述 |
|---|---|
| `stdagent init` | 创建 `.stdai/` + `config.toml` + 示例 standards |
| `stdagent pull` | 更新 `.stdai/cache/` 中的 Git 源 |
| `stdagent sync` | 核心命令：pull -> parse -> convert -> 向外扩散 |
| `stdagent status` | 显示 targets 状态、drift 检测、最后同步时间 |
| `stdagent clean` | 清空根目录生成文件，保留 `.stdai/` |
| `stdagent version` | 打印版本与构建信息 |

详见 [commands.md](commands.md)。

## 8. 同步与转换逻辑

流程：

```
git pull (auto_pull=true)
  -> 读取 .stdai/standards/ 全部文件
  -> 解析 frontmatter
  -> 对每个 enabled target:
       convert=true  执行平台特定转换（Claude hooks/policy；Codex AGENTS.md 尾部追加 Rules Reference 菜单等）
       convert=false 直接 raw copy
  -> backup（如开启）
  -> 写入根目录 + 平台目录
  -> 更新 .stdai/state.json（last_sync、checksum）
```

注入 Footer：当 `inject = true` 时，在生成文件末尾追加 `Generated by stdagent`
Footer；`inject_whatis` 控制是否包含详细说明（来源、命令、文件作用）。

详见 [conversion-rules.md](conversion-rules.md)。

## 9. 非功能需求

- 技术栈：Go 1.26+，单 binary（goreleaser 发布）
- 依赖：cobra、go-toml/v2、go-git/v5、gopkg.in/yaml.v3
- 兼容性：macOS / Linux / Windows（symlink 失败自动 fallback copy）
- 安全：自动 backup、`.stdaiignore` 支持、私有 Git 认证（SSH key、HTTPS token）
- 性能：小项目 < 2 秒，大项目 < 5 秒
- 可扩展：Target 实现 interface，新增 target 只需添加 transformer

## 10. 用户故事

1. 作为开发者，`stdagent init` 后立刻获得干净的 `.stdai/` 结构和示例文件
2. 作为多 AI 用户，修改一次 `.stdai/standards/` 或 Git 源头，自动同步到所有
   enabled 平台
3. 作为团队 Lead，新成员只需 `git clone` + `stdagent sync` 即可获得一致配置
4. 作为 Claude + Codex 重度用户，工具智能处理两者规则差异，而非简单复制
5. 作为注重项目干净的用户，工具只在 `.stdai/` 内管理，生成文件精准扩散到
   根目录与平台目录

## 11. Edge Cases

- 私有 Git 仓库认证失败：报错 + 提示配置 SSH/token，不静默 fallback
- Windows symlink 权限不足：自动 fallback copy + warn 日志
- 无 Frontmatter 文件：降级为纯 markdown rule（type 推断为 rules）
- 冲突：根目录已存在用户手写的 CLAUDE.md，sync 前 backup 到
  `.stdai/backups/<timestamp>/CLAUDE.md` 后再覆盖；首次提示用户
- 大型 monorepo：v1.1 支持 per-subdir `.stdai/config.toml`
- 敏感规则：推荐 private Git + `.stdaiignore`

## 12. 路线图

详见 [roadmap.md](roadmap.md)。

- v1.0 MVP：Tier 1 9 个 target、基础同步流
- v1.1：MCP 配置转换、pre-commit hook、drift auto-fix、Tier 2
- v1.2：Web Dashboard、能力对比报告、VSCode/Cursor 扩展
- v1.3：团队协作面板、Tier 3 工具
