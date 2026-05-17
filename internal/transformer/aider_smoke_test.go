package transformer

import (
	"testing"

	"std-ai/internal/config"
	"std-ai/internal/parser"
)

func TestAiderNoop(t *testing.T) {
	tr := &Aider{}
	cfg := &config.Config{Inject: true}
	docs := []*parser.Document{{Type: parser.TypeRules, Name: "x", Body: "y"}}
	plan, _ := tr.Plan(docs, cfg)
	if len(plan.Files) != 0 {
		t.Errorf("aider should be noop, got %d files", len(plan.Files))
	}
}
