package procresolve

// Options configures ResolveFromPID. ListProcs and Lsof are injectable for
// tests; production may wire live process listing and lsof.
type Options struct {
	GrokHome  string
	CodexHome string
	MaxDepth  int
	ListProcs func() []Proc
	Lsof      func(pid int) []string

	// EnrichInfo, when true and Kind is grok with a SessionID, calls
	// LookupGrokInfo to fill GrokTitle / GrokModel. Opt-in; default false.
	EnrichInfo bool
	// LookupGrokInfo is optional. Production CLI may default to
	// agent/grok/sessions.Info; tests inject a pure stub.
	LookupGrokInfo func(home, sessionID string) (title, model string, err error)
}

// Proc is one process row in a snapshot.
type Proc struct {
	PID  int
	PPID int
	Cmd  string
}

// ProcNode is a classified process in the descendant tree of the input pid.
type ProcNode struct {
	PID  int
	PPID int
	Role string // input | agent-run | agent-run-serve | grok | codex | other
	Cmd  string
}

// Result is the outcome of ResolveFromPID.
type Result struct {
	InputPID   int
	Kind       string // grok | codex | none
	SessionID  string
	Source     string // open-files | open-files+tree
	Confidence string // hard | "" when none
	RunnerPID  int
	RunnerCmd  string
	Tree       []ProcNode
	Warnings   []string
	// GrokTitle / GrokModel are filled only when EnrichInfo is true and
	// LookupGrokInfo succeeds for a grok hard hit.
	GrokTitle string
	GrokModel string
}

// TreeFormatOptions controls FormatTree connector style.
type TreeFormatOptions struct {
	// ASCII, when true, uses +-- / `-- / | instead of box-drawing.
	ASCII bool
}
