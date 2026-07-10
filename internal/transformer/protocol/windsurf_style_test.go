package protocol

import (
	"strings"
	"testing"

	"github.com/StringKe/std-agent/internal/config"
	"github.com/StringKe/std-agent/internal/parser"
	"github.com/StringKe/std-agent/internal/writer"
)

// windsurfTestCfg 返回不注入 marker 的最小 cfg，glossary 默认开
func windsurfTestCfg() *config.Config {
	return &config.Config{
		Inject:             false,
		InjectWhatIs:       false,
		InjectTypeGlossary: true,
	}
}

func windsurfAdapter() Adapter {
	return Adapter{
		Name:                 "windsurf",
		RulesDir:             ".windsurf/rules",
		SkillsDir:            ".windsurf/skills",
		CommandsDir:          ".windsurf/workflows",
		FallbackDir:          ".windsurf/rules",
		InjectExplainer:      true,
		InjectStdaiTypeField: true,
		InjectTypeGlossary:   true,
		RuleTriggerMode:      TriggerTrigger,
		MaxBytesPerFile:      12000,
		SkillSupportedFields: []string{"name", "description"},
	}
}

func continueLikeAdapter() Adapter {
	return Adapter{
		Name:                 "continue-dev",
		RulesDir:             ".continue/rules",
		SkillsDir:            "", // 无原生 skill -> 走 BuildDegradedSkillPackage
		CommandsDir:          ".continue/prompts",
		CommandsFileSuffix:   ".prompt.md",
		CommandFrontmatter:   []string{"name", "description", "version", "invokable"},
		FallbackDir:          ".continue/rules",
		InjectExplainer:      true,
		InjectStdaiTypeField: true,
		InjectTypeGlossary:   true,
		RuleTriggerMode:      TriggerTrigger,
	}
}

func antigravityLikeAdapter() Adapter {
	return Adapter{
		Name:                 "antigravity",
		RulesDir:             ".agents/rules",
		SkillsDir:            "", // 无原生 skill
		SkillsAsRule:         true,
		CommandsDir:          ".agents/workflows",
		FallbackDir:          ".agents/rules",
		InjectExplainer:      true,
		InjectStdaiTypeField: true,
		InjectTypeGlossary:   true,
		RuleTriggerMode:      TriggerTrigger,
		MaxBytesPerFile:      12000,
	}
}

func findWindsurfStyleOp(plan *writer.Plan, target string) *writer.FileOp {
	for i := range plan.Files {
		if plan.Files[i].Path == target {
			return &plan.Files[i]
		}
	}
	return nil
}

func windsurfStylePlanPaths(plan *writer.Plan) []string {
	out := make([]string, 0, len(plan.Files))
	for _, f := range plan.Files {
		out = append(out, f.Path)
	}
	return out
}

func TestWindsurfStyle_RuleTriggerAlwaysOn(t *testing.T) {
	doc := &parser.Document{
		Type:        parser.TypeRules,
		Name:        "core",
		AlwaysApply: true,
		Body:        "BODY",
	}
	plan, err := WindsurfStyle{}.Plan([]*parser.Document{doc}, windsurfAdapter(), windsurfTestCfg())
	if err != nil {
		t.Fatal(err)
	}
	op := findWindsurfStyleOp(plan, ".windsurf/rules/core.md")
	if op == nil {
		t.Fatalf("missing rule file; got %v", windsurfStylePlanPaths(plan))
	}
	if !strings.Contains(string(op.Content), "trigger: always_on") {
		t.Errorf("expected always_on trigger, got:\n%s", string(op.Content))
	}
}

func TestWindsurfStyle_RuleTriggerGlob(t *testing.T) {
	doc := &parser.Document{
		Type:    parser.TypeRules,
		Name:    "go-rules",
		ApplyTo: []string{"**/*.go"},
		Body:    "BODY",
	}
	plan, _ := WindsurfStyle{}.Plan([]*parser.Document{doc}, windsurfAdapter(), windsurfTestCfg())
	op := findWindsurfStyleOp(plan, ".windsurf/rules/go-rules.md")
	if op == nil {
		t.Fatalf("missing rule file; got %v", windsurfStylePlanPaths(plan))
	}
	c := string(op.Content)
	if !strings.Contains(c, "trigger: glob") {
		t.Errorf("expected glob trigger, got:\n%s", c)
	}
	if !strings.Contains(c, "**/*.go") {
		t.Errorf("expected globs entry, got:\n%s", c)
	}
}

