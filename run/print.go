package run

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

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
		writeHumanTrace(w, summary, detail)
	}

	return nil
}

func writeHumanTrace(w io.Writer, summary trace.AgentTraceSummary, detail *trace.AgentTraceDetail) {
	sessionID := extractSessionID(detail)
	runner := emptyDefault(summary.AgentRunnerID, "agent")
	lines := summary.LogLineCount

	fmt.Fprintln(w)
	fmt.Fprintf(w, "═══════════════════════════════════════════════════════════════\n")
	if sessionID != "" {
		fmt.Fprintf(w, "  Session: %s\n", sessionID)
	}
	fmt.Fprintf(w, "  Events:  %d lines  ·  %s  ·  %s status\n", lines, runner, summary.Status)
	fmt.Fprintf(w, "═══════════════════════════════════════════════════════════════\n")
	fmt.Fprintln(w)

	if detail == nil || len(detail.Messages) == 0 {
		fmt.Fprintln(w, "  (no messages)")
		return
	}

	for i, msg := range detail.Messages {
		writeHumanMessage(w, i+1, msg)
	}

	fmt.Fprintln(w, "───────────────────────────────────────────────────────────────")
	if summary.Status == "completed" && summary.Error == "" {
		fmt.Fprintln(w, "  ✓ Done")
	} else if summary.Status == "failed" {
		fmt.Fprintf(w, "  ✗ Failed: %s\n", summary.Error)
	} else {
		fmt.Fprintf(w, "  ○ Status: %s\n", summary.Status)
	}
	fmt.Fprintln(w, "───────────────────────────────────────────────────────────────")
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
