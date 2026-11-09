package consumers

import "encoding/json"

// ModelID is the consumer that tracks the session's model route.
const ModelID = "model"

// ModelSelection is one provider/model route a session can run on.
type ModelSelection struct {
	Provider string `json:"provider,omitempty"`
	Model    string `json:"model,omitempty"`
	// ReasoningEffort is only known once a request used the route.
	ReasoningEffort string `json:"reasoningEffort,omitempty"`
}

// ModelState is the durable model-selection intent plus the route a request
// actually used. It mirrors DSH's own modelSelection projection: a selection
// the user made is pending until a request header confirms it, and the header
// is what records the effort the request really ran with.
type ModelState struct {
	// LastUsed is the route of the last committed request.
	LastUsed *ModelSelection `json:"lastUsed,omitempty"`
	// Pending is a selection announced but not yet used.
	Pending *ModelSelection `json:"pending,omitempty"`
	// Seq is the row LastUsed or Pending came from.
	Seq int64 `json:"seq,omitempty"`
}

// Model folds model/selection and request/header rows into the live route.
//
// The fold is deliberately identical to DSH's projection so the Messages tab
// and the daemon agree on what the session runs: request/header wins for the
// used route, model/selection sets the pending one, and a header that confirms
// the pending route clears it.
func Model() Consumer {
	return &Typed[ModelState]{
		Name:  ModelID,
		Ver:   1,
		Watch: []string{"model/selection", "request/header"},
		Apply: func(state *ModelState, ev Event, _ Point) error {
			switch ev.Type {
			case "model/selection":
				fields := decodeObject(ev.Data)
				selection := selectionFrom(fields)
				if selection == nil || sameSelection(state.Pending, selection) {
					return nil
				}
				state.Pending = selection
				state.Seq = ev.Seq
				return nil
			case "request/header":
				header := objectField(decodeObject(ev.Data), "header")
				config := objectField(header, "config")
				selection := selectionFrom(config)
				if selection == nil {
					return nil
				}
				if sameSelection(state.Pending, selection) {
					state.Pending = nil
				}
				if sameSelection(state.LastUsed, selection) {
					return nil
				}
				state.LastUsed = selection
				state.Seq = ev.Seq
				return nil
			}
			return nil
		},
	}
}

// selectionFrom reads a provider/model/reasoningEffort triple. A row without a
// model is not a route, so it folds to a no-op.
func selectionFrom(fields map[string]json.RawMessage) *ModelSelection {
	model := text(fields, "model")
	if model == "" {
		return nil
	}
	provider := text(fields, "provider")
	if provider == "" {
		return nil
	}
	return &ModelSelection{
		Provider:        provider,
		Model:           model,
		ReasoningEffort: text(fields, "reasoningEffort"),
	}
}

// sameSelection compares two routes field by field, the way DSH does.
func sameSelection(left, right *ModelSelection) bool {
	if left == nil || right == nil {
		return left == right
	}
	return left.Provider == right.Provider &&
		left.Model == right.Model &&
		left.ReasoningEffort == right.ReasoningEffort
}

// Model returns the folded route.
func (s Snapshot) Model() ModelState {
	var state ModelState
	_ = s.Value(ModelID, &state)
	return state
}

// Next is the route the session is heading for: a pending selection if there is
// one, otherwise the route last used.
func (s ModelState) Next() *ModelSelection {
	if s.Pending != nil {
		return s.Pending
	}
	return s.LastUsed
}
