package run

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/xhd2015/agent-pro/agent/event/print"
	"github.com/xhd2015/agent-pro/agent_trace/types"
	"github.com/xhd2015/agent-pro/trace"
	"github.com/xhd2015/dot-pkgs/go-pkgs/logs"
)

const TRUNCATE_LINE_MAX = 1024

func printHumanReport(source trace.Source, descriptions []string) error {
	return writeHumanReport(os.Stdout, source, descriptions)
}

func FormatMessage(msg types.AgentTraceMessage) string {
	var buf strings.Builder
	writeHumanMessage(&buf, 0, msg)
	return strings.TrimRight(buf.String(), "\n")
}

func FormatMessageCompact(msg types.AgentTraceMessage) string {
	return print.FormatMessageCompact(msg)
}

func FormatTraceLine(line string) string {
	return print.FormatTraceLine(line)
}

func writeHumanReport(w io.Writer, source trace.Source, descriptions []string) error {
	summaries, err := source.List()
	if err != nil {
		return err
	}

	for _, summary := range summaries {
		detail, err := source.Get(summary.ID)
		if err != nil {
			fmt.Fprintf(w, "  detail_error: %v\n", err)
			continue
		}
		if err := writeHumanTrace(w, summary, detail); err != nil {
			return err
		}
	}

	return nil
}

func writeHumanTrace(w io.Writer, summary trace.AgentTraceSummary, detail *trace.AgentTraceDetail) error {
	sessionID := extractSessionID(detail)
	runner := emptyDefault(summary.AgentRunnerID, "agent")
	lines := summary.LogLineCount

	fmt.Fprintln(w)
	fmt.Fprintf(w, "═══════════════════════════════════════════════════════════════\n")
	if sessionID != "" {
		fmt.Fprintf(w, "  Session: %s\n", sessionID)
	}
	fmt.Fprintf(w, "  Events:  %d lines  ·  %s  ·  %s\n", lines, runner, statusDisplay(summary.Status))
	fmt.Fprintf(w, "═══════════════════════════════════════════════════════════════\n")
	fmt.Fprintln(w)

	if detail == nil || len(detail.Messages) == 0 {
		fmt.Fprintln(w, "  (no messages)")
	} else {
		for i, msg := range detail.Messages {
			writeHumanMessage(w, i+1, msg)
		}
	}

	if summary.Status != "running" || detail == nil || detail.Metadata.LogPath == "" {
		fmt.Fprintln(w, "───────────────────────────────────────────────────────────────")
		printFinalStatus(w, summary)
		fmt.Fprintln(w, "───────────────────────────────────────────────────────────────")
		return nil
	}

	logPath := detail.Metadata.LogPath
	metaPath := filepath.Join(filepath.Dir(logPath), "metadata.json")

	watchCtx, watchCancel := context.WithCancel(context.Background())
	defer watchCancel()
	statusCtx, statusCancel := context.WithCancel(context.Background())
	defer statusCancel()

	sigCtx, _ := signal.NotifyContext(context.Background(), os.Interrupt)
	go func() {
		<-sigCtx.Done()
		watchCancel()
		statusCancel()
	}()

	nextNum := len(detail.Messages)
	var mu sync.Mutex
	statusDone := make(chan struct{})
	var statusDoneOnce sync.Once
	signalStatusDone := func() {
		statusDoneOnce.Do(func() {
			statusCancel()
			close(statusDone)
		})
	}

	go func() {
		logs.WatchFileEvents(statusCtx, metaPath, logs.WatchFileEventsOptions{
			DisableDebounce: true,
		}, func(ev logs.FileEvent) error {
			data, err := os.ReadFile(metaPath)
			if err != nil {
				return nil
			}
			var meta trace.AgentTraceMetadata
			if err := json.Unmarshal(data, &meta); err != nil {
				return nil
			}
			if meta.Status != "running" {
				signalStatusDone()
			}
			return nil
		})
	}()

	watchErr := make(chan error, 1)
	go func() {
		watchErr <- logs.WatchLine(watchCtx, logPath, logs.WatchLineOptions{}, func(line string) error {
			parsed, ok := types.ParseAgentTraceLine(json.RawMessage(line))
			if !ok {
				return nil
			}
			mu.Lock()
			nextNum++
			if parsed.Message != nil {
				writeHumanMessage(w, nextNum, *parsed.Message)
			} else if parsed.Activity != nil {
				writeHumanMessage(w, nextNum, types.AgentTraceMessage{
					Role:     types.RoleToolCall,
					ToolCall: parsed.Activity,
				})
			}
			mu.Unlock()
			return nil
		})
	}()

	<-statusDone

	time.Sleep(200 * time.Millisecond)
	watchCancel()

	data, _ := os.ReadFile(metaPath)
	var finalMeta trace.AgentTraceMetadata
	json.Unmarshal(data, &finalMeta)
	fmt.Fprintln(w, "───────────────────────────────────────────────────────────────")
	printFinalStatus(w, trace.AgentTraceSummary{AgentTraceMetadata: finalMeta})
	fmt.Fprintln(w, "───────────────────────────────────────────────────────────────")

	return <-watchErr
}

