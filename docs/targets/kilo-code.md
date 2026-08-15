# Target: Kilo Code

调研日期: 2026-05-17（2026-07-10 重大更新：确认为 kilo.ai 新平台/opencode fork，
skills GA、commands 单数目录、kilo.jsonc instructions[] 闭环已实现，见下方全文修订）
官方仓库: https://github.com/Kilo-Org/kilocode
公司: Kilo Code Inc

## 1. 摘要

2026-07-10 核实：Kilo Code 官方站点为 kilo.ai，新平台**以 opencode 为底座 fork**（而非
2026-05-17 版本描述的"Cline 活跃 fork"）。CLI/新平台线与旧 VS Code 插件线（`.kilocode/`
路径）是否长期并存两套配置尚未验证，标 UNKNOWN（见 §9）。

配置形态最显著的特征：`.kilo/rules/` 目录下的 `.md` 文件**不会被自动扫描**，必须被项目
`kilo.jsonc` 的 `instructions[]` 数组**显式引用**才会加载（`instructions[]` 支持 glob 模式，
如 `.kilo/rules/*.md`）。同时为 Cline/Roo 老用户保留 `.kilocode/rules/` 路径做向后兼容。

项目根 `AGENTS.md` **恒加载**（不受 `instructions[]` 限制，与 `.kilo/rules/*.md` 需显式声明
形成对比）。

transformer（`internal/transformer/kilo_code.go`）已实现 kilo.jsonc 闭环：只要
`.kilo/rules/` 下有产物（rules 或 glossary），Plan 额外追加一个 `JSONMerge` FileOp 把
`{"instructions":[".kilo/rules/*.md"]}` 合并写入 `kilo.jsonc`（保留用户已有键，仅追加该
glob）。v0.0.4 文档中"v0.0.5 待办：kilo.jsonc 写入"已落地，不再是待办项。

协议归属：协议族 D（Clinerules）。与 cline / roo-code 共用 `Clinerules` Protocol，
adapter 字段差异表见 §7。

## 2. 配置文件路径

| 类型 | 路径 | 说明 |
|---|---|---|
| 项目级 rules（主） | `<repo>/.kilo/rules/` 目录 | 必须被 kilo.jsonc 的 `instructions[]` 显式引用才加载 |
| 项目级 rules（兼容） | `<repo>/.kilocode/rules/` 目录 | 向后兼容老（VS Code 插件线）路径 |
| mode-specific rules | `<repo>/.kilocode/rules-{mode}/` 目录 | 角色切换时加载（architect / code / debug 等） |
| skills | `<repo>/.kilo/skills/<name>/SKILL.md` | 原生 Agent Skills 标准包 |
| commands | `<repo>/.kilo/commands/<name>.md` | 官方文档复数目录；源码仍兼容单数 `.kilo/command/` |
| kilo.jsonc | `<repo>/kilo.jsonc` | 项目级配置，`instructions[]` 引用 rules 路径（支持 glob） |
| 全局配置 | `~/.config/kilo/kilo.jsonc` | 用户级配置 |
| 根文件（消费） | `<repo>/AGENTS.md` | 恒加载，不受 `instructions[]` 限制 |

## 3. 文件格式与 frontmatter

- 文件格式：纯 Markdown（`.md`），rule body 无强制 frontmatter
- frontmatter 方言：沿用 Cline 上游约定（`paths: ["src/**"]` YAML list 做 glob 激活），
  官方文档对 frontmatter 字段无强制规范
- 与 cline 的差异：kilo 用 `.kilo/rules/` 子目录组织（不用文件名数字前缀），
  rules 加载顺序由 kilo.jsonc `instructions[]` 数组顺序决定

## 4. skills / commands / subagents 原生支持

