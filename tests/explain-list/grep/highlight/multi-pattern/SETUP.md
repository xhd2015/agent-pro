# Scenario

**Feature**: multi-pattern --color highlights every provided pattern on kept cards

```
# Q has both "docker" and "kubernetes"; --grep docker --grep kubernetes --color
-> bold-red around both match spans
```

## Preconditions

- One session whose body contains both patterns (distinct non-overlapping spans).

## Steps

1. Seed dual-pattern session.
2. Args: `list --grep docker --grep kubernetes --color`.
3. Assert both original-case spans wrapped in bold red.

## Context

- Every non-overlapping occurrence of every provided pattern is highlighted.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"list", "--grep", "docker", "--grep", "kubernetes", "--color"}
	req.Sessions = []SessionSeed{
		simpleSession(
			"2026-07-17-15-00-00-hl-multi-mmmmmmmm",
			"opencode", "deepseek-chat",
			"Compare docker and kubernetes networking",
			"Both orchestrate containers.",
		),
	}
	return nil
}
```
