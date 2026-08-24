package sessions

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
)

// SnapshotHelp is the text for `agent-pro grok session snapshot --help`.
const SnapshotHelp = `Usage: agent-pro grok session snapshot (<session-id> | --tab SEL | --tab-index N) [OPTIONS]

Print currently visible pane text for a live Grok session host.
Does not focus the pane or switch Desktop. No resume when no host.

When the Grok id is bound to a live agent-run grok-tty session, prefers that
TTY snapshot (sanitized single frame). Otherwise uses iTerm2 Contents
(same as: kool iterm2 contents <iterm-uuid>). Bare grok (not under agent-run)
always uses iTerm.

Session source (exactly one):
  <session-id>          explicit Grok session id
  --tab SEL             1-based tab index, or next|left|right (right ≡ next)
  --tab-index N         0-based tab index in this iTerm window

Options:
  --index N             select candidate N when multiple tabs host the same session
                        (positional <session-id> only; not with --tab/--tab-index)
  --json                emit {"session_id","iterm_session_id","app","source","contents"} (no ANSI)
  -o, --output FILE     write output to FILE instead of stdout
  --dry-run             resolve only; do not capture pane text
  --iterm               force iTerm Contents (skip agent-run prefer path)
  -h,--help             show help

Tab selectors use the same window/tab discovery as: kool iterm2 window status.
Relative next/left/right do not wrap; edges error.
`

// SnapshotCommandHelpLine is the parent `agent-pro grok session` help row.
const SnapshotCommandHelpLine = `  snapshot …             capture visible pane text (--tab / --tab-index / <id>)`

// SnapshotOpts drives Snapshot / RunSnapshot. Nil hooks use production probes.
type SnapshotOpts struct {
	Index  *int
	DryRun bool
	JSON   bool
	Output string // -o/--output path; empty → stdout

	// ForceITerm skips the agent-run prefer path (--iterm).
	ForceITerm bool

	// AgentRunHome overrides AGENT_RUN_HOME / ~/.agent-run for production prefer.
	AgentRunHome string

	// TabFrom, when set, short-circuits host discovery to this already-resolved tab.
	TabFrom *TabResolveResult

	ListProcs func() []FocusProc
	Lsof      func(int) []string
	ListITerm func() ([]iterm2.SessionRef, error)

	// Tab resolve hooks (nil → production). Used by RunSnapshot for --tab/--tab-index.
	CurrentSessionID func() string
	ControllingTTY   func() string
	AncestorTTYs     func() []string

	// Contents dumps visible pane text. Nil → iterm2.Contents.
	Contents func(sessionID string, cfg *iterm2.ContentsConfig) (iterm2.ContentsResult, error)

	// AgentRunSnapshot, when set, is tried before iTerm (unless ForceITerm).
	// Return (nil, nil) for soft miss; error containing "ambiguous" → warning + fall back.
	AgentRunSnapshot func(grokSessionID string) (*AgentRunSnapshotResult, error)

	// Warning, when set, receives soft-fallback messages (e.g. ambiguous mapping).
	Warning func(string)
}

// SnapshotResult is the outcome of a successful Snapshot resolve (and optional capture).
type SnapshotResult struct {
	SessionID         string // Grok session id
	ITermSessionID    string
	App               string
	Contents          string
	Source            string // SnapshotSourceAgentRun | SnapshotSourceITerm
	AgentRunSessionID string
	Candidate         FocusCandidate
}

type snapshotJSON struct {
	SessionID         string `json:"session_id"`
	ITermSessionID    string `json:"iterm_session_id"`
	App               string `json:"app"`
	Source            string `json:"source"`
	AgentRunSessionID string `json:"agent_run_session_id,omitempty"`
	Contents          string `json:"contents"`
}

