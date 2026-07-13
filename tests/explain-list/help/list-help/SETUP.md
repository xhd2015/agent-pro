# Scenario

**Feature**: explain list --help mentions --limit and --color

```
explain list --help -> usage text includes --limit and --color
```

## Preconditions

- No sessions required for help.

## Steps

1. Args: `list --help`.
2. Sessions empty.

## Context

- Soft content match (implementer owns exact help wording).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"list", "--help"}
	req.Sessions = nil
	return nil
}
```
