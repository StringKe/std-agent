package transformer

import (
	"strings"

	"github.com/StringKe/std-agent/internal/config"
	"github.com/StringKe/std-agent/internal/parser"
	"github.com/StringKe/std-agent/internal/transformer/protocol"
	"github.com/StringKe/std-agent/internal/writer"
)

func init() {
	Register(&Windsurf{})
}

// Windsurf 是 Windsurf（已并入 Cognition/Devin Desktop）transformer
//
// 协议族：WindsurfStyle（无根文件，rules 用 trigger frontmatter，原生 skills 包，
// commands 落 .windsurf/workflows/*.md）。
//
// Devin 双写过渡：Windsurf 并入 Cognition 后文档以 Devin Desktop 呈现，规则目录
// 首选 `.devin/rules/`（`.windsurf/rules/` 降为 fallback 仍被读取）。rules 与
// skills 同步双写 `.devin/`，两边字节一致。workflows 仍只写 `.windsurf/workflows/`
// （Devin Local 不消费 workflows）。
type Windsurf struct{}

// Name 返回 "windsurf"
func (w *Windsurf) Name() string { return "windsurf" }

// Plan 计算输出，并把 .windsurf/rules/ 与 .windsurf/skills/ 镜像到 .devin/
func (w *Windsurf) Plan(docs []*parser.Document, cfg *config.Config) (*writer.Plan, error) {
	plan, err := protocol.WindsurfStyle{}.Plan(FilterDocs(docs, w.Name()), windsurfAdapter, cfg)
	if err != nil {
		return nil, err
	}
	mirrored := make([]writer.FileOp, 0, len(plan.Files))
	for _, f := range plan.Files {
		if rest, ok := strings.CutPrefix(f.Path, ".windsurf/rules/"); ok {
			dup := f
			dup.Path = ".devin/rules/" + rest
			mirrored = append(mirrored, dup)
		}
		if rest, ok := strings.CutPrefix(f.Path, ".windsurf/skills/"); ok {
			dup := f
			dup.Path = ".devin/skills/" + rest
			mirrored = append(mirrored, dup)
		}
	}
	plan.Files = append(plan.Files, mirrored...)
	return plan, nil
}

// windsurfAdapter 配置 WindsurfStyle 协议族的 windsurf 变体
var windsurfAdapter = protocol.Adapter{
	Name:                 "windsurf",
	RulesDir:             ".windsurf/rules",
	SkillsDir:            ".windsurf/skills",
	CommandsDir:          ".windsurf/workflows",
	CommandsFileSuffix:   ".md",
	SubagentsDir:         ".devin/agents",
	FallbackDir:          ".windsurf/rules",
	InjectExplainer:      true,
	InjectStdaiTypeField: true,
	RuleTriggerMode:      protocol.TriggerTrigger,
	InjectTypeGlossary:   true,
	MaxBytesPerFile:      12000, // windsurf 单 rule 字符上限（保持与原实现等价）
	SkillSupportedFields: []string{"name", "description"},
}
