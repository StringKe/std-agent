package transformer

import (
	"strings"
	"testing"

	"github.com/StringKe/std-agent/internal/config"
	"github.com/StringKe/std-agent/internal/parser"
)

func TestCodexAGENTSMd(t *testing.T) {
	tr := &Codex{}
	cfg := &config.Config{Inject: true, InjectWhatIs: false}
	docs := []*parser.Document{
		{Type: parser.TypeRules, Name: "rule-a", Description: "Rule A 描述", Body: "body of a"},
		{Type: parser.TypeRules, Name: "rule-b", Description: "Rule B 描述", Body: "body of b"},
	}
	plan, _ := tr.Plan(docs, cfg)
	// 现行行为：nonRoot rules 全文 inline 到 AGENTS.md（amp / warp 同风格），
	// 不再产 .codex/memories/（与 Codex 官方语义冲突，已废弃）
	paths := pathSet(plan)
	if !paths["AGENTS.md"] {
		t.Error("missing AGENTS.md")
	}
	main, _ := contentOf(plan, "AGENTS.md")
	for _, want := range []string{
		"rule-a",
		"Rule A 描述",
		"body of a",
		"rule-b",
		"body of b",
	} {
		if !strings.Contains(main, want) {
			t.Errorf("missing %q in AGENTS.md:\n%s", want, main)
		}
	}
	// 无 manifest 段（RulesDir 为空时全 inline，无文件引用清单）
	if strings.Contains(main, "Reference Rules") {
		t.Error("AGENTS.md should not contain Reference Rules manifest after inline migration")
	}
}

// TestCodexNoMemoriesOutput：防回归。codex plan 在 .codex/ 下只允许官方
// 文档化的项目级配置面 .codex/agents/*.toml（developers.openai.com/codex/subagents）；
// .codex/memories/ 等其他路径仍禁止（memories 是 ~/.codex/memories/ 用户级自动系统）。
func TestCodexNoMemoriesOutput(t *testing.T) {
	tr := &Codex{}
	cfg := &config.Config{Inject: true, InjectWhatIs: false}
	docs := []*parser.Document{
		{Type: parser.TypeRules, Name: "r", Body: "rule body"},
		{Type: parser.TypeSkills, Name: "s", Description: "skill", Body: "skill body"},
		{Type: parser.TypeCommands, Name: "c", Description: "cmd", Body: "cmd body"},
		{Type: parser.TypeReferences, Name: "ref", Description: "ref", Body: "ref body"},
		{Type: parser.TypeSubagents, Name: "sub", Description: "sub", Body: "sub body"},
	}
	plan, _ := tr.Plan(docs, cfg)
	for p := range pathSet(plan) {
		if strings.HasPrefix(p, ".codex/") && !strings.HasPrefix(p, ".codex/agents/") {
			t.Errorf("codex plan must not write into .codex/ outside agents/ (official Team Config namespace), got %s", p)
		}
	}
}

// TestCodexReferencesAndSubagentsFallback：references / subagents 降级落 .agents/<subdir>/
func TestCodexReferencesAndSubagentsFallback(t *testing.T) {
	tr := &Codex{}
	cfg := &config.Config{Inject: false}
	docs := []*parser.Document{
		{Type: parser.TypeReferences, Name: "design", Description: "设计参考", Body: "ref body"},
		{Type: parser.TypeSubagents, Name: "reviewer", Description: "审查代理", Body: "sub body"},
	}
	plan, _ := tr.Plan(docs, cfg)
	paths := pathSet(plan)
	for _, want := range []string{".agents/references/design.md", ".agents/subagents/reviewer.md"} {
		if !paths[want] {
			t.Errorf("missing %s in plan: %v", want, paths)
		}
	}
}

