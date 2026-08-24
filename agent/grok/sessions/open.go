package sessions

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
)

// OpenHelp is the text for `agent-pro grok session open --help`.
const OpenHelp = `Usage: agent-pro grok session open (<session-id> | --tab SEL | --tab-index N) [OPTIONS]

Focus the iTerm2 tab that already hosts this Grok session when one exists.
Otherwise open a new iTerm2 window and run: grok --resume <session-id>
When the Grok id is bound in agent-run (grok/grok-tty runner_session_id),
prefers agent-run (live → focus; exited → agent-run resume) instead of bare
grok --resume.

Session source (exactly one):
  <session-id>          explicit Grok session id
  --tab SEL             1-based tab index, or next|left|right (right ≡ next)
  --tab-index N         0-based tab index in this iTerm window

Options:
  --index N             select candidate N when multiple tabs host the same session
                        (positional <session-id> only; not with --tab/--tab-index)
  --dir DIR             workspace for resume (default: session cwd)
  --no-agent-run        force bare grok --resume (skip agent-run prefer)
  --dry-run             resolve only; do not focus or open a window
  -h,--help             show help

Tab selectors use the same window/tab discovery as: kool iterm2 window status.
Relative next/left/right do not wrap; edges error.
A successful --tab/--tab-index resolve focuses that tab (never resumes).
`

// OpenCommandHelpLine is the parent `agent-pro grok session` help row.
const OpenCommandHelpLine = `  open   …               focus hosting tab or resume (--tab / --tab-index / <id>)`

const (
	OpenActionFocused = "focused"
	OpenActionResumed = "resumed"
)

// OpenOpts drives Open / RunOpen. Nil hooks use production probes and launchers.
type OpenOpts struct {
	Index  *int
	Dir    string // optional workspace override for resume
	DryRun bool

	// NoAgentRun skips agent-run prefer (--no-agent-run).
	NoAgentRun bool
	// AgentRunHome overrides AGENT_RUN_HOME / ~/.agent-run for production prefer.
	AgentRunHome string
	// AgentRunBin overrides LookPath("agent-run") for resume/run window commands.
	AgentRunBin string

	// TabFrom, when set, short-circuits focus to this already-resolved tab
	// (used by RunOpen after ResolveFromTab). Resume is skipped.
	TabFrom *TabResolveResult

	ListProcs  func() []FocusProc
	Lsof       func(int) []string
	ListITerm  func() ([]iterm2.SessionRef, error)
	FocusITerm func(iterm2.SessionRef) error

	// Tab resolve hooks (nil → production). Used by RunOpen for --tab/--tab-index.
	CurrentSessionID func() string
	ControllingTTY   func() string
	AncestorTTYs     func() []string

	OpenInNewWindow func(dir, followUp string) error
	GrokBin         string
	LookPath        func(file string) (string, error)
	Env             []string

	// AgentRunOpen replaces production agent-run prefer (tests).
	AgentRunOpen func(grokSessionID, prompt string) (*AgentRunOpenResult, error)

	Stderr io.Writer // warnings (live-but-no-iTerm); nil → os.Stderr
}

// OpenResult is the outcome of a successful Open.
type OpenResult struct {
	Action            string // OpenActionFocused | OpenActionResumed
	Candidate         FocusCandidate
	CWD               string
	Command           string
	ViaAgentRun       bool
	AgentRunSessionID string
	AgentRunMode      string
}

// OpenFake is the deterministic injected boundary used by open tests.
type OpenFake struct {
	FocusFake
	Opened           []string // "dir|followUp" entries
	CurrentSessionID string
	ControllingTTY   string

	AgentRunByID  map[string]*AgentRunOpenResult
	AgentRunErr   error
	AgentRunCalls []string
}

// OpenOpts returns OpenOpts wired to this fake.
func (f *OpenFake) OpenOpts() *OpenOpts {
	fo := f.FocusFake.Opts()
	opts := &OpenOpts{
		ListProcs:  fo.ListProcs,
		Lsof:       fo.Lsof,
		ListITerm:  fo.ListITerm,
		FocusITerm: fo.FocusITerm,
		CurrentSessionID: func() string {
			return f.CurrentSessionID
		},
		ControllingTTY: func() string {
			return f.ControllingTTY
		},
		AncestorTTYs: func() []string { return nil },
		OpenInNewWindow: func(dir, followUp string) error {
			f.Opened = append(f.Opened, dir+"|"+followUp)
			return nil
		},
		GrokBin:     "/usr/local/bin/grok",
		AgentRunBin: "/usr/local/bin/agent-run",
		LookPath: func(file string) (string, error) {
			switch file {
			case "agent-run":
				return "/usr/local/bin/agent-run", nil
			default:
				return "/usr/local/bin/grok", nil
			}
		},
		Env: []string{"PATH=/usr/bin"},
	}
	if f.AgentRunByID != nil || f.AgentRunErr != nil {
		opts.AgentRunOpen = func(grokSessionID, prompt string) (*AgentRunOpenResult, error) {
			f.AgentRunCalls = append(f.AgentRunCalls, grokSessionID+"|"+prompt)
			if f.AgentRunErr != nil {
				return nil, f.AgentRunErr
			}
			hit, ok := f.AgentRunByID[grokSessionID]
			if !ok || hit == nil {
				return nil, nil
			}
			out := *hit
			if out.Mode == "" {
				out.Mode = AgentRunOpenModeFocus
				out.Focused = true
			}
			return &out, nil
		}
	}
	return opts
}

