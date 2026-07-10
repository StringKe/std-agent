package transformer

import (
	"strings"
	"testing"

	"github.com/StringKe/std-agent/internal/config"
	"github.com/StringKe/std-agent/internal/parser"
)

func TestGeminiCommandToml(t *testing.T) {
	tr := &Gemini{}
	cfg := &config.Config{Inject: true}
	docs := []*parser.Document{
		{Type: parser.TypeCommands, Name: "review", Description: "Run review", Body: "Please review the diff."},
	}
	plan, _ := tr.Plan(docs, cfg)
	c, ok := contentOf(plan, ".gemini/commands/review.toml")
	if !ok {
		t.Fatalf("missing .gemini/commands/review.toml")
	}
	if !strings.Contains(c, `description = "Run review"`) {
		t.Errorf("missing description: %s", c)
	}
	if !strings.Contains(c, "prompt = '''") {
		t.Errorf("missing prompt body: %s", c)
	}
}