func TestCodexCommandsAsSkill(t *testing.T) {
	tr := &Codex{}
	cfg := &config.Config{Inject: false}
	docs := []*parser.Document{
		{Type: parser.TypeCommands, Name: "review", Description: "Run code review", Body: "Steps..."},
	}
	plan, _ := tr.Plan(docs, cfg)
	c, ok := contentOf(plan, ".agents/skills/commands/review/SKILL.md")
	if !ok {
		t.Fatalf("expected .agents/skills/commands/review/SKILL.md, paths: %v", pathSet(plan))
	}
	for _, want := range []string{"name: review", "std-agent-type: commands", "/review", "Run code review"} {
		if !strings.Contains(c, want) {
			t.Errorf("missing %q in:\n%s", want, c)
		}
	}
	assertSkillYAMLFrontmatterFirst(t, c)
	root, ok := contentOf(plan, "AGENTS.md")
	if ok && strings.Contains(root, "Run code review") {
		t.Errorf("commands must stay out of shared AGENTS.md:\n%s", root)
	}
}

// TestCodexSkillYAMLFrontmatterFirst：Codex 要求 SKILL.md 以 --- 开头的 YAML frontmatter；
// marker 必须在 frontmatter 闭合之后，否则 formatter 会把 --- 当 HR 破坏格式。
func TestCodexSkillYAMLFrontmatterFirst(t *testing.T) {
	tr := &Codex{}
	cfg := &config.Config{Inject: true, InjectWhatIs: false}
	docs := []*parser.Document{
		{Type: parser.TypeSkills, Name: "lint", Description: "Lint the repo", Body: "Steps"},
		{Type: parser.TypeCommands, Name: "ship", Description: "Ship it", Body: "Ship steps"},
	}
	plan, _ := tr.Plan(docs, cfg)
	for _, path := range []string{
		".agents/skills/lint/SKILL.md",
		".agents/skills/commands/ship/SKILL.md",
	} {
		c, ok := contentOf(plan, path)
		if !ok {
			t.Fatalf("missing %s, paths: %v", path, pathSet(plan))
		}
		assertSkillYAMLFrontmatterFirst(t, c)
	}
}

func assertSkillYAMLFrontmatterFirst(t *testing.T, content string) {
	t.Helper()
	if !strings.HasPrefix(content, "---\n") {
		t.Fatalf("SKILL.md must start with YAML frontmatter ---, got:\n%s", content)
	}
	if !strings.Contains(content[4:], "\n---\n") {
		t.Fatalf("SKILL.md missing closing --- for frontmatter, got:\n%s", content)
	}
}

func TestCodexCommandsSkillDoesNotCollideWithSkills(t *testing.T) {
	tr := &Codex{}
	cfg := &config.Config{Inject: false}
	// 同 name 的 skill 与 command 同时存在，子目录隔离避免冲突（v3）
	docs := []*parser.Document{
		{Type: parser.TypeSkills, Name: "review", Description: "Skill version", Body: "skill body"},
		{Type: parser.TypeCommands, Name: "review", Description: "Cmd version", Body: "cmd body"},
	}
	plan, _ := tr.Plan(docs, cfg)
	paths := pathSet(plan)
	if !paths[".agents/skills/review/SKILL.md"] {
		t.Errorf("missing skill path, paths: %v", paths)
	}
	if !paths[".agents/skills/commands/review/SKILL.md"] {
		t.Errorf("missing command-as-skill path, paths: %v", paths)
	}
}

func TestCodexAGENTSMdInlineWhenNoRoot(t *testing.T) {
	tr := &Codex{}
	cfg := &config.Config{Inject: false}
	docs := []*parser.Document{
		{Type: parser.TypeRules, Name: "naming", Description: "命名规范", Body: "naming body"},
	}
	plan, _ := tr.Plan(docs, cfg)
	main, _ := contentOf(plan, "AGENTS.md")
	// 无 root rule 时仍有占位标题，nonRoot rule 全文 inline 其后
	for _, want := range []string{"Project AGENTS Manifest", "naming body"} {
		if !strings.Contains(main, want) {
			t.Errorf("missing %q in AGENTS.md:\n%s", want, main)
		}
	}
}

