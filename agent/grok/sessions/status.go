package sessions

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/xhd2015/agent-pro/pkgs/procresolve"
)

// LivePID is one live process hard-hitting a session via open files.
type LivePID struct {
	PID  int
	Name string // filepath.Base(argv0)
	Cmd  string // full command line
}

// LiveOptions injects process listing and open-file probes for tests.
// When hooks are nil, production uses procresolve.ListLiveProcs / LiveLsof.
type LiveOptions struct {
	ListProcs func() []procresolve.Proc
	Lsof      func(int) []string
}

// SessionStatus is dual-signal liveness for one Grok session.
type SessionStatus struct {
	SessionID  string
	FileActive bool
	PIDs       []LivePID
	PIDChecked bool
	State      string // "running" | "marked-active" | "inactive"
}

// IsFileActive reports whether sessionID is listed in active_sessions.json.
// Accepts object form {"sessions":[{sessionId|session_id:...},...]} and a bare
// JSON array of the same entry objects. Missing file, {}, or empty sessions → inactive.
func IsFileActive(grokHome, sessionID string) (bool, error) {
	return isSessionActive(grokHome, sessionID)
}

// LivePIDsForSession scans ListProcs for grok runners only (basename grok,
// excluding pure `grok update`). For each runner, Lsof open paths are checked
// for a hard hit on sessionID (same path rules as procresolve).
// Returns all matches sorted by PID ascending.
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
		if !procresolve.IsGrokRunner(p.Cmd) {
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

// Status returns dual-signal status for a session.
// Finds the session first (unknown id → error). When checkPID, scans live PIDs.
func Status(grokHome, sessionID string, checkPID bool, live *LiveOptions) (*SessionStatus, error) {
	sessionID = strings.TrimSpace(sessionID)
	if _, err := Find(grokHome, sessionID); err != nil {
		return nil, err
	}

	fileActive, err := IsFileActive(grokHome, sessionID)
	if err != nil {
		return nil, err
	}

	st := &SessionStatus{
		SessionID:  sessionID,
		FileActive: fileActive,
		PIDChecked: checkPID,
	}

	if checkPID {
		pids, err := LivePIDsForSession(sessionID, live)
		if err != nil {
			return nil, err
		}
		st.PIDs = pids
	}

	st.State = rollupState(st)
	return st, nil
}

func rollupState(st *SessionStatus) string {
	if st == nil {
		return "inactive"
	}
	if st.PIDChecked && len(st.PIDs) > 0 {
		return "running"
	}
	if st.FileActive {
		return "marked-active"
	}
	return "inactive"
}

func openFilesHitSession(paths []string, sessionID string) bool {
	want := strings.ToLower(strings.TrimSpace(sessionID))
	if want == "" {
		return false
	}
	for _, p := range paths {
		kind, id, ok := procresolve.ParseSessionFromPath(p)
		if !ok || kind != "grok" {
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

// FormatStatusText renders human-readable session status for CLI.
func FormatStatusText(st *SessionStatus) string {
	if st == nil {
		return ""
	}
	var b strings.Builder
	fmt.Fprintf(&b, "State: %s\n", st.State)
	fileWord := "no"
	if st.FileActive {
		fileWord = "yes"
	}
	fmt.Fprintf(&b, "File: %s\n", fileWord)
	if !st.PIDChecked {
		fmt.Fprintf(&b, "PIDs: skipped (--no-pid)\n")
	} else if len(st.PIDs) == 0 {
		fmt.Fprintf(&b, "PIDs: none\n")
	} else {
		fmt.Fprintf(&b, "PIDs:\n")
		for _, p := range st.PIDs {
			fmt.Fprintf(&b, "  %d %s\n", p.PID, p.Name)
		}
	}
	return strings.TrimRight(b.String(), "\n")
}

// FormatStatusJSON renders status as a JSON object (no ANSI).
func FormatStatusJSON(st *SessionStatus) (string, error) {
	if st == nil {
		return "", fmt.Errorf("nil SessionStatus")
	}
	type pidJSON struct {
		PID  int    `json:"pid"`
		Name string `json:"name"`
		Cmd  string `json:"cmd"`
	}
	doc := struct {
		SessionID  string    `json:"session_id"`
		State      string    `json:"state"`
		FileActive bool      `json:"file_active"`
		PIDChecked bool      `json:"pid_checked"`
		PIDs       []pidJSON `json:"pids"`
	}{
		SessionID:  st.SessionID,
		State:      st.State,
		FileActive: st.FileActive,
		PIDChecked: st.PIDChecked,
		PIDs:       make([]pidJSON, 0, len(st.PIDs)),
	}
	for _, p := range st.PIDs {
		doc.PIDs = append(doc.PIDs, pidJSON{PID: p.PID, Name: p.Name, Cmd: p.Cmd})
	}
	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// FormatActiveBlock renders the dual-signal Active section for session info.
func FormatActiveBlock(st *SessionStatus) string {
	if st == nil {
		return ""
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Active:\n")
	fileWord := "no"
	if st.FileActive {
		fileWord = "yes (file active)"
	}
	fmt.Fprintf(&b, "  File: %s\n", fileWord)
	if !st.PIDChecked {
		fmt.Fprintf(&b, "  PIDs: skipped (--no-pid)\n")
	} else if len(st.PIDs) == 0 {
		fmt.Fprintf(&b, "  PIDs: none\n")
	} else {
		fmt.Fprintf(&b, "  PIDs:\n")
		for _, p := range st.PIDs {
			fmt.Fprintf(&b, "    %d %s\n", p.PID, p.Name)
		}
	}
	return strings.TrimRight(b.String(), "\n")
}
