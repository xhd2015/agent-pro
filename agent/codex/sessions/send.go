package sessions

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/xhd2015/agent-pro/pkgs/agenttty"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
	lessflags "github.com/xhd2015/less-flags"
)

// SendHelp is the text for `agent-pro codex session send --help`.
const SendHelp = `Usage: agent-pro codex session send [text] (--session-id <id> | --tab SEL | --tab-index N) [OPTIONS]

Type text and/or key sequences into the live iTerm2 pane that hosts a Codex session.
Same write-text path as: kool iterm2 session <iterm-uuid> send …
By default requires a hosting iTerm tab. With --open, resumes in a new
window when no host is found, waits for the tab to appear, then sends.
When the Codex id is bound in agent-run (codex/codex-tty runner_session_id),
--session-id prefers agent-run auto-send-or-resume directly (live → send
queue; exited → resume) with no iTerm discovery or SendText.
--tab / --tab-index still target iTerm panes explicitly.

Session source (exactly one):
  --session-id ID       Codex session id
  --tab SEL             1-based tab index, or next|left|right (right ≡ next)
  --tab-index N         0-based tab index in this iTerm window

Options:
  --index N             select candidate N when multiple tabs host the same session
                        (--session-id only; not with --tab/--tab-index)
  --no-submit           write without newline (stage; user presses Enter)
  --focus               switch to the session's window/tab before writing
  --no-ctrl-u           do not prefix Ctrl-U (default prefixes Ctrl-U)
  --open                if no hosting tab: resume in a new window, then send
  --no-agent-run        force iTerm path (skip agent-run prefer for --session-id)
  --dir DIR             workspace for --open resume (default: session cwd)
  --dry-run             resolve only; do not open or call SendText
  --enter               append Enter (\n) to the send sequence
  --up,--down,--left,--right   append arrow key (CSI)
  --esc                 append Escape
  --ctrl-c,--ctrl-d     append Ctrl-C / Ctrl-D
  --text STR            append text in sequence order (interleaves with keys)
  -h,--help             show help

At least one of [text], --text, or a key flag is required.
Sequence flags (--enter/--up/--text/…) keep CLI order and may repeat.
Positional [text], when present, is always appended last after the sequence.
Tab selectors use the same window/tab discovery as: kool iterm2 window status.
--open cannot be combined with --tab/--tab-index.
`

// SendCommandHelpLine is the parent `agent-pro codex session` help row.
const SendCommandHelpLine = `  send   …               type text into hosting pane (--session-id / --tab / --open)`

// DefaultSendOpenWait is how long --open waits for a hosting tab after resume.
const DefaultSendOpenWait = 120 * time.Second

// sendOpenPollInterval is the DiscoverFocusHosting poll gap after --open resume.
const sendOpenPollInterval = 200 * time.Millisecond

// SendOpts drives Send / RunSend. Nil hooks use production probes and SendText.
type SendOpts struct {
	Index    *int
	Dir      string
	DryRun   bool
	Open     bool
	Focus    bool
	NoSubmit bool
	NoCtrlU  bool

	// NoAgentRun skips agent-run prefer on --open (--no-agent-run).
	NoAgentRun bool
	// AgentRunHome overrides AGENT_RUN_HOME / ~/.agent-run for production prefer.
	AgentRunHome string
	// AgentRunBin overrides LookPath("agent-run") for resume/run window commands.
	AgentRunBin string

	// OpenWait caps how long --open polls for a hosting tab. Zero → DefaultSendOpenWait.
	OpenWait time.Duration

	// TabFrom, when set, short-circuits host discovery to this already-resolved tab.
	TabFrom *TabResolveResult

	ListProcs func() []FocusProc
	Lsof      func(int) []string
	ListITerm func() ([]iterm2.SessionRef, error)

	// Tab resolve hooks (nil → production). Used by RunSend for --tab/--tab-index.
	CurrentSessionID func() string
	ControllingTTY   func() string
	AncestorTTYs     func() []string

	// SendText writes into an iTerm session. Nil → iterm2.SendText.
	SendText func(sessionID, text string, opts iterm2.SendTextOptions, cfg *iterm2.SendTextConfig) error

	OpenInNewWindow func(dir, followUp string) error
	CodexBin         string
	LookPath        func(file string) (string, error)
	Env             []string

	// AgentRunOpen replaces production agent-run prefer+deliver (tests).
	// Return (nil, nil) soft miss; ambiguous error → warning + fall back.
	AgentRunOpen func(codexSessionID, prompt string) (*AgentRunOpenResult, error)

	// Sleep / Now inject the --open wait loop (tests). Nil → time.Sleep / time.Now.
	Sleep func(d time.Duration)
	Now   func() time.Time

	Stderr io.Writer // warnings (live-but-no-iTerm); nil → os.Stderr
}