func writeHumanMessage(w io.Writer, n int, msg types.AgentTraceMessage) {
	if msg.ToolCall != nil {
		tc := msg.ToolCall
		tool := strings.ToLower(tc.ToolName)
		summary := strings.TrimSpace(tc.Summary)

		icon, label := toolIcon(tool)
		fmt.Fprintf(w, "[%d]  %-4s %s\n", n, icon, label)

		// show summary
		for _, line := range strings.Split(summary, "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			fmt.Fprintf(w, "     %s\n", truncateLine(line, TRUNCATE_LINE_MAX))
		}

		// show file changes
		if len(tc.FileChanges) > 0 {
			for _, fc := range tc.FileChanges {
				fmt.Fprintf(w, "     →  %s %s\n", fc.Kind, shortPath(fc.Path))
			}
		}

		if tc.Status == types.StatusFailed {
			fmt.Fprintf(w, "     ✗  FAILED\n")
		}
		fmt.Fprintln(w)
	} else {
		text := strings.TrimSpace(msg.Content)
		if text == "" {
			return
		}
		fmt.Fprintf(w, "[%d]  💬   ASSISTANT\n", n)
		for _, line := range strings.Split(text, "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			fmt.Fprintf(w, "     %s\n", truncateLine(line, TRUNCATE_LINE_MAX))
		}
		fmt.Fprintln(w)
	}
}

func writeHumanMessageCompact(w io.Writer, msg types.AgentTraceMessage) {
	if msg.ToolCall != nil {
		tc := msg.ToolCall
		tool := strings.ToLower(tc.ToolName)
		summary := strings.TrimSpace(tc.Summary)

		icon, label := toolIcon(tool)
		fmt.Fprintf(w, "%-4s %s\n", icon, label)

		for _, line := range strings.Split(summary, "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			fmt.Fprintf(w, "  %s\n", truncateLine(line, TRUNCATE_LINE_MAX))
		}

		if len(tc.FileChanges) > 0 {
			for _, fc := range tc.FileChanges {
				fmt.Fprintf(w, "  →  %s %s\n", fc.Kind, shortPath(fc.Path))
			}
		}

		if tc.Status == types.StatusFailed {
			fmt.Fprintf(w, "  ✗  FAILED\n")
		}
		fmt.Fprintln(w)
	} else {
		text := strings.TrimSpace(msg.Content)
		if text == "" {
			return
		}
		fmt.Fprintf(w, "💬   ASSISTANT\n")
		for _, line := range strings.Split(text, "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			fmt.Fprintf(w, "  %s\n", truncateLine(line, TRUNCATE_LINE_MAX))
		}
		fmt.Fprintln(w)
	}
}

func toolIcon(tool string) (string, string) {
	switch tool {
	case "bash", "shell", "execute", "exec", "run":
		return "⚡", "RUN"
	case "read", "read_file", "read file":
		return "📖", "READ"
	case "write", "edit", "write_file", "write file", "patch":
		return "✏️", "EDIT"
	case "list", "ls", "list_files", "list files":
		return "🔍", "LIST"
	case "glob", "search":
		return "🔎", "SEARCH"
	case "grep":
		return "🔎", "GREP"
	case "task":
		return "🤖", "TASK"
	case "todowrite", "todo":
		return "📋", "TODO"
	default:
		return "🔧", strings.ToUpper(tool)
	}
}

func extractSessionID(detail *trace.AgentTraceDetail) string {
	if detail == nil {
		return ""
	}
	for _, raw := range detail.RawLines {
		var m struct {
			SessionID string `json:"sessionID"`
		}
		if err := json.Unmarshal(raw, &m); err == nil && m.SessionID != "" {
			return m.SessionID
		}
	}
	return ""
}

func truncateLine(s string, max int) string {
	if len(s) <= max {
		return s
	}
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max]) + "..."
}

func shortPath(path string) string {
	parts := strings.Split(path, string(os.PathSeparator))
	if len(parts) > 2 {
		parts = parts[len(parts)-2:]
	}
	return strings.Join(parts, "/")
}

func emptyDefault(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func statusDisplay(status string) string {
	switch status {
	case "running":
		return "⚡ running"
	case "completed":
		return "✓ completed"
	case "failed":
		return "✗ failed"
	case "stopped":
		return "⊘ stopped"
	default:
		return status + " status"
	}
}

func printFinalStatus(w io.Writer, summary trace.AgentTraceSummary) {
	if summary.Status == "completed" && summary.Error == "" {
		fmt.Fprintln(w, "  ✓ Done")
	} else if summary.Status == "failed" {
		fmt.Fprintf(w, "  ✗ Failed: %s\n", summary.Error)
	} else {
		fmt.Fprintf(w, "  ○ Status: %s\n", summary.Status)
	}
}
