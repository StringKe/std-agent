package protocol

import (
	"encoding/json"
	"fmt"
	"path"
	"strings"

	"github.com/StringKe/std-agent/internal/config"
	"github.com/StringKe/std-agent/internal/parser"
	"github.com/StringKe/std-agent/internal/transformer/transformerutil"
	"github.com/StringKe/std-agent/internal/writer"
)

// Copilot 是 GitHub Copilot 协议实现。
//
// 落点：
//   - .github/copilot-instructions.md（根文件，rules root body + manifest 段 + glossary 头部）
//   - .github/instructions/<n>.instructions.md（path-specific rules，applyTo frontmatter）
//   - .github/prompts/<n>.prompt.md（slash commands）
//   - .github/agents/<n>.agent.md（subagents 原生）
//   - .vscode/mcp.json（顶级键 `servers`，与 Claude 的 mcpServers 不同）
//
// skills：Copilot Agent Skills 已 GA（cloud agent / code review / CLI / VS Code），
// SkillsDir 非空时输出原生 .github/skills/<n>/SKILL.md 标准包。
//
// graceful degradation：
//   - references 不原生支持，落到 .github/instructions/references/<n>.instructions.md
//     （保留 .instructions.md 特殊后缀；frontmatter alwaysApply=false / std-agent-type=<type>）
//   - SkillsDir 为空时 skills 同走 .instructions.md 降级（历史行为）
//   - SubagentInvokeCmd 非空时 subagent 走 CLI 调用降级（body 含 shell 调用指引）
type Copilot struct{}

// Plan 实现 Protocol 接口
func (Copilot) Plan(docs []*parser.Document, adapter Adapter, cfg *config.Config) (*writer.Plan, error) {
	if adapter.Name == "" {
		return nil, fmt.Errorf("copilot protocol: adapter.Name is empty")
	}
	plan := &writer.Plan{Target: adapter.Name}
	if adapter.Disabled {
		return plan, nil
	}

	if cfg != nil && cfg.MCP != nil && len(cfg.MCP.Servers) > 0 {
		plan.Files = append(plan.Files, buildCopilotMCP(cfg.MCP, adapter))
	}

	rules := transformerutil.FilterByType(docs, parser.TypeRules)
	skills := transformerutil.FilterByType(docs, parser.TypeSkills)
	commands := transformerutil.FilterByType(docs, parser.TypeCommands)
	references := transformerutil.FilterByType(docs, parser.TypeReferences)
	subagents := transformerutil.FilterByType(docs, parser.TypeSubagents)
	transformerutil.SortDocs(rules)
	transformerutil.SortDocs(skills)
	transformerutil.SortDocs(commands)
	transformerutil.SortDocs(references)
	transformerutil.SortDocs(subagents)

	// rules 拆分：root 或无 applyTo -> 根文件；其余 -> path-specific .instructions.md
	var general, pathSpecific []*parser.Document
	for _, d := range rules {
		if d.Root || len(transformerutil.EffectiveApplyTo(d, adapter.Name)) == 0 {
			general = append(general, d)
		} else {
			pathSpecific = append(pathSpecific, d)
		}
	}

	if op, ok := buildCopilotRoot(general, pathSpecific, adapter, cfg); ok {
		plan.Files = append(plan.Files, op)
	}
	for _, d := range pathSpecific {
		plan.Files = append(plan.Files, buildCopilotPathInstruction(d, adapter, cfg))
	}
	for _, d := range commands {
		plan.Files = append(plan.Files, buildCopilotPrompt(d, adapter, cfg))
	}
	for _, d := range subagents {
		plan.Files = append(plan.Files, buildCopilotSubagent(d, adapter, cfg))
	}
	for _, d := range skills {
		if adapter.SkillsDir != "" {
			plan.Files = append(plan.Files, BuildNativeSkillPackage(d, adapter, cfg)...)
		} else {
			plan.Files = append(plan.Files, buildCopilotFallbackInstruction(d, adapter, cfg))
		}
	}
	for _, d := range references {
		plan.Files = append(plan.Files, buildCopilotFallbackInstruction(d, adapter, cfg))
	}
	return plan, nil
}

// buildCopilotMCP 写 .vscode/mcp.json。VS Code 工作区 MCP 顶级键是 "servers"，
// 与 Claude Code 的 .mcp.json 顶级键 "mcpServers" 不同。
func buildCopilotMCP(mcp *config.MCPConfig, adapter Adapter) writer.FileOp {
	topKey := adapter.MCPTopKey
	if topKey == "" {
		topKey = "servers"
	}
	wrapped := map[string]map[string]config.MCPServer{
		topKey: mcp.Servers,
	}
	body, _ := json.MarshalIndent(wrapped, "", "  ")
	body = append(body, '\n')
	dest := adapter.MCPPath
	if dest == "" {
		dest = ".vscode/mcp.json"
	}
	return writer.FileOp{Path: dest, Content: body}
}

