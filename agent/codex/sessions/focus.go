package sessions

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/xhd2015/agent-pro/pkgs/procresolve"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
)

// FocusCandidate is one iTerm2 tab that hosts the Codex session.
// Order is the stable 0-based --index order.
type FocusCandidate struct {
	WindowID    string
	WindowTitle string
	TabIndex    int
	SessionID   string // iTerm session id
	TTY         string
	PID         int
}

// FocusOpts drives DiscoverFocusHosting. Nil hooks use production probes.
type FocusOpts struct {
	Index      *int
	ListProcs  func() []FocusProc
	Lsof       func(int) []string
	ListITerm  func() ([]iterm2.SessionRef, error)
	FocusITerm func(iterm2.SessionRef) error
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

// FocusDiscovery is the live-hosting probe result for a known Codex session.
type FocusDiscovery struct {
	Candidates []FocusCandidate
	LiveCount  int
}

// LivePID is one live process hard-hitting a session via open files.
type LivePID struct {
	PID  int
	Name string
	Cmd  string
}

// LiveOptions injects process listing and open-file probes for tests.
type LiveOptions struct {
	ListProcs func() []procresolve.Proc
	Lsof      func(int) []string
}

// LivePIDsForSession scans ListProcs for codex runners only. For each runner,
// Lsof open paths are checked for a hard hit on sessionID (kind codex).
func LivePIDsForSession(sessionID string, opts *LiveOptions) ([]LivePID, error) {
	sessionID = strings.TrimSpace(sessionID)
	listProcs := procresolve.ListLiveProcs
	lsof := procresolve.LiveLsof
	if opts != nil {
		if opts.ListProcs != nil {
			listProcs = opts.ListProcs
		}
		if opts.Lsof != nil {
			lsof = opts.Lsof
		}
	}

	procs := listProcs()
	var hits []LivePID
	for _, p := range procs {
		if !procresolve.IsCodexRunner(p.Cmd) {
			continue
		}
		paths := lsof(p.PID)
		if !openFilesHitSession(paths, sessionID) {
			continue
		}
		hits = append(hits, LivePID{
			PID:  p.PID,
			Name: argv0Base(p.Cmd),
			Cmd:  p.Cmd,
		})
	}
	sort.Slice(hits, func(i, j int) bool {
		return hits[i].PID < hits[j].PID
	})
	return hits, nil
}

func openFilesHitSession(paths []string, sessionID string) bool {
	want := strings.ToLower(strings.TrimSpace(sessionID))
	if want == "" {
		return false
	}
	for _, p := range paths {
		kind, id, ok := procresolve.ParseSessionFromPath(p)
		if !ok || kind != "codex" {
			continue
		}
		if strings.EqualFold(id, want) {
			return true
		}
	}
	return false
}

func argv0Base(cmd string) string {
	fields := strings.Fields(cmd)
	if len(fields) == 0 {
		return ""
	}
	return filepath.Base(fields[0])
}

// DiscoverFocusHosting maps a known Codex session to iTerm tabs via live PID → TTY.
// It does not call Find and never focuses. Session existence is the caller's job.
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
	fmt.Fprintf(&b, "\nSpecify one with:\n  agent-pro codex session focus %s --index <%s>", sessionID, validFocusIndexes(candidates))
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
