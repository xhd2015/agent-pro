package commit_msg

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// SanitizeResult is the structured outcome of post-parse commit message sanitization.
type SanitizeResult struct {
	Msg      CommitMsg
	Changes  []string
	Rejected bool
	Reason   string
}

var (
	// Meta labels on a title line (optional surrounding bold).
	reMetaTitleChars = regexp.MustCompile(`(?i)^\*{0,2}Title\s*\(\s*\d+\s*chars?\s*\)\s*:\*{0,2}\s*`)
	reMetaTitle      = regexp.MustCompile(`(?i)^\*{0,2}Title\s*:\*{0,2}\s*`)
	reMetaCommitMsg  = regexp.MustCompile(`(?i)^\*{0,2}Commit\s+message\s*:\*{0,2}\s*`)
	reMetaDescription = regexp.MustCompile(`(?i)^\*{0,2}Description\s*:\*{0,2}\s*`)
	// Trailing meta char-count annotations, e.g. " (48 chars)".
	reTrailingCharCount = regexp.MustCompile(`(?i)\s*\(\s*\d+\s*chars?\s*\)\s*$`)
	// Tool noise: todowrite … completed.
	reTodoWriteNoise = regexp.MustCompile(`(?is)\btodowrite\b.*\bcompleted\b`)
	// git commit -m "..." / -m '...' tokens (non-greedy quoted args).
	reGitCommitM = regexp.MustCompile(`(?is)^\s*(?:` + "`" + `)?\s*git\s+commit\b`)
	reDashMArg   = regexp.MustCompile(`-m\s+(?:"((?:\\.|[^"\\])*)"|'((?:\\.|[^'\\])*)')`)
)

// Sanitize applies post-parse anti-pattern cleanup to agent commit-message text.
// On success returns a usable CommitMsg. On unusable garbage returns Rejected=true.
func Sanitize(raw string) SanitizeResult {
	var changes []string
	text := strings.TrimSpace(raw)
	if text == "" {
		return SanitizeResult{Rejected: true, Reason: "empty commit message"}
	}

	// Rule 2: whole-blob JSON with title → CommitMsg.
	msg, ok := tryParseJSONCommitMsg(text)
	if ok {
		changes = append(changes, "json")
	} else {
		// Rule 3: strip leading/trailing fence lines, then retry JSON.
		fenced := stripFenceLines(text)
		if fenced != text {
			changes = append(changes, "fence")
			text = fenced
		}
		if msg2, ok2 := tryParseJSONCommitMsg(text); ok2 {
			msg = msg2
			changes = append(changes, "json")
			ok = true
		}
	}

	if !ok {
		// Non-JSON blob path: treat as a single text that may become title (+ desc via -m).
		blob := text
		blob = stripFenceLines(blob)
		blob = stripMetaLabels(blob)
		if blob != text {
			changes = append(changes, "meta")
		}
		blob = strings.TrimSpace(blob)

		// Unwrap outer matching wrappers on the whole blob before git -m extract.
		unwrapped, didUnwrap := unwrapOuter(blob)
		if didUnwrap {
			changes = append(changes, "unwrap")
			blob = unwrapped
		}

		if isGitCommitWrapper(blob) {
			if m, extracted := extractGitCommitM(blob); extracted {
				msg = m
				changes = append(changes, "git-m")
			} else {
				msg = CommitMsg{Title: blob}
			}
		} else {
			// Split into title + description on first blank line if present.
			msg = splitTitleDesc(blob)
		}
	}

	// Field-level cleanup (JSON path and non-JSON path).
	msg.Title = strings.TrimSpace(msg.Title)
	msg.Description = strings.TrimSpace(msg.Description)

	// Strip meta labels from title (and description if needed).
	if cleaned := stripMetaLabels(msg.Title); cleaned != msg.Title {
		msg.Title = strings.TrimSpace(cleaned)
		changes = append(changes, "meta-title")
	}
	if cleaned := stripMetaLabels(msg.Description); cleaned != msg.Description {
		msg.Description = strings.TrimSpace(cleaned)
		changes = append(changes, "meta-desc")
	}

	// If title still looks like a git commit wrapper (JSON title edge case).
	if isGitCommitWrapper(msg.Title) {
		if m, extracted := extractGitCommitM(msg.Title); extracted {
			// Prefer extracted body over existing description when multi -m.
			if m.Description != "" {
				msg = m
			} else {
				msg.Title = m.Title
			}
			changes = append(changes, "git-m-title")
		}
	}

	// Unwrap outer `...` / **...** on title and description (rule 5).
	if u, did := unwrapOuter(msg.Title); did {
		msg.Title = u
		changes = append(changes, "unwrap-title")
	} else if u, did := stripUnclosedLeadingTick(msg.Title); did {
		msg.Title = u
		changes = append(changes, "leading-tick")
	}
	if u, did := unwrapOuter(msg.Description); did {
		msg.Description = u
		changes = append(changes, "unwrap-desc")
	} else if u, did := stripUnclosedLeadingTick(msg.Description); did {
		msg.Description = u
		changes = append(changes, "leading-tick-desc")
	}

	// Drop trailing char-count annotations when clearly meta (rule 7).
	if cleaned := reTrailingCharCount.ReplaceAllString(msg.Title, ""); cleaned != msg.Title {
		msg.Title = strings.TrimSpace(cleaned)
		changes = append(changes, "char-count")
	}
	if cleaned := reTrailingCharCount.ReplaceAllString(msg.Description, ""); cleaned != msg.Description {
		msg.Description = strings.TrimSpace(cleaned)
		changes = append(changes, "char-count-desc")
	}

	msg.Title = strings.TrimSpace(msg.Title)
	msg.Description = strings.TrimSpace(msg.Description)

	// Rule 8: reject empty or tool noise.
	if msg.Title == "" {
		return SanitizeResult{Rejected: true, Reason: "empty title after sanitize", Changes: changes}
	}
	combined := msg.Title
	if msg.Description != "" {
		combined = msg.Title + "\n\n" + msg.Description
	}
	if isToolNoise(combined) || isToolNoise(raw) {
		return SanitizeResult{
			Rejected: true,
			Reason:   "unusable commit message (tool noise)",
			Changes:  changes,
		}
	}

	// Never truncate by length (rule 9). Preserve inner backticks (rule 10) by only
	// unwrapping matching outer pairs above.
	return SanitizeResult{Msg: msg, Changes: changes}
}

