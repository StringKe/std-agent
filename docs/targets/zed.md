# Target: Zed

调研日期: 2026-08-21
官方文档: https://zed.dev/docs/ai/instructions ；https://zed.dev/docs/ai/skills
官方仓库: https://github.com/zed-industries/zed

## 1. 摘要

Zed Agent 以 `AGENTS.md` 为项目指令主文件，项目 skills 只扫描 `.agents/skills/<name>/SKILL.md`。skills 必须扁平，子目录不会被发现。

项目指令按存在的第一个文件生效：`.rules`、`.cursorrules`、`.windsurfrules`、`.clinerules`、`.github/copilot-instructions.md`、`AGENT.md`、`AGENTS.md`、`CLAUDE.md`、`GEMINI.md`。stdagent 只写共享 `AGENTS.md`；多 target 同时开启时，若仓库里已有更靠前的兼容文件，Zed 可能不读 `AGENTS.md`。

## 2. std-agent 映射

| std-agent 类型 | 落点 |
|---|---|
| rules | 全文 inline 到共享 `AGENTS.md`；嵌套 root 写 `<path>/AGENTS.md` |
| skills | `.agents/skills/<name>/SKILL.md`（共享路径，字段集与 Codex 相同） |
| commands | `.zed/commands/<name>.md`（不进 `.agents/skills/commands/`，避免违反扁平 skills） |
| references | `.zed/references/<name>.md` |
| subagents | `.zed/subagents/<name>.md` |

官方 skill frontmatter：`name`、`description`、`disable-model-invocation`。共享 skills 路径不写后者，以免与 Codex 字节不一致。description 软上限 1024 字节；全部 skill 的 name+description 合计 50KB。

## 3. UNKNOWN

- Zed 是否自动发现 `.zed/` fallback 目录
- 官方是否有 AGENTS.md 字节硬限
