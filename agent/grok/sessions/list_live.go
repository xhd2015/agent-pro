package sessions

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/xhd2015/agent-pro/pkgs/itermsnapshot"
	"github.com/xhd2015/agent-pro/pkgs/procresolve"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2/snapshot"
)

// ListLiveHelp is the text for `agent-pro grok session list-live --help`.
const ListLiveHelp = `Usage: agent-pro grok session list-live [OPTIONS]

List Grok session ids that currently have a hosting iTerm2 tab.
Discovery: live grok processes (open-files hard hit) → TTY → iTerm match.
Sessions with a live PID but no iTerm tab are omitted.
One row per session id; multiple tabs show as w=… t=…(+N).

Options:
  --json        machine-readable JSON (no ANSI)
  --limit N     show at most N sessions (0 = unlimited)
  -h,--help     show help
`

// ListLiveCommandHelpLine is the parent `agent-pro grok session` help row.
const ListLiveCommandHelpLine = `  list-live …            list Grok ids hosted in iTerm tabs`

// LiveHostingRow is one Grok session currently hosted in iTerm.
type LiveHostingRow struct {
	SessionID  string
	ITerm      string // w=<id> t=<tab>(+N)
	Title      string // from summary.json generated_title; empty → "-"
	Workspace  string
	Candidates []FocusCandidate
}

// ListLiveOpts drives ListLive / RunListLive. Nil hooks use production probes.
type ListLiveOpts struct {
	JSON  bool
	Limit int // 0 = unlimited

	ListProcs func() []FocusProc
	Lsof      func(int) []string
	ListITerm func() ([]iterm2.SessionRef, error)

	// PaneByTTY supplies optional pane cwd for WORKSPACE (nil → path-derived summary).
	PaneByTTY func() (map[string]LivePaneInfo, error)

	// CaptureInventory, when both ListITerm and PaneByTTY are nil, runs once
	// and feeds hosting refs + pane cwd. Nil → fast path: one ListITerm
	// AppleScript and empty panes (title/cwd from open-file summary / FindSession).
	CaptureInventory func() (panes map[string]LivePaneInfo, refs []iterm2.SessionRef, err error)

	// FindSession resolves workspace when pane cwd and path-derived summary cwd are empty.
	// Title/cwd normally come from summary.json beside the lsof hard-hit session dir.
	FindSession func(sessionID string) (cwd string, ok bool)
	GrokHome    string
}

// LivePaneInfo is optional iTerm pane metadata keyed by normalized TTY.
type LivePaneInfo struct {
	Idle *bool
	Cwd  string
}

// ListLiveFake is the deterministic injected boundary for list-live tests.
type ListLiveFake struct {
	FocusFake
	PaneByTTY map[string]LivePaneInfo
	// CwdBySession overrides FindSession when set.
	CwdBySession map[string]string
}

// ListLiveOpts returns ListLiveOpts wired to this fake.
func (f *ListLiveFake) ListLiveOpts() *ListLiveOpts {
	fo := f.FocusFake.Opts()
	return &ListLiveOpts{
		ListProcs: fo.ListProcs,
		Lsof:      fo.Lsof,
		ListITerm: fo.ListITerm,
		PaneByTTY: func() (map[string]LivePaneInfo, error) {
			if f.PaneByTTY == nil {
				return map[string]LivePaneInfo{}, nil
			}
			out := make(map[string]LivePaneInfo, len(f.PaneByTTY))
			for k, v := range f.PaneByTTY {
				out[k] = v
			}
			return out, nil
		},
		FindSession: func(sessionID string) (string, bool) {
			if f.CwdBySession == nil {
				return "", false
			}
			cwd, ok := f.CwdBySession[sessionID]
			return cwd, ok
		},
	}
}

