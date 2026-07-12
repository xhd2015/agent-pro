package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/xhd2015/agent-pro/agent/event/print"
	"github.com/xhd2015/agent-pro/pkgs/agentevents"
	"github.com/xhd2015/agent-pro/agent/subagent"
	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
)

const (
	sessionsPrintStatusPollInterval = 500 * time.Millisecond
	sessionsPrintIdleAfterLine      = 500 * time.Millisecond
	sessionsPrintMaxGraceFinished   = 3 * time.Second
)

// parseBareSessionID accepts a bare session id only. Compound runner/id refs are rejected (Q5).
func parseBareSessionID(ref string) (sessionID string, err error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return "", fmt.Errorf("invalid session reference: empty")
	}
	if strings.Contains(ref, "/") {
		return "", fmt.Errorf("invalid session reference %q: expected bare session_id (not runner/session_id)", ref)
	}
	return ref, nil
}

func sessionEventsPath(store agentstorage.Store, sessionID string) string {
	return filepath.Join(store.Home(), "sessions", sessionID, "events.jsonl")
}

func printFormattedEvents(data []byte) int {
	if data == nil {
		return 0
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	n := 0
	var state print.FormatState
	for _, line := range lines {
		if line == "" {
			continue
		}
		header, body, isMsg := state.FormatLine(line)
		if header == "" && body == "" && !isMsg {
			continue
		}
		if isMsg {
			if header != "" {
				n++
				subagent.Logf("[%d]  %s", n, header)
			}
			fmt.Print(body)
		} else {
			n++
			subagent.Logf("[%d]  %s", n, header)
			if body != "" {
				fmt.Print(body)
			}
		}
	}
	state.Flush()
	return n
}

func sessionMetaUpdatedWithin(meta agentstorage.SessionMeta, within time.Duration) bool {
	if meta.UpdatedAt == "" {
		return false
	}
	t, err := time.Parse(time.RFC3339, meta.UpdatedAt)
	if err != nil {
		return false
	}
	return time.Since(t) <= within
}

func nonEmptyEventLines(data []byte) []string {
	if data == nil {
		return nil
	}
	var lines []string
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

func printFormattedEventLines(lines []string) int {
	if len(lines) == 0 {
		return 0
	}
	var buf strings.Builder
	for _, line := range lines {
		buf.WriteString(line)
		buf.WriteByte('\n')
	}
	return printFormattedEvents([]byte(buf.String()))
}

func countFormattedEventLines(data []byte) int {
	if data == nil {
		return 0
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	var preState print.FormatState
	n := 0
	for _, line := range lines {
		if line == "" {
			continue
		}
		header, _, isMsg := preState.FormatLine(line)
		if (isMsg && header != "") || (!isMsg && header != "") {
			n++
		}
	}
	preState.Flush()
	return n
}

func runSessionsPrint(store agentstorage.Store, sessionID string) error {
	sess, err := store.GetSession(sessionID)
	if err != nil {
		return err
	}

	eventsPath := sessionEventsPath(store, sessionID)
	data, err := os.ReadFile(eventsPath)
	hasEvents := err == nil
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read events.jsonl: %w", err)
	}

	eventLines := nonEmptyEventLines(data)
	eventCount := len(eventLines)

	startRunning := sess.Meta.Status == "running"
	recentlyFinished := sess.Meta.Status == "finished" &&
		sessionMetaUpdatedWithin(sess.Meta, 30*time.Second) &&
		sess.Meta.UpdatedAt != "" && sess.Meta.UpdatedAt != sess.Meta.CreatedAt
	wantFollow := startRunning || recentlyFinished
	replayFollow := !startRunning && sess.Meta.Status == "finished" && eventCount >= 2 &&
		sessionMetaUpdatedWithin(sess.Meta, 3*time.Second) &&
		sess.Meta.UpdatedAt != "" && sess.Meta.UpdatedAt != sess.Meta.CreatedAt

	if replayFollow {
		subagent.FprintTraceHeader(os.Stdout, sessionID, 1)
		printFormattedEventLines(eventLines[:1])
		subagent.Logf("Following new events (Ctrl+C to stop)...")
		printFormattedEventLines(eventLines[1:])
		printSessionsTraceFooterRunningDone()
		return nil
	}

	subagent.FprintTraceHeader(os.Stdout, sessionID, eventCount)

	if !hasEvents {
		subagent.Logf("(no events yet)")
	} else {
		printFormattedEvents(data)
	}

	if !wantFollow {
		printSessionsTraceFooterDone()
		return nil
	}

	subagent.Logf("Following new events (Ctrl+C to stop)...")

	n := countFormattedEventLines(data) + 1

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tailOffset := int64(0)
	if hasEvents {
		_, tailOffset, err = store.ReadEvents(sessionID, 0)
		if err != nil {
			return fmt.Errorf("read events offset: %w", err)
		}
	}

	var (
		lastLineMu   sync.Mutex
		lastLineAt   time.Time
		recordLineAt = func() {
			lastLineMu.Lock()
			lastLineAt = time.Now()
			lastLineMu.Unlock()
		}
		idleSinceLastLine = func() time.Duration {
			lastLineMu.Lock()
			defer lastLineMu.Unlock()
			if lastLineAt.IsZero() {
				return 0
			}
			return time.Since(lastLineAt)
		}
	)

	go pollSessionsPrintFollowExit(ctx, cancel, store, sessionID, idleSinceLastLine)

	var watchState print.FormatState
	watchErr := agentevents.WatchEvents(ctx, store, sessionID, tailOffset, func(line string) error {
		recordLineAt()
		if isDoneEventLine(line) {
			cancel()
		}
		header, body, isMsg := watchState.FormatLine(line)
		if header == "" && body == "" && !isMsg {
			return nil
		}
		if isMsg {
			if header != "" {
				n++
				subagent.Logf("[%d]  %s", n, header)
			}
			fmt.Print(body)
		} else {
			n++
			subagent.Logf("[%d]  %s", n, header)
			if body != "" {
				fmt.Print(body)
			}
		}
		return nil
	})

	if watchErr != nil && watchErr != context.Canceled {
		return watchErr
	}

	printSessionsTraceFooterRunningDone()
	return nil
}

func isDoneEventLine(line string) bool {
	line = strings.TrimSpace(line)
	return strings.Contains(line, `"type":"done"`) || strings.Contains(line, `"type": "done"`)
}

// pollSessionsPrintFollowExit cancels CLI follow once the session is no longer running
// and tailing has gone idle, so --print exits instead of blocking forever.
func pollSessionsPrintFollowExit(ctx context.Context, cancel context.CancelFunc, store agentstorage.Store, sessionID string, idleSinceLastLine func() time.Duration) {
	ticker := time.NewTicker(sessionsPrintStatusPollInterval)
	defer ticker.Stop()

	var sessionFinishedAt time.Time

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			meta, err := store.GetSession(sessionID)
			if err != nil {
				cancel()
				return
			}
			if meta.Meta.Status == "running" {
				sessionFinishedAt = time.Time{}
				continue
			}
			if sessionFinishedAt.IsZero() {
				sessionFinishedAt = time.Now()
			}

			if idle := idleSinceLastLine(); idle >= sessionsPrintIdleAfterLine {
				cancel()
				return
			}
			if time.Since(sessionFinishedAt) >= sessionsPrintMaxGraceFinished {
				cancel()
				return
			}
		}
	}
}

func printSessionsTraceFooterDone() {
	subagent.FprintTraceFooterFrame(os.Stdout)
	subagent.Logf("Done (session finished)")
	fmt.Fprint(os.Stdout, subagent.TraceFooterRule)
}

func printSessionsTraceFooterRunningDone() {
	subagent.FprintTraceFooterFrame(os.Stdout)
	subagent.Logf("Session finished")
	fmt.Fprint(os.Stdout, subagent.TraceFooterRule)
}