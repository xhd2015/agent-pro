// Package msgfmt formats an ordered chat transcript (oldest → newest) into a
// plain-text block suitable for injecting into an agent prompt.
//
// Pure and SeaTalk-unaware: no I/O, env, cwd, or process globals.
package msgfmt

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// DefaultMaxPerMessageRunes is the default per-message body rune cap when
// Options.MaxPerMessageRunes is 0 or negative.
const DefaultMaxPerMessageRunes = 1000

// truncationMarker is the single-rune Unicode ellipsis used when a body is
// shortened (U+2026). Not ASCII "..." and not "[truncated]".
const truncationMarker = "…"

// Message is one chat message in oldest → newest order.
type Message struct {
	ID     string
	Sender string
	Text   string
}

// Options controls selection and size caps. Zero fields select package defaults:
// body cap DefaultMaxPerMessageRunes, no MaxMessages count cap, no total budget.
type Options struct {
	// MaxPerMessageRunes caps each message body. 0 or negative → DefaultMaxPerMessageRunes.
	MaxPerMessageRunes int
	// MaxMessages, when >0, keeps the latest N messages. 0 means no count cap.
	MaxMessages int
	// TotalBudgetRunes, when >0, is a rune budget on the full formatted block.
	// 0 means no total budget. Oldest messages are dropped first; the last
	// message is always kept even if the block still exceeds the budget.
	TotalBudgetRunes int
}

// Result is a structured view of what FormatDetailed produced.
type Result struct {
	Text            string
	Shown           int
	SourceCount     int
	OldestDropped   int // SourceCount - Shown
	BodiesTruncated int // messages whose body was shortened (among shown)
	LastMessageID   string // newest input message ID ("" if none / empty id)
}

// prepared is one message after body capping, ready for block assembly.
type prepared struct {
	msg       Message
	truncated bool
}

// Format returns the formatted chat-history block (same as FormatDetailed.Text).
func Format(msgs []Message, opts Options) string {
	return FormatDetailed(msgs, opts).Text
}

// FormatDetailed formats msgs under opts and returns text plus selection metadata.
//
// Pipeline: (1) MaxMessages keeps latest N → (2) per-body rune cap →
// (3) TotalBudgetRunes drops oldest from the full formatted block.
func FormatDetailed(msgs []Message, opts Options) Result {
	sourceCount := len(msgs)
	lastID := ""
	if sourceCount > 0 {
		lastID = msgs[sourceCount-1].ID
	}

	if sourceCount == 0 {
		return Result{}
	}

	// (a) count-cap: keep latest N when MaxMessages > 0
	selected := msgs
	if opts.MaxMessages > 0 && len(selected) > opts.MaxMessages {
		selected = selected[len(selected)-opts.MaxMessages:]
	}

	maxBody := opts.MaxPerMessageRunes
	if maxBody <= 0 {
		maxBody = DefaultMaxPerMessageRunes
	}

	// (b) per-message body rune cap
	prep := make([]prepared, len(selected))
	for i, m := range selected {
		body, trunc := capBody(m.Text, maxBody)
		m.Text = body
		prep[i] = prepared{msg: m, truncated: trunc}
	}

	// (c) total budget on full formatted block: drop oldest until fits,
	// always keep at least the last message.
	for len(prep) > 1 && opts.TotalBudgetRunes > 0 {
		text := buildBlock(sourceCount, prep)
		if utf8.RuneCountInString(text) <= opts.TotalBudgetRunes {
			break
		}
		prep = prep[1:]
	}

	// Final block (may still exceed budget when only the last message remains).
	text := buildBlock(sourceCount, prep)

	bodiesTruncated := 0
	for _, p := range prep {
		if p.truncated {
			bodiesTruncated++
		}
	}
	shown := len(prep)
	return Result{
		Text:            text,
		Shown:           shown,
		SourceCount:     sourceCount,
		OldestDropped:   sourceCount - shown,
		BodiesTruncated: bodiesTruncated,
		LastMessageID:   lastID,
	}
}

// capBody shortens body so its rune length is at most max.
// Truncated form is prefix of (max-1) runes + "…"; result rune count == max.
func capBody(body string, max int) (string, bool) {
	if max <= 0 {
		max = DefaultMaxPerMessageRunes
	}
	n := utf8.RuneCountInString(body)
	if n <= max {
		return body, false
	}
	// Keep max-1 content runes + one-rune marker.
	keep := max - 1
	if keep < 0 {
		keep = 0
	}
	if keep == 0 {
		return truncationMarker, true
	}
	runes := []rune(body)
	return string(runes[:keep]) + truncationMarker, true
}

// formatLine builds one message line per omit rules for empty id/sender.
func formatLine(m Message) string {
	switch {
	case m.ID != "" && m.Sender != "":
		// Two spaces before '['.
		return fmt.Sprintf("message_id=%s  [%s] : %s", m.ID, m.Sender, m.Text)
	case m.ID != "" && m.Sender == "":
		return fmt.Sprintf("message_id=%s : %s", m.ID, m.Text)
	case m.ID == "" && m.Sender != "":
		return fmt.Sprintf("[%s] : %s", m.Sender, m.Text)
	default:
		return m.Text
	}
}

// headerFor returns the singular header only when source and shown are both 1;
// otherwise the multi form "showing K of N".
func headerFor(sourceCount, shown int) string {
	if sourceCount == 1 && shown == 1 {
		return "Chat history (1 message):"
	}
	return fmt.Sprintf("Chat history (showing %d of %d):", shown, sourceCount)
}

// buildBlock assembles header + message lines, each line ending with '\n'.
func buildBlock(sourceCount int, prep []prepared) string {
	if len(prep) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString(headerFor(sourceCount, len(prep)))
	b.WriteByte('\n')
	for _, p := range prep {
		b.WriteString(formatLine(p.msg))
		b.WriteByte('\n')
	}
	return b.String()
}
