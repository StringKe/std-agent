package protocol

import (
	"path"
	"strings"

	"github.com/StringKe/std-agent/internal/config"
	"github.com/StringKe/std-agent/internal/parser"
	"github.com/StringKe/std-agent/internal/transformer/transformerutil"
	"github.com/StringKe/std-agent/internal/writer"
)

// defaultFallbackSubdir 返回 doc type 对应的默认 fallback 子目录名。
// rules 落到 FallbackDir 根（无子目录），其余 type 都走专属子目录。
func defaultFallbackSubdir(t parser.DocType) string {
	switch t {
	case parser.TypeSkills:
		return "skills"
	case parser.TypeCommands:
		return "commands"
	case parser.TypeReferences:
		return "references"
	case parser.TypeSubagents:
		return "subagents"
	}
	return ""
}

// BuildDegradedFileOp 把不原生支持的 type doc 写入 fallback 路径。
//
// 路径规则（v3：子目录隔离，无私有前缀）：
//
//	<FallbackDir or RulesDir>/<subdir>/<name>.md
//
// subdir 优先取 adapter.FallbackSubdir[doc.Type]，缺失则按 defaultFallbackSubdir 推断。
//
// frontmatter 行为：
//   - description（若 doc.Description 非空）
//   - std-agent-type: <type>（若 adapter.InjectStdaiTypeField=true）
//
// body 行为：
//   - 若 explainer 启用（adapter.InjectExplainer，可被 InjectExplainerOverride 覆盖），
//     在 body 头部插入 HTML 注释段（不污染规范字段）
//
// 该 helper 仅产单个 FileOp，用于 rule-equivalent fallback。
// SKILL.md 标准形式 fallback 走 BuildDegradedSkillPackage。
func BuildDegradedFileOp(doc *parser.Document, adapter Adapter, cfg *config.Config) writer.FileOp {
	fullPath := degradedFilePath(doc, adapter)

	var fm transformerutil.FmBuilder
	if doc.Description != "" {
		fm.Add("description", doc.Description)
	}
	if adapter.InjectStdaiTypeField {
		fm.Add("std-agent-type", string(doc.Type))
	}

	body := doc.Body
	if shouldInjectExplainer(doc.Type, adapter) {
		body = renderExplainerHeader(doc.Type, doc.Name, adapter.Name) + "\n\n" + body
	}
	if adapter.SubagentInvokeCmd != "" && doc.Type == parser.TypeSubagents {
		body = renderSubagentInvokeBody(doc, adapter.SubagentInvokeCmd, adapter.Name)
	}

	opts := transformerutil.MakeOpts(cfg, adapter.Name, doc.Path, false)
	return transformerutil.BuildMarkdownFile(fullPath, fm.String(), body, opts)
}

// BuildDegradedSkillPackage 把 fallback 的 doc 写为 Agent Skills 标准
// <skillDir>/<name>/SKILL.md 形式（用于支持 Agent Skills 的 target 接收
// references / commands / subagents fallback）。
//
// 路径：<adapter.SkillsDir 或 fallback>/<name>/SKILL.md，
// 同时把 SkillFiles 一并 fan-out 到同目录下。
//
// frontmatter 含 Agent Skills 规范字段（name / description / license /
// compatibility / metadata）+ std-agent-type 私有标识（若 InjectStdaiTypeField=true）。
// body 头部含 explainer HTML 注释（若 InjectExplainer=true）。
func BuildDegradedSkillPackage(doc *parser.Document, adapter Adapter, cfg *config.Config) []writer.FileOp {
	skillDir := degradedSkillDir(doc, adapter)

	var fm transformerutil.FmBuilder
	fm.Add("name", doc.Name)
	fm.Add("description", transformerutil.MergeDescription(doc.Description, doc.WhenToUse))
	fm.Add("license", doc.License)
	fm.Add("compatibility", doc.Compatibility)
	fm.AddMap("metadata", doc.Metadata)
	if adapter.InjectStdaiTypeField {
		fm.Add("std-agent-type", string(doc.Type))
	}

	body := doc.Body
	if shouldInjectExplainer(doc.Type, adapter) {
		body = renderExplainerHeader(doc.Type, doc.Name, adapter.Name) + "\n\n" + body
	}
	if adapter.SubagentInvokeCmd != "" && doc.Type == parser.TypeSubagents {
		body = renderSubagentInvokeBody(doc, adapter.SubagentInvokeCmd, adapter.Name)
	}

	opts := transformerutil.MakeOpts(cfg, adapter.Name, doc.Path, false)
	skillMd := transformerutil.BuildMarkdownFile(skillDir+"/SKILL.md", fm.String(), body, opts)
	return transformerutil.BuildSkillPackage(skillDir, skillMd, doc.SkillFiles)
}

