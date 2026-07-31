package agentrunapi

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
)

// ProcRow is one process snapshot row for tree TTY collection.
type ProcRow struct {
	PID  int
	PPID int
	TTY  string // e.g. /dev/ttys148, "??", or ""
	Cmd  string
}

// FocusCandidate is one iTerm match for a session after tree TTY resolution.
type FocusCandidate struct {
	Index   int // 0-based list index (CLI --index uses the same)
	Ref     iterm2.SessionRef
	PID     int
	TTY     string
	Source  string // e.g. "tree"
	CmdHint string
}

// FocusOpts drives FindITermForSession / FocusSession.
type FocusOpts struct {
	Store     agentstorage.Store
	SessionID string
	Index     *int // nil = require unique match
	DryRun    bool

	// Injectables (nil => production implementations):
	ListProcs  func() []ProcRow
	ListITerm  func() ([]iterm2.SessionRef, error)
	FocusITerm func(iterm2.SessionRef) error
}

// CollectTTYsFromTree returns real TTYs from ancestors + descendants of rootPID.
// Skips "??", blank, and whitespace-only TTY fields. Order is ancestors (root
// toward parents) then descendants (BFS); each real TTY appears at most once.
func CollectTTYsFromTree(procs []ProcRow, rootPID int) []string {
	if rootPID <= 0 || len(procs) == 0 {
		return nil
	}
	byPID := make(map[int]ProcRow, len(procs))
	children := make(map[int][]int, len(procs))
	for _, p := range procs {
		byPID[p.PID] = p
		if p.PPID != 0 || p.PID != p.PPID {
			children[p.PPID] = append(children[p.PPID], p.PID)
		}
	}
	if _, ok := byPID[rootPID]; !ok {
		return nil
	}

	seenPID := map[int]bool{rootPID: true}
	var ordered []ProcRow

	// Ancestors: walk PPID chain from root toward parents (excluding root itself first
	// for ancestor pass; root is added with descendants/root collection).
	// Include root + ancestors + descendants.
	// Ancestor walk (parents of root).
	cur := rootPID
	var ancestors []ProcRow
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

	// Root
	ordered = append(ordered, byPID[rootPID])
	// Ancestors after root for walk order (parent chain from nearest to farthest)
	ordered = append(ordered, ancestors...)

	// Descendants BFS
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
		// Use normalized form for dedup key but keep original-ish path.
		norm := iterm2.NormalizeTTY(tty)
		if norm == "" || seenTTY[norm] {
			continue
		}
		seenTTY[norm] = true
		out = append(out, norm)
	}
	return out
}

// FindITermForSession resolves session -> TTYs -> iTerm refs; does not focus.
func FindITermForSession(opts FocusOpts) ([]FocusCandidate, error) {
	sessionID := strings.TrimSpace(opts.SessionID)
	if sessionID == "" {
		return nil, fmt.Errorf("session_id is required")
	}
	if opts.Store == nil {
		return nil, fmt.Errorf("store is required")
	}

	sess, err := opts.Store.GetSession(sessionID)
	if err != nil {
		return nil, err
	}
	meta := sess.Meta
	runner := strings.TrimSpace(meta.Runner)
	if runner == "" {
		runner = "grok-tty"
	}
	termID := strings.TrimSpace(meta.TerminalSessionID)
	if termID == "" {
		termID = sessionID
	}

	rootPID, err := readRegistryPID(opts.Store.Home(), runner, termID)
	if err != nil {
		return nil, err
	}
	if rootPID <= 0 {
		return nil, fmt.Errorf("no candidates: registry pid missing for terminal session %s", termID)
	}

	listProcs := opts.ListProcs
	if listProcs == nil {
		listProcs = listLiveProcRows
	}
	procs := listProcs()
	ttys := CollectTTYsFromTree(procs, rootPID)

	// Map normalized TTY -> process row for CmdHint/PID.
	ttyOwner := map[string]ProcRow{}
	for _, p := range procs {
		n := iterm2.NormalizeTTY(strings.TrimSpace(p.TTY))
		if n == "" || strings.TrimSpace(p.TTY) == "??" {
			continue
		}
		if _, ok := ttyOwner[n]; !ok {
			ttyOwner[n] = p
		}
	}

	listITerm := opts.ListITerm
	if listITerm == nil {
		listITerm = listITermSessions
	}
	refs, err := listITerm()
	if err != nil {
		return nil, err
	}

	matched := iterm2.FindByTTY(refs, ttys)
	cands := make([]FocusCandidate, 0, len(matched))
	for i, ref := range matched {
		tty := iterm2.NormalizeTTY(ref.TTY)
		owner, ok := ttyOwner[tty]
		cand := FocusCandidate{
			Index:  i,
			Ref:    ref,
			TTY:    tty,
			Source: "tree",
		}
		if ok {
			cand.PID = owner.PID
			cand.CmdHint = owner.Cmd
		}
		cands = append(cands, cand)
	}
	return cands, nil
}