// SanitizeOrError runs Sanitize and returns an error when the message is rejected.
func SanitizeOrError(raw string) (CommitMsg, error) {
	res := Sanitize(raw)
	if res.Rejected {
		reason := res.Reason
		if reason == "" {
			reason = "unusable"
		}
		return CommitMsg{}, fmt.Errorf("%s", reason)
	}
	return res.Msg, nil
}

func tryParseJSONCommitMsg(text string) (CommitMsg, bool) {
	text = strings.TrimSpace(text)
	if text == "" {
		return CommitMsg{}, false
	}
	// Prefer extractJSONFromText so fenced/embedded JSON still works.
	jsonText := extractJSONFromText(text)
	if jsonText == "" {
		return CommitMsg{}, false
	}
	var msg CommitMsg
	if err := json.Unmarshal([]byte(jsonText), &msg); err != nil || msg.Title == "" {
		return CommitMsg{}, false
	}
	msg.Title = strings.TrimSpace(msg.Title)
	msg.Description = strings.TrimSpace(msg.Description)
	return msg, true
}

// stripFenceLines removes leading/trailing markdown code-fence lines (``` / ```json).
func stripFenceLines(text string) string {
	lines := strings.Split(text, "\n")
	// Leading fences
	for len(lines) > 0 {
		trimmed := strings.TrimSpace(lines[0])
		if isFenceLine(trimmed) {
			lines = lines[1:]
			continue
		}
		break
	}
	// Trailing fences
	for len(lines) > 0 {
		trimmed := strings.TrimSpace(lines[len(lines)-1])
		if isFenceLine(trimmed) {
			lines = lines[:len(lines)-1]
			continue
		}
		break
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

func isFenceLine(s string) bool {
	if s == "```" || strings.HasPrefix(s, "```") {
		// ``` or ```json / ```go etc. — whole line is a fence marker
		rest := strings.TrimPrefix(s, "```")
		rest = strings.TrimSpace(rest)
		// fence language token: alnum only, no spaces required but allow simple lang
		if rest == "" {
			return true
		}
		for _, r := range rest {
			if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '+' && r != '-' && r != '_' {
				return false
			}
		}
		return true
	}
	return false
}

// stripMetaLabels removes leading meta labels from text (first line / whole string).
func stripMetaLabels(text string) string {
	s := strings.TrimSpace(text)
	// Apply repeatedly until stable (at most a few times).
	for i := 0; i < 4; i++ {
		orig := s
		// Prefer the more specific Title (N chars): form first.
		if loc := reMetaTitleChars.FindStringIndex(s); loc != nil && loc[0] == 0 {
			s = strings.TrimSpace(s[loc[1]:])
			continue
		}
		if loc := reMetaTitle.FindStringIndex(s); loc != nil && loc[0] == 0 {
			s = strings.TrimSpace(s[loc[1]:])
			continue
		}
		if loc := reMetaCommitMsg.FindStringIndex(s); loc != nil && loc[0] == 0 {
			s = strings.TrimSpace(s[loc[1]:])
			continue
		}
		// Description label: only strip if it is the whole first segment start
		// and the string looks like a meta-only prefix (not a legitimate "Description: ..." commit subject).
		// Design lists **Description:** as a meta label to strip.
		if loc := reMetaDescription.FindStringIndex(s); loc != nil && loc[0] == 0 {
			s = strings.TrimSpace(s[loc[1]:])
			continue
		}
		if s == orig {
			break
		}
	}
	return s
}

// unwrapOuter removes a single matching outer wrapper: `...` or **...**.
// Does not strip legitimate inner backticks.
func unwrapOuter(s string) (string, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return s, false
	}
	// Outer bold **...**
	if strings.HasPrefix(s, "**") && strings.HasSuffix(s, "**") && len(s) > 4 {
		inner := s[2 : len(s)-2]
		// Avoid unwrapping if the "inner" itself is empty or still only markers oddly.
		// Allow inner content to contain ** (rare).
		return strings.TrimSpace(inner), true
	}
	// Outer single backticks: whole string wrapped in one pair of `
	if strings.HasPrefix(s, "`") && strings.HasSuffix(s, "`") && len(s) > 1 {
		// Matching outer: starts and ends with ` and the first/last form a pair wrapping the rest.
		// For "feat: add `--open` flag" we must NOT unwrap because the first ` is not an outer wrapper
		// of the whole string — the string starts with 'f', not with a wrapping pair.
		// Wait: "feat: add `--open` flag" does NOT start with `.
		// What about "`feat: add `--open` flag`"? That would be outer wrap with inner ticks.
		// Design: unwrap outer matching. For outer wrap with inner ticks, first and last chars are `.
		// Risk: "`--open`" alone would unwrap to --open which is correct for outer-only.
		// For legitimate "feat: add `--open` flag" — no leading/trailing outer ticks on whole field.
		inner := s[1 : len(s)-1]
		// If inner still has unmatched structure we still unwrap one outer layer (ordered rule).
		return strings.TrimSpace(inner), true
	}
	return s, false
}

// stripUnclosedLeadingTick strips a single leading backtick when there is no matching trailing one.
func stripUnclosedLeadingTick(s string) (string, bool) {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "`") && !strings.HasSuffix(s, "`") {
		return strings.TrimSpace(s[1:]), true
	}
	return s, false
}

