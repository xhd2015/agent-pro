package run

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/xhd2015/agent-pro/agent_trace/types"
	"github.com/xhd2015/agent-pro/trace"
)

func printHumanReport(source trace.Source, descriptions []string) error {
	return writeHumanReport(os.Stdout, source, descriptions)
}

func FormatMessage(msg types.AgentTraceMessage) string {
	var buf strings.Builder
	writeHumanMessage(&buf, 0, msg)
	return strings.TrimRight(buf.String(), "\n")
}

func FormatMessageCompact(msg types.AgentTraceMessage) string {
	var buf strings.Builder
	writeHumanMessageCompact(&buf, msg)
	return strings.TrimRight(buf.String(), "\n")
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
		if err := writeHumanTrace(w, source, summary, detail); err != nil {
			return err
		}
	}

	return nil
}

func writeHumanTrace(w io.Writer, source trace.Source, summary trace.AgentTraceSummary, detail *trace.AgentTraceDetail) error {
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

	if summary.Status == "running" && detail != nil && detail.Metadata.LogPath != "" {
		return followTraceSession(w, source, summary.ID, detail.Metadata.LogPath, len(detail.Messages), summary)
	}

	fmt.Fprintln(w, "───────────────────────────────────────────────────────────────")
	printFinalStatus(w, summary)
	fmt.Fprintln(w, "───────────────────────────────────────────────────────────────")
	return nil
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
			fmt.Fprintf(w, "     %s\n", truncateLine(line, 70))
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
			fmt.Fprintf(w, "     %s\n", truncateLine(line, 70))
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
			fmt.Fprintf(w, "  %s\n", truncateLine(line, 70))
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
			fmt.Fprintf(w, "  %s\n", truncateLine(line, 70))
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
	if len(s) > max {
		return s[:max-3] + "..."
	}
	return s
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

func followTraceSession(w io.Writer, source trace.Source, id string, logPath string, printedCount int, summary trace.AgentTraceSummary) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}
	defer watcher.Close()

	dir := logPath[:strings.LastIndex(logPath, string(os.PathSeparator))]
	if err := watcher.Add(dir); err != nil {
		return fmt.Errorf("failed to watch directory %s: %w", dir, err)
	}
	if _, err := os.Stat(logPath); err == nil {
		if err := watcher.Add(logPath); err != nil {
			return fmt.Errorf("failed to watch file %s: %w", logPath, err)
		}
	}

	statusTicker := time.NewTicker(500 * time.Millisecond)
	defer statusTicker.Stop()

	flushPending := func() error {
		detail, err := source.Get(id)
		if err != nil {
			return err
		}
		newCount := len(detail.Messages)
		if newCount > printedCount {
			for i := printedCount; i < newCount; i++ {
				writeHumanMessage(w, i+1, detail.Messages[i])
			}
			printedCount = newCount
		}
		if detail.Metadata.Status != "running" {
			fmt.Fprintln(w, "───────────────────────────────────────────────────────────────")
			printFinalStatus(w, trace.AgentTraceSummary{
				AgentTraceMetadata: detail.Metadata,
			})
			fmt.Fprintln(w, "───────────────────────────────────────────────────────────────")
			cancel()
			return nil
		}
		return nil
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}
			if event.Op&fsnotify.Create == fsnotify.Create && event.Name == logPath {
				if err := watcher.Add(logPath); err != nil {
					return fmt.Errorf("failed to watch newly created file %s: %w", logPath, err)
				}
			}
			if event.Name == logPath && (event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create) {
				if err := flushPending(); err != nil {
					fmt.Fprintf(os.Stderr, "Error reading trace: %v\n", err)
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			return fmt.Errorf("watcher error: %w", err)
		case <-statusTicker.C:
			if err := flushPending(); err != nil {
				fmt.Fprintf(os.Stderr, "Error polling trace: %v\n", err)
			}
		case <-ctx.Done():
			return nil
		}
	}
}
