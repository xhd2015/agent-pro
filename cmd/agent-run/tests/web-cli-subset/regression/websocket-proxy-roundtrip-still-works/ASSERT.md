## Expected

- Websocket receives PTY transcript and echoed browser input (same contract as websocket-proxy/round-trip-io).

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if !containsAny(resp.WSOutput, req.RegistryTranscript, "echo:hello from browser", "hello from browser") {
		t.Fatalf("websocket proxy regression failed; output=%q error=%q", resp.WSOutput, resp.WSError)
	}
}
```
