package protocol

import (
	"encoding/json"
	"path"
	"strings"

	"github.com/StringKe/std-agent/internal/config"
	"github.com/StringKe/std-agent/internal/parser"
	"github.com/StringKe/std-agent/internal/transformer/transformerutil"
	"github.com/StringKe/std-agent/internal/writer"
)

// Cursor 是 Cursor IDE 的协议族实现。
//
// 原生支持：rules（.cursor/rules/<name>.mdc 后缀） / skills（Agent Skills 标准
// .cursor/skills/<name>/SKILL.md） / commands（.cursor/commands/<name>.md） /
// MCP（.cursor/mcp.json 顶层 mcpServers）。
//
// subagents：Cursor 已原生支持 .cursor/agents/<name>.md（frontmatter name /
// description / model，另有 readonly / is_background 可选字段），SubagentsDir
// 非空时走原生落点，为空时降级到 .cursor/rules/subagents/<name>.md。
//
// 不原生支持的 type 走 graceful degradation：
//   - references -> .cursor/rules/references/<name>.md（子目录隔离 + frontmatter std-agent-type: references）
//
// 方言要点（与 ClaudeMD / AgentsMD 不同）：
//   - 文件后缀 .mdc（不是 .md）—rules 类专属
//   - globs frontmatter 是逗号分隔字符串（GlobsCommaString），不是 YAML list
//   - alwaysApply frontmatter 是 bool
//   - Cursor 无单一根文件；InjectTypeGlossary=true 时把 glossary 落到
//     .cursor/rules/glossary.md（无前缀，靠子目录隔离 + frontmatter std-agent-type: glossary）
type Cursor struct{}

