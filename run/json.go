package run

import (
	"encoding/json"
	"io"
	"strings"

	"github.com/xhd2015/agent-pro/trace"
)

type traceJSONFilters struct {
	TraceID         string
	Command         string
	DelegationLabel string
	ParentTraceID   string
}

type traceJSONReport struct {
	Sources []string                  `json:"sources"`
	Traces  []trace.AgentTraceSummary `json:"traces"`
}

func printTraceJSONReport(w io.Writer, source trace.Source, descriptions []string, filters traceJSONFilters) error {
	summaries, err := source.List()
	if err != nil {
		return err
	}
	summaries = filterTraceJSONSummaries(summaries, filters)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(traceJSONReport{
		Sources: descriptions,
		Traces:  summaries,
	})
}

func filterTraceJSONSummaries(summaries []trace.AgentTraceSummary, filters traceJSONFilters) []trace.AgentTraceSummary {
	traceID := strings.TrimSpace(filters.TraceID)
	command := strings.TrimSpace(filters.Command)
	label := strings.TrimSpace(filters.DelegationLabel)
	parentTraceID := strings.TrimSpace(filters.ParentTraceID)
	if traceID == "" && command == "" && label == "" && parentTraceID == "" {
		return summaries
	}
	var out []trace.AgentTraceSummary
	for _, summary := range summaries {
		if traceID != "" && !traceIDMatches(summary.ID, traceID) {
			continue
		}
		if command != "" && summary.Command != command {
			continue
		}
		if parentTraceID != "" && !traceIDMatches(summary.ParentTraceID, parentTraceID) {
			continue
		}
		if label != "" && !summaryOrChildHasDelegationLabel(summary, label) {
			continue
		}
		out = append(out, summary)
	}
	return out
}

func traceIDMatches(summaryID, want string) bool {
	summaryID = strings.TrimSpace(summaryID)
	want = strings.TrimSpace(want)
	if summaryID == want {
		return true
	}
	if _, raw, ok := strings.Cut(summaryID, ":"); ok && raw == want {
		return true
	}
	return false
}

func summaryOrChildHasDelegationLabel(summary trace.AgentTraceSummary, label string) bool {
	if summary.DelegationLabel == label || summary.DelegationID == label {
		return true
	}
	for _, child := range summary.Children {
		if child.DelegationLabel == label || child.DelegationID == label {
			return true
		}
	}
	return false
}
