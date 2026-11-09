package consumers

import "encoding/json"

// QuestionID is the consumer that tracks asks and approvals waiting on a human.
const QuestionID = "question"

// ask_user_question is the tool whose call blocks on a person.
const askUserQuestionTool = "ask_user_question"

// A session cannot have many live prompts; the cap only guards against an
// unbounded state if a log is replayed with results missing.
const (
	maxOpenAsks      = 8
	maxOpenApprovals = 8
)

// QuestionState is what the session is currently waiting on.
type QuestionState struct {
	// Asks are ask_user_question calls without a result yet.
	Asks []AskState `json:"asks,omitempty"`
	// Approvals are approval requests without a decision yet.
	Approvals []ApprovalState `json:"approvals,omitempty"`
	// Asked and Answered count ask_user_question calls over the session's life.
	Asked    int64 `json:"asked,omitempty"`
	Answered int64 `json:"answered,omitempty"`
	// ApprovalsAsked and ApprovalsDecided count approval requests.
	ApprovalsAsked   int64 `json:"approvalsAsked,omitempty"`
	ApprovalsDecided int64 `json:"approvalsDecided,omitempty"`
	// Turn is the turn currently running, used when a row omits its turn.
	Turn int64 `json:"turn,omitempty"`
}

// AskState is one pending ask_user_question call.
type AskState struct {
	CallID string        `json:"callId"`
	Seq    int64         `json:"seq"`
	Turn   int64         `json:"turn,omitempty"`
	Step   int64         `json:"step,omitempty"`
	Items  []AskQuestion `json:"questions,omitempty"`
}

// AskQuestion is one question inside a pending ask.
type AskQuestion struct {
	ID       string   `json:"id,omitempty"`
	Header   string   `json:"header,omitempty"`
	Question string   `json:"question,omitempty"`
	Options  []string `json:"options,omitempty"`
	Multi    bool     `json:"multiSelect,omitempty"`
}

// ApprovalState is one pending approval request.
type ApprovalState struct {
	ID     string `json:"id"`
	CallID string `json:"callId,omitempty"`
	Tool   string `json:"tool,omitempty"`
	Reason string `json:"reason,omitempty"`
	Seq    int64  `json:"seq"`
	Turn   int64  `json:"turn,omitempty"`
}

// Question folds the open-tool set into pending prompts.
//
// A tool call is closed by the tool/result that carries its toolCallId, and an
// approval is closed by approval/decided. Anything still open when its turn
// ends was abandoned (the turn was aborted or the process died), so turn/end
// clears it; otherwise a long session would accumulate dead prompts forever.
func Question() Consumer {
	return &Typed[QuestionState]{
		Name:  QuestionID,
		Ver:   1,
		Watch: []string{"tool/call", "tool/result", "approval/asked", "approval/decided", "turn/start", "turn/end"},
		Apply: func(state *QuestionState, ev Event, _ Point) error {
			fields := decodeObject(ev.Data)
			if fields == nil {
				return nil
			}
			switch ev.Type {
			case "turn/start":
				state.Turn = integer(fields, "turn")
			case "tool/call":
				if text(fields, "name") != askUserQuestionTool {
					return nil
				}
				callID := text(fields, "callId")
				if callID == "" || hasAsk(state.Asks, callID) {
					return nil
				}
				state.Asked++
				ask := AskState{
					CallID: callID,
					Seq:    ev.Seq,
					Turn:   turnOf(fields, state),
					Step:   integer(fields, "step"),
					Items:  askItems(argumentObject(fields, "arguments")),
				}
				state.Asks = append(state.Asks, ask)
				if len(state.Asks) > maxOpenAsks {
					state.Asks = state.Asks[len(state.Asks)-maxOpenAsks:]
				}
			case "tool/result":
				callID := text(objectField(fields, "message"), "toolCallId")
				if callID == "" {
					callID = text(fields, "toolCallId")
				}
				if callID == "" {
					return nil
				}
				if removed := removeAsk(&state.Asks, callID); removed {
					state.Answered++
				}
			case "approval/asked":
				id := text(fields, "id")
				if id == "" || hasApproval(state.Approvals, id) {
					return nil
				}
				state.ApprovalsAsked++
				state.Approvals = append(state.Approvals, ApprovalState{
					ID:     id,
					CallID: text(fields, "callId"),
					Tool:   text(fields, "toolName"),
					Reason: text(fields, "reason"),
					Seq:    ev.Seq,
					Turn:   turnOf(fields, state),
				})
				if len(state.Approvals) > maxOpenApprovals {
					state.Approvals = state.Approvals[len(state.Approvals)-maxOpenApprovals:]
				}
			case "approval/decided":
				id := text(fields, "id")
				if id == "" {
					return nil
				}
				state.ApprovalsDecided++
				removeApproval(&state.Approvals, id)
			case "turn/end":
				turn := integer(fields, "turn")
				state.Turn = turn
				state.Asks = pruneAsks(state.Asks, turn)
				state.Approvals = pruneApprovals(state.Approvals, turn)
			}
			return nil
		},
	}
}

// Question returns the folded pending prompts.
func (s Snapshot) Question() QuestionState {
	var state QuestionState
	_ = s.Value(QuestionID, &state)
	return state
}

// Pending reports whether a person's input is needed. It is only meaningful
// when the snapshot is CaughtUp.
func (s Snapshot) Pending() bool {
	state := s.Question()
	return len(state.Asks) > 0 || len(state.Approvals) > 0
}

// turnOf reads a row's turn, falling back to the turn the session is in when
// the row does not carry one.
func turnOf(fields map[string]json.RawMessage, state *QuestionState) int64 {
	if turn := integer(fields, "turn"); turn != 0 {
		return turn
	}
	return state.Turn
}

func askItems(arguments map[string]json.RawMessage) []AskQuestion {
	questions := objectArray(arguments, "questions")
	if len(questions) == 0 {
		return nil
	}
	items := make([]AskQuestion, 0, len(questions))
	for _, question := range questions {
		item := AskQuestion{
			ID:       text(question, "id"),
			Header:   text(question, "header"),
			Question: text(question, "question"),
			Multi:    boolean(question, "multiSelect"),
		}
		for _, option := range objectArray(question, "options") {
			if label := text(option, "label"); label != "" {
				item.Options = append(item.Options, label)
			}
		}
		items = append(items, item)
	}
	return items
}

func hasAsk(asks []AskState, callID string) bool {
	for _, ask := range asks {
		if ask.CallID == callID {
			return true
		}
	}
	return false
}

func removeAsk(asks *[]AskState, callID string) bool {
	for i, ask := range *asks {
		if ask.CallID == callID {
			*asks = append((*asks)[:i], (*asks)[i+1:]...)
			return true
		}
	}
	return false
}

func hasApproval(approvals []ApprovalState, id string) bool {
	for _, approval := range approvals {
		if approval.ID == id {
			return true
		}
	}
	return false
}

func removeApproval(approvals *[]ApprovalState, id string) {
	for i, approval := range *approvals {
		if approval.ID == id {
			*approvals = append((*approvals)[:i], (*approvals)[i+1:]...)
			return
		}
	}
}

func pruneAsks(asks []AskState, turn int64) []AskState {
	kept := asks[:0]
	for _, ask := range asks {
		if ask.Turn > turn {
			kept = append(kept, ask)
		}
	}
	return kept
}

func pruneApprovals(approvals []ApprovalState, turn int64) []ApprovalState {
	kept := approvals[:0]
	for _, approval := range approvals {
		if approval.Turn > turn {
			kept = append(kept, approval)
		}
	}
	return kept
}
