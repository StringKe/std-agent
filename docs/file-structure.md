# 目录结构与扩散原则

## 设计哲学

Std-Agent 严格区分两类目录：

- **内部管理区** `.stdai/` 工具自己拥有，开发者只通过 `stdagent` CLI 间接修改
- **向外扩散区** 项目根 + `.claude/` / `.codex/` / `.cursor/` / `.github/` / `.windsurf/` /
  `.gemini/` 等平台目录，由 `stdagent sync` 写入；这些是 AI 工具实际消费的位置

核心约束：**工具只能写内部管理区与已声明的扩散区**。任何其他位置（如
`src/`、`scripts/`、`docs/`、`README.md`）禁止被 stdagent 触碰。

## 完整目录布局

```
my-project/
├── .stdai/                           内部管理区
│   ├── config.toml                   唯一配置文件
│   ├── state.json                    运行时状态（最后同步时间、checksums、生成文件清单）
│   ├── standards/                    单一真相源（本地缓存或手写）
│   │   ├── rules/
│   │   │   ├── coding-style.md
│   │   │   ├── git-commit.md
│   │   │   └── ...
│   │   ├── skills/
│   │   │   ├── code-review/
│   │   │   │   └── SKILL.md
│   │   │   └── ...
│   │   ├── commands/
│   │   │   ├── review.md
│   │   │   └── ...
│   │   └── references/
│   │       ├── architecture-overview.md
│   │       └── ...
│   ├── cache/                        远端 Git 源缓存（git clone 副本）
│   │   └── default/                  按 sources.<name> 命名的子目录
│   │       └── standards/
│   ├── backups/                      sync 前备份
│   │   └── 2026-05-07T17-30-00Z/
│   │       ├── CLAUDE.md
│   │       └── AGENTS.md
│   └── logs/                         可选 verbose 日志
├── CLAUDE.md                         向外扩散：Claude Code 项目根
├── AGENTS.md                         向外扩散：Codex / Aider / Cursor fallback / OpenCode
├── GEMINI.md                         向外扩散：Gemini CLI
├── .claude/                          向外扩散：Claude Code 专用
│   ├── rules/
│   ├── skills/
│   ├── commands/
│   └── agents/
├── .codex/                           向外扩散：Codex 专用
│   └── rules/
├── .cursor/                          向外扩散：Cursor 专用
│   ├── rules/                        .mdc 文件
│   └── mcp.json
├── .github/                          向外扩散：Copilot 专用
│   ├── copilot-instructions.md
│   └── instructions/
│       └── *.instructions.md
├── .windsurf/                        向外扩散：Windsurf 专用（如适用）
│   └── rules/
├── .windsurfrules                    向外扩散：Windsurf 项目级
├── .gemini/                          向外扩散：Gemini CLI
│   └── commands/
├── .clinerules/                      向外扩散：Cline
│   └── *.md
├── opencode.json                     向外扩散：OpenCode
└── ...                               项目其他文件，stdagent 永不触碰
```

## 写入清单（白名单）

`stdagent sync` 在每次执行时，只能写入以下路径之一：

| 范围 | 路径 | 说明 |
|---|---|---|
| 项目根标记文件 | `CLAUDE.md` `AGENTS.md` `GEMINI.md` `.windsurfrules` `opencode.json` | 各 target 要求位于根的入口 |
| Claude Code | `.claude/{rules,skills,commands,agents}/` `.claude/settings.json` `.mcp.json` | 见 targets/claude-code.md |
| Codex | `.codex/rules/` | 见 targets/codex.md |
| Cursor | `.cursor/rules/` `.cursor/mcp.json` | MDC 格式 |
| GitHub Copilot | `.github/copilot-instructions.md` `.github/instructions/` `.github/prompts/` `.github/chatmodes/` | 见 targets/github-copilot.md |
| Windsurf | `.windsurf/` `.windsurfrules` | 见 targets/windsurf.md |
| Gemini CLI | `.gemini/commands/` `.gemini/settings.json` | 见 targets/gemini-cli.md |
| Aider | `.aider.conf.yml`（如开启）| 不强制改写，可选 |
| Cline | `.clinerules/` | 多文件目录 |
| OpenCode | `opencode.json` | 见 targets/opencode.md |

任何不在白名单的路径，sync 不写也不删。

## 备份策略

- 每次 sync 之前，对将被覆盖的根目录文件（`CLAUDE.md` 等）以及平台目录，
  做快照到 `.stdai/backups/<RFC3339-utc>/`
- 备份保留 N 份（默认 5），超过自动清理最老
- `stdagent clean` 不删 backups
- 根 `.gitignore` 由 `gitignore` 配置维护（默认 `generated`）；详见下文

## .stdaiignore

`gitignore` 风格的 glob 文件，匹配的源文件不参与 parse / 扩散。`stdagent init` 默认生成模板。

- 文件位置：项目根 `.stdaiignore`
- pattern 路径相对 `.stdai/standards/`（与远端 source 合并后的逻辑根）
- 支持 doublestar `**` / `*` / `?`，`#` 行注释
- 行首 `!` 反向匹配 UNKNOWN（v1.3 仅支持正向 ignore，需要白名单走 `[targets].include` 或 frontmatter `targets`）

示例：

```
# 草稿 / WIP
rules/draft-*.md
**/wip-*.md

# 内部 only，不扩散
references/internal-*.md
```

被 ignore 的文件会在 sync 输出 `[ignore] skipped N files via .stdaiignore` 警告。

## 与 .gitignore 的关系

`init` 与 `sync` 维护根 `.gitignore` 中的 `# BEGIN stdagent` / `# END stdagent` 块，块外条目不改。这是对「只写内部管理区与已声明扩散区」的显式例外：只改该 managed 块。

`gitignore` 三种模式：

- `generated`（默认）：忽略运行时文件 + 全部可重建产物
- `portable`：同 `generated`，但保留公约集合 `AGENTS.md` 与 `.agents/`
- `off`：不改 `.gitignore`

`.stdai/config.toml` 与 `.stdai/standards/` 应被提交到 git。已跟踪的扩散文件不会因为新增 ignore pattern 自动从 index 消失。

## 反模式

禁止在以下情况发生写入：

- 写入项目其他源代码目录（`src/`、`pkg/`、`app/` 等）
- 在根目录创建 `README.md`、`LICENSE`、`Makefile` 等通用文件
- 在 `.stdai/` 之外创建临时文件、日志、缓存
- 修改用户已有的非生成文件（通过 marker 行 + checksum 检测）
