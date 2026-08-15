# Target: Windsurf (Codeium / Cognition Devin Desktop)

调研日期: 2026-05-07（2026-07-10 补充调研；2026-08-15 复核 `.windsurf/{rules,skills,workflows}` 仍为官方主路径）
官方文档: https://docs.windsurf.com/ ; https://docs.devin.ai/desktop/devin-desktop-faq

## 1. 摘要

Windsurf 配置以 `.windsurf/` 子目录为核心，rules、workflows、skills 三件套并存。
Cascade Memories 是会话级软记忆。`.windsurfrules` legacy 文件存在 `.windsurf/rules/`
时被现代格式取代。

2026-06-02 起 Windsurf 已并入 Cognition 的 Devin 产品线，以 "Devin Desktop" 名称发行
下一代版本，构建在原 Windsurf IDE 基座上。规则目录首选顺序随之调整：新版本优先读取
`.devin/rules/`，旧的 `.windsurfrules` 与 `.windsurf/rules/` 仍继续读取（不强制迁移，
现有仓库无需改名即可继续工作）。workflows 与 skills 目录未受合并影响，仍是
`.windsurf/workflows/` 与 `.windsurf/skills/`。

transformer 已落地双写过渡：`.windsurf/rules/*.md` 产出的每个文件都原样镜像一份到
`.devin/rules/`，两处字节内容一致，避免用户在新旧版本客户端间切换时丢规则
（`internal/transformer/windsurf.go`）。

`AGENTS.md` 由同一 Rules 引擎处理：根级当 always-on rule，子目录中的 AGENTS.md
自动当作针对该目录的 glob rule。

## 2. 配置文件路径

