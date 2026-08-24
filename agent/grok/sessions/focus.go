package sessions

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"

	"github.com/xhd2015/agent-pro/pkgs/procresolve"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
)

// ErrNotFound is returned when the Grok session id is unknown, no live grok
// process hosts the session, or no iTerm2 tab matches that process TTY.
var ErrNotFound = errors.New("not found")

// FocusHelp is the text for `agent-pro grok session focus --help`.
const FocusHelp = `Usage: agent-pro grok session focus <session-id> [--index N]
  --index N   select candidate N when multiple tabs host the same session
  -h,--help   show help
`

// FocusCommandHelpLine is the parent `agent-pro grok session` help row.
const FocusCommandHelpLine = `  focus  <session-id>   focus the iTerm2 tab that hosts this Grok session`

// FocusProc is one process-table row used to walk from a live grok PID to a TTY.
type FocusProc struct {
	PID  int
	PPID int
	TTY  string // e.g. /dev/ttys148, ttys148, "??", or ""
	Cmd  string
}

// FocusCandidate is one iTerm2 tab that hosts the Grok session.
// Order is the stable 0-based --index order.
type FocusCandidate struct {
	WindowID    string
	WindowTitle string
	TabIndex    int
	SessionID   string // iTerm session id
	TTY         string
	PID         int
}

// FocusOpts drives Focus. Nil hooks use production process / iTerm probes.
type FocusOpts struct {
	Index      *int
	ListProcs  func() []FocusProc
	Lsof       func(int) []string
	ListITerm  func() ([]iterm2.SessionRef, error)
	FocusITerm func(iterm2.SessionRef) error
}

// FocusResult is the selected candidate after a successful focus.
type FocusResult struct {
	Candidate FocusCandidate
}

// FocusFake is the deterministic injected process + iTerm boundary used by tests.
type FocusFake struct {
	Procs          []FocusProc
	OpenFiles      map[int][]string
	ITerm          []iterm2.SessionRef
	Focused        []string
	ListProcsCalls int
	LsofCalls      int
	ListITermCalls int
}

// Opts returns FocusOpts wired to this fake. Safe to call more than once.
func (f *FocusFake) Opts() *FocusOpts {
	return &FocusOpts{
		ListProcs: func() []FocusProc {
			f.ListProcsCalls++
			return append([]FocusProc(nil), f.Procs...)
		},
		Lsof: func(pid int) []string {
			f.LsofCalls++
			if f.OpenFiles == nil {
				return nil
			}
			return f.OpenFiles[pid]
		},
		ListITerm: func() ([]iterm2.SessionRef, error) {
			f.ListITermCalls++
			return append([]iterm2.SessionRef(nil), f.ITerm...), nil
		},
		FocusITerm: func(ref iterm2.SessionRef) error {
			f.Focused = append(f.Focused, ref.SessionID)
			return nil
		},
	}
}

// FocusDiscovery is the live-hosting probe result for a known Grok session.
// Candidates may be empty when there is no live PID, no TTY, or no iTerm match.
type FocusDiscovery struct {
	Candidates []FocusCandidate
	LiveCount  int // live grok PIDs hard-hitting the session (before TTY/iTerm)
}

