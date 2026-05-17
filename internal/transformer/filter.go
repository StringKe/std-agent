package transformer

import (
	"std-ai/internal/parser"
)

// FilterDocs 按 target 名筛选适用的 docs，返回新 slice
func FilterDocs(docs []*parser.Document, target string) []*parser.Document {
	out := make([]*parser.Document, 0, len(docs))
	for _, d := range docs {
		if targetApplies(d, target) {
			out = append(out, d)
		}
	}
	return out
}

func targetApplies(d *parser.Document, target string) bool {
	if len(d.ExcludeTargets) > 0 {
		for _, t := range d.ExcludeTargets {
			if t == target {
				return false
			}
		}
		return true
	}
	if len(d.Targets) == 0 {
		return true
	}
	for _, t := range d.Targets {
		if t == target {
			return true
		}
	}
	return false
}
