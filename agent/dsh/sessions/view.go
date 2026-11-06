package sessions

// View is a cached display projection of one session artifact. Rows include
// thinking/present/ask/tool summaries that Transcript omits.
type View struct {
	Summary
	ReasoningEffort string
	Todos           []Todo
	Rows            []Message
}

// Message is one display line of a session preview.
type Message struct {
	Seq     int64    `json:"seq"`
	Time    int64    `json:"time"`
	Kind    string   `json:"kind"` // user | assistant | tool | thinking | error | aborted
	Text    string   `json:"text,omitempty"`
	Sum     string   `json:"sum,omitempty"`
	Tool    string   `json:"tool,omitempty"`
	Present *Present `json:"present,omitempty"`
	Ask     *Ask     `json:"ask,omitempty"`
	Images  []Image  `json:"images,omitempty"`
}

type Image struct {
	MimeType string `json:"mimeType"`
	Index    int    `json:"index"`
}

type PresentFile struct {
	Path        string `json:"path"`
	Description string `json:"description,omitempty"`
}

type Present struct {
	Files []PresentFile `json:"files"`
}

type AskOption struct {
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`
}

type AskQuestion struct {
	ID          string      `json:"id"`
	Question    string      `json:"question"`
	Header      string      `json:"header,omitempty"`
	Detail      string      `json:"detail,omitempty"`
	Options     []AskOption `json:"options,omitempty"`
	MultiSelect bool        `json:"multiSelect,omitempty"`
}

type AskAnswer struct {
	ID       string   `json:"id"`
	Selected []string `json:"selected"`
	Custom   string   `json:"custom,omitempty"`
}

type Ask struct {
	CallID    string        `json:"callId,omitempty"`
	Questions []AskQuestion `json:"questions"`
	Answers   []AskAnswer   `json:"answers,omitempty"`
}

type Todo struct {
	Content string `json:"content"`
	Status  string `json:"status"`
}

// View returns the cached display projection for id. An unchanged artifact is
// not re-decoded or re-projected.
func (s *Store) View(id string) (*View, error) {
	body, err := s.loadInspectBody(id)
	if err != nil {
		return nil, err
	}
	inspectBodies.mu.Lock()
	if body.view != nil {
		view := body.view
		inspectBodies.mu.Unlock()
		return view, nil
	}
	inspectBodies.mu.Unlock()
	if err := s.ensureInspectEvents(body); err != nil {
		return nil, err
	}
	view := ProjectInspection(body.inspection)
	inspectBodies.mu.Lock()
	if body.view == nil {
		body.view = view
		dropInspectEventsLocked(body)
	}
	out := body.view
	inspectBodies.mu.Unlock()
	return out, nil
}

// ProjectInspection builds a View from an already-decoded Inspection.
func ProjectInspection(inspection *Inspection) *View {
	if inspection == nil {
		return &View{}
	}
	return &View{
		Summary:         inspection.Summary,
		ReasoningEffort: lastReasoningEffort(inspection.Events),
		Todos:           latestAgentTodos(inspection.Events),
		Rows:            Project(inspection.Events),
	}
}
