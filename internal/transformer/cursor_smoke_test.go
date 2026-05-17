package transformer

import (
	"strings"
	"testing"

	"std-ai/internal/config"
	"std-ai/internal/parser"
)

func TestCursorRuleModes(t *testing.T) {
	tr := &Cursor{}
	cfg := &config.Config{Inject: false}
	cases := []struct {
		name string
		doc  *parser.Document
		want string
	}{
		{"always", &parser.Document{Type: parser.TypeRules, Name: "x", AlwaysApply: true, Body: "b"}, "alwaysApply: true"},
		{"glob", &parser.Document{Type: parser.TypeRules, Name: "x", ApplyTo: []string{"**/*.go", "**/*.ts"}, Body: "b"}, "**/*.go,**/*.ts"},
		{"agent-req", &parser.Document{Type: parser.TypeRules, Name: "x", Description: "use when X", Body: "b"}, "description: use when X"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			plan, _ := tr.Plan([]*parser.Document{tc.doc}, cfg)
			if len(plan.Files) != 1 {
				t.Fatalf("expected 1 file, got %d", len(plan.Files))
			}
			c := string(plan.Files[0].Content)
			if !strings.Contains(c, tc.want) {
				t.Errorf("missing %q in:\n%s", tc.want, c)
			}
		})
	}
}