// SendResult is the outcome of a successful Send resolve (and optional write).
type SendResult struct {
	SessionID         string
	ITermSessionID    string
	Text              string
	Opened            bool // true when --open resumed a new window
	ViaAgentRun       bool
	AgentRunSessionID string
	AgentRunMode      string // send | resume | run
	Candidate         FocusCandidate
	CWD               string
	Command           string
}

// SendFake is the deterministic injected boundary used by send tests.
type SendFake struct {
	FocusFake
	CurrentSessionID string
	ControllingTTY   string

	Opened       []string // "dir|followUp"
	SendCalls    []SendCall
	SendErr      error
	AfterOpen    func(*SendFake) // mutate Procs/ITerm after resume (tests)
	SleepCalls   int
	Clock        time.Time // advanced by Sleep when non-zero start set via InitClock
	clockStarted bool

	// AgentRunByID maps codex session id → prefer hit. Missing key = soft miss.
	AgentRunByID  map[string]*AgentRunOpenResult
	AgentRunErr   error
	AgentRunCalls []string
}

// SendCall records one SendText invocation.
type SendCall struct {
	SessionID string
	Text      string
	Opts      iterm2.SendTextOptions
}

// InitClock sets the fake clock start used by Sleep/Now injects.
func (f *SendFake) InitClock(t time.Time) {
	f.Clock = t
	f.clockStarted = true
}

// SendOpts returns SendOpts wired to this fake.
func (f *SendFake) SendOpts() *SendOpts {
	fo := f.FocusFake.Opts()
	opts := &SendOpts{
		ListProcs: fo.ListProcs,
		Lsof:      fo.Lsof,
		ListITerm: fo.ListITerm,
		CurrentSessionID: func() string {
			return f.CurrentSessionID
		},
		ControllingTTY: func() string {
			return f.ControllingTTY
		},
		AncestorTTYs: func() []string { return nil },
		SendText: func(sessionID, text string, opts iterm2.SendTextOptions, _ *iterm2.SendTextConfig) error {
			f.SendCalls = append(f.SendCalls, SendCall{SessionID: sessionID, Text: text, Opts: opts})
			if f.SendErr != nil {
				return f.SendErr
			}
			return nil
		},
		OpenInNewWindow: func(dir, followUp string) error {
			f.Opened = append(f.Opened, dir+"|"+followUp)
			if f.AfterOpen != nil {
				f.AfterOpen(f)
			}
			return nil
		},
		CodexBin:     "/usr/local/bin/codex",
		AgentRunBin: "/usr/local/bin/agent-run",
		LookPath: func(file string) (string, error) {
			switch file {
			case "agent-run":
				return "/usr/local/bin/agent-run", nil
			default:
				return "/usr/local/bin/codex", nil
			}
		},
		Env: []string{"PATH=/usr/bin"},
		Sleep: func(d time.Duration) {
			f.SleepCalls++
			if f.clockStarted {
				f.Clock = f.Clock.Add(d)
			}
		},
		Now: func() time.Time {
			if f.clockStarted {
				return f.Clock
			}
			return time.Now()
		},
	}
	if f.AgentRunByID != nil || f.AgentRunErr != nil {
		opts.AgentRunOpen = func(codexSessionID, prompt string) (*AgentRunOpenResult, error) {
			f.AgentRunCalls = append(f.AgentRunCalls, codexSessionID+"|"+prompt)
			if f.AgentRunErr != nil {
				return nil, f.AgentRunErr
			}
			hit, ok := f.AgentRunByID[codexSessionID]
			if !ok || hit == nil {
				return nil, nil
			}
			out := *hit
			if out.Mode == "" {
				if strings.TrimSpace(prompt) == "" {
					out.Mode = AgentRunOpenModeFocus
					out.Focused = true
				} else {
					out.Mode = AgentRunOpenModeSend
					out.Delivered = true
				}
			}
			return &out, nil
		}
	}
	return opts
}

