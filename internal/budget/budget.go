// Package budget 负责 LLM 上下文消耗的提醒与限额检查。
//
// 不同 AI CLI 工具对 rule / skill / command 文件大小有自己的硬上限或软建议，
// 加上跨工具通用的"context budget 友好"软建议（rules 不超 8000 字符等），
// 本包统一表达为 Limit 数组并在 sync 时由 runner 发出 WARN。
//
// 字符 vs token：本包统计字节数（len(string)），不做 tokenizer 估算；中英文
// 混合 LLM tokenization 经验值约 2-4 字节/token，超 8000 字符约 2000-4000 tokens，
// 已足够触发"context 较大"提醒。
package budget

import (
	"fmt"

	"github.com/StringKe/std-agent/internal/parser"
)

// Limit 描述某 target 在某 kind 上的字符上限
type Limit struct {
	Target string // 适用 target；"*" 表示通用
	Kind   string // rule / skill / command / agents-md-total
	Soft   int    // 软警告阈值；0 表示不检查
	Hard   int    // 硬上限；0 表示不检查
	Note   string // 用户可读说明
}

// Limits 是所有已知限额，由 CheckDocument / CheckTotalRules / CheckRootFile 遍历检查。
//
// 数值原则（2026-07 全 target 官方文档核查，来源见 docs/sdlc/2026-07-10/spec-refresh/spec.md）：
// Hard 只填官方文档化的上限，官方没写就置 0——宁缺毋假；Soft 是跨工具通用的
// context 友好建议。三种超限语义在 Note 里写清：截断（部分内容丢）/ 拒载（整文件
// 或整字段失效）/ 淘汰（按优先级弃置低序内容）。
//
//nolint:gochecknoglobals // 全局只读配置，无并发风险
var Limits = []Limit{
	// 通用软建议（context budget 友好，无官方依据，纯经验值）
	{"*", "rule", 8000, 0, "rule 单文件 > 8000 字符会显著消耗 LLM context；考虑拆分或转 skill"},
	{"*", "skill", 20000, 0, "SKILL.md > 20000 字符（约 5000 tokens 软上限）；把内容拆到 references/ 或 scripts/（Agent Skills 规范建议正文 < 500 行）"},
	{"*", "command", 4000, 0, "command 主体 > 4000 字符会显著消耗 context；历史上 Copilot Code Review 截前 4000 字符的规则已于 2026-06-12 官方移除，此条现为纯建议"},

	// Agent Skills 规范字段硬限（超限硬拒载：skill 不被索引；
	// Copilot / Cline / Roo / OpenCode / Augment 等共用同一规范）
	{"*", "skill-name", 0, 64, "Agent Skills 规范：name ≤ 64 字符（小写字母数字连字符，须等于目录名），超限 skill 不被索引"},
	{"*", "skill-description", 0, 1024, "Agent Skills 规范：description ≤ 1024 字符，超限 skill 不被索引"},
	{"claude-code", "skill-listing", 0, 1536, "Claude Code skill listing 截断：description + when_to_use 合计 > 1536 字符时列表中被截断（调用时仍读全文），触发词要前置"},

	// target 硬上限（语义见 Note）
	{"codex", "agents-md-total", 0, 32768, "codex project_doc_max_bytes 默认 32768：root->cwd 整条链（含嵌套 AGENTS.md）累计字节，超限后按链序整文件停止追加（链尾先丢）；该值是 config.toml 可调默认值而非硬上限"},
	{"kimi-code", "agents-md-total", 0, 32768, "Kimi Code AGENTS.md 层级发现 32KiB 预算：项目根到 cwd 逐级合并，leaf-first 分配（超限祖先层先丢）"},
	{"cursor", "rule", 80000, 100000, "Cursor 单 rule 文件 100000 字符上限（服务端下发，可能变动），超限截断并提示；>80000 接近上限"},
	{"windsurf", "rule", 0, 12000, "Windsurf workspace rule 单文件上限 12000 字符（per-file 非总量）；超限行为官方未定义，legacy .windsurfrules 实测为截断"},
	{"windsurf", "global-rules", 0, 6000, "Windsurf global_rules.md 上限 6000 字符"},
	{"antigravity", "rule", 0, 12000, "Antigravity rule 单文件上限 12000 字符（官方文档）"},
	{"antigravity", "command", 0, 12000, "Antigravity workflow 单文件上限 12000 字符（官方文档：Workflow files are limited to 12,000 characters each）"},
	{"copilot", "subagent", 0, 30000, "Copilot .agent.md 正文上限 30000 字符（GitHub.com custom agents 参考文档）"},
	{"augment-code", "rules-total", 0, 49512, "Augment Workspace Guidelines + Rules 合计上限 49512 字符；超限按优先级淘汰（manual -> always/auto -> .augment-guidelines 顺序应用直至限额，后序静默弃置）"},

	// 根文件体积（CLAUDE.md / AGENTS.md / GEMINI.md / .github/copilot-instructions.md）
	// 这些文件**每次 AI session 启动都会被整体加载到 system prompt**，体积直接影响
	// 起始 context 占用，与按需加载的 .claude/rules/<name>.md 不同。root rule body
	// 写得过大会导致 session 一开始就吃掉大量 token，建议把详细规则拆到非 root rule。
	{"claude-code", "root-file", 8000, 0, "CLAUDE.md 每次 session 启动整体加载且官方明确全量不截断（无硬上限，软指导 under 200 lines）；> 8k 字符（~2k tokens）建议把详细规则拆到非 root rule，用 @import 按需加载"},
	{"codex", "root-file", 8000, 32768, "AGENTS.md 由 codex 启动整体加载；project_doc_max_bytes 默认 32768（链累计口径、可调），codex rules 全文 inline 到 AGENTS.md，> 8k（~2k tokens）建议精简规则或对低优先级 rule 关闭 codex target"},
	{"cursor", "root-file", 80000, 100000, "Cursor 把 AGENTS.md / CLAUDE.md 按单 rule 同一 100000 字符上限处理，超限截断（限额服务端下发可能变动）"},
	{"kimi-code", "root-file", 8000, 32768, "AGENTS.md 由 Kimi Code 层级发现加载，全链 32KiB 预算（leaf-first），kimi-code rules 全文 inline 到 AGENTS.md，> 8k（~2k tokens）建议精简规则或对低优先级 rule 关闭 kimi-code target"},
	{"gemini", "root-file", 8000, 0, "GEMINI.md 由 Gemini CLI 启动加载到 system prompt，官方无字节上限文档；> 8k 字符建议精简根文件，把详细规则拆到非 root rule"},
	{"copilot", "root-file", 8000, 0, "copilot-instructions.md 无硬性字符上限（历史截断规则已于 2026-06-12 移除）；官方软指导不超过约 2 页，> 8k 建议精简"},
}