// Open focuses an existing iTerm host tab for sessionID, or resumes the session
// in a new iTerm2 window when no hosting tab is found.
// When opts.TabFrom is set, focuses that tab directly (no rediscovery, no resume).
func Open(grokHome, sessionID string, opts *OpenOpts) (*OpenResult, error) {
	if opts == nil {
		opts = &OpenOpts{}
	}
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return nil, fmt.Errorf("session id is required")
	}

	info, err := Info(grokHome, sessionID)
	if err != nil {
		return nil, err
	}

	if opts.TabFrom != nil {
		selected := focusCandidateFromTab(opts.TabFrom)
		return focusOpenCandidate(selected, opts)
	}

	focusOpts := &FocusOpts{
		Index:      opts.Index,
		ListProcs:  opts.ListProcs,
		Lsof:       opts.Lsof,
		ListITerm:  opts.ListITerm,
		FocusITerm: opts.FocusITerm,
	}
	disc, err := DiscoverFocusHosting(sessionID, focusOpts)
	if err != nil {
		return nil, err
	}
	if disc != nil && len(disc.Candidates) > 0 {
		selected, selErr := selectFocusCandidate(sessionID, disc.Candidates, opts.Index)
		if selErr != nil {
			msg := strings.ReplaceAll(selErr.Error(),
				"agent-pro grok session focus",
				"agent-pro grok session open")
			return nil, fmt.Errorf("%s", msg)
		}
		return focusOpenCandidate(selected, opts)
	}

	stderr := opts.Stderr
	if stderr == nil {
		stderr = os.Stderr
	}

	arHit, arWarn, arErr := preferAgentRunOpen(agentRunOpenHooks{
		NoAgentRun:      opts.NoAgentRun,
		AgentRunHome:    opts.AgentRunHome,
		DryRun:          opts.DryRun,
		Dir:             opts.Dir,
		Info:            info,
		SkipProduction:  opts.ListProcs != nil || opts.ListITerm != nil,
		OpenInNewWindow: opts.OpenInNewWindow,
		AgentRunBin:     opts.AgentRunBin,
		LookPath:        opts.LookPath,
		Stderr:          stderr,
		AgentRunOpen:    opts.AgentRunOpen,
	}, sessionID, "")
	writeAgentRunOpenWarn(stderr, arWarn)
	if arErr != nil {
		return nil, arErr
	}
	if arHit != nil {
		cmd := arHit.Command
		if arHit.Opened && !opts.DryRun && opts.AgentRunOpen != nil {
			if cmd == "" {
				cmd = "agent-run run --session-id " + arHit.AgentRunSessionID + " --auto-send-or-resume --open"
			}
			openFn := opts.OpenInNewWindow
			if openFn == nil {
				openFn = defaultOpenInNewWindow
			}
			if err := openFn(arHit.CWD, cmd); err != nil {
				return nil, fmt.Errorf("agent-run deliver failed: open window: %w", err)
			}
		}
		if arHit.Focused {
			return &OpenResult{
				Action:            OpenActionFocused,
				ViaAgentRun:       true,
				AgentRunSessionID: arHit.AgentRunSessionID,
				AgentRunMode:      arHit.Mode,
				CWD:               arHit.CWD,
			}, nil
		}
		return &OpenResult{
			Action:            OpenActionResumed,
			ViaAgentRun:       true,
			AgentRunSessionID: arHit.AgentRunSessionID,
			AgentRunMode:      arHit.Mode,
			CWD:               arHit.CWD,
			Command:           cmd,
		}, nil
	}

	cwd, err := resolveOpenCWD(info, opts.Dir)
	if err != nil {
		return nil, err
	}
	bin, err := resolveForkGrokBin(opts.GrokBin, opts.LookPath)
	if err != nil {
		return nil, err
	}
	argv := []string{"--resume", sessionID}
	cmdLine := quotedForkCommandLine(bin, argv)

	if disc != nil && disc.LiveCount > 0 {
		fmt.Fprintf(stderr, "warning: session has live grok PID(s) but no matching iTerm tab; opening a new window\n")
	}

	if opts.DryRun {
		return &OpenResult{
			Action:  OpenActionResumed,
			CWD:     cwd,
			Command: cmdLine,
		}, nil
	}

	openFn := opts.OpenInNewWindow
	if openFn == nil {
		openFn = defaultOpenInNewWindow
	}
	if err := openFn(cwd, cmdLine); err != nil {
		return nil, fmt.Errorf("open new window: %w", err)
	}
	return &OpenResult{
		Action:  OpenActionResumed,
		CWD:     cwd,
		Command: cmdLine,
	}, nil
}