// FocusSession finds candidates, applies 0/1/multi policy, focuses unless DryRun.
func FocusSession(opts FocusOpts) (FocusCandidate, error) {
	cands, err := FindITermForSession(opts)
	if err != nil {
		return FocusCandidate{}, err
	}
	var chosen FocusCandidate
	switch len(cands) {
	case 0:
		return FocusCandidate{}, fmt.Errorf("no iTerm candidates found for session %s", strings.TrimSpace(opts.SessionID))
	case 1:
		if opts.Index != nil {
			idx := *opts.Index
			if idx != 0 {
				return FocusCandidate{}, fmt.Errorf("index %d out of range (0 candidates index 0 only)", idx)
			}
		}
		chosen = cands[0]
	default:
		if opts.Index == nil {
			var b strings.Builder
			fmt.Fprintf(&b, "multiple iTerm candidates (%d); specify --index N:\n", len(cands))
			for _, c := range cands {
				fmt.Fprintf(&b, "  --index %d  window=%s tab=%d tty=%s name=%s\n",
					c.Index, c.Ref.WindowID, c.Ref.TabIndex, c.TTY, c.Ref.Name)
			}
			return FocusCandidate{}, fmt.Errorf("%s", strings.TrimSuffix(b.String(), "\n"))
		}
		idx := *opts.Index
		if idx < 0 || idx >= len(cands) {
			return FocusCandidate{}, fmt.Errorf("index %d out of range (0..%d)", idx, len(cands)-1)
		}
		chosen = cands[idx]
	}

	if opts.DryRun {
		return chosen, nil
	}
	focusFn := opts.FocusITerm
	if focusFn == nil {
		focusFn = func(ref iterm2.SessionRef) error {
			return iterm2.Focus(ref, nil)
		}
	}
	if err := focusFn(chosen.Ref); err != nil {
		return FocusCandidate{}, err
	}
	return chosen, nil
}

// readRegistryPID loads the serve PID from {home}/{runner}-registry/{termID}.json.
func readRegistryPID(home, runner, termID string) (int, error) {
	path := filepath.Join(home, runner+"-registry", termID+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, fmt.Errorf("registry entry not found: %s", path)
		}
		return 0, err
	}
	var entry struct {
		PID int `json:"pid"`
	}
	if err := json.Unmarshal(data, &entry); err != nil {
		return 0, fmt.Errorf("parse registry %s: %w", path, err)
	}
	return entry.PID, nil
}

// listLiveProcRows is the production ListProcs: ps with pid/ppid/tty/command.
func listLiveProcRows() []ProcRow {
	// macOS/BSD: tty column is bare (ttys148) or ?? ; Linux similar with -o tty=.
	cmd := exec.Command("ps", "-ax", "-o", "pid=,ppid=,tty=,command=")
	out, err := cmd.Output()
	if err != nil {
		cmd = exec.Command("ps", "-x", "-o", "pid=,ppid=,tty=,command=")
		out, err = cmd.Output()
		if err != nil {
			return nil
		}
	}
	return parsePSProcRows(out)
}

func parsePSProcRows(out []byte) []ProcRow {
	var rows []ProcRow
	sc := bufio.NewScanner(bytes.NewReader(out))
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		// pid ppid tty command...
		if len(fields) < 3 {
			continue
		}
		pid, err1 := strconv.Atoi(fields[0])
		ppid, err2 := strconv.Atoi(fields[1])
		if err1 != nil || err2 != nil {
			continue
		}
		tty := fields[2]
		cmdStr := ""
		if len(fields) > 3 {
			cmdStr = strings.Join(fields[3:], " ")
		}
		rows = append(rows, ProcRow{PID: pid, PPID: ppid, TTY: tty, Cmd: cmdStr})
	}
	return rows
}

// listITermSessions is the production ListITerm via AppleScript session list.
func listITermSessions() ([]iterm2.SessionRef, error) {
	script := iterm2.BuildSessionListScript()
	cmd := exec.Command("osascript", "-e", script)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("list iTerm sessions: %w", err)
	}
	return iterm2.ParseSessionListOutput(string(out))
}