// ListLive returns Grok sessions that have at least one hosting iTerm tab.
// Order is sorted by session id. Rows without a resolved sid never appear.
//
// External probes (ps / lsof / iTerm session list) run once per ListLive call
// and are reused across sid discovery and every DiscoverFocusHosting join.
// ListITerm is prefetched in parallel with sid discovery. TITLE/WORKSPACE come
// from summary.json beside the lsof hard-hit session dir (no WalkDir index).
func ListLive(opts *ListLiveOpts) ([]LiveHostingRow, error) {
	if opts == nil {
		opts = &ListLiveOpts{}
	}
	opts = withSharedListLiveProbes(opts)
	focusOpts := &FocusOpts{
		ListProcs: opts.ListProcs,
		Lsof:      opts.Lsof,
		ListITerm: opts.ListITerm,
	}

	// Overlap AppleScript session-list with lsof sid discovery.
	if opts.ListITerm != nil {
		go func() { _, _ = opts.ListITerm() }()
	}

	sids, metaIndex, err := discoverLiveGrokSessions(opts)
	if err != nil {
		return nil, err
	}

	paneByTTY, _ := loadPaneByTTY(opts)

	rows := make([]LiveHostingRow, 0, len(sids))
	for _, sid := range sids {
		disc, err := DiscoverFocusHosting(sid, focusOpts)
		if err != nil {
			return nil, err
		}
		if disc == nil || len(disc.Candidates) == 0 {
			continue
		}
		cands := disc.Candidates
		row := LiveHostingRow{
			SessionID:  sid,
			ITerm:      formatLiveHostingITerm(cands),
			Candidates: cands,
		}
		primary := cands[0]
		tty := iterm2.NormalizeTTY(primary.TTY)
		if info, ok := paneByTTY[tty]; ok {
			row.Workspace = strings.TrimSpace(info.Cwd)
		}
		if meta, ok := metaIndex[sid]; ok {
			if row.Title == "" {
				row.Title = strings.TrimSpace(meta.Title)
			}
			if row.Workspace == "" {
				row.Workspace = strings.TrimSpace(meta.Cwd)
			}
		}
		if row.Workspace == "" && opts.FindSession != nil {
			if cwd, ok := opts.FindSession(sid); ok {
				row.Workspace = cwd
			}
		}
		rows = append(rows, row)
	}

	sort.Slice(rows, func(i, j int) bool {
		return rows[i].SessionID < rows[j].SessionID
	})
	if opts.Limit > 0 && len(rows) > opts.Limit {
		rows = rows[:opts.Limit]
	}
	return rows, nil
}

type liveSessionMeta struct {
	Title string
	Cwd   string
}

// RunListLive implements `agent-pro grok session list-live`.
func RunListLive(args []string, stdout, stderr io.Writer, grokHome string, opts *ListLiveOpts) error {
	if stdout == nil {
		stdout = io.Discard
	}
	if stderr == nil {
		stderr = io.Discard
	}
	if opts == nil {
		opts = &ListLiveOpts{}
	}
	if strings.TrimSpace(opts.GrokHome) == "" {
		opts.GrokHome = grokHome
	}

	jsonOut, limit, help, err := parseListLiveArgs(args)
	if err != nil {
		return err
	}
	if help {
		fmt.Fprint(stdout, ListLiveHelp)
		if !strings.HasSuffix(ListLiveHelp, "\n") {
			fmt.Fprint(stdout, "\n")
		}
		return nil
	}
	opts.JSON = jsonOut
	opts.Limit = limit

	rows, err := ListLive(opts)
	if err != nil {
		return err
	}
	if opts.JSON {
		return writeListLiveJSON(stdout, rows)
	}
	out := FormatListLiveTable(rows)
	fmt.Fprint(stdout, out)
	if out != "" && !strings.HasSuffix(out, "\n") {
		fmt.Fprint(stdout, "\n")
	}
	return nil
}

func parseListLiveArgs(args []string) (jsonOut bool, limit int, help bool, err error) {
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch a {
		case "-h", "--help":
			help = true
		case "--json":
			jsonOut = true
		case "--limit":
			if i+1 >= len(args) {
				return false, 0, false, fmt.Errorf("--limit requires a value")
			}
			i++
			n, perr := parsePositiveInt(args[i])
			if perr != nil {
				return false, 0, false, fmt.Errorf("--limit: %w", perr)
			}
			limit = n
		default:
			if strings.HasPrefix(a, "--limit=") {
				n, perr := parsePositiveInt(strings.TrimPrefix(a, "--limit="))
				if perr != nil {
					return false, 0, false, fmt.Errorf("--limit: %w", perr)
				}
				limit = n
				continue
			}
			return false, 0, false, fmt.Errorf("unexpected argument: %s", a)
		}
	}
	return jsonOut, limit, help, nil
}

