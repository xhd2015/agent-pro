package sessions

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// SessionStats aggregates identity, counts, latency, tool times, and task
// aggregates for one Grok session from summary/signals/events/updates.
type SessionStats struct {
	ID, Title, CWD, Model, Agent string
	CreatedAt, LastActiveAt      time.Time

	Turns, UserMessages, AssistantMessages int
	ThinkingBlocks                         int
	ToolCalls, ToolCompleted, ToolFailed   int
	Compactions, Cancellations, Errors     int

	SessionDurationSec int
	AvgResponseMs      int
	AvgTTFTMs          int

	Tools           []ToolStat
	BackgroundTasks *TaskAgg
	Subagents       *TaskAgg

	// Per-item samples for Top background / Top subagent sections.
	// Command is stored full; display truncation is applied in FormatStatsTextOpts.
	BackgroundTaskItems []BackgroundTaskItem
	SubagentItems       []SubagentItem

	Sources StatsSources
}

// BackgroundTaskItem is one background task sample for Top-N tables.
type BackgroundTaskItem struct {
	DurationMs  float64
	Command     string // full, never store-truncate
	Description string
	ExitCode    *int
	Kind, CWD   string
}

// SubagentItem is one subagent sample for Top-N tables (spawn join + finish).
type SubagentItem struct {
	DurationMs                   float64
	ID, Description, Type        string
	Status, Model                string
	ToolCalls, Turns, TokensUsed int
}

// ToolStat holds per-tool handler duration aggregates from events.jsonl.
type ToolStat struct {
	Name           string
	Count          int
	Success, Error int
	AvgMs, MedMs   float64
	MinMs, MaxMs   float64
}

// TaskAgg holds count and duration aggregates for background tasks or subagents.
type TaskAgg struct {
	Count        int
	AvgMs, MaxMs float64
}

// StatsSources records which sidecar files contributed to SessionStats.
type StatsSources struct {
	Summary, Signals, Events, Updates bool
	Warnings                          []string
}

// statsSignals is the subset of signals.json used by Stats (counts + latency).
type statsSignals struct {
	TurnCount              int `json:"turnCount"`
	UserMessageCount       int `json:"userMessageCount"`
	AssistantMessageCount  int `json:"assistantMessageCount"`
	ToolCallCount          int `json:"toolCallCount"`
	ToolFailureCount       int `json:"toolFailureCount"`
	ErrorCount             int `json:"errorCount"`
	CancellationCount      int `json:"cancellationCount"`
	CompactionCount        int `json:"compactionCount"`
	SessionDurationSeconds int `json:"sessionDurationSeconds"`
	AvgResponseTimeMs      int `json:"avgResponseTimeMs"`
	AvgTimeToFirstTokenMs  int `json:"avgTimeToFirstTokenMs"`
}

