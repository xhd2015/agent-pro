# msgfmt — SeaTalk-unaware chat history formatter

Classic TDD doctests for greenfield package
`github.com/xhd2015/agent-pro/pkgs/msgfmt` (**RED** until implementer lands
`Format` / `FormatDetailed`).

Pure in-process library: messages in → formatted chat-history block out.
**No** SeaTalk types, open-inject header, pathfmt, localbot, or SYSTEM.md.

# DSN (Domain Specific Notion)

**msgfmt** formats an ordered chat transcript (oldest → newest) into a plain-text
block suitable for injecting into an agent prompt. Callers may pre-shorten paths
into `Message.Text`; this package never imports pathfmt or SeaTalk.

```
msgs []Message (oldest → newest) + Options
  -> Format        -> string (same as Result.Text)
  -> FormatDetailed -> Result{Text, Shown, SourceCount, OldestDropped,
                              BodiesTruncated, LastMessageID}

line shape (generic):
  message_id=<id>  [<sender>] : <text>
  (omit empty id / empty sender cleanly — see locked line rules)

block shape (non-empty):
  <header>\n
  <line>\n
  ...
```

**Participants**

- **Caller** — bot/listen layer that already collected generic messages.
- **`Format` / `FormatDetailed`** — pure functions; no env, cwd, globals, I/O.
- **`Options`** — selection + size caps (zero fields → package defaults).
- **`Result`** — structured view of what was shown vs dropped/truncated.

**Locked types (implementer contract)**

```text
package msgfmt

const DefaultMaxPerMessageRunes = 1000

type Message struct {
	ID     string
	Sender string
	Text   string
}

type Options struct {
	MaxPerMessageRunes int // 0 or negative → DefaultMaxPerMessageRunes (1000)
	MaxMessages        int // 0 → no count cap; >0 keep latest N
	TotalBudgetRunes   int // 0 → no total budget on the formatted block
}

type Result struct {
	Text            string
	Shown           int
	SourceCount     int
	OldestDropped   int // SourceCount - Shown
	BodiesTruncated int // messages whose body was shortened
	LastMessageID   string // newest input message ID ("" if none / empty id)
}

func Format(msgs []Message, opts Options) string
func FormatDetailed(msgs []Message, opts Options) Result
```

**Behaviors (locked)**

1. **Order** — input is oldest → newest. `MaxMessages` keeps the **latest** N.
2. **Pipeline** — (a) count-cap latest N → (b) per-message body rune cap →
   (c) drop oldest under `TotalBudgetRunes` on the **full formatted block**,
   always preferring to keep the last (trigger) message.
3. **Body cap** — each message **body** (`Text`) is at most
   `MaxPerMessageRunes` runes (default 1000). Over-long bodies become
   `prefix…` where `…` is U+2026 and `utf8.RuneCountInString(body) == max`.
   Marker is **only** `…` (not `...`, not `[truncated]`).
4. **Empty** — `nil` or empty `msgs` → `""`; Result zeros / empty strings.
5. **Line omit rules** (clean empty fields):

| ID | Sender | Line |
|----|--------|------|
| set | set | `message_id=<id>  [<sender>] : <text>` |
| set | empty | `message_id=<id> : <text>` |
| empty | set | `[<sender>] : <text>` |
| empty | empty | `<text>` |

6. **Header labels**:
   - empty output: no header
   - `SourceCount == 1 && Shown == 1`: `Chat history (1 message):`
   - otherwise: `Chat history (showing K of N):` with `K=Shown`, `N=SourceCount`
7. **Block newline** — non-empty `Text` is header + `\n` + lines each ending in
   `\n` (trailing newline after the last message line).
8. **`Format` ≡ `FormatDetailed(...).Text`**.
9. **Budget edge** — if only the last message remains and the block still exceeds
   `TotalBudgetRunes`, keep that last message (already body-capped).
10. **Parallel-safe** — pure; no `Setenv` / `Chdir` / process globals.

**Out of scope**

- Open-inject framing / SeaTalk headers
- pathfmt / path shortening
- short/medium/long named presets as public API (optional internals only)
- L3 e2e / CLI

## Version

0.0.2

## Decision Tree

