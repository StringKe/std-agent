# Target: Qwen Code (Alibaba)

调研日期: 2026-05-17
官方仓库: https://github.com/QwenLM/qwen-code
官方主页: https://github.com/QwenLM/qwen-code

## 1. 摘要

Qwen Code 是阿里巴巴通义实验室（QwenLM）维护的 CLI Agent，基于 Google
Gemini CLI fork 而成，针对 Qwen 大模型生态做了适配。22k stars，是
中国大陆开发者覆盖率最高的开源 AI 编码 CLI 之一（Qwen API 国内可用性
稳定，无需翻墙）。

配置体系延续 Gemini CLI 的设计：根目录上下文文件 `QWEN.md`（命名沿
gemini-cli 的 `GEMINI.md` 模板），fallback 兼容读 `AGENTS.md`（跨工具
基础层）；自定义命令落 `.qwen/commands/`。与 gemini-cli 的关键差异：
commands 用普通 markdown 而非 TOML，更接近 codex / claude-code 风格。

## 2. 配置文件路径

| 类别 | 路径 | 说明 |
|---|---|---|
| 全局上下文 | `~/.qwen/QWEN.md` | 用户级，自动加载 |
| 项目上下文 | `<repo>/QWEN.md` | 项目级，优先级最高 |
| 跨工具 fallback | `<repo>/AGENTS.md` | 当 QWEN.md 缺失时读 |
| 嵌套上下文 | `<repo>/<subdir>/QWEN.md` | 子目录限定，沿祖先链加载 |
| 用户 settings | `~/.qwen/settings.json` | UNKNOWN（与 gemini-cli 同源） |
| 项目自定义命令 | `<repo>/.qwen/commands/*.md` | slash 命令 markdown |

## 3. 文件格式

| 文件 | 扩展 | frontmatter |
|---|---|---|
| `QWEN.md` | `.md` | 无 frontmatter，纯指令文本 |
| `AGENTS.md`（fallback） | `.md` | 无 frontmatter |
| `.qwen/commands/*.md` | `.md` | `description` / `argument-hint` / `tools` / `model`（与 codex / claude-code 对齐） |

与 gemini-cli 的格式差异：

- gemini-cli commands 用 TOML（`description` + `prompt` 双字段）
- qwen-code commands 用 markdown（与 codex / claude-code 统一）

## 4. std-agent 四类映射

| std-agent 类型 | Qwen Code 落点 | 加载方式 |
|---|---|---|
| rules | 项目根 `QWEN.md`，nonRoot rule 全部 inline 拼接到主文件（无子目录 rules） | 自动加载，无 frontmatter |
| skills | fallback `.qwen/rules/skills/<name>/SKILL.md`（Agent Skills 标准包） | std-agent 降级，AI 按 description 触发 |
| commands | `.qwen/commands/<name>.md`，frontmatter 描述参数 | `/` 触发 |
| references | fallback `.qwen/rules/references/<name>.md` | std-agent 降级 |
| subagents | fallback `.qwen/rules/subagents/<name>.md` | std-agent 降级 |

嵌套 root（`std-agent/standards/rules/*.md` 带 `nestedPath`）写到对应子目录的
`QWEN.md`，不带 manifest（与 codex / gemini-cli 一致）。

## 5. 转换器实现要点

1. 主输出走根 `QWEN.md`（不复用 codex 的 `AGENTS.md` 根文件，保持私有根
   避免与跨工具内容串扰；Qwen 自身有 `AGENTS.md` fallback 兜底）
2. RulesDir 留空：与 gemini-cli 同源，无原生子目录 rules，所有 nonRoot
   rule body 直接 inline 拼接到 `QWEN.md`
3. commands 走原生 markdown 到 `.qwen/commands/<name>.md`，frontmatter
   字段与 codex 对齐（无需 TOML 转换）
4. SkillsAsRule 设为 false（与 amp 同理）：RulesDir 为空时若开
   SkillsAsRule 会把 skill 写到仓库根，因此走 BuildDegradedSkillPackage
   到 `.qwen/rules/skills/<name>/SKILL.md`
5. references / subagents 走 BuildDegradedFileOp 到
   `.qwen/rules/{references,subagents}/<name>.md`
6. 不写 `settings.json`：MCP / 全局配置不在 std-agent 范围内

## 6. 信息来源

- https://github.com/QwenLM/qwen-code （访问日期 2026-05-17）
- /tmp/std-agent-protocol-research.md 行 22 / 77（Qwen Code 段）

## 7. 已确认

- Qwen Code 是 Gemini CLI fork，22k stars
- 优先读 `QWEN.md`，fallback 兼容 `AGENTS.md`
- commands 落 `.qwen/commands/`（markdown 格式）
- 中国大陆开发者覆盖率高（Qwen API 国内稳定可用）

## 8. UNKNOWN

- commands frontmatter 完整字段集（沿 gemini-cli TOML 字段还是另立
  字段集）：研究文件未明确，转换器先用 codex 风格 frontmatter 输出
- skills 是否有原生目录结构（公开文档未提及）：转换器走降级
- `~/.qwen/settings.json` 的字段集（推测与 gemini-cli `~/.gemini/settings.json`
  接近，但未确认是否同 schema）
- 嵌套 `QWEN.md` 的祖先链加载边界（与 gemini-cli 一致到 `.git` 边界？UNKNOWN）
