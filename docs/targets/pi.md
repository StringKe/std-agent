# Target: Pi

调研日期: 2026-05-17（2026-07-10 补充调研）
维护方: earendil-works (Mario Zechner)
仓库: https://github.com/earendil-works/pi

## 1. 摘要

pi 是 Mario Zechner（libgdx 作者）开发的高度可扩展 AI agent runner，
定位为"AGENTS.md + Agent Skills 标准严格执行者"。pi 不发明私有协议，
直接消费业界既成事实标准：根 `AGENTS.md` 是项目主指令文件（`CLAUDE.md`
作为等价别名同样被识别），Agent Skills 标准（`SKILL.md` + 子文件包）按
Anthropic / AAIF 规范严格落地。

设计哲学差异（相对 codex / claude-code）：pi 把"系统提示替换 / 追加"做成
显式机制（`SYSTEM.md` 整段替换，`APPEND_SYSTEM.md` 追加而不替换），其他
工具只做隐式叠加。Skill frontmatter 字段不容忍偏离规范，写错不识别。

**加载方向只向上、不向下**：pi 在启动时从三个来源加载 AGENTS.md（或
CLAUDE.md）并全部拼接：`~/.pi/agent/AGENTS.md`（全局）、从 cwd 向上walk
的各级 parent 目录、cwd 本身。pi **不做子目录向下发现**，即位于 cwd 下方
子目录的 AGENTS.md 不会被自动纳入（除非用户显式把 cwd 切到那个子目录）。
此前调研记录的"子目录限定、近端胜出"描述不准确，已修正。

## 2. 配置文件路径

| 类别 | 路径 | 说明 |
|---|---|---|
| 全局 AGENTS | `~/.pi/agent/AGENTS.md` | 全局默认指令 |
| 项目 AGENTS | `<repo>/AGENTS.md` | 项目主指令，从 cwd 向上 walk 到各级 parent 后与全局拼接 |
| Parent AGENTS | `<repo 上层各目录>/AGENTS.md` | 只向上 walk，不向下发现子目录 |
| Prompts | `<repo>/.pi/prompts/<name>.md` | slash 模板，调用方式 `/<name>` |
| 项目 Skills | `<repo>/.pi/skills/<name>/SKILL.md` | Agent Skills 标准包；发现范围含 cwd 向上至 git root
  的各级祖先目录 |
| 共享 Skills | `<repo>/.agents/skills/<name>/SKILL.md` | 与 codex / claude / crush 共享生态位，同样按祖先目录链发现 |
| 全局 Skills | `~/.pi/agent/skills/`、`~/.agents/skills/` | 用户级 skill |
| 系统提示（替换） | `<repo>/.pi/SYSTEM.md` 或 `~/.pi/agent/SYSTEM.md` | 整段替换 pi 内置 system prompt，项目级优先于全局级 |
| 系统提示（追加） | `<repo>/.pi/APPEND_SYSTEM.md` 或 `~/.pi/agent/APPEND_SYSTEM.md` | 追加而不替换内置 system prompt |

加载顺序：`~/.pi/agent/AGENTS.md`（全局） + 从 cwd 向上 walk 的各级 parent
AGENTS.md + cwd 自身 AGENTS.md，全部拼接注入，不做向下子目录扫描。

## 3. 文件格式

| 文件 | 格式 | frontmatter |
|---|---|---|
| `AGENTS.md` | Markdown | 无（全文即指令） |
| `.pi/prompts/<n>.md` | Markdown | 可选（pi 接受 description / argument-hint） |
| `.pi/skills/<n>/SKILL.md` | Markdown | 严格 Agent Skills 字段（name / description / license / compatibility / metadata） |
| `SYSTEM.md` / `APPEND_SYSTEM.md` | Markdown | 无 |

pi 不约束 AGENTS.md 字节上限，但单文件超过 ~32KB 时建议拆多级 parent
AGENTS.md 控制注入预算（注意：拆分只在向上 walk 链路上有效，向下的子目录
拆分不会被自动发现）。

## 4. std-agent 四类映射

| std-agent 类型 | pi 落点 |
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
3. `NestedSupported: true`：transformer 仍会为带 `NestedPath` 的 root rule
   产出 `<NestedPath>/AGENTS.md`。这类文件只有当用户把 cwd 切到该子目录
   运行 pi 才会被读取（pi 只向上 walk，不主动向下发现），效果与直接在该
   子目录内运行 pi 等价，用途有限但无害，保留产出供该场景使用
4. SkillsDir = `.pi/skills`，原生 Agent Skills 标准包（项目级）；
   stdagent 不主动写 `.agents/skills/` 共享路径，避免与 codex 适配器
   产生路径冲突
5. CommandsDir = `.pi/prompts`，原生 markdown + 可选 frontmatter
6. FallbackDir = `.pi/rules`：subagents 等无原生支持的 type 走子目录
   隔离 fallback
7. InjectExplainer / InjectStdaiTypeField / InjectTypeGlossary 全开：
   pi 严格执行规范，元数据完整有助于工具识别 stdagent 生成内容
8. 不写 `SYSTEM.md` / `APPEND_SYSTEM.md`：这两个文件是系统提示替换 / 追加
   入口，语义上不对应任何 std-agent 类型，不在四类映射范围，stdagent 不
   主动写

## 6. 信息来源

- https://github.com/earendil-works/pi（仓库主页，访问日期 2026-05-17 / 2026-07-10 复核）
- https://github.com/earendil-works/pi/blob/main/packages/coding-agent/README.md（
  AGENTS.md 加载顺序、SYSTEM.md / APPEND_SYSTEM.md 路径与语义，2026-07-10 新增）
- https://github.com/earendil-works/pi/blob/main/packages/coding-agent/docs/skills.md（
  skills 沿祖先目录链发现，2026-07-10 新增）

## 7. 已确认（2026-07-10 新增）

- AGENTS.md 加载只从 `~/.pi/agent/`、cwd 向上各级 parent、cwd 自身三处拼接，
  不做向下子目录发现
- pi 同时识别 `CLAUDE.md` 作为 AGENTS.md 的等价别名
- `SYSTEM.md` 路径为项目级 `.pi/SYSTEM.md` 或全局 `~/.pi/agent/SYSTEM.md`，
  整段替换默认 system prompt；`APPEND_SYSTEM.md` 走同样两级路径，追加而不替换
- 项目级优先于全局级（`.pi/SYSTEM.md` 优先于 `~/.pi/agent/SYSTEM.md`）
- skills 发现范围含 cwd 向上至 git root 的各级祖先目录（非仅 cwd 本身）

## 8. UNKNOWN

- Pi 与 `.agents/skills/`（共享路径）和 `.pi/skills/` 同时存在时的优先级/合并
  关系：实测 pi 同时读两个目录，stdagent v1 仅写 `.pi/skills/`，由用户在需要
  跨工具复用时手动迁移
- `.pi/rules/` fallback 目录本身是否被 pi 以任何形式扫描（当前假设不被扫描，
  仅供人工/其他工具检索）
