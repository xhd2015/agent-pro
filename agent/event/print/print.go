package print

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	eventtypes "github.com/xhd2015/agent-pro/agent/event/types"
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
	// PRIMARY: try AgentEvent (match known action types only, so old native
	// formats like opencode type="text" or pi type="message_start" fall through)
	var event eventtypes.AgentEvent
	if err := json.Unmarshal([]byte(trimmed), &event); err == nil && isAgentEventType(event.Type) {
		return FormatAgentEvent(event)
	}
	// FALLBACK: old adapter system
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

// isAgentEventType returns true for known AgentEvent action types, so that
// old native event formats (e.g. opencode type="text", pi type="message_start")
// are not matched by the AgentEvent primary path.
func isAgentEventType(t eventtypes.ActionType) bool {
	switch t {
	case eventtypes.ActionThink, eventtypes.ActionToolCall, eventtypes.ActionMessage,
		eventtypes.ActionError, eventtypes.ActionDone,
		eventtypes.ActionStepStart, eventtypes.ActionStepFinish,
		eventtypes.ActionSleep:
		return true
	}
	return false
}

// FormatAgentEvent formats an AgentEvent into a human-readable string.
func FormatAgentEvent(event eventtypes.AgentEvent) string {
	switch event.Type {
	case eventtypes.ActionToolCall:
		icon, label := toolIcon(event.Tool)
		var parts []string
		parts = append(parts, fmt.Sprintf("%s %s", icon, label))
		if event.Text != "" {
			parts = append(parts, event.Text)
		}
		if event.Output != "" {
			parts = append(parts, event.Output)
		}
		if len(event.Changes) > 0 {
			for _, c := range event.Changes {
				parts = append(parts, c.Kind+" "+c.Path)
			}
		}
		if event.ExitCode != nil && *event.ExitCode != 0 {
			parts = append(parts, "FAILED")
		}
		return strings.Join(parts, "\n")
	case eventtypes.ActionMessage:
		return fmt.Sprintf("💬 %s", event.Text)
	case eventtypes.ActionThink:
		return fmt.Sprintf("💭 %s", event.Text)
	case eventtypes.ActionError:
		return fmt.Sprintf("❌ %s\n✗ FAILED", event.Text)
	case eventtypes.ActionStepStart:
		return "▶ STEP START"
	case eventtypes.ActionStepFinish:
		return "◼ STEP FINISH"
	default:
		return fmt.Sprintf("[%s] %s", event.Type, event.Text)
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

// FormatState provides streaming coalescing of consecutive assistant message
// deltas into a single output block. Use FormatLine for each event line in a
// loop; it returns a header, body, and a flag indicating the output type.
// Call Flush after the last line to close any open message block.
//
// Usage:
//
//	var state print.FormatState
//	for _, line := range lines {
//	    header, body, isMsg := state.FormatLine(line)
//	    if header == "" && body == "" && !isMsg {
//	        continue
//	    }
//	    if isMsg {
//	        if header != "" {
//	            n++
//	            Logf("[%d]  %s", n, header)
//	        }
//	        fmt.Print(body)
//	    } else {
//	        n++
//	        Logf("[%d]  %s", n, header)
//	        if body != "" {
//	            fmt.Print(body)
//	        }
//	    }
//	}
//	state.Flush()
type FormatState struct {
	msgStarted bool
}

// FormatLine processes one JSONL event line.
//
// Returns:
//   - header: the block header line (e.g. "💬   ASSISTANT", "⚡ RUN", "▶ STEP START")
//   - body:   the text to print after the header (without trailing newline)
//   - isMsg:  true if this is an assistant message (first delta or continuation);
//             false for non-message events (tool calls, steps, etc.)
//
// For the first assistant message delta, header and body are both non-empty.
// For subsequent deltas, only body is non-empty (continuation, no header).
// For non-message events, header is the formatted block and body is empty.
// Returns all-empty when nothing to output.
func (s *FormatState) FormatLine(line string) (header string, body string, isMsg bool) {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" || !strings.HasPrefix(trimmed, "{") {
		return "", "", false
	}

	// --- PRIMARY: AgentEvent path ---
	var event eventtypes.AgentEvent
	if err := json.Unmarshal([]byte(trimmed), &event); err == nil && isAgentEventType(event.Type) {
		if event.Type == eventtypes.ActionMessage {
			if !s.msgStarted {
				s.msgStarted = true
				return "💬   ASSISTANT", "  " + event.Text, true
			}
			return "", event.Text, true
		}
		s.closeMsg()
		return FormatAgentEvent(event), "", false
	}

	// --- FALLBACK: adapter path ---
	parsed, ok := types.ParseAgentTraceLine(json.RawMessage(trimmed))
	if !ok {
		return "", "", false
	}
	if parsed.Message != nil && parsed.Message.Role == types.RoleAssistant {
		if !s.msgStarted {
			s.msgStarted = true
			return "💬   ASSISTANT", "  " + parsed.Message.Content, true
		}
		return "", parsed.Message.Content, true
	}

	s.closeMsg()
	var formatted string
	if parsed.Message != nil {
		formatted = FormatMessageCompact(*parsed.Message)
	} else if parsed.Activity != nil {
		formatted = FormatMessageCompact(types.AgentTraceMessage{
			Role:     types.RoleToolCall,
			ToolCall: parsed.Activity,
		})
	}
	return formatted, "", false
}

// Flush closes any pending message block by printing a trailing newline.
func (s *FormatState) Flush() {
	s.closeMsg()
}

func (s *FormatState) closeMsg() {
	if s.msgStarted {
		fmt.Print("\n")
		s.msgStarted = false
	}
}
