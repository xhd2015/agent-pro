package agenttty

import (
	"regexp"
	"strings"
)

// GrokSectionKind is a vertical TUI phase in top→bottom order.
type GrokSectionKind string

const (
	GrokSectionHeader              GrokSectionKind = "header"
	GrokSectionActivityStrip       GrokSectionKind = "activity_strip"
	GrokSectionBody                GrokSectionKind = "body"
	GrokSectionWorkedFor           GrokSectionKind = "worked_for"
	GrokSectionRecap               GrokSectionKind = "recap"
	GrokSectionRunningIndicator    GrokSectionKind = "running_indicator"
	GrokSectionStatusAboveComposer GrokSectionKind = "status_above_composer" // Waiting/Responding/Thinking…/tool spinner
	GrokSectionTip                 GrokSectionKind = "tip"                   // e.g. "Tight on space? Try /compact-mode"
	GrokSectionComposer            GrokSectionKind = "composer"
	GrokSectionHelpFooter          GrokSectionKind = "help_footer"
)

// GrokSection is one contiguous span of snapshot lines.
type GrokSection struct {
	Kind  GrokSectionKind
	Lines []string
}

// GrokFrame is the stateful parse of a Grok TUI snapshot (ordered sections).
type GrokFrame struct {
	Sections []GrokSection
}

// WorkedForMeta is typed crumbs for a WorkedFor section line.
type WorkedForMeta struct {
	Duration string
	Stop     bool
	Hooks    string // optional raw "[hooks: N]"
}

var (
	reGrokQuota          = regexp.MustCompile(`(?i)\d+(?:\.\d+)?\s*[KMGT]?B?\s*/\s*\d+(?:\.\d+)?\s*[KMGT]?B?\b`)
	reGrokWorkedForStop  = regexp.MustCompile(`(?i)^Worked for\s+(\S+)\s+stop(?:\s+(\[hooks:\s*\d+\]))?\s*$`)
	reGrokWorkedForLoose = regexp.MustCompile(`(?i)^Worked for\s+(\S+)\b`)
	reGrokRunningCmd     = regexp.MustCompile(`(?i)\d+\s+commands?\s+still\s+running`)
	reGrokLiveThinking   = regexp.MustCompile(`(?i)(?:◆\s*)?Thinking(?:…|\.\.\.)|(?:[⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏]\s*)Thinking`)
	reGrokWaitingStatus  = regexp.MustCompile(`(?i)waiting for (a )?response`)
	reGrokBrailleSpinner = regexp.MustCompile(`[⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏]`)
	reGrokShortDuration  = regexp.MustCompile(`\d+(?:\.\d+)?s`)
)

type grokParsePhase int

const (
	grokPhaseHeader grokParsePhase = iota
	grokPhaseActivity
	grokPhaseBody
	grokPhaseAfterWorkedFor // recap | running | tip | composer
	grokPhaseTip
	grokPhaseComposer
	grokPhaseHelp
)