// Stats locates a session by exact UUID and aggregates stats from summary,
// signals.json, events.jsonl, and updates.jsonl. Missing optional files do not
// fail; they set Sources flags and append human-readable warnings.
func Stats(grokHome, sessionID string) (*SessionStats, error) {
	session, err := Find(grokHome, sessionID)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(session.Path)
	if err != nil {
		return nil, err
	}
	data = []byte(strings.TrimSpace(string(data)))
	if len(data) == 0 {
		return nil, sessionNotFoundError(sessionID)
	}

	var summary grokSummary
	if err := json.Unmarshal(data, &summary); err != nil {
		return nil, fmt.Errorf("parse summary for %s: %w", sessionID, err)
	}

	createdAt, _ := parseTimestamp(summary.CreatedAt)

	st := &SessionStats{
		ID:           session.ID,
		Title:        session.Title,
		CWD:          session.CWD,
		Model:        strings.TrimSpace(summary.CurrentModelID),
		Agent:        strings.TrimSpace(summary.AgentName),
		CreatedAt:    createdAt,
		LastActiveAt: session.LastActiveAt,
		Sources: StatsSources{
			Summary: true,
		},
	}

	sessionDir := filepath.Dir(session.Path)
	signalsPath := filepath.Join(sessionDir, "signals.json")
	eventsPath := filepath.Join(sessionDir, "events.jsonl")
	updatesPath := filepath.Join(sessionDir, "updates.jsonl")

	hasSignals := fileExists(signalsPath)
	hasEvents := fileExists(eventsPath)
	hasUpdates := fileExists(updatesPath)

	if hasSignals {
		if sig, ok := parseStatsSignalsFile(signalsPath); ok {
			st.Sources.Signals = true
			st.Turns = sig.TurnCount
			st.UserMessages = sig.UserMessageCount
			st.AssistantMessages = sig.AssistantMessageCount
			st.ToolCalls = sig.ToolCallCount
			st.ToolFailed = sig.ToolFailureCount
			st.Errors = sig.ErrorCount
			st.Cancellations = sig.CancellationCount
			st.Compactions = sig.CompactionCount
			st.SessionDurationSec = sig.SessionDurationSeconds
			st.AvgResponseMs = sig.AvgResponseTimeMs
			st.AvgTTFTMs = sig.AvgTimeToFirstTokenMs
		} else {
			// File exists but unreadable — treat as missing for Sources.
			st.Sources.Warnings = append(st.Sources.Warnings,
				"signals.json present but unreadable; session rollups omitted")
		}
	} else {
		st.Sources.Warnings = append(st.Sources.Warnings,
			"signals.json missing; session rollups omitted")
	}

	if hasEvents {
		if tools, completed, started, ok := parseEventsJSONL(eventsPath); ok {
			st.Sources.Events = true
			st.Tools = tools
			st.ToolCompleted = completed
			// Prefer signals toolCallCount when signals were loaded.
			if !st.Sources.Signals {
				st.ToolCalls = started
			}
		} else {
			st.Sources.Warnings = append(st.Sources.Warnings,
				"events.jsonl present but unreadable; per-tool duration omitted")
		}
	} else {
		st.Sources.Warnings = append(st.Sources.Warnings,
			"events.jsonl missing; per-tool duration omitted")
	}

	if hasUpdates {
		thinking, bg, sub, bgItems, subItems, ok := parseUpdatesJSONL(updatesPath)
		if ok {
			st.Sources.Updates = true
			st.ThinkingBlocks = thinking
			if bg != nil && bg.Count > 0 {
				st.BackgroundTasks = bg
			}
			if sub != nil && sub.Count > 0 {
				st.Subagents = sub
			}
			st.BackgroundTaskItems = bgItems
			st.SubagentItems = subItems
		} else {
			st.Sources.Warnings = append(st.Sources.Warnings,
				"updates.jsonl present but unreadable; thinking / background task sections omitted")
		}
	} else {
		st.Sources.Warnings = append(st.Sources.Warnings,
			"updates.jsonl missing; thinking / background task sections omitted")
	}

	return st, nil
}

// FormatStatsOptions configures human stats text (pretty durations, color, tops).
type FormatStatsOptions struct {
	Home      string
	Now       time.Time
	ColorMode string // "auto" | "always" | "never"; empty → treat as "never" in tests
	TopN      int    // default 5 at CLI; 0 hides Top tools/background/subagent sections
}

// FormatStatsText is a thin wrapper for callers that do not pass color/top options.
// Uses ColorMode "never" and TopN 5.
func FormatStatsText(stats *SessionStats, home string, now time.Time) string {
	return FormatStatsTextOpts(stats, FormatStatsOptions{
		Home:      home,
		Now:       now,
		ColorMode: "never",
		TopN:      5,
	})
}

