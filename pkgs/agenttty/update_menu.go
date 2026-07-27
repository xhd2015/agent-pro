package agenttty

import "strings"

// IsBlockingUpdateMenu reports whether text is the Codex full-screen
// "Update available" menu modal (numbered Skip / Update now options).
// Residual banners that still say "Update available" without menu options return false.
//
// Directory trust prompts also show "Press enter to continue" and must NOT match
// (they would otherwise block send as "codex update available").
func IsBlockingUpdateMenu(text string) bool {
	if strings.TrimSpace(text) == "" {
		return false
	}
	lower := strings.ToLower(text)
	compact := compactWritableText(lower)

	// Trust modal is not the update menu even though both use an enter footer.
	if strings.Contains(compact, "doyoutrustthecontentsofthisdirectory") ||
		(strings.Contains(lower, "do you trust the contents") &&
			(strings.Contains(lower, "yes, continue") || strings.Contains(compact, "yescontinue"))) {
		return false
	}

	hasUpdate := strings.Contains(lower, "update available") ||
		strings.Contains(compact, "updateavailable")
	hasSkipUntil := strings.Contains(lower, "skip until next version") ||
		strings.Contains(compact, "skipuntilnextversion")
	hasNumberedUpdate := strings.Contains(lower, "1. update") ||
		strings.Contains(compact, "1.update") ||
		strings.Contains(lower, "1. update now") ||
		strings.Contains(lower, "update now")
	hasNumberedSkip := strings.Contains(lower, "2. skip") ||
		strings.Contains(compact, "2.skip")
	hasEnterFooter := strings.Contains(lower, "press enter to continue") ||
		strings.Contains(compact, "pressentertocontinue")

	// Explicit skip-until option is update-menu only.
	if hasSkipUntil {
		return true
	}
	// Numbered Update now + Skip is the classic modal.
	if hasNumberedUpdate && hasNumberedSkip {
		return true
	}
	// "Update available" chrome with enter footer or numbered options.
	if hasUpdate && (hasEnterFooter || hasNumberedUpdate || hasNumberedSkip) {
		return true
	}
	return false
}

// UpdateMenuSelection returns which menu option the › / U+203A marker is on:
// UPDATE_NOW, SKIP, SKIP_UNTIL_NEXT, or "" when no menu selection is visible.
func UpdateMenuSelection(text string) string {
	if !IsBlockingUpdateMenu(text) {
		return ""
	}
	return selectionFromLines(text)
}

func selectionFromLines(text string) string {
	// Prefer the most specific match among selection-marked lines.
	// Only treat lines that look like menu options (numbered or option keywords).
	var found string
	for _, line := range strings.Split(text, "\n") {
		sel, ok := selectionOnLine(line)
		if !ok {
			continue
		}
		// Prefer SKIP_UNTIL_NEXT over SKIP over UPDATE_NOW if multiple (shouldn't happen).
		switch sel {
		case "SKIP_UNTIL_NEXT":
			return sel
		case "SKIP":
			if found != "SKIP_UNTIL_NEXT" {
				found = sel
			}
		case "UPDATE_NOW":
			if found == "" {
				found = sel
			}
		}
	}
	return found
}

func selectionOnLine(line string) (string, bool) {
	// Selection marker is › or U+203A (live Codex 0.143.0).
	idx := strings.Index(line, "›")
	if idx < 0 {
		idx = strings.Index(line, "\u203a")
	}
	if idx < 0 {
		return "", false
	}
	// Content after the selection marker.
	rest := line[idx:]
	// Drop the marker rune itself for matching.
	if strings.HasPrefix(rest, "›") {
		rest = rest[len("›"):]
	} else if strings.HasPrefix(rest, "\u203a") {
		rest = rest[len("\u203a"):]
	}
	lower := strings.ToLower(strings.TrimSpace(rest))
	compact := compactWritableText(lower)

	// Must look like a menu option line, not the main chat prompt after ›.
	looksLikeOption := strings.Contains(lower, "update now") ||
		strings.Contains(lower, "skip") ||
		strings.HasPrefix(lower, "1.") ||
		strings.HasPrefix(lower, "2.") ||
		strings.HasPrefix(lower, "3.") ||
		strings.Contains(compact, "1.update") ||
		strings.Contains(compact, "2.skip") ||
		strings.Contains(compact, "3.skip")
	if !looksLikeOption {
		return "", false
	}

	// Order: until-next before bare Skip, then Update now.
	if strings.Contains(lower, "until") ||
		strings.Contains(compact, "skipuntil") {
		return "SKIP_UNTIL_NEXT", true
	}
	if strings.Contains(lower, "skip") || strings.Contains(compact, "2.skip") {
		return "SKIP", true
	}
	if strings.Contains(lower, "update") || strings.Contains(compact, "1.update") {
		return "UPDATE_NOW", true
	}
	return "", false
}
