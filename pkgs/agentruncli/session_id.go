package agentruncli

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/xhd2015/agent-pro/pkgs/agenttty"
	"github.com/xhd2015/agent-pro/pkgs/ttywatch"
)

const autoSessionIDBaseMaxRunes = 128
const autoSessionIDFallbackBase = "sess"

// slugifyPrompt turns a prompt into a session-id base slug:
// lowercase, non [a-z0-9] → '-', collapse/trim '-', empty → "sess",
// max 128 runes with trailing '-' re-trimmed after truncate.
func slugifyPrompt(prompt string) string {
	s := strings.ToLower(prompt)
	var b strings.Builder
	b.Grow(len(s))
	prevDash := false
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			prevDash = false
			continue
		}
		if !prevDash {
			b.WriteByte('-')
			prevDash = true
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		return autoSessionIDFallbackBase
	}
	if utf8.RuneCountInString(out) > autoSessionIDBaseMaxRunes {
		runes := []rune(out)
		out = strings.TrimRight(string(runes[:autoSessionIDBaseMaxRunes]), "-")
		if out == "" {
			return autoSessionIDFallbackBase
		}
	}
	return out
}

// generateAutoSessionID builds base-YYYYMMDD-HHMMSS (local), then -1, -2, … on conflict.
func generateAutoSessionID(prompt, runner, home string) (string, error) {
	base := slugifyPrompt(prompt)
	ts := time.Now().Local().Format("20060102-150405")
	candidate := base + "-" + ts
	if !sessionIDTaken(home, runner, candidate) {
		return candidate, nil
	}
	for n := 1; n < 10000; n++ {
		id := fmt.Sprintf("%s-%d", candidate, n)
		if !sessionIDTaken(home, runner, id) {
			return id, nil
		}
	}
	return "", fmt.Errorf("could not allocate free session id for base %q", base)
}

// sessionIDTaken reports whether storage already has the id, or (for TTY runners)
// a live terminal registry entry holds it.
func sessionIDTaken(home, runner, id string) bool {
	metaPath := filepath.Join(home, "sessions", id, "meta.json")
	if _, err := os.Stat(metaPath); err == nil {
		return true
	}
	// Also treat an existing session directory as taken (meta may be mid-write).
	dir := filepath.Join(home, "sessions", id)
	if st, err := os.Stat(dir); err == nil && st.IsDir() {
		return true
	}
	if !agenttty.IsTTYRunner(runner) {
		return false
	}
	provider, ok := agenttty.Get(runner)
	if !ok {
		return false
	}
	cfg := ttywatch.RegistryConfig{Home: home, Subdir: provider.RegistryDir}
	entry, err := ttywatch.ReadRegistry(cfg, id)
	if err != nil || entry == nil {
		return false
	}
	return registryEntryLive(entry.ListenAddr)
}

func registryEntryLive(addr string) bool {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return false
	}
	conn, err := net.DialTimeout("tcp", addr, 200*time.Millisecond)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}
