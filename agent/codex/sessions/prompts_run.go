package sessions

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/xhd2015/dot-pkgs/go-pkgs/computer-use/macos/space"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2/tabselect"
)

const (
	codexDefaultPromptsListLimit = 10
	codexPromptsMissingTS        = "[—]"
	codexPromptsSep              = "────────────────────────────────────────"
	codexPromptsEllipsis         = "…"
)

// PromptsHelp is the text for `agent-pro codex session prompts --help`.
const PromptsHelp = `Usage:
  agent-pro codex session prompts (<session-id> | --session-id ID | --tab SEL | --tab-index N | --this-tab)
    [--first] [--grep P]... [--exclude Q] [--head N | --tail N] [--max-body N]
    [--color|--no-color]
  agent-pro codex session prompts [--this-window | --this-space]
    [--first] [--recent <window>] [--limit N]
    [--grep P]... [--exclude Q] [--head N | --tail N] [--max-body N]
    [--color|--no-color]
  agent-pro codex session prompts [--recent <window>] [--limit N]
    [--first] [--grep P]... [--exclude Q] [--head N | --tail N] [--max-body N]
    [--color|--no-color]

Show user prompts only as compact lines:
  [YYYY-MM-DD HH:MM:SS] prompt text…

Same selection matrix and filters as: agent-pro grok session prompts.
`

// PromptsCommandHelpLine is the parent `agent-pro codex session` help row.
const PromptsCommandHelpLine = `  prompts [session-id]  list user prompts (one session or recent multi)`

// PromptsOpts injects probes for RunPrompts. Nil hooks use production.
type PromptsOpts struct {
	ListLive         func() ([]LiveHostingRow, error)
	CurrentWindow    func() (iterm2.WindowStatus, error)
	ListUserSpaces   func() ([]space.UserSpace, error)
	SpaceIndexForWin func(windowID uint64) (int, error)
	HomeDir          func() string
	Now              func() time.Time
	ListProcs        func() []FocusProc
	Lsof             func(int) []string
	ListITerm        func() ([]iterm2.SessionRef, error)
	CurrentSessionID func() string
	ControllingTTY   func() string
	AncestorTTYs     func() []string
}

type codexUserPrompt struct {
	Index     int
	Timestamp time.Time
	Text      string
}

type codexSessionPrompts struct {
	Session
	UserPrompts   []codexUserPrompt
	OmittedBefore int
	OmittedAfter  int
}

// RunPrompts implements `agent-pro codex session prompts`.
func RunPrompts(args []string, stdout, stderr io.Writer, codexHome string, opts *PromptsOpts) error {
	if stdout == nil {
		stdout = io.Discard
	}
	if stderr == nil {
		stderr = io.Discard
	}
	parsed, err := parseCodexPromptsArgs(args)
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

	singleID, err := resolveCodexPromptsSingleSession(parsed, &runOpts)
	if err != nil {
		return err
	}
	if singleID != "" {
		if parsed.RecentSet || parsed.LimitSet || parsed.ThisWindow || parsed.ThisSpace {
			return fmt.Errorf("session id cannot be combined with --recent, --limit, --this-window, or --this-space")
		}
		sp, err := loadCodexSessionPrompts(codexHome, singleID)
		if err != nil {
			return err
		}
		sp.UserPrompts, sp.OmittedBefore, sp.OmittedAfter, err = filterCodexUserPrompts(sp.UserPrompts, parsed)
		if err != nil {
			return err
		}
		bw := bufio.NewWriter(stdout)
		defer bw.Flush()
		return writeCodexPromptsText(bw, sp, parsed, now, home, false)
	}

	ids, hostScoped, err := resolveCodexPromptsMultiIDs(parsed, codexHome, stderr, &runOpts)
	if err != nil {
		return err
	}
	bw := bufio.NewWriterSize(stdout, 1024)
	defer bw.Flush()
	fmt.Fprintln(stderr, "scanning sessions…")
	return streamCodexPromptsList(bw, stderr, codexHome, ids, hostScoped, parsed, now, home)
}

