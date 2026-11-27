package sessions

// View is a cached display projection of one session artifact. Rows include
// thinking/present/ask/tool summaries that Transcript omits.
type View struct {
	Summary
	ReasoningEffort string
	Todos           []Todo
	Rows            []Message
	// Truncated means Rows come from a suffix, not the whole log.
	Truncated bool
}

// Message is one display line of a session preview.
type Message struct {
	Seq  int64  `json:"seq"`
	Time int64  `json:"time"`
	Kind string `json:"kind"` // user | assistant | tool | thinking | error | aborted | command | context
	Text string `json:"text,omitempty"`
	Sum  string `json:"sum,omitempty"`
	Tool string `json:"tool,omitempty"`
	// Ref is the identity a row shares with the row that supersedes it, so a
	// live reader can replace the pending form with the settled one instead of
	// showing both. Empty for rows that arrive once and never change.
	Ref string `json:"ref,omitempty"`
	// Injection is set on a context row: the durable source a non-user
	// `user/message` carried (a skill's instructions, the catalog, workspace
	// instructions, a runtime snapshot, a notice, a recall). It is nil for
	// every other kind.
	Injection *Injection `json:"injection,omitempty"`
	Present   *Present   `json:"present,omitempty"`
	Ask       *Ask       `json:"ask,omitempty"`
	Images    []Image    `json:"images,omitempty"`
	// ResumedAt is when a paused row was continued. Zero means the row is still
	// the open pause; a later turn/start stamps it.
	ResumedAt int64 `json:"resumedAt,omitempty"`
}

// Injection names the context a user-role request carried besides the human
// prompt. Label is what a reader recognises the producer by (a skill name, the
// instruction files, recalled session labels); Source and Form are the durable
// `source.kind` and `source.form`, kept for the row's icon and for a body the
// client wants to present structurally. A reader that does not know a source
// still shows its kind, so a newer DSH producer never disappears.
type Injection struct {
	Source  string `json:"source"`
	Form    string `json:"form,omitempty"`
	Label   string `json:"label,omitempty"`
	Summary string `json:"summary,omitempty"`
}

type Image struct {
	MimeType string `json:"mimeType"`
	Index    int    `json:"index"`
	// SHA256 is the content address of an attachment-backed image, so callers
	// can read the bytes from the attachment store instead of re-reading the
	// log. Empty for inline (data:) images.
	SHA256 string `json:"sha256,omitempty"`
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
