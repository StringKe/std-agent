package protocol

import (
	"strings"
	"testing"

	"std-ai/internal/config"
	"std-ai/internal/parser"
	"std-ai/internal/writer"
)

// agentsMDTestCfg 返回不注入 marker 的最小 cfg，让断言只关心 body 内容
func agentsMDTestCfg() *config.Config {
	return &config.Config{Inject: false, InjectWhatIs: false}
}

// findAgentsMDOp 在 plan.Files 里找匹配 path 的 FileOp；找不到返回 nil
func findAgentsMDOp(plan *writer.Plan, p string) *writer.FileOp {
	for i := range plan.Files {
		if plan.Files[i].Path == p {
			return &plan.Files[i]
		}
	}
	return nil
}

// agentsMDPlanPaths 提取 plan 中所有 op path，便于断言诊断
func agentsMDPlanPaths(plan *writer.Plan) []string {
	out := make([]string, 0, len(plan.Files))
	for _, f := range plan.Files {
		out = append(out, f.Path)
	}
	return out
}

// codexLikeAdapter 是 codex 风格 adapter 的 fixture
func codexLikeAdapter() Adapter {
	return Adapter{
		Name:                  "codex",
		RootFileName:          "AGENTS.md",
		ManifestSection:       "Reference Rules",
		NestedSupported:       true,
		RulesDir:              ".codex/memories",
		SkillsDir:             ".agents/skills",
		FallbackDir:           ".codex/memories",
		InjectExplainer:       true,
		InjectStdaiTypeField:  true,
		InjectTypeGlossary:    false,
		SkillSupportedFields:  []string{"name", "description", "license", "compatibility", "metadata"},
		CommandsAsSkillPrefix: "cmd-",
		InjectCommandsToRoot:  true,
	}
}

// TestAgentsMD_DisabledReturnsEmptyPlan：Disabled=true 时不输出任何 FileOp
func TestAgentsMD_DisabledReturnsEmptyPlan(t *testing.T) {
	docs := []*parser.Document{{Type: parser.TypeRules, Name: "r", Body: "B"}}
	plan, err := AgentsMD{}.Plan(docs, Adapter{Name: "x", Disabled: true}, agentsMDTestCfg())
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(plan.Files) != 0 {
		t.Errorf("Disabled=true should produce 0 files, got %d: %v", len(plan.Files), agentsMDPlanPaths(plan))
	}
	if plan.Target != "x" {
		t.Errorf("Target=%q, want %q", plan.Target, "x")
	}
}

// TestAgentsMD_EmptyDocsReturnsEmptyPlan：无 docs 时返回空 Plan
func TestAgentsMD_EmptyDocsReturnsEmptyPlan(t *testing.T) {
	plan, err := AgentsMD{}.Plan(nil, codexLikeAdapter(), agentsMDTestCfg())
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(plan.Files) != 0 {
		t.Errorf("empty docs should produce 0 files, got %v", agentsMDPlanPaths(plan))
	}
}

// TestAgentsMD_RootBodyWithManifest：root rule + nonRoot rule 拼出根文件 + manifest 段
func TestAgentsMD_RootBodyWithManifest(t *testing.T) {
	docs := []*parser.Document{
		{Type: parser.TypeRules, Name: "root", Root: true, Body: "# Project\n\nProject overview."},
		{Type: parser.TypeRules, Name: "style", Description: "code style", Body: "use tabs"},
	}
	plan, err := AgentsMD{}.Plan(docs, codexLikeAdapter(), agentsMDTestCfg())
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	root := findAgentsMDOp(plan, "AGENTS.md")
	if root == nil {
		t.Fatalf("AGENTS.md not produced, paths=%v", agentsMDPlanPaths(plan))
	}
	if !root.IsRoot {
		t.Error("AGENTS.md FileOp.IsRoot should be true")
	}
	content := string(root.Content)
	if !strings.Contains(content, "Project overview.") {
		t.Errorf("root body missing project overview:\n%s", content)
	}
	if !strings.Contains(content, "## Reference Rules") {
		t.Errorf("manifest section title missing:\n%s", content)
	}
	if !strings.Contains(content, ".codex/memories/style.md") {
		t.Errorf("manifest entry missing:\n%s", content)
	}
	// nonRoot rule fan-out 到 RulesDir
	if findAgentsMDOp(plan, ".codex/memories/style.md") == nil {
		t.Errorf("nonRoot rule fan-out missing, paths=%v", agentsMDPlanPaths(plan))
	}
}

