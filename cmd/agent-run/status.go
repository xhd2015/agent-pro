package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
	"github.com/xhd2015/agent-pro/pkgs/agenttty"
	"github.com/xhd2015/agent-pro/pkgs/ttywatch"
	"github.com/xhd2015/less-gen/flags"
)

const statusHelp = `
Usage: agent-run status [OPTIONS] [<session-id|runner/session_id>]

Show agent-run home (bare) or multi-layer session status.

With no arguments, prints the agent-run home path.

With a session id (or runner/session_id ref), probes storage, process,
terminal, runner (bound + exited), and resume readiness.

Options:
  --json       output multi-layer status as JSON
  -h, --help   show help
`

// sessionStatusReport is the multi-layer probe result for status / resume gates.
type sessionStatusReport struct {
	Session   string             `json:"session"`
	Status    string             `json:"status"`
	Workspace string             `json:"workspace,omitempty"`
	Process   processLayerReport `json:"process"`
	Terminal  terminalLayerReport `json:"terminal"`
	Runner    runnerLayerReport  `json:"runner"`
	Resume    resumeLayerReport  `json:"resume"`
}

type processLayerReport struct {
	Status string `json:"status"` // alive|dead|unknown
	PID    int    `json:"pid,omitempty"`
	Kind   string `json:"kind,omitempty"`
}

type terminalLayerReport struct {
	Status   string `json:"status"` // reachable|unreachable|missing
	ID       string `json:"id,omitempty"`
	Listen   string `json:"listen,omitempty"`
	Screen   string `json:"screen,omitempty"`
	Sendable string `json:"sendable,omitempty"` // yes|no
}

type runnerLayerReport struct {
	Status    string `json:"status"` // binding|bound|unbound
	Kind      string `json:"kind,omitempty"`
	SessionID string `json:"session_id,omitempty"`
	// Exited is true/false when known; nil means unknown (JSON null).
	Exited *bool `json:"exited"`
}

type resumeLayerReport struct {
	Ready  bool   `json:"ready"`
	Reason string `json:"reason,omitempty"`
}

func runStatus(args []string) error {
	var jsonFlag bool
	remaining, err := flags.Bool("--json", &jsonFlag).
		Help("-h,--help", statusHelp).
		Parse(args)
	if err != nil {
		return err
	}
	store, err := openStore()
	if err != nil {
		return err
	}
	if len(remaining) == 0 {
		fmt.Printf("home: %s\n", store.Home())
		return nil
	}
	ref := strings.TrimSpace(remaining[0])
	if ref == "" {
		fmt.Printf("home: %s\n", store.Home())
		return nil
	}
	meta, err := resolveSessionMeta(store, ref)
	if err != nil {
		return err
	}
	report := probeSessionStatus(store, meta)
	if jsonFlag {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(report); err != nil {
			return err
		}
		// Encode already appends a trailing newline.
		return nil
	}
	printSessionStatusHuman(report)
	return nil
}

// resolveSessionMeta resolves <session-id> or <runner>/<session_id>.
func resolveSessionMeta(store agentstorage.Store, ref string) (agentstorage.SessionMeta, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return agentstorage.SessionMeta{}, fmt.Errorf("session not found: empty")
	}
	if runner, id, ok := splitRunnerSessionRef(ref); ok {
		sess, err := store.GetSession(runner, id)
		if err != nil {
			return agentstorage.SessionMeta{}, fmt.Errorf("session not found: %s", ref)
		}
		return sess.Meta, nil
	}
	// Bare session id: unique match across runners.
	all, err := listAllSessions(store)
	if err != nil {
		return agentstorage.SessionMeta{}, err
	}
	var matches []agentstorage.SessionMeta
	for _, m := range all {
		if m.SessionID == ref {
			matches = append(matches, m)
		}
	}
	if len(matches) == 0 {
		return agentstorage.SessionMeta{}, fmt.Errorf("session not found: %s", ref)
	}
	if len(matches) > 1 {
		var refs []string
		for _, m := range matches {
			refs = append(refs, m.Runner+"/"+m.SessionID)
		}
		return agentstorage.SessionMeta{}, fmt.Errorf("session id %q is ambiguous; use one of: %s", ref, strings.Join(refs, ", "))
	}
	return matches[0], nil
}