// SnapshotFake is the deterministic injected boundary used by snapshot tests.
type SnapshotFake struct {
	FocusFake
	CurrentSessionID string
	ControllingTTY   string

	ContentsByID  map[string]iterm2.ContentsResult // keyed by SessionUUID(id) and raw id
	ContentsErr   error                            // returned when no map hit (and no default)
	ContentsCalls []string

	// AgentRunByID maps grok session id → prefer-path hit. Missing key = soft miss.
	AgentRunByID    map[string]*AgentRunSnapshotResult
	AgentRunErr     error // returned for every AgentRunSnapshot call when set
	AgentRunCalls   []string
	AgentRunEnabled bool // when true, wire AgentRunSnapshot even if AgentRunByID is nil
}

// SnapshotOpts returns SnapshotOpts wired to this fake.
func (f *SnapshotFake) SnapshotOpts() *SnapshotOpts {
	fo := f.FocusFake.Opts()
	opts := &SnapshotOpts{
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
		Contents: func(sessionID string, _ *iterm2.ContentsConfig) (iterm2.ContentsResult, error) {
			f.ContentsCalls = append(f.ContentsCalls, sessionID)
			if f.ContentsErr != nil {
				return iterm2.ContentsResult{}, f.ContentsErr
			}
			uuid := iterm2.SessionUUID(sessionID)
			if f.ContentsByID != nil {
				if res, ok := f.ContentsByID[sessionID]; ok {
					return res, nil
				}
				if res, ok := f.ContentsByID[uuid]; ok {
					return res, nil
				}
			}
			return iterm2.ContentsResult{
				SessionID: uuid,
				App:       iterm2.CanonicalITermAppSystem,
				Contents:  "fixture pane text for " + uuid,
			}, nil
		},
	}
	if f.AgentRunEnabled || f.AgentRunByID != nil || f.AgentRunErr != nil {
		opts.AgentRunSnapshot = func(grokSessionID string) (*AgentRunSnapshotResult, error) {
			f.AgentRunCalls = append(f.AgentRunCalls, grokSessionID)
			if f.AgentRunErr != nil {
				return nil, f.AgentRunErr
			}
			if f.AgentRunByID == nil {
				return nil, nil
			}
			hit, ok := f.AgentRunByID[grokSessionID]
			if !ok {
				return nil, nil
			}
			return hit, nil
		}
	}
	return opts
}

