package agentruncli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/agent-pro/pkgs/agenttty"
)

// validateEnvFlags checks each -e/--env value is KEY=VALUE with non-empty KEY.
func validateEnvFlags(entries []string) error {
	for _, e := range entries {
		if err := validateEnvFlag(e); err != nil {
			return err
		}
	}
	return nil
}

func validateEnvFlag(entry string) error {
	// Keep internal spaces in values; only trim outer whitespace for presence.
	entry = strings.TrimSpace(entry)
	if entry == "" {
		return fmt.Errorf("invalid -e/--env: empty value; expected KEY=VALUE")
	}
	eq := strings.IndexByte(entry, '=')
	if eq < 0 {
		return fmt.Errorf("invalid -e/--env %q: missing '='; expected KEY=VALUE", entry)
	}
	key := entry[:eq]
	if key == "" {
		return fmt.Errorf("invalid -e/--env %q: empty key; expected KEY=VALUE", entry)
	}
	return nil
}

// LoadEnvFile reads KEY=VALUE lines from path for --env-file delivery.
// Empty / whitespace-only path → (nil, nil). Blank lines and # comments are skipped.
// Each non-comment line must pass validateEnvFlag.
func LoadEnvFile(path string) ([]string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, nil
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("read --env-file %s: %w", path, err)
	}
	defer f.Close()

	var out []string
	sc := bufio.NewScanner(f)
	// Allow long PATH= lines (default 64KiB is plenty; bump for safety).
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	lineNo := 0
	for sc.Scan() {
		lineNo++
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if err := validateEnvFlag(line); err != nil {
			return nil, fmt.Errorf("--env-file %s:%d: %w", path, lineNo, err)
		}
		out = append(out, line)
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("read --env-file %s: %w", path, err)
	}
	return out, nil
}

// MergeEnvEntries concatenates file then CLI entries (CLI last-win per key at
// child apply time). Empty inputs stay nil.
func MergeEnvEntries(fileEntries, cliEntries []string) []string {
	return appendStringLists(fileEntries, cliEntries)
}

// resolveRunEnv merges optional --env-file contents with -e/--env flags.
// CLI entries are validated first; file lines are validated while loading.
// Result: file entries then CLI (CLI overrides on duplicate keys).
func resolveRunEnv(envFile string, cliEntries []string) ([]string, error) {
	if err := validateEnvFlags(cliEntries); err != nil {
		return nil, err
	}
	cliEntries = normalizeEnvEntries(cliEntries)
	fileEntries, err := LoadEnvFile(envFile)
	if err != nil {
		return nil, err
	}
	fileEntries = normalizeEnvEntries(fileEntries)
	return MergeEnvEntries(fileEntries, cliEntries), nil
}

// resolvePrependPaths converts CLI --prepend-path values to absolute paths.
// Missing directories are soft-allowed. Empty values are a hard error.
func resolvePrependPaths(paths []string) ([]string, error) {
	if len(paths) == 0 {
		return nil, nil
	}
	out := make([]string, 0, len(paths))
	for _, p := range paths {
		p = strings.TrimSpace(p)
		if p == "" {
			return nil, fmt.Errorf("--prepend-path requires a non-empty directory")
		}
		abs, err := filepath.Abs(p)
		if err != nil {
			return nil, fmt.Errorf("--prepend-path %q: %w", p, err)
		}
		out = append(out, abs)
	}
	return out, nil
}

// resolveAgentRunnerConfigHomeAbs returns an absolute path when flag is set.
func resolveAgentRunnerConfigHomeAbs(flagValue string) (string, error) {
	v := strings.TrimSpace(flagValue)
	if v == "" {
		return "", nil
	}
	abs, err := filepath.Abs(v)
	if err != nil {
		return "", fmt.Errorf("--agent-runner-config-home %q: %w", v, err)
	}
	return abs, nil
}

// requireTTYForSessionEnv hard-errors when session env flags are used on non-TTY runners.
// envFile, when non-empty, counts as session env (same TTY-only rule as -e).
func requireTTYForSessionEnv(runner string, prependPaths, envEntries []string, envFile ...string) error {
	hasEnvFile := false
	if len(envFile) > 0 && strings.TrimSpace(envFile[0]) != "" {
		hasEnvFile = true
	}
	if len(prependPaths) == 0 && len(envEntries) == 0 && !hasEnvFile {
		return nil
	}
	if agenttty.IsTTYRunner(runner) {
		return nil
	}
	return fmt.Errorf("--prepend-path, -e/--env, and --env-file are only supported for TTY runners (got %s); non-TTY runners are not supported", runner)
}

// requireTTYForColor hard-errors when --color is used on a non-TTY runner.
func requireTTYForColor(runner string, color bool) error {
	if !color {
		return nil
	}
	if agenttty.IsTTYRunner(runner) {
		return nil
	}
	return fmt.Errorf("--color is only supported for TTY runners (got %s); non-TTY runners are not supported", runner)
}

// normalizeEnvEntries trims entries while preserving KEY=VALUE body (no key trim past first =).
func normalizeEnvEntries(entries []string) []string {
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

// appendUniqueOrder preserves order and appends extra after base (no dedup of paths).
func appendStringLists(base, extra []string) []string {
	if len(base) == 0 && len(extra) == 0 {
		return nil
	}
	out := make([]string, 0, len(base)+len(extra))
	out = append(out, base...)
	out = append(out, extra...)
	return out
}