// ParseGrokFrame segments plain snapshot text into ordered TUI sections.
// Expected phases: Header → optional ActivityStrip → Body → optional WorkedFor
// → optional Recap → optional RunningIndicator → optional StatusAboveComposer
// → optional Tip → Composer → HelpFooter.
func ParseGrokFrame(plain string) GrokFrame {
	lines := strings.Split(plain, "\n")
	var sections []GrokSection
	phase := grokPhaseHeader
	var cur *GrokSection

	flush := func() {
		if cur == nil {
			return
		}
		sections = append(sections, *cur)
		cur = nil
	}
	start := func(kind GrokSectionKind, line string) {
		flush()
		cur = &GrokSection{Kind: kind, Lines: []string{line}}
	}
	add := func(line string) {
		if cur == nil {
			start(GrokSectionBody, line)
			return
		}
		cur.Lines = append(cur.Lines, line)
	}
	ensureKind := func(kind GrokSectionKind, line string) {
		if cur != nil && cur.Kind == kind {
			add(line)
			return
		}
		start(kind, line)
	}

	for _, line := range lines {
		trim := strings.TrimSpace(line)
		lower := strings.ToLower(trim)

		switch phase {
		case grokPhaseHeader:
			if isGrokHeaderLine(trim) {
				ensureKind(GrokSectionHeader, line)
				phase = grokPhaseActivity
				continue
			}
			if isGrokActivityStripLine(trim, lower) {
				start(GrokSectionActivityStrip, line)
				phase = grokPhaseActivity
				continue
			}
			if handled, next := grokOpenPostHeader(trim, lower, line, start); handled {
				phase = next
				continue
			}
			if trim != "" {
				start(GrokSectionBody, line)
				phase = grokPhaseBody
			}

		case grokPhaseActivity:
			if isGrokActivityStripLine(trim, lower) {
				ensureKind(GrokSectionActivityStrip, line)
				continue
			}
			if isGrokHeaderLine(trim) && cur != nil && cur.Kind == GrokSectionHeader {
				add(line)
				continue
			}
			if handled, next := grokOpenPostHeader(trim, lower, line, start); handled {
				phase = next
				continue
			}
			if trim != "" || cur == nil {
				ensureKind(GrokSectionBody, line)
			} else {
				add(line)
			}
			phase = grokPhaseBody

		case grokPhaseBody:
			if handled, next := grokOpenFromBody(trim, lower, line, start); handled {
				phase = next
				continue
			}
			if isGrokActivityStripLine(trim, lower) {
				start(GrokSectionActivityStrip, line)
				phase = grokPhaseActivity
				continue
			}
			ensureKind(GrokSectionBody, line)

		case grokPhaseAfterWorkedFor:
			if isGrokRecapOpener(trim, lower) || (cur != nil && cur.Kind == GrokSectionRecap && !isGrokComposerOpener(trim) && !isGrokHelpFooterLine(trim, lower) && !isGrokRunningIndicatorLine(trim, lower) && !isGrokStatusAboveComposerLine(trim, lower) && !isGrokTipLine(trim, lower)) {
				if isGrokRecapOpener(trim, lower) || (cur != nil && cur.Kind == GrokSectionRecap) {
					ensureKind(GrokSectionRecap, line)
					continue
				}
			}
			if isGrokRunningIndicatorLine(trim, lower) {
				start(GrokSectionRunningIndicator, line)
				phase = grokPhaseTip
				continue
			}
			if isGrokStatusAboveComposerLine(trim, lower) {
				start(GrokSectionStatusAboveComposer, line)
				phase = grokPhaseTip
				continue
			}
			if isGrokTipLine(trim, lower) {
				start(GrokSectionTip, line)
				phase = grokPhaseTip
				continue
			}
			if isGrokComposerOpener(trim) {
				start(GrokSectionComposer, line)
				phase = grokPhaseComposer
				continue
			}
			if isGrokHelpFooterLine(trim, lower) {
				start(GrokSectionHelpFooter, line)
				phase = grokPhaseHelp
				continue
			}
			// Padding between recap and composer stays with recap when open.
			if cur != nil && cur.Kind == GrokSectionRecap {
				add(line)
				continue
			}
			if trim != "" {
				ensureKind(GrokSectionBody, line)
			}

		case grokPhaseTip:
			if isGrokStatusAboveComposerLine(trim, lower) {
				ensureKind(GrokSectionStatusAboveComposer, line)
				continue
			}
			if isGrokTipLine(trim, lower) {
				ensureKind(GrokSectionTip, line)
				continue
			}
			if isGrokComposerOpener(trim) {
				start(GrokSectionComposer, line)
				phase = grokPhaseComposer
				continue
			}
			if isGrokHelpFooterLine(trim, lower) {
				start(GrokSectionHelpFooter, line)
				phase = grokPhaseHelp
				continue
			}
			// Blank padding before composer stays with tip or status.
			if cur != nil && (cur.Kind == GrokSectionTip || cur.Kind == GrokSectionStatusAboveComposer || cur.Kind == GrokSectionRunningIndicator) && trim == "" {
				add(line)
				continue
			}
			if trim != "" {
				ensureKind(GrokSectionBody, line)
				phase = grokPhaseBody
			}

		case grokPhaseComposer:
			if isGrokHelpFooterLine(trim, lower) {
				start(GrokSectionHelpFooter, line)
				phase = grokPhaseHelp
				continue
			}
			if isGrokStatusAboveComposerLine(trim, lower) {
				start(GrokSectionStatusAboveComposer, line)
				phase = grokPhaseTip
				continue
			}
			if isGrokRunningIndicatorLine(trim, lower) {
				start(GrokSectionRunningIndicator, line)
				phase = grokPhaseTip
				continue
			}
			// Do not swallow busy/transcript lines into the composer box section
			// (synthetic fixtures may put "thinking about…" after the box line).
			if trim != "" && !isGrokComposerContinuationLine(trim) && grokSectionHasLiveThinking(trim) {
				start(GrokSectionBody, line)
				phase = grokPhaseBody
				continue
			}
			ensureKind(GrokSectionComposer, line)

		case grokPhaseHelp:
			ensureKind(GrokSectionHelpFooter, line)
		}
	}
	flush()
	return GrokFrame{Sections: sections}
}

