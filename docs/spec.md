# std-agent Specification

stdagent 跨工具 AI 配置同步标准与转换规则的权威文档。本 spec 定义：

- std-agent 自身的源文件 schema（YAML frontmatter + Markdown）
- 四种内容类型（rules / skills / commands / references）的语义
- 23 个目标工具中相同概念的差异
- 转换器实现策略与边界情况

读者：工具实现者、高级用户、新 target 集成方。

## Part 1: std-agent 标准

### 1.1 文件位置约定

```
my-project/
├── .stdai/
│   ├── config.toml                  主配置（toml）
│   └── standards/                   源数据根
│       ├── rules/<name>.md          type: rules
│       ├── skills/<name>/           type: skills（目录 package，详见 1.7）
│       │   ├── SKILL.md             frontmatter + 主文本（必填）
│       │   ├── scripts/             可执行代码（可选）
│       │   ├── references/          按需加载文档（可选）
│       │   ├── assets/              模板/数据/图片（可选）
│       │   ├── templates/           代码模板（可选）
│       │   └── examples/            使用示例（可选）
│       ├── commands/<name>.md       type: commands
│       ├── references/<name>.md     type: references
│       └── mcp.json                 MCP 服务器配置（独立 JSON）
```

### 1.2 命名规则

- name 必须 kebab-case：正则 `^[a-z0-9][a-z0-9-]*$`
- 同 type 内 name 唯一
- skill 目录名等于 SKILL.md frontmatter `name`
- 文件无 frontmatter 时降级为 type=rules，name 由文件名推导（去扩展名 + 小写 + `_`/`.` 替换为 `-`）

### 1.3 YAML Frontmatter 完整 Schema

```yaml
---
# 必填
type: rules                  # rules | skills | commands | references
name: coding-style           # kebab-case，同 type 内唯一

# 元信息（可选）
version: "1.2"               # semver
description: |               # 单行或多行
  General coding style.
priority: high               # high | normal | low (default: normal)
tags: [style, security]      # 自由标签

# target 限定（互斥：targets 与 exclude_targets 二选一）
targets:                     # 白名单
  - claude-code
  - codex
exclude_targets: []          # 黑名单

# 路径限定（rules / skills 有效）
applyTo:                     # gitignore-style glob 列表
  - "**/*.ts"
alwaysApply: false           # rules: true 时无视 applyTo 始终生效

# command 专属
allowed_tools:               # slash 内可用工具白名单
  - Read
  - "Bash(git status:*)"
argument_hint: "<file path>"

# 通用模型限定（commands / skills）
model: sonnet

# skills 专属
disable_model_invocation: false   # true 时仅响应显式 slash 调用
---

正文 Markdown 从这里开始...
```

### 1.4 字段适用矩阵

| 字段 | 类型 | 默认 | 必填 | 适用 type |
|---|---|---|---|---|
| `type` | enum | - | 是 | 全部 |
| `name` | string (kebab-case) | 文件名推导 | 否（推荐填） | 全部 |
| `version` | string | "1.0.0" | 否 | 全部 |
| `description` | string | "" | model_decision rule 必填 | 全部 |
| `priority` | enum | normal | 否 | 全部 |
| `tags` | string[] | [] | 否 | 全部 |
| `targets` | string[] | [] | 否（与 exclude_targets 互斥） | 全部 |
| `exclude_targets` | string[] | [] | 否 | 全部 |
| `applyTo` | glob[] | [] | 否 | rules / skills |
| `alwaysApply` | bool | false | 否 | rules |
| `allowed_tools` | string[] | [] | 否 | commands |
| `argument_hint` | string | "" | 否 | commands |
| `model` | string | "" | 否 | commands / skills |
| `disable_model_invocation` | bool | false | 否 | skills |

### 1.5 校验规则

stdagent parse 阶段强制校验，违反整体 sync 中止：

1. frontmatter 是合法 YAML（缺失则降级 type=rules + WARN）
2. type 在四种枚举内
3. name 满足 kebab-case 正则
4. 同 type 下 name 唯一
5. targets 与 exclude_targets 不能同时非空
6. targets / exclude_targets 内值必须是合法 target 名（23 个）
7. priority 在 enum 内