// CheckDocument 对单个 Document 做 budget 检查（body 字节 + skill 字段长度），
// 返回 WARN 消息列表
func CheckDocument(d *parser.Document) []string {
	if d == nil {
		return nil
	}
	n := d.BodyBytes
	if n == 0 {
		n = len([]byte(d.Body))
	}
	kind := docKind(d.Type)
	if kind == "" {
		return nil
	}
	var out []string
	out = append(out, checkKind(kind, d.Path, n)...)

	// skill frontmatter 字段长度（Agent Skills 规范硬拒载 + Claude listing 截断）
	if d.Type == parser.TypeSkills {
		out = append(out, checkKind("skill-name", d.Path, len(d.Name))...)
		out = append(out, checkKind("skill-description", d.Path, len(d.Description))...)
		out = append(out, checkKind("skill-listing", d.Path, len(d.Description)+len(d.WhenToUse))...)
	}
	return out
}

// checkKind 用 Limits 表校验某 kind 的字节/字符数
func checkKind(kind, path string, n int) []string {
	var out []string
	for _, l := range Limits {
		if l.Kind != kind {
			continue
		}
		if l.Hard > 0 && n > l.Hard {
			out = append(out, fmt.Sprintf("HARD %s [%s] %s %d > %d (%s)", path, l.Target, kind, n, l.Hard, l.Note))
			continue
		}
		if l.Soft > 0 && n > l.Soft {
			out = append(out, fmt.Sprintf("SOFT %s [%s] %s %d > %d (%s)", path, l.Target, kind, n, l.Soft, l.Note))
		}
	}
	return out
}

// CheckTotalRules 对所有 type=rules Document body 字节求和，与总量类上限比对：
//   - codex agents-md-total：project_doc_max_bytes 链累计（超限链尾整文件不加载）
//   - augment-code rules-total：Workspace Guidelines + Rules 合计（超限按优先级淘汰）
func CheckTotalRules(docs []*parser.Document) []string {
	total := 0
	for _, d := range docs {
		if d.Type != parser.TypeRules {
			continue
		}
		n := d.BodyBytes
		if n == 0 {
			n = len([]byte(d.Body))
		}
		total += n
	}
	var out []string
	for _, l := range Limits {
		if (l.Kind != "agents-md-total" && l.Kind != "rules-total") || l.Hard == 0 {
			continue
		}
		if total > l.Hard {
			out = append(out, fmt.Sprintf("HARD [%s] total rules %d > %d (%s)", l.Target, total, l.Hard, l.Note))
		}
	}
	return out
}

// CheckRootFile 对 transformer 生成的根文件（CLAUDE.md / AGENTS.md / 等）做体积检查。
//
// path 是相对项目根的路径（用于 WARN 输出定位），target 是 transformer 名（"claude-code" 等），
// sizeBytes 是 final content 字节数（含 stdagent header/footer marker）。返回 WARN 字符串列表。
//
// 与 CheckDocument 不同：CheckDocument 检查源端 std rule body 字节，CheckRootFile 检查
// 输出端实际写盘的根文件字节。两者都重要：源 rule 大 -> 单 rule 占 context；根文件大 ->
// 启动即占 context。
func CheckRootFile(target, path string, sizeBytes int) []string {
	var out []string
	for _, l := range Limits {
		if l.Kind != "root-file" || l.Target != target {
			continue
		}
		if l.Hard > 0 && sizeBytes > l.Hard {
			out = append(out, fmt.Sprintf("HARD %s [%s] root-file %d > %d (%s)", path, target, sizeBytes, l.Hard, l.Note))
			continue
		}
		if l.Soft > 0 && sizeBytes > l.Soft {
			out = append(out, fmt.Sprintf("SOFT %s [%s] root-file %d > %d (%s)", path, target, sizeBytes, l.Soft, l.Note))
		}
	}
	return out
}

// docKind 把 parser.DocType 转 budget kind
func docKind(t parser.DocType) string {
	switch t {
	case parser.TypeRules:
		return "rule"
	case parser.TypeSkills:
		return "skill"
	case parser.TypeCommands:
		return "command"
	case parser.TypeSubagents:
		return "subagent"
	}
	return ""
}
