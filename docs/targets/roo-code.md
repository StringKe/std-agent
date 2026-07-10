# Target: Roo Code

调研日期: 2026-05-17（2026-07-10 更新：skills GA 已迁移、原生 commands 目录、嵌套/单文件优先级回填）
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
| skills | `<repo>/.roo/skills/<name>/SKILL.md` | 2026-05-15 GA 的 Agent Skills 标准包，官方扫描路径 |
| commands（原生） | `<repo>/.roo/commands/<name>.md` | 原生 slash commands，frontmatter `description` / `argument-hint` / `mode` |

## 3. 文件格式与 frontmatter

- 文件格式：纯 Markdown（`.md`）。社区主流 roo 项目**不写 frontmatter**，
  与 cline 一样直接读 body 全文
- frontmatter 方言：Cline 上游约定 `paths: ["src/**"]`（YAML list）做 glob 条件激活，
  Roo Code 沿用此约定，但官方文档对 frontmatter 字段无强制规范
- 与 cline 的关键差异：cline 用文件名数字前缀（`100-` / `500-` / `900-`）暗示加载优先级；
  roo-code 用**子目录组织**（`.roo/rules/security/` / `.roo/rules/style/`）替代数字前缀
- **`.roo/rules/` 内部不递归扫描子目录**：只读顶层 `.md` 文件，`.roo/rules/skills/`、
  `.roo/rules/workflows/` 等降级子目录完全不被读取（2026-07-10 确认，
  https://github.com/RooCodeInc/Roo-Code/pull/10446）。skills / commands 必须走各自原生目录，
  不能再靠 `.roo/rules/` 子目录降级
- `.roo/rules/` 目录存在时，单文件 fallback `.roorules` 被忽略（目录优先，2026-07-10 确认，
  解决 §9 旧 UNKNOWN）

## 4. skills / commands / subagents 原生支持

| std-agent 类型 | Roo Code 原生 | std-agent 落点 |
|---|---|---|
| rules | YES（核心） | `.roo/rules/<name>.md` |
| commands | YES（原生 slash commands，2026-07-10 确认） | `.roo/commands/<name>.md`，frontmatter `description` / `argument-hint` / `mode` |
| skills | YES（2026-05-15 GA，2026-07-10 确认） | `.roo/skills/<n>/SKILL.md`（Agent Skills 标准包） |
| references | NO | `.roo/rules/references/<n>.md`（fallback） |
| subagents | 通过 mode-switching 间接实现，但非 stdagent 形态 | `.roo/rules/subagents/<n>.md`（fallback） |

transformer（`internal/transformer/roo_code.go` `rooCodeAdapter`）已迁移到原生落点：
`SkillsDir=".roo/skills"`、`CommandsDir=".roo/commands"`；`buildWorkflow` 目前只写
`description` / `argument-hint` 两个 frontmatter 字段，官方 `mode` 字段尚未接入（跟进项，非本轮范围）。

## 5. 字节限制

无明确字节限制。社区实践：单 rule 文件 < 4KB，全部 rules 合计 < 20KB
（与 cline 一致，超出会占用 context window）。stdagent v0.0.4 不强制此限制。

## 6. stdagent 落点（2026-07-10 更新，与 `internal/transformer/roo_code.go` 一致）

- RulesDir：`.roo/rules`（不递归子目录，见 §3）
- SkillsDir：`.roo/skills`（原生 Agent Skills 标准包，已迁移，不再走 `.roo/rules/skills/` fallback）
- CommandsDir：`.roo/commands`（原生 slash commands，已迁移，不再走 `.roo/rules/workflows/`）
- FallbackDir：`.roo/rules`（references / subagents 仍无原生落点，自动加 subdir 降级）
- SingleFileFallback：`.roorules`（`.roo/rules/` 目录存在时被忽略，见 §3）
- 数字前缀：无（与 cline 区别）
- glossary：`.roo/rules/glossary.md`（frontmatter `std-agent-type: glossary`）

## 6.1 嵌套 AGENTS.md（2026-07-10 新增）

Roo Code 支持读取子目录 AGENTS.md，但受 `enableSubfolderRules` 设置控制，**默认关闭**。
用户需在 Roo Code 设置中手动开启才会向下扫描子目录。stdagent 当前 `rooCodeAdapter` 未设置
`NestedSupported`（零值 false），与"默认关闭"的官方行为一致，无需改动。

## 7. 与 cline / kilo-code 的关系

| 维度 | Cline | Roo Code | Kilo Code |
|---|---|---|---|
| 主目录 | `.clinerules/` | `.roo/rules/` | `.kilo/rules/` |
| 单文件 fallback | `.clinerules` | `.roorules`（目录存在时被忽略） | 无 |
| mode-specific | `.clinerules-{mode}` | `.roo/rules-{mode}/` + `.roorules-{mode}` | `.kilocode/rules-{mode}/` |
| 数字前缀 | 100/500/900 | 无（用子目录） | 无 |
| 全局位置 | 无 | `~/.roo/` | `~/.config/kilo/kilo.jsonc` |
| 读根 AGENTS.md | 是（根目录，不读嵌套） | 是（额外 context，不读嵌套子目录，见 §6.1） | 是（恒加载，见 kilo-code.md） |

三者共用 `Clinerules` Protocol，差异完全靠 adapter 配置表达，无额外代码。

## 8. 信息来源

- https://github.com/RooCodeInc/Roo-Code
- https://docs.roocode.com/features/custom-instructions（rules / mode-specific 文档）
- https://docs.roocode.com/features/skills（2026-07-10 新增，skills GA 确认）
- https://github.com/RooCodeInc/Roo-Code/pull/10446（2026-07-10 新增，`.roo/rules/` 不递归子目录、
  单文件 fallback 被忽略的确认）
- 协议族对比：/tmp/std-agent-protocol-research.md §2.D

## 9. UNKNOWN

- mode-specific rules 是否同时加载基础 rules（推断"叠加"，未实证）
- AGENTS.md 在 Roo Code 中的合并顺序与权重
- 原生 commands frontmatter `mode` 字段的具体取值与行为细节，transformer 尚未接入
  （已知字段存在，具体语义未核实）