func TestWindsurfStyle_RuleTriggerModelDecision(t *testing.T) {
	doc := &parser.Document{
		Type:        parser.TypeRules,
		Name:        "review",
		Description: "use when reviewing code",
		Body:        "BODY",
	}
	plan, _ := WindsurfStyle{}.Plan([]*parser.Document{doc}, windsurfAdapter(), windsurfTestCfg())
	op := findWindsurfStyleOp(plan, ".windsurf/rules/review.md")
	if op == nil {
		t.Fatalf("missing rule file; got %v", windsurfStylePlanPaths(plan))
	}
	c := string(op.Content)
	if !strings.Contains(c, "trigger: model_decision") {
		t.Errorf("expected model_decision trigger, got:\n%s", c)
	}
	if !strings.Contains(c, "description: use when reviewing code") {
		t.Errorf("expected description field, got:\n%s", c)
	}
}

func TestWindsurfStyle_RuleTriggerManual(t *testing.T) {
	doc := &parser.Document{
		Type: parser.TypeRules,
		Name: "fallback",
		Body: "BODY",
	}
	plan, _ := WindsurfStyle{}.Plan([]*parser.Document{doc}, windsurfAdapter(), windsurfTestCfg())
	op := findWindsurfStyleOp(plan, ".windsurf/rules/fallback.md")
	if op == nil {
		t.Fatalf("missing rule file; got %v", windsurfStylePlanPaths(plan))
	}
	if !strings.Contains(string(op.Content), "trigger: manual") {
		t.Errorf("expected manual trigger, got:\n%s", string(op.Content))
	}
}

func TestWindsurfStyle_SkillNativeWindsurfPath(t *testing.T) {
	doc := &parser.Document{
		Type:        parser.TypeSkills,
		Name:        "code-review",
		Description: "review changed code",
		License:     "MIT",
		Body:        "BODY",
	}
	plan, _ := WindsurfStyle{}.Plan([]*parser.Document{doc}, windsurfAdapter(), windsurfTestCfg())
	op := findWindsurfStyleOp(plan, ".windsurf/skills/code-review/SKILL.md")
	if op == nil {
		t.Fatalf("expected native skill; got %v", windsurfStylePlanPaths(plan))
	}
	c := string(op.Content)
	if !strings.Contains(c, "name: code-review") {
		t.Errorf("expected name frontmatter, got:\n%s", c)
	}
	// Windsurf 白名单仅 name + description；license 应被过滤
	if strings.Contains(c, "license:") {
		t.Errorf("license should be excluded by SkillSupportedFields, got:\n%s", c)
	}
}

func TestWindsurfStyle_SkillContinueDegradedSkillPackage(t *testing.T) {
	// Continue 无原生 skill，未启用 SkillsAsRule -> 走 Agent Skills 标准 fallback
	doc := &parser.Document{
		Type:        parser.TypeSkills,
		Name:        "code-review",
		Description: "review changed code",
		Body:        "BODY",
	}
	plan, _ := WindsurfStyle{}.Plan([]*parser.Document{doc}, continueLikeAdapter(), windsurfTestCfg())
	op := findWindsurfStyleOp(plan, ".continue/rules/skills/code-review/SKILL.md")
	if op == nil {
		t.Fatalf("expected degraded skill; got %v", windsurfStylePlanPaths(plan))
	}
	c := string(op.Content)
	if !strings.Contains(c, "std-agent-type: skills") {
		t.Errorf("expected std-agent-type field, got:\n%s", c)
	}
	if !strings.Contains(c, "<!-- std-agent degraded skills: code-review") {
		t.Errorf("expected explainer header, got:\n%s", c)
	}
}