// DiscoverFocusHosting maps a known Grok session to iTerm tabs via live PID → TTY.
// It does not call Find and never focuses. Session existence is the caller's job.
//
// Production (ListITerm nil) uses a TTY-targeted iTerm scan (FindSessionsByTTY)
// instead of dumping every pane. Injected ListITerm keeps the full-list path
// (tests and ListLive shared inventory).
func DiscoverFocusHosting(sessionID string, opts *FocusOpts) (*FocusDiscovery, error) {
	if opts == nil {
		opts = &FocusOpts{}
	}
	listProcs := opts.ListProcs
	if listProcs == nil {
		listProcs = listLiveFocusProcs
	}
	lsof := opts.Lsof
	if lsof == nil {
		lsof = procresolve.LiveLsof
	}
	procs := listProcs()

	live, err := LivePIDsForSession(sessionID, &LiveOptions{
		ListProcs: func() []procresolve.Proc {
			out := make([]procresolve.Proc, 0, len(procs))
			for _, p := range procs {
				out = append(out, procresolve.Proc{PID: p.PID, PPID: p.PPID, Cmd: p.Cmd})
			}
			return out
		},
		Lsof: lsof,
	})
	if err != nil {
		return nil, err
	}
	out := &FocusDiscovery{LiveCount: len(live)}
	if len(live) == 0 {
		return out, nil
	}

	var ttys []string
	seenTTY := map[string]bool{}
	for _, p := range live {
		for _, tty := range collectTTYsFromTree(procs, p.PID) {
			if seenTTY[tty] {
				continue
			}
			seenTTY[tty] = true
			ttys = append(ttys, tty)
		}
	}
	if len(ttys) == 0 {
		return out, nil
	}

	var matched []iterm2.SessionRef
	if opts.ListITerm != nil {
		refs, listErr := opts.ListITerm()
		if listErr != nil {
			return nil, listErr
		}
		matched = uniqueSessionRefs(iterm2.FindByTTY(refs, ttys))
	} else {
		refs, findErr := iterm2.FindSessionsByTTY(ttys)
		if findErr != nil {
			return nil, findErr
		}
		matched = uniqueSessionRefs(refs)
	}
	if len(matched) == 0 {
		return out, nil
	}

	ttyOwner := map[string]FocusProc{}
	for _, p := range procs {
		n := iterm2.NormalizeTTY(strings.TrimSpace(p.TTY))
		if n == "" || strings.TrimSpace(p.TTY) == "??" {
			continue
		}
		if _, ok := ttyOwner[n]; !ok {
			ttyOwner[n] = p
		}
	}

	candidates := make([]FocusCandidate, 0, len(matched))
	for _, ref := range matched {
		tty := iterm2.NormalizeTTY(ref.TTY)
		cand := FocusCandidate{
			WindowID:    ref.WindowID,
			WindowTitle: ref.WindowName,
			TabIndex:    ref.TabIndex,
			SessionID:   ref.SessionID,
			TTY:         tty,
		}
		if owner, ok := ttyOwner[tty]; ok {
			cand.PID = owner.PID
		}
		candidates = append(candidates, cand)
	}
	out.Candidates = candidates
	return out, nil
}

// Focus finds the Grok session, maps live grok PIDs to iTerm tabs via TTY,
// and focuses one existing tab. It never creates a window, tab, or session.
func Focus(grokHome, sessionID string, opts *FocusOpts) (*FocusResult, error) {
	if opts == nil {
		opts = &FocusOpts{}
	}
	if _, err := Find(grokHome, sessionID); err != nil {
		if isSessionNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	disc, err := DiscoverFocusHosting(sessionID, opts)
	if err != nil {
		return nil, err
	}
	if disc == nil || len(disc.Candidates) == 0 {
		return nil, ErrNotFound
	}

	selected, err := selectFocusCandidate(sessionID, disc.Candidates, opts.Index)
	if err != nil {
		return nil, err
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
	return &FocusResult{Candidate: selected}, nil
}

// RunFocus implements `agent-pro grok session focus` with injectable writers
// and process/iTerm hooks. A nil opts uses production probes.
func RunFocus(args []string, stdout io.Writer, grokHome string, opts *FocusOpts) error {
	if stdout == nil {
		stdout = io.Discard
	}
	sessionID, index, help, err := parseFocusArgs(args)
	if err != nil {
		return err
	}
	if help {
		txt := FocusHelp
		if !strings.HasSuffix(txt, "\n") {
			txt += "\n"
		}
		_, _ = io.WriteString(stdout, txt)
		return nil
	}

	runOpts := FocusOpts{}
	if opts != nil {
		runOpts = *opts
	}
	runOpts.Index = index
	result, err := Focus(grokHome, sessionID, &runOpts)
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "focused: window %s, tab %d\n", result.Candidate.WindowID, result.Candidate.TabIndex)
	return nil
}

func parseFocusArgs(args []string) (sessionID string, index *int, help bool, err error) {
	var positional []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "-h" || arg == "--help" {
			return "", nil, true, nil
		}
		if arg == "--index" {
			if i+1 >= len(args) {
				return "", nil, false, fmt.Errorf("--index must be an integer")
			}
			n, convErr := strconv.Atoi(args[i+1])
			if convErr != nil {
				return "", nil, false, fmt.Errorf("--index must be an integer")
			}
			index = &n
			i++
			continue
		}
		if strings.HasPrefix(arg, "--index=") {
			n, convErr := strconv.Atoi(strings.TrimPrefix(arg, "--index="))
			if convErr != nil {
				return "", nil, false, fmt.Errorf("--index must be an integer")
			}
			index = &n
			continue
		}
		if strings.HasPrefix(arg, "-") {
			return "", nil, false, fmt.Errorf("unrecognized flag: %s", arg)
		}
		positional = append(positional, arg)
	}
	if len(positional) != 1 {
		return "", nil, false, fmt.Errorf("expected exactly one session id, got %d arguments", len(positional))
	}
	sessionID = strings.TrimSpace(positional[0])
	if sessionID == "" {
		return "", nil, false, fmt.Errorf("session id is required")
	}
	return sessionID, index, false, nil
}

