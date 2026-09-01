package sessions

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/xhd2015/dot-pkgs/go-pkgs/computer-use/macos/space"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2/tabselect"
)

// PromptsHelp is the text for `agent-pro grok session prompts --help`.
const PromptsHelp = `Usage:
  agent-pro grok session prompts (<session-id> | --session-id ID | --tab SEL | --tab-index N | --this-tab)
    [--first] [--grep P]... [--exclude Q] [--head N | --tail N] [--max-body N]
    [--color|--no-color]
  agent-pro grok session prompts [--this-window | --this-space]
    [--first] [--recent <window>] [--limit N]
    [--grep P]... [--exclude Q] [--head N | --tail N] [--max-body N]
    [--color|--no-color]
  agent-pro grok session prompts [--recent <window>] [--limit N]
    [--first] [--grep P]... [--exclude Q] [--head N | --tail N] [--max-body N]
    [--color|--no-color]
  agent-pro grok sessions prompts …   (alias)

Show user prompts only as compact lines:
  [YYYY-MM-DD HH:MM:SS] prompt text…

Single mode: all user prompts for one session (full history), optional text filters.
Multi mode (no session source): newest sessions by last_active, with selection matrix:

  (no flags)              last 10 sessions that have prompts
  --limit N               last N sessions that have prompts (N >= 1)
  --recent Nd|Nh|Nm       all sessions with ≥1 in-window user prompt (no default cap)
  --recent W --limit N    in-window sessions only, stop at N
  --this-window           live hosts in this iTerm window (no default cap)
  --this-space            live hosts on this macOS Mission Control desktop (no default cap)
  --main                  only main-agent class sessions (skip subagents)

Filter pipeline (per session, after recent window): grep keep → exclude drop → head|tail.
Sessions with zero survivors are skipped and do not count toward --limit.
Head and tail are mutually exclusive; N >= 1. Empty --grep/--exclude patterns error.
--first is sugar for --head 1 (mutually exclusive with --head/--tail).

Body length: full collapsed text by default. --max-body N soft-caps each body to
N runes + … (N >= 1). With --grep, full body + highlight unless --max-body, which
windows around the first pattern's first match within N runes.

Multi layout: session header, prompt lines, separator rule between sessions, footer.
Output streams session-by-session (not buffered until the end).

Session source (exactly one when scoping a session or host set):
  <session-id>          explicit Grok session id
  --session-id ID       same as positional id
  --tab SEL             1-based tab index, or next|left|right|current
  --tab-index N         0-based tab index in this iTerm window
  --this-tab            alias for --tab current
  --this-window         all live Grok hosts in this iTerm window
  --this-space          all live Grok hosts on this macOS Space

Options:
  --recent WINDOW   time window: Nd, Nh, or Nm (e.g. 1d, 2h, 30m)
  --limit N         session limit (see matrix above; must be >= 1)
  --first           only the first user prompt per session after text filters
  --main            only main-agent class sessions (alias: --main-agent)
  --grep P          keep prompts whose text contains P (repeatable; AND on the
                    same prompt; case-insensitive literal)
  --exclude Q       drop prompts whose text matches Q (case-insensitive literal)
  --head N          first N prompts per session after text filters (N >= 1)
  --tail N          last N prompts per session after text filters (N >= 1)
  --max-body N      soft-cap each prompt body to N runes + … (N >= 1; default: full)
  --color           force ANSI color on (even when stdout is not a TTY)
  --no-color        force ANSI color off
  -h,--help         show help

Color (auto by default): TTY on unless NO_COLOR is set; --color/--no-color override.
With --grep and color on, match spans are bold-red; omission markers are dim.
Headers, timestamps, separators, and footer are dim when color is on.

Notes:
  - session id / --tab / --this-tab cannot be combined with --recent, --limit,
    --this-window, or --this-space
  - sessions with zero user prompts (or zero in-window prompts) are skipped
`

// PromptsCommandHelpLine is the parent `agent-pro grok session` help row.
const PromptsCommandHelpLine = `  prompts [session-id]  list user prompts (one session or recent multi)`

// PromptsOpts injects probes for RunPrompts. Nil hooks use production.
type PromptsOpts struct {
	ListLive         func() ([]LiveHostingRow, error)
	CurrentWindow    func() (iterm2.WindowStatus, error)
	ListUserSpaces   func() ([]space.UserSpace, error)
	SpaceIndexForWin func(windowID uint64) (int, error)
	HomeDir          func() string
	Now              func() time.Time
	// Tab resolve hooks (nil → production). Used for --tab/--this-tab.
	ListProcs        func() []FocusProc
	Lsof             func(int) []string
	ListITerm        func() ([]iterm2.SessionRef, error)
	CurrentSessionID func() string
	ControllingTTY   func() string
	AncestorTTYs     func() []string
	SessionMeta      func(sessionID string) (TabSessionMeta, bool)
}