// Plan 按 type 分桶生成 FileOp。
func (c Cursor) Plan(docs []*parser.Document, adapter Adapter, cfg *config.Config) (*writer.Plan, error) {
	plan := &writer.Plan{Target: adapter.Name}
	if adapter.Disabled {
		return plan, nil
	}

	if cfg.MCP != nil && len(cfg.MCP.Servers) > 0 && adapter.MCPPath != "" {
		plan.Files = append(plan.Files, c.buildMCPJSON(cfg.MCP, adapter))
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

	if g := c.buildGlossary(adapter, cfg); g != nil {
		plan.Files = append(plan.Files, *g)
	}

	for _, d := range rules {
		plan.Files = append(plan.Files, c.buildRule(d, adapter, cfg))
	}
	for _, d := range skills {
		plan.Files = append(plan.Files, c.buildSkill(d, adapter, cfg)...)
	}
	for _, d := range commands {
		plan.Files = append(plan.Files, c.buildCommand(d, adapter, cfg))
	}
	// references：走独立子目录 .cursor/references/<name>.md（v3：不借 skills 路径，
	// 否则 cursor 会按 skill 触发逻辑加载 references 破坏"按需查阅"语义）
	for _, d := range references {
		plan.Files = append(plan.Files, BuildDegradedFileOp(d, adapter, cfg))
	}
	// subagents：SubagentsDir 非空走原生 .cursor/agents/<name>.md，否则路径降级
	for _, d := range subagents {
		if adapter.SubagentsDir != "" {
			plan.Files = append(plan.Files, c.buildSubagent(d, adapter, cfg))
		} else {
			plan.Files = append(plan.Files, BuildDegradedFileOp(d, adapter, cfg))
		}
	}
	return plan, nil
}

// buildSubagent 生成 .cursor/agents/<name>.md（Cursor 原生 subagent）
//
// frontmatter：name / description / model / readonly / is_background
// （https://cursor.com/docs/subagents.md）。body 是 subagent 系统提示词。
func (c Cursor) buildSubagent(d *parser.Document, adapter Adapter, cfg *config.Config) writer.FileOp {
	var fm transformerutil.FmBuilder
	fm.Add("name", d.Name)
	fm.Add("description", transformerutil.MergeDescription(d.Description, d.WhenToUse))
	fm.Add("model", d.Model)
	if d.ReadOnly {
		fm.AddBool("readonly", true)
	}
	if d.Background {
		fm.AddBool("is_background", true)
	}
	opts := transformerutil.MakeOpts(cfg, adapter.Name, d.Path, false)
	return transformerutil.BuildMarkdownFile(
		transformerutil.FilePath(adapter.SubagentsDir, d.Name, ".md"),
		fm.String(), d.Body, opts,
	)
}

// buildMCPJSON 生成 .cursor/mcp.json，顶层键固定 mcpServers
func (c Cursor) buildMCPJSON(mcp *config.MCPConfig, adapter Adapter) writer.FileOp {
	topKey := adapter.MCPTopKey
	if topKey == "" {
		topKey = "mcpServers"
	}
	wrapper := map[string]map[string]config.MCPServer{topKey: mcp.Servers}
	body, _ := json.MarshalIndent(wrapper, "", "  ")
	body = append(body, '\n')
	return writer.FileOp{Path: adapter.MCPPath, Content: body}
}

// buildRule 生成 .cursor/rules/<name>.mdc，含 alwaysApply / globs（逗号分隔字符串）/ description frontmatter
func (c Cursor) buildRule(d *parser.Document, adapter Adapter, cfg *config.Config) writer.FileOp {
	var fm transformerutil.FmBuilder
	if adapter.SupportsAlwaysApply && d.AlwaysApply {
		fm.AddBool("alwaysApply", true)
	}
	applyTo := transformerutil.EffectiveApplyTo(d, adapter.Name)
	if adapter.GlobsFieldName != "" && len(applyTo) > 0 {
		// Cursor globs 是逗号分隔字符串
		fm.Add(adapter.GlobsFieldName, joinComma(applyTo))
	}
	if adapter.SupportsDescription {
		fm.Add("description", d.Description)
	}
	opts := transformerutil.MakeOpts(cfg, adapter.Name, d.Path, false)
	return transformerutil.BuildMarkdownFile(
		transformerutil.FilePath(adapter.RulesDir, d.Name, ".mdc"),
		fm.String(), d.Body, opts,
	)
}

// buildSkill 生成 .cursor/skills/<name>/SKILL.md（Agent Skills 标准）
func (c Cursor) buildSkill(d *parser.Document, adapter Adapter, cfg *config.Config) []writer.FileOp {
	skillDir := transformerutil.SkillDir(adapter.SkillsDir, d.Name)
	var fm transformerutil.FmBuilder
	fm.Add("name", d.Name)
	fm.Add("description", transformerutil.MergeDescription(d.Description, d.WhenToUse))
	applyTo := transformerutil.EffectiveApplyTo(d, adapter.Name)
	if len(applyTo) > 0 {
		fm.AddList("paths", applyTo)
	}
	if d.DisableModelInvocation {
		fm.AddBool("disable-model-invocation", true)
	}
	fm.Add("license", d.License)
	fm.Add("compatibility", d.Compatibility)
	fm.AddMap("metadata", d.Metadata)
	opts := transformerutil.MakeOpts(cfg, adapter.Name, d.Path, false)
	skillMd := transformerutil.BuildMarkdownFile(skillDir+"/SKILL.md", fm.String(), d.Body, opts)
	return transformerutil.BuildSkillPackage(skillDir, skillMd, d.SkillFiles)
}

// buildCommand 生成 .cursor/commands/<name>.md
func (c Cursor) buildCommand(d *parser.Document, adapter Adapter, cfg *config.Config) writer.FileOp {
	opts := transformerutil.MakeOpts(cfg, adapter.Name, d.Path, false)
	body := d.Body
	if d.Description != "" {
		body = d.Description + "\n\n" + d.Body
	}
	return transformerutil.BuildMarkdownFile(
		transformerutil.FilePath(adapter.CommandsDir, d.Name, ".md"),
		"", body, opts,
	)
}

// buildGlossary 在 InjectTypeGlossary=true 时把 type glossary 落到
// .cursor/rules/glossary.md（无下划线前缀，靠子目录隔离 + frontmatter std-agent-type: glossary）。
// Cursor 无单一根文件，glossary 不能 inline；落点选 RulesDir 顶层。
//
// 返回 nil 表示不输出 glossary 文件。
func (c Cursor) buildGlossary(adapter Adapter, cfg *config.Config) *writer.FileOp {
	if !adapter.InjectTypeGlossary {
		return nil
	}
	body := RenderGlossaryFor(adapter)
	if body == "" {
		return nil
	}
	var fm transformerutil.FmBuilder
	fm.Add("std-agent-type", "glossary")
	opts := transformerutil.MakeOpts(cfg, adapter.Name, "", false)
	op := transformerutil.BuildMarkdownFile(
		path.Join(adapter.RulesDir, "glossary.md"),
		fm.String(), body, opts,
	)
	return &op
}

// joinComma 把字符串 slice 用逗号拼接。
// 不复用 transformerutil.CommaJoin 仅为本协议方言显式化，未来可能切换分隔符。
func joinComma(items []string) string {
	return strings.Join(items, ",")
}
