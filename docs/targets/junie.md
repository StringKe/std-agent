# Target: Junie (JetBrains)

调研日期: 2026-08-21
官方文档: https://junie.jetbrains.com/docs/guidelines-and-memory.html ；https://junie.jetbrains.com/docs/agent-skills.html
官方仓库: https://github.com/JetBrains/junie-guidelines

## 1. 摘要

Junie 是 JetBrains 的编码 agent（IDE + CLI，2026-06 GA）。CLI 发现 guidelines 的顺序：

1. `.junie/AGENTS.md`
2. 否则根 `AGENTS.md`，并合并 `.junie/playbook.md` 与 `.junie/rules/*.md`
3. 遗留 `.junie/guidelines.md` 或 `.junie/guidelines/`

stdagent 写共享根 `AGENTS.md`，不复制 `.junie/AGENTS.md`，避免双份注入。non-root rules 写 `.junie/rules/`。skills 原生 `.junie/skills/<name>/SKILL.md`。

Junie CLI 可提示把其他 agent 的 skills 导入 `.junie/skills/`，但不会把 `.agents/skills/` 当作原生扫描根。

## 2. std-agent 映射

| std-agent 类型 | 落点 |
|---|---|
| rules | root 写共享 `AGENTS.md`；其余写 `.junie/rules/<name>.md` |
| skills | `.junie/skills/<name>/SKILL.md`（官方 frontmatter：`name` + `description`） |
| commands | `.junie/commands/<name>.md`（无官方 slash 命令目录文档） |
| references | `.junie/references/<name>.md` |
| subagents | `.junie/subagents/<name>.md` |

## 3. UNKNOWN

- 嵌套 `AGENTS.md` 是否自动加载（adapter `NestedSupported=false`）
- 官方 commands / subagents 文件路径
- `.junie/rules/*.md` 的 frontmatter 方言