// RunPrompts implements `agent-pro grok session prompts`.
func RunPrompts(args []string, stdout, stderr io.Writer, grokHome string, opts *PromptsOpts) error {
	if stdout == nil {
		stdout = io.Discard
	}
	if stderr == nil {
		stderr = io.Discard
	}
	parsed, err := parsePromptsArgs(args)
	if err != nil {
		return err
	}
	if parsed.Help {
		txt := PromptsHelp
		if !strings.HasSuffix(txt, "\n") {
			txt += "\n"
		}
		_, err = io.WriteString(stdout, txt)
		return err
	}

	runOpts := PromptsOpts{}
	if opts != nil {
		runOpts = *opts
	}
	nowFn := runOpts.Now
	if nowFn == nil {
		nowFn = time.Now
	}
	homeFn := runOpts.HomeDir
	if homeFn == nil {
		homeFn = func() string {
			h, _ := os.UserHomeDir()
			return h
		}
	}
	now := nowFn()
	home := homeFn()

	filterOpts := FilterUserPromptsOptions{
		Grep:       parsed.Grep,
		GrepSet:    parsed.GrepSet,
		Exclude:    parsed.Exclude,
		ExcludeSet: parsed.ExcludeSet,
		Head:       parsed.Head,
		HeadSet:    parsed.HeadSet,
		Tail:       parsed.Tail,
		TailSet:    parsed.TailSet,
	}
	fmtOpts := FormatPromptsOptions{
		Now:          now,
		Home:         home,
		Window:       parsed.Recent,
		Limit:        parsed.Limit,
		RecentSet:    parsed.RecentSet,
		LimitSet:     parsed.LimitSet,
		ColorMode:    parsed.ColorMode,
		Grep:         parsed.Grep,
		GrepSet:      parsed.GrepSet,
		MaxBodyRunes: parsed.MaxBody,
		MaxBodySet:   parsed.MaxBodySet,
	}

	singleID, err := resolvePromptsSingleSession(parsed, grokHome, &runOpts)
	if err != nil {
		return err
	}
	if singleID != "" {
		if parsed.RecentSet || parsed.LimitSet || parsed.ThisWindow || parsed.ThisSpace {
			return fmt.Errorf("session id cannot be combined with --recent, --limit, --this-window, or --this-space")
		}
		sp, err := Prompts(grokHome, singleID)
		if err != nil {
			return err
		}
		if parsed.MainOnly && isSubAgentClass(sp.Session) {
			return fmt.Errorf("session %s is a subagent; omit --main or choose a main-agent session", singleID)
		}
		hasFilter := parsed.GrepSet || parsed.ExcludeSet || parsed.HeadSet || parsed.TailSet
		if hasFilter {
			kept, ob, oa, err := FilterUserPrompts(sp.UserPrompts, filterOpts)
			if err != nil {
				return err
			}
			sp.UserPrompts = kept
			sp.OmittedBefore = ob
			sp.OmittedAfter = oa
		}
		bw := bufio.NewWriter(stdout)
		defer bw.Flush()
		return WritePromptsText(bw, sp, fmtOpts)
	}

	listOpts := ListPromptsOptions{
		Now:        now,
		Recent:     parsed.Recent,
		RecentSet:  parsed.RecentSet,
		Limit:      parsed.Limit,
		LimitSet:   parsed.LimitSet,
		Home:       home,
		MainOnly:   parsed.MainOnly,
		Grep:       parsed.Grep,
		GrepSet:    parsed.GrepSet,
		Exclude:    parsed.Exclude,
		ExcludeSet: parsed.ExcludeSet,
		Head:       parsed.Head,
		HeadSet:    parsed.HeadSet,
		Tail:       parsed.Tail,
		TailSet:    parsed.TailSet,
	}

	if parsed.ThisWindow || parsed.ThisSpace {
		ids, err := resolvePromptsHostScopeIDs(parsed, stderr, &runOpts)
		if err != nil {
			return err
		}
		listOpts.OnlySessionIDs = ids
		listOpts.OnlySessionIDsSet = true
	}

	bw := bufio.NewWriterSize(stdout, 1024)
	defer bw.Flush()
	fmt.Fprintln(stderr, "scanning sessions…")
	return StreamPromptsList(bw, grokHome, listOpts, fmtOpts)
}

