package agentrunapi

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
	"github.com/xhd2015/agent-pro/pkgs/agenttty"
	"github.com/xhd2015/tty-watch/pkgs/ttywatch"
)

// LifecycleProbe is the production TTY/registry lifecycle probe for Classify.
// It mirrors agent-run CLI probeSessionStatus for ResumeReady and RunnerExited.
//
// When Opts.Probe / Classify probe is nil, this function is used (not EmptyProbe).
//
// For unbound codex runners, best-effort EnsureCodexRunnerBound runs first so
// pollers/status re-reads pick up runner_session_id without callers inventing
// bind logic. Bind misses never fail the probe.
func LifecycleProbe(store agentstorage.Store, meta agentstorage.SessionMeta) (ProbeReport, error) {
	if store == nil {
		return ProbeReport{}, nil
	}
	// Persist codex runner_session_id when discoverable (scrollback / rollout).
	if strings.TrimSpace(meta.RunnerSessionID) == "" && isCodexRunner(meta.Runner) {
		meta, _ = EnsureCodexRunnerBound(store, meta, productionCodexBindOpts(store, meta))
	}
	layers := probeLifecycleLayers(store, meta)
	return ProbeReport{
		ResumeReady:  layers.resumeReady,
		RunnerExited: layers.exited,
	}, nil
}

// EmptyProbe always returns unknown lifecycle (ResumeReady false, RunnerExited nil).
// Use in unit tests that need ModeRun for a found session without TTY side effects.
func EmptyProbe(store agentstorage.Store, meta agentstorage.SessionMeta) (ProbeReport, error) {
	_ = store
	_ = meta
	return ProbeReport{}, nil
}

// lifecycleLayers is the internal multi-layer subset needed for Classify.
type lifecycleLayers struct {
	processStatus    string // serve PID: alive|dead|unknown
	processPID       int    // serve / registry PID
	commandPID       int    // PTY agent child (0 unknown)
	commandExited    bool   // registry recorded PTY Wait done
	commandAlive     *bool  // nil unknown; true/false when commandPID known
	terminalStatus   string // reachable|unreachable|missing
	terminalSendable string // yes|no
	runnerStatus     string // binding|bound|unbound
	scrollback       string
	exited           *bool
	resumeReady      bool
}

func probeLifecycleLayers(store agentstorage.Store, meta agentstorage.SessionMeta) lifecycleLayers {
	var layers lifecycleLayers

	termID := strings.TrimSpace(meta.TerminalSessionID)
	var ttySess *agenttty.TTYSession
	if termID != "" {
		if s, err := agenttty.ResolveByTerminalID(store.Home(), termID); err == nil {
			ttySess = s
		}
	}
	if ttySess == nil && termID == "" {
		if s, err := agenttty.ResolveTerminalStatus(store, meta.Runner, meta.SessionID); err == nil {
			ttySess = s
			termID = s.TerminalSessionID
		}
	}

	if ttySess != nil {
		layers.processPID = ttySess.Registry.PID
		if ttySess.Registry.PID > 0 {
			if ttywatch.ProcessAlive(ttySess.Registry.PID) {
				layers.processStatus = "alive"
			} else {
				layers.processStatus = "dead"
			}
		} else {
			layers.processStatus = "unknown"
		}
		// Phase 2: agent child PID from registry (preferred over serve PID).
		layers.commandPID = ttySess.Registry.CommandPID
		layers.commandExited = ttySess.Registry.CommandExited
		if layers.commandExited {
			f := false
			layers.commandAlive = &f
		} else if layers.commandPID > 0 {
			alive := ttywatch.ProcessAlive(layers.commandPID)
			layers.commandAlive = &alive
		}
		if ttySess.TCPReachable {
			layers.terminalStatus = "reachable"
			provider, ok := agenttty.Get(ttySess.RunnerID)
			if ok && provider.CheckWritable != nil {
				scrollbackText, err := ttywatch.SnapshotText(ttySess.Registry.ListenAddr, ttySess.TerminalSessionID)
				if err == nil && len(scrollbackText) > 0 {
					layers.scrollback = scrollbackText
					writable := provider.CheckWritable([]byte(scrollbackText))
					if writable.Ready {
						layers.terminalSendable = "yes"
					} else {
						layers.terminalSendable = "no"
					}
				} else {
					layers.terminalSendable = "no"
				}
			} else {
				if scrollbackText, err := ttywatch.SnapshotText(ttySess.Registry.ListenAddr, ttySess.TerminalSessionID); err == nil {
					layers.scrollback = scrollbackText
				}
				layers.terminalSendable = "no"
			}
		} else {
			layers.terminalStatus = "unreachable"
			layers.terminalSendable = "no"
		}
	} else if termID != "" {
		layers.processStatus = "dead"
		layers.terminalStatus = "missing"
		layers.terminalSendable = "no"
	} else {
		layers.processStatus = "unknown"
		layers.terminalStatus = "missing"
		layers.terminalSendable = "no"
	}

	runnerSID := strings.TrimSpace(meta.RunnerSessionID)
	if runnerSID != "" {
		layers.runnerStatus = "bound"
	} else if bindInProgress(store.Home(), meta.SessionID) {
		layers.runnerStatus = "binding"
	} else {
		layers.runnerStatus = "unbound"
	}

	layers.exited = computeRunnerExited(layers, meta)
	bound := layers.runnerStatus == "bound"
	exitedTrue := layers.exited != nil && *layers.exited
	layers.resumeReady = bound && exitedTrue
	return layers
}

