# Target: Qwen Code (Alibaba)

调研日期: 2026-05-17，2026-07-10 复核更新
官方仓库: https://github.com/QwenLM/qwen-code
官方主页: https://github.com/QwenLM/qwen-code

## 1. 摘要

Qwen Code 是阿里巴巴通义实验室（QwenLM）维护的 CLI Agent，基于 Google
Gemini CLI fork 而成，针对 Qwen 大模型生态做了适配。22k stars，是
中国大陆开发者覆盖率最高的开源 AI 编码 CLI 之一（Qwen API 国内可用性
稳定，无需翻墙）。

配置体系延续 Gemini CLI 的设计：根目录上下文文件 `QWEN.md`（命名沿
gemini-cli 的 `GEMINI.md` 模板）。2026-07 复核修正：`AGENTS.md` **不是**
仅当 `QWEN.md` 缺失时才读的 fallback，而是与 `QWEN.md` **叠加共存**
（两者都读并合并注入 context，源码 `loadHierarchicalMemory` 同时加载）。
自定义命令落 `.qwen/commands/`，与 gemini-cli 的关键差异：commands 用
普通 markdown 而非 TOML，更接近 codex / claude-code 风格。

2026-07 复核新增两项 P0 修复：
- Skills 已 GA，原生落点 `.qwen/skills/<name>/SKILL.md`（旧文档"降级"已过时）
- `.qwen/rules/` 是官方原生 RulesDir（源码 `loadRules`，支持 frontmatter `paths:`
  条件规则），旧版把它当纯降级 FallbackDir 是错误的；nonRoot rules 现已从
  inline 改为写独立文件 `.qwen/rules/<n>.md`
来源：https://github.com/QwenLM/qwen-code/blob/main/docs/users/features/memory.md
、https://qwenlm.github.io/qwen-code-docs/en/users/features/skills/

## 2. 配置文件路径

| 类别 | 路径 | 说明 |
|---|---|---|
| 全局上下文 | `~/.qwen/QWEN.md` | 用户级，自动加载 |
| 项目上下文 | `<repo>/QWEN.md` | 项目级，优先级最高 |
| 跨工具叠加 | `<repo>/AGENTS.md` | 与 `QWEN.md` 同时加载并合并，非"缺失时才读"的 fallback |
| 嵌套上下文 | `<repo>/<subdir>/QWEN.md` | 子目录限定；**仅 cwd 向上链发现，官方源码无向下预扫描**（与 gemini-cli 的 JIT 主动扫描不同，需 cwd 实际位于该子目录才生效） |
| 项目原生 rules | `<repo>/.qwen/rules/*.md` | 官方 `loadRules` 原生目录，frontmatter 支持 `paths:` 条件规则（2026-07 修正，非降级） |
| 项目原生 skills | `<repo>/.qwen/skills/<name>/SKILL.md` | Agent Skills 标准，frontmatter 另支持 `priority` 字段（transformer 暂未渲染） |
| 用户 settings | `~/.qwen/settings.json` | UNKNOWN（与 gemini-cli 同源） |
| 项目自定义命令 | `<repo>/.qwen/commands/*.md` | slash 命令 markdown |

## 3. 文件格式

| 文件 | 扩展 | frontmatter |
|---|---|---|
| `QWEN.md` | `.md` | 无 frontmatter，纯指令文本 |
| `AGENTS.md`（并行叠加） | `.md` | 无 frontmatter |
| `.qwen/rules/*.md` | `.md` | `paths`（可选，条件生效范围） |
| `.qwen/skills/<name>/SKILL.md` | `.md` | Agent Skills 标准字段 + 官方扩展 `priority`（可选） |
| `.qwen/commands/*.md` | `.md` | `description` / `argument-hint` / `tools` / `model`（与 codex / claude-code 对齐） |

与 gemini-cli 的格式差异：

- gemini-cli commands 用 TOML（`description` + `prompt` 双字段）；qwen-code commands 用 markdown
- gemini-cli 无原生 RulesDir（全 inline）；qwen-code 有原生 `.qwen/rules/`

