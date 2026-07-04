package run

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func runGrokDebugEnabled() bool {
	v := os.Getenv("LLM_MOCK_RUN_GROK_DEBUG")
	return v == "1" || v == "true" || v == "yes"
}

func runGrokDebugf(format string, args ...any) {
	if !runGrokDebugEnabled() {
		return
	}
	fmt.Fprintf(os.Stderr, "llm-mock[run-grok]: "+format+"\n", args...)
}

func describeSessionRoots(grokHome string) string {
	sessionsRoot := filepath.Join(grokHome, "sessions")
	entries, err := os.ReadDir(sessionsRoot)
	if err != nil {
		return fmt.Sprintf("sessions: (missing: %v)", err)
	}
	if len(entries) == 0 {
		return "sessions: (empty)"
	}
	var parts []string
	for _, ent := range entries {
		if !ent.IsDir() {
			continue
		}
		enc := ent.Name()
		subs, _ := os.ReadDir(filepath.Join(sessionsRoot, enc))
		var subNames []string
		for _, sub := range subs {
			if !sub.IsDir() {
				continue
			}
			hasEvents := "(no events.jsonl)"
			if _, err := os.Stat(filepath.Join(sessionsRoot, enc, sub.Name(), "events.jsonl")); err == nil {
				hasEvents = "(has events.jsonl)"
			}
			subNames = append(subNames, sub.Name()+hasEvents)
		}
		parts = append(parts, fmt.Sprintf("%s: [%s]", enc, joinComma(subNames)))
	}
	if len(parts) == 0 {
		return "sessions: (no encoded cwd dirs)"
	}
	return "sessions{" + joinComma(parts) + "}"
}

func joinComma(items []string) string {
	if len(items) == 0 {
		return ""
	}
	out := items[0]
	for i := 1; i < len(items); i++ {
		out += ", " + items[i]
	}
	return out
}

func since(start time.Time) time.Duration {
	return time.Since(start).Round(time.Millisecond)
}