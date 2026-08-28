package sessions

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/pathfmt"
)

// StatusCommandHelpLine is the parent `agent-pro codex session` help row.
const StatusCommandHelpLine = `  status <session-id>   show PID liveness + rollout path (File always no)`

// SessionStatus is liveness for one Codex session.
// FileActive is always false: Codex has no active_sessions.json equivalent.
// State is "running" | "inactive" (never "marked-active" without a file signal).
type SessionStatus struct {
	SessionID  string
	Path       string // absolute rollout path from Find (display via TildeHome)
	FileActive bool
	PIDs       []LivePID
	PIDChecked bool
	State      string // "running" | "inactive"
}

// Status returns liveness for a session.
// Finds the session first (unknown id → error). When checkPID, scans live PIDs.
// FileActive is always false.
func Status(codexHome, sessionID string, checkPID bool, live *LiveOptions) (*SessionStatus, error) {
	sessionID = strings.TrimSpace(sessionID)
	path, err := Find(codexHome, sessionID)
	if err != nil {
		return nil, err
	}

	st := &SessionStatus{
		SessionID:  sessionID,
		Path:       path,
		FileActive: false,
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

// FormatStatusText renders human-readable session status for CLI.
// Path is display-shortened with pathfmt.TildeHome (not for I/O).
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
	if p := strings.TrimSpace(st.Path); p != "" {
		fmt.Fprintf(&b, "Path: %s\n", pathfmt.TildeHome(p))
	}
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
// path is the absolute rollout path (no tilde).
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
		Path       string    `json:"path"`
		PIDs       []pidJSON `json:"pids"`
	}{
		SessionID:  st.SessionID,
		State:      st.State,
		FileActive: st.FileActive,
		PIDChecked: st.PIDChecked,
		Path:       st.Path,
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

// FormatActiveBlock renders the Active section for session info.
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
