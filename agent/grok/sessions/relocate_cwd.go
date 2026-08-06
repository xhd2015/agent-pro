package sessions

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// RelocateCWDOptions configures RelocateCWD.
type RelocateCWDOptions struct {
	// GrokHome is the Grok home directory. Empty → $GROK_HOME or ~/.grok.
	GrokHome string
}

// RelocateCWDResult reports paths after a successful relocate.
type RelocateCWDResult struct {
	OldCWD, NewCWD, OldSessionDir, NewSessionDir string
	FilesTouched                                 []string // optional
}

// RelocateCWD relocates a Grok session's workspace cwd by updating explicit JSON
// fields and moving the session directory under the URL-encoded new cwd key.
//
// Call shape: sessionID first (required), targetDir second (required), opts last (nil OK).
func RelocateCWD(sessionID, targetDir string, opts *RelocateCWDOptions) (*RelocateCWDResult, error) {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return nil, fmt.Errorf("session id is required")
	}

	grokHome, err := resolveRelocateGrokHome(opts)
	if err != nil {
		return nil, err
	}

	session, err := Find(grokHome, sessionID)
	if err != nil {
		return nil, err
	}
	oldSessionDir := filepath.Dir(session.Path)

	active, err := IsFileActive(grokHome, sessionID)
	if err != nil {
		return nil, err
	}
	if active {
		return nil, fmt.Errorf("session %s is active; cannot relocate", sessionID)
	}

	if strings.TrimSpace(targetDir) == "" {
		return nil, fmt.Errorf("target directory is required")
	}
	newCWD, err := filepath.Abs(targetDir)
	if err != nil {
		return nil, fmt.Errorf("resolve target directory: %w", err)
	}
	st, err := os.Stat(newCWD)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("target directory does not exist: %s", newCWD)
		}
		return nil, fmt.Errorf("stat target directory: %w", err)
	}
	if !st.IsDir() {
		return nil, fmt.Errorf("target is not a directory: %s", newCWD)
	}

	oldCWD := strings.TrimSpace(session.CWD)
	if oldCWD == "" {
		// Fallback: decode parent directory key under sessions/.
		if decoded, ok := decodeSessionParentCWD(oldSessionDir); ok {
			oldCWD = decoded
		} else {
			return nil, fmt.Errorf("session %s has empty cwd", sessionID)
		}
	}
	if abs, err := filepath.Abs(oldCWD); err == nil {
		oldCWD = abs
	}

	newSessionDir := filepath.Join(grokHome, "sessions", encodeSessionCWDKey(newCWD), sessionID)
	if fileExists(newSessionDir) && filepath.Clean(newSessionDir) != filepath.Clean(oldSessionDir) {
		return nil, fmt.Errorf("destination session path already exists: %s", newSessionDir)
	}

	// Already at target (same canonical cwd and already under encoded new key).
	if filepath.Clean(oldCWD) == filepath.Clean(newCWD) &&
		filepath.Clean(oldSessionDir) == filepath.Clean(newSessionDir) {
		return &RelocateCWDResult{
			OldCWD:         oldCWD,
			NewCWD:         newCWD,
			OldSessionDir:  oldSessionDir,
			NewSessionDir:  newSessionDir,
		}, nil
	}

	summaryPath := filepath.Join(oldSessionDir, "summary.json")
	if err := updateSummaryCWD(summaryPath, oldCWD, newCWD); err != nil {
		return nil, err
	}

	promptPath := filepath.Join(oldSessionDir, "prompt_context.json")
	promptUpdated := false
	if fileExists(promptPath) {
		if err := updatePromptWorkingDirectory(promptPath, newCWD); err != nil {
			return nil, err
		}
		promptUpdated = true
	}

	if err := os.MkdirAll(filepath.Dir(newSessionDir), 0o755); err != nil {
		return nil, fmt.Errorf("create destination parent: %w", err)
	}
	if err := os.Rename(oldSessionDir, newSessionDir); err != nil {
		return nil, fmt.Errorf("move session directory: %w", err)
	}

	var filesTouched []string
	filesTouched = append(filesTouched, filepath.Join(newSessionDir, "summary.json"))
	if promptUpdated {
		filesTouched = append(filesTouched, filepath.Join(newSessionDir, "prompt_context.json"))
	}

	return &RelocateCWDResult{
		OldCWD:         oldCWD,
		NewCWD:         newCWD,
		OldSessionDir:  oldSessionDir,
		NewSessionDir:  newSessionDir,
		FilesTouched:   filesTouched,
	}, nil
}

