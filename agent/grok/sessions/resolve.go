package sessions

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/xhd2015/agent-pro/pkgs/procresolve"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
	"github.com/xhd2015/less-gen/flags"
)

// ResolveHelp is the text for `agent-pro grok session resolve --help`.
const ResolveHelp = `Usage: agent-pro grok session resolve [OPTIONS]

Resolve a Grok session id either by walking ancestors to the nearest
grok runner (default), or from a sibling iTerm2 tab in this window.

Options:
  --pid PID         start pid for ancestor walk (default: current process)
  --tab SEL         1-based tab index, or next|left|right (right ≡ next)
  --tab-index N     0-based tab index in this iTerm window
  --dry-run         print resolution plan ([dry-run] lines); same discovery path
  -v,--verbose      print detail fields on stderr (ancestor or tab)
  --json            print session id + detail fields as JSON
  -h,--help         show help

Exactly one session source: ancestor walk (default / --pid), or --tab, or
--tab-index. --pid cannot combine with --tab/--tab-index.
Relative next/left/right do not wrap; edges error.
Tab discovery matches: kool iterm2 window status.
When a parent and its child subagent share a tab, the parent id is returned;
unrelated multiple grok sessions on the same tab still refuse.
`

// ResolveCommandHelpLine is the parent `agent-pro grok session` help row.
const ResolveCommandHelpLine = `  resolve               resolve Grok session id (ancestor walk or --tab)`

// ResolveOpts drives RunResolve. Nil hooks use live process listing and lsof.
// PID is the default start pid when --pid is omitted (production: os.Getpid()).
type ResolveOpts struct {
	Stdout, Stderr io.Writer
	PID            int
	ListProcs      func() []procresolve.Proc
	Lsof           func(pid int) []string
	GrokHome       string

	// Tab path hooks (nil → production probes used by ResolveFromTab).
	ListFocusProcs   func() []FocusProc
	ListITerm        func() ([]iterm2.SessionRef, error)
	CurrentSessionID func() string
	ControllingTTY   func() string
	AncestorTTYs     func() []string
	SessionMeta      func(sessionID string) (TabSessionMeta, bool)
}

// ResolveDetails is the machine-readable / verbose detail set for a hit.
type ResolveDetails struct {
	SessionID   string `json:"session_id"`
	Mode        string `json:"mode,omitempty"` // "ancestor" | "tab"
	StartPID    int    `json:"start_pid,omitempty"`
	AncestorPID int    `json:"ancestor_pid,omitempty"`
	RunnerPID   int    `json:"runner_pid"`
	Source      string `json:"source"`
	Confidence  string `json:"confidence"`
	WindowID    string `json:"window_id,omitempty"`
	TabIndex    int    `json:"tab_index,omitempty"`
	TTY         string `json:"tty,omitempty"`
}

// RunResolve implements `agent-pro grok session resolve`.
// It does not os.Exit and does not print an `Error:` / `agent-pro:` prefix.
func RunResolve(args []string, opts *ResolveOpts) error {
	if opts == nil {
		opts = &ResolveOpts{}
	}
	applyResolveDefaults(opts)

	var pidFlag *int
	var tabFlag *string
	var tabIndexFlag *int
	var dryRun bool
	var verbose bool
	var jsonOut bool

	stdout := opts.Stdout
	remaining, err := flags.Int("--pid", &pidFlag).
		String("--tab", &tabFlag).
		Int("--tab-index", &tabIndexFlag).
		Bool("--dry-run", &dryRun).
		Bool("-v,--verbose", &verbose).
		Bool("--json", &jsonOut).
		HelpFunc("-h,--help", func() {
			txt := strings.TrimPrefix(ResolveHelp, "\n")
			fmt.Fprint(stdout, txt)
			if !strings.HasSuffix(txt, "\n") {
				fmt.Fprintln(stdout)
			}
		}).
		HelpNoExit().
		Parse(args)
	if err == flags.ErrHelp {
		return nil
	}
	if err != nil {
		return err
	}
	if len(remaining) > 0 {
		return fmt.Errorf("unexpected argument: %s", remaining[0])
	}

	tabSet := tabFlag != nil
	tabIndexSet := tabIndexFlag != nil
	if tabSet && tabIndexSet {
		return fmt.Errorf("--tab and --tab-index cannot be specified together")
	}
	if pidFlag != nil && (tabSet || tabIndexSet) {
		return fmt.Errorf("--pid and --tab/--tab-index cannot be specified together")
	}

	var details ResolveDetails
	if tabSet || tabIndexSet {
		var sel TabSelector
		if tabSet {
			sel, err = ParseTabFlag(*tabFlag)
		} else {
			sel, err = ParseTabIndexFlag(*tabIndexFlag)
		}
		if err != nil {
			return err
		}
		tabOpts := &TabResolveOpts{
			ListProcs:        opts.ListFocusProcs,
			Lsof:             opts.Lsof,
			ListITerm:        opts.ListITerm,
			CurrentSessionID: opts.CurrentSessionID,
			ControllingTTY:   opts.ControllingTTY,
			AncestorTTYs:     opts.AncestorTTYs,
			GrokHome:         opts.GrokHome,
			SessionMeta:      opts.SessionMeta,
		}
		tr, err := ResolveFromTab(sel, tabOpts)
		if err != nil {
			return err
		}
		details = ResolveDetails{
			SessionID:  tr.SessionID,
			Mode:       "tab",
			RunnerPID:  tr.RunnerPID,
			Source:     tr.Source,
			Confidence: tr.Confidence,
			WindowID:   tr.WindowID,
			TabIndex:   tr.TabIndex,
			TTY:        tr.TTY,
		}
	} else {
		startPID := opts.PID
		if pidFlag != nil {
			startPID = *pidFlag
		}

		resolveOpts := procresolve.Options{
			GrokHome:  opts.GrokHome,
			ListProcs: opts.ListProcs,
			Lsof:      opts.Lsof,
		}

		anc, ok := procresolve.FindAncestorGrok(startPID, resolveOpts)
		result, err := procresolve.ResolveFromAncestors(startPID, resolveOpts)
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("no ancestor grok")
		}
		if result == nil || strings.TrimSpace(result.SessionID) == "" {
			return fmt.Errorf("session not resolved")
		}

		details = ResolveDetails{
			SessionID:   result.SessionID,
			Mode:        "ancestor",
			StartPID:    startPID,
			AncestorPID: anc.PID,
			RunnerPID:   result.RunnerPID,
			Source:      result.Source,
			Confidence:  result.Confidence,
		}
	}

	// One pipeline; gate only the consumer-facing success shape.
	if jsonOut {
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(details); err != nil {
			return err
		}
		return nil
	}
	if dryRun {
		writeResolveDryRunPlan(stdout, details)
		return nil
	}

	fmt.Fprintln(stdout, details.SessionID)
	if verbose {
		writeResolveVerbose(opts.Stderr, details)
	}
	return nil
}