// TestAgentsMD_GlossaryPrependedToRoot：InjectTypeGlossary=true 时 root 头部含 glossary 段
func TestAgentsMD_GlossaryPrependedToRoot(t *testing.T) {
	a := codexLikeAdapter()
	a.InjectTypeGlossary = true
	docs := []*parser.Document{
		{Type: parser.TypeRules, Name: "root", Root: true, Body: "PROJECT_OVERVIEW"},
	}
	plan, err := AgentsMD{}.Plan(docs, a, agentsMDTestCfg())
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	root := findAgentsMDOp(plan, "AGENTS.md")
	if root == nil {
		t.Fatalf("root not found, paths=%v", agentsMDPlanPaths(plan))
	}
	content := string(root.Content)
	if !strings.Contains(content, "std-ai type glossary auto-injected") {
		t.Errorf("expected glossary marker, got:\n%s", content)
	}
	// glossary 必须在 project overview 之前
	gIdx := strings.Index(content, "std-ai type glossary")
	pIdx := strings.Index(content, "PROJECT_OVERVIEW")
	if gIdx < 0 || pIdx < 0 || gIdx >= pIdx {
		t.Errorf("glossary should appear before project overview, gIdx=%d pIdx=%d", gIdx, pIdx)
	}
}

// TestAgentsMD_NestedRootNoManifestNoGlossary：nested root 文件不含 manifest / glossary
func TestAgentsMD_NestedRootNoManifestNoGlossary(t *testing.T) {
	a := codexLikeAdapter()
	a.InjectTypeGlossary = true
	docs := []*parser.Document{
		{
			Type:       parser.TypeRules,
			Name:       "subroot",
			NestedPath: "service/api",
			Body:       "Subdir context",
		},
	}
	plan, err := AgentsMD{}.Plan(docs, a, agentsMDTestCfg())
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	nested := findAgentsMDOp(plan, "service/api/AGENTS.md")
	if nested == nil {
		t.Fatalf("nested AGENTS.md not produced, paths=%v", agentsMDPlanPaths(plan))
	}
	content := string(nested.Content)
	if strings.Contains(content, "std-ai type glossary") {
		t.Errorf("nested root should NOT contain glossary, got:\n%s", content)
	}
	if strings.Contains(content, "## Reference Rules") {
		t.Errorf("nested root should NOT contain manifest, got:\n%s", content)
	}
	if !strings.Contains(content, "Subdir context") {
		t.Errorf("nested root body missing, got:\n%s", content)
	}
}

// TestAgentsMD_SkillNativePath：SkillsDir 非空时走原生 Agent Skills 路径
func TestAgentsMD_SkillNativePath(t *testing.T) {
	docs := []*parser.Document{
		{Type: parser.TypeSkills, Name: "code-review", Description: "Review code", Body: "skill body"},
	}
	plan, err := AgentsMD{}.Plan(docs, codexLikeAdapter(), agentsMDTestCfg())
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	op := findAgentsMDOp(plan, ".agents/skills/code-review/SKILL.md")
	if op == nil {
		t.Fatalf("native skill path missing, paths=%v", agentsMDPlanPaths(plan))
	}
	content := string(op.Content)
	if !strings.Contains(content, "name: code-review") {
		t.Errorf("expected name frontmatter, got:\n%s", content)
	}
	// 不应有 std-ai-type 字段（原生路径）
	if strings.Contains(content, "std-ai-type:") {
		t.Errorf("native skill should NOT have std-ai-type field, got:\n%s", content)
	}
}

