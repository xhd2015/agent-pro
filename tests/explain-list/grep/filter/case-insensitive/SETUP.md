# Scenario

**Feature**: --grep is case-insensitive; listed body keeps original casing

```
# session Q has "Docker"; explain list --grep docker
-> session kept; stdout still shows "Docker" (not lowercased)
```

## Preconditions

- One session with mixed-case body containing `Docker`.

## Steps

1. Seed session with Q `How does Docker networking work?`.
2. Args: `list --grep docker`.
3. Assert session listed and original `Docker` casing present.

## Context

- Match is case-insensitive literal; re-casing bodies is forbidden.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"list", "--grep", "docker"}
	req.Sessions = []SessionSeed{
		simpleSession(
			"2026-07-14-09-00-00-caseci-aaaaaaaa",
			"opencode", "deepseek-chat",
			"How does Docker networking work?",
			"Containers share a bridge network.",
		),
	}
	return nil
}
```