func loadCodexSessionPrompts(codexHome, sessionID string) (*codexSessionPrompts, error) {
	path, err := Find(codexHome, sessionID)
	if err != nil {
		return nil, err
	}
	msgs, err := loadChatMessagesFromRollout(path)
	if err != nil {
		return nil, err
	}
	var prompts []codexUserPrompt
	idx := 0
	for _, m := range msgs {
		if m.Kind != MessageKindUser {
			continue
		}
		text := strings.TrimSpace(m.Text)
		if text == "" {
			continue
		}
		idx++
		prompts = append(prompts, codexUserPrompt{Index: idx, Timestamp: m.Timestamp, Text: text})
	}
	sess := Session{ID: sessionID, Path: path}
	if sessions, derr := discoverSessions(codexHome); derr == nil {
		for _, s := range sessions {
			if s.ID == sessionID {
				sess = s
				break
			}
		}
	}
	return &codexSessionPrompts{Session: sess, UserPrompts: prompts}, nil
}

func resolveCodexPromptsSingleSession(parsed codexPromptsArgs, opts *PromptsOpts) (string, error) {
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
		})
		return id, err
	}
	return "", nil
}

func resolveCodexPromptsMultiIDs(parsed codexPromptsArgs, codexHome string, stderr io.Writer, opts *PromptsOpts) ([]string, bool, error) {
	if parsed.ThisWindow || parsed.ThisSpace {
		ids, err := resolveCodexHostScopeIDs(parsed, stderr, opts)
		return ids, true, err
	}
	sessions, err := discoverSessions(codexHome)
	if err != nil {
		return nil, false, err
	}
	sort.Slice(sessions, func(i, j int) bool {
		if sessions[i].StartedAt.Equal(sessions[j].StartedAt) {
			return sessions[i].ID > sessions[j].ID
		}
		return sessions[i].StartedAt.After(sessions[j].StartedAt)
	})
	ids := make([]string, 0, len(sessions))
	for _, s := range sessions {
		ids = append(ids, s.ID)
	}
	return ids, false, nil
}

