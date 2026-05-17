package transformer

import (
	"strings"
	"testing"

	"std-ai/internal/config"
	"std-ai/internal/parser"
)

func TestCodexAGENTSMd(t *testing.T) {
	tr := &Codex{}
	cfg := &config.Config{Inject: true, InjectWhatIs: false}
	docs := []*parser.Document{
		{Type: parser.TypeRules, Name: "rule-a", Description: "Rule A 描述", Body: "body of a"},
		{Type: parser.TypeRules, Name: "rule-b", Description: "Rule B 描述", Body: "body of b"},
	}
	plan, _ := tr.Plan(docs, cfg)
	// 现行行为：AGENTS.md（自描述清单，不内联 ## section）+ .codex/memories/<name>.md per nonRoot
	paths := pathSet(plan)
	if !paths["AGENTS.md"] {
		t.Error("missing AGENTS.md")
	}
	if !paths[".codex/memories/rule-a.md"] || !paths[".codex/memories/rule-b.md"] {
		t.Errorf("expected nonRoot rules spilled to .codex/memories/, got %v", paths)
	}
	main, _ := contentOf(plan, "AGENTS.md")
	for _, want := range []string{
		"Project AGENTS Manifest",
		"Reference Rules",
		".codex/memories/rule-a.md",
		"Rule A 描述",
		".codex/memories/rule-b.md",
	} {
		if !strings.Contains(main, want) {
			t.Errorf("missing %q in AGENTS.md:\n%s", want, main)
		}
	}
	// 不应再 ## section 内联 nonRoot rule body
	if strings.Contains(main, "body of a") {
		t.Error("nonRoot rule body should not be inlined in AGENTS.md")
	}
}

func TestCodexAGENTSMdAllSpilled(t *testing.T) {
	tr := &Codex{}
	cfg := &config.Config{Inject: false}
	docs := []*parser.Document{
		{Type: parser.TypeRules, Name: "aaa-tiny", Body: "tiny"},
		{Type: parser.TypeRules, Name: "zzz-huge", Body: strings.Repeat("x", 1000)},
	}
	plan, _ := tr.Plan(docs, cfg)
	paths := pathSet(plan)
	for _, want := range []string{"AGENTS.md", ".codex/memories/aaa-tiny.md", ".codex/memories/zzz-huge.md"} {
		if !paths[want] {
			t.Errorf("missing %s in plan: %v", want, paths)
		}
	}
	main, _ := contentOf(plan, "AGENTS.md")
	if !strings.Contains(main, "Reference Rules") {
		t.Error("expected Reference Rules manifest in AGENTS.md")
	}
}

func TestCodexCommandsAsSkill(t *testing.T) {
	tr := &Codex{}
	cfg := &config.Config{Inject: false}
	docs := []*parser.Document{
		{Type: parser.TypeCommands, Name: "review", Description: "Run code review", Body: "Steps..."},
	}
	plan, _ := tr.Plan(docs, cfg)
	c, ok := contentOf(plan, ".agents/skills/cmd-review/SKILL.md")
	if !ok {
		t.Fatalf("expected .agents/skills/cmd-review/SKILL.md, paths: %v", pathSet(plan))
	}
	for _, want := range []string{"name: cmd-review", "/review", "Run code review"} {
		if !strings.Contains(c, want) {
			t.Errorf("missing %q in:\n%s", want, c)
		}
	}
}

func TestCodexCommandsSkillDoesNotCollideWithSkills(t *testing.T) {
	tr := &Codex{}
	cfg := &config.Config{Inject: false}
	// 同 name 的 skill 与 command 同时存在，应输出到不同路径
	docs := []*parser.Document{
		{Type: parser.TypeSkills, Name: "review", Description: "Skill version", Body: "skill body"},
		{Type: parser.TypeCommands, Name: "review", Description: "Cmd version", Body: "cmd body"},
	}
	plan, _ := tr.Plan(docs, cfg)
	paths := pathSet(plan)
	if !paths[".agents/skills/review/SKILL.md"] {
		t.Errorf("missing skill path, paths: %v", paths)
	}
	if !paths[".agents/skills/cmd-review/SKILL.md"] {
		t.Errorf("missing command-as-skill path, paths: %v", paths)
	}
}

func TestCodexAGENTSMdManifestWhenNoRoot(t *testing.T) {
	tr := &Codex{}
	cfg := &config.Config{Inject: false}
	docs := []*parser.Document{
		{Type: parser.TypeRules, Name: "naming", Description: "命名规范", Body: "naming body"},
	}
	plan, _ := tr.Plan(docs, cfg)
	main, _ := contentOf(plan, "AGENTS.md")
	if !strings.Contains(main, "Reference Rules") {
		t.Errorf("无 root rule 时应自动追加 Reference Rules:\n%s", main)
	}
}