// sendHostResolution is the resolved host / open path for Send.
type sendHostResolution struct {
	Candidate         FocusCandidate
	Opened            bool
	ViaAgentRun       bool
	AgentRunSessionID string
	AgentRunMode      string
	Delivered         bool // agent-run already delivered prompt; skip SendText
	CWD               string
	Command           string
}

// Send resolves a live iTerm host for sessionID and types text into it.
// When opts.Open is set and no host exists, resumes like Open then waits for a host.
// When opts.TabFrom is set, uses that tab (no rediscovery, no --open resume).
func Send(codexHome, sessionID, text string, opts *SendOpts) (*SendResult, error) {
	if opts == nil {
		opts = &SendOpts{}
	}
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return nil, fmt.Errorf("session id is required")
	}

	info, err := Info(codexHome, sessionID, 3)
	if err != nil {
		return nil, err
	}

	host, err := resolveSendHost(info, sessionID, text, opts)
	if err != nil {
		return nil, err
	}

	itermID := strings.TrimSpace(host.Candidate.SessionID)
	result := &SendResult{
		SessionID:         sessionID,
		ITermSessionID:    iterm2.SessionUUID(itermID),
		Text:              text,
		Opened:            host.Opened,
		ViaAgentRun:       host.ViaAgentRun,
		AgentRunSessionID: host.AgentRunSessionID,
		AgentRunMode:      host.AgentRunMode,
		Candidate:         host.Candidate,
		CWD:               host.CWD,
		Command:           host.Command,
	}

	if opts.DryRun || host.Delivered {
		return result, nil
	}

	if itermID == "" {
		return nil, fmt.Errorf("no hosting iTerm tab for session %s", sessionID)
	}

	sendFn := opts.SendText
	if sendFn == nil {
		sendFn = iterm2.SendText
	}
	if err := sendITermCodexText(sendFn, itermID, text, opts); err != nil {
		if errors.Is(err, iterm2.ErrSessionNotFound) || strings.Contains(strings.ToLower(err.Error()), "session not found") {
			return nil, fmt.Errorf("session not found: %s", iterm2.SessionUUID(itermID))
		}
		return nil, err
	}
	result.ITermSessionID = iterm2.SessionUUID(itermID)
	return result, nil
}

// sendITermCodexText types into the Codex composer via iTerm2 SendText.
//
// Real Codex TUI often treats a single AppleScript "write text …" (with newline)
// as typing only — the same failure mode agenttty.InjectMessage documents for
// codex-tty. Default submit therefore: type without newline, settle, Enter,
// settle, Enter again. --no-submit skips the Enter steps.
func sendITermCodexText(sendFn func(sessionID, text string, opts iterm2.SendTextOptions, cfg *iterm2.SendTextConfig) error, itermID, text string, opts *SendOpts) error {
	typeOnly := iterm2.SendTextOptions{
		Focus:    opts.Focus,
		NoSubmit: true,
		NoCtrlU:  opts.NoCtrlU,
	}
	if err := sendFn(itermID, text, typeOnly, nil); err != nil {
		return err
	}
	if opts.NoSubmit {
		return nil
	}
	sleepFn := opts.Sleep
	if sleepFn == nil {
		sleepFn = time.Sleep
	}
	enter := iterm2.SendTextOptions{Focus: false, NoSubmit: false, NoCtrlU: true}
	sleepFn(agenttty.CodexSubmitSettle)
	if err := sendFn(itermID, "", enter, nil); err != nil {
		return err
	}
	sleepFn(agenttty.CodexSubmitSettle)
	return sendFn(itermID, "", enter, nil)
}

