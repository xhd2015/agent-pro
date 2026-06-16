package print

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/xhd2015/agent-pro/agent_trace/types"
	_ "github.com/xhd2015/agent-pro/agent_trace/opencode"
	_ "github.com/xhd2015/agent-pro/agent_trace/pi"
)

const TRUNCATE_LINE_MAX = 1024

func FormatTraceLine(line string) string {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" || !strings.HasPrefix(trimmed, "{") {
		return ""
	}
	parsed, ok := types.ParseAgentTraceLine(json.RawMessage(trimmed))
	if !ok {
		return ""
	}
	if parsed.Message != nil {
		return FormatMessageCompact(*parsed.Message)
	}
	if parsed.Activity != nil {
		return FormatMessageCompact(types.AgentTraceMessage{
			Role:     types.RoleToolCall,
			ToolCall: parsed.Activity,
		})
	}
	return ""
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

func writeHumanMessage(w io.Writer, n int, msg types.AgentTraceMessage) {
	if msg.ToolCall != nil {
		tc := msg.ToolCall
		tool := strings.ToLower(tc.ToolName)
		summary := strings.TrimSpace(tc.Summary)

		icon, label := toolIcon(tool)
		fmt.Fprintf(w, "[%d]  %-4s %s\n", n, icon, label)

		for _, line := range strings.Split(summary, "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			fmt.Fprintf(w, "     %s\n", truncateLine(line, TRUNCATE_LINE_MAX))
		}

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
