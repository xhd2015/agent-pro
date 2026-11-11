// Package sessionactivity reports the live phase of an agent session.
//
// Eluc (and any other caller) should use Read rather than re-implementing
// per-runner log scans. Kinds are idle, thinking, responding, tool_calling,
// waiting, error, aborted, and not_found.
//
// DSH phases are composed from three independent observations, because no
// single one covers every session: the agent registry says whether a turn is
// running, but only for the sessions a daemon has loaded; the folded session
// log says whether the newest turn ever closed, for any session with a log; and
// the log's write lease tells an open turn that is still executing from one
// whose process is gone. A session the registry does not list is therefore not
// idle: it is an interrupted error when its log left a turn open and no writer
// holds the lease, a running turn when a writer does, and an Unanswered
// placeholder when the log has not been folded yet. not_found means the log is
// gone, never merely "not loaded by the daemon".
package sessionactivity

import (
	"fmt"
	"strings"

	dshsessions "github.com/xhd2015/agent-pro/agent/dsh/sessions"
)

const (
	RunnerDSH   = "dsh"
	RunnerGrok  = "grok"
	RunnerCodex = "codex"
)

// Kind is the live phase of one session.
type Kind string

const (
	Idle        Kind = "idle"
	Thinking    Kind = "thinking"
	Responding  Kind = "responding"
	ToolCalling Kind = "tool_calling"
	Waiting     Kind = "waiting"
	ErrorKind   Kind = "error"
	Aborted     Kind = "aborted"
	NotFound    Kind = "not_found"
)

// Snapshot is one session's current phase plus durable facts the caller
// already needed for display.
type Snapshot struct {
	Runner    string `json:"runner"`
	SessionID string `json:"session_id"`
	Kind      Kind   `json:"kind"`
	Detail    string `json:"detail,omitempty"`
	Live      bool   `json:"live"`
	Workspace string `json:"workspace,omitempty"`
	Title     string `json:"title,omitempty"`
	UpdatedAt int64  `json:"updated_at,omitempty"`
	Outcome   string `json:"outcome,omitempty"`
	// LiveStale marks a phase read from a log whose open turn is newer than the
	// liveness answer: Live is false because the answer could not testify about
	// that turn, not because the turn failed. A caller must not render a stale
	// read as an error.
	LiveStale bool `json:"live_stale,omitempty"`
	// Unanswered marks a placeholder snapshot taken while the evidence behind
	// this phase was missing: the status source did not answer (RPC timeout,
	// daemon still starting), the folded log has not caught up, or the write
	// lease could not be observed. Kind stays ErrorKind so every existing
	// consumer keeps its fallback, but a caller that can show "not known yet"
	// must not present this as the session's real state.
	Unanswered bool `json:"unanswered,omitempty"`
}

// Probe is the injectable entry point. Nil fields use the live agent homes.
type Probe struct {
	DSH   func(sessionID string) (Snapshot, error)
	Grok  func(sessionID string) (Snapshot, error)
	Codex func(sessionID string) (Snapshot, error)

	// RunStatus decides liveness for the default DSH reader when set; nil uses
	// the installed default from SetDefaultRunStatus. It does not apply when DSH
	// is set, because that reader answers liveness itself, and it never applies
	// to Grok or Codex.
	RunStatus RunStatus

	// Outcome and Writers complete that reader's verdict: whether the newest
	// turn in the folded log closed, and whether a process still holds the log's
	// write lease. Nil uses the installed defaults from SetDefaultOutcome and
	// SetDefaultWriter, and a missing hook makes the reader abstain rather than
	// call a turn dead. Like RunStatus they do not apply when DSH is set, and
	// they never apply to Grok or Codex.
	Outcome DurableOutcomeSource
	Writers WriterSource
}

// Default is the production probe: every nil field reads the live agent home,
// and the DSH reader therefore also consults the installed RunStatus, Outcome
// and Writer hooks.
func Default() Probe {
	return Probe{}
}

// Read reports one session's activity using the production probe.
func Read(runner, sessionID string) (Snapshot, error) {
	return Default().Read(runner, sessionID)
}

// Read reports one session's activity.
func (p Probe) Read(runner, sessionID string) (Snapshot, error) {
	runner = strings.ToLower(strings.TrimSpace(runner))
	sessionID = strings.TrimSpace(sessionID)
	if runner == "" {
		return Snapshot{}, fmt.Errorf("runner is required")
	}
	if sessionID == "" {
		return Snapshot{}, fmt.Errorf("session id is required")
	}
	var fn func(string) (Snapshot, error)
	switch runner {
	case RunnerDSH:
		if p.DSH != nil {
			fn = p.DSH
		} else {
			src := dshSourcesOf(p)
			fn = func(id string) (Snapshot, error) { return readDSHStatus(id, src) }
		}
	case RunnerGrok:
		fn = p.Grok
		if fn == nil {
			fn = readGrok
		}
	case RunnerCodex:
		fn = p.Codex
		if fn == nil {
			fn = readCodex
		}
	default:
		return Snapshot{}, fmt.Errorf("unknown runner: %s", runner)
	}
	snap, err := fn(sessionID)
	if err != nil {
		return Snapshot{}, err
	}
	snap.Runner = runner
	snap.SessionID = sessionID
	return snap, nil
}

func snapshotFromDSH(inspection *dshsessions.Inspection, live dshsessions.Live) Snapshot {
	var act dshsessions.Activity
	if live.Known {
		act = dshsessions.Classify(inspection.Events, live)
	} else {
		act = dshsessions.ClassifyLog(inspection.Events)
		act.Live = live.Running
	}
	return Snapshot{
		Runner:    RunnerDSH,
		SessionID: inspection.ID,
		Kind:      Kind(act.Kind),
		Detail:    act.Detail,
		Live:      act.Live,
		LiveStale: act.LiveStale,
		Workspace: inspection.Cwd,
		Title:     inspection.Title,
		UpdatedAt: inspection.UpdatedAt,
		Outcome:   inspection.Outcome,
	}
}