// FormatStatsTextOpts renders SessionStats as sectioned human-readable text.
// Optional tool / background / subagent sections are omitted when empty.
// Pretty durations on all duration fields; tool table sorted by N desc then
// name; Top-N sections when TopN > 0; ANSI when shouldColor(ColorMode).
// Trailing newline is trimmed (CLI may re-add via fmt.Println).
func FormatStatsTextOpts(stats *SessionStats, opts FormatStatsOptions) string {
	if stats == nil {
		return ""
	}
	_ = opts.Home
	_ = opts.Now

	colorMode := strings.TrimSpace(opts.ColorMode)
	if colorMode == "" {
		colorMode = "never"
	}
	useColor := shouldColor(colorMode)
	topN := opts.TopN

	const (
		reset = "\x1b[0m"
		dim   = "\x1b[2m"
		green = "\x1b[32m"
		red   = "\x1b[31m"
	)
	paint := func(code, s string) string {
		if !useColor || s == "" {
			return s
		}
		return code + s + reset
	}
	section := func(title string) string {
		return paint(dim, title)
	}
	label := func(s string) string {
		return paint(dim, s)
	}
	numIfErr := func(n int, s string) string {
		if useColor && n > 0 {
			return paint(red, s)
		}
		return s
	}

	var b strings.Builder
	fmt.Fprintf(&b, "%s %s\n", label("Session:"), stats.ID)

	title := strings.TrimSpace(stats.Title)
	if title == "" {
		title = "(untitled)"
	}
	fmt.Fprintf(&b, "%s %s\n", label("Title:"), title)

	if stats.CWD != "" {
		fmt.Fprintf(&b, "%s %s\n", label("CWD:"), stats.CWD)
	}
	if stats.Model != "" {
		fmt.Fprintf(&b, "%s %s\n", label("Model:"), stats.Model)
	}
	if stats.Agent != "" {
		fmt.Fprintf(&b, "%s %s\n", label("Agent:"), stats.Agent)
	}

	fmt.Fprintf(&b, "\n%s\n", section("Counts"))
	fmt.Fprintf(&b, "  %s %d\n", label("Turns:"), stats.Turns)
	fmt.Fprintf(&b, "  %s %d\n", label("User messages:"), stats.UserMessages)
	fmt.Fprintf(&b, "  %s %d\n", label("Assistant messages:"), stats.AssistantMessages)
	fmt.Fprintf(&b, "  %s %d\n", label("Thinking blocks:"), stats.ThinkingBlocks)
	fmt.Fprintf(&b, "  %s %d\n", label("Tool calls:"), stats.ToolCalls)
	fmt.Fprintf(&b, "  %s %d\n", label("Tool completed:"), stats.ToolCompleted)
	fmt.Fprintf(&b, "  %s %s\n", label("Tool failed:"),
		numIfErr(stats.ToolFailed, fmt.Sprintf("%d", stats.ToolFailed)))
	fmt.Fprintf(&b, "  %s %d\n", label("Compactions:"), stats.Compactions)
	fmt.Fprintf(&b, "  %s %d\n", label("Cancellations:"), stats.Cancellations)
	fmt.Fprintf(&b, "  %s %s\n", label("Errors:"),
		numIfErr(stats.Errors, fmt.Sprintf("%d", stats.Errors)))

	fmt.Fprintf(&b, "\n%s\n", section("Latency"))
	fmt.Fprintf(&b, "  %s %s\n", label("Duration:"), formatPrettySec(stats.SessionDurationSec))
	fmt.Fprintf(&b, "  %s %s\n", label("Avg response:"), formatPrettyMs(float64(stats.AvgResponseMs)))
	fmt.Fprintf(&b, "  %s %s\n", label("Avg time-to-first:"), formatPrettyMs(float64(stats.AvgTTFTMs)))

	if len(stats.Tools) > 0 {
		fmt.Fprintf(&b, "\n%s\n", section("Tool handler time"))
		writeToolTable(&b, stats.Tools, useColor)
	}

	if topN > 0 && len(stats.Tools) > 0 {
		fmt.Fprintf(&b, "\n%s\n", section("Top tools by total handler time"))
		writeTopTools(&b, stats.Tools, topN)
	}

	if stats.BackgroundTasks != nil && stats.BackgroundTasks.Count > 0 {
		fmt.Fprintf(&b, "\n%s\n", section("Background tasks"))
		fmt.Fprintf(&b, "  %s %d\n", label("Count:"), stats.BackgroundTasks.Count)
		fmt.Fprintf(&b, "  %s %s\n", label("Avg:"), formatPrettyMs(stats.BackgroundTasks.AvgMs))
		fmt.Fprintf(&b, "  %s %s\n", label("Max:"), formatPrettyMs(stats.BackgroundTasks.MaxMs))
	}

	if topN > 0 && len(stats.BackgroundTaskItems) > 0 {
		fmt.Fprintf(&b, "\n%s\n", section("Top background tasks"))
		writeTopBackgroundTasks(&b, stats.BackgroundTaskItems, topN)
	}

	if stats.Subagents != nil && stats.Subagents.Count > 0 {
		fmt.Fprintf(&b, "\n%s\n", section("Subagents"))
		fmt.Fprintf(&b, "  %s %d\n", label("Count:"), stats.Subagents.Count)
		fmt.Fprintf(&b, "  %s %s\n", label("Avg:"), formatPrettyMs(stats.Subagents.AvgMs))
		fmt.Fprintf(&b, "  %s %s\n", label("Max:"), formatPrettyMs(stats.Subagents.MaxMs))
	}

	if topN > 0 && len(stats.SubagentItems) > 0 {
		fmt.Fprintf(&b, "\n%s\n", section("Top subagents"))
		writeTopSubagents(&b, stats.SubagentItems, topN)
	}

	fmt.Fprintf(&b, "\n%s\n", section("Sources"))
	writeSourceLine(&b, "summary", stats.Sources.Summary, useColor, green, reset)
	writeSourceLine(&b, "signals", stats.Sources.Signals, useColor, green, reset)
	writeSourceLine(&b, "events", stats.Sources.Events, useColor, green, reset)
	writeSourceLine(&b, "updates", stats.Sources.Updates, useColor, green, reset)
	if len(stats.Sources.Warnings) > 0 {
		fmt.Fprintf(&b, "  %s\n", label("warnings:"))
		for _, w := range stats.Sources.Warnings {
			fmt.Fprintf(&b, "    - %s\n", w)
		}
	}

	return strings.TrimRight(b.String(), "\n")
}

