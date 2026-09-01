package sessions

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/xhd2015/agent-pro/pkgs/procresolve"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2/tabselect"
)

// TabResolveResult is a Codex session resolved from an iTerm tab.
type TabResolveResult struct {
	SessionID    string
	RunnerPID    int
	Source       string
	Confidence   string
	WindowID     string
	WindowName   string
	TabIndex     int // 1-based iTerm index
	TabListPos   int // 0-based position in window tab list
	TTY          string
	ITermSession string
}

// TabResolveOpts injects probes for ResolveFromTab. Nil hooks use production.
type TabResolveOpts struct {
	ListProcs        func() []FocusProc
	Lsof             func(int) []string
	ListITerm        func() ([]iterm2.SessionRef, error)
	CurrentSessionID func() string
	ControllingTTY   func() string
	AncestorTTYs     func() []string
}

// ResolveFromTab finds the target tab in the current iTerm window and resolves
// exactly one Codex session id from codex runners on that tab's TTY(s).
// Multiple unrelated Codex sessions on the same tab refuse (fail closed).
func ResolveFromTab(sel tabselect.TabSelector, opts *TabResolveOpts) (*TabResolveResult, error) {
	if opts == nil {
		opts = &TabResolveOpts{}
	}
	listProcs := opts.ListProcs
	if listProcs == nil {
		listProcs = listLiveFocusProcs
	}
	lsof := opts.Lsof
	if lsof == nil {
		lsof = procresolve.LiveLsof
	}

	refs, cfg, err := loadTabITermRefs(opts)
	if err != nil {
		if errors.Is(err, iterm2.ErrNotInSession) {
			return nil, fmt.Errorf("iterm2: not inside an iTerm2 session (no ITERM_SESSION_ID and no matching TTY)")
		}
		return nil, err
	}

	cached := refs
	cfg.ListSessions = func() ([]iterm2.SessionRef, error) {
		return cached, nil
	}
	st, err := iterm2.CurrentWindowStatusWith(cfg)
	if err != nil {
		if errors.Is(err, iterm2.ErrNotInSession) {
			return nil, fmt.Errorf("iterm2: not inside an iTerm2 session (no ITERM_SESSION_ID and no matching TTY)")
		}
		return nil, err
	}
	if len(st.Tabs) == 0 {
		return nil, fmt.Errorf("no tabs in window %s", st.WindowID)
	}

	row, listPos, err := tabselect.SelectWindowTab(st, sel)
	if err != nil {
		return nil, err
	}

	ttys := tabTTYs(refs, st.WindowID, row.Index)
	if len(ttys) == 0 {
		tty := strings.TrimSpace(row.TTY)
		if tty != "" {
			ttys = []string{iterm2.NormalizeTTY(tty)}
		}
	}
	if len(ttys) == 0 {
		return nil, fmt.Errorf("no tty on tab %d", row.Index)
	}

	procs := listProcs()
	hits, err := codexSessionsOnTTYs(ttys, procs, lsof)
	if err != nil {
		return nil, err
	}
	if len(hits) == 0 {
		return nil, fmt.Errorf("no codex session on tab %d (tty %s)", row.Index, displayTabTTYs(ttys))
	}
	if len(hits) > 1 {
		return nil, formatMultiCodexTabError(row.Index, hits)
	}

	hit := hits[0]
	return &TabResolveResult{
		SessionID:    hit.SessionID,
		RunnerPID:    hit.RunnerPID,
		Source:       hit.Source,
		Confidence:   "hard",
		WindowID:     st.WindowID,
		WindowName:   st.WindowName,
		TabIndex:     row.Index,
		TabListPos:   listPos,
		TTY:          hit.TTY,
		ITermSession: row.SessionID,
	}, nil
}

