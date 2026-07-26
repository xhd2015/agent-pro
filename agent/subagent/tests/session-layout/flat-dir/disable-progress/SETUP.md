# Scenario

**Feature**: ProgressEnabled=false disables progress reporting infrastructure

```
# no progress/, no PROGRESS_FILE, no report-progress watch
subagent.Run(SessionLayout{ProgressEnabled:false}) -> skip progress feature
```

## Preconditions

- Flat session dir; questions enabled.

## Steps

1. Set `ProgressEnabled: false` on layout.

## Context

- events.jsonl and messages.jsonl still expected.

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
		MessagesPath:     "messages.jsonl",
		QuestionsEnabled: true,
		ProgressEnabled:  false,
	}
	return nil
}
```
