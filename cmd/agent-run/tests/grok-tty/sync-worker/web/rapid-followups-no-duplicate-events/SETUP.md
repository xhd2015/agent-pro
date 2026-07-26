# Scenario

**Bug**: I1 — reproduces web_d6b4b203cc9ff71a duplicate follow-up events

```
POST hello? -> wait done
  -> POST what did I say? (2s gap, within 90s tail overlap)
  -> events.jsonl file-level count: 1 user per prompt
```

## Steps

1. Configure web grok-tty env with llm-mock chrome hook.
2. Start `agent-run web`; POST initial session with `hello?`.
3. Schedule turn 1 completion; wait for first `done`.
4. POST follow-up `what did I say?` after `FollowUpGap`.
5. Schedule turn 2 completion; wait for second `done`.

```go
import (
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Mode = "web-rapid-followups"
	req.PromptA = "hello?"
	req.PromptB = "what did I say?"
	req.ReplyA = "sync-web-reply-hello"
	req.ReplyB = "sync-web-reply-recall"
	req.FollowUpGap = 2 * time.Second
	req.CompletionDelayTurn1 = 1200 * time.Millisecond
	req.CompletionDelayTurn2 = 1200 * time.Millisecond
	req.ChromeHoldSeconds = 60
	return nil
}
```