func resolveSendHost(info *SessionInfo, sessionID, text string, opts *SendOpts) (sendHostResolution, error) {
	if opts.TabFrom != nil {
		return sendHostResolution{Candidate: focusCandidateFromTab(opts.TabFrom)}, nil
	}

	stderr := opts.Stderr
	if stderr == nil {
		stderr = os.Stderr
	}

	// Prefer agent-run BEFORE any iTerm discovery when --session-id is managed.
	arHit, arWarn, arErr := preferAgentRunOpen(agentRunOpenHooks{
		NoAgentRun:      opts.NoAgentRun,
		AgentRunHome:    opts.AgentRunHome,
		DryRun:          opts.DryRun,
		NoSubmit:        opts.NoSubmit,
		Dir:             opts.Dir,
		Info:            info,
		SkipProduction:  opts.ListProcs != nil || opts.ListITerm != nil,
		OpenInNewWindow: opts.OpenInNewWindow,
		AgentRunBin:     opts.AgentRunBin,
		LookPath:        opts.LookPath,
		Stderr:          stderr,
		AgentRunOpen:    opts.AgentRunOpen,
	}, sessionID, text)
	writeAgentRunOpenWarn(stderr, arWarn)
	if arErr != nil {
		// Managed deliver failed — never fall through to bare codex --resume.
		return sendHostResolution{}, arErr
	}
	if arHit != nil {
		cmd := arHit.Command
		if arHit.Opened && !opts.DryRun && opts.AgentRunOpen != nil {
			// Injectable prefer returns metadata only; launch here.
			if cmd == "" {
				cmd = "agent-run run --session-id " + arHit.AgentRunSessionID + " --auto-send-or-resume --open"
			}
			openFn := opts.OpenInNewWindow
			if openFn == nil {
				openFn = defaultOpenInNewWindow
			}
			if err := openFn(arHit.CWD, cmd); err != nil {
				return sendHostResolution{}, fmt.Errorf("agent-run deliver failed: open window: %w", err)
			}
		}
		return sendHostResolution{
			Opened:            arHit.Opened,
			ViaAgentRun:       true,
			AgentRunSessionID: arHit.AgentRunSessionID,
			AgentRunMode:      arHit.Mode,
			Delivered:         arHit.Delivered,
			CWD:               arHit.CWD,
			Command:           cmd,
		}, nil
	}

	focusOpts := &FocusOpts{
		Index:     opts.Index,
		ListProcs: opts.ListProcs,
		Lsof:      opts.Lsof,
		ListITerm: opts.ListITerm,
	}
	disc, err := DiscoverFocusHosting(sessionID, focusOpts)
	if err != nil {
		return sendHostResolution{}, err
	}
	if disc != nil && len(disc.Candidates) > 0 {
		cand, selErr := selectFocusCandidate(sessionID, disc.Candidates, opts.Index)
		if selErr != nil {
			msg := strings.ReplaceAll(selErr.Error(),
				"agent-pro codex session focus",
				"agent-pro codex session send")
			return sendHostResolution{}, fmt.Errorf("%s", msg)
		}
		return sendHostResolution{Candidate: cand}, nil
	}

	if !opts.Open {
		return sendHostResolution{}, fmt.Errorf("no hosting iTerm tab for session %s", sessionID)
	}

	cwd, err := resolveOpenCWD(info, opts.Dir)
	if err != nil {
		return sendHostResolution{}, err
	}
	bin, err := resolveCodexBin(opts.CodexBin, opts.LookPath)
	if err != nil {
		return sendHostResolution{}, err
	}
	argv := resumeCodexArgv(sessionID)
	cmdLine := quotedForkCommandLine(bin, argv)

	if disc != nil && disc.LiveCount > 0 {
		fmt.Fprintf(stderr, "warning: session has live codex PID(s) but no matching iTerm tab; opening a new window\n")
	}

	if opts.DryRun {
		return sendHostResolution{Opened: true, CWD: cwd, Command: cmdLine}, nil
	}

	openFn := opts.OpenInNewWindow
	if openFn == nil {
		openFn = defaultOpenInNewWindow
	}
	if err := openFn(cwd, cmdLine); err != nil {
		return sendHostResolution{}, fmt.Errorf("open new window: %w", err)
	}

	cand, err := waitForSendHostingTab(sessionID, opts)
	if err != nil {
		return sendHostResolution{Opened: true, CWD: cwd, Command: cmdLine}, err
	}
	return sendHostResolution{Candidate: cand, Opened: true, CWD: cwd, Command: cmdLine}, nil
}