func splitRunnerSessionRef(ref string) (runner, sessionID string, ok bool) {
	i := strings.IndexByte(ref, '/')
	if i <= 0 || i >= len(ref)-1 {
		return "", "", false
	}
	// Only first slash: runner/session_id (session ids should not contain /).
	runner = ref[:i]
	sessionID = ref[i+1:]
	if strings.Contains(sessionID, "/") {
		return "", "", false
	}
	return runner, sessionID, true
}

func probeSessionStatus(store agentstorage.Store, meta agentstorage.SessionMeta) sessionStatusReport {
	report := sessionStatusReport{
		Session:   meta.Runner + "/" + meta.SessionID,
		Status:    meta.Status,
		Workspace: meta.Workspace,
	}

	// --- process + terminal layers ---
	termID := strings.TrimSpace(meta.TerminalSessionID)
	report.Terminal.ID = termID
	report.Process.Kind = "serve"

	var ttySess *agenttty.TTYSession
	if termID != "" {
		if s, err := agenttty.ResolveByTerminalID(store.Home(), termID); err == nil {
			ttySess = s
		}
	}
	if ttySess == nil && termID == "" {
		// Try meta/registry scan used by terminal status probes.
		if s, err := agenttty.ResolveTerminalStatus(store, meta.Runner, meta.SessionID); err == nil {
			ttySess = s
			termID = s.TerminalSessionID
			report.Terminal.ID = termID
		}
	}

	// scrollbackSnapshot is terminal text used for sendable + exited heuristics.
	// Keep-alive serve stays TCP-reachable after agent /exit; markers live here.
	var scrollbackSnapshot string

	if ttySess != nil {
		report.Process.PID = ttySess.Registry.PID
		report.Terminal.Listen = ttySess.Registry.ListenAddr
		if ttySess.Registry.PID > 0 {
			if ttywatch.ProcessAlive(ttySess.Registry.PID) {
				report.Process.Status = "alive"
			} else {
				report.Process.Status = "dead"
			}
		} else {
			report.Process.Status = "unknown"
		}
		if ttySess.TCPReachable {
			report.Terminal.Status = "reachable"
			screen := ttySess.ScreenStatus
			provider, ok := agenttty.Get(ttySess.RunnerID)
			if ok && provider.CheckWritable != nil {
				scrollbackText, err := ttywatch.SnapshotText(ttySess.Registry.ListenAddr, ttySess.TerminalSessionID)
				if err == nil && len(scrollbackText) > 0 {
					scrollbackSnapshot = scrollbackText
					scrollback := []byte(scrollbackText)
					writable := provider.CheckWritable(scrollback)
					if writable.Ready {
						report.Terminal.Sendable = "yes"
					} else {
						report.Terminal.Sendable = "no"
					}
					if provider.DetectScreenStatus != nil && (screen == "" || screen == "unknown") {
						if live := provider.DetectScreenStatus(scrollback); live != "" {
							screen = live
						}
					}
				} else {
					report.Terminal.Sendable = "no"
				}
			} else {
				// Still try to snapshot for exit markers even without CheckWritable.
				if scrollbackText, err := ttywatch.SnapshotText(ttySess.Registry.ListenAddr, ttySess.TerminalSessionID); err == nil {
					scrollbackSnapshot = scrollbackText
				}
				report.Terminal.Sendable = "no"
			}
			if screen == "" {
				screen = "unknown"
			}
			report.Terminal.Screen = screen
		} else {
			report.Terminal.Status = "unreachable"
			report.Terminal.Sendable = "no"
			if ttySess.ScreenStatus != "" {
				report.Terminal.Screen = ttySess.ScreenStatus
			} else {
				report.Terminal.Screen = "unknown"
			}
		}
	} else if termID != "" {
		// Meta points at a terminal, but registry is gone.
		report.Process.Status = "dead"
		report.Terminal.Status = "missing"
		report.Terminal.Sendable = "no"
	} else {
		report.Process.Status = "unknown"
		report.Terminal.Status = "missing"
		report.Terminal.Sendable = "no"
	}

	// --- runner layer ---
	runnerSID := strings.TrimSpace(meta.RunnerSessionID)
	report.Runner.Kind = runnerKindFromMeta(meta.Runner)
	if runnerSID != "" {
		report.Runner.Status = "bound"
		report.Runner.SessionID = runnerSID
	} else if bindInProgress(store.Home(), meta.Runner, meta.SessionID) {
		// Mid-open background bind: durable bind.json state=in_progress.
		report.Runner.Status = "binding"
	} else {
		report.Runner.Status = "unbound"
	}

	exited := computeRunnerExited(report, meta, scrollbackSnapshot)
	report.Runner.Exited = exited

	// --- resume layer ---
	bound := report.Runner.Status == "bound"
	exitedTrue := exited != nil && *exited
	if bound && exitedTrue {
		report.Resume.Ready = true
	} else {
		report.Resume.Ready = false
		switch {
		case report.Runner.Status == "binding":
			report.Resume.Reason = "runner session bind in progress"
		case !bound:
			report.Resume.Reason = "runner session not bound (missing runner_session_id)"
		case exited != nil && !*exited:
			report.Resume.Reason = "runner still active (exited: false); use send, not resume"
		default:
			report.Resume.Reason = "runner exited state unknown; cannot resume"
		}
	}
	return report
}