func focusCandidateFromTab(tr *TabResolveResult) FocusCandidate {
	if tr == nil {
		return FocusCandidate{}
	}
	return FocusCandidate{
		WindowID:    tr.WindowID,
		WindowTitle: tr.WindowName,
		TabIndex:    tr.TabIndex,
		SessionID:   tr.ITermSession,
		TTY:         tr.TTY,
		PID:         tr.RunnerPID,
	}
}

func focusOpenCandidate(selected FocusCandidate, opts *OpenOpts) (*OpenResult, error) {
	if opts.DryRun {
		return &OpenResult{
			Action:    OpenActionFocused,
			Candidate: selected,
		}, nil
	}
	focusFn := opts.FocusITerm
	if focusFn == nil {
		focusFn = func(ref iterm2.SessionRef) error {
			return iterm2.Focus(ref, nil)
		}
	}
	if err := focusFn(iterm2.SessionRef{
		WindowID:  selected.WindowID,
		TabIndex:  selected.TabIndex,
		SessionID: selected.SessionID,
		TTY:       selected.TTY,
	}); err != nil {
		return nil, err
	}
	return &OpenResult{Action: OpenActionFocused, Candidate: selected}, nil
}

// RunOpen implements `agent-pro grok session open` with injectable writers/hooks.
func RunOpen(args []string, stdout, stderr io.Writer, grokHome string, opts *OpenOpts) error {
	if stdout == nil {
		stdout = io.Discard
	}
	if stderr == nil {
		stderr = io.Discard
	}
	parsed, err := parseOpenArgs(args)
	if err != nil {
		return err
	}
	if parsed.Help {
		txt := OpenHelp
		if !strings.HasSuffix(txt, "\n") {
			txt += "\n"
		}
		_, _ = io.WriteString(stdout, txt)
		return nil
	}

	runOpts := OpenOpts{}
	if opts != nil {
		runOpts = *opts
	}
	runOpts.Index = parsed.Index
	runOpts.Dir = parsed.Dir
	runOpts.DryRun = parsed.DryRun
	runOpts.NoAgentRun = parsed.NoAgentRun
	if runOpts.Stderr == nil {
		runOpts.Stderr = stderr
	}

	sessionID, tabMeta, err := ResolveSessionSource(parsed.Positional, parsed.Tab, parsed.TabIndex, &SessionSourceOpts{
		ListProcs:        runOpts.ListProcs,
		Lsof:             runOpts.Lsof,
		ListITerm:        runOpts.ListITerm,
		CurrentSessionID: runOpts.CurrentSessionID,
		ControllingTTY:   runOpts.ControllingTTY,
		AncestorTTYs:     runOpts.AncestorTTYs,
	})
	if err != nil {
		return err
	}
	if tabMeta != nil {
		if parsed.Index != nil {
			return fmt.Errorf("--index cannot be combined with --tab/--tab-index")
		}
		runOpts.TabFrom = tabMeta
	}

	result, err := Open(grokHome, sessionID, &runOpts)
	if err != nil {
		return err
	}
	switch result.Action {
	case OpenActionFocused:
		if parsed.DryRun {
			if result.ViaAgentRun {
				fmt.Fprintln(stdout, "Would focus via agent-run")
				fmt.Fprintf(stdout, "  grok id:       %s\n", sessionID)
				fmt.Fprintf(stdout, "  agent-run id:  %s\n", result.AgentRunSessionID)
				return nil
			}
			fmt.Fprintf(stdout, "Would focus: window %s, tab %d\n", result.Candidate.WindowID, result.Candidate.TabIndex)
			return nil
		}
		if result.ViaAgentRun {
			fmt.Fprintf(stdout, "focused: agent-run %s\n", result.AgentRunSessionID)
		} else {
			fmt.Fprintf(stdout, "focused: window %s, tab %d\n", result.Candidate.WindowID, result.Candidate.TabIndex)
		}
	case OpenActionResumed:
		if parsed.DryRun {
			fmt.Fprintln(stdout, "Would open new iTerm2 window")
			fmt.Fprintf(stdout, "  grok id:  %s\n", sessionID)
			if result.ViaAgentRun {
				fmt.Fprintf(stdout, "  agent-run id: %s\n", result.AgentRunSessionID)
			}
			fmt.Fprintf(stdout, "  cwd:      %s\n", result.CWD)
			fmt.Fprintf(stdout, "  command:  %s\n", result.Command)
			return nil
		}
		if result.ViaAgentRun {
			fmt.Fprintf(stdout, "opened: new window; agent-run resume %s\n", result.AgentRunSessionID)
		} else {
			fmt.Fprintf(stdout, "opened: new window; resuming %s\n", sessionID)
		}
	default:
		return fmt.Errorf("internal: unknown open action %q", result.Action)
	}
	return nil
}