| std-agent 类型 | Kilo Code 原生 | std-agent 落点 |
|---|---|---|
| rules | YES（核心，需 kilo.jsonc `instructions[]` 引用） | `.kilo/rules/<name>.md` |
| commands | YES | `.kilo/commands/<name>.md`（2026-08-15 对齐官方文档复数） |
| skills | YES（原生 Agent Skills，2026-07-10 确认） | `.kilo/skills/<n>/SKILL.md` |
| references | NO | `.kilo/rules/references/<n>.md`（fallback） |
| subagents | mode-switching 间接实现，非 stdagent 形态 | `.kilo/rules/subagents/<n>.md`（fallback） |

## 5. 字节限制

无明确字节限制。社区实践沿用 Cline 约定（单 rule < 4KB，合计 < 20KB），
stdagent 不强制此限制。

## 6. stdagent 落点（2026-07-10 更新，与 `internal/transformer/kilo_code.go` 一致）

- RulesDir：`.kilo/rules`
- SkillsDir：`.kilo/skills`（原生 Agent Skills，已迁移）
- CommandsDir：`.kilo/command`（原生 slash commands，单数目录，已迁移）
- FallbackDir：`.kilo/rules`（references / subagents 自动加 subdir，仍无原生落点）
- SingleFileFallback：`""`（kilo 无单文件 fallback，与 cline `.clinerules` / roo `.roorules` 不同）
- 数字前缀：无（与 cline 区别，跟 roo 一致）
- glossary：`.kilo/rules/glossary.md`（frontmatter `std-agent-type: glossary`）

### kilo.jsonc 写入（已落地，不再是待办）

`KiloCode.Plan`（`internal/transformer/kilo_code.go`）在 `.kilo/rules/` 下有产物（rules 或
glossary）时，额外追加一个 `writer.FileOp{Path: "kilo.jsonc", JSONMerge: true}`，内容为
`{"instructions":[".kilo/rules/*.md"]}`。writer 侧 JSONMerge 合并逻辑保留用户既有 `kilo.jsonc`
键，仅追加/合并 `instructions` 数组元素；若用户 `kilo.jsonc` 已含注释（JSONC），writer 跳过
并 WARN，不破坏用户配置。旧版"v0.0.5 待办"描述已过时，闭环已实现。

## 7. 与 cline / roo-code 的关系

| 维度 | Cline | Roo Code | Kilo Code |
|---|---|---|---|
| 主目录 | `.clinerules/` | `.roo/rules/` | `.kilo/rules/` |
| 单文件 fallback | `.clinerules` | `.roorules` | 无 |
| 入口配置 | 无（自动加载） | 无（自动加载） | `kilo.jsonc` instructions[] 显式引用（支持 glob） |
| mode-specific | `.clinerules-{mode}` | `.roo/rules-{mode}/` + `.roorules-{mode}` | `.kilocode/rules-{mode}/` |
| 数字前缀 | 100/500/900 | 无（用子目录） | 无（用 instructions[] 顺序） |
| 全局位置 | 无 | `~/.roo/` | `~/.config/kilo/kilo.jsonc` |
| 根文件消费 | AGENTS.md（根目录，不读嵌套） | AGENTS.md（额外 context，不读嵌套） | AGENTS.md（恒加载，不受 instructions[] 限制） |

三者共用 `Clinerules` Protocol，差异完全靠 adapter 配置表达，无额外代码。

## 8. 信息来源

- https://github.com/Kilo-Org/kilocode
- https://kilo.ai/docs/customize/skills（2026-07-10 新增，skills GA 确认）
- https://kilo.ai/docs/customize/custom-instructions（2026-07-10 新增，`instructions[]` 显式
  引用与 glob 支持确认）
- 协议族对比：/tmp/std-agent-protocol-research.md §2.D（行 34, 113-121）

## 9. UNKNOWN

- Kilo VS Code 插件线（`.kilocode/`）与 CLI 新平台（`.kilo/`，kilo.ai/opencode 底座）是否
  长期并存两套配置，影响是否应拆成双 target
- Kilo Code 是否继承 opencode 底座的动态嵌套 AGENTS.md 解析能力（未核实）
- mode-specific rules 是否同时加载基础 rules（推断"叠加"，未实证）
