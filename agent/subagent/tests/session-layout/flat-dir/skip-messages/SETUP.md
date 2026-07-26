# Scenario

**Feature**: empty MessagesPath skips messages.jsonl writes

```
# host owns prompt history elsewhere
subagent.Run(SessionLayout{MessagesPath:""}) -> no messages.jsonl
```

## Preconditions

- Flat session dir with other artifacts still written.

## Steps

1. Configure flat dir with `MessagesPath: ""`.

## Context

- events.jsonl should still be created.

```go
import (
	"testing"

	"github.com/xhd2015/agent-pro/agent/subagent"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	configureFlatDirBase(t, req)
	req.Layout = subagent.SessionLayout{
		Dir:              req.SessionDir,
		MessagesPath:     "",
		QuestionsEnabled: true,
		ProgressEnabled:  true,
	}
	return nil
}
```
