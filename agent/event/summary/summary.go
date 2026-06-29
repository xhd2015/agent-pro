package summary

import (
	"fmt"
	"strings"
)

func CompactTraceOutput(output string) string {
	if len([]byte(output)) <= 4000 && strings.Count(output, "\n") <= 40 {
		return output
	}
	lines := strings.Split(strings.TrimRight(output, "\n"), "\n")
	head := traceJoinLines(lines, 0, minTraceInt(10, len(lines)))
	tailStart := len(lines) - minTraceInt(10, len(lines))
	if tailStart < 10 {
		tailStart = 10
	}
	tail := traceJoinLines(lines, tailStart, len(lines))
	return fmt.Sprintf("[omitted: %d lines, %d bytes]\n--- first lines ---\n%s\n--- last lines ---\n%s", len(lines), len([]byte(output)), head, tail)
}

func traceJoinLines(lines []string, start, end int) string {
	if start >= end {
		return ""
	}
	out := make([]string, 0, end-start)
	for _, line := range lines[start:end] {
		out = append(out, truncateTraceLine(line))
	}
	return strings.Join(out, "\n")
}

func truncateTraceLine(line string) string {
	runes := []rune(line)
	if len(runes) <= 260 {
		return line
	}
	return string(runes[:260]) + "...<line truncated>"
}

func minTraceInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func TitleFromIdentifier(id string) string {
	if strings.TrimSpace(id) == "" {
		return "Tool"
	}
	parts := strings.FieldsFunc(id, func(r rune) bool {
		return r == '_' || r == '-' || r == '.'
	})
	for i, part := range parts {
		if part == "" {
			continue
		}
		parts[i] = strings.ToUpper(part[:1]) + part[1:]
	}
	return strings.Join(parts, " ")
}