func resolveRelocateGrokHome(opts *RelocateCWDOptions) (string, error) {
	if opts != nil {
		if home := strings.TrimSpace(opts.GrokHome); home != "" {
			return filepath.Abs(home)
		}
	}
	if home := strings.TrimSpace(os.Getenv("GROK_HOME")); home != "" {
		return filepath.Abs(home)
	}
	userHome, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve user home: %w", err)
	}
	return filepath.Join(userHome, ".grok"), nil
}

// encodeSessionCWDKey matches Grok layout: url.PathEscape of an absolute cwd
// (e.g. /tmp/ws → %2Ftmp%2Fws).
func encodeSessionCWDKey(cwd string) string {
	return url.PathEscape(cwd)
}

func decodeSessionParentCWD(sessionDir string) (string, bool) {
	parent := filepath.Base(filepath.Dir(sessionDir))
	if parent == "" || parent == "." || parent == "sessions" {
		return "", false
	}
	decoded, err := url.PathUnescape(parent)
	if err != nil || decoded == "" {
		return "", false
	}
	return decoded, true
}

// isSessionActive reports whether sessionID is listed in active_sessions.json.
//
// Accepts object form {"sessions":[{sessionId|session_id:...},...]} and a bare
// JSON array of the same entry objects. Missing file, {}, or empty sessions → inactive.
func isSessionActive(grokHome, sessionID string) (bool, error) {
	path := filepath.Join(grokHome, "active_sessions.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	data = []byte(strings.TrimSpace(string(data)))
	if len(data) == 0 {
		return false, nil
	}

	sessionID = strings.TrimSpace(sessionID)

	// Object form: {"sessions":[...]}
	var obj struct {
		Sessions []json.RawMessage `json:"sessions"`
	}
	if err := json.Unmarshal(data, &obj); err == nil {
		// Distinguish bare arrays (which also unmarshal into empty objects in some cases)
		// by checking the first non-space byte.
		if len(data) > 0 && data[0] == '{' {
			for _, raw := range obj.Sessions {
				if entrySessionID(raw) == sessionID {
					return true, nil
				}
			}
			return false, nil
		}
	}

	// Bare array form: [{sessionId:...},...]
	var arr []json.RawMessage
	if err := json.Unmarshal(data, &arr); err == nil {
		for _, raw := range arr {
			if entrySessionID(raw) == sessionID {
				return true, nil
			}
		}
		return false, nil
	}

	// Unknown shape: treat as inactive (do not block relocate).
	return false, nil
}

func entrySessionID(raw json.RawMessage) string {
	var e struct {
		SessionID  string `json:"sessionId"`
		Session_ID string `json:"session_id"`
	}
	if err := json.Unmarshal(raw, &e); err != nil {
		return ""
	}
	id := strings.TrimSpace(e.SessionID)
	if id == "" {
		id = strings.TrimSpace(e.Session_ID)
	}
	return id
}

func updateSummaryCWD(path, oldCWD, newCWD string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read summary: %w", err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return fmt.Errorf("parse summary: %w", err)
	}

	info, ok := m["info"].(map[string]any)
	if !ok || info == nil {
		info = map[string]any{}
		m["info"] = info
	}
	info["cwd"] = newCWD

	if gitRoot, ok := m["git_root_dir"].(string); ok {
		if gitRoot == oldCWD || filepath.Clean(gitRoot) == filepath.Clean(oldCWD) {
			m["git_root_dir"] = newCWD
		}
	}

	return writeJSONFile(path, m)
}

func updatePromptWorkingDirectory(path, newCWD string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read prompt_context: %w", err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return fmt.Errorf("parse prompt_context: %w", err)
	}
	m["working_directory"] = newCWD
	return writeJSONFile(path, m)
}

func writeJSONFile(path string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	if err := os.WriteFile(path, b, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}