// TestAgentsMD_SkillFallbackPath：SkillsDir 为空时走 BuildDegradedSkillPackage
func TestAgentsMD_SkillFallbackPath(t *testing.T) {
	a := Adapter{
		Name:                 "amp",
		RootFileName:         "AGENTS.md",
		RulesDir:             "",
		SkillsDir:            "",
		FallbackDir:          ".amp/rules",
		InjectExplainer:      true,
		InjectStdaiTypeField: true,
	}
	docs := []*parser.Document{
		{Type: parser.TypeSkills, Name: "x", Body: "B"},
	}
	plan, err := AgentsMD{}.Plan(docs, a, agentsMDTestCfg())
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	op := findAgentsMDOp(plan, ".amp/rules/skills/x/SKILL.md")
	if op == nil {
		t.Fatalf("fallback skill path missing, paths=%v", agentsMDPlanPaths(plan))
	}
	content := string(op.Content)
	if !strings.Contains(content, "std-ai-type: skills") {
		t.Errorf("fallback skill should contain std-ai-type, got:\n%s", content)
	}
	if !strings.Contains(content, "<!-- std-ai degraded skills:") {
		t.Errorf("fallback skill should contain explainer, got:\n%s", content)
	}
}

// TestAgentsMD_CommandsInjectedToRoot：InjectCommandsToRoot=true 时 commands 段并入 root
func TestAgentsMD_CommandsInjectedToRoot(t *testing.T) {
	a := codexLikeAdapter()
	a.InjectCommandsToRoot = true
	a.CommandsAsSkillPrefix = "" // 关掉 skill prefix 才会走 inject 分支
	docs := []*parser.Document{
		{Type: parser.TypeRules, Name: "root", Root: true, Body: "PROJECT"},
		{Type: parser.TypeCommands, Name: "release", Description: "Release patch", Body: "release steps"},
	}
	cfg := &config.Config{Inject: true, InjectWhatIs: false}
	plan, err := AgentsMD{}.Plan(docs, a, cfg)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	root := findAgentsMDOp(plan, "AGENTS.md")
	if root == nil {
		t.Fatalf("AGENTS.md missing, paths=%v", agentsMDPlanPaths(plan))
	}
	content := string(root.Content)
	if !strings.Contains(content, "## Slash Commands") {
		t.Errorf("commands not injected to root, got:\n%s", content)
	}
	if !strings.Contains(content, "### /release") {
		t.Errorf("command entry missing, got:\n%s", content)
	}
	// 验证 inject 发生在 footer marker 之前
	mIdx := strings.LastIndex(content, "<!-- /Generated by stdagent -->")
	cIdx := strings.Index(content, "## Slash Commands")
	if mIdx >= 0 && cIdx > mIdx {
		t.Errorf("commands should be injected BEFORE footer marker, cIdx=%d mIdx=%d", cIdx, mIdx)
	}
	// 不应该有独立 command 文件
	if findAgentsMDOp(plan, ".codex/memories/commands/release.md") != nil {
		t.Errorf("InjectCommandsToRoot=true should not produce independent command file")
	}
}

// TestAgentsMD_CommandsAsSkillPrefix：CommandsAsSkillPrefix 非空时 command 降级为 skill
func TestAgentsMD_CommandsAsSkillPrefix(t *testing.T) {
	a := codexLikeAdapter()
	a.InjectCommandsToRoot = false // 关掉 inject 才走 skill prefix
	docs := []*parser.Document{
		{Type: parser.TypeCommands, Name: "release", Description: "Release patch", Body: "steps"},
	}
	plan, err := AgentsMD{}.Plan(docs, a, agentsMDTestCfg())
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	op := findAgentsMDOp(plan, ".agents/skills/cmd-release/SKILL.md")
	if op == nil {
		t.Fatalf("command-as-skill path missing, paths=%v", agentsMDPlanPaths(plan))
	}
	content := string(op.Content)
	if !strings.Contains(content, "name: cmd-release") {
		t.Errorf("expected name=cmd-release frontmatter, got:\n%s", content)
	}
	if !strings.Contains(content, "/release") {
		t.Errorf("description should mention slash invocation, got:\n%s", content)
	}
}