func TestWindsurfStyle_SkillAntigravityAsRule(t *testing.T) {
	// Antigravity 启用 SkillsAsRule，skill 降级为 .agents/rules/skill-<name>.md
	doc := &parser.Document{
		Type:        parser.TypeSkills,
		Name:        "scout",
		Description: "scout the codebase",
		Body:        "BODY",
	}
	plan, _ := WindsurfStyle{}.Plan([]*parser.Document{doc}, antigravityLikeAdapter(), windsurfTestCfg())
	op := findWindsurfStyleOp(plan, ".agents/rules/skill-scout.md")
	if op == nil {
		t.Fatalf("expected skill-as-rule path; got %v", windsurfStylePlanPaths(plan))
	}
	c := string(op.Content)
	if !strings.Contains(c, "trigger: model_decision") {
		t.Errorf("expected model_decision trigger, got:\n%s", c)
	}
	if !strings.Contains(c, "description: scout the codebase") {
		t.Errorf("expected description, got:\n%s", c)
	}
}

func TestWindsurfStyle_CommandPathWindsurf(t *testing.T) {
	doc := &parser.Document{
		Type:        parser.TypeCommands,
		Name:        "release",
		Description: "release a patch version",
		Body:        "STEPS",
	}
	plan, _ := WindsurfStyle{}.Plan([]*parser.Document{doc}, windsurfAdapter(), windsurfTestCfg())
	op := findWindsurfStyleOp(plan, ".windsurf/workflows/release.md")
	if op == nil {
		t.Fatalf("expected windsurf workflow; got %v", windsurfStylePlanPaths(plan))
	}
	c := string(op.Content)
	// windsurf 无 frontmatter；description 应进 body 头部
	if strings.HasPrefix(c, "---") {
		t.Errorf("windsurf workflow should not have frontmatter, got:\n%s", c)
	}
	if !strings.Contains(c, "release a patch version") {
		t.Errorf("expected description prepended to body, got:\n%s", c)
	}
}

func TestWindsurfStyle_CommandPathContinuePrompt(t *testing.T) {
	doc := &parser.Document{
		Type:        parser.TypeCommands,
		Name:        "release",
		Version:     "1.0",
		Description: "release patch",
		Body:        "STEPS",
	}
	plan, _ := WindsurfStyle{}.Plan([]*parser.Document{doc}, continueLikeAdapter(), windsurfTestCfg())
	op := findWindsurfStyleOp(plan, ".continue/prompts/release.prompt.md")
	if op == nil {
		t.Fatalf("expected continue prompt; got %v", windsurfStylePlanPaths(plan))
	}
	c := string(op.Content)
	if !strings.Contains(c, "name: release") {
		t.Errorf("expected name field, got:\n%s", c)
	}
	if !strings.Contains(c, "invokable: true") {
		t.Errorf("expected invokable=true, got:\n%s", c)
	}
	if !strings.Contains(c, "version: ") {
		t.Errorf("expected version field, got:\n%s", c)
	}
}

func TestWindsurfStyle_CommandPathAntigravity(t *testing.T) {
	doc := &parser.Document{
		Type:        parser.TypeCommands,
		Name:        "release",
		Description: "release",
		Body:        "STEPS",
	}
	plan, _ := WindsurfStyle{}.Plan([]*parser.Document{doc}, antigravityLikeAdapter(), windsurfTestCfg())
	if findWindsurfStyleOp(plan, ".agents/workflows/release.md") == nil {
		t.Fatalf("expected antigravity workflow; got %v", windsurfStylePlanPaths(plan))
	}
}

func TestWindsurfStyle_ReferencesFallbackSubdir(t *testing.T) {
	doc := &parser.Document{
		Type:        parser.TypeReferences,
		Name:        "transformer-design",
		Description: "transformer architecture notes",
		Body:        "REF BODY",
	}
	plan, _ := WindsurfStyle{}.Plan([]*parser.Document{doc}, windsurfAdapter(), windsurfTestCfg())
	wantPath := ".windsurf/rules/references/transformer-design.md"
	op := findWindsurfStyleOp(plan, wantPath)
	if op == nil {
		t.Fatalf("expected references fallback %q; got %v", wantPath, windsurfStylePlanPaths(plan))
	}
	c := string(op.Content)
	if !strings.Contains(c, "std-agent-type: references") {
		t.Errorf("expected std-agent-type frontmatter, got:\n%s", c)
	}
	if !strings.Contains(c, "<!-- std-agent degraded references: transformer-design") {
		t.Errorf("expected explainer, got:\n%s", c)
	}
	for _, forbidden := range []string{"_ref_", "_skill_", "_subagent_"} {
		if strings.Contains(wantPath, forbidden) {
			t.Errorf("path %q must not contain forbidden prefix %q", wantPath, forbidden)
		}
	}
}

