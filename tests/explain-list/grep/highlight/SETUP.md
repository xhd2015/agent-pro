# Scenario

**Feature**: with greps, color on paints match spans bold red in Q/A bodies

```
# explain list --grep P [--color]
-> filter as usual; when color on: \x1b[1;31m…\x1b[0m around match text
# when color off: filter only, no ANSI on bodies
```

## Preconditions

- Shared short fixtures with known casing for match spans.
- Color-on leaves pass `--color`; color-off omits it (piped non-TTY).

## Steps

1. Leaves seed body text containing the pattern(s).
2. Assert bold-red SGR around original-case match text, or absence of ANSI.

## Context

- H1: bold red `\x1b[1;31m`. H2: per-line highlight. H3: leftmost-first,
  no nested/overlapping spans. Highlight every non-overlapping occurrence of
  every provided pattern on kept cards.
- Q/A labels still use cyan/green when color on.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if req.Bin == "" {
		t.Fatalf("grep/highlight setup: explain binary not built")
	}
	return nil
}

// highlightDockerSession has mixed-case "Docker" in the Q body.
func highlightDockerSession() SessionSeed {
	return simpleSession(
		"2026-07-16-14-00-00-hl-docker-aaaaaaaa",
		"opencode", "deepseek-chat",
		"Using Docker Compose for local stacks",
		"A short answer without the pattern.",
	)
}
```