```
tests/msgfmt/
├── DOCTEST.md
├── SETUP.md
├── empty/                              # no messages → empty string
│   ├── SETUP.md
│   ├── nil-msgs/
│   └── empty-slice/
├── line-shape/                         # single-line field omit rules + singular label
│   ├── SETUP.md
│   ├── full-fields/
│   ├── omit-empty-id/
│   ├── omit-empty-sender/
│   └── text-only/
├── body-cap/                           # MaxPerMessageRunes / default 1000 / unicode
│   ├── SETUP.md
│   ├── exactly-default-max/
│   ├── over-default-max/
│   ├── custom-max/
│   └── unicode-counts-runes/
├── max-messages/                       # keep latest N + multi label K of N
│   ├── SETUP.md
│   ├── under-limit-shows-all/
│   └── keeps-latest-with-label/
├── total-budget/                       # TotalBudgetRunes drop oldest, keep last
│   ├── SETUP.md
│   ├── under-budget-shows-all/
│   ├── drops-oldest-first/
│   └── keeps-last-when-tight/
└── format-detailed/                    # Result metadata + Format parity
    ├── SETUP.md
    ├── result-counts/
    ├── bodies-truncated-count/
    ├── last-message-id/
    └── format-equals-result-text/
```

Parameter ranking (most → least significant):

1. **Input emptiness** — empty/`nil` vs at least one message
2. **Selection** — all / `MaxMessages` / `TotalBudgetRunes`
3. **Line fields** — id/sender present or omitted
4. **Body size** — under / at / over `MaxPerMessageRunes` (runes, not bytes)
5. **Observation API** — `Format` string vs `FormatDetailed` Result fields

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `empty/nil-msgs` | `nil` slice → `""`; Result all-zero |
| 2 | `empty/empty-slice` | empty non-nil slice → `""`; Result all-zero |
| 3 | `line-shape/full-fields` | id+sender+text; singular header; full line shape |
| 4 | `line-shape/omit-empty-id` | empty ID → `[sender] : text` (no `message_id=`) |
| 5 | `line-shape/omit-empty-sender` | empty Sender → `message_id=<id> : text` (no brackets) |
| 6 | `line-shape/text-only` | empty ID+Sender → body text only on the line |
| 7 | `body-cap/exactly-default-max` | 1000-rune body unchanged; `DefaultMaxPerMessageRunes=1000` |
| 8 | `body-cap/over-default-max` | 1001+ runes → body ≤1000 with trailing `…` |
| 9 | `body-cap/custom-max` | `MaxPerMessageRunes=5`, long body → 4 runes + `…` |
| 10 | `body-cap/unicode-counts-runes` | multi-byte runes counted as runes, not bytes |
| 11 | `max-messages/under-limit-shows-all` | 3 msgs, MaxMessages=5 → all 3; showing 3 of 3 |
| 12 | `max-messages/keeps-latest-with-label` | 10 msgs, MaxMessages=3 → last 3; showing 3 of 10 |
| 13 | `total-budget/under-budget-shows-all` | budget large → no drops |
| 14 | `total-budget/drops-oldest-first` | tight budget drops oldest; last body remains |
| 15 | `total-budget/keeps-last-when-tight` | budget < one full block still keeps last message |
| 16 | `format-detailed/result-counts` | Shown / SourceCount / OldestDropped after MaxMessages |
| 17 | `format-detailed/bodies-truncated-count` | BodiesTruncated counts truncated messages |
| 18 | `format-detailed/last-message-id` | LastMessageID = newest input message ID |
| 19 | `format-detailed/format-equals-result-text` | `Format` == `FormatDetailed.Text` |

## How to Run

From module root `…/agent-pro-master-2026-08-07` (directory that contains `go.mod`
for `github.com/xhd2015/agent-pro`):

```sh
doctest vet ./tests/msgfmt
doctest test -v ./tests/msgfmt
doctest test -v ./tests/msgfmt/body-cap/over-default-max
```

Classic TDD: **`doctest vet` must pass**; **`doctest test` is expected RED** until
`pkgs/msgfmt` exists with the locked API.

```go
import (
	"fmt"
	"testing"
	"unicode/utf8"

	"github.com/xhd2015/agent-pro/pkgs/msgfmt"
	"github.com/xhd2015/doctest/session"
)

// Request is the pure library input for one leaf.
type Request struct {
	Msgs []msgfmt.Message
	Opts msgfmt.Options
}

// Response captures both Format and FormatDetailed observations.
type Response struct {
	Text   string
	Detail msgfmt.Result
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	_ = d
	if req == nil {
		return nil, fmt.Errorf("nil request")
	}
	detail := msgfmt.FormatDetailed(req.Msgs, req.Opts)
	text := msgfmt.Format(req.Msgs, req.Opts)
	return &Response{Text: text, Detail: detail}, nil
}

// runeRepeat builds a string of n copies of the first rune in s (ASCII-safe helper).
func runeRepeat(s string, n int) string {
	if n <= 0 {
		return ""
	}
	r, _ := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError && len(s) == 0 {
		return ""
	}
	buf := make([]rune, n)
	for i := range buf {
		buf[i] = r
	}
	return string(buf)
}
```