// bindInProgress reports durable open-bind state for concurrent status.
func bindInProgress(home, runner, sessionID string) bool {
	path := filepath.Join(home, "sessions", runner, sessionID, "bind.json")
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

func runnerKindFromMeta(runner string) string {
	switch {
	case strings.HasPrefix(runner, "grok"):
		return "grok"
	case strings.HasPrefix(runner, "codex"):
		return "codex"
	default:
		if runner == "" {
			return ""
		}
		return runner
	}
}

// computeRunnerExited returns true/false when known, nil when unknown.
//
// Keep-alive __serve__ remains TCP-reachable after the agent child exits
// (zombie serve). Do not treat terminal.status==reachable alone as "still
// active"; inspect scrollback exit markers, sendable idle, and serve children.
func computeRunnerExited(report sessionStatusReport, meta agentstorage.SessionMeta, scrollback string) *bool {
	t := true
	f := false
	if report.Terminal.Status == "reachable" {
		// Truly live agent TUI: idle/sendable input prompt.
		if report.Terminal.Sendable == "yes" {
			return &f
		}
		// Zombie-after-/exit: exit footer / [Terminal exited] in snapshot.
		// Prefer scrollback markers over "no child" alone — fixtures (and some
		// hosts) register a non-serve PID that legitimately has zero children.
		if scrollbackSuggestsAgentExited(scrollback) {
			return &t
		}
		// Keep-alive serve still reachable but agent child is gone, and
		// snapshot is non-empty without a live idle prompt (sendable already no).
		if report.Terminal.Sendable == "no" &&
			strings.TrimSpace(scrollback) != "" &&
			!scrollbackLooksLiveAgent(scrollback) &&
			report.Process.Status == "alive" && report.Process.PID > 0 &&
			!processHasChildren(report.Process.PID) {
			return &t
		}
		// Reachable but no exit signal (busy/loading/unknown/snapshot miss) → still active.
		return &f
	}
	// Bound + terminal missing/unreachable → treat as exited.
	if report.Runner.Status == "bound" {
		if report.Terminal.Status == "missing" || report.Terminal.Status == "unreachable" {
			return &t
		}
		if report.Process.Status == "dead" {
			return &t
		}
	}
	// Finished storage with no live terminal → exited.
	if meta.Status == "finished" || meta.Status == "error" {
		if report.Terminal.Status != "reachable" {
			return &t
		}
	}
	// Unbound/binding without live terminal: unknown exited.
	if (report.Runner.Status == "unbound" || report.Runner.Status == "binding") && report.Terminal.Status != "reachable" {
		return nil
	}
	return nil
}

// scrollbackSuggestsAgentExited reports post-/exit markers in terminal snapshot.
// Matches left-aligned or indented "[Terminal exited]" and grok/codex resume footers.
func scrollbackSuggestsAgentExited(scrollback string) bool {
	if strings.TrimSpace(scrollback) == "" {
		return false
	}
	// Snapshot may retain ANSI; plain substring match is enough for markers.
	if strings.Contains(scrollback, "[Terminal exited]") {
		return true
	}
	// "Resume this session with:" / "grok --resume <id>" / "codex --resume <id>"
	lower := strings.ToLower(scrollback)
	if strings.Contains(lower, "grok --resume") || strings.Contains(lower, "codex --resume") {
		return true
	}
	if strings.Contains(lower, "resume this session with") {
		return true
	}
	return false
}

// scrollbackLooksLiveAgent reports idle TUI prompt markers (agent still interactive).
func scrollbackLooksLiveAgent(scrollback string) bool {
	if strings.TrimSpace(scrollback) == "" {
		return false
	}
	if strings.Contains(scrollback, "Grok \u203a") || strings.Contains(scrollback, "Grok ›") ||
		strings.Contains(scrollback, "Grok >") {
		return true
	}
	// Codex / generic prompt glyphs often mean interactive TUI.
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

// processHasChildren reports whether pid has at least one child process.
// Used to detect zombie keep-alive serve (alive PID, agent child gone).
func processHasChildren(pid int) bool {
	if pid <= 0 {
		return false
	}
	// pgrep -P lists direct children; exit 1 means none (or error).
	cmd := exec.Command("pgrep", "-P", strconv.Itoa(pid))
	out, err := cmd.Output()
	if err != nil {
		// No children or pgrep unavailable — treat as no children only when
		// output is empty; on hard errors prefer "has children" (not exited)
		// only if we cannot tell — actually: empty → no children.
		if len(bytes.TrimSpace(out)) == 0 {
			return false
		}
		return true
	}
	return len(bytes.TrimSpace(out)) > 0
}

func printSessionStatusHuman(r sessionStatusReport) {
	// Use column-aligned labels similar to the design layout.
	fmt.Printf("session:   %s\n", r.Session)
	fmt.Printf("status:    %s\n", r.Status)
	if r.Workspace != "" {
		fmt.Printf("workspace: %s\n", r.Workspace)
	}
	fmt.Println()
	fmt.Println("process:")
	fmt.Printf("  status:  %s\n", r.Process.Status)
	if r.Process.PID > 0 {
		fmt.Printf("  pid:     %d\n", r.Process.PID)
	}
	if r.Process.Kind != "" {
		fmt.Printf("  kind:    %s\n", r.Process.Kind)
	}
	fmt.Println()
	fmt.Println("terminal:")
	fmt.Printf("  status:   %s\n", r.Terminal.Status)
	if r.Terminal.ID != "" {
		fmt.Printf("  id:       %s\n", r.Terminal.ID)
	}
	if r.Terminal.Listen != "" {
		fmt.Printf("  listen:   %s\n", r.Terminal.Listen)
	}
	if r.Terminal.Screen != "" {
		fmt.Printf("  screen:   %s\n", r.Terminal.Screen)
	}
	if r.Terminal.Sendable != "" {
		fmt.Printf("  sendable: %s\n", r.Terminal.Sendable)
	}
	fmt.Println()
	fmt.Println("runner:")
	fmt.Printf("  status:     %s\n", r.Runner.Status)
	if r.Runner.Kind != "" {
		fmt.Printf("  kind:       %s\n", r.Runner.Kind)
	}
	fmt.Printf("  session_id: %s\n", r.Runner.SessionID)
	exitedStr := "unknown"
	if r.Runner.Exited != nil {
		if *r.Runner.Exited {
			exitedStr = "true"
		} else {
			exitedStr = "false"
		}
	}
	fmt.Printf("  exited:     %s\n", exitedStr)
	fmt.Println()
	fmt.Println("resume:")
	if r.Resume.Ready {
		fmt.Printf("  ready: yes\n")
	} else {
		fmt.Printf("  ready: no\n")
		if r.Resume.Reason != "" {
			fmt.Printf("  reason: %s\n", r.Resume.Reason)
		}
	}
}
