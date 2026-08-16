package protocol

import (
	"strings"
	"testing"

	"github.com/StringKe/std-agent/internal/parser"
)

func TestRenderGlobs(t *testing.T) {
	cases := []struct {
		name   string
		field  string
		format GlobsFormat
		vals   []string
		want   string
	}{
		{
			name:   "empty field skip",
			field:  "",
			format: GlobsList,
			vals:   []string{"**/*.go"},
			want:   "",
		},
		{
			name:   "empty vals skip",
			field:  "globs",
			format: GlobsList,
			vals:   nil,
			want:   "",
		},
		{
			name:   "list format",
			field:  "globs",
			format: GlobsList,
			vals:   []string{"**/*.go", "**/*.md"},
			want:   "globs:\n  - \"**/*.go\"\n  - \"**/*.md\"\n",
		},
		{
			name:   "comma string format",
			field:  "applyTo",
			format: GlobsCommaString,
			vals:   []string{"**/*.go", "**/*.md"},
			want:   "applyTo: \"**/*.go,**/*.md\"\n",
		},
		{
			name:   "paths field with list (quoted due to glob char)",
			field:  "paths",
			format: GlobsList,
			vals:   []string{"src/**"},
			want:   "paths:\n  - \"src/**\"\n",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := RenderGlobs(tc.field, tc.format, tc.vals)
			if got != tc.want {
				t.Errorf("RenderGlobs() got:\n%q\nwant:\n%q", got, tc.want)
			}
		})
	}
}

func TestRenderTriggerFrontmatter_TriggerNone(t *testing.T) {
	got := RenderTriggerFrontmatter(TriggerNone, &parser.Document{AlwaysApply: true})
	if got != "" {
		t.Errorf("TriggerNone should return empty, got %q", got)
	}
}

func TestRenderTriggerFrontmatter_AlwaysApply(t *testing.T) {
	cases := []struct {
		name string
		doc  *parser.Document
		want string
	}{
		{"AlwaysApply=true", &parser.Document{AlwaysApply: true}, "alwaysApply: true\n"},
		{"AlwaysApply=false", &parser.Document{AlwaysApply: false}, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := RenderTriggerFrontmatter(TriggerAlwaysApply, tc.doc)
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestRenderTriggerFrontmatter_Trigger(t *testing.T) {
	cases := []struct {
		name string
		doc  *parser.Document
		want string
	}{
		{
			name: "always_on",
			doc:  &parser.Document{AlwaysApply: true},
			want: "trigger: always_on\n",
		},
		{
			name: "glob with apply",
			doc:  &parser.Document{ApplyTo: []string{"**/*.go"}},
			want: "trigger: glob\nglobs:\n  - \"**/*.go\"\n",
		},
		{
			name: "model_decision",
			doc:  &parser.Document{Description: "use when reviewing"},
			want: "trigger: model_decision\ndescription: use when reviewing\n",
		},
		{
			name: "manual fallback",
			doc:  &parser.Document{},
			want: "trigger: manual\n",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := RenderTriggerFrontmatter(TriggerTrigger, tc.doc)
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestRenderTriggerFrontmatter_ApplyTo(t *testing.T) {
	cases := []struct {
		name string
		doc  *parser.Document
		want string
	}{
		{
			name: "single glob",
			doc:  &parser.Document{ApplyTo: []string{"**/*.go"}},
			want: "applyTo: \"**/*.go\"\n",
		},
		{
			name: "multi glob comma",
			doc:  &parser.Document{ApplyTo: []string{"**/*.go", "**/*.md"}},
			want: "applyTo: \"**/*.go,**/*.md\"\n",
		},
		{
			name: "empty apply skip",
			doc:  &parser.Document{},
			want: "",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := RenderTriggerFrontmatter(TriggerApplyTo, tc.doc)
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestRenderTriggerFrontmatter_Inclusion(t *testing.T) {
	cases := []struct {
		name string
		doc  *parser.Document
		want string
	}{
		{
			name: "always",
			doc:  &parser.Document{AlwaysApply: true},
			want: "inclusion: always\n",
		},
		{
			name: "fileMatch single",
			doc:  &parser.Document{ApplyTo: []string{"**/*.go"}},
			want: "inclusion: fileMatch\nfileMatchPattern: \"**/*.go\"\n",
		},
		{
			name: "fileMatch multi",
			doc:  &parser.Document{ApplyTo: []string{"**/*.ts", "**/*.tsx"}},
			want: "inclusion: fileMatch\nfileMatchPattern:\n  - \"**/*.ts\"\n  - \"**/*.tsx\"\n",
		},
		{
			name: "auto",
			doc:  &parser.Document{Name: "api-design", Description: "REST API patterns"},
			want: "inclusion: auto\nname: api-design\ndescription: REST API patterns\n",
		},
		{
			name: "manual",
			doc:  &parser.Document{},
			want: "inclusion: manual\n",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := RenderTriggerFrontmatter(TriggerInclusion, tc.doc)
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestRenderTriggerFrontmatter_NilDoc(t *testing.T) {
	if got := RenderTriggerFrontmatter(TriggerTrigger, nil); got != "" {
		t.Errorf("nil doc should return empty, got %q", got)
	}
}

func TestFmBodyOnlyTrimsWrapper(t *testing.T) {
	// 间接验证 fmBodyOnly：RenderGlobs 单字段输出不含 "---" 包裹
	got := RenderGlobs("globs", GlobsList, []string{"a"})
	if strings.Contains(got, "---") {
		t.Errorf("output should not contain frontmatter wrapper, got %q", got)
	}
}