// TestAgentsMD_ReferencesFallbackNoPrefix：references fallback 路径无 _ref_ 前缀
func TestAgentsMD_ReferencesFallbackNoPrefix(t *testing.T) {
	a := codexLikeAdapter()
	a.SkillsDir = "" // 强制走 BuildDegradedFileOp（rule-equivalent），不走 SKILL package
	docs := []*parser.Document{
		{Type: parser.TypeReferences, Name: "design", Body: "background notes"},
	}
	plan, err := AgentsMD{}.Plan(docs, a, agentsMDTestCfg())
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	op := findAgentsMDOp(plan, ".codex/memories/references/design.md")
	if op == nil {
		t.Fatalf("references fallback path missing, paths=%v", agentsMDPlanPaths(plan))
	}
	for _, bad := range []string{"_ref_", "_skill_", "_command_", "_subagent_"} {
		if strings.Contains(op.Path, bad) {
			t.Errorf("path %q should not contain forbidden prefix %q", op.Path, bad)
		}
	}
	content := string(op.Content)
	if !strings.Contains(content, "std-ai-type: references") {
		t.Errorf("expected std-ai-type=references, got:\n%s", content)
	}
}

// TestAgentsMD_ReferencesAgentSkillsFallback：SkillsDir 非空时 references fallback 走 Agent Skills 包
func TestAgentsMD_ReferencesAgentSkillsFallback(t *testing.T) {
	a := codexLikeAdapter()
	// ReferencesDir 空 + SkillsDir 非空 -> BuildDegradedSkillPackage
	docs := []*parser.Document{
		{Type: parser.TypeReferences, Name: "design", Description: "design notes", Body: "B"},
	}
	plan, err := AgentsMD{}.Plan(docs, a, agentsMDTestCfg())
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	op := findAgentsMDOp(plan, ".agents/skills/design/SKILL.md")
	if op == nil {
		t.Fatalf("Agent Skills references fallback missing, paths=%v", agentsMDPlanPaths(plan))
	}
	content := string(op.Content)
	if !strings.Contains(content, "std-ai-type: references") {
		t.Errorf("expected std-ai-type=references, got:\n%s", content)
	}
}

// TestAgentsMD_SubagentsFallbackNoPrefix：subagents fallback 路径无 _subagent_ 前缀
func TestAgentsMD_SubagentsFallbackNoPrefix(t *testing.T) {
	a := codexLikeAdapter()
	docs := []*parser.Document{
		{Type: parser.TypeSubagents, Name: "reviewer", Body: "subagent body"},
	}
	plan, err := AgentsMD{}.Plan(docs, a, agentsMDTestCfg())
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	op := findAgentsMDOp(plan, ".codex/memories/subagents/reviewer.md")
	if op == nil {
		t.Fatalf("subagents fallback path missing, paths=%v", agentsMDPlanPaths(plan))
	}
	for _, bad := range []string{"_subagent_", "_skill_", "_command_"} {
		if strings.Contains(op.Path, bad) {
			t.Errorf("path %q should not contain forbidden prefix %q", op.Path, bad)
		}
	}
}

// TestAgentsMD_SubagentNativeDir：SubagentsDir 非空时走原生路径
func TestAgentsMD_SubagentNativeDir(t *testing.T) {
	a := codexLikeAdapter()
	a.SubagentsDir = ".opencode/agents"
	docs := []*parser.Document{
		{Type: parser.TypeSubagents, Name: "reviewer", Description: "code reviewer", Body: "B"},
	}
	plan, err := AgentsMD{}.Plan(docs, a, agentsMDTestCfg())
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	op := findAgentsMDOp(plan, ".opencode/agents/reviewer.md")
	if op == nil {
		t.Fatalf("native subagent path missing, paths=%v", agentsMDPlanPaths(plan))
	}
	content := string(op.Content)
	if !strings.Contains(content, "name: reviewer") {
		t.Errorf("expected name frontmatter, got:\n%s", content)
	}
}