// Snapshot resolves a live iTerm host for sessionID and captures visible pane text.
// When opts.TabFrom is set, uses that tab's iTerm session (no rediscovery).
// Unlike Open, missing hosts are a hard error (no resume).
func Snapshot(grokHome, sessionID string, opts *SnapshotOpts) (*SnapshotResult, error) {
	if opts == nil {
		opts = &SnapshotOpts{}
	}
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return nil, fmt.Errorf("session id is required")
	}

	if _, err := Find(grokHome, sessionID); err != nil {
		return nil, err
	}

	if hit, warn, ok := preferAgentRunSnapshot(opts, sessionID); ok {
		return &SnapshotResult{
			SessionID:         sessionID,
			Source:            SnapshotSourceAgentRun,
			AgentRunSessionID: hit.AgentRunSessionID,
			Contents:          hit.Contents,
		}, nil
	} else if warn != "" && opts.Warning != nil {
		opts.Warning(warn)
	}

	// Production capture: one AppleScript (TTY → meta+contents) when --index
	// is not needed. Tries each live TTY until one pane hits (ancestor TTYs
	// may not map to an iTerm session). Skips full ListITerm + Contents walk.
	if opts.TabFrom == nil && opts.ListITerm == nil && opts.Contents == nil &&
		!opts.DryRun && opts.Index == nil {
		if ttys, ok := liveTTYsForSession(sessionID, opts.ListProcs, opts.Lsof); ok {
			for _, tty := range ttys {
				cap, capErr := iterm2.CaptureByTTY(tty, nil)
				if capErr != nil {
					continue
				}
				ref := cap.Ref
				return &SnapshotResult{
					SessionID:      sessionID,
					ITermSessionID: iterm2.SessionUUID(ref.SessionID),
					App:            cap.App,
					Contents:       cap.Contents,
					Source:         SnapshotSourceITerm,
					Candidate: FocusCandidate{
						WindowID:    ref.WindowID,
						WindowTitle: ref.WindowName,
						TabIndex:    ref.TabIndex,
						SessionID:   ref.SessionID,
						TTY:         iterm2.NormalizeTTY(ref.TTY),
					},
				}, nil
			}
			// Fall through to DiscoverFocusHosting on soft miss.
		}
	}

	var selected FocusCandidate
	if opts.TabFrom != nil {
		selected = focusCandidateFromTab(opts.TabFrom)
	} else {
		focusOpts := &FocusOpts{
			Index:     opts.Index,
			ListProcs: opts.ListProcs,
			Lsof:      opts.Lsof,
			ListITerm: opts.ListITerm,
		}
		disc, err := DiscoverFocusHosting(sessionID, focusOpts)
		if err != nil {
			return nil, err
		}
		if disc == nil || len(disc.Candidates) == 0 {
			return nil, fmt.Errorf("no hosting iTerm tab for session %s", sessionID)
		}
		cand, selErr := selectFocusCandidate(sessionID, disc.Candidates, opts.Index)
		if selErr != nil {
			msg := strings.ReplaceAll(selErr.Error(),
				"agent-pro grok session focus",
				"agent-pro grok session snapshot")
			return nil, fmt.Errorf("%s", msg)
		}
		selected = cand
	}

	itermID := strings.TrimSpace(selected.SessionID)
	if itermID == "" && strings.TrimSpace(selected.TTY) == "" {
		return nil, fmt.Errorf("no hosting iTerm tab for session %s", sessionID)
	}

	result := &SnapshotResult{
		SessionID:      sessionID,
		ITermSessionID: iterm2.SessionUUID(itermID),
		Source:         SnapshotSourceITerm,
		Candidate:      selected,
	}

	if opts.DryRun {
		return result, nil
	}

	// Production: TTY early-exit Contents (avoids a second full UUID walk).
	// Injected Contents keeps the UUID Contents path for tests.
	if opts.Contents != nil {
		if itermID == "" {
			return nil, fmt.Errorf("no hosting iTerm tab for session %s", sessionID)
		}
		res, err := opts.Contents(itermID, nil)
		if err != nil {
			if errors.Is(err, iterm2.ErrSessionNotFound) || strings.Contains(strings.ToLower(err.Error()), "session not found") {
				return nil, fmt.Errorf("session not found: %s", iterm2.SessionUUID(itermID))
			}
			return nil, err
		}
		result.ITermSessionID = res.SessionID
		if strings.TrimSpace(result.ITermSessionID) == "" {
			result.ITermSessionID = iterm2.SessionUUID(itermID)
		}
		result.App = res.App
		result.Contents = res.Contents
		return result, nil
	}

	tty := strings.TrimSpace(selected.TTY)
	if tty != "" {
		res, err := iterm2.ContentsByTTY(tty, nil)
		if err != nil {
			if errors.Is(err, iterm2.ErrSessionNotFound) || strings.Contains(strings.ToLower(err.Error()), "session not found") ||
				strings.Contains(strings.ToLower(err.Error()), "tty not found") {
				return nil, fmt.Errorf("no hosting iTerm tab for session %s", sessionID)
			}
			return nil, err
		}
		if strings.TrimSpace(result.ITermSessionID) == "" {
			result.ITermSessionID = iterm2.SessionUUID(itermID)
		}
		result.App = res.App
		result.Contents = res.Contents
		return result, nil
	}

	res, err := iterm2.Contents(itermID, nil)
	if err != nil {
		if errors.Is(err, iterm2.ErrSessionNotFound) || strings.Contains(strings.ToLower(err.Error()), "session not found") {
			return nil, fmt.Errorf("session not found: %s", iterm2.SessionUUID(itermID))
		}
		return nil, err
	}
	result.ITermSessionID = res.SessionID
	if strings.TrimSpace(result.ITermSessionID) == "" {
		result.ITermSessionID = iterm2.SessionUUID(itermID)
	}
	result.App = res.App
	result.Contents = res.Contents
	return result, nil
}

