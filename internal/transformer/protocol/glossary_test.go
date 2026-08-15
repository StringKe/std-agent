package protocol

import (
	"strings"
	"testing"

	"github.com/StringKe/std-agent/internal/config"
)

func TestRenderGlossaryFor_Disabled(t *testing.T) {
	got := RenderGlossaryFor(Adapter{InjectTypeGlossary: false}, &config.Config{InjectTypeGlossary: true})
	if got != "" {
		t.Errorf("disabled should return empty, got %q", got)
	}
}

func TestRenderGlossaryFor_Enabled(t *testing.T) {
	got := RenderGlossaryFor(Adapter{InjectTypeGlossary: true}, &config.Config{InjectTypeGlossary: true})
	if got == "" {
		t.Fatal("enabled should return non-empty")
	}
	if !strings.Contains(got, "std-agent 类型速查") {
		t.Errorf("expected glossary title, got:\n%s", got)
	}
	for _, marker := range []string{"rules", "skills", "commands", "references", "subagents"} {
		if !strings.Contains(got, marker) {
			t.Errorf("expected type %q in glossary, got:\n%s", marker, got)
		}
	}
}

func TestRenderGlossaryFor_AutoInjectMarker(t *testing.T) {
	got := RenderGlossaryFor(Adapter{InjectTypeGlossary: true}, &config.Config{InjectTypeGlossary: true})
	if !strings.Contains(got, "std-agent type glossary auto-injected") {
		t.Errorf("expected auto-inject marker comment, got:\n%s", got)
	}
}

func TestRenderGlossaryFor_ConfigDisabled(t *testing.T) {
	got := RenderGlossaryFor(Adapter{InjectTypeGlossary: true}, &config.Config{InjectTypeGlossary: false})
	if got != "" {
		t.Errorf("config disabled should return empty, got %q", got)
	}
}
