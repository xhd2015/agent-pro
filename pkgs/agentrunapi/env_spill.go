package agentrunapi

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

// EnvFileSpillMinRunes is the exclusive lower bound for auto-spill of session
// env to --env-file: any single KEY=VALUE entry whose rune count is > this
// value triggers spilling the full env list. Locked: 64.
const EnvFileSpillMinRunes = 64

// EnvSpillOpts configures MaybeSpillEnv (shared by BuildFollowUpCommand).
type EnvSpillOpts struct {
	// SpillDir is where the file is written. Empty uses
	// filepath.Join(os.TempDir(), "agent-run-env-spill").
	SpillDir string
	// SessionID if non-empty is sanitized into the filename (env-<id>-*.env).
	SessionID string
	// Force writes even when no entry exceeds EnvFileSpillMinRunes.
	Force bool
}

// ShouldSpillEnv reports whether any KEY=VALUE entry exceeds EnvFileSpillMinRunes
// runes (after TrimSpace). Empty / whitespace-only entries are ignored.
func ShouldSpillEnv(entries []string) bool {
	for _, e := range entries {
		e = strings.TrimSpace(e)
		if e == "" {
			continue
		}
		if utf8.RuneCountInString(e) > EnvFileSpillMinRunes {
			return true
		}
	}
	return false
}

// MaybeSpillEnv writes all non-empty env entries to a file for --env-file
// delivery when ShouldSpillEnv is true (or Force). Under the threshold and
// !Force: returns ("", false, nil) and writes nothing.
func MaybeSpillEnv(entries []string, opts EnvSpillOpts) (path string, spilled bool, err error) {
	normalized := normalizeFollowUpEnvEntries(entries)
	if !opts.Force && !ShouldSpillEnv(normalized) {
		return "", false, nil
	}
	if len(normalized) == 0 && !opts.Force {
		return "", false, nil
	}
	path, err = spillFollowUpEnv(normalized, opts)
	if err != nil {
		return "", false, err
	}
	return path, true, nil
}

func normalizeFollowUpEnvEntries(entries []string) []string {
	if len(entries) == 0 {
		return nil
	}
	out := make([]string, 0, len(entries))
	for _, e := range entries {
		e = strings.TrimSpace(e)
		if e == "" {
			continue
		}
		out = append(out, e)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// spillFollowUpEnv writes KEY=VALUE lines under SpillDir (or production temp
// default) and returns an absolute path for --env-file.
func spillFollowUpEnv(entries []string, opts EnvSpillOpts) (string, error) {
	dir := strings.TrimSpace(opts.SpillDir)
	if dir == "" {
		dir = filepath.Join(os.TempDir(), "agent-run-env-spill")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("env spill dir: %w", err)
	}
	pattern := "env-*.env"
	if sid := strings.TrimSpace(opts.SessionID); sid != "" {
		safe := strings.Map(func(r rune) rune {
			if r == '/' || r == '\\' || r == 0 {
				return '_'
			}
			return r
		}, sid)
		pattern = "env-" + safe + "-*.env"
	}
	f, err := os.CreateTemp(dir, pattern)
	if err != nil {
		return "", fmt.Errorf("create env spill file: %w", err)
	}
	path := f.Name()
	body := strings.Join(entries, "\n")
	if body != "" {
		body += "\n"
	}
	if _, err := f.WriteString(body); err != nil {
		f.Close()
		_ = os.Remove(path)
		return "", fmt.Errorf("write env spill file: %w", err)
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(path)
		return "", fmt.Errorf("close env spill file: %w", err)
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return path, nil
	}
	return abs, nil
}