// TestAgentsMD_SubagentCLIFallback：SubagentInvokeCmd 非空时 body 含 shell 调用指引
func TestAgentsMD_SubagentCLIFallback(t *testing.T) {
	a := Adapter{
		Name:              "copilot",
		RootFileName:      "AGENTS.md",
		RulesDir:          ".github/instructions",
		FallbackDir:       ".github/instructions",
		InjectExplainer:   true,
		SubagentInvokeCmd: "claude --agent {name}",
	}
	docs := []*parser.Document{
		{Type: parser.TypeSubagents, Name: "reviewer", Description: "code reviewer", Body: "instructions"},
	}
	plan, err := AgentsMD{}.Plan(docs, a, agentsMDTestCfg())
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	op := findAgentsMDOp(plan, ".github/instructions/subagents/reviewer.md")
	if op == nil {
		t.Fatalf("subagent CLI fallback path missing, paths=%v", agentsMDPlanPaths(plan))
	}
	content := string(op.Content)
	if !strings.Contains(content, "claude --agent reviewer") {
		t.Errorf("expected CLI invocation with substituted name, got:\n%s", content)
	}
	if !strings.Contains(content, "```bash") {
		t.Errorf("expected fenced shell block, got:\n%s", content)
	}
}

// TestAgentsMD_RuleTriggerMode：RuleTriggerMode=Trigger 时 nonRoot rule 写 trigger frontmatter
func TestAgentsMD_RuleTriggerMode(t *testing.T) {
	a := Adapter{
		Name:            "antigravity",
		RootFileName:    "AGENTS.md",
		RulesDir:        ".agents/rules",
		RuleTriggerMode: TriggerTrigger,
	}
	docs := []*parser.Document{
		{Type: parser.TypeRules, Name: "always", AlwaysApply: true, Body: "B"},
		{Type: parser.TypeRules, Name: "glob-r", ApplyTo: []string{"**/*.go"}, Body: "B"},
		{Type: parser.TypeRules, Name: "model", Description: "use when reviewing", Body: "B"},
		{Type: parser.TypeRules, Name: "manual-r", Body: "B"},
	}
	plan, err := AgentsMD{}.Plan(docs, a, agentsMDTestCfg())
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	cases := map[string]string{
		".agents/rules/always.md":   "trigger: always_on",
		".agents/rules/glob-r.md":   "trigger: glob",
		".agents/rules/model.md":    "trigger: model_decision",
		".agents/rules/manual-r.md": "trigger: manual",
	}
	for p, want := range cases {
		op := findAgentsMDOp(plan, p)
		if op == nil {
			t.Errorf("rule file %q missing, paths=%v", p, agentsMDPlanPaths(plan))
			continue
		}
		if !strings.Contains(string(op.Content), want) {
			t.Errorf("rule %q expected to contain %q, got:\n%s", p, want, op.Content)
		}
	}
}

// TestAgentsMD_SkillsAsRule：SkillsAsRule=true 时 skill 降级为 rule
func TestAgentsMD_SkillsAsRule(t *testing.T) {
	a := Adapter{
		Name:         "antigravity",
		RootFileName: "AGENTS.md",
		RulesDir:     ".agents/rules",
		SkillsAsRule: true,
	}
	docs := []*parser.Document{
		{Type: parser.TypeSkills, Name: "review", Description: "review skill", Body: "B"},
	}
	plan, err := AgentsMD{}.Plan(docs, a, agentsMDTestCfg())
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	op := findAgentsMDOp(plan, ".agents/rules/skill-review.md")
	if op == nil {
		t.Fatalf("skill-as-rule path missing, paths=%v", agentsMDPlanPaths(plan))
	}
	content := string(op.Content)
	if !strings.Contains(content, "trigger: model_decision") {
		t.Errorf("expected trigger: model_decision, got:\n%s", content)
	}
}