func resolveCodexHostScopeIDs(parsed codexPromptsArgs, stderr io.Writer, opts *PromptsOpts) ([]string, error) {
	listLive := opts.ListLive
	if listLive == nil {
		listLive = func() ([]LiveHostingRow, error) { return ListLive(nil) }
	}
	rows, err := listLive()
	if err != nil {
		return nil, err
	}
	currentWindow := opts.CurrentWindow
	if currentWindow == nil {
		currentWindow = func() (iterm2.WindowStatus, error) { return iterm2.CurrentWindowStatus() }
	}
	listSpaces := opts.ListUserSpaces
	if listSpaces == nil {
		listSpaces = func() ([]space.UserSpace, error) { return space.ListUserSpaces() }
	}
	spaceIdx := opts.SpaceIndexForWin
	if spaceIdx == nil {
		spaceIdx = func(windowID uint64) (int, error) { return space.SpaceIndexForWindow(windowID) }
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
		winIDs := codexLiveRowWindowIDs(row)
		keep := false
		if parsed.ThisWindow {
			for _, wid := range winIDs {
				if wid == wantWindow {
					keep = true
					break
				}
			}
		} else {
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

func codexLiveRowWindowIDs(row LiveHostingRow) []string {
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

func streamCodexPromptsList(w, stderr io.Writer, codexHome string, ids []string, hostScoped bool, parsed codexPromptsArgs, now time.Time, home string) error {
	_ = stderr
	capN, hasCap := codexPromptsSessionCap(parsed, hostScoped)
	nSess := 0
	totalMsgs := 0
	wroteAny := false
	for _, id := range ids {
		sp, err := loadCodexSessionPrompts(codexHome, id)
		if err != nil {
			continue
		}
		prompts := sp.UserPrompts
		if parsed.RecentSet {
			prompts = filterCodexPromptsInWindow(prompts, now, parsed.Recent)
		}
		prompts, ob, oa, err := filterCodexUserPrompts(prompts, parsed)
		if err != nil {
			return err
		}
		if len(prompts) == 0 {
			continue
		}
		sp.UserPrompts = prompts
		sp.OmittedBefore = ob
		sp.OmittedAfter = oa
		if nSess > 0 {
			if _, err := io.WriteString(w, "\n"+codexPromptsSep+"\n\n"); err != nil {
				return err
			}
		}
		title := strings.TrimSpace(sp.Title)
		if title == "" {
			title = "(untitled)"
		}
		header := fmt.Sprintf("── %s  ·  %s  ·  %s  ·  %s",
			sp.ID, formatRelativeTime(sp.StartedAt, now), title, shortenCodexPath(sp.CWD, home))
		if _, err := fmt.Fprintln(w, header); err != nil {
			return err
		}
		if err := writeCodexPromptsText(w, sp, parsed, now, home, true); err != nil {
			return err
		}
		if f, ok := w.(interface{ Flush() error }); ok {
			_ = f.Flush()
		}
		nSess++
		totalMsgs += len(prompts)
		wroteAny = true
		if hasCap && nSess >= capN {
			break
		}
	}
	if !wroteAny {
		_, err := io.WriteString(w, "No user prompts found\n")
		return err
	}
	footer := fmt.Sprintf("%d sessions, %d user messages", nSess, totalMsgs)
	if parsed.RecentSet && parsed.Recent > 0 {
		footer += fmt.Sprintf(" (recent %s)", formatCodexWindowShort(parsed.Recent))
	}
	if parsed.LimitSet && parsed.Limit > 0 {
		footer += fmt.Sprintf(" (limit %d)", parsed.Limit)
	}
	_, err := fmt.Fprintln(w, footer)
	return err
}

func codexPromptsSessionCap(parsed codexPromptsArgs, hostScoped bool) (int, bool) {
	if hostScoped && !parsed.LimitSet && !parsed.RecentSet {
		return 0, false
	}
	if parsed.RecentSet {
		if parsed.LimitSet {
			return parsed.Limit, true
		}
		return 0, false
	}
	if parsed.LimitSet {
		return parsed.Limit, true
	}
	return codexDefaultPromptsListLimit, true
}

func filterCodexPromptsInWindow(prompts []codexUserPrompt, now time.Time, window time.Duration) []codexUserPrompt {
	start := now.Add(-window)
	var out []codexUserPrompt
	for _, p := range prompts {
		if p.Timestamp.IsZero() {
			continue
		}
		if p.Timestamp.Before(start) || p.Timestamp.After(now) {
			continue
		}
		out = append(out, p)
	}
	return out
}

func filterCodexUserPrompts(prompts []codexUserPrompt, parsed codexPromptsArgs) ([]codexUserPrompt, int, int, error) {
	kept := prompts
	if parsed.GrepSet {
		var filtered []codexUserPrompt
		for _, p := range kept {
			if codexTextContainsAllCI(p.Text, parsed.Grep) {
				filtered = append(filtered, p)
			}
		}
		kept = filtered
	}
	if parsed.ExcludeSet {
		var filtered []codexUserPrompt
		for _, p := range kept {
			if strings.Contains(strings.ToLower(p.Text), strings.ToLower(parsed.Exclude)) {
				continue
			}
			filtered = append(filtered, p)
		}
		kept = filtered
	}
	omittedBefore, omittedAfter := 0, 0
	if parsed.HeadSet {
		if len(kept) > parsed.Head {
			omittedAfter = len(kept) - parsed.Head
			kept = append([]codexUserPrompt(nil), kept[:parsed.Head]...)
		}
	} else if parsed.TailSet {
		if len(kept) > parsed.Tail {
			omittedBefore = len(kept) - parsed.Tail
			kept = append([]codexUserPrompt(nil), kept[len(kept)-parsed.Tail:]...)
		}
	}
	if kept == nil {
		kept = []codexUserPrompt{}
	}
	return kept, omittedBefore, omittedAfter, nil
}

func writeCodexPromptsText(w io.Writer, sp *codexSessionPrompts, parsed codexPromptsArgs, now time.Time, home string, bodyOnly bool) error {
	_ = now
	_ = home
	if !bodyOnly && (sp == nil || len(sp.UserPrompts) == 0) {
		_, err := io.WriteString(w, "No user prompts found\n")
		return err
	}
	if sp.OmittedBefore > 0 {
		if _, err := fmt.Fprintf(w, "(...%d omitted...)\n", sp.OmittedBefore); err != nil {
			return err
		}
	}
	for _, p := range sp.UserPrompts {
		ts := codexPromptsMissingTS
		if !p.Timestamp.IsZero() {
			ts = p.Timestamp.In(time.Local).Format("2006-01-02 15:04:05")
		}
		body := collapseCodexPromptBody(p.Text)
		if parsed.MaxBodySet && parsed.MaxBody > 0 {
			body = softCapCodexRunes(body, parsed.MaxBody)
		}
		if _, err := fmt.Fprintf(w, "[%s] %s\n", ts, body); err != nil {
			return err
		}
	}
	if sp.OmittedAfter > 0 {
		if _, err := fmt.Fprintf(w, "(...%d omitted...)\n", sp.OmittedAfter); err != nil {
			return err
		}
	}
	return nil
}

func collapseCodexPromptBody(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func softCapCodexRunes(s string, n int) string {
	if n < 1 || utf8.RuneCountInString(s) <= n {
		return s
	}
	runes := []rune(s)
	return string(runes[:n]) + codexPromptsEllipsis
}

func codexTextContainsAllCI(text string, patterns []string) bool {
	lower := strings.ToLower(text)
	for _, p := range patterns {
		if !strings.Contains(lower, strings.ToLower(p)) {
			return false
		}
	}
	return true
}

func shortenCodexPath(path, home string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return "-"
	}
	home = strings.TrimSpace(home)
	if home != "" && (path == home || strings.HasPrefix(path, home+"/")) {
		return "~" + strings.TrimPrefix(path, home)
	}
	return path
}

func formatCodexWindowShort(d time.Duration) string {
	if d%(24*time.Hour) == 0 {
		return fmt.Sprintf("%dd", int(d/(24*time.Hour)))
	}
	if d%time.Hour == 0 {
		return fmt.Sprintf("%dh", int(d/time.Hour))
	}
	return fmt.Sprintf("%dm", int(d/time.Minute))
}

var codexRecentWindowRE = regexp.MustCompile(`(?i)^([0-9]+)([dhm])$`)

func parseCodexRecentWindow(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	m := codexRecentWindowRE.FindStringSubmatch(s)
	if m == nil {
		return 0, fmt.Errorf("invalid recent window %q; use Nd, Nh, or Nm (e.g. 1d, 2h, 30m)", s)
	}
	n, err := strconv.Atoi(m[1])
	if err != nil || n <= 0 {
		return 0, fmt.Errorf("invalid recent window %q; use Nd, Nh, or Nm with a positive number", s)
	}
	switch strings.ToLower(m[2]) {
	case "d":
		return time.Duration(n) * 24 * time.Hour, nil
	case "h":
		return time.Duration(n) * time.Hour, nil
	case "m":
		return time.Duration(n) * time.Minute, nil
	default:
		return 0, fmt.Errorf("invalid recent window %q; use Nd, Nh, or Nm (e.g. 1d, 2h, 30m)", s)
	}
}

type codexPromptsArgs struct {
	Positional []string
	SessionID  *string
	Tab        *string
	TabIndex   *int
	ThisTab    bool
	ThisWindow bool
	ThisSpace  bool
	First      bool
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

func parseCodexPromptsArgs(args []string) (codexPromptsArgs, error) {
	var out codexPromptsArgs
	var colorFlag, noColorFlag bool
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "-h" || arg == "--help" {
			out.Help = true
			return out, nil
		}
		switch {
		case arg == "--color":
			colorFlag = true
		case arg == "--no-color":
			noColorFlag = true
		case arg == "--first":
			out.First = true
		case arg == "--this-tab":
			out.ThisTab = true
		case arg == "--this-window":
			out.ThisWindow = true
		case arg == "--this-space":
			out.ThisSpace = true
		case arg == "--session-id" || strings.HasPrefix(arg, "--session-id="):
			raw, _, err := takeMessagesFlagValue(arg, "--session-id", args, &i)
			if err != nil {
				return out, err
			}
			out.SessionID = &raw
		case arg == "--tab" || strings.HasPrefix(arg, "--tab="):
			raw, _, err := takeMessagesFlagValue(arg, "--tab", args, &i)
			if err != nil {
				return out, err
			}
			out.Tab = &raw
		case arg == "--tab-index" || strings.HasPrefix(arg, "--tab-index="):
			raw, _, err := takeMessagesFlagValue(arg, "--tab-index", args, &i)
			if err != nil {
				return out, err
			}
			n, err := strconv.Atoi(raw)
			if err != nil {
				return out, fmt.Errorf("--tab-index must be an integer")
			}
			out.TabIndex = &n
		case arg == "--recent" || strings.HasPrefix(arg, "--recent="):
			raw, _, err := takeMessagesFlagValue(arg, "--recent", args, &i)
			if err != nil {
				return out, err
			}
			w, err := parseCodexRecentWindow(raw)
			if err != nil {
				return out, err
			}
			out.Recent = w
			out.RecentSet = true
		case arg == "--limit" || strings.HasPrefix(arg, "--limit="):
			raw, _, err := takeMessagesFlagValue(arg, "--limit", args, &i)
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
		case arg == "--grep" || strings.HasPrefix(arg, "--grep="):
			raw, _, err := takeMessagesFlagValue(arg, "--grep", args, &i)
			if err != nil {
				return out, err
			}
			out.Grep = append(out.Grep, raw)
			out.GrepSet = true
		case arg == "--exclude" || strings.HasPrefix(arg, "--exclude="):
			raw, _, err := takeMessagesFlagValue(arg, "--exclude", args, &i)
			if err != nil {
				return out, err
			}
			out.Exclude = raw
			out.ExcludeSet = true
		case arg == "--head" || strings.HasPrefix(arg, "--head="):
			raw, _, err := takeMessagesFlagValue(arg, "--head", args, &i)
			if err != nil {
				return out, err
			}
			n, err := strconv.Atoi(raw)
			if err != nil {
				return out, fmt.Errorf("--head must be an integer")
			}
			out.Head = n
			out.HeadSet = true
		case arg == "--tail" || strings.HasPrefix(arg, "--tail="):
			raw, _, err := takeMessagesFlagValue(arg, "--tail", args, &i)
			if err != nil {
				return out, err
			}
			n, err := strconv.Atoi(raw)
			if err != nil {
				return out, fmt.Errorf("--tail must be an integer")
			}
			out.Tail = n
			out.TailSet = true
		case arg == "--max-body" || strings.HasPrefix(arg, "--max-body="):
			raw, _, err := takeMessagesFlagValue(arg, "--max-body", args, &i)
			if err != nil {
				return out, err
			}
			n, err := strconv.Atoi(raw)
			if err != nil {
				return out, fmt.Errorf("--max-body must be an integer")
			}
			out.MaxBody = n
			out.MaxBodySet = true
		case strings.HasPrefix(arg, "-"):
			return out, fmt.Errorf("unknown flag: %s", arg)
		default:
			out.Positional = append(out.Positional, arg)
		}
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
