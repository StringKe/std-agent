package transformer

import (
	"testing"

	"github.com/StringKe/std-agent/internal/config"
	"github.com/StringKe/std-agent/internal/parser"
)

func TestClinePriorityPrefix(t *testing.T) {
	tr := &Cline{}
	cfg := &config.Config{Inject: false}
	docs := []*parser.Document{
		{Type: parser.TypeRules, Name: "a", Priority: parser.PriorityHigh, Body: "h"},
		{Type: parser.TypeRules, Name: "b", Priority: parser.PriorityNormal, Body: "n"},
		{Type: parser.TypeRules, Name: "c", Priority: parser.PriorityLow, Body: "l"},
	}
	plan, _ := tr.Plan(docs, cfg)
	paths := pathSet(plan)
	for _, want := range []string{".clinerules/100-a.md", ".clinerules/500-b.md", ".clinerules/900-c.md"} {
		if !paths[want] {
			t.Errorf("missing %s, paths: %v", want, paths)
		}
	}
}