func writeSourceLine(b *strings.Builder, name string, ok bool, useColor bool, green, reset string) {
	mark := "no"
	if ok {
		mark = "yes"
		if useColor {
			mark = green + "yes" + reset
		}
	}
	fmt.Fprintf(b, "  %s: %s\n", name, mark)
}

func writeToolTable(b *strings.Builder, tools []ToolStat, useColor bool) {
	sorted := append([]ToolStat(nil), tools...)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Count != sorted[j].Count {
			return sorted[i].Count > sorted[j].Count
		}
		return sorted[i].Name < sorted[j].Name
	})

	nameW := len("NAME")
	for _, t := range sorted {
		if len(t.Name) > nameW {
			nameW = len(t.Name)
		}
	}
	if nameW < 4 {
		nameW = 4
	}

	const (
		reset = "\x1b[0m"
		red   = "\x1b[31m"
	)

	// Header: NAME  N  SUCCESS  ERROR  AVG  MED  MIN  MAX
	fmt.Fprintf(b, "  %-*s  %5s  %7s  %5s  %8s  %8s  %8s  %8s\n",
		nameW, "NAME", "N", "SUCCESS", "ERROR", "AVG", "MED", "MIN", "MAX")

	for _, t := range sorted {
		errStr := fmt.Sprintf("%5d", t.Error)
		if useColor && t.Error > 0 {
			// right-align digits inside ANSI so column width stays stable
			errStr = fmt.Sprintf("%s%5d%s", red, t.Error, reset)
		}
		fmt.Fprintf(b, "  %-*s  %5d  %7d  %s  %8s  %8s  %8s  %8s\n",
			nameW, t.Name, t.Count, t.Success, errStr,
			formatPrettyMs(t.AvgMs), formatPrettyMs(t.MedMs),
			formatPrettyMs(t.MinMs), formatPrettyMs(t.MaxMs))
	}
}

func writeTopTools(b *strings.Builder, tools []ToolStat, topN int) {
	type ranked struct {
		name  string
		total float64
		n     int
		avg   float64
	}
	items := make([]ranked, 0, len(tools))
	for _, t := range tools {
		items = append(items, ranked{
			name:  t.Name,
			total: float64(t.Count) * t.AvgMs,
			n:     t.Count,
			avg:   t.AvgMs,
		})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].total != items[j].total {
			return items[i].total > items[j].total
		}
		return items[i].name < items[j].name
	})
	if topN > len(items) {
		topN = len(items)
	}
	nameW := len("NAME")
	for i := 0; i < topN; i++ {
		if len(items[i].name) > nameW {
			nameW = len(items[i].name)
		}
	}
	fmt.Fprintf(b, "  %2s  %-*s  %8s  %5s  %8s\n", "#", nameW, "NAME", "TOTAL", "N", "AVG")
	for i := 0; i < topN; i++ {
		it := items[i]
		fmt.Fprintf(b, "  %2d  %-*s  %8s  %5d  %8s\n",
			i+1, nameW, it.name, formatPrettyMs(it.total), it.n, formatPrettyMs(it.avg))
	}
}