type openArgs struct {
	Positional []string
	Index      *int
	Tab        *string
	TabIndex   *int
	Dir        string
	DryRun     bool
	NoAgentRun bool
	Help       bool
}

func parseOpenArgs(args []string) (openArgs, error) {
	var out openArgs
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "-h" || arg == "--help" {
			out.Help = true
			return out, nil
		}
		if arg == "--dry-run" {
			out.DryRun = true
			continue
		}
		if arg == "--no-agent-run" {
			out.NoAgentRun = true
			continue
		}
		if arg == "--index" {
			if i+1 >= len(args) {
				return out, fmt.Errorf("--index must be an integer")
			}
			n, convErr := strconv.Atoi(args[i+1])
			if convErr != nil {
				return out, fmt.Errorf("--index must be an integer")
			}
			out.Index = &n
			i++
			continue
		}
		if strings.HasPrefix(arg, "--index=") {
			n, convErr := strconv.Atoi(strings.TrimPrefix(arg, "--index="))
			if convErr != nil {
				return out, fmt.Errorf("--index must be an integer")
			}
			out.Index = &n
			continue
		}
		if arg == "--tab" {
			if i+1 >= len(args) {
				return out, fmt.Errorf("--tab requires a value (1-based index, or next|left|right)")
			}
			v := args[i+1]
			out.Tab = &v
			i++
			continue
		}
		if strings.HasPrefix(arg, "--tab=") {
			v := strings.TrimPrefix(arg, "--tab=")
			out.Tab = &v
			continue
		}
		if arg == "--tab-index" {
			if i+1 >= len(args) {
				return out, fmt.Errorf("--tab-index must be an integer")
			}
			n, convErr := strconv.Atoi(args[i+1])
			if convErr != nil {
				return out, fmt.Errorf("--tab-index must be an integer")
			}
			out.TabIndex = &n
			i++
			continue
		}
		if strings.HasPrefix(arg, "--tab-index=") {
			n, convErr := strconv.Atoi(strings.TrimPrefix(arg, "--tab-index="))
			if convErr != nil {
				return out, fmt.Errorf("--tab-index must be an integer")
			}
			out.TabIndex = &n
			continue
		}
		if arg == "--dir" {
			if i+1 >= len(args) {
				return out, fmt.Errorf("--dir requires a path")
			}
			out.Dir = args[i+1]
			i++
			continue
		}
		if strings.HasPrefix(arg, "--dir=") {
			out.Dir = strings.TrimPrefix(arg, "--dir=")
			continue
		}
		if strings.HasPrefix(arg, "-") {
			return out, fmt.Errorf("unrecognized flag: %s", arg)
		}
		out.Positional = append(out.Positional, arg)
	}
	return out, nil
}

func resolveOpenCWD(info *SessionInfo, dirOverride string) (string, error) {
	cwd := strings.TrimSpace(dirOverride)
	if cwd == "" && info != nil {
		cwd = strings.TrimSpace(info.CWD)
	}
	if cwd == "" {
		id := ""
		if info != nil {
			id = info.ID
		}
		return "", fmt.Errorf("session %s has empty cwd; pass --dir", id)
	}
	abs, err := filepath.Abs(cwd)
	if err != nil {
		return "", fmt.Errorf("workspace dir: %w", err)
	}
	if real, e := filepath.EvalSymlinks(abs); e == nil {
		abs = real
	}
	st, err := os.Stat(abs)
	if err != nil {
		return "", fmt.Errorf("workspace dir: %w", err)
	}
	if !st.IsDir() {
		return "", fmt.Errorf("workspace dir: not a directory: %s", abs)
	}
	return abs, nil
}

func defaultOpenInNewWindow(dir, followUp string) error {
	return iterm2.OpenConfig(dir, &iterm2.Config{
		Mode:             iterm2.ModeForceNew,
		FollowUpCommands: []string{followUp},
		SafeInputIgnore:  true,
	})
}