| 类别 | 路径 | 说明 |
|---|---|---|
| 全局 rules | `~/.codeium/windsurf/memories/global_rules.md` | 单文件，所有 workspace 恒生效 |
| 工作区 rules（Devin Desktop 首选，2026-06+） | `.devin/rules/*.md` | 新版本优先读取，stdagent 与 `.windsurf/rules/` 双写镜像 |
| 工作区 rules（legacy，仍读） | `.windsurf/rules/*.md`（每文件一个 rule） | workspace、子目录、向上至 git root 全部扫描 |
| 系统 rules（Enterprise） | macOS `/Library/Application Support/Windsurf/rules/`、Linux/WSL `/etc/windsurf/rules/`、Windows `C:\ProgramData\Windsurf\rules\` | IT 部署，read-only |
| 工作区 workflows | `.windsurf/workflows/*.md` | 子目录 + 父目录至 git root 全部发现，Devin 合并未影响该目录 |
| 全局 workflows | `~/.codeium/windsurf/global_workflows/*.md` | 不入仓 |
| 工作区 skills | `.windsurf/skills/<name>/SKILL.md` + 附属文件 | 每 skill 一个目录，Devin 合并未影响该目录 |
| 全局 skills | `~/.codeium/windsurf/skills/<name>/SKILL.md` | 用户级 |
| 用户 MCP | `~/.codeium/windsurf/mcp_config.json` | 顶层键 `mcpServers` |
| Memories（自动） | `~/.codeium/windsurf/memories/` 下 workspace-specific 子结构 | 仅本机本工作区，不可分享 |
| AGENTS.md | 工作区任意目录 | Rules 引擎处理（root=always-on，子目录=glob） |
| `.windsurfrules`（legacy） | 项目根 | Wave 8 之前格式；存在 `.windsurf/rules/` 时被取代 |

## 3. 文件格式与 frontmatter

现代 rules / workflows / skills 都是 markdown + YAML frontmatter。Rules frontmatter 关键字段：

| 字段 | 类型 | 必填条件 |
|---|---|---|
| `trigger` | enum: `always_on` / `model_decision` / `glob` / `manual` | 必填 |
| `globs` | string[] | `trigger=glob` 时必填 |
| `description` | string | `trigger=model_decision` 时必填 |

`global_rules.md` 与根级 `AGENTS.md` **不使用 frontmatter**，恒 always-on。

`.windsurfrules` 是纯文本，常用编号列表，全部 always-on，无 frontmatter。

`.devin/rules/*.md` 沿用与 `.windsurf/rules/*.md` 相同的 frontmatter 方言（同一份 Rules
引擎消费两个目录），字段定义不变。

## 4. 4 种触发模式

| 模式 | trigger | 行为 | 上下文成本 |
|---|---|---|---|
| Always On | `always_on` | 全文进 system prompt | 每条消息 |
| Model Decision | `model_decision` | system prompt 只放 description，按需读全文 | description 恒在 |
| Glob | `glob` | 仅当 Cascade 读/写匹配 globs 的文件时拉入 | 命中文件才挂载 |
| Manual | `manual` | 不进 system prompt，靠 `@rule-name` 触发 | 仅 @mention |

## 5. 字符上限

| 范围 | 路径 | 上限 |
|---|---|---|
| 全局 rules | `global_rules.md` | 6000 字符 |
| 工作区 rules（单文件） | `.windsurf/rules/*.md` 每文件 | 12000 字符 |

`.devin/rules/*.md` 是 `.windsurf/rules/*.md` 的字节级镜像，transformer 按同一
12000 字符上限计算，未见官方为 `.devin/rules/` 单独公布上限。

## 6. MCP 配置

| 类别 | 路径 | 说明 |
|---|---|---|
| 用户级 | `~/.codeium/windsurf/mcp_config.json` | 顶层键 `mcpServers`（不是 `servers`） |
| 项目级 | INSUFFICIENT-EVIDENCE | 官方文档未给出明确路径；候选 `.windsurf/mcp.json` / `.windsurf/mcp_config.json` 均无背书 |

支持 stdio 与 HTTP 两类：

```jsonc
{
  "mcpServers": {
    "stdio-server": {
      "command": "/abs/path/bin",
      "args": ["--flag"],
      "env": { "K": "V" }
    },
    "remote-server": {
      "serverUrl": "https://...",
      "headers": { "Authorization": "Bearer ${env:TOKEN}" }
    }
  }
}
```

变量插值：`${env:VAR}`、`${file:/path}`，作用于 `command / args / env / serverUrl / url / headers`。

Cascade 同时可见 100 个工具上限。

## 7. 加载机制

- 嵌套：`.windsurf/rules` 与 `.windsurf/workflows` 在当前 workspace、子目录、向上至 git root
  全部被发现；多 workspace 时按最短相对路径去重
- `.devin/rules` 遵循同一发现规则（当前实测行为，Devin Desktop 版本以 `.devin/rules/`
  为首选查找目录，找不到时退回 `.windsurf/rules/`）
- Workflows 优先级：System > Workspace > Global > Built-in（同名时高层覆盖）
- `.windsurfrules` 与 `.windsurf/rules/` 共存时现代格式优先，legacy 被忽略

## 8. std-agent 四类映射

| std-agent 类型 | Windsurf 落点 |
|---|---|
| rules（无 applyTo） | `.windsurf/rules/<name>.md` + 镜像 `.devin/rules/<name>.md`，`trigger: always_on` |
| rules（有 applyTo） | `.windsurf/rules/<name>.md` + 镜像 `.devin/rules/<name>.md`，`trigger: glob`，`globs:` 来自 std `applyTo` |
| skills | `.windsurf/skills/<name>/SKILL.md` + 附属 |
| commands | `.windsurf/workflows/<name>.md`（slash 触发 `/<name>`） |
| references | 不主动写入；推荐作为 rules 段落 |

## 9. 转换器实现要点

1. 每条 std rule 输出独立 `.windsurf/rules/<name>.md`，按 std frontmatter 决定 `trigger`：
   - `alwaysApply: true` -> `trigger: always_on`
   - `applyTo` 非空 -> `trigger: glob` + `globs: <list>`
   - 否则 -> `trigger: model_decision` + `description` 必填
2. 单文件超过 12000 字符 -> 拆分为多个，name 加后缀 `-part1` `-part2`
3. `.windsurfrules` 不再生成（推送 modern 格式）
4. workflows 输出 `.windsurf/workflows/<name>.md`，文件名即 slash 名
5. skills 输出 `.windsurf/skills/<name>/SKILL.md`
6. Windsurf 自动消费其他启用 producer 经 runner canonicalize 的共享 AGENTS.md
7. v1.0 不写 MCP（用户级 path 已知，项目级未知，等 v1.1 项目级路径确认后落地）
8. Devin 双写：`Windsurf.Plan` 在 `WindsurfStyle` 协议产出的 Plan 基础上，把所有
   `.windsurf/rules/` 前缀的 `FileOp` 复制一份，路径改写为 `.devin/rules/` 前缀，
   内容字节不变（`internal/transformer/windsurf.go`）。workflows / skills 不做镜像，
   因为 Devin 合并只影响 rules 目录优先级。
9. 双写是过渡方案：待 Devin Desktop 生态稳定、`.windsurf/rules/` 读取占比下降后，
   再评估是否只写 `.devin/rules/`，或拆出独立 devin target；当前维持双写零成本兼容。

## 10. 信息来源

- https://docs.windsurf.com/windsurf/cascade/memories
- https://docs.windsurf.com/windsurf/cascade/mcp
- https://docs.windsurf.com/windsurf/skills
- https://docs.windsurf.com/windsurf/workflows
- https://windsurf.com/editor/directory
- https://github.com/Windsurf-Samples/cascade-customizations-catalog
- https://devin.ai/blog/windsurf-is-now-devin-desktop/（Windsurf 并入 Devin Desktop 公告）
- https://docs.devin.ai/desktop/devin-desktop-faq（`.devin/rules/` 优先、`.windsurf/rules/` 向后兼容）

## 11. 已确认与剩余 UNKNOWN

已确认：
- 用户级 MCP 路径 `~/.codeium/windsurf/mcp_config.json`，顶层键 `mcpServers`
- MCP schema 含 stdio (command/args/env) 与 HTTP (serverUrl/url/headers)
- 变量插值 `${env:VAR}` `${file:/path}`
- Cascade 同时可见 100 个工具上限
- Windsurf 已并入 Cognition Devin Desktop（2026-06-02 起），`.devin/rules/` 为新版本
  首选规则目录，`.windsurfrules` 与 `.windsurf/rules/` 仍向后兼容读取，无需强制迁移
- workflows / skills 目录不受 Devin 合并影响

剩余 UNKNOWN：
- 项目级 MCP 配置路径
- skills `SKILL.md` frontmatter 完整字段
- 是否有 hooks 机制
- Cascade Memories 自动捕获策略
- `.devin/rules/*.md` 单文件是否有独立于 12000 字符的官方上限（未见公开文档单独说明）
