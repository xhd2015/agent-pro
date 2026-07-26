# Scenario

**Feature**: QuestionsEnabled=false disables questions FIFO infrastructure

```
# no questions/, no QUESTION_FIFO, no QUESTIONS stdout footer
subagent.Run(SessionLayout{QuestionsEnabled:false}) -> skip questions feature
```

## Preconditions

- Flat session dir; other features enabled.

## Steps

1. Set `QuestionsEnabled: false` on layout.

## Context

- stdout must not end with QUESTIONS block.

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
		QuestionsEnabled: false,
		ProgressEnabled:  true,
	}
	return nil
}
```
