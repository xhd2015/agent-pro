package run

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/xhd2015/agent-pro/trace"
)

func printTraceReport(source trace.Source, descriptions []string, messageLimit int) error {
	return writeTraceReport(os.Stdout, source, descriptions, messageLimit)
}

func writeTraceReport(w io.Writer, source trace.Source, descriptions []string, messageLimit int) error {
	if messageLimit < 0 {
		messageLimit = 0
	}
	fmt.Fprintln(w, "Agent trace sources:")
	if len(descriptions) == 0 {
		fmt.Fprintln(w, "  none discovered")
	} else {
		for _, desc := range descriptions {
			fmt.Fprintf(w, "  %s\n", desc)
		}
	}
	summaries, err := source.List()
	if err != nil {
		return err
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Traces:")
	if len(summaries) == 0 {
		fmt.Fprintln(w, "  none")
		return nil
	}
	for _, summary := range summaries {
		writeTraceSummary(w, summary)
		if messageLimit == 0 {
			continue
		}
		detail, err := source.Get(summary.ID)
		if err != nil {
			fmt.Fprintf(w, "  detail_error: %v\n", err)
			continue
		}
		writeRecentMessages(w, detail, messageLimit)
	}
	return nil
}

func writeTraceSummary(w io.Writer, summary trace.AgentTraceSummary) {
	fmt.Fprintf(w, "- %s %s %s", summary.ID, emptyDefault(summary.Command, "agent"), emptyDefault(summary.Status, "unknown"))
	if summary.AgentRunnerID != "" || summary.Model != "" {
		fmt.Fprintf(w, " %s/%s", emptyDefault(summary.AgentRunnerID, "agent"), emptyDefault(summary.Model, "default"))
	}
	fmt.Fprintln(w)
	if summary.DelegationLabel != "" || summary.DelegationID != "" {
		fmt.Fprintf(w, "  delegation: %s\n", emptyDefault(firstNonEmpty(summary.DelegationLabel, summary.DelegationID), "-"))
	}
	if summary.ParentTraceID != "" {
		fmt.Fprintf(w, "  parent: %s\n", summary.ParentTraceID)
	}
	if summary.Workspace != "" {
		fmt.Fprintf(w, "  workspace: %s\n", summary.Workspace)
	}
	fmt.Fprintf(w, "  lines: %d\n", summary.LogLineCount)
	if len(summary.Children) > 0 {
		fmt.Fprintln(w, "  children:")
		for _, child := range summary.Children {
			label := firstNonEmpty(child.DelegationLabel, child.DelegationID)
			if label != "" {
				label = " label=" + label
			}
			fmt.Fprintf(w, "    - %s %s %s%s\n", child.ID, emptyDefault(child.Command, "agent"), emptyDefault(child.Status, "unknown"), label)
		}
	}
}

func writeRecentMessages(w io.Writer, detail *trace.AgentTraceDetail, limit int) {
	if detail == nil || len(detail.Messages) == 0 {
		return
	}
	start := len(detail.Messages) - limit
	if start < 0 {
		start = 0
	}
	fmt.Fprintln(w, "  recent_messages:")
	for _, msg := range detail.Messages[start:] {
		switch {
		case msg.ToolCall != nil:
			fmt.Fprintf(w, "    tool_call: %s\n", trimOneLine(msg.ToolCall.Summary, 160))
		default:
			fmt.Fprintf(w, "    %s: %s\n", emptyDefault(string(msg.Role), "message"), trimOneLine(msg.Content, 220))
		}
	}
}

func trimOneLine(value string, limit int) string {
	value = strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
	if limit > 0 && len(value) > limit {
		return value[:limit] + "..."
	}
	if value == "" {
		return "-"
	}
	return value
}

func emptyDefault(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
