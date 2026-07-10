# Target: Roo Code

调研日期: 2026-05-17
官方仓库: https://github.com/RooCodeInc/Roo-Code
公司: Roo Code Inc

## 1. 摘要

Roo Code 是 VS Code AI 编码扩展，源自 Cline 的 fork（早期名为 Roo Cline），
2026 年第一季度独立品牌运营，GitHub stars 18k+。功能上接近 Cline 的超集，
增加了 mode-specific rules（架构师 / 工程师 / 测试等多角色切换）与 workflows。

协议归属：协议族 D（Clinerules）。与 cline / kilo-code 共用 `Clinerules` Protocol，
仅 `RulesDir` / `SingleFileFallback` 等 adapter 字段不同。

## 2. 配置文件路径

| 类型 | 路径 | 说明 |
|---|---|---|
| 项目级 rules（主） | `<repo>/.roo/rules/` 目录 | 子目录形式，加载所有 `.md` 文件 |
| 项目级 rules（fallback） | `<repo>/.roorules` | 单文件，向后兼容老项目；目录存在时优先目录 |
| mode-specific rules | `<repo>/.roo/rules-{mode}/` 目录 | 仅在切换到对应 mode 时加载 |
| mode-specific 单文件 | `<repo>/.roorules-{mode}` | mode 单文件 fallback |
| workflows | `<repo>/.roo/rules/workflows/` 目录 | 自定义命令，`/<name>` 触发 |
| 全局 rules | `~/.roo/` 目录 | OS 级全局规则 |
| AGENTS.md | 同时读取项目根 `AGENTS.md` | 作为额外 context 注入（非 first-class） |

## 3. 文件格式与 frontmatter

- 文件格式：纯 Markdown（`.md`）。社区主流 roo 项目**不写 frontmatter**，
  与 cline 一样直接读 body 全文
- frontmatter 方言：Cline 上游约定 `paths: ["src/**"]`（YAML list）做 glob 条件激活，
  Roo Code 沿用此约定，但官方文档对 frontmatter 字段无强制规范
- 与 cline 的关键差异：cline 用文件名数字前缀（`100-` / `500-` / `900-`）暗示加载优先级；
  roo-code 用**子目录组织**（`.roo/rules/security/` / `.roo/rules/style/`）替代数字前缀

## 4. skills / commands / subagents 原生支持

| std-agent 类型 | Roo Code 原生 | std-agent 落点 |
|---|---|---|
| rules | YES（核心） | `.roo/rules/<name>.md` |
| commands | 部分（workflows） | `.roo/rules/workflows/<name>.md` |
| skills | NO（无 Agent Skills 协议） | `.roo/rules/skills/<n>/SKILL.md`（fallback，含 explainer 注释） |
| references | NO | `.roo/rules/references/<n>.md`（fallback） |
| subagents | 通过 mode-switching 间接实现，但非 stdagent 形态 | `.roo/rules/subagents/<n>.md`（fallback） |

workflows 类似 commands 但需用户输入 `/<name>` 触发，不自动执行。

## 5. 字节限制

无明确字节限制。社区实践：单 rule 文件 < 4KB，全部 rules 合计 < 20KB
（与 cline 一致，超出会占用 context window）。stdagent v0.0.4 不强制此限制。

## 6. stdagent 落点

- RulesDir：`.roo/rules`
- CommandsDir：`.roo/rules/workflows`
- FallbackDir：`.roo/rules`（skills/references/subagents 自动加 subdir）
- SingleFileFallback：`.roorules`（保留，v0.0.4 默认走目录形式）
- 数字前缀：无（与 cline 区别）
- glossary：`.roo/rules/glossary.md`（frontmatter `std-agent-type: glossary`）

## 7. 与 cline / kilo-code 的关系

| 维度 | Cline | Roo Code | Kilo Code |
|---|---|---|---|
| 主目录 | `.clinerules/` | `.roo/rules/` | `.kilo/rules/` |
| 单文件 fallback | `.clinerules` | `.roorules` | 无 |
| mode-specific | `.clinerules-{mode}` | `.roo/rules-{mode}/` + `.roorules-{mode}` | `.kilo/rules-{mode}/` |
| 数字前缀 | 100/500/900 | 无（用子目录） | 无 |
| 全局位置 | 无 | `~/.roo/` | `~/.config/kilo/kilo.jsonc` |
| 读 AGENTS.md | 否 | 是 | 否 |

三者共用 `Clinerules` Protocol，差异完全靠 adapter 配置表达，无额外代码。

## 8. 信息来源

- https://github.com/RooCodeInc/Roo-Code
- https://docs.roocode.com/features/custom-instructions（rules / mode-specific 文档）
- 协议族对比：/tmp/std-agent-protocol-research.md §2.D

## 9. UNKNOWN

- `.roo/rules/` 与 `.roorules` 同时存在时的优先级（社区推断为目录优先，未在官方 doc 明确）
- mode-specific rules 是否同时加载基础 rules（推断"叠加"，未实证）
- AGENTS.md 在 Roo Code 中的合并顺序与权重
