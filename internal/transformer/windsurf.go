package transformer

import (
	"std-ai/internal/config"
	"std-ai/internal/parser"
	"std-ai/internal/transformer/protocol"
	"std-ai/internal/writer"
)

func init() {
	Register(&Windsurf{})
}

// Windsurf 是 Codeium Windsurf transformer
//
// 协议族：WindsurfStyle（无根文件，rules 用 trigger frontmatter，原生 skills 包，
// commands 落 .windsurf/workflows/*.md）。
type Windsurf struct{}

// Name 返回 "windsurf"
func (w *Windsurf) Name() string { return "windsurf" }

// Plan 计算输出
func (w *Windsurf) Plan(docs []*parser.Document, cfg *config.Config) (*writer.Plan, error) {
	return protocol.WindsurfStyle{}.Plan(FilterDocs(docs, w.Name()), windsurfAdapter, cfg)
}

// windsurfAdapter 配置 WindsurfStyle 协议族的 windsurf 变体
var windsurfAdapter = protocol.Adapter{
	Name:                 "windsurf",
	RulesDir:             ".windsurf/rules",
	SkillsDir:            ".windsurf/skills",
	CommandsDir:          ".windsurf/workflows",
	CommandsFileSuffix:   ".md",
	FallbackDir:          ".windsurf/rules",
	InjectExplainer:      true,
	InjectStdaiTypeField: true,
	RuleTriggerMode:      protocol.TriggerTrigger,
	InjectTypeGlossary:   true,
	MaxBytesPerFile:      12000, // windsurf 单 rule 字符上限（保持与原实现等价）
	SkillSupportedFields: []string{"name", "description"},
}
