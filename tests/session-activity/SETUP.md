# Scenario

**Feature**: derive live agent activity from a durable log or last chat message

```
events + live  -> activity kind
last chat kind + live -> activity kind
```

## Preconditions

- Classification is pure: no DSH home, iTerm, or process list.

## Steps

1. Classify the supplied DSH events or chat last-message.

## Context

- Default source is DSH unless a grouping sets `chat`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Source = "dsh"
	return nil
}
```
