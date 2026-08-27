package agenttty

import "strings"

// InputBoxStatus is live composer occupancy: empty, occupied, or unknown.
type InputBoxStatus string

const (
	InputBoxEmpty    InputBoxStatus = "empty"
	InputBoxOccupied InputBoxStatus = "occupied"
	InputBoxUnknown  InputBoxStatus = "unknown"
)

func (s InputBoxStatus) String() string { return string(s) }

const (
	codexGlyphSingle = "\u203a" // ›
	codexGlyphDouble = "\u00bb" // »
	grokGlyph        = "\u276f" // ❯ (HEAVY RIGHT-POINTING ANGLE QUOTATION MARK ORNAMENT)
	grokGlyphAlt     = "\u2771" // ❱ (requirement alias; live TUI uses U+276F)
	// Codex empty-composer glue: placeholder on the same line as the model footer.
	codexFooterGlue = " medium \u00b7 " // " medium · "
)

// DetectInputBox classifies live composer occupancy from snapshot text.
// It uses the last ›/» (Codex) or ❯ (Grok). Empty text or no glyph is unknown.
// Codex treats " medium · " on the glyph line as empty; Grok ignores that glue.
func DetectInputBox(text string) InputBoxStatus {
	if text == "" {
		return InputBoxUnknown
	}
	plain := StripANSI([]byte(text))
	idx, glyph, family := lastComposerGlyph(plain)
	if idx < 0 {
		return InputBoxUnknown
	}
	line := lineAt(plain, idx)
	remainder := remainderAfterGlyph(plain, idx, glyph)
	if family == "grok" {
		user := grokRemainderUserText(remainder)
		if user == "" || isGrokComposerPlaceholder(user) {
			return InputBoxEmpty
		}
		return InputBoxOccupied
	}
	if strings.Contains(line, codexFooterGlue) || strings.TrimSpace(remainder) == "" {
		return InputBoxEmpty
	}
	return InputBoxOccupied
}

// InputBoxReport maps a probe token to the locked CLI human line (no trailing
// newline) and JSON input_box value.
func InputBoxReport(token string) (human, jsonVal string) {
	token = strings.TrimSpace(token)
	if token == "" {
		token = string(InputBoxUnknown)
	}
	return "input box: " + token, token
}

func lastComposerGlyph(text string) (idx int, glyph, family string) {
	idx = -1
	if i := strings.LastIndex(text, codexGlyphSingle); i > idx {
		idx, glyph, family = i, codexGlyphSingle, "codex"
	}
	if i := strings.LastIndex(text, codexGlyphDouble); i > idx {
		idx, glyph, family = i, codexGlyphDouble, "codex"
	}
	if i := strings.LastIndex(text, grokGlyph); i > idx {
		idx, glyph, family = i, grokGlyph, "grok"
	}
	if i := strings.LastIndex(text, grokGlyphAlt); i > idx {
		idx, glyph, family = i, grokGlyphAlt, "grok"
	}
	return idx, glyph, family
}

func lineAt(text string, idx int) string {
	start := 0
	if i := strings.LastIndex(text[:idx], "\n"); i >= 0 {
		start = i + 1
	}
	end := len(text)
	if i := strings.IndexByte(text[idx:], '\n'); i >= 0 {
		end = idx + i
	}
	return text[start:end]
}

// grokRemainderUserText is the composer draft after ❯, ignoring box borders
// and block-cursor glyphs (live chrome is `│ ❯ … │` plus a cursor).
func grokRemainderUserText(remainder string) string {
	s := strings.TrimSpace(remainder)
	s = strings.TrimRightFunc(s, grokComposerChromeRune)
	return strings.TrimSpace(s)
}

// isGrokComposerPlaceholder reports idle chrome hints shown when the box is empty.
func isGrokComposerPlaceholder(user string) bool {
	switch strings.ToLower(strings.TrimSpace(user)) {
	case "build anything", "add a follow-up":
		return true
	default:
		return false
	}
}

func grokComposerChromeRune(r rune) bool {
	switch {
	case r >= 0x2500 && r <= 0x257F: // box drawing
		return true
	case r >= 0x2580 && r <= 0x259F: // block elements (cursor)
		return true
	case r == 0x25A0, r == 0x25A1, r == 0x25AE, r == 0x25AF:
		return true
	default:
		return false
	}
}

// LastComposerRemainder is the text after the last ›/»/❯ on its line
// (ANSI stripped). Empty if there is no composer glyph.
func LastComposerRemainder(text string) string {
	plain := StripANSI([]byte(text))
	idx, glyph, _ := lastComposerGlyph(plain)
	if idx < 0 {
		return ""
	}
	return remainderAfterGlyph(plain, idx, glyph)
}

func remainderAfterGlyph(text string, idx int, glyph string) string {
	start := idx + len(glyph)
	if start >= len(text) {
		return ""
	}
	end := len(text)
	if i := strings.IndexByte(text[start:], '\n'); i >= 0 {
		end = start + i
	}
	return text[start:end]
}
