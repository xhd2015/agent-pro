## Expected
- The command accepts `--dir`.
- Host opencode config is not created.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    if resp.HostConfigExists {
        t.Fatal("fake-opencode wrote host opencode config")
    }
}
```

