package transformer

import (
	"std-ai/internal/writer"
)

func pathSet(plan *writer.Plan) map[string]bool {
	out := make(map[string]bool, len(plan.Files))
	for _, f := range plan.Files {
		out[f.Path] = true
	}
	return out
}

func contentOf(plan *writer.Plan, path string) (string, bool) {
	for _, f := range plan.Files {
		if f.Path == path {
			return string(f.Content), true
		}
	}
	return "", false
}
