## Expected
- No opencode config directory is created under temporary HOME.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    if resp.HostConfigExists {
        t.Fatal("fake-opencode wrote host config")
    }
}
```

