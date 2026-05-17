# Target: Pi

调研日期: 2026-05-17
维护方: earendil-works (Mario Zechner)
仓库: https://github.com/earendil-works/pi

## 1. 摘要

pi 是 Mario Zechner（libgdx 作者）开发的高度可扩展 AI agent runner，
定位为"AGENTS.md + Agent Skills 标准严格执行者"。pi 不发明私有协议，
直接消费业界既成事实标准：根 `AGENTS.md` 是项目主指令文件，Agent
Skills 标准（`SKILL.md` + 子文件包）按 Anthropic / AAIF 规范严格落地。

设计哲学差异（相对 codex / claude-code）：pi 把"加载顺序 + 系统提示
替换"做成显式机制（`SYSTEM.md` 可替换或追加内置 system prompt），
其他工具只做隐式叠加。Skill frontmatter 字段不容忍偏离规范，写错
不识别。

## 2. 配置文件路径

| 类别 | 路径 | 说明 |
|---|---|---|
| 全局 AGENTS | `~/.pi/agent/AGENTS.md` | 全局默认指令 |
| 项目 AGENTS | `<repo>/AGENTS.md` | 项目主指令，沿 cwd 向 parent 叠加 |
| 嵌套 AGENTS | `<repo>/<subdir>/AGENTS.md` | 子目录限定，近端胜出 |
| Prompts | `<repo>/.pi/prompts/<name>.md` | slash 模板，调用方式 `/<name>` |
| 项目 Skills | `<repo>/.pi/skills/<name>/SKILL.md` | Agent Skills 标准包 |
| 共享 Skills | `<repo>/.agents/skills/<name>/SKILL.md` | 与 codex / claude / crush 共享生态位 |
| 系统提示 | `<repo>/SYSTEM.md` 或 `~/.pi/agent/SYSTEM.md` | 替换/追加 pi 内置 system prompt |

加载顺序：`~/.pi/agent/` -> parent dirs（git root 向下叠加）-> cwd。

## 3. 文件格式

| 文件 | 格式 | frontmatter |
|---|---|---|
| `AGENTS.md` | Markdown | 无（全文即指令） |
| `.pi/prompts/<n>.md` | Markdown | 可选（pi 接受 description / argument-hint） |
| `.pi/skills/<n>/SKILL.md` | Markdown | 严格 Agent Skills 字段（name / description / license / compatibility / metadata） |
| `SYSTEM.md` | Markdown | 无 |

pi 不约束 AGENTS.md 字节上限，但单文件超过 ~32KB 时建议拆 nested
AGENTS.md 控制注入预算。

## 4. std-ai 四类映射

| std-ai 类型 | pi 落点 |
|---|---|
| rules | 根 `AGENTS.md`（所有 nonRoot rules inline 拼接） |
| skills | `<repo>/.pi/skills/<n>/SKILL.md` + 同目录辅助文件 |
| commands | `<repo>/.pi/prompts/<n>.md`，调用方式 `/<n>` |
| references | 降级为 Agent Skills 包写到 `.pi/skills/<n>/SKILL.md`（pi 严格执行 skill 协议，单文件 references 包装为 skill） |
| subagents | 降级到 `.pi/rules/subagents/<n>.md`（pi 无原生 subagent，rule-equivalent fallback） |

## 5. 转换器实现要点

1. 主输出：项目根 `AGENTS.md`，由 `inject` footer 标识为 stdagent 生成
2. RulesDir 留空：pi 无子目录 rules，所有 nonRoot rule body 直接 inline
   到 AGENTS.md（amp / warp 风格）
3. SkillsDir = `.pi/skills`，原生 Agent Skills 标准包（项目级）；
   stdagent 不主动写 `.agents/skills/` 共享路径，避免与 codex 适配器
   产生路径冲突
4. CommandsDir = `.pi/prompts`，原生 markdown + 可选 frontmatter
5. FallbackDir = `.pi/rules`：subagents 等无原生支持的 type 走子目录
   隔离 fallback
6. InjectExplainer / InjectStdaiTypeField / InjectTypeGlossary 全开：
   pi 严格执行规范，元数据完整有助于工具识别 stdagent 生成内容
7. 不写 `SYSTEM.md`：pi 把 SYSTEM.md 视为系统提示替换入口，不在四类
   映射范围

## 6. 信息来源

- https://github.com/earendil-works/pi（仓库主页，访问日期 2026-05-17）
- /tmp/std-ai-protocol-research.md（行 29，2026-05-17 调研笔记）

## 7. UNKNOWN

- `SYSTEM.md` 替换 vs 追加的具体策略（fenced block? 标记?）公开文档
  描述简略，stdagent 不主动写
- pi 是否支持 `AGENTS.override.md`（grok 风格 per-dir 覆盖）：当前
  按"无覆盖"实现，仅依赖 nested 叠加
- Pi 与 `.agents/skills/`（共享路径）的优先级关系：实测 pi 同时读两
  个目录，stdagent v1 仅写 `.pi/skills/`，由用户在需要跨工具复用时
  手动迁移
