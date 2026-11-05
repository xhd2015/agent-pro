package agentstorage

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/agent-pro/agent/codex/adhoc"
)

// LookupCodexConnection distinguishes a missing session from an existing native
// session. Both are nil connections, but only the former may adopt a new target.
func LookupCodexConnection(home, id string) (*adhoc.Connection, bool, error) {
	if home == "" || id == "" || id == "." || id == ".." || strings.ContainsAny(id, "/\\\x00") {
		return nil, false, errors.New("invalid Codex connection session reference")
	}
	raw, err := os.ReadFile(filepath.Join(home, "sessions", id, "meta.json"))
	if os.IsNotExist(err) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, errors.New("cannot read bound Codex session")
	}
	var meta SessionMeta
	if json.Unmarshal(raw, &meta) != nil {
		return nil, false, errors.New("invalid bound Codex session metadata")
	}
	if meta.CodexConnection == nil {
		return nil, true, nil
	}
	if err := meta.CodexConnection.Validate(); err != nil {
		return nil, true, err
	}
	return meta.CodexConnection.Clone(), true, nil
}

// ReadCodexConnection is a read-only final-spawn lookup that fails closed when
// the binding is missing; it never substitutes the current user preferences.
func ReadCodexConnection(home, id string) (*adhoc.Connection, error) {
	c, found, err := LookupCodexConnection(home, id)
	if err != nil {
		return nil, err
	}
	if !found || c == nil {
		return nil, errors.New("Codex session has no ad hoc connection binding")
	}
	return c, nil
}