// RunSnapshot implements `agent-pro grok session snapshot` with injectable writers/hooks.
func RunSnapshot(args []string, stdout, stderr io.Writer, grokHome string, opts *SnapshotOpts) error {
	if stdout == nil {
		stdout = io.Discard
	}
	if stderr == nil {
		stderr = io.Discard
	}
	parsed, err := parseSnapshotArgs(args)
	if err != nil {
		return err
	}
	if parsed.Help {
		txt := SnapshotHelp
		if !strings.HasSuffix(txt, "\n") {
			txt += "\n"
		}
		_, _ = io.WriteString(stdout, txt)
		return nil
	}

	runOpts := SnapshotOpts{}
	if opts != nil {
		runOpts = *opts
	}
	runOpts.Index = parsed.Index
	runOpts.DryRun = parsed.DryRun
	runOpts.JSON = parsed.JSON
	runOpts.Output = parsed.Output
	runOpts.ForceITerm = parsed.ForceITerm
	if runOpts.Warning == nil {
		runOpts.Warning = func(msg string) {
			fmt.Fprintln(stderr, msg)
		}
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

	result, err := Snapshot(grokHome, sessionID, &runOpts)
	if err != nil {
		return err
	}

	if parsed.DryRun {
		if result.Source == SnapshotSourceAgentRun {
			fmt.Fprintf(stdout, "Would capture via agent-run\n")
			fmt.Fprintf(stdout, "  grok id:        %s\n", result.SessionID)
			fmt.Fprintf(stdout, "  agent-run id:   %s\n", result.AgentRunSessionID)
			fmt.Fprintf(stdout, "  source:         %s\n", result.Source)
			return nil
		}
		fmt.Fprintf(stdout, "Would capture: window %s, tab %d\n", result.Candidate.WindowID, result.Candidate.TabIndex)
		fmt.Fprintf(stdout, "  grok id:   %s\n", result.SessionID)
		fmt.Fprintf(stdout, "  iterm id:  %s\n", result.ITermSessionID)
		fmt.Fprintf(stdout, "  source:    %s\n", result.Source)
		return nil
	}

	var body []byte
	if parsed.JSON {
		var buf bytes.Buffer
		enc := json.NewEncoder(&buf)
		enc.SetEscapeHTML(false)
		if err := enc.Encode(snapshotJSON{
			SessionID:         result.SessionID,
			ITermSessionID:    result.ITermSessionID,
			App:               result.App,
			Source:            result.Source,
			AgentRunSessionID: result.AgentRunSessionID,
			Contents:          result.Contents,
		}); err != nil {
			return err
		}
		body = buf.Bytes()
	} else {
		body = []byte(result.Contents)
		if len(body) > 0 && body[len(body)-1] != '\n' {
			body = append(body, '\n')
		}
	}

	if strings.TrimSpace(parsed.Output) != "" {
		if err := os.WriteFile(parsed.Output, body, 0o644); err != nil {
			return err
		}
		return nil
	}
	_, err = stdout.Write(body)
	return err
}

type snapshotArgs struct {
	Positional []string
	Index      *int
	Tab        *string
	TabIndex   *int
	JSON       bool
	Output     string
	DryRun     bool
	ForceITerm bool
	Help       bool
}

func parseSnapshotArgs(args []string) (snapshotArgs, error) {
	var out snapshotArgs
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
		if arg == "--iterm" {
			out.ForceITerm = true
			continue
		}
		if arg == "--json" {
			out.JSON = true
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
		if arg == "-o" || arg == "--output" {
			if i+1 >= len(args) {
				return out, fmt.Errorf("%s requires a path", arg)
			}
			out.Output = args[i+1]
			i++
			continue
		}
		if strings.HasPrefix(arg, "--output=") {
			out.Output = strings.TrimPrefix(arg, "--output=")
			continue
		}
		if strings.HasPrefix(arg, "-o=") {
			out.Output = strings.TrimPrefix(arg, "-o=")
			continue
		}
		if strings.HasPrefix(arg, "-") {
			return out, fmt.Errorf("unrecognized flag: %s", arg)
		}
		out.Positional = append(out.Positional, arg)
	}
	return out, nil
}
