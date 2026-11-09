package consumers

// OutcomeID is the consumer that tracks how the session's last turn ended and
// whether one is still running.
const OutcomeID = "outcome"

// Turn outcome kinds reported by turn/end.
const (
	OutcomeCompleted   = "completed"
	OutcomeMaxTokens   = "max-tokens"
	OutcomeError       = "error"
	OutcomeAborted     = "aborted"
	OutcomeInterrupted = "interrupted"
)

// OutcomeState is the session's run state plus the last reported failure.
type OutcomeState struct {
	// Running is true between turn/start and turn/end.
	Running bool `json:"running,omitempty"`
	// Turn is the running turn, or the last finished one.
	Turn int64 `json:"turn,omitempty"`
	// Last is the last finished turn's outcome kind.
	Last string `json:"last,omitempty"`
	// Turns counts finished turns.
	Turns int64 `json:"turns,omitempty"`
	// Errors counts turns that ended with an error.
	Errors int64 `json:"errors,omitempty"`
	// Retries counts llm retries, which is the shape of a flaky turn.
	Retries int64 `json:"retries,omitempty"`
	// Error is the last failure reported by a turn or an llm retry.
	Error *FailureState `json:"error,omitempty"`
}

// FailureState is a reported failure.
type FailureState struct {
	Message string `json:"message,omitempty"`
	Code    string `json:"code,omitempty"`
	// Seq is the row that reported it.
	Seq int64 `json:"seq,omitempty"`
}

// Outcome folds turn boundaries into a run state and the last failure.
//
// A turn that never ended (the log simply stops, e.g. the process was killed)
// stays Running until a later turn/end closes it, so callers must only trust
// !Running when the snapshot is CaughtUp.
func Outcome() Consumer {
	return &Typed[OutcomeState]{
		Name:  OutcomeID,
		Ver:   1,
		Watch: []string{"turn/start", "turn/end", "llm/retry"},
		Apply: func(state *OutcomeState, ev Event, _ Point) error {
			fields := decodeObject(ev.Data)
			if fields == nil {
				return nil
			}
			switch ev.Type {
			case "turn/start":
				state.Running = true
				state.Turn = integer(fields, "turn")
			case "turn/end":
				state.Running = false
				state.Turn = integer(fields, "turn")
				state.Turns++
				reason := objectField(fields, "reason")
				kind := text(reason, "kind")
				if kind == "" {
					kind = OutcomeCompleted
				}
				state.Last = kind
				if kind != OutcomeError {
					return nil
				}
				state.Errors++
				if failure := objectField(reason, "error"); failure != nil {
					state.Error = &FailureState{
						Message: text(failure, "message"),
						Code:    text(failure, "code"),
						Seq:     ev.Seq,
					}
				} else if state.Error == nil {
					state.Error = &FailureState{Seq: ev.Seq}
				}
			case "llm/retry":
				// A retry that later succeeds still leaves the session usable,
				// so the count is the signal and the failure is just the last
				// one on record.
				state.Retries++
				failure := objectField(fields, "failure")
				if failure == nil {
					return nil
				}
				state.Error = &FailureState{
					Message: text(failure, "message"),
					Code:    text(failure, "code"),
					Seq:     ev.Seq,
				}
			}
			return nil
		},
	}
}

// Outcome returns the folded run state.
func (s Snapshot) Outcome() OutcomeState {
	var state OutcomeState
	_ = s.Value(OutcomeID, &state)
	return state
}

// Failure returns the last reported failure, or nil when there is none.
func (s Snapshot) Failure() *FailureState {
	return s.Outcome().Error
}
