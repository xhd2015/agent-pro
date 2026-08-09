# Scenario

**Feature**: empty store with --grep still uses empty-store message

```
# no sessions; explain list --grep anything
-> "No explain sessions yet.\n" (not "No matching…"); exit 0
```

## Preconditions

- Isolated ConfigHome with no sessions.

## Steps

1. Leave Sessions empty.
2. Args: `list --grep anything`.
3. Assert empty-store wording, not no-match wording.

## Context

- Empty store path is unchanged when greps are present.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"list", "--grep", "anything"}
	req.Sessions = nil
	return nil
}
```
