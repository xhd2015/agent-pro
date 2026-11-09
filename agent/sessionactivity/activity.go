// Package sessionactivity reports the live phase of an agent session.
//
// Eluc (and any other caller) should use Read rather than re-implementing
// per-runner log scans. Kinds are idle, thinking, responding, tool_calling,
// waiting, error, aborted, and not_found.
//
// DSH logs are read in two depths. A session session/status lists is tailed and
// classified with its liveness, so a failed or aborted turn survives the
// process exiting; a session the answered map does not list is answered from
// its log header, because the daemon has no state for it and its tail would
// only cost time. not_found therefore means the log is gone, never merely "not
// loaded by the daemon".
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
}

// Default is the production probe: every nil field reads the live agent home,
// and the DSH reader therefore also consults the installed RunStatus hook.
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
			rs := p.runStatusHook()
			fn = func(id string) (Snapshot, error) { return readDSHStatus(id, rs) }
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

func snapshotFromDSH(inspection *dshsessions.Inspection, live bool, liveKnown bool) Snapshot {
	var act dshsessions.Activity
	if liveKnown {
		act = dshsessions.Classify(inspection.Events, live)
	} else {
		act = dshsessions.ClassifyLog(inspection.Events)
		act.Live = live
	}
	return Snapshot{
		Runner:    RunnerDSH,
		SessionID: inspection.ID,
		Kind:      Kind(act.Kind),
		Detail:    act.Detail,
		Live:      act.Live,
		Workspace: inspection.Cwd,
		Title:     inspection.Title,
		UpdatedAt: inspection.UpdatedAt,
		Outcome:   inspection.Outcome,
	}
}