func waitForSendHostingTab(sessionID string, opts *SendOpts) (FocusCandidate, error) {
	wait := opts.OpenWait
	if wait <= 0 {
		wait = DefaultSendOpenWait
	}
	nowFn := opts.Now
	if nowFn == nil {
		nowFn = time.Now
	}
	sleepFn := opts.Sleep
	if sleepFn == nil {
		sleepFn = time.Sleep
	}

	deadline := nowFn().Add(wait)
	focusOpts := &FocusOpts{
		Index:     opts.Index,
		ListProcs: opts.ListProcs,
		Lsof:      opts.Lsof,
		ListITerm: opts.ListITerm,
	}

	for {
		disc, err := DiscoverFocusHosting(sessionID, focusOpts)
		if err != nil {
			return FocusCandidate{}, err
		}
		if disc != nil && len(disc.Candidates) > 0 {
			cand, selErr := selectFocusCandidate(sessionID, disc.Candidates, opts.Index)
			if selErr != nil {
				msg := strings.ReplaceAll(selErr.Error(),
					"agent-pro codex session focus",
					"agent-pro codex session send")
				return FocusCandidate{}, fmt.Errorf("%s", msg)
			}
			return cand, nil
		}
		if !nowFn().Before(deadline) {
			return FocusCandidate{}, fmt.Errorf("timed out waiting for hosting tab after open (%s)", wait)
		}
		sleepFn(sendOpenPollInterval)
	}
}