func isGitCommitWrapper(s string) bool {
	return reGitCommitM.MatchString(strings.TrimSpace(s))
}

// extractGitCommitM parses git commit -m "..." [-m "..."]* into title + description.
// 1st -m → title; remaining -m → description joined by \n\n.
func extractGitCommitM(s string) (CommitMsg, bool) {
	s = strings.TrimSpace(s)
	// Drop outer backticks if still present.
	if u, did := unwrapOuter(s); did {
		s = u
	}
	if !isGitCommitWrapper(s) {
		return CommitMsg{}, false
	}
	matches := reDashMArg.FindAllStringSubmatch(s, -1)
	if len(matches) == 0 {
		return CommitMsg{}, false
	}
	var args []string
	for _, m := range matches {
		arg := m[1]
		if arg == "" {
			arg = m[2]
		}
		// Unescape simple \" sequences.
		arg = strings.ReplaceAll(arg, `\"`, `"`)
		arg = strings.ReplaceAll(arg, `\'`, `'`)
		args = append(args, strings.TrimSpace(arg))
	}
	if len(args) == 0 || args[0] == "" {
		return CommitMsg{}, false
	}
	msg := CommitMsg{Title: args[0]}
	if len(args) > 1 {
		msg.Description = strings.Join(args[1:], "\n\n")
	}
	return msg, true
}

func splitTitleDesc(blob string) CommitMsg {
	blob = strings.TrimSpace(blob)
	// Prefer \n\n split (formatted message style).
	if parts := strings.SplitN(blob, "\n\n", 2); len(parts) == 2 {
		return CommitMsg{
			Title:       strings.TrimSpace(parts[0]),
			Description: strings.TrimSpace(parts[1]),
		}
	}
	// Single-line or multi-line without blank: first line title, rest description.
	lines := strings.Split(blob, "\n")
	if len(lines) == 1 {
		return CommitMsg{Title: strings.TrimSpace(lines[0])}
	}
	title := strings.TrimSpace(lines[0])
	rest := strings.TrimSpace(strings.Join(lines[1:], "\n"))
	return CommitMsg{Title: title, Description: rest}
}

func isToolNoise(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	if reTodoWriteNoise.MatchString(s) {
		return true
	}
	// Also catch compact form without requiring much between tokens.
	lower := strings.ToLower(s)
	if strings.Contains(lower, "todowrite") && strings.Contains(lower, "completed") {
		return true
	}
	return false
}