func parsePositiveInt(s string) (int, error) {
	var n int
	if _, err := fmt.Sscanf(strings.TrimSpace(s), "%d", &n); err != nil {
		return 0, fmt.Errorf("invalid integer %q", s)
	}
	if n < 0 {
		return 0, fmt.Errorf("must be >= 0")
	}
	return n, nil
}

// withSharedListLiveProbes returns a shallow copy of opts whose ListProcs,
// Lsof, and iTerm inventory hooks share one snapshot for the duration of ListLive.
// Nil hooks resolve to production probes before sharing.
//
// When both ListITerm and PaneByTTY are nil:
//   - CaptureInventory set → one load feeds hosting + panes
//   - else → one ListITerm AppleScript; panes empty (skip Capture)
func withSharedListLiveProbes(opts *ListLiveOpts) *ListLiveOpts {
	if opts == nil {
		opts = &ListLiveOpts{}
	}
	out := *opts

	listProcs := opts.ListProcs
	if listProcs == nil {
		listProcs = listLiveFocusProcs
	}
	procs := listProcs()
	out.ListProcs = func() []FocusProc {
		return procs
	}

	lsof := opts.Lsof
	productionLsof := lsof == nil
	if productionLsof {
		lsof = procresolve.LiveLsof
	}
	lsofCache := map[int][]string{}
	if productionLsof {
		// One bulk lsof for all grok runners; seed the per-pid cache.
		var pids []int
		for _, p := range procs {
			if procresolve.IsGrokRunner(p.Cmd) {
				pids = append(pids, p.PID)
			}
		}
		for pid, paths := range procresolve.LiveLsofMany(pids) {
			lsofCache[pid] = paths
		}
	}
	out.Lsof = func(pid int) []string {
		if paths, ok := lsofCache[pid]; ok {
			return paths
		}
		paths := lsof(pid)
		lsofCache[pid] = paths
		return paths
	}

	if opts.ListITerm == nil && opts.PaneByTTY == nil {
		if opts.CaptureInventory != nil {
			bindUnifiedITermInventory(&out, opts.CaptureInventory)
			return &out
		}
		// Fast production path: skip full-window session dump + Capture enrich.
		// One TTY-targeted AppleScript for grok-host TTYs; title/cwd from open-file summary.
		ttys := grokHostingTTYs(procs)
		bindMemoListITerm(&out, func() ([]iterm2.SessionRef, error) {
			return iterm2.FindSessionsByTTY(ttys)
		})
		out.PaneByTTY = func() (map[string]LivePaneInfo, error) {
			return map[string]LivePaneInfo{}, nil
		}
		return &out
	}

	listITerm := opts.ListITerm
	if listITerm == nil {
		ttys := grokHostingTTYs(procs)
		listITerm = func() ([]iterm2.SessionRef, error) {
			return iterm2.FindSessionsByTTY(ttys)
		}
	}
	bindMemoListITerm(&out, listITerm)

	return &out
}

// grokHostingTTYs returns normalized TTYs for live grok runners (including
// ancestor/descendant TTYs), used for one FindSessionsByTTY inventory.
func grokHostingTTYs(procs []FocusProc) []string {
	seen := map[string]bool{}
	var ttys []string
	for _, p := range procs {
		if !procresolve.IsGrokRunner(p.Cmd) {
			continue
		}
		for _, tty := range collectTTYsFromTree(procs, p.PID) {
			if tty == "" || seen[tty] {
				continue
			}
			seen[tty] = true
			ttys = append(ttys, tty)
		}
	}
	return ttys
}

func bindMemoListITerm(out *ListLiveOpts, listITerm func() ([]iterm2.SessionRef, error)) {
	var (
		once sync.Once
		refs []iterm2.SessionRef
		err  error
	)
	out.ListITerm = func() ([]iterm2.SessionRef, error) {
		once.Do(func() {
			refs, err = listITerm()
		})
		return refs, err
	}
}