func grokOpenPostHeader(trim, lower, line string, start func(GrokSectionKind, string)) (bool, grokParsePhase) {
	return grokOpenFromBody(trim, lower, line, start)
}

func grokOpenFromBody(trim, lower, line string, start func(GrokSectionKind, string)) (bool, grokParsePhase) {
	if meta, ok := parseGrokWorkedForLine(trim); ok && meta.Stop {
		start(GrokSectionWorkedFor, line)
		return true, grokPhaseAfterWorkedFor
	}
	if isGrokRecapOpener(trim, lower) {
		start(GrokSectionRecap, line)
		return true, grokPhaseAfterWorkedFor
	}
	if isGrokRunningIndicatorLine(trim, lower) {
		start(GrokSectionRunningIndicator, line)
		return true, grokPhaseTip
	}
	if isGrokStatusAboveComposerLine(trim, lower) {
		start(GrokSectionStatusAboveComposer, line)
		return true, grokPhaseTip
	}
	if isGrokTipLine(trim, lower) {
		start(GrokSectionTip, line)
		return true, grokPhaseTip
	}
	if isGrokComposerOpener(trim) {
		start(GrokSectionComposer, line)
		return true, grokPhaseComposer
	}
	if isGrokHelpFooterLine(trim, lower) {
		start(GrokSectionHelpFooter, line)
		return true, grokPhaseHelp
	}
	return false, grokPhaseBody
}

func isGrokHeaderLine(trim string) bool {
	if trim == "" {
		return false
	}
	if reGrokQuota.MatchString(trim) {
		return true
	}
	if strings.HasPrefix(trim, "~/") || strings.HasPrefix(trim, "⎇") {
		return true
	}
	return false
}

func isGrokActivityStripLine(trim, lower string) bool {
	if trim == "" {
		return false
	}
	if strings.Contains(lower, "▾ tasks") || (strings.Contains(lower, "tasks") && strings.Contains(trim, "▾")) {
		return true
	}
	if strings.Contains(lower, "⸬ task") || strings.HasPrefix(lower, "task ") {
		return true
	}
	if strings.Contains(lower, "sub-agent") || strings.Contains(lower, "subagent") {
		return true
	}
	if strings.Contains(lower, "long running") || strings.Contains(lower, "background terminal") {
		return true
	}
	if strings.Contains(lower, "starting session") {
		return true
	}
	return false
}

