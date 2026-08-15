package transformer

import (
	"path"

	"github.com/StringKe/std-agent/internal/config"
	"github.com/StringKe/std-agent/internal/parser"
	"github.com/StringKe/std-agent/internal/transformer/protocol"
	"github.com/StringKe/std-agent/internal/writer"
)

const sharedAgentsFileName = "AGENTS.md"

// CanonicalizeSharedAGENTS replaces every target-specific AGENTS.md output with
// one target-neutral rules document. A rule is included when it applies to at
// least one enabled AGENTS.md producer; target-specific capabilities remain in
// their native sidecar files.
func CanonicalizeSharedAGENTS(plans []*writer.Plan, docs []*parser.Document, cfg *config.Config) error {
	var producers []*writer.Plan
	for _, plan := range plans {
		if planContainsAGENTS(plan) {
			producers = append(producers, plan)
		}
	}
	if len(producers) == 0 {
		return nil
	}

	seen := map[*parser.Document]struct{}{}
	var sharedRules []*parser.Document
	for _, plan := range producers {
		for _, doc := range FilterDocs(docs, plan.Target) {
			if doc.Type != parser.TypeRules {
				continue
			}
			if _, ok := seen[doc]; ok {
				continue
			}
			seen[doc] = struct{}{}
			sharedRules = append(sharedRules, doc)
		}
	}

	canonical, err := (protocol.AgentsMD{}).Plan(sharedRules, protocol.Adapter{
		RootFileName:       sharedAgentsFileName,
		NestedSupported:    true,
		InjectTypeGlossary: true,
	}, cfg)
	if err != nil {
		return err
	}

	for _, plan := range producers {
		files := make([]writer.FileOp, 0, len(plan.Files)+len(canonical.Files))
		for _, op := range plan.Files {
			if path.Base(op.Path) != sharedAgentsFileName {
				files = append(files, op)
			}
		}
		files = append(files, canonical.Files...)
		plan.Files = files
	}
	return nil
}

func planContainsAGENTS(plan *writer.Plan) bool {
	for _, op := range plan.Files {
		if path.Base(op.Path) == sharedAgentsFileName {
			return true
		}
	}
	return false
}
