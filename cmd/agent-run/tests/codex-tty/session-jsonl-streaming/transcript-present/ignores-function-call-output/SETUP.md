# Scenario

**Feature**: `function_call_output` is diagnostic context, not assistant stdout by default

```
rollout JSONL response_item.function_call_output
  -> diagnostic tool output is observed by tailer
  -> no assistant stdout line is emitted for that output
```

## Preconditions

- The transcript also contains a real assistant message so the run has useful structured output.

## Steps

1. Seed a `function_call_output` record and an assistant message.
2. Assert stdout includes the assistant message but not the function output marker.

```go
import "testing"

const codexFunctionOutputText = "JSONL_FUNCTION_OUTPUT_SHOULD_NOT_PRINT"
const codexAssistantAfterFunctionText = "JSONL_ASSISTANT_AFTER_FUNCTION_OUTPUT"

func Setup(t *testing.T, req *Request) error {
	seedCodexTranscript(t, req,
		codexFunctionCallOutputLine(codexFunctionOutputText),
		codexAgentMessageLine(codexAssistantAfterFunctionText),
	)
	req.StreamProbeSubstring = codexAssistantAfterFunctionText
	return nil
}
```