func resolvePromptsSingleSession(parsed promptsArgs, grokHome string, opts *PromptsOpts) (string, error) {
	sources := 0
	if len(parsed.Positional) > 0 {
		sources++
	}
	if parsed.SessionID != nil {
		sources++
	}
	if parsed.Tab != nil || parsed.TabIndex != nil || parsed.ThisTab {
		sources++
	}
	if parsed.ThisWindow {
		sources++
	}
	if parsed.ThisSpace {
		sources++
	}
	if sources > 1 {
		return "", fmt.Errorf("exactly one session source: <session-id>|--session-id|--tab|--tab-index|--this-tab|--this-window|--this-space")
	}
	if len(parsed.Positional) > 1 {
		return "", fmt.Errorf("expected at most one session id, got %d arguments", len(parsed.Positional))
	}
	if len(parsed.Positional) == 1 {
		id := strings.TrimSpace(parsed.Positional[0])
		if id == "" {
			return "", fmt.Errorf("session id is required")
		}
		return id, nil
	}
	if parsed.SessionID != nil {
		id := strings.TrimSpace(*parsed.SessionID)
		if id == "" {
			return "", fmt.Errorf("session id is required")
		}
		return id, nil
	}
	if parsed.ThisTab || parsed.Tab != nil || parsed.TabIndex != nil {
		var tabFlag *string
		var tabIndex *int
		if parsed.ThisTab {
			cur := "current"
			tabFlag = &cur
		} else {
			tabFlag = parsed.Tab
			tabIndex = parsed.TabIndex
		}
		id, _, err := ResolveSessionSource(nil, tabFlag, tabIndex, &SessionSourceOpts{
			ListProcs:        opts.ListProcs,
			Lsof:             opts.Lsof,
			ListITerm:        opts.ListITerm,
			CurrentSessionID: opts.CurrentSessionID,
			ControllingTTY:   opts.ControllingTTY,
			AncestorTTYs:     opts.AncestorTTYs,
			GrokHome:         grokHome,
			SessionMeta:      opts.SessionMeta,
		})
		return id, err
	}
	return "", nil
}