// RunSend implements `agent-pro codex session send` with injectable writers/hooks.
func RunSend(args []string, stdout, stderr io.Writer, codexHome string, opts *SendOpts) error {
	if stdout == nil {
		stdout = io.Discard
	}
	if stderr == nil {
		stderr = io.Discard
	}
	parsed, err := parseSendArgs(args)
	if err != nil {
		return err
	}
	if parsed.Help {
		txt := SendHelp
		if !strings.HasSuffix(txt, "\n") {
			txt += "\n"
		}
		_, _ = io.WriteString(stdout, txt)
		return nil
	}
	payload, hasTextContent, hasEnter, err := composeSendPayload(parsed.Seq, parsed.Text)
	if err != nil {
		return err
	}
	if payload == "" {
		return fmt.Errorf("send: missing text or key (--enter/--up/--text/…)")
	}
	if parsed.Open && (parsed.Tab != nil || parsed.TabIndex != nil) {
		return fmt.Errorf("--open cannot be combined with --tab/--tab-index")
	}

	runOpts := SendOpts{}
	if opts != nil {
		runOpts = *opts
	}
	runOpts.Index = parsed.Index
	runOpts.Dir = parsed.Dir
	runOpts.DryRun = parsed.DryRun
	runOpts.Open = parsed.Open
	runOpts.NoAgentRun = parsed.NoAgentRun
	runOpts.Focus = parsed.Focus
	runOpts.NoSubmit = parsed.NoSubmit
	runOpts.NoCtrlU = parsed.NoCtrlU
	if !hasTextContent {
		runOpts.NoCtrlU = true
		runOpts.NoSubmit = true
	}
	if hasEnter {
		runOpts.NoSubmit = true
	}
	if runOpts.Stderr == nil {
		runOpts.Stderr = stderr
	}

	sessionID, tabMeta, err := resolveSendSessionSource(parsed.SessionID, parsed.Tab, parsed.TabIndex, &SessionSourceOpts{
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

	result, err := Send(codexHome, sessionID, payload, &runOpts)
	if err != nil {
		return err
	}

	if parsed.DryRun {
		if result.ViaAgentRun && result.Opened {
			fmt.Fprintln(stdout, "Would open new iTerm2 window via agent-run")
			fmt.Fprintf(stdout, "  codex id:       %s\n", sessionID)
			fmt.Fprintf(stdout, "  agent-run id:  %s\n", result.AgentRunSessionID)
			fmt.Fprintf(stdout, "  mode:          %s\n", result.AgentRunMode)
			fmt.Fprintf(stdout, "  cwd:           %s\n", result.CWD)
			fmt.Fprintf(stdout, "  command:       %s\n", result.Command)
		} else if result.ViaAgentRun {
			fmt.Fprintln(stdout, "Would send via agent-run")
			fmt.Fprintf(stdout, "  codex id:       %s\n", sessionID)
			fmt.Fprintf(stdout, "  agent-run id:  %s\n", result.AgentRunSessionID)
			fmt.Fprintf(stdout, "  mode:          %s\n", result.AgentRunMode)
		} else if result.Opened {
			fmt.Fprintln(stdout, "Would open new iTerm2 window")
			fmt.Fprintf(stdout, "  codex id:  %s\n", sessionID)
			fmt.Fprintf(stdout, "  cwd:      %s\n", result.CWD)
			fmt.Fprintf(stdout, "  command:  %s\n", result.Command)
		} else {
			fmt.Fprintf(stdout, "Would send to: window %s, tab %d\n", result.Candidate.WindowID, result.Candidate.TabIndex)
			fmt.Fprintf(stdout, "  codex id:   %s\n", result.SessionID)
			fmt.Fprintf(stdout, "  iterm id:  %s\n", result.ITermSessionID)
		}
		fmt.Fprintf(stdout, "  text:      %q\n", result.Text)
		return nil
	}

	if result.Opened {
		if result.ViaAgentRun {
			fmt.Fprintf(stdout, "opened: new window; agent-run resume %s\n", result.AgentRunSessionID)
		} else {
			fmt.Fprintf(stdout, "opened: new window; resuming %s\n", sessionID)
		}
	}
	fmt.Fprintf(stdout, "sent to session %s\n", sessionID)
	return nil
}

func resolveSendSessionSource(sessionIDFlag *string, tabFlag *string, tabIndexFlag *int, opts *SessionSourceOpts) (string, *TabResolveResult, error) {
	if sessionIDFlag != nil {
		if tabFlag != nil || tabIndexFlag != nil {
			return "", nil, fmt.Errorf("--session-id cannot be combined with --tab/--tab-index")
		}
		id := strings.TrimSpace(*sessionIDFlag)
		if id == "" {
			return "", nil, fmt.Errorf("session id is required")
		}
		return id, nil, nil
	}
	if tabFlag == nil && tabIndexFlag == nil {
		return "", nil, fmt.Errorf("expected --session-id, or --tab / --tab-index")
	}
	return ResolveSessionSource(nil, tabFlag, tabIndexFlag, opts)
}

type sendArgs struct {
	Text       string // positional text; always appended last after Seq
	Seq        lessflags.Flags
	SessionID  *string
	Index      *int
	Tab        *string
	TabIndex   *int
	Dir        string
	DryRun     bool
	Open       bool
	NoAgentRun bool
	Focus      bool
	NoSubmit   bool
	NoCtrlU    bool
	Help       bool
}

func parseSendArgs(args []string) (sendArgs, error) {
	var out sendArgs
	remain, err := lessflags.
		String("--session-id", &out.SessionID).
		String("--tab", &out.Tab).
		Int("--tab-index", &out.TabIndex).
		Int("--index", &out.Index).
		String("--dir", &out.Dir).
		Bool("--open", &out.Open).
		Bool("--no-agent-run", &out.NoAgentRun).
		Bool("--focus", &out.Focus).
		Bool("--no-submit", &out.NoSubmit).
		Bool("--no-ctrl-u", &out.NoCtrlU).
		Bool("--dry-run", &out.DryRun).
		Group(
			lessflags.CollectParsedFlags(&out.Seq).
				Bool("--enter").
				Bool("--up,--down,--left,--right").
				Bool("--esc").
				Bool("--ctrl-c,--ctrl-d").
				String("--text"),
		).
		HelpFunc("-h,--help", func() {}).
		HelpNoExit().
		Parse(args)
	if err != nil {
		if err == lessflags.ErrHelp {
			out.Help = true
			return out, nil
		}
		return out, err
	}
	switch len(remain) {
	case 0:
		// missing payload checked in RunSend after help
	case 1:
		out.Text = remain[0]
	default:
		return out, fmt.Errorf("send: unexpected arguments: %s", strings.Join(remain[1:], " "))
	}
	return out, nil
}

// composeSendPayload builds the write payload: Seq in CLI order, then positional.
// hasTextContent is true when positional or --text contributed; hasEnter when --enter appears.
func composeSendPayload(seq lessflags.Flags, positional string) (payload string, hasTextContent bool, hasEnter bool, err error) {
	var b strings.Builder
	for _, fl := range seq {
		switch fl.Flag {
		case "--enter":
			b.WriteByte('\n')
			hasEnter = true
		case "--up":
			b.WriteString("\x1b[A")
		case "--down":
			b.WriteString("\x1b[B")
		case "--right":
			b.WriteString("\x1b[C")
		case "--left":
			b.WriteString("\x1b[D")
		case "--esc":
			b.WriteByte(0x1b)
		case "--ctrl-c":
			b.WriteByte(0x03)
		case "--ctrl-d":
			b.WriteByte(0x04)
		case "--text":
			b.WriteString(fl.Value)
			hasTextContent = true
		default:
			return "", false, false, fmt.Errorf("unsupported send sequence flag: %s", fl.Flag)
		}
	}
	if positional != "" {
		b.WriteString(positional)
		hasTextContent = true
	}
	return b.String(), hasTextContent, hasEnter, nil
}
