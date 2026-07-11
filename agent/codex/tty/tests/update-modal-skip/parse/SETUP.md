# Scenario

**Feature**: offline ParseStatusSnapshot on signed post-Skip status screen

```
05-status-fields.snapshot.txt -> ParseStatusSnapshot monthly/credits/reset
```

## Preconditions

No live Codex; fixture only.

## Steps

1. Set `Op=parse`.
2. Leaf sets fixture basename.

## Context

Ensures field regexes still match real TUI chrome (PROTOCOL continue_status).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Op = "parse"
	return nil
}
```
