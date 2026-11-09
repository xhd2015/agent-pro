package consumers

import "strings"

// TitleID is the consumer that tracks the session title.
const TitleID = "title"

// TitleState is the latest title the session announced.
type TitleState struct {
	Title string `json:"title,omitempty"`
	// Source is "user", "provider", or "fallback" as reported by the session.
	Source string `json:"source,omitempty"`
	// Seq is the row the title came from.
	Seq int64 `json:"seq,omitempty"`
}

// Title folds session/title rows. The last titled row wins, which is what the
// session itself shows: user renames are just later rows.
func Title() Consumer {
	return &Typed[TitleState]{
		Name:  TitleID,
		Ver:   1,
		Watch: []string{"session/title"},
		Apply: func(state *TitleState, ev Event, _ Point) error {
			fields := decodeObject(ev.Data)
			if fields == nil {
				return nil
			}
			title := strings.TrimSpace(text(fields, "title"))
			if title == "" {
				return nil
			}
			state.Title = title
			state.Seq = ev.Seq
			if source := objectField(fields, "source"); source != nil {
				state.Source = text(source, "kind")
			}
			return nil
		},
	}
}

// Title returns the folded title, or "" when the session never had one.
func (s Snapshot) Title() string {
	var state TitleState
	if !s.Value(TitleID, &state) {
		return ""
	}
	return state.Title
}

// TitleState returns the folded title state.
func (s Snapshot) TitleState() TitleState {
	var state TitleState
	_ = s.Value(TitleID, &state)
	return state
}
