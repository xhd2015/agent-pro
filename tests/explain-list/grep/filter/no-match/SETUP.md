# Scenario

**Feature**: non-empty store with zero grep matches prints dedicated message

```
# sessions about cats; explain list --grep docker
-> stdout exactly "No matching explain sessions.\n"; exit 0
```

## Preconditions

- At least one valid session that does not match the pattern.

## Steps

1. Seed two non-matching sessions.
2. Args: `list --grep docker`.
3. Assert dedicated no-match message (not empty-store wording).

## Context

- Exit 0; distinct from `No explain sessions yet.`

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"list", "--grep", "docker"}
	req.Sessions = []SessionSeed{
		simpleSession(
			"2026-07-15-12-00-00-cat0-aaaaaaaa",
			"opencode", "deepseek-chat",
			"about cats marker-cat-0", "meow",
		),
		simpleSession(
			"2026-07-15-13-00-00-cat1-bbbbbbbb",
			"opencode", "deepseek-chat",
			"more cats marker-cat-1", "purr",
		),
	}
	return nil
}
```
