package agentstorage

import "strings"

// Provider family keys stored in SessionMeta.RunnerSessions.
const (
	RunnerFamilyCodex = "codex"
	RunnerFamilyGrok  = "grok"
)

// RunnerFamily returns the provider family for a runner name.
// Codex aliases (codex, codex-tty) → "codex"; grok aliases → "grok";
// anything else → the trimmed runner string.
func RunnerFamily(runner string) string {
	r := strings.TrimSpace(runner)
	if r == "" {
		return ""
	}
	if IsCodexRunner(r) {
		return RunnerFamilyCodex
	}
	if IsGrokRunner(r) {
		return RunnerFamilyGrok
	}
	return r
}

// SameRunnerFamily reports whether a and b are the same provider family.
// Empty names are not the same family as anything (including each other).
func SameRunnerFamily(a, b string) bool {
	fa := RunnerFamily(a)
	fb := RunnerFamily(b)
	return fa != "" && fa == fb
}

// HydrateRunnerSessions fills RunnerSessions[family(runner)] from the live
// RunnerSessionID when that map slot is empty. In-memory only; does not persist.
func (m *SessionMeta) HydrateRunnerSessions() {
	if m == nil {
		return
	}
	id := strings.TrimSpace(m.RunnerSessionID)
	if id == "" {
		return
	}
	fam := RunnerFamily(m.Runner)
	if fam == "" {
		return
	}
	if strings.TrimSpace(m.RunnerSessions[fam]) != "" {
		return
	}
	if m.RunnerSessions == nil {
		m.RunnerSessions = map[string]string{}
	}
	m.RunnerSessions[fam] = id
}

// SetRunnerSessionBind sets the live runner_session_id and the map slot for
// the current runner's family. Empty id only updates the live slot.
func (m *SessionMeta) SetRunnerSessionBind(id string) {
	if m == nil {
		return
	}
	id = strings.TrimSpace(id)
	m.RunnerSessionID = id
	if id == "" {
		return
	}
	fam := RunnerFamily(m.Runner)
	if fam == "" {
		return
	}
	if m.RunnerSessions == nil {
		m.RunnerSessions = map[string]string{}
	}
	m.RunnerSessions[fam] = id
}

// ClearCurrentFamilyBind clears the live slot and RunnerSessions[family(runner)].
// Other families are left intact.
func (m *SessionMeta) ClearCurrentFamilyBind() {
	if m == nil {
		return
	}
	m.RunnerSessionID = ""
	fam := RunnerFamily(m.Runner)
	if fam == "" || m.RunnerSessions == nil {
		return
	}
	delete(m.RunnerSessions, fam)
	if len(m.RunnerSessions) == 0 {
		m.RunnerSessions = nil
	}
}

// RunnerSessionIDForFamily returns the bound provider id for family, using the
// map first and the live slot when the current runner is that family.
func (m SessionMeta) RunnerSessionIDForFamily(family string) string {
	family = strings.TrimSpace(family)
	if family == "" {
		return ""
	}
	if m.RunnerSessions != nil {
		if id := strings.TrimSpace(m.RunnerSessions[family]); id != "" {
			return id
		}
	}
	if RunnerFamily(m.Runner) == family {
		return strings.TrimSpace(m.RunnerSessionID)
	}
	return ""
}
