package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/xhd2015/agent-pro/agent/event/print"
	"github.com/xhd2015/agent-pro/pkgs/agentevents"
	"github.com/xhd2015/agent-pro/agent/subagent"
	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
)

func parseSessionRef(ref string) (runner, sessionID string, err error) {
	i := strings.Index(ref, "/")
	if i <= 0 || i == len(ref)-1 {
		return "", "", fmt.Errorf("invalid session reference %q: expected runner/session_id", ref)
	}
	return ref[:i], ref[i+1:], nil
}

func sessionEventsPath(store agentstorage.Store, runner, sessionID string) string {
	return filepath.Join(store.Home(), "sessions", runner, sessionID, "events.jsonl")
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

func runSessionsPrint(store agentstorage.Store, runner, sessionID string) error {
	sess, err := store.GetSession(runner, sessionID)
	if err != nil {
		return err
	}

	eventsPath := sessionEventsPath(store, runner, sessionID)
	data, err := os.ReadFile(eventsPath)
	hasEvents := err == nil
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read events.jsonl: %w", err)
	}

	eventLines := nonEmptyEventLines(data)
	eventCount := len(eventLines)

	startRunning := sess.Meta.Status == "running"
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

	if !startRunning {
		printSessionsTraceFooterDone()
		return nil
	}

	subagent.Logf("Following new events (Ctrl+C to stop)...")

	n := countFormattedEventLines(data) + 1

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tailOffset := int64(0)
	if hasEvents {
		_, tailOffset, err = store.ReadEvents(runner, sessionID, 0)
		if err != nil {
			return fmt.Errorf("read events offset: %w", err)
		}
	}

	var watchState print.FormatState
	watchErr := agentevents.WatchEvents(ctx, store, runner, sessionID, tailOffset, func(line string) error {
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