func parseGrokWorkedForLine(trim string) (WorkedForMeta, bool) {
	if trim == "" {
		return WorkedForMeta{}, false
	}
	compact := strings.Join(strings.Fields(trim), " ")
	if m := reGrokWorkedForStop.FindStringSubmatch(compact); len(m) >= 2 {
		meta := WorkedForMeta{Duration: m[1], Stop: true}
		if len(m) >= 3 {
			meta.Hooks = m[2]
		}
		return meta, true
	}
	if m := reGrokWorkedForLoose.FindStringSubmatch(compact); len(m) >= 2 {
		return WorkedForMeta{Duration: m[1], Stop: false}, true
	}
	return WorkedForMeta{}, false
}

func isGrokRecapOpener(trim, lower string) bool {
	if trim == "" {
		return false
	}
	if strings.Contains(lower, "◆ recap") || strings.Contains(lower, "♦ recap") {
		return true
	}
	return strings.Contains(trim, "Recap") && (strings.Contains(trim, "┃") || strings.Contains(trim, "◆"))
}

func isGrokRunningIndicatorLine(trim, lower string) bool {
	if trim == "" {
		return false
	}
	if reGrokRunningCmd.MatchString(lower) {
		return true
	}
	if strings.Contains(lower, "still running") && strings.Contains(lower, "interrupt") {
		return true
	}
	if strings.Contains(lower, "send a message to interrupt") {
		return true
	}
	return false
}

// isGrokStatusAboveComposerLine reports the live status row immediately above
// the composer: Waiting / Responding / Thinking… / foreground tool spinner.
// Transcript "┃ ◆ Thinking…" without a braille spinner stays in Body.
// "Starting session…" stays ActivityStrip / boot, not this section.
func isGrokStatusAboveComposerLine(trim, lower string) bool {
	if trim == "" {
		return false
	}
	if strings.Contains(lower, "starting session") {
		return false
	}
	if isGrokRunningIndicatorLine(trim, lower) {
		return false
	}
	if reGrokWaitingStatus.MatchString(trim) {
		return true
	}
	if strings.Contains(trim, "Responding…") || strings.Contains(trim, "Responding...") {
		return true
	}
	// Live Thinking… status row: braille spinner + Thinking… + duration.
	if (strings.Contains(trim, "Thinking…") || strings.Contains(trim, "Thinking...")) &&
		!strings.Contains(lower, "thought for") &&
		reGrokBrailleSpinner.MatchString(trim) &&
		reGrokShortDuration.MatchString(trim) {
		return true
	}
	// Foreground tool-run spinner (e.g. "⠙ Clear cache… 4.1s … [stop]").
	if reGrokBrailleSpinner.MatchString(trim) &&
		reGrokShortDuration.MatchString(trim) &&
		(strings.Contains(trim, "…") || strings.Contains(trim, "...")) &&
		!strings.Contains(lower, "thinking") &&
		!reGrokWaitingStatus.MatchString(trim) &&
		!strings.Contains(lower, "responding") {
		return true
	}
	return false
}

func isGrokTipLine(trim, lower string) bool {
	if trim == "" {
		return false
	}
	if strings.Contains(lower, "tight on space") {
		return true
	}
	if strings.Contains(lower, "/compact-mode") || strings.Contains(lower, "try /compact") {
		return true
	}
	return false
}

// isGrokComposerOpener reports the live boxed input chrome only.
// Transcript user lines like "❯ list files…" must stay in Body, not open Composer.
func isGrokComposerOpener(trim string) bool {
	if trim == "" {
		return false
	}
	if strings.HasPrefix(trim, "╭") {
		return true
	}
	// Boxed live composer: "│ ❯ … │" (box drawing + glyph), not bare transcript ❯.
	return strings.Contains(trim, "│") && strings.Contains(trim, "❯")
}