func loadTabITermRefs(opts *TabResolveOpts) ([]iterm2.SessionRef, *iterm2.CurrentStatusConfig, error) {
	cfg := &iterm2.CurrentStatusConfig{}
	if opts.CurrentSessionID != nil {
		cfg.SessionID = opts.CurrentSessionID
	}
	if opts.ControllingTTY != nil {
		cfg.ControllingTTY = opts.ControllingTTY
	}
	if opts.AncestorTTYs != nil {
		cfg.AncestorTTYs = opts.AncestorTTYs
	}

	if opts.ListITerm != nil {
		refs, err := opts.ListITerm()
		return refs, cfg, err
	}

	sid := ""
	if cfg.SessionID != nil {
		sid = strings.TrimSpace(cfg.SessionID())
	} else if v := strings.TrimSpace(os.Getenv("ITERM_SESSION_ID")); v != "" {
		sid = v
	}
	if sid != "" {
		refs, err := iterm2.ListSessionsInWindowByUUID(sid)
		if err == nil && len(iterm2.FindBySessionUUID(refs, sid)) > 0 {
			return refs, cfg, nil
		}
	}
	refs, err := iterm2.ListSessions()
	return refs, cfg, err
}

type tabCodexHit struct {
	SessionID string
	RunnerPID int
	Source    string
	TTY       string
	OpenPath  string
}


func tabTTYs(refs []iterm2.SessionRef, windowID string, tabIndex int) []string {
	seen := map[string]bool{}
	var out []string
	for _, ref := range refs {
		if ref.WindowID != windowID || ref.TabIndex != tabIndex {
			continue
		}
		n := iterm2.NormalizeTTY(ref.TTY)
		if n == "" || seen[n] {
			continue
		}
		seen[n] = true
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}

func codexSessionsOnTTYs(ttys []string, procs []FocusProc, lsof func(int) []string) ([]tabCodexHit, error) {
	want := map[string]bool{}
	for _, t := range ttys {
		n := iterm2.NormalizeTTY(t)
		if n != "" {
			want[n] = true
		}
	}
	bySID := map[string]tabCodexHit{}
	var order []string
	for _, p := range procs {
		if !procresolve.IsCodexRunner(p.Cmd) {
			continue
		}
		// Codex runners (often agent-run wrapped) may show TTY "??"; match if
		// any ancestor/descendant TTY hits the tab.
		treeTTYs := collectTTYsFromTree(procs, p.PID)
		matchedTTY := ""
		for _, tty := range treeTTYs {
			if want[tty] {
				matchedTTY = tty
				break
			}
		}
		if matchedTTY == "" {
			continue
		}
		files := lsof(p.PID)
		for _, f := range files {
			kind, sid, ok := procresolve.ParseSessionFromPath(f)
			if !ok || kind != "codex" {
				continue
			}
			sid = strings.TrimSpace(sid)
			if sid == "" {
				continue
			}
			if _, exists := bySID[sid]; exists {
				continue
			}
			bySID[sid] = tabCodexHit{
				SessionID: sid,
				RunnerPID: p.PID,
				Source:    "open-files",
				TTY:       matchedTTY,
				OpenPath:  f,
			}
			order = append(order, sid)
		}
	}
	out := make([]tabCodexHit, 0, len(order))
	for _, sid := range order {
		out = append(out, bySID[sid])
	}
	return out, nil
}

func displayTabTTYs(ttys []string) string {
	if len(ttys) == 0 {
		return "-"
	}
	return strings.Join(ttys, ", ")
}

func formatMultiCodexTabError(tabIndex int, hits []tabCodexHit) error {
	var b strings.Builder
	fmt.Fprintf(&b, "multiple codex sessions on tab %d\n", tabIndex)
	for i, h := range hits {
		fmt.Fprintf(&b, "  [%d] %s  pid %d  tty %s\n", i, h.SessionID, h.RunnerPID, h.TTY)
	}
	fmt.Fprintf(&b, "Refuse to guess; pass an explicit session id")
	return errors.New(strings.TrimRight(b.String(), "\n"))
}