func applyResolveDefaults(opts *ResolveOpts) {
	if opts.Stdout == nil {
		opts.Stdout = os.Stdout
	}
	if opts.Stderr == nil {
		opts.Stderr = os.Stderr
	}
	if opts.PID == 0 {
		opts.PID = os.Getpid()
	}
	if opts.ListProcs == nil {
		opts.ListProcs = procresolve.ListLiveProcs
	}
	if opts.Lsof == nil {
		opts.Lsof = procresolve.LiveLsof
	}
}

func writeResolveDryRunPlan(w io.Writer, d ResolveDetails) {
	if d.Mode == "tab" {
		fmt.Fprintf(w, "[dry-run] mode:          tab\n")
		fmt.Fprintf(w, "[dry-run] window:        %s\n", d.WindowID)
		fmt.Fprintf(w, "[dry-run] tab index:     %d\n", d.TabIndex)
		fmt.Fprintf(w, "[dry-run] tty:           %s\n", d.TTY)
		fmt.Fprintf(w, "[dry-run] runner pid:    %d\n", d.RunnerPID)
		fmt.Fprintf(w, "[dry-run] would resolve: %s\n", d.SessionID)
		fmt.Fprintf(w, "[dry-run] source:        %s\n", d.Source)
		fmt.Fprintf(w, "[dry-run] confidence:    %s\n", d.Confidence)
		return
	}
	fmt.Fprintf(w, "[dry-run] start pid:     %d\n", d.StartPID)
	fmt.Fprintf(w, "[dry-run] ancestor pid:  %d\n", d.AncestorPID)
	fmt.Fprintf(w, "[dry-run] runner pid:    %d\n", d.RunnerPID)
	fmt.Fprintf(w, "[dry-run] would resolve: %s\n", d.SessionID)
	fmt.Fprintf(w, "[dry-run] source:        %s\n", d.Source)
	fmt.Fprintf(w, "[dry-run] confidence:    %s\n", d.Confidence)
}

func writeResolveVerbose(w io.Writer, d ResolveDetails) {
	if d.Mode == "tab" {
		fmt.Fprintf(w, "mode:         tab\n")
		fmt.Fprintf(w, "window:       %s\n", d.WindowID)
		fmt.Fprintf(w, "tab index:    %d\n", d.TabIndex)
		fmt.Fprintf(w, "tty:          %s\n", d.TTY)
		fmt.Fprintf(w, "runner pid:   %d\n", d.RunnerPID)
		fmt.Fprintf(w, "source:       %s\n", d.Source)
		fmt.Fprintf(w, "confidence:   %s\n", d.Confidence)
		return
	}
	fmt.Fprintf(w, "start pid:    %d\n", d.StartPID)
	fmt.Fprintf(w, "ancestor pid: %d\n", d.AncestorPID)
	fmt.Fprintf(w, "runner pid:   %d\n", d.RunnerPID)
	fmt.Fprintf(w, "source:       %s\n", d.Source)
	fmt.Fprintf(w, "confidence:   %s\n", d.Confidence)
}