// isGrokComposerContinuationLine reports lines that belong inside an open composer
// section (box borders / model footer), not transcript or status chrome.
func isGrokComposerContinuationLine(trim string) bool {
	if trim == "" {
		return true
	}
	if strings.HasPrefix(trim, "╭") || strings.HasPrefix(trim, "╰") || strings.HasPrefix(trim, "│") {
		return true
	}
	lower := strings.ToLower(trim)
	if strings.Contains(lower, "always-approve") || strings.Contains(trim, "Grok 4.") ||
		strings.Contains(trim, "Mock Model") || strings.Contains(trim, "Composer ") {
		return true
	}
	return false
}

func isGrokHelpFooterLine(trim, lower string) bool {
	if trim == "" {
		return false
	}
	if strings.Contains(trim, "Ctrl+.:shortcuts") || strings.Contains(trim, "Ctrl+.:") {
		return true
	}
	if strings.Contains(trim, "Shift+Tab:mode") || strings.Contains(lower, "space:prompt") {
		return true
	}
	if strings.Contains(trim, "Enter:send") || strings.Contains(trim, "Enter:open") || strings.HasPrefix(trim, "Enter:") {
		return true
	}
	if strings.Contains(trim, "Ctrl+e:") || strings.Contains(trim, "Ctrl+c:cancel") || strings.Contains(trim, "Ctrl+;:queue") {
		return true
	}
	return false
}

func (f GrokFrame) hasKind(kind GrokSectionKind) bool {
	for _, s := range f.Sections {
		if s.Kind == kind {
			return true
		}
	}
	return false
}

func (f GrokFrame) kinds() []GrokSectionKind {
	out := make([]GrokSectionKind, 0, len(f.Sections))
	for _, s := range f.Sections {
		out = append(out, s.Kind)
	}
	return out
}

func (f GrokFrame) workedForStop() bool {
	for _, s := range f.Sections {
		if s.Kind != GrokSectionWorkedFor {
			continue
		}
		for _, ln := range s.Lines {
			if meta, ok := parseGrokWorkedForLine(strings.TrimSpace(ln)); ok && meta.Stop {
				return true
			}
		}
	}
	return false
}

func (f GrokFrame) hasRunningIndicator() bool {
	return f.hasKind(GrokSectionRunningIndicator)
}

func (f GrokFrame) hasStatusAboveComposer() bool {
	return f.hasKind(GrokSectionStatusAboveComposer)
}

func grokSectionHasLiveThinking(text string) bool {
	if reGrokLiveThinking.MatchString(text) {
		return true
	}
	lower := strings.ToLower(text)
	return strings.Contains(lower, "thinking about")
}

func (f GrokFrame) hasLiveThinkingBeforeSettledTurn() bool {
	for _, s := range f.Sections {
		switch s.Kind {
		case GrokSectionWorkedFor:
			for _, ln := range s.Lines {
				if meta, ok := parseGrokWorkedForLine(strings.TrimSpace(ln)); ok && meta.Stop {
					return false
				}
			}
		case GrokSectionActivityStrip, GrokSectionBody:
			if grokSectionHasLiveThinking(strings.Join(s.Lines, "\n")) {
				return true
			}
		}
	}
	return false
}

// judgeGrokFrameBusy returns a definitive busy/idle decision from sections when possible.
// ok=false means the frame has no usable composer/status chrome yet.
// StatusAboveComposer / RunningIndicator beat an earlier WorkedFor+bare-stop.
// WorkedFor+bare-stop beats historical body "thinking" crumbs.
// A visible boxed composer with no busy sections is idle.
func judgeGrokFrameBusy(f GrokFrame) (busy bool, ok bool) {
	if f.hasRunningIndicator() {
		return true, true
	}
	if f.hasStatusAboveComposer() {
		return true, true
	}
	if f.workedForStop() {
		return false, true
	}
	if f.hasLiveThinkingBeforeSettledTurn() {
		return true, true
	}
	if f.hasKind(GrokSectionComposer) {
		return false, true
	}
	return false, false
}