func selectFocusCandidate(sessionID string, candidates []FocusCandidate, index *int) (FocusCandidate, error) {
	if index == nil && len(candidates) > 1 {
		return FocusCandidate{}, formatMultipleFocusError(sessionID, candidates)
	}
	selected := 0
	if index != nil {
		selected = *index
		if selected < 0 || selected >= len(candidates) {
			return FocusCandidate{}, formatIndexOOBError(selected, candidates)
		}
	}
	return candidates[selected], nil
}

func formatMultipleFocusError(sessionID string, candidates []FocusCandidate) error {
	var b strings.Builder
	fmt.Fprintf(&b, "multiple iTerm2 tabs host session %s\n\n", sessionID)
	printFocusCandidates(&b, candidates)
	fmt.Fprintf(&b, "\nSpecify one with:\n  agent-pro grok session focus %s --index <%s>", sessionID, validFocusIndexes(candidates))
	return errors.New(b.String())
}

func formatIndexOOBError(index int, candidates []FocusCandidate) error {
	var b strings.Builder
	fmt.Fprintf(&b, "--index %d is out of range (valid indexes: %s)\n\n", index, validFocusIndexes(candidates))
	printFocusCandidates(&b, candidates)
	return errors.New(strings.TrimRight(b.String(), "\n"))
}

func printFocusCandidates(w io.Writer, candidates []FocusCandidate) {
	for i, c := range candidates {
		fmt.Fprintf(w, "  [%d] window %s (%q) tab %d tty %s session %s\n", i, c.WindowID, c.WindowTitle, c.TabIndex, c.TTY, c.SessionID)
	}
}

func validFocusIndexes(candidates []FocusCandidate) string {
	indexes := make([]string, len(candidates))
	for i := range candidates {
		indexes[i] = strconv.Itoa(i)
	}
	return strings.Join(indexes, "|")
}

func isSessionNotFound(err error) bool {
	return err != nil && strings.Contains(err.Error(), "grok session not found")
}

// collectTTYsFromTree returns real TTYs from the root PID plus its ancestors
// and descendants. Skips "??", blank, and whitespace-only TTY fields.
func collectTTYsFromTree(procs []FocusProc, rootPID int) []string {
	if rootPID <= 0 || len(procs) == 0 {
		return nil
	}
	byPID := make(map[int]FocusProc, len(procs))
	children := make(map[int][]int, len(procs))
	for _, p := range procs {
		byPID[p.PID] = p
		if p.PPID == 0 && p.PID == p.PPID {
			continue
		}
		children[p.PPID] = append(children[p.PPID], p.PID)
	}
	if _, ok := byPID[rootPID]; !ok {
		return nil
	}

	seenPID := map[int]bool{rootPID: true}
	var ordered []FocusProc

	cur := rootPID
	var ancestors []FocusProc
	for {
		row, ok := byPID[cur]
		if !ok {
			break
		}
		if cur != rootPID {
			ancestors = append(ancestors, row)
			seenPID[cur] = true
		}
		if row.PPID == 0 || row.PPID == cur {
			break
		}
		if seenPID[row.PPID] {
			break
		}
		cur = row.PPID
	}

	ordered = append(ordered, byPID[rootPID])
	ordered = append(ordered, ancestors...)

	queue := []int{rootPID}
	for len(queue) > 0 {
		pid := queue[0]
		queue = queue[1:]
		for _, child := range children[pid] {
			if seenPID[child] {
				continue
			}
			seenPID[child] = true
			if row, ok := byPID[child]; ok {
				ordered = append(ordered, row)
			}
			queue = append(queue, child)
		}
	}

	seenTTY := map[string]bool{}
	var out []string
	for _, row := range ordered {
		tty := strings.TrimSpace(row.TTY)
		if tty == "" || tty == "??" {
			continue
		}
		norm := iterm2.NormalizeTTY(tty)
		if norm == "" || seenTTY[norm] {
			continue
		}
		seenTTY[norm] = true
		out = append(out, norm)
	}
	return out
}