// TestAgentsMD_SpilloverInlineRulesOversize：RulesDir 为空 + MaxBytesPerFile>0 + 超阈值 rule -> spill 到独立文件
func TestAgentsMD_SpilloverInlineRulesOversize(t *testing.T) {
	a := Adapter{
		Name:            "amp",
		RootFileName:    "AGENTS.md",
		RulesDir:        "", // 强制 inline
		FallbackDir:     ".amp/rules",
		MaxBytesPerFile: 100,
	}
	bigBody := strings.Repeat("X", 500)
	docs := []*parser.Document{
		{Type: parser.TypeRules, Name: "big", Body: bigBody, BodyBytes: len(bigBody)},
		{Type: parser.TypeRules, Name: "small", Body: "tiny", BodyBytes: 4},
	}
	plan, err := AgentsMD{}.Plan(docs, a, agentsMDTestCfg())
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	// big 应该被 spill 到独立文件
	bigOp := findAgentsMDOp(plan, ".amp/rules/big.md")
	if bigOp == nil {
		t.Fatalf("oversized rule should spill to standalone file, paths=%v", agentsMDPlanPaths(plan))
	}
	if !strings.Contains(string(bigOp.Content), "XXX") {
		t.Errorf("spilled rule should contain body, got prefix=%q", string(bigOp.Content)[:50])
	}
	// small 应该 inline 到 root
	root := findAgentsMDOp(plan, "AGENTS.md")
	if root == nil {
		t.Fatalf("root missing")
	}
	rootContent := string(root.Content)
	if !strings.Contains(rootContent, "tiny") {
		t.Errorf("small rule should inline in root, got:\n%s", rootContent)
	}
	if strings.Contains(rootContent, "XXXXXX") {
		t.Errorf("oversized rule should NOT inline in root")
	}
}

// TestAgentsMD_NoRootFileNameSkipsRoot：RootFileName="" 时不写根文件
func TestAgentsMD_NoRootFileNameSkipsRoot(t *testing.T) {
	a := Adapter{
		Name:         "weird",
		RootFileName: "",
		RulesDir:     ".weird/rules",
	}
	docs := []*parser.Document{
		{Type: parser.TypeRules, Name: "r", Body: "B"},
	}
	plan, err := AgentsMD{}.Plan(docs, a, agentsMDTestCfg())
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	for _, op := range plan.Files {
		if op.IsRoot {
			t.Errorf("RootFileName=\"\" should produce no IsRoot file, got %q", op.Path)
		}
	}
}

// TestAgentsMD_BuildCommandsSection：辅助函数渲染输出含期望段落
func TestAgentsMD_BuildCommandsSection(t *testing.T) {
	cmds := []*parser.Document{
		{Name: "release", Description: "Release patch", Body: "do step"},
	}
	got := BuildCommandsSection(cmds)
	if !strings.Contains(got, "## Slash Commands") {
		t.Errorf("missing title, got:\n%s", got)
	}
	if !strings.Contains(got, "### /release") {
		t.Errorf("missing command heading, got:\n%s", got)
	}
}

// TestAgentsMD_InjectBeforeFooter：marker 存在则插入到 marker 之前；不存在则追加
func TestAgentsMD_InjectBeforeFooter(t *testing.T) {
	withMarker := []byte("HEAD\n<!-- /Generated by stdagent -->\n")
	got := InjectBeforeFooter(withMarker, "INJECT")
	if !strings.Contains(string(got), "INJECT<!-- /Generated by stdagent -->") {
		t.Errorf("expected INJECT before marker, got:\n%s", got)
	}

	noMarker := []byte("HEAD\n")
	got2 := InjectBeforeFooter(noMarker, "INJECT")
	if string(got2) != "HEAD\nINJECT" {
		t.Errorf("no marker should append, got %q", got2)
	}
}