// buildCopilotRoot 拼根文件 .github/copilot-instructions.md
//
// 顺序：glossary 头部（若开启）+ root rule body 或 fallback 标题 + JoinAGENTSStyle
// 拼非 path-specific rules + manifest section 引用 path-specific rules。
// 返回 (FileOp, true) 当有内容；空 docs 全部无内容时返回 (FileOp{}, false)。
func buildCopilotRoot(general, pathSpecific []*parser.Document, adapter Adapter, cfg *config.Config) (writer.FileOp, bool) {
	if adapter.RootFileName == "" && len(general) == 0 && len(pathSpecific) == 0 {
		return writer.FileOp{}, false
	}
	rootPath := adapter.RootFileName
	if rootPath == "" {
		rootPath = ".github/copilot-instructions.md"
	}

	roots, nonRoot := transformerutil.PartitionRoot(general)
	var body strings.Builder
	body.WriteString(RenderGlossaryFor(adapter))
	switch {
	case len(roots) > 0:
		body.WriteString(transformerutil.RenderRootBody(roots))
		if len(nonRoot) > 0 {
			body.WriteString("\n")
			body.WriteString(transformerutil.JoinAGENTSStyle("", nonRoot))
		}
	case len(nonRoot) > 0:
		body.WriteString(transformerutil.JoinAGENTSStyle("Copilot Repository Instructions", nonRoot))
	}

	if len(pathSpecific) > 0 {
		title := adapter.ManifestSection
		if title == "" {
			title = "Path-Specific Instructions"
		}
		body.WriteString(transformerutil.BuildRuleManifestSection(
			title,
			adapter.Name,
			pathSpecific,
			func(d *parser.Document) string {
				return transformerutil.FilePath(".github/instructions", d.Name, ".instructions.md")
			},
			false,
		))
	}

	if body.Len() == 0 {
		return writer.FileOp{}, false
	}

	opts := transformerutil.MakeOpts(cfg, adapter.Name, "", true)
	op := transformerutil.BuildMarkdownFile(rootPath, "", body.String(), opts)
	op.IsRoot = true
	return op, true
}

// buildCopilotPathInstruction 写 path-specific rule 到 .github/instructions/<n>.instructions.md
//
// frontmatter 用 applyTo 字段（不是 globs / paths）。format 由 adapter.GlobsFieldFormat 决定：
//   - GlobsCommaString -> applyTo: "glob1,glob2"
//   - GlobsList -> applyTo: [- "glob1", - "glob2"]
//
// 默认 GlobsCommaString（与现有 copilot.go 行为一致）。
func buildCopilotPathInstruction(d *parser.Document, adapter Adapter, cfg *config.Config) writer.FileOp {
	applyTo := transformerutil.EffectiveApplyTo(d, adapter.Name)
	var fm transformerutil.FmBuilder
	switch adapter.GlobsFieldFormat {
	case GlobsList:
		fm.AddList("applyTo", applyTo)
	default:
		fm.Add("applyTo", strings.Join(applyTo, ","))
	}
	body := d.Body
	if d.Description != "" {
		body = d.Description + "\n\n" + d.Body
	}
	opts := transformerutil.MakeOpts(cfg, adapter.Name, d.Path, false)
	return transformerutil.BuildMarkdownFile(
		transformerutil.FilePath(".github/instructions", d.Name, ".instructions.md"),
		fm.String(), body, opts,
	)
}

// buildCopilotPrompt 写 slash command 到 .github/prompts/<n>.prompt.md
func buildCopilotPrompt(d *parser.Document, adapter Adapter, cfg *config.Config) writer.FileOp {
	var fm transformerutil.FmBuilder
	fm.Add("description", d.Description)
	fm.Add("argument-hint", d.ArgumentHint)
	fm.AddList("tools", d.AllowedTools)
	fm.Add("model", d.Model)
	opts := transformerutil.MakeOpts(cfg, adapter.Name, d.Path, false)
	return transformerutil.BuildMarkdownFile(
		transformerutil.FilePath(".github/prompts", d.Name, ".prompt.md"),
		fm.String(), d.Body, opts,
	)
}

