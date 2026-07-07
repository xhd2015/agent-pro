## Expected

- Upstream fake ptywrap records browser keyboard input bytes.
- Websocket client receives echoed PTY output.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	gotInput := strings.Join(req.RegistryInputs, "|")
	if !strings.Contains(gotInput, "hello from browser") {
		t.Fatalf("keyboard input did not reach PTY; upstream inputs=%q", gotInput)
	}
	if !containsAny(resp.WSOutput, "echo:hello from browser", req.RegistryTranscript) {
		t.Fatalf("websocket did not return PTY echo; output=%q", resp.WSOutput)
	}
}
```
