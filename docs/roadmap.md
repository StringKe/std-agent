# 路线图

## v1.0 MVP — 已落地

Tier 1 9 个 target 的核心同步流，覆盖 85%+ 用户场景。

- 已落地 `init` `pull` `sync` `status` `clean` `version` 6 个命令
- 已落地 YAML frontmatter + Markdown 解析与校验
- 已落地 9 个 target transformer（claude-code / codex / cursor / copilot /
  windsurf / gemini / aider / cline / opencode）
- 已落地 git 远端源拉取与本地缓存
- 已落地 backup + drift 检测
- 已落地 inject footer + inject_whatis
- 已落地 CI 多平台测试与跨平台编译
- 已落地 goreleaser 多平台 release（6 个 OS × ARCH）

## v1.1 — 已落地

扩展同步范围 + 维护体验。

- 已落地 MCP 转换：`.stdai/standards/mcp.json` 单源分发到 `.mcp.json` /
  `.cursor/mcp.json` / `.vscode/mcp.json`（顶级键 mcpServers vs servers 自动适配）
- 已落地 monorepo 支持：cwd 向上 walk 找 `.stdai/config.toml`
- 已落地 `stdagent fix` drift auto-fix 命令
- 已落地 `stdagent install-hook` git pre-commit 钩子
- 已落地 `stdagent upgrade` 自我升级（sha256 校验 + 原子替换）
- 已落地 git-cliff changelog（按 conventional commits 分组，无 email 暴露）
- 已落地 Tier 2 transformer：`continue-dev` + `antigravity`

## v1.2 — 进行中

### 已落地

- SKILL package multi-file 输出：claude-code / codex / cursor / windsurf 四个
  目录形 transformer 写完整 SKILL 目录（含 scripts/ references/ assets/
  templates/ examples/ 子目录的辅助文件）
- frontmatter schema 扩展：parser 接受 10 个 SKILL package 字段（when_to_use /
  arguments / effort / context / agent / shell / hooks / license / compatibility /
  metadata）
- AGENTS.md ## Slash Commands 段：codex transformer 在 AGENTS.md 末尾追加
  commands 段，aider 通过 `read:` 引用 AGENTS.md 时也能看到 slash 命令
- source.Local 收集 skills/ 子树下的非 markdown 辅助文件
- source.Git 复用 Local.Files 自动支持 SKILL package 辅助文件
- runner.collectSkillPackageFiles 把 skill 目录辅助文件附到对应 Document
- metadata YAML 嵌套写入到 4 个目录形 target frontmatter（FmBuilder.AddMap）
- copilot/opencode 单文件 skill 检测到 SkillFiles 通过 op.Reason 输出 WARN
- runner 收集 plan.Files 中 WARN reason 到 res.Warnings；cli/sync 输出到 stderr
- internal/budget 包：LLM 上下文消耗检查，rule/skill/command 各自软硬上限
  WARN（详见 docs/spec.md 4.10）
- `.stdaiignore`：gitignore 风格 glob，runner parse 前过滤源文件，doublestar 支持 `**`；
  `stdagent init` 默认生成模板（详见 docs/spec.md 4.11）
- hooks 转换：`.stdai/standards/hooks.json` 经 runner 注入 Config.Hooks，
  claude-code / codex transformer 输出 `.claude/stdagent-hooks.json` /
  `.codex/stdagent-hooks.json` 中间文件，由 `stdagent apply-hooks` 命令把 hooks 字段
  merge 到 `.claude/settings.json`（详见 docs/spec.md 4.12）
- 测试覆盖率：budget 100%、transformer 94.4%、parser 93.7%、source 84.5%、
  writer 83.2%、config 81.6%、state 73.9%、cli 70.2%、runner 68.8%

### 计划中

- 全局工具链：Homebrew Tap + Scoop bucket（已撤回，待用户需求时重启）
- Web Dashboard：可视化展示 std 文件、target diff、drift 状态
- 能力对比报告：哪些 rule 在哪些 target 生效
- VSCode/Cursor 扩展：直接编辑 `.stdai/standards/` 并预览 target 输出
- 团队级策略中心：组织级强制规则下发
- 远端 git 源也支持 SKILL package 辅助文件（v1.2 仅本地源）
- codex hooks schema 自动适配（v1.5 计划，目前 UNKNOWN）

## v1.3+ — 远期

- Tier 3 工具：JetBrains AI/Junie、Amazon Q、Tabnine、Replit Agent、Roo Code、
  Kiro、Zencoder、Augment、Lovable/Bolt.new/v0、CodeGPT、Qodo、Snyk Code、Warp
- AI 辅助编辑：根据用户描述自动生成 std 文件
- 共享市场：发布/安装 std 文件包（社区贡献）

## 不在路线图

- 自研 AI agent 运行时（保持工具中立，只做配置同步）
- 替代任何 target 工具的核心功能（hooks runner、MCP server 等）
- 闭源商业版（保持 OSS）