// bindUnifiedITermInventory wires ListITerm + PaneByTTY from one inventory load.
func bindUnifiedITermInventory(out *ListLiveOpts, capture func() (map[string]LivePaneInfo, []iterm2.SessionRef, error)) {
	var (
		once  sync.Once
		panes map[string]LivePaneInfo
		refs  []iterm2.SessionRef
		err   error
	)
	load := func() {
		once.Do(func() {
			panes, refs, err = capture()
			if panes == nil {
				panes = map[string]LivePaneInfo{}
			}
			// Hosting must not fail solely because inventory soft-missed: fall back
			// to the lighter session-list script when refs are empty.
			if len(refs) == 0 {
				refs, err = listLiveITermSessions()
			}
		})
	}
	out.ListITerm = func() ([]iterm2.SessionRef, error) {
		load()
		return refs, err
	}
	out.PaneByTTY = func() (map[string]LivePaneInfo, error) {
		load()
		// Pane enrich is soft: never fail the list for inventory errors.
		return panes, nil
	}
}

// sessionRefsFromSnapshot maps a Capture inventory to SessionRef rows used by
// DiscoverFocusHosting (same WindowID/TabIndex/TTY shape as ListITerm).
func sessionRefsFromSnapshot(snap *snapshot.Snapshot) []iterm2.SessionRef {
	if snap == nil {
		return nil
	}
	var out []iterm2.SessionRef
	for wi := range snap.Windows {
		win := &snap.Windows[wi]
		wid := ""
		if win.WindowID != 0 {
			wid = strconv.FormatUint(win.WindowID, 10)
		} else if win.Index != 0 {
			wid = strconv.Itoa(win.Index)
		}
		for ti := range win.Tabs {
			tab := &win.Tabs[ti]
			for si := range tab.Sessions {
				sess := &tab.Sessions[si]
				tabIndex := tab.Index
				if tabIndex == 0 && sess.TabIndex != 0 {
					tabIndex = sess.TabIndex
				}
				out = append(out, iterm2.SessionRef{
					WindowID:   wid,
					WindowName: win.Name,
					TabIndex:   tabIndex,
					SessionID:  sess.ID,
					TTY:        sess.TTY,
					Name:       sess.Name,
				})
			}
		}
	}
	return out
}

// discoverLiveGrokSessions finds live Grok sids via open-file hard hits and
// loads TITLE/CWD from summary.json beside each hit (remapped into GrokHome
// when set). No WalkDir over the sessions tree.
func discoverLiveGrokSessions(opts *ListLiveOpts) ([]string, map[string]liveSessionMeta, error) {
	listProcs := opts.ListProcs
	if listProcs == nil {
		listProcs = listLiveFocusProcs
	}
	lsof := opts.Lsof
	if lsof == nil {
		lsof = procresolve.LiveLsof
	}

	seen := map[string]struct{}{}
	meta := map[string]liveSessionMeta{}
	var ids []string
	for _, p := range listProcs() {
		if !procresolve.IsGrokRunner(p.Cmd) {
			continue
		}
		for _, path := range lsof(p.PID) {
			kind, sid, ok := procresolve.ParseSessionFromPath(path)
			if !ok || kind != "grok" {
				continue
			}
			sid = strings.ToLower(strings.TrimSpace(sid))
			if sid == "" {
				continue
			}
			if _, exists := seen[sid]; !exists {
				seen[sid] = struct{}{}
				ids = append(ids, sid)
			}
			if _, has := meta[sid]; !has {
				if m, ok := metaFromOpenPath(path, opts.GrokHome); ok {
					meta[sid] = m
				}
			}
		}
	}
	sort.Strings(ids)
	return ids, meta, nil
}

// metaFromOpenPath reads summary.json for the session dir implied by an open-file
// path. When grokHome is set, remaps …/.grok/sessions/<rel> → grokHome/sessions/<rel>
// so injected fixtures (fixture paths + temp GrokHome) still resolve.
func metaFromOpenPath(openPath, grokHome string) (liveSessionMeta, bool) {
	sessionDir, _, ok := procresolve.GrokSessionDirFromPath(openPath)
	if !ok || strings.TrimSpace(sessionDir) == "" {
		return liveSessionMeta{}, false
	}
	sessionDir = remapGrokSessionDir(sessionDir, grokHome)
	sumPath := filepath.Join(sessionDir, "summary.json")
	session, ok := parseSummaryFile(sumPath)
	if !ok {
		return liveSessionMeta{}, false
	}
	return liveSessionMeta{
		Title: strings.TrimSpace(session.Title),
		Cwd:   strings.TrimSpace(session.CWD),
	}, true
}