func uniqueSessionRefs(refs []iterm2.SessionRef) []iterm2.SessionRef {
	seen := map[string]bool{}
	var out []iterm2.SessionRef
	for _, ref := range refs {
		key := strings.TrimSpace(ref.SessionID)
		if key == "" {
			key = ref.WindowID + "\t" + strconv.Itoa(ref.TabIndex) + "\t" + iterm2.NormalizeTTY(ref.TTY)
		}
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, ref)
	}
	return out
}

func listLiveFocusProcs() []FocusProc {
	cmd := exec.Command("ps", "-ax", "-o", "pid=,ppid=,tty=,command=")
	out, err := cmd.Output()
	if err != nil {
		cmd = exec.Command("ps", "-x", "-o", "pid=,ppid=,tty=,command=")
		out, err = cmd.Output()
		if err != nil {
			return nil
		}
	}
	return parseFocusProcRows(out)
}

func parseFocusProcRows(out []byte) []FocusProc {
	var rows []FocusProc
	sc := bufio.NewScanner(bytes.NewReader(out))
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		pid, err1 := strconv.Atoi(fields[0])
		ppid, err2 := strconv.Atoi(fields[1])
		if err1 != nil || err2 != nil {
			continue
		}
		cmdStr := ""
		if len(fields) > 3 {
			cmdStr = strings.Join(fields[3:], " ")
		}
		rows = append(rows, FocusProc{PID: pid, PPID: ppid, TTY: fields[2], Cmd: cmdStr})
	}
	return rows
}

func listLiveITermSessions() ([]iterm2.SessionRef, error) {
	script := iterm2.BuildSessionListScript()
	cmd := exec.Command("osascript", "-e", script)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("list iTerm sessions: %w", err)
	}
	return iterm2.ParseSessionListOutput(string(out))
}

// liveTTYsForSession returns normalized TTYs for live grok PIDs hosting
// sessionID. ok is false when there is no live host or no usable TTY.
func liveTTYsForSession(sessionID string, listProcs func() []FocusProc, lsof func(int) []string) ([]string, bool) {
	if listProcs == nil {
		listProcs = listLiveFocusProcs
	}
	if lsof == nil {
		lsof = procresolve.LiveLsof
	}
	procs := listProcs()
	live, err := LivePIDsForSession(sessionID, &LiveOptions{
		ListProcs: func() []procresolve.Proc {
			out := make([]procresolve.Proc, 0, len(procs))
			for _, p := range procs {
				out = append(out, procresolve.Proc{PID: p.PID, PPID: p.PPID, Cmd: p.Cmd})
			}
			return out
		},
		Lsof: lsof,
	})
	if err != nil || len(live) == 0 {
		return nil, false
	}
	seen := map[string]bool{}
	var ttys []string
	for _, p := range live {
		for _, tty := range collectTTYsFromTree(procs, p.PID) {
			n := iterm2.NormalizeTTY(tty)
			if n == "" || seen[n] {
				continue
			}
			seen[n] = true
			ttys = append(ttys, n)
		}
	}
	if len(ttys) == 0 {
		return nil, false
	}
	return ttys, true
}