// TestCodexSubagentTOML 验证 subagents 追加 Codex 原生可消费的 .codex/agents/<n>.toml
func TestCodexSubagentTOML(t *testing.T) {
	tr := &Codex{}
	cfg := &config.Config{Inject: true}
	docs := []*parser.Document{
		{Type: parser.TypeSubagents, Name: "reviewer", Description: "Review specialist", Model: "gpt-5.3-codex", Body: "Review carefully.", Path: "subagents/reviewer.md"},
	}
	plan, err := tr.Plan(docs, cfg)
	if err != nil {
		t.Fatal(err)
	}
	c, ok := contentOf(plan, ".codex/agents/reviewer.toml")
	if !ok {
		t.Fatalf("missing .codex/agents/reviewer.toml, paths: %v", pathSet(plan))
	}
	for _, want := range []string{
		`name = "reviewer"`,
		`description = "Review specialist"`,
		`model = "gpt-5.3-codex"`,
		"developer_instructions = '''",
		"Review carefully.",
	} {
		if !strings.Contains(c, want) {
			t.Errorf("missing %q in TOML:\n%s", want, c)
		}
	}
	// 过渡期保留 Markdown 降级产物
	if !pathSet(plan)[".agents/subagents/reviewer.md"] {
		t.Errorf("transition markdown output should be kept, paths: %v", pathSet(plan))
	}
}

// TestCodexImplicitInvocationSidecar 验证 disable_model_invocation 写到
// agents/openai.yaml，而不是污染共享 SKILL.md。
// 来源：https://learn.chatgpt.com/codex/build-skills
func TestCodexImplicitInvocationSidecar(t *testing.T) {
	docs := []*parser.Document{{
		Type:                   parser.TypeSkills,
		Name:                   "review",
		Description:            "Review code",
		DisableModelInvocation: true,
		Body:                   "body",
	}}
	cfg := &config.Config{Inject: false}
	codexPlan, err := (&Codex{}).Plan(docs, cfg)
	if err != nil {
		t.Fatal(err)
	}
	ampPlan, err := (&Amp{}).Plan(docs, cfg)
	if err != nil {
		t.Fatal(err)
	}

	yamlPath := ".agents/skills/review/agents/openai.yaml"
	c, ok := contentOf(codexPlan, yamlPath)
	if !ok {
		t.Fatalf("codex must emit %s when disable_model_invocation, paths: %v", yamlPath, pathSet(codexPlan))
	}
	for _, want := range []string{
		"policy:",
		"allow_implicit_invocation: false",
	} {
		if !strings.Contains(c, want) {
			t.Errorf("missing %q in openai.yaml:\n%s", want, c)
		}
	}

	skill, ok := contentOf(codexPlan, ".agents/skills/review/SKILL.md")
	if !ok {
		t.Fatal("missing SKILL.md")
	}
	if strings.Contains(skill, "allow_implicit_invocation") {
		t.Errorf("SKILL.md must not carry allow_implicit_invocation:\n%s", skill)
	}

	ampSkill, ok := contentOf(ampPlan, ".agents/skills/review/SKILL.md")
	if !ok {
		t.Fatal("amp missing shared SKILL.md")
	}
	if skill != ampSkill {
		t.Errorf("shared SKILL.md must stay byte-identical between codex and amp")
	}
	if pathSet(ampPlan)[yamlPath] {
		t.Errorf("amp must not emit Codex-only sidecar %s", yamlPath)
	}
}

func TestCodexImplicitInvocationSidecarAbsentByDefault(t *testing.T) {
	plan, err := (&Codex{}).Plan([]*parser.Document{{
		Type:        parser.TypeSkills,
		Name:        "review",
		Description: "Review code",
		Body:        "body",
	}}, &config.Config{Inject: false})
	if err != nil {
		t.Fatal(err)
	}
	if pathSet(plan)[".agents/skills/review/agents/openai.yaml"] {
		t.Error("openai.yaml must not be written when disable_model_invocation is unset")
	}
}

func TestCodexCommandSkillImplicitInvocationSidecar(t *testing.T) {
	plan, err := (&Codex{}).Plan([]*parser.Document{{
		Type:                   parser.TypeCommands,
		Name:                   "ship",
		Description:            "Ship it",
		DisableModelInvocation: true,
		Body:                   "ship",
	}}, &config.Config{Inject: false})
	if err != nil {
		t.Fatal(err)
	}
	if !pathSet(plan)[".agents/skills/commands/ship/agents/openai.yaml"] {
		t.Errorf("command-as-skill must also emit openai.yaml, paths: %v", pathSet(plan))
	}
}
