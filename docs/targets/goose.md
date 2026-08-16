# Target: Goose (AAIF)

调研日期: 2026-08-16
官方文档: https://goose-docs.ai/
官方仓库: https://github.com/aaif-goose/goose

## 1. 摘要

Goose 是 Linux Foundation Agentic AI Foundation 的开源本地 agent（原 Block goose）。
默认项目上下文文件是 `AGENTS.md`，其次是 `.goosehints`。stdagent 只写共享 `AGENTS.md`，不写 `.goosehints`，避免双份注入。

项目 skills 官方推荐 `.agents/skills/<name>/SKILL.md`。Goose 还兼容 `.goose/skills/` 与 `.claude/skills/`，但共享 `.agents/skills/` 必须与 Codex / Amp / Warp / Kimi / Antigravity 字节一致。

## 2. 配置路径

| 类别 | 路径 | 说明 |
|---|---|---|
| 项目上下文 | `AGENTS.md` | 默认；支持嵌套，每级先 AGENTS.md 再 .goosehints |
| 全局上下文 | `~/.config/goose/AGENTS.md` | `CONTEXT_FILE_NAMES` 可改 |
| 项目 skills | `.agents/skills/<name>/SKILL.md` | 官方推荐 |
| 兼容 skills | `.goose/skills/`、`.claude/skills/` | 向后兼容，stdagent 不写 |
| 全局 skills | `~/.agents/skills/` | 用户级 |
| 插件 skills | `~/.agents/plugins/<name>/` | 不由 stdagent 生成 |

## 3. std-agent 映射

| std-agent 类型 | Goose 落点 |
|---|---|
| rules | 全文 inline 到共享 `AGENTS.md`；嵌套 root 写 `<path>/AGENTS.md` |
| skills | `.agents/skills/<name>/SKILL.md`（共享路径） |
| commands | `.agents/skills/commands/<name>/SKILL.md`（无原生 commands） |
| references | `.goose/references/<name>.md` |
| subagents | `.goose/subagents/<name>.md`（无官方文件化 agent 目录） |

## 4. 来源

- https://goose-docs.ai/docs/guides/context-engineering/using-skills/
- https://goose-docs.ai/docs/guides/context-engineering/using-goosehints/
- https://github.com/aaif-goose/goose

## 5. UNKNOWN

- 无官方 AGENTS.md 字节硬限
- 无官方文件化 subagent schema
- `.goose/` 根目录除 `skills/` 外是否还有自动扫描