// degradedFilePath 计算 fallback 文件路径
func degradedFilePath(doc *parser.Document, adapter Adapter) string {
	dir := adapter.FallbackDir
	if dir == "" {
		dir = adapter.RulesDir
	}
	sub := ""
	if adapter.FallbackSubdir != nil {
		if v, ok := adapter.FallbackSubdir[doc.Type]; ok {
			sub = v
		} else {
			sub = defaultFallbackSubdir(doc.Type)
		}
	} else {
		sub = defaultFallbackSubdir(doc.Type)
	}
	if sub == "" {
		return path.Join(dir, doc.Name+".md")
	}
	return path.Join(dir, sub, doc.Name+".md")
}

// degradedSkillDir 计算 fallback 到 Agent Skills 标准形式时的 skill 目录
func degradedSkillDir(doc *parser.Document, adapter Adapter) string {
	base := adapter.SkillsDir
	if base == "" {
		dir := adapter.FallbackDir
		if dir == "" {
			dir = adapter.RulesDir
		}
		sub := defaultFallbackSubdir(parser.TypeSkills)
		base = path.Join(dir, sub)
	}
	return path.Join(base, doc.Name)
}

// shouldInjectExplainer 综合 adapter.InjectExplainer 与 InjectExplainerOverride
// 决定 doc 是否应注入 explainer header
func shouldInjectExplainer(t parser.DocType, adapter Adapter) bool {
	if v, ok := adapter.InjectExplainerOverride[t]; ok {
		return v
	}
	return adapter.InjectExplainer
}

// renderExplainerHeader 生成 HTML 注释元解释段。
//
// 注入位置：frontmatter 之后、body 之前（由 BuildMarkdownFile 拼接顺序保证）。
// 输出内容按 type 切换 semantics 说明，方便 AI 工具阅读时理解该文件不是普通 rule。
func renderExplainerHeader(t parser.DocType, name, target string) string {
	semantics := explainerSemanticsFor(t)
	var b strings.Builder
	b.WriteString("<!-- std-agent degraded ")
	b.WriteString(string(t))
	b.WriteString(": ")
	b.WriteString(name)
	b.WriteString(" -->\n")
	b.WriteString("<!-- Target tool \"")
	b.WriteString(target)
	b.WriteString("\" does not natively support this type. ")
	b.WriteString(semantics)
	b.WriteString(" -->")
	return b.String()
}

// explainerSemanticsFor 返回 type 在 fallback 场景下的语义说明文本
func explainerSemanticsFor(t parser.DocType) string {
	switch t {
	case parser.TypeSkills:
		return "Skill is an on-demand capability pack; AI loads it when the description matches user intent."
	case parser.TypeCommands:
		return "Command is a user-invoked template; AI does not auto-trigger it."
	case parser.TypeReferences:
		return "Reference is background material; AI consults it only when needed, not auto-loaded."
	case parser.TypeSubagents:
		return "Subagent is an isolated sub-agent definition; AI delegates work via spawn or CLI invocation."
	}
	return "See .stdai/standards/ for the source type."
}

// renderSubagentInvokeBody 生成"通过 CLI 调用其他 AI 实现 subagent"的 body。
//
// invokeCmd 含 {name} 占位（如 "claude --agent {name}"），替换后嵌入 shell
// fenced block 中。原 doc.Body 作为"Subagent prompt body"段附在末尾，便于
// 用户阅读子代理实际指令。
func renderSubagentInvokeBody(doc *parser.Document, invokeCmd, adapterName string) string {
	cmd := strings.ReplaceAll(invokeCmd, "{name}", doc.Name)
	var b strings.Builder
	b.WriteString("<!-- std-agent degraded subagents: ")
	b.WriteString(doc.Name)
	b.WriteString(" via CLI invocation -->\n")
	b.WriteString("<!-- Target tool \"")
	b.WriteString(adapterName)
	b.WriteString("\" does not natively support sub-agents. ")
	b.WriteString("To delegate this subagent, the AI should spawn it by running the following shell command in its execution environment. -->\n\n")
	b.WriteString("# Subagent: ")
	b.WriteString(doc.Name)
	b.WriteString("\n\n")
	if doc.Description != "" {
		b.WriteString(doc.Description)
		b.WriteString("\n\n")
	}
	b.WriteString("## How to spawn\n\n")
	b.WriteString("This subagent runs in an isolated context. When the user requests its work, execute:\n\n")
	b.WriteString("```bash\n")
	b.WriteString(cmd)
	b.WriteString(" \"<task description>\"\n")
	b.WriteString("```\n\n")
	b.WriteString("The target subagent definition lives at `.stdai/standards/subagents/")
	b.WriteString(doc.Name)
	b.WriteString(".md`.\n\n")
	b.WriteString("## Subagent prompt body\n\n")
	b.WriteString(strings.TrimSpace(doc.Body))
	b.WriteString("\n")
	return b.String()
}