func bindInProgress(home, sessionID string) bool {
	path := filepath.Join(home, "sessions", sessionID, "bind.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	var st struct {
		State string `json:"state"`
	}
	if json.Unmarshal(data, &st) != nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(st.State), "in_progress")
}

// computeRunnerExited returns true/false when known, nil when unknown.
// Parity with agentruncli: keep-alive serve may stay reachable after /exit.
//
// Signal priority:
//  1. command_pid / command_exited (PTY agent child) — primary
//  2. scrollback exit footer / [Terminal exited]
//  3. sendable + zombie serve (no children)
//
// Serve PID alone never means "agent live" under keep-alive.
func computeRunnerExited(layers lifecycleLayers, meta agentstorage.SessionMeta) *bool {
	t := true
	f := false

	// --- Primary: durable command_pid / command_exited from registry ---
	if layers.commandExited {
		return &t
	}
	if layers.commandAlive != nil {
		if !*layers.commandAlive {
			return &t // agent child dead
		}
		// Agent child still running → live (even if residual footers appear).
		return &f
	}

	if layers.terminalStatus == "reachable" {
		// Prefer exit markers over sendable (writable should already be "no"
		// after exit; this is belt-and-suspenders for residual glyphs).
		if scrollbackSuggestsAgentExited(layers.scrollback, meta.Runner) {
			return &t
		}
		if layers.terminalSendable == "yes" {
			return &f
		}
		if layers.terminalSendable == "no" &&
			strings.TrimSpace(layers.scrollback) != "" &&
			!scrollbackLooksLiveAgent(layers.scrollback) &&
			layers.processStatus == "alive" && layers.processPID > 0 &&
			!processHasChildren(layers.processPID) {
			return &t
		}
		// Reachable but no exit signal → still active.
		return &f
	}
	if layers.runnerStatus == "bound" {
		if layers.terminalStatus == "missing" || layers.terminalStatus == "unreachable" {
			return &t
		}
		if layers.processStatus == "dead" {
			return &t
		}
	}
	if meta.Status == "finished" || meta.Status == "error" {
		if layers.terminalStatus != "reachable" {
			return &t
		}
	}
	if (layers.runnerStatus == "unbound" || layers.runnerStatus == "binding") && layers.terminalStatus != "reachable" {
		return nil
	}
	return nil
}

// scrollbackSuggestsAgentExited delegates to agenttty (runner-scoped footers).
func scrollbackSuggestsAgentExited(scrollback, runner string) bool {
	return agenttty.ScrollbackSuggestsAgentExited(scrollback, runner)
}

func scrollbackLooksLiveAgent(scrollback string) bool {
	if strings.TrimSpace(scrollback) == "" {
		return false
	}
	if strings.Contains(scrollback, "Grok \u203a") || strings.Contains(scrollback, "Grok ›") ||
		strings.Contains(scrollback, "Grok >") {
		return true
	}
	if strings.Contains(scrollback, "\u203a") || strings.Contains(scrollback, "›") ||
		strings.Contains(scrollback, "❯") {
		return true
	}
	lower := strings.ToLower(scrollback)
	if strings.Contains(lower, "response:") || strings.Contains(lower, "submitted:") {
		return true
	}
	return false
}

func processHasChildren(pid int) bool {
	if pid <= 0 {
		return false
	}
	cmd := exec.Command("pgrep", "-P", strconv.Itoa(pid))
	out, err := cmd.Output()
	if err != nil {
		if len(bytes.TrimSpace(out)) == 0 {
			return false
		}
		return true
	}
	return len(bytes.TrimSpace(out)) > 0
}