// writeTopBackgroundTasks prints:
//
//	#  DURATION  EXIT  COMMAND
//
// COMMAND is display-truncated at 200 runes; EXIT is integer or "-" if nil.
func writeTopBackgroundTasks(b *strings.Builder, items []BackgroundTaskItem, topN int) {
	sorted := append([]BackgroundTaskItem(nil), items...)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].DurationMs != sorted[j].DurationMs {
			return sorted[i].DurationMs > sorted[j].DurationMs
		}
		return sorted[i].Command < sorted[j].Command
	})
	if topN > len(sorted) {
		topN = len(sorted)
	}
	fmt.Fprintf(b, "  %2s  %8s  %4s  %s\n", "#", "DURATION", "EXIT", "COMMAND")
	for i := 0; i < topN; i++ {
		it := sorted[i]
		exitStr := "-"
		if it.ExitCode != nil {
			exitStr = strconv.Itoa(*it.ExitCode)
		}
		cmd := truncateRunes(it.Command, 200)
		if cmd == "" {
			cmd = "(unnamed)"
		}
		fmt.Fprintf(b, "  %2d  %8s  %4s  %s\n",
			i+1, formatPrettyMs(it.DurationMs), exitStr, cmd)
	}
}

// writeTopSubagents prints:
//
//	#  DURATION  STATUS     TYPE              TOOLS  TURNS  DESC
//
// DESC is description if set else ID; display-truncated at 64 runes.
func writeTopSubagents(b *strings.Builder, items []SubagentItem, topN int) {
	sorted := append([]SubagentItem(nil), items...)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].DurationMs != sorted[j].DurationMs {
			return sorted[i].DurationMs > sorted[j].DurationMs
		}
		// Stable tie-break: description then ID.
		di, dj := sorted[i].Description, sorted[j].Description
		if di == "" {
			di = sorted[i].ID
		}
		if dj == "" {
			dj = sorted[j].ID
		}
		return di < dj
	})
	if topN > len(sorted) {
		topN = len(sorted)
	}
	// Locked contiguous header (STATUS width 9 left, TYPE width 16 left).
	fmt.Fprintf(b, "  %2s  %8s  %-9s  %-16s  %5s  %5s  %s\n",
		"#", "DURATION", "STATUS", "TYPE", "TOOLS", "TURNS", "DESC")
	for i := 0; i < topN; i++ {
		it := sorted[i]
		desc := strings.TrimSpace(it.Description)
		if desc == "" {
			desc = strings.TrimSpace(it.ID)
		}
		desc = truncateRunes(desc, 64)
		if desc == "" {
			desc = "(unnamed)"
		}
		fmt.Fprintf(b, "  %2d  %8s  %-9s  %-16s  %5d  %5d  %s\n",
			i+1, formatPrettyMs(it.DurationMs), it.Status, it.Type,
			it.ToolCalls, it.Turns, desc)
	}
}

