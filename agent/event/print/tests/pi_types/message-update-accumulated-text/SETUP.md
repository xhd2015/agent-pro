## Preconditions
- A `message_update` event where the accumulated text in `message.content[0].text` is much larger than the `assistantMessageEvent.delta`.
- This mimics the actual pi streaming behavior: each event carries the full accumulated text.
- After the fix, FormatTraceLine must use only the delta, NOT the full accumulated Content text, preventing duplication in `--trace` output.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Line = `{"type":"message_update","message":{"role":"assistant","content":[{"type":"text","text":"The user has given me a detailed requirement for creating a macOS menu bar app. I need to design a comprehensive doctest tree for this feature."}]},"assistantMessageEvent":{"type":"text_delta","delta":" feature."}}`
	return nil
}
```