func TestWindsurfStyle_SubagentsFallbackSubdir(t *testing.T) {
	doc := &parser.Document{
		Type:        parser.TypeSubagents,
		Name:        "reviewer",
		Description: "code reviewer subagent",
		Body:        "SUB BODY",
	}
	plan, _ := WindsurfStyle{}.Plan([]*parser.Document{doc}, antigravityLikeAdapter(), windsurfTestCfg())
	wantPath := ".agents/rules/subagents/reviewer.md"
	op := findWindsurfStyleOp(plan, wantPath)
	if op == nil {
		t.Fatalf("expected subagents fallback %q; got %v", wantPath, windsurfStylePlanPaths(plan))
	}
	c := string(op.Content)
	if !strings.Contains(c, "std-agent-type: subagents") {
		t.Errorf("expected std-agent-type frontmatter, got:\n%s", c)
	}
}

func TestWindsurfStyle_GlossaryLandsInRulesDir(t *testing.T) {
	plan, _ := WindsurfStyle{}.Plan([]*parser.Document{
		{Type: parser.TypeRules, Name: "x", Body: "B"},
	}, windsurfAdapter(), windsurfTestCfg())
	op := findWindsurfStyleOp(plan, ".windsurf/rules/glossary.md")
	if op == nil {
		t.Fatalf("expected glossary; got %v", windsurfStylePlanPaths(plan))
	}
	c := string(op.Content)
	if !strings.Contains(c, "std-agent 类型速查") {
		t.Errorf("expected glossary content, got:\n%s", c)
	}
	if !strings.Contains(c, "std-agent-type: glossary") {
		t.Errorf("expected std-agent-type glossary frontmatter, got:\n%s", c)
	}
}

func TestWindsurfStyle_GlossarySkippedWhenDisabled(t *testing.T) {
	cfg := windsurfTestCfg()
	cfg.InjectTypeGlossary = false
	plan, _ := WindsurfStyle{}.Plan([]*parser.Document{
		{Type: parser.TypeRules, Name: "x", Body: "B"},
	}, windsurfAdapter(), cfg)
	if findWindsurfStyleOp(plan, ".windsurf/rules/glossary.md") != nil {
		t.Error("cfg.InjectTypeGlossary=false should suppress glossary")
	}
}

func TestWindsurfStyle_DisabledAdapter(t *testing.T) {
	adapter := windsurfAdapter()
	adapter.Disabled = true
	plan, _ := WindsurfStyle{}.Plan([]*parser.Document{
		{Type: parser.TypeRules, Name: "x", Body: "B"},
	}, adapter, windsurfTestCfg())
	if len(plan.Files) != 0 {
		t.Errorf("Disabled adapter should produce 0 files, got %d", len(plan.Files))
	}
}

func TestWindsurfStyle_EmptyDocs(t *testing.T) {
	plan, _ := WindsurfStyle{}.Plan(nil, windsurfAdapter(), windsurfTestCfg())
	if len(plan.Files) != 0 {
		t.Errorf("empty docs should produce 0 files, got %d", len(plan.Files))
	}
}

func TestWindsurfStyle_RuleMaxBytesWarn(t *testing.T) {
	adapter := windsurfAdapter()
	adapter.MaxBytesPerFile = 50
	doc := &parser.Document{
		Type: parser.TypeRules,
		Name: "big",
		Body: strings.Repeat("BIG ", 200),
	}
	plan, _ := WindsurfStyle{}.Plan([]*parser.Document{doc}, adapter, windsurfTestCfg())
	op := findWindsurfStyleOp(plan, ".windsurf/rules/big.md")
	if op == nil {
		t.Fatalf("missing rule file; got %v", windsurfStylePlanPaths(plan))
	}
	if !strings.Contains(op.Reason, "exceeds 50 chars") {
		t.Errorf("expected MaxBytes warn reason, got %q", op.Reason)
	}
}
