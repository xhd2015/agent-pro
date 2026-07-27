# Scenario

**Feature**: assistant `response_item.message` output_text records stream as assistant text

```
rollout JSONL response_item
  -> payload.type=message, role=assistant
  -> content output_text fragments
  -> assistant stdout
```

## Preconditions

- Non-assistant roles are outside this leaf; this leaf covers the valid assistant role.

## Steps

1. Seed the rollout JSONL with an assistant `response_item.message`.
2. Assert output text content is emitted to stdout.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

const codexResponseItemText = "JSONL_RESPONSE_ITEM_ASSISTANT"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	seedCodexTranscript(t, req, codexAssistantResponseItemLine(codexResponseItemText))
	req.StreamProbeSubstring = codexResponseItemText
	return nil
}
```
