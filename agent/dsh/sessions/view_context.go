package sessions

import (
	"encoding/json"
	"strings"
)

// TranscriptContext is the display row for a non-user `user/message`: context
// injected into model history beside the human prompt. The row keeps the
// injection's own seq, so it sits where the model received it.
const TranscriptContext = "context"

// contextSource is the handful of durable source fields a row header needs.
// Everything else in the source stays in the log; a form the client wants to
// present structurally reads it from the source's own record, not from here.
type contextSource struct {
	Kind    string `json:"kind"`
	Form    string `json:"form"`
	Name    string `json:"name"`
	Summary string `json:"summary"`
	Changes []struct {
		Path string `json:"path"`
	} `json:"changes"`
	References []struct {
		Label string `json:"label"`
	} `json:"references"`
}

// messageInjection reports whether event is context injected into model
// history, and what a reader should call its producer.
//
// A human prompt (`source.kind == "user"`) is the ordinary user row, and a
// compaction checkpoint is the compaction fold's row; both return false. Every
// other source is context: a skill's instructions, the skill catalog, workspace
// instructions, a runtime snapshot, a job notice, a session recall, a goal
// continuation, or a source this build has never seen. Unknown kinds still
// produce a row labelled by their durable kind, so a newer DSH producer is
// visible instead of silently dropped.
func messageInjection(event Event) (*Injection, bool) {
	if event.Type != "user/message" {
		return nil, false
	}
	var data struct {
		Source json.RawMessage `json:"source"`
	}
	if json.Unmarshal(event.Data, &data) != nil {
		return nil, false
	}
	var source contextSource
	if len(data.Source) > 0 {
		_ = json.Unmarshal(data.Source, &source)
	}
	kind := strings.TrimSpace(source.Kind)
	if kind == TranscriptUser || kind == compactionCheckpointKind {
		return nil, false
	}
	return &Injection{
		Source:  kind,
		Form:    strings.TrimSpace(source.Form),
		Label:   injectionLabel(kind, source),
		Summary: strings.TrimSpace(source.Summary),
	}, true
}

// injectionLabel is what a reader recognises the producer by. A skill is named
// by its skill name, instructions by the files they reconciled, and a recall by
// the sessions it cited; every other producer is named by its durable kind.
func injectionLabel(kind string, source contextSource) string {
	switch kind {
	case "skill-invocation":
		return strings.TrimSpace(source.Name)
	case "agent-instructions":
		paths := make([]string, 0, len(source.Changes))
		for _, change := range source.Changes {
			paths = append(paths, change.Path)
		}
		return joinLabels(paths)
	case "session-reference":
		labels := make([]string, 0, len(source.References))
		for _, reference := range source.References {
			labels = append(labels, reference.Label)
		}
		return joinLabels(labels)
	}
	return kind
}

// joinLabels joins non-empty values in first-seen order, without repeating one
// the source listed twice.
func joinLabels(values []string) string {
	kept := make([]string, 0, len(values))
	seen := make(map[string]bool, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		kept = append(kept, value)
	}
	return strings.Join(kept, ", ")
}
