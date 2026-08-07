# Scenario

**Feature**: pure chat-history formatter (`pkgs/msgfmt`) — L2 in-process

```
msgs []Message + Options
  -> msgfmt.Format / FormatDetailed
  -> Text block + Result metadata
```

## Preconditions

- Package `github.com/xhd2015/agent-pro/pkgs/msgfmt` exports the locked API in
  root `DOCTEST.md` (Message, Options, Result, Format, FormatDetailed,
  `DefaultMaxPerMessageRunes = 1000`).
- Tests are pure Go: **no** `t.Setenv`, **no** `Chdir`, **no** filesystem, **no**
  network, **no** SeaTalk imports.
- Input order is oldest → newest on every leaf unless stated otherwise.
- Zero `Options` fields mean: body cap default 1000; no message-count cap; no
  total budget.

## Steps

1. Root `Setup` normalizes zero-value `Options` fields; leaves own `Msgs`.
2. Leaf `Setup` builds messages and options for one scenario.
3. Root `Run` calls `FormatDetailed` and `Format`.
4. Leaf `Assert` checks `Text`, line shape, and/or `Result` fields.

## Context

- Truncation marker is Unicode ellipsis `…` (U+2026), one rune.
- Multi-message header always uses `Chat history (showing K of N):`.
- Singular header only when exactly one source message is shown:
  `Chat history (1 message):`.
- Helpers below are shared by all leaves.

```go
import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/xhd2015/agent-pro/pkgs/msgfmt"
)

const truncationMarker = "…" // U+2026

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// Pure library suite: no temp dirs / env. Preserve any leaf-prefilled opts
	// while keeping zero-value semantics (0 MaxPerMessageRunes → package default).
	// Leaves always own Msgs (nil vs empty is a deliberate leaf choice).
	req.Opts = msgfmt.Options{
		MaxPerMessageRunes: req.Opts.MaxPerMessageRunes,
		MaxMessages:        req.Opts.MaxMessages,
		TotalBudgetRunes:   req.Opts.TotalBudgetRunes,
	}
	return nil
}

func msg(id, sender, text string) msgfmt.Message {
	return msgfmt.Message{ID: id, Sender: sender, Text: text}
}

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected Run error: %v", err)
	}
}

func assertResp(t *testing.T, resp *Response) {
	t.Helper()
	if resp == nil {
		t.Fatal("nil response")
	}
}

func assertEqualString(t *testing.T, field, got, want string) {
	t.Helper()
	if got != want {
		t.Fatalf("%s:\n got: %q\nwant: %q", field, got, want)
	}
}

func assertEqualInt(t *testing.T, field string, got, want int) {
	t.Helper()
	if got != want {
		t.Fatalf("%s: got %d, want %d", field, got, want)
	}
}

func assertContains(t *testing.T, got, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Fatalf("expected %q in:\n%s", want, got)
	}
}

func assertNotContains(t *testing.T, got, want string) {
	t.Helper()
	if strings.Contains(got, want) {
		t.Fatalf("expected NOT to contain %q in:\n%s", want, got)
	}
}

func assertFormatEqualsDetail(t *testing.T, resp *Response) {
	t.Helper()
	assertEqualString(t, "Format vs FormatDetailed.Text", resp.Text, resp.Detail.Text)
}

func assertEmptyResult(t *testing.T, resp *Response) {
	t.Helper()
	assertEqualString(t, "Text", resp.Text, "")
	assertEqualString(t, "Detail.Text", resp.Detail.Text, "")
	assertEqualInt(t, "Shown", resp.Detail.Shown, 0)
	assertEqualInt(t, "SourceCount", resp.Detail.SourceCount, 0)
	assertEqualInt(t, "OldestDropped", resp.Detail.OldestDropped, 0)
	assertEqualInt(t, "BodiesTruncated", resp.Detail.BodiesTruncated, 0)
	assertEqualString(t, "LastMessageID", resp.Detail.LastMessageID, "")
	assertFormatEqualsDetail(t, resp)
}

// bodyRuneCountOfLine returns the rune count of the body after the last " : "
// separator, or of the whole line if no separator is present (text-only lines).
func bodyRuneCountOfLine(line string) int {
	const sep = " : "
	if i := strings.LastIndex(line, sep); i >= 0 {
		return utf8.RuneCountInString(line[i+len(sep):])
	}
	return utf8.RuneCountInString(line)
}

// messageLines strips the header line and returns body lines without trailing empties.
func messageLines(text string) []string {
	text = strings.TrimSuffix(text, "\n")
	parts := strings.Split(text, "\n")
	if len(parts) <= 1 {
		return nil
	}
	return parts[1:]
}
```
