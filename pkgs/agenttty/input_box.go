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
		if strings.TrimSpace(remainder) == "" {
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
