// Package itermsnapshot wraps P1 shell/iterm2/snapshot Capture with
// procresolve agent attach for busy panes only. Composition only — does not
// mutate snapshot.SnapshotSession. No kool import.
package itermsnapshot

import (
	"github.com/xhd2015/agent-pro/pkgs/procresolve"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2/snapshot"
)

// CaptureOpts controls agent-enrich Capture. L2 injects Snapshot + ResolveFromPID.
type CaptureOpts struct {
	// NoEnrich skips procresolve agent attach entirely.
	NoEnrich bool

	// Snapshot, when non-nil, is the base inventory; BaseCapture is not called.
	// L2 tests inject a fixture *snapshot.Snapshot (no AppleScript / live ps).
	Snapshot *snapshot.Snapshot

	// BaseCapture obtains the base inventory when Snapshot is nil.
	// Production default: snapshot.Capture (or equivalent via snapshot.Collector).
	// Hard errors from BaseCapture propagate from Capture.
	BaseCapture func() (*snapshot.Snapshot, []string, error)

	// ResolveFromPID attaches agents for busy panes.
	// Production default: procresolve.ResolveFromPID with live ListProcs/Lsof
	// and EnrichInfo=true (kool parity). Soft on error/nil/none.
	ResolveFromPID func(pid int) (*procresolve.Result, error)
}

// Result is the enriched view: base Snapshot plus agents by session ID.
type Result struct {
	Snapshot *snapshot.Snapshot
	// Agents keyed by SnapshotSession.ID (iTerm session UUID).
	// Nil or empty when no agents attached.
	Agents map[string]*SessionAgent
}

// SessionAgent is the procresolve-derived agent for one busy pane.
type SessionAgent struct {
	Kind      string // grok | codex
	SessionID string
	Title     string // from procresolve.Result.GrokTitle when present
	Tree      []AgentTreeNode
}

// AgentTreeNode is one process in the agent process tree.
type AgentTreeNode struct {
	PID  int
	PPID int
	Role string // input | agent-run | … | grok | codex | other
	Cmd  string
}

// Capture runs base inventory (inject Snapshot or BaseCapture / default), then
// attaches agents for busy panes unless NoEnrich.
// Returns (*Result, warnings, error). Hard error only from base capture.
// Agent resolve failures are soft (no entry in Agents).
func Capture(opts CaptureOpts) (*Result, []string, error) {
	var (
		snap *snapshot.Snapshot
		warn []string
		err  error
	)

	if opts.Snapshot != nil {
		snap = opts.Snapshot
	} else {
		base := opts.BaseCapture
		if base == nil {
			base = snapshot.Capture
		}
		snap, warn, err = base()
		if err != nil {
			return nil, warn, err
		}
	}

	res := &Result{
		Snapshot: snap,
	}
	if opts.NoEnrich || snap == nil {
		return res, warn, nil
	}

	resolve := opts.ResolveFromPID
	if resolve == nil {
		// One process-table snapshot shared by every busy-pane resolve.
		procs := procresolve.ListLiveProcs()
		resolve = func(pid int) (*procresolve.Result, error) {
			return procresolve.ResolveFromPID(pid, procresolve.Options{
				ListProcs:  func() []procresolve.Proc { return procs },
				Lsof:       procresolve.LiveLsof,
				EnrichInfo: true,
			})
		}
	}

	agents := make(map[string]*SessionAgent)
	for i := range snap.Windows {
		w := &snap.Windows[i]
		for j := range w.Tabs {
			tab := &w.Tabs[j]
			for k := range tab.Sessions {
				sess := &tab.Sessions[k]
				ag := attachAgent(sess, resolve)
				if ag != nil {
					agents[sess.ID] = ag
				}
			}
		}
	}
	if len(agents) > 0 {
		res.Agents = agents
	}
	return res, warn, nil
}

// attachAgent implements busy-only soft attach (kool attachAgent parity).
// Does not mutate sess.
func attachAgent(sess *snapshot.SnapshotSession, resolve func(int) (*procresolve.Result, error)) *SessionAgent {
	// Busy-only: Idle non-nil and false. Idle=true or Idle=nil → skip.
	if sess.Idle == nil || *sess.Idle {
		return nil
	}

	// Prefer PID, else ShellPID; require positive.
	var pid int
	if sess.PID != nil {
		pid = *sess.PID
	} else if sess.ShellPID != nil {
		pid = *sess.ShellPID
	}
	if pid <= 0 {
		return nil
	}

	r, err := resolve(pid)
	if err != nil || r == nil {
		return nil
	}
	if r.Kind == "" || r.Kind == "none" || r.SessionID == "" {
		return nil
	}

	tree := make([]AgentTreeNode, len(r.Tree))
	for i, n := range r.Tree {
		tree[i] = AgentTreeNode{
			PID:  n.PID,
			PPID: n.PPID,
			Role: n.Role,
			Cmd:  n.Cmd,
		}
	}
	return &SessionAgent{
		Kind:      r.Kind,
		SessionID: r.SessionID,
		Title:     r.GrokTitle,
		Tree:      tree,
	}
}
