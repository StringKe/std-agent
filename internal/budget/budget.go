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

	"std-ai/internal/parser"
)

// Limit 描述某 target 在某 kind 上的字符上限
type Limit struct {
	Target string // 适用 target；"*" 表示通用
	Kind   string // rule / skill / command / agents-md-total
	Soft   int    // 软警告阈值；0 表示不检查
	Hard   int    // 硬上限；0 表示不检查
	Note   string // 用户可读说明
}

// Limits 是所有已知限额，由 CheckDocument / CheckTotalRules 遍历检查
//
//nolint:gochecknoglobals // 全局只读配置，无并发风险
var Limits = []Limit{
	// 通用软建议（context budget 友好）
	{"*", "rule", 8000, 0, "rule 单文件 > 8000 字符会显著消耗 LLM context；考虑拆分或转 skill"},
	{"*", "skill", 20000, 0, "SKILL.md > 20000 字符（约 5000 tokens 软上限）；把内容拆到 references/ 或 scripts/"},
	{"*", "command", 4000, 0, "command 主体 > 4000 字符；GitHub Copilot Code Review 仅读前 4000 字符"},

	// target 硬上限（部分由 transformer 自行处理 spill / 截断）
	{"codex", "agents-md-total", 0, 32768, "AGENTS.md 总字节超过 32768，将自动 spill 到 .codex/rules/"},
	{"cursor", "rule", 80000, 100000, "Cursor rule 字符上限 100000，>80000 接近上限；超 100000 触发截断"},
	{"windsurf", "rule", 0, 12000, "Windsurf rule 单文件硬上限 12000 字符"},
	{"antigravity", "rule", 0, 12000, "Antigravity rule 单文件硬上限 12000 字符"},
	{"windsurf", "global-rules", 0, 6000, "Windsurf global_rules.md 上限 6000 字符"},
}

// CheckDocument 对单个 Document body 做 budget 检查，返回 WARN 消息列表
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
	for _, l := range Limits {
		if l.Kind != kind {
			continue
		}
		if l.Hard > 0 && n > l.Hard {
			out = append(out, fmt.Sprintf("HARD %s [%s] %d > %d (%s)", d.Path, l.Target, n, l.Hard, l.Note))
			continue
		}
		if l.Soft > 0 && n > l.Soft {
			out = append(out, fmt.Sprintf("SOFT %s [%s] %d > %d (%s)", d.Path, l.Target, n, l.Soft, l.Note))
		}
	}
	return out
}

// CheckTotalRules 对所有 type=rules Document body 字节求和，
// 与 codex AGENTS.md 总字节硬上限比对（spill 已自动处理，但提醒用户拆分）
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
		if l.Kind != "agents-md-total" || l.Hard == 0 {
			continue
		}
		if total > l.Hard {
			out = append(out, fmt.Sprintf("HARD AGENTS.md [%s] total rules %d > %d (%s)", l.Target, total, l.Hard, l.Note))
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
	}
	return ""
}