合法 target 名（23 个）：

```
claude-code  codex  cursor  copilot  windsurf  gemini  aider  cline  opencode
roo-code  crush  amp  warp  factory
continue-dev  antigravity  qwen-code  pi  kilo-code  augment-code
jules  grok-build  kimi-code
```

### 1.7 SKILL Package（目录形式）

四家工具（Anthropic Claude Code、OpenAI Codex、Cursor、Windsurf）以及更多生态
已收敛到 **Agent Skills 开放标准**（[agentskills.io](https://agentskills.io)，
Anthropic 发起后开源）。std-agent 的 skill 直接对齐此标准。

#### 1.7.1 目录结构

```
.stdai/standards/skills/code-review/
├── SKILL.md                必填，frontmatter + 主文本
├── scripts/                可执行代码（progressive disclosure 第 3 段加载）
│   └── lint.sh
├── references/             按需加载文档
│   └── checklist.md
├── assets/                 模板 / 数据 / 图片
│   └── pr-template.md
├── templates/              代码模板（Claude Code 习惯）
└── examples/               使用示例
```

子目录无强制约定，但建议遵循 `scripts/` `references/` `assets/` `templates/`
`examples/` 五个标准目录名以最大化跨工具兼容。

#### 1.7.2 SKILL.md frontmatter 完整 schema

最小核（agentskills 标准必填）：

| 字段 | 约束 | 说明 |
|---|---|---|
| `name` | 小写字母+数字+连字符，<= 64，必须与父目录名一致 | skill 标识 |
| `description` | <= 1024 字符；含"做什么 + 何时用" | progressive disclosure 第 1 段恒载 |

agentskills 可选字段（多家共通）：

| 字段 | 类型 | 说明 |
|---|---|---|
| `license` | string | 许可证（如 MIT） |
| `compatibility` | string / list | 工具兼容性声明 |
| `metadata` | map | 自由键值（author / version / tags 等） |
| `allowed_tools` | string / list | 实验字段，skill 激活时免授权工具 |

Claude Code 私有扩展（其他 target 转换时降级或丢弃）：

| 字段 | 类型 | 转换策略 |
|---|---|---|
| `when_to_use` | string | std-agent 拼接到 description 末尾 |
| `paths` | glob list | 直接写入支持的 target；不支持的丢弃 |
| `disable_model_invocation` | bool | Cursor / Claude Code 直传；其他 target 转化为 description hint |
| `argument_hint` | string | commands transformer 已有；skills 在 Claude Code 直传 |
| `arguments` | string / list | Claude Code 专属；其他 target 丢弃 |
| `model` | string | Claude Code / Cursor 直传；其他丢弃 |
| `effort` | enum: low/medium/high/xhigh/max | Claude Code 专属 |
| `context` | enum: fork | Claude Code 专属（子代理隔离） |
| `agent` | string | Claude Code 专属（context: fork 时指定子代理类型） |
| `shell` | enum: bash/powershell | Claude Code 专属 |
| `hooks` | object | Claude Code 专属 |

std-agent SKILL.md 接受所有上述字段；每个 transformer 按 target 能力裁剪。
未识别字段保留在 metadata 自由通道。

#### 1.7.3 progressive disclosure 三段式加载

```
启动:  仅载 name + description（约 100 tokens）
命中:  整 SKILL.md（建议 <= 5000 tokens / 500 行）
按需:  scripts/ references/ assets/ 子文件（模型按相对路径读）
```

写 SKILL.md 时遵循建议：

- description 写"何时用 + 做什么"（不超 1024 字符），让模型在第 1 段判断是否拉取
- SKILL.md 主文本 <= 500 行（Claude Code 软限制；其他 target 不强制但建议）
- 长内容拆到 `references/<topic>.md`，主文本里用相对 markdown 链接 `[ref](references/topic.md)`
- 脚本不嵌入 SKILL.md，放 `scripts/<file>` 由模型按需读取

#### 1.7.4 引用机制

- markdown 链接：`[checklist](references/checklist.md)`（相对于 SKILL.md）
- 脚本路径：`scripts/lint.sh`（相对）
- Claude Code 还支持：`${CLAUDE_SKILL_DIR}/scripts/lint.sh` 绝对路径替换、`` !`cmd` `` 行内
  shell 注入；其他 target 不识别这两类语法，std-agent 转换器对 Claude Code 直传，对其他降级为
  普通文本

#### 1.7.5 调用方式

| 模式 | 触发 | 配置 |
|---|---|---|
| 自动（model decision） | 模型读 description 决定 | 默认行为 |
| 显式 slash | `/skill-name` | Claude Code / Cursor / Codex 支持 |
| 显式 @-mention | `@skill-name` | Windsurf 唯一方式 |
| 仅 slash（非自动） | `disable_model_invocation: true` | Claude Code / Cursor 支持 |

### 1.7.6 实施进度（v1.2 已落地）

parser 与 runner 层
- parser 接受全部 10 个 SKILL package 字段（when_to_use / arguments /
  effort / context / agent / shell / hooks / license / compatibility /
  metadata）并存入 `Document`
- source.Local.Files 收集 `skills/` 子树下所有非 markdown 文件
  （scripts/ references/ assets/ 等）
- runner 在 parse 阶段跳过 `skills/<name>/<subdir>/*.md`（视为 SKILL
  package 辅助文件而非独立 Document）
- runner.collectSkillPackageFiles 把同 skill 目录辅助文件附到
  `Document.SkillFiles`
- 远端 git 源（source.Git.Files）通过复用 Local.Files，自动也支持 SKILL
  package 辅助文件

transformer 输出层
- claude-code / codex / cursor / windsurf 四个目录形 transformer 输出完整
  skill package（SKILL.md + 全部辅助子目录），按各 target 字段支持范围裁剪
  frontmatter
- 不支持 `when_to_use` 字段的 target（codex / cursor / windsurf）自动把
  when_to_use 拼到 description 末尾（`MergeDescription` helper）
- claude-code / codex / cursor / windsurf 输出 `metadata` 自由 map（YAML
  嵌套），claude-code 还输出 `hooks`（其他 target 私有字段未承诺，丢弃）
- copilot / opencode 单文件 skill 检测到 `SkillFiles` 非空时通过
  `op.Reason` WARN 提示用户（`.github/agents/<n>.agent.md` 与
  `.opencode/agents/<n>.md` 不支持 SKILL package 子目录）

诊断与可观察性
- runner.Sync 在 transformer.Plan 后扫描 `plan.Files`，把带 `WARN`
  前缀的 `Reason` 收集到 `res.Warnings`
- cli/sync.go 在末尾把 `res.Warnings` 输出到 stderr，让 SkillFiles 被忽略
  等降级行为对用户可见

### 1.8 MCP 配置（独立 JSON）

`.stdai/standards/mcp.json` 不走 frontmatter，是独立 JSON：

```json
{
  "version": "1.0",
  "servers": {
    "github": {
      "type": "stdio",
      "command": "gh",
      "args": ["api"],
      "env": {"TOKEN": "${env:GH_TOKEN}"}
    },
    "linear": {
      "type": "http",
      "url": "https://mcp.linear.app/sse",
      "headers": {"Authorization": "Bearer ${env:LINEAR_TOKEN}"}
    }
  }
}
```

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `version` | string | 是 | "1.0" |
| `servers` | object | 是 | 键为 server 名 |
| `servers[name].type` | enum: stdio/http/sse | 否 | 推断 |
| `servers[name].command` | string | type=stdio 必填 | 可执行路径 |
| `servers[name].args` | string[] | 否 | stdio 启动参数 |
| `servers[name].env` | object | 否 | stdio 环境变量 |
| `servers[name].url` | string | type=http/sse 必填 | server URL |
| `servers[name].headers` | object | 否 | http/sse 自定义 header |

## Part 2: 四种概念的语义

### 2.1 rules（规则）

**定义**：常驻或条件触发的 system prompt 增量。

- 文本直接进 LLM context
- 全局（始终注入）或路径条件（命中文件时注入）
- 不主动执行，仅描述约束

**std-agent 落地**：`.stdai/standards/rules/<name>.md`

**典型用例**：编码风格、git commit 规范、安全检查、测试约定。

### 2.2 skills（可复用能力）

**定义**：按 [agentskills.io](https://agentskills.io) 开放标准的目录 package，
模型按需拉取的"专项指令包"。

- 目录结构：`SKILL.md` + 可选 `scripts/` `references/` `assets/` `templates/` `examples/`
- progressive disclosure 三段式加载：description（恒载）-> SKILL.md（命中）-> 子文件（按需）
- LLM 根据 frontmatter description 决定是否拉取
- 跨工具共识：Anthropic / OpenAI / Cursor / Windsurf / 其他生态都消费同一目录

**std-agent 落地**：`.stdai/standards/skills/<name>/`（package 形式）

**完整 schema 见 1.7 节**。

**典型用例**：code review checklist、安全审计步骤、API 设计模板、部署 runbook。

### 2.3 commands（slash 命令）

**定义**：用户显式触发的提示模板，通常映射为 `/<name>`。

- 不自动注入，由用户输入触发
- 接受参数（`$1` `$ARGUMENTS`）
- 可限制可用工具

**std-agent 落地**：`.stdai/standards/commands/<name>.md`

**典型用例**：`/review` 跑代码审查、`/explain` 解释选中代码、`/deploy` 启动部署流。

### 2.4 references（参考资料）

**定义**：人类可读的支撑文档，不主动进 LLM context。

- 不参与自动加载
- 通过 `@<path>` import 或 `docs:` 索引引用

**std-agent 落地**：`.stdai/standards/references/<name>.md`

**典型用例**：架构概览、API 端点列表、术语表。

## Part 3: 23 个目标工具中的概念差异

### 3.1 总览矩阵（11 列）

| 概念 | claude-code | codex | cursor | copilot | windsurf | gemini | aider | cline | opencode | continue-dev | antigravity |
|---|---|---|---|---|---|---|---|---|---|---|---|
| **rules** (无 applyTo) | `.claude/rules/<n>.md` | 拼到 `AGENTS.md` | `.cursor/rules/<n>.mdc` (Always/AgentReq) | `.github/copilot-instructions.md` | `.windsurf/rules/<n>.md` (always_on) | 拼到 `GEMINI.md` | `read:` 引用 AGENTS.md | `.clinerules/<NNN>-<n>.md` | `AGENTS.md` 复用 | `.continue/rules/<n>.md` | `.agents/rules/<n>.md` (always_on) |
| **rules** (有 applyTo) | 同上 + frontmatter applyTo | `<sub>/AGENTS.md` | `.cursor/rules/<n>.mdc` (globs) | `.github/instructions/<n>.instructions.md` (applyTo) | `.windsurf/rules/<n>.md` (glob) | `<sub>/GEMINI.md` | 同 rules | `.clinerules/<NNN>-<n>.md` (paths) | applyTo 信息丢弃 | `.continue/rules/<n>.md` (globs) | `.agents/rules/<n>.md` (glob) |
| **skills** | `.claude/skills/<n>/SKILL.md` | `.agents/skills/<n>/SKILL.md` | `.cursor/skills/<n>/SKILL.md` | `.github/agents/<n>.agent.md` | `.windsurf/skills/<n>/SKILL.md` | 降级 commands | 不支持 | `.clinerules/workflows/<n>.md` | `.opencode/agents/<n>.md` (mode=subagent) | 降级 model_decision rule | 降级 model_decision rule |
| **commands** | `.claude/commands/<n>.md` | `.agents/skills/cmd-<n>/SKILL.md`（降级为 skill） | `.cursor/commands/<n>.md` | `.github/prompts/<n>.prompt.md` | `.windsurf/workflows/<n>.md` | `.gemini/commands/<n>.toml` | 不支持 | `.clinerules/workflows/<n>.md` | `.opencode/commands/<n>.md` | `.continue/prompts/<n>.prompt.md` | `.agents/workflows/<n>.md` |
| **references** | `@<path>` import | 不写 | 不写 | 不写 | 不写 | `@<path>` import | `read:` 引用 | `memory-bank/<6 文件>` | `@<path>` 变量替换 | rule 内嵌 / `docs:` 索引 | rule 正文 `@<file>` 引用 |
| **MCP** | `.mcp.json` (mcpServers) | 跳过 | `.cursor/mcp.json` (mcpServers) | `.vscode/mcp.json` (servers) | 跳过（用户级） | 跳过 | 不支持 | 不支持 | 跳过 | 跳过 | 跳过（用户级） |

### 3.2 关键差异

#### 同名不同义：rule

| 工具 | rule 是什么 |
|---|---|
| Claude Code | 路径条件加载的 markdown，frontmatter 含 applyTo |
| Codex | 没有独立 rule，AGENTS.md 全文即 rule |
| Cursor | `.mdc`，4 种激活模式（Always / Auto-Attached / Agent Requested / Manual） |
| Copilot | `instructions.md` 带 applyTo glob |
| Windsurf | trigger 4 模式（always_on / glob / model_decision / manual） |
| Antigravity | trigger 4 模式（同 Windsurf） |
| Continue | name + globs + alwaysApply |
| Cline | 目录文件自动合并，frontmatter `paths:` 条件激活 |
| Aider | 不支持，仅 `read: AGENTS.md` 显式声明 |

stdagent 的统一 frontmatter（applyTo + alwaysApply + description）覆盖以上差异，
每个 transformer 推导对应字段。

#### 同名不同义：skill

| 工具 | skill 是什么 |
|---|---|
| Claude Code | `.claude/skills/<n>/SKILL.md`，frontmatter name/description/model/tools |
| Codex / Antigravity | `.agents/skills/`（OpenAI agents 协议，部分 schema UNKNOWN） |
| Cursor | `.cursor/skills/<n>/SKILL.md`，含 disable-model-invocation |
| Copilot | `.github/agents/<n>.agent.md`（VS Code 旧 chat-modes） |
| Windsurf | `.windsurf/skills/<n>/SKILL.md` |
| OpenCode | `.opencode/agents/<n>.md`（mode=subagent） |
| Cline | 无原生概念，降级为 workflow |
| Continue / Gemini / Aider | 无概念，降级为 model_decision rule 或跳过 |

#### 同义不同名：commands

| 工具 | 文件路径 | 内部叫法 |
|---|---|---|
| Claude Code | `.claude/commands/<n>.md` | slash command |
| Cursor | `.cursor/commands/<n>.md` | slash command |
| Copilot | `.github/prompts/<n>.prompt.md` | prompt file |
| Windsurf | `.windsurf/workflows/<n>.md` | workflow |
| Gemini | `.gemini/commands/<n>.toml` | command（toml 格式） |
| OpenCode | `.opencode/commands/<n>.md` | command |
| Continue | `.continue/prompts/<n>.prompt.md` | prompt |
| Antigravity | `.agents/workflows/<n>.md` | workflow |
| Cline | `.clinerules/workflows/<n>.md` | workflow |
| Codex | 无（自定义 prompt 已 deprecated） | - |
| Aider | 无（不可扩展） | - |

stdagent 的 `commands/` 类型按各 target 路径与格式输出。

### 3.3 主入口文件

| 文件 | 写入者 | 消费方 |
|---|---|---|
| `CLAUDE.md` | claude-code | Claude Code |
| `AGENTS.md` | enabled producer plans + runner canonicalization | Codex / Cursor (fallback) / Copilot Coding Agent / OpenCode / Aider (经 read:) / Windsurf / Antigravity 等 |
| `GEMINI.md` | gemini | Gemini CLI |
| `.clinerules/` 多文件 | cline | Cline |
| `.continue/rules/` 多文件 | continue-dev | Continue.dev |
| `.agents/rules/` 多文件 | antigravity | Antigravity |
| `.cursor/rules/*.mdc` | cursor | Cursor |
| `.github/copilot-instructions.md` | copilot | Copilot 全家桶 |
| `.windsurf/rules/*.md` | windsurf | Windsurf |

`AGENTS.md` 被多个工具消费。所有 producer plan 在写入前由 runner 统一为 target-neutral rules 内容。

## Part 4: 转换实现策略

### 4.1 数据流

```
.stdai/standards/<type>/<name>.md
    -> parser.Parse(file)
    -> Document{Type, Name, frontmatter, Body}
    -> transformer.FilterDocs(target) 仅适用本 target 的子集
    -> all transformer.Plan(docs, cfg)
    -> shared-path canonicalization + collision validation
    -> Plan{Files: [FileOp...]}
    -> writer.Apply(plan) 原子写盘 / dry-run / drift 跳过
```

每个 target 实现 `transformer.Transformer` 接口：

```go
type Transformer interface {
    Name() string
    Plan(docs []*parser.Document, cfg *config.Config) (*writer.Plan, error)
}
```

### 4.2 frontmatter 字段映射

详见 [conversion-rules.md](conversion-rules.md) 第 3 节。

### 4.3 文件大小上限

| 目标 | 上限 | 处理 |
|---|---|---|
| Codex AGENTS.md | project_doc_max_bytes 默认 32768 字节 | WARN；精简 rules 或调整 target filter |
| Cursor 通用 rule | 100000 字符 | WARN |
| Windsurf 单 rule | 12000 字符 | WARN |
| Antigravity 单 rule / workflow | 12000 字符 | WARN |

### 4.4 不写主入口

aider / opencode / antigravity 等不生成根 AGENTS.md；它们复用其他启用 producer
生成并由 runner canonicalize 的共享文件。

### 4.5 降级处理

完整降级链（按 target 能力链式选择）：

```
commands:
  原生 slash 支持      -> 写 commands 文件
  不支持 commands 但支持 skills -> 写 skill (description 加 slash 调用 hint)
  都不支持             -> noop + WARN

skills:
  原生 skill 支持      -> 写 skill 包
  不支持 skill 但支持 rules -> 写 model_decision rule (description 必填)
  都不支持             -> 跳过 + WARN

references:
  原生 import 支持     -> @path 引用
  不支持               -> 跳过（不污染主入口）
```

具体 target 表：

| 目标 | 降级类 | 策略 |
|---|---|---|
| codex | commands | 降级为 skill 写到 `.agents/skills/cmd-<n>/SKILL.md`（description 含 slash 调用 hint） |
| aider | skills / commands | noop；aider 不支持项目级扩展，AGENTS.md 只承载 shared rules |
| continue-dev | skills | 写到 `.continue/rules/skill-<n>.md`（model_decision rule） |
| antigravity | skills | 写到 `.agents/rules/skill-<n>.md`（model_decision rule） |
| gemini | skills | 跳过（无原生 skill） |
| copilot / opencode | skills 含 SkillFiles | 单文件 agent.md 不支持子目录；SkillFiles 被忽略，runner 收集 WARN 输出到 stderr |

### 4.6 Footer 注入

当 `inject = true`，文件首部注入：

```html
<!-- Generated by stdagent vX.Y.Z. Source: ... -->
```

文件末尾：`<!-- /Generated by stdagent -->`。

当 `inject_whatis = true` 附加 "About This File" 段。

JSON 文件（`.mcp.json` / `.cursor/mcp.json` / `.vscode/mcp.json`）不写 markdown marker，
drift 检测靠 sha256 checksum。

TOML 文件（`.gemini/commands/*.toml`）用 `# Generated by stdagent ...` 注释 marker。

### 4.7 备份与冲突

每次 sync 前 `writer.Backup.Snapshot` 把将被覆盖的文件复制到 `.stdai/backups/<RFC3339-utc>/`。
保留份数由 `backup_keep` 控制（默认 5）。

冲突检测：

1. 检测目标文件首尾的 stdagent marker
2. 不存在 marker（用户手写）：备份后写入 + stderr WARN
3. 存在 marker：直接覆盖，不输出 WARN

### 4.8 monorepo

`stdagent <command>` 不显式 `--config` 时，从 cwd 向上 walk 找最近的 `.stdai/config.toml`，
返回 `(cfgPath, projectRoot)`，所有路径操作基于 projectRoot。

子目录任意位置跑 stdagent 都能定位到 monorepo 项目根。

### 4.9 自我升级

`stdagent upgrade` 流程：

1. 调 `https://api.github.com/repos/StringKe/std-agent/releases/latest` 拿 tag
2. 比对当前 `versionStr`，已最新且非 `--force` 时跳过
3. 下载平台对应归档：`std-agent_<ver>_<os>_<arch>.<tar.gz|zip>`
4. 下载 `checksums.txt` 校验 sha256
5. 解包提取 `stdagent` binary
6. `inconshreveable/go-update` 原子替换当前 executable（含 Windows 运行中 .exe 处理）

`--version vX.Y.Z` 可指定目标版本（含降级）；`GITHUB_TOKEN` 环境变量自动用于 GH API（提升 rate limit）。

### 4.10 LLM 上下文 Budget 提醒

`internal/budget` 包对每个 std 源文件做字符上限检查，runner 在 parse 后
把 SOFT / HARD WARN 收集到 `res.Warnings`（前缀 `[budget]`），cli 输出到
stderr。

| 类型 | 软上限 | 硬上限 | 说明 |
|---|---|---|---|
| rule（任意 target） | 8000 字符 | - | 超过会显著消耗 LLM context；建议拆分或转 skill |
| skill（任意 target） | 20000 字符 | - | 超 agentskills 5000 tokens 软上限；建议拆到 references/ |
| command（任意 target） | 4000 字符 | - | 通用 context 软建议 |
| codex AGENTS.md 总字节 | - | 32768 | WARN，不自动 spill |
| cursor rule | 80000 | 100000 | 硬上限触发截断 |
| windsurf rule | - | 12000 | 单文件硬上限 |
| antigravity rule | - | 12000 | 单文件硬上限 |
| windsurf global_rules.md | - | 6000 | 全局文件硬上限 |

输出形态（stderr）：

```
[warn] [budget] SOFT rules/huge.md [*] 9000 > 8000 (rule 单文件 > 8000 字符...)
[warn] [budget] HARD AGENTS.md [codex] total rules 40000 > 32768 (...)
```

`Document.BodyBytes` 字段在 parser.Parse 时填充；budget 包优先用此值，
否则 fallback 到 `len([]byte(d.Body))`。

字符与 token：限额检查使用最终字节数；`stdagent budget` 另提供基于字符规则的
token 近似，`--rendered` 报告实际 root 与 sidecar 体积。
已足够触发"context 较大"提醒。后续如需更精确估算可接 OpenAI tiktoken
等 tokenizer。

### 4.11 .stdaiignore

`gitignore` 风格 glob，控制源端哪些文件不参与 sync。

- 文件位置：项目根 `.stdaiignore`
- pattern 路径相对 `.stdai/standards/`（与远端 source 合并后的逻辑根）
- 支持 doublestar `**` / `*` / `?`
- 行首 `#` 注释，空行忽略
- 当前版本仅支持正向 ignore；白名单需走 `[targets].include` 或 frontmatter `targets`
- 被 ignore 的文件在 sync 输出 `[ignore] skipped N files via .stdaiignore` 警告

`stdagent init` 默认生成模板（含示例注释），用户按需取消注释。


## 附录: 文档索引

- [README.md](../README.md) 项目入口与快速开始
- [docs/prd.md](prd.md) 产品需求
- [docs/architecture.md](architecture.md) 模块划分与数据流
- [docs/commands.md](commands.md) CLI 命令规范
- [docs/conversion-rules.md](conversion-rules.md) 转换矩阵 + 字段映射
- [docs/file-structure.md](file-structure.md) 目录结构原则
- [docs/format-spec.md](format-spec.md) frontmatter 详细 schema（本 spec 的补充）
- [docs/roadmap.md](roadmap.md) 路线图
- [docs/targets/](targets/) 23 个目标工具调研

---

本 spec 版本 1.0；兼容 stdagent v0.1.0+。