// buildCopilotSubagent 写 subagent。
//
// 默认走原生 .github/agents/<n>.agent.md（copilot 原生支持 subagents）。
// 当 adapter.SubagentInvokeCmd 非空时，进入 Phase 4.2 的 CLI 调用降级形态 B：
// body 含 shell 调用指引，文件落到 .github/agents/<n>.agent.md（保留 .agent.md
// 后缀让 copilot 把它当 agent 加载，body 内的指引告诉 AI 通过 shell 调用其他 CLI 执行）。
func buildCopilotSubagent(d *parser.Document, adapter Adapter, cfg *config.Config) writer.FileOp {
	var fm transformerutil.FmBuilder
	fm.Add("description", transformerutil.MergeDescription(d.Description, d.WhenToUse))
	fm.AddList("tools", d.AllowedTools)
	fm.Add("model", d.Model)
	// 官方 .agent.md 扩展字段（GitHub.com custom agents 参考文档）：
	// disable-model-invocation 取代已废弃的 infer；user-invocable 默认 true
	if d.DisableModelInvocation {
		fm.AddBool("disable-model-invocation", true)
	}
	if d.UserInvocable != nil && !*d.UserInvocable {
		fm.AddBool("user-invocable", false)
	}
	if adapter.InjectStdaiTypeField && adapter.SubagentInvokeCmd != "" {
		fm.Add("std-agent-type", string(parser.TypeSubagents))
	}

	body := d.Body
	if adapter.SubagentInvokeCmd != "" {
		body = renderSubagentInvokeBody(d, adapter.SubagentInvokeCmd, adapter.Name)
	}

	opts := transformerutil.MakeOpts(cfg, adapter.Name, d.Path, false)
	op := transformerutil.BuildMarkdownFile(
		transformerutil.FilePath(".github/agents", d.Name, ".agent.md"),
		fm.String(), body, opts,
	)
	if len(d.SkillFiles) > 0 {
		op.Reason = fmt.Sprintf("WARN: %d SKILL package 辅助文件被忽略，copilot .agent.md 是单文件不支持子目录", len(d.SkillFiles))
	}
	return op
}

// buildCopilotFallbackInstruction 写 skills / references fallback 文件
//
// 路径：.github/instructions/<subdir>/<n>.instructions.md，subdir 按
// FallbackSubdir 优先，缺失走 defaultFallbackSubdir("skills" / "references")。
// 保留 .instructions.md 后缀让 copilot 把它当 path-specific instruction 加载，
// applyTo 留空（不限制路径，AI 按需读取）。
//
// frontmatter：
//   - description（若 doc.Description 非空）
//   - applyTo: ""（显式空字符串，表明不限路径）
//   - std-agent-type: <type>（若 adapter.InjectStdaiTypeField=true）
//
// body 头部插入 HTML 注释 explainer（若 adapter.InjectExplainer=true，
// 受 InjectExplainerOverride 覆盖）。
func buildCopilotFallbackInstruction(d *parser.Document, adapter Adapter, cfg *config.Config) writer.FileOp {
	fullPath := copilotFallbackPath(d, adapter)

	var fm transformerutil.FmBuilder
	if d.Description != "" {
		fm.Add("description", d.Description)
	}
	fm.AddRaw("applyTo", `""`)
	if adapter.InjectStdaiTypeField {
		fm.Add("std-agent-type", string(d.Type))
	}

	body := d.Body
	if shouldInjectExplainer(d.Type, adapter) {
		body = renderExplainerHeader(d.Type, d.Name, adapter.Name) + "\n\n" + body
	}
	opts := transformerutil.MakeOpts(cfg, adapter.Name, d.Path, false)
	op := transformerutil.BuildMarkdownFile(fullPath, fm.String(), body, opts)
	// SkillFiles 非空时 copilot 无法承载（.instructions.md 是单文件），runner 收集 WARN。
	if d.Type == parser.TypeSkills && len(d.SkillFiles) > 0 {
		op.Reason = fmt.Sprintf("WARN: %d SKILL package 辅助文件被忽略，copilot .instructions.md 是单文件不支持子目录", len(d.SkillFiles))
	}
	return op
}

// copilotFallbackPath 计算 skills / references fallback 文件路径。
// 与 BuildDegradedFileOp 不同：保留 .instructions.md 后缀（copilot 方言要求）。
func copilotFallbackPath(d *parser.Document, adapter Adapter) string {
	dir := adapter.FallbackDir
	if dir == "" {
		dir = ".github/instructions"
	}
	sub := ""
	if adapter.FallbackSubdir != nil {
		if v, ok := adapter.FallbackSubdir[d.Type]; ok {
			sub = v
		} else {
			sub = defaultFallbackSubdir(d.Type)
		}
	} else {
		sub = defaultFallbackSubdir(d.Type)
	}
	if sub == "" {
		return path.Join(dir, d.Name+".instructions.md")
	}
	return path.Join(dir, sub, d.Name+".instructions.md")
}
