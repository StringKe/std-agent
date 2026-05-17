# Target: Kilo Code

调研日期: 2026-05-17
官方仓库: https://github.com/Kilo-Org/kilocode
公司: Kilo Code Inc

## 1. 摘要

Kilo Code 是 VS Code AI 编码扩展，源自 Cline 的活跃 fork（与 Roo Code 并列为 Cline
生态两大下游 fork）。配置形态最显著的特征：在 `.kilo/rules/` 目录之上加了一层
`kilo.jsonc` 配置文件，由 `instructions[]` 数组**显式引用**子文件，而非自动加载所有
`.md`。同时为 Cline 老用户保留 `.kilocode/rules/` 路径做向后兼容。

协议归属：协议族 D（Clinerules）。与 cline / roo-code 共用 `Clinerules` Protocol，
adapter 字段差异表见 §7。

## 2. 配置文件路径

| 类型 | 路径 | 说明 |
|---|---|---|
| 项目级 rules（主） | `<repo>/.kilo/rules/` 目录 | 由 kilo.jsonc 的 instructions[] 引用 |
| 项目级 rules（兼容） | `<repo>/.kilocode/rules/` 目录 | 向后兼容老路径 |
| mode-specific rules | `<repo>/.kilocode/rules-{mode}/` 目录 | 角色切换时加载（architect / code / debug 等） |
| workflows | `<repo>/.kilo/rules/workflows/` 目录 | 自定义命令，`/<name>` 触发 |
| kilo.jsonc | `<repo>/kilo.jsonc` | 项目级配置，含 `instructions[]` 引用 rules 路径 |
| 全局配置 | `~/.config/kilo/kilo.jsonc` | 用户级配置 |

## 3. 文件格式与 frontmatter

- 文件格式：纯 Markdown（`.md`），rule body 无强制 frontmatter
- frontmatter 方言：沿用 Cline 上游约定（`paths: ["src/**"]` YAML list 做 glob 激活），
  官方文档对 frontmatter 字段无强制规范
- 与 cline 的差异：kilo 用 `.kilo/rules/` 子目录组织（不用文件名数字前缀），
  rules 加载顺序由 kilo.jsonc `instructions[]` 数组顺序决定

## 4. skills / commands / subagents 原生支持

| std-ai 类型 | Kilo Code 原生 | std-ai 落点 |
|---|---|---|
| rules | YES（核心） | `.kilo/rules/<name>.md` |
| commands | 部分（workflows） | `.kilo/rules/workflows/<name>.md` |
| skills | NO | `.kilo/rules/skills/<n>/SKILL.md`（fallback，含 explainer 注释） |
| references | NO | `.kilo/rules/references/<n>.md`（fallback） |
| subagents | mode-switching 间接实现，非 stdagent 形态 | `.kilo/rules/subagents/<n>.md`（fallback） |

## 5. 字节限制

无明确字节限制。社区实践沿用 Cline 约定（单 rule < 4KB，合计 < 20KB），
stdagent v0.0.4 不强制此限制。

## 6. stdagent 落点（v0.0.4）

- RulesDir：`.kilo/rules`
- CommandsDir：`.kilo/rules/workflows`
- FallbackDir：`.kilo/rules`（skills / references / subagents 自动加 subdir）
- SingleFileFallback：`""`（kilo 无单文件 fallback，与 cline `.clinerules` / roo `.roorules` 不同）
- 数字前缀：无（与 cline 区别，跟 roo 一致）
- glossary：`.kilo/rules/glossary.md`（frontmatter `std-ai-type: glossary`）

### v0.0.5 待办：kilo.jsonc 写入

Kilo Code 真正读取的入口是 `kilo.jsonc` 的 `instructions[]` 数组，而非自动扫描 `.kilo/rules/`。
v0.0.4 暂未实现 `kilo.jsonc` 的写入，用户需手动添加 `instructions[]` 条目引用
stdagent 输出的 rule 文件。后续版本将通过 `AdditionalConfigWriter` hook 完成
kilo.jsonc 的合并写入（保留用户既有键，仅维护 stdagent 注入段）。

## 7. 与 cline / roo-code 的关系

| 维度 | Cline | Roo Code | Kilo Code |
|---|---|---|---|
| 主目录 | `.clinerules/` | `.roo/rules/` | `.kilo/rules/` |
| 单文件 fallback | `.clinerules` | `.roorules` | 无 |
| 入口配置 | 无（自动加载） | 无（自动加载） | `kilo.jsonc` instructions[] 显式引用 |
| mode-specific | `.clinerules-{mode}` | `.roo/rules-{mode}/` + `.roorules-{mode}` | `.kilocode/rules-{mode}/` |
| 数字前缀 | 100/500/900 | 无（用子目录） | 无（用 instructions[] 顺序） |
| 全局位置 | 无 | `~/.roo/` | `~/.config/kilo/kilo.jsonc` |

三者共用 `Clinerules` Protocol，差异完全靠 adapter 配置表达，无额外代码。

## 8. 信息来源

- https://github.com/Kilo-Org/kilocode
- 协议族对比：/tmp/std-ai-protocol-research.md §2.D（行 34, 113-121）

## 9. UNKNOWN

- kilo.jsonc 的 `instructions[]` 是否支持 glob 模式（如 `.kilo/rules/*.md`）尚未在官方文档验证
- `.kilo/rules/` 与 `.kilocode/rules/` 同时存在时的优先级（推断为 `.kilo` 新路径优先）
- mode-specific rules 是否同时加载基础 rules（推断"叠加"，未实证）