// formatPrettyMs formats a millisecond duration for human text.
// 0 → 0ms; <1s → integer ms; ≥1s non-whole → one decimal s; whole seconds and
// larger use compact units (3s, 2m, 2m57s, 2h39m12s).
func formatPrettyMs(ms float64) string {
	if ms < 0 || math.IsNaN(ms) {
		ms = 0
	}
	if ms < 1000 {
		return fmt.Sprintf("%dms", int(math.Round(ms)))
	}

	totalSecF := ms / 1000.0
	// Use rounded whole seconds for multi-unit breakdown.
	totalSec := int64(math.Round(totalSecF))
	if totalSec < 0 {
		totalSec = 0
	}

	if totalSec >= 3600 {
		h := totalSec / 3600
		rem := totalSec % 3600
		m := rem / 60
		s := rem % 60
		switch {
		case m == 0 && s == 0:
			return fmt.Sprintf("%dh", h)
		case s == 0:
			return fmt.Sprintf("%dh%dm", h, m)
		case m == 0:
			return fmt.Sprintf("%dh%ds", h, s)
		default:
			return fmt.Sprintf("%dh%dm%ds", h, m, s)
		}
	}

	if totalSec >= 60 {
		m := totalSec / 60
		s := totalSec % 60
		if s == 0 {
			return fmt.Sprintf("%dm", m)
		}
		return fmt.Sprintf("%dm%ds", m, s)
	}

	// 1s .. 59.x s
	// Prefer whole-second form when the ms value is (nearly) a whole second.
	if math.Abs(totalSecF-math.Round(totalSecF)) < 1e-6 {
		return fmt.Sprintf("%ds", int(math.Round(totalSecF)))
	}
	oneDec := math.Round(totalSecF*10) / 10
	if math.Abs(oneDec-math.Round(oneDec)) < 1e-9 {
		return fmt.Sprintf("%ds", int(math.Round(oneDec)))
	}
	return fmt.Sprintf("%.1fs", oneDec)
}

func formatPrettySec(sec int) string {
	return formatPrettyMs(float64(sec) * 1000)
}

func truncateRunes(s string, max int) string {
	if max <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	if max <= 1 {
		return string(r[:max])
	}
	return string(r[:max-1]) + "…"
}

func sourceMark(ok bool) string {
	if ok {
		return "yes"
	}
	return "no"
}

func parseStatsSignalsFile(path string) (statsSignals, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return statsSignals{}, false
	}
	data = []byte(strings.TrimSpace(string(data)))
	if len(data) == 0 {
		return statsSignals{}, false
	}
	var sig statsSignals
	if err := json.Unmarshal(data, &sig); err != nil {
		return statsSignals{}, false
	}
	return sig, true
}

type toolAccum struct {
	name      string
	durations []float64
	success   int
	errCount  int
}

