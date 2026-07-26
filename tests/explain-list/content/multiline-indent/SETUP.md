# Scenario

**Feature**: multi-line message bodies preserve newlines and indent continuations

```
# user "hello\n\nworld"; assistant multi-line with blank line
explain list -> first line after Q/A label; non-empty continuations 6 spaces;
               blank segments emit pure \n (no spaces)
```

## Preconditions

- One session with newline-containing Q and multi-line A (including a blank line).

## Steps

1. Seed messages with internal newlines and a blank line.
2. Run list.
3. Assert preserved layout: 6-space indent on non-empty continuations; pure blank lines.

## Context

- Algorithm: split on `\n`; line 0 is `   {label}  {line0}`; empty segments → `\n` only;
  non-empty continuations → exactly 6 spaces + segment.
- Classic TDD: current product collapses whitespace → this leaf is RED until implement.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"list"}
	req.Sessions = []SessionSeed{
		simpleSession(
			"2026-07-13-11-00-00-multiline-dddddddd",
			"opencode", "deepseek-chat",
			"hello\n\nworld",
			"first\nsecond\n\nthird",
		),
	}
	return nil
}
```