func resolvePromptsHostScopeIDs(parsed promptsArgs, stderr io.Writer, opts *PromptsOpts) ([]string, error) {
	listLive := opts.ListLive
	if listLive == nil {
		listLive = func() ([]LiveHostingRow, error) {
			return ListLive(nil)
		}
	}
	rows, err := listLive()
	if err != nil {
		return nil, err
	}

	currentWindow := opts.CurrentWindow
	if currentWindow == nil {
		currentWindow = func() (iterm2.WindowStatus, error) {
			return iterm2.CurrentWindowStatus()
		}
	}
	listSpaces := opts.ListUserSpaces
	if listSpaces == nil {
		listSpaces = func() ([]space.UserSpace, error) {
			return space.ListUserSpaces()
		}
	}
	spaceIdx := opts.SpaceIndexForWin
	if spaceIdx == nil {
		spaceIdx = func(windowID uint64) (int, error) {
			return space.SpaceIndexForWindow(windowID)
		}
	}

	wantWindow := ""
	wantSpace := -1
	if parsed.ThisWindow {
		st, err := currentWindow()
		if err != nil {
			return nil, err
		}
		wantWindow = strings.TrimSpace(st.WindowID)
		if wantWindow == "" {
			return nil, fmt.Errorf("current iTerm window id unknown")
		}
	}
	if parsed.ThisSpace {
		spaces, err := listSpaces()
		if err != nil {
			return nil, err
		}
		found := false
		for _, s := range spaces {
			if s.Current {
				wantSpace = s.Index
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("current macOS Space unknown")
		}
	}

	seen := map[string]bool{}
	var ids []string
	spaceCache := map[string]int{}
	spaceErr := map[string]bool{}
	for _, row := range rows {
		sid := strings.TrimSpace(row.SessionID)
		if sid == "" || seen[sid] {
			continue
		}
		winIDs := liveRowWindowIDs(row)
		keep := false
		if parsed.ThisWindow {
			for _, wid := range winIDs {
				if wid == wantWindow {
					keep = true
					break
				}
			}
		} else if parsed.ThisSpace {
			for _, wid := range winIDs {
				idx, ok := spaceCache[wid]
				if !ok {
					if spaceErr[wid] {
						continue
					}
					u, convErr := strconv.ParseUint(wid, 10, 64)
					if convErr != nil {
						fmt.Fprintf(stderr, "warning: skip window %s: invalid id\n", wid)
						spaceErr[wid] = true
						continue
					}
					var serr error
					idx, serr = spaceIdx(u)
					if serr != nil {
						fmt.Fprintf(stderr, "warning: skip window %s: %v\n", wid, serr)
						spaceErr[wid] = true
						continue
					}
					spaceCache[wid] = idx
				}
				if idx == wantSpace {
					keep = true
					break
				}
			}
		}
		if !keep {
			continue
		}
		seen[sid] = true
		ids = append(ids, sid)
	}
	return ids, nil
}

func liveRowWindowIDs(row LiveHostingRow) []string {
	seen := map[string]bool{}
	var out []string
	for _, c := range row.Candidates {
		wid := strings.TrimSpace(c.WindowID)
		if wid == "" || seen[wid] {
			continue
		}
		seen[wid] = true
		out = append(out, wid)
	}
	if len(out) == 0 {
		// Fallback: parse w=<id> from ITerm column.
		s := strings.TrimSpace(row.ITerm)
		if i := strings.Index(s, "w="); i >= 0 {
			rest := s[i+2:]
			end := len(rest)
			for j, r := range rest {
				if r == ' ' || r == 't' || r == '(' {
					end = j
					break
				}
			}
			wid := strings.TrimSpace(rest[:end])
			if wid != "" {
				out = append(out, wid)
			}
		}
	}
	return out
}

type promptsArgs struct {
	Positional []string
	SessionID  *string
	Tab        *string
	TabIndex   *int
	ThisTab    bool
	ThisWindow bool
	ThisSpace  bool
	First      bool
	MainOnly   bool
	Recent     time.Duration
	RecentSet  bool
	Limit      int
	LimitSet   bool
	Grep       []string
	GrepSet    bool
	Exclude    string
	ExcludeSet bool
	Head       int
	HeadSet    bool
	Tail       int
	TailSet    bool
	MaxBody    int
	MaxBodySet bool
	ColorMode  string
	Help       bool
}

func parsePromptsArgs(args []string) (promptsArgs, error) {
	var out promptsArgs
	var colorFlag, noColorFlag bool
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "-h" || arg == "--help" {
			out.Help = true
			return out, nil
		}
		if arg == "--color" {
			colorFlag = true
			continue
		}
		if arg == "--no-color" {
			noColorFlag = true
			continue
		}
		if arg == "--first" {
			out.First = true
			continue
		}
		if arg == "--main" || arg == "--main-agent" {
			out.MainOnly = true
			continue
		}
		if arg == "--this-tab" {
			out.ThisTab = true
			continue
		}
		if arg == "--this-window" {
			out.ThisWindow = true
			continue
		}
		if arg == "--this-space" {
			out.ThisSpace = true
			continue
		}
		if arg == "--session-id" || strings.HasPrefix(arg, "--session-id=") {
			raw, _, err := takeFlagValue(arg, "--session-id", args, &i)
			if err != nil {
				return out, err
			}
			out.SessionID = &raw
			continue
		}
		if arg == "--tab" || strings.HasPrefix(arg, "--tab=") {
			raw, _, err := takeFlagValue(arg, "--tab", args, &i)
			if err != nil {
				return out, err
			}
			out.Tab = &raw
			continue
		}
		if arg == "--tab-index" || strings.HasPrefix(arg, "--tab-index=") {
			raw, _, err := takeFlagValue(arg, "--tab-index", args, &i)
			if err != nil {
				return out, err
			}
			n, err := strconv.Atoi(raw)
			if err != nil {
				return out, fmt.Errorf("--tab-index must be an integer")
			}
			out.TabIndex = &n
			continue
		}
		if arg == "--recent" || strings.HasPrefix(arg, "--recent=") {
			raw, _, err := takeFlagValue(arg, "--recent", args, &i)
			if err != nil {
				return out, err
			}
			w, err := ParseRecentWindow(raw)
			if err != nil {
				return out, err
			}
			out.Recent = w
			out.RecentSet = true
			continue
		}
		if arg == "--limit" || strings.HasPrefix(arg, "--limit=") {
			raw, _, err := takeFlagValue(arg, "--limit", args, &i)
			if err != nil {
				return out, err
			}
			n, err := strconv.Atoi(raw)
			if err != nil {
				return out, fmt.Errorf("--limit must be an integer")
			}
			if n < 1 {
				return out, fmt.Errorf("--limit must be >= 1")
			}
			out.Limit = n
			out.LimitSet = true
			continue
		}
		if arg == "--grep" || strings.HasPrefix(arg, "--grep=") {
			raw, _, err := takeFlagValue(arg, "--grep", args, &i)
			if err != nil {
				return out, err
			}
			out.Grep = append(out.Grep, raw)
			out.GrepSet = true
			continue
		}
		if arg == "--exclude" || strings.HasPrefix(arg, "--exclude=") {
			raw, _, err := takeFlagValue(arg, "--exclude", args, &i)
			if err != nil {
				return out, err
			}
			out.Exclude = raw
			out.ExcludeSet = true
			continue
		}
		if arg == "--head" || strings.HasPrefix(arg, "--head=") {
			raw, _, err := takeFlagValue(arg, "--head", args, &i)
			if err != nil {
				return out, err
			}
			n, err := strconv.Atoi(raw)
			if err != nil {
				return out, fmt.Errorf("--head must be an integer")
			}
			out.Head = n
			out.HeadSet = true
			continue
		}
		if arg == "--tail" || strings.HasPrefix(arg, "--tail=") {
			raw, _, err := takeFlagValue(arg, "--tail", args, &i)
			if err != nil {
				return out, err
			}
			n, err := strconv.Atoi(raw)
			if err != nil {
				return out, fmt.Errorf("--tail must be an integer")
			}
			out.Tail = n
			out.TailSet = true
			continue
		}
		if arg == "--max-body" || strings.HasPrefix(arg, "--max-body=") {
			raw, _, err := takeFlagValue(arg, "--max-body", args, &i)
			if err != nil {
				return out, err
			}
			n, err := strconv.Atoi(raw)
			if err != nil {
				return out, fmt.Errorf("--max-body must be an integer")
			}
			out.MaxBody = n
			out.MaxBodySet = true
			continue
		}
		if strings.HasPrefix(arg, "-") {
			return out, fmt.Errorf("unknown flag: %s", arg)
		}
		out.Positional = append(out.Positional, arg)
	}

	if colorFlag && noColorFlag {
		return out, fmt.Errorf("--color and --no-color cannot be specified together")
	}
	switch {
	case colorFlag:
		out.ColorMode = "always"
	case noColorFlag:
		out.ColorMode = "never"
	default:
		out.ColorMode = "auto"
	}

	if out.Tab != nil && out.TabIndex != nil {
		return out, fmt.Errorf("--tab and --tab-index cannot be specified together")
	}
	if out.ThisTab && (out.Tab != nil || out.TabIndex != nil) {
		return out, fmt.Errorf("--this-tab cannot be combined with --tab/--tab-index")
	}
	if out.ThisWindow && out.ThisSpace {
		return out, fmt.Errorf("--this-window and --this-space cannot be specified together")
	}

	if out.First {
		if out.HeadSet || out.TailSet {
			return out, fmt.Errorf("--first and --head/--tail are mutually exclusive")
		}
		out.Head = 1
		out.HeadSet = true
	}
	if out.HeadSet && out.TailSet {
		return out, fmt.Errorf("--head and --tail are mutually exclusive")
	}
	if out.HeadSet && out.Head < 1 {
		return out, fmt.Errorf("--head must be >= 1 (got %d)", out.Head)
	}
	if out.TailSet && out.Tail < 1 {
		return out, fmt.Errorf("--tail must be >= 1 (got %d)", out.Tail)
	}
	if out.MaxBodySet && out.MaxBody < 1 {
		return out, fmt.Errorf("--max-body must be >= 1 (got %d)", out.MaxBody)
	}
	if out.GrepSet {
		for _, p := range out.Grep {
			if strings.TrimSpace(p) == "" {
				return out, fmt.Errorf("--grep pattern must not be empty")
			}
		}
		if len(out.Grep) == 0 {
			return out, fmt.Errorf("--grep pattern must not be empty")
		}
	}
	if out.ExcludeSet && out.Exclude == "" {
		return out, fmt.Errorf("--exclude pattern must not be empty")
	}
	// Validate --tab early when provided (including current).
	if out.Tab != nil {
		if _, err := tabselect.ParseTabFlag(*out.Tab); err != nil {
			return out, err
		}
	}
	if out.TabIndex != nil {
		if _, err := tabselect.ParseTabIndexFlag(*out.TabIndex); err != nil {
			return out, err
		}
	}
	return out, nil
}