func remapGrokSessionDir(sessionDir, grokHome string) string {
	grokHome = strings.TrimSpace(grokHome)
	if grokHome == "" {
		return sessionDir
	}
	slash := filepath.ToSlash(sessionDir)
	const marker = "/.grok/sessions/"
	idx := strings.Index(slash, marker)
	if idx < 0 {
		return sessionDir
	}
	rel := slash[idx+len(marker):]
	return filepath.Join(grokHome, "sessions", filepath.FromSlash(rel))
}

func loadPaneByTTY(opts *ListLiveOpts) (map[string]LivePaneInfo, error) {
	if opts.PaneByTTY != nil {
		return opts.PaneByTTY()
	}
	res, _, err := itermsnapshot.Capture(itermsnapshot.CaptureOpts{NoEnrich: true})
	if err != nil || res == nil || res.Snapshot == nil {
		return map[string]LivePaneInfo{}, err
	}
	return paneInfoFromSnapshot(res.Snapshot), nil
}

func paneInfoFromSnapshot(snap *snapshot.Snapshot) map[string]LivePaneInfo {
	out := map[string]LivePaneInfo{}
	if snap == nil {
		return out
	}
	for wi := range snap.Windows {
		win := &snap.Windows[wi]
		for ti := range win.Tabs {
			tab := &win.Tabs[ti]
			for si := range tab.Sessions {
				sess := &tab.Sessions[si]
				tty := iterm2.NormalizeTTY(sess.TTY)
				if tty == "" {
					continue
				}
				info := LivePaneInfo{Idle: sess.Idle}
				if sess.Cwd != nil {
					info.Cwd = strings.TrimSpace(*sess.Cwd)
				}
				if _, exists := out[tty]; !exists {
					out[tty] = info
				}
			}
		}
	}
	return out
}

func formatLiveHostingITerm(cands []FocusCandidate) string {
	if len(cands) == 0 {
		return "-"
	}
	c := cands[0]
	wid := strings.TrimSpace(c.WindowID)
	if wid == "" {
		wid = "0"
	}
	base := fmt.Sprintf("w=%s t=%d", wid, c.TabIndex)
	if len(cands) > 1 {
		base = fmt.Sprintf("%s(+%d)", base, len(cands)-1)
	}
	return base
}

// FormatListLiveTable renders the human table + footer.
func FormatListLiveTable(rows []LiveHostingRow) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%-38s  %-14s  %-42s  %s\n", "SESSION_ID", "ITERM", "TITLE", "WORKSPACE")
	for _, r := range rows {
		title := strings.TrimSpace(r.Title)
		if title == "" {
			title = "-"
		} else {
			title = truncateTitle(title)
		}
		ws := r.Workspace
		if ws == "" {
			ws = "-"
		}
		fmt.Fprintf(&b, "%-38s  %-14s  %-42s  %s\n", r.SessionID, r.ITerm, title, ws)
	}
	fmt.Fprintf(&b, "\n%d sessions\n", len(rows))
	return b.String()
}

type listLiveJSONEnvelope struct {
	Sessions []listLiveJSONRow `json:"sessions"`
	Summary  listLiveJSONSum   `json:"summary"`
}

type listLiveJSONRow struct {
	SessionID string `json:"session_id"`
	ITerm     string `json:"iterm"`
	Title     string `json:"title"`
	Workspace string `json:"workspace"`
	Hosts     int    `json:"hosts"`
}

type listLiveJSONSum struct {
	Count int `json:"count"`
}

func writeListLiveJSON(w io.Writer, rows []LiveHostingRow) error {
	env := listLiveJSONEnvelope{
		Sessions: make([]listLiveJSONRow, 0, len(rows)),
		Summary:  listLiveJSONSum{Count: len(rows)},
	}
	for _, r := range rows {
		title := strings.TrimSpace(r.Title)
		if title == "" {
			title = "-"
		}
		ws := r.Workspace
		if ws == "" {
			ws = "-"
		}
		env.Sessions = append(env.Sessions, listLiveJSONRow{
			SessionID: r.SessionID,
			ITerm:     r.ITerm,
			Title:     title,
			Workspace: ws,
			Hosts:     len(r.Candidates),
		})
	}
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	return enc.Encode(env)
}