func parseEventsJSONL(path string) (tools []ToolStat, completed, started int, ok bool) {
	f, err := os.Open(path)
	if err != nil {
		return nil, 0, 0, false
	}
	defer f.Close()

	byName := make(map[string]*toolAccum)
	var order []string

	scanner := bufio.NewScanner(f)
	// Allow long event lines.
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var ev map[string]any
		if err := json.Unmarshal([]byte(line), &ev); err != nil {
			continue
		}
		typ, _ := ev["type"].(string)
		switch typ {
		case "tool_started":
			started++
		case "tool_completed":
			completed++
			name, _ := ev["tool_name"].(string)
			if name == "" {
				name = "(unknown)"
			}
			acc, exists := byName[name]
			if !exists {
				acc = &toolAccum{name: name}
				byName[name] = acc
				order = append(order, name)
			}
			if d, ok := asFloat(ev["duration_ms"]); ok {
				acc.durations = append(acc.durations, d)
			}
			outcome, _ := ev["outcome"].(string)
			switch strings.ToLower(strings.TrimSpace(outcome)) {
			case "error", "failed", "failure":
				acc.errCount++
			default:
				// success and anything else count as success for the Success field
				// when outcome is explicitly "success"; unknown still increments Count
				// via duration. Prefer explicit success.
				if strings.EqualFold(outcome, "success") || outcome == "" {
					acc.success++
				} else {
					// non-success non-error outcomes still count as neither? fixtures
					// only use success/error. Treat non-error as success.
					acc.success++
				}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, 0, 0, false
	}

	// Sort tool names for stable output.
	sort.Strings(order)
	tools = make([]ToolStat, 0, len(order))
	for _, name := range order {
		acc := byName[name]
		ts := ToolStat{
			Name:    name,
			Count:   len(acc.durations),
			Success: acc.success,
			Error:   acc.errCount,
		}
		// If durations empty but we had outcomes, Count should still reflect calls.
		// Fixtures always include duration_ms; use max of durations vs success+error.
		if n := acc.success + acc.errCount; n > ts.Count {
			ts.Count = n
		}
		if len(acc.durations) > 0 {
			ts.MinMs, ts.MaxMs, ts.AvgMs, ts.MedMs = durationStats(acc.durations)
		}
		tools = append(tools, ts)
	}
	return tools, completed, started, true
}

func durationStats(durs []float64) (min, max, avg, med float64) {
	if len(durs) == 0 {
		return 0, 0, 0, 0
	}
	sorted := append([]float64(nil), durs...)
	sort.Float64s(sorted)
	min = sorted[0]
	max = sorted[len(sorted)-1]
	var sum float64
	for _, d := range sorted {
		sum += d
	}
	avg = sum / float64(len(sorted))
	med = medianSorted(sorted)
	return min, max, avg, med
}

func medianSorted(sorted []float64) float64 {
	n := len(sorted)
	if n == 0 {
		return 0
	}
	if n%2 == 1 {
		return sorted[n/2]
	}
	return (sorted[n/2-1] + sorted[n/2]) / 2
}

type spawnMeta struct {
	Description string
	Type        string
	Model       string
}

func parseUpdatesJSONL(path string) (thinking int, bg, sub *TaskAgg, bgItems []BackgroundTaskItem, subItems []SubagentItem, ok bool) {
	f, err := os.Open(path)
	if err != nil {
		return 0, nil, nil, nil, nil, false
	}
	defer f.Close()

	inThoughtRun := false
	var bgDurations []float64
	var subDurations []float64

	// task_backgrounded: task_id → description
	taskDescByID := make(map[string]string)
	// subagent_spawned: id → description, type, model
	spawnByID := make(map[string]spawnMeta)

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var raw map[string]any
		if err := json.Unmarshal([]byte(line), &raw); err != nil {
			continue
		}

		update := extractSessionUpdate(raw)
		if update == nil {
			continue
		}

		kind, _ := update["sessionUpdate"].(string)
		switch kind {
		case "agent_thought_chunk":
			if !inThoughtRun {
				thinking++
				inThoughtRun = true
			}
		case "task_backgrounded":
			inThoughtRun = false
			if id, ok := stringField(update, "task_id"); ok {
				if desc, _ := stringField(update, "description"); desc != "" {
					taskDescByID[id] = desc
				}
			}
		case "task_completed":
			inThoughtRun = false
			if ms, ok := taskWallMs(update); ok {
				bgDurations = append(bgDurations, ms)
				item := BackgroundTaskItem{
					DurationMs: ms,
					Command:    taskSnapshotString(update, "command"),
					Kind:       taskSnapshotString(update, "kind"),
					CWD:        taskSnapshotString(update, "cwd"),
				}
				if code, ok := taskSnapshotExitCode(update); ok {
					c := code
					item.ExitCode = &c
				}
				// Join description by task_id when present.
				if id, ok := stringField(update, "task_id"); ok {
					if d, found := taskDescByID[id]; found {
						item.Description = d
					}
				}
				if item.Description == "" {
					if d := taskSnapshotString(update, "description"); d != "" {
						item.Description = d
					}
				}
				bgItems = append(bgItems, item)
			}
		case "subagent_spawned":
			inThoughtRun = false
			id, _ := stringField(update, "subagent_id")
			if id == "" {
				id, _ = stringField(update, "id")
			}
			if id == "" {
				break
			}
			meta := spawnMeta{}
			meta.Description, _ = stringField(update, "description")
			// Real Grok wire uses subagent_type; accept legacy "type" as fallback.
			meta.Type, _ = stringField(update, "subagent_type")
			if meta.Type == "" {
				meta.Type, _ = stringField(update, "type")
			}
			meta.Model, _ = stringField(update, "model")
			spawnByID[id] = meta
		case "subagent_finished":
			inThoughtRun = false
			if d, ok := asFloat(update["duration_ms"]); ok {
				subDurations = append(subDurations, d)
				item := SubagentItem{DurationMs: d}
				if id, ok := stringField(update, "subagent_id"); ok {
					item.ID = id
				} else if id, ok := stringField(update, "child_session_id"); ok {
					item.ID = id
				} else if id, ok := stringField(update, "id"); ok {
					item.ID = id
				}
				if s, ok := stringField(update, "status"); ok {
					item.Status = s
				}
				if n, ok := asInt(update["tool_calls"]); ok {
					item.ToolCalls = n
				}
				if n, ok := asInt(update["turns"]); ok {
					item.Turns = n
				}
				if n, ok := asInt(update["tokens_used"]); ok {
					item.TokensUsed = n
				}
				// Finish may carry its own description (legacy).
				if desc, ok := stringField(update, "description"); ok {
					item.Description = desc
				}
				// Join spawn meta by id.
				if item.ID != "" {
					if meta, found := spawnByID[item.ID]; found {
						if item.Description == "" {
							item.Description = meta.Description
						}
						if item.Type == "" {
							item.Type = meta.Type
						}
						if item.Model == "" {
							item.Model = meta.Model
						}
					}
				}
				subItems = append(subItems, item)
			}
		default:
			// Any other sessionUpdate ends a thought run.
			if kind != "" {
				inThoughtRun = false
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return 0, nil, nil, nil, nil, false
	}

	if len(bgDurations) > 0 {
		bg = &TaskAgg{Count: len(bgDurations)}
		var sum, max float64
		for _, d := range bgDurations {
			sum += d
			if d > max {
				max = d
			}
		}
		bg.AvgMs = sum / float64(len(bgDurations))
		bg.MaxMs = max
	}
	if len(subDurations) > 0 {
		sub = &TaskAgg{Count: len(subDurations)}
		var sum, max float64
		for _, d := range subDurations {
			sum += d
			if d > max {
				max = d
			}
		}
		sub.AvgMs = sum / float64(len(subDurations))
		sub.MaxMs = max
	}
	return thinking, bg, sub, bgItems, subItems, true
}

func stringField(m map[string]any, key string) (string, bool) {
	if m == nil {
		return "", false
	}
	s, ok := m[key].(string)
	if !ok {
		return "", false
	}
	s = strings.TrimSpace(s)
	if s == "" {
		return "", false
	}
	return s, true
}

func taskSnapshotString(update map[string]any, key string) string {
	snap, _ := update["task_snapshot"].(map[string]any)
	if snap == nil {
		return ""
	}
	s, _ := stringField(snap, key)
	return s
}

func taskSnapshotExitCode(update map[string]any) (int, bool) {
	snap, _ := update["task_snapshot"].(map[string]any)
	if snap == nil {
		return 0, false
	}
	return asInt(snap["exit_code"])
}

func asInt(v any) (int, bool) {
	switch n := v.(type) {
	case int:
		return n, true
	case int64:
		return int(n), true
	case float64:
		return int(n), true
	case float32:
		return int(n), true
	case json.Number:
		i, err := n.Int64()
		return int(i), err == nil
	case string:
		i, err := strconv.Atoi(strings.TrimSpace(n))
		return i, err == nil
	default:
		return 0, false
	}
}

// extractSessionUpdate returns the map that holds sessionUpdate, either the
// top-level object (flat) or params.update (nested envelope).
func extractSessionUpdate(raw map[string]any) map[string]any {
	if _, ok := raw["sessionUpdate"]; ok {
		return raw
	}
	params, _ := raw["params"].(map[string]any)
	if params == nil {
		return nil
	}
	update, _ := params["update"].(map[string]any)
	if update == nil {
		return nil
	}
	if _, ok := update["sessionUpdate"]; ok {
		return update
	}
	return nil
}

func taskWallMs(update map[string]any) (float64, bool) {
	snap, _ := update["task_snapshot"].(map[string]any)
	if snap == nil {
		return 0, false
	}
	start, ok1 := epochMs(snap["start_time"])
	end, ok2 := epochMs(snap["end_time"])
	if !ok1 || !ok2 {
		return 0, false
	}
	return end - start, true
}

func epochMs(v any) (float64, bool) {
	m, ok := v.(map[string]any)
	if !ok {
		return 0, false
	}
	secs, okS := asFloat(m["secs_since_epoch"])
	if !okS {
		return 0, false
	}
	nanos, _ := asFloat(m["nanos_since_epoch"])
	return secs*1000 + nanos/1e6, true
}

func asFloat(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case json.Number:
		f, err := n.Float64()
		return f, err == nil
	case string:
		// rare
		var f float64
		_, err := fmt.Sscanf(n, "%f", &f)
		return f, err == nil && !math.IsNaN(f)
	default:
		return 0, false
	}
}