## 4. std-agent 五类映射（实际实现，`internal/transformer/qwen_code.go`）

| std-agent 类型 | Qwen Code 落点 | 加载方式 |
|---|---|---|
| rules（root） | 项目根 `QWEN.md` | 自动加载，无 frontmatter |
| rules（nonRoot） | `.qwen/rules/<name>.md`（原生目录，非降级，2026-07 修正） | 自动扫描，frontmatter `paths` 条件生效 |
| skills | `.qwen/skills/<name>/SKILL.md`（原生 Agent Skills 标准包，2026-07 修正为原生非降级） | AI 按 description 触发 |
| commands | `.qwen/commands/<name>.md`，frontmatter 描述参数 | `/` 触发 |
| references | fallback `.qwen/rules/references/<name>.md` | std-agent 降级 |
| subagents | `.qwen/agents/<name>.md` | 官方原生（https://qwenlm.github.io/qwen-code-docs/en/users/features/sub-agents/） |

嵌套 root（源文档带 `NestedPath`）写到对应子目录的 `QWEN.md`，不带 manifest
（与 codex / gemini-cli 一致）；但 Qwen Code 本身只做 cwd 向上链发现，写入的
嵌套 `QWEN.md` 只有当用户实际工作目录位于该子目录时才会被读到。

## 5. 转换器实现要点

1. 主输出走根 `QWEN.md`（不复用 codex 的 `AGENTS.md` 根文件，保持私有根
   避免与跨工具内容串扰；Qwen 自身 `AGENTS.md` 与 `QWEN.md` 并行叠加，非互斥）
2. `RulesDir=".qwen/rules"`（2026-07 由留空改为原生目录）：nonRoot rule
   现在写独立文件，frontmatter `paths` 字段对应官方条件规则语法
   （`GlobsFieldName="paths"`, `GlobsFieldFormat=GlobsList`）
3. commands 走原生 markdown 到 `.qwen/commands/<name>.md`，frontmatter
   字段与 codex 对齐（无需 TOML 转换）
4. skills：`SkillsDir=".qwen/skills"` 已设置，走 `BuildNativeSkillPackage`
   原生落点（2026-07 由降级改为原生，旧文档"SkillsAsRule 设为 false 走降级"已过时）
5. references / subagents 仍走 `BuildDegradedFileOp` 到
   `.qwen/rules/{references,subagents}/<name>.md`
6. 不写 `settings.json`：MCP / 全局配置不在 std-agent 范围内

## 6. 信息来源

- https://github.com/QwenLM/qwen-code （访问日期 2026-05-17）
- https://github.com/QwenLM/qwen-code/blob/main/docs/users/features/memory.md
- https://qwenlm.github.io/qwen-code-docs/en/users/features/skills/

## 7. 已确认（2026-07-10 复核更新）

- Qwen Code 是 Gemini CLI fork，22k stars
- `QWEN.md` 与 `AGENTS.md` 并行叠加加载（非 fallback 互斥关系）
- `.qwen/rules/` 是官方原生 RulesDir，支持 `paths:` 条件规则
- `.qwen/skills/` 是官方原生 SkillsDir，GA 已启用；frontmatter 支持 `name` / `description` / `priority`
- commands 落 `.qwen/commands/`（markdown 格式）
- 中国大陆开发者覆盖率高（Qwen API 国内稳定可用）
- 嵌套 `QWEN.md` 仅 cwd 向上链发现，无向下预扫描（区别于 gemini-cli 的 JIT）

## 8. UNKNOWN（2026-07 复核仍未证实）

- skills frontmatter `priority` 字段的取值范围与作用（官方文档提及字段名，未给出语义细节，transformer 暂未渲染）
- `~/.qwen/settings.json` 的字段集（推测与 gemini-cli `~/.gemini/settings.json`
  接近，但未确认是否同 schema）
- 根文件（QWEN.md）与 skills / commands 的字节上限及截断行为（官方均无数值文档）
