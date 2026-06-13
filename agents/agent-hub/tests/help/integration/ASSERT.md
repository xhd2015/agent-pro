## Expected
- Exit code 0.
- Stdout contains "Usage:", "integration", lists status/install/uninstall/enable/disable subcommands.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertExitCode(t, resp, 0)
    assertContains(t, resp.Stdout, "Usage:")
    assertContains(t, resp.Stdout, "integration")
    assertContains(t, resp.Stdout, "status")
    assertContains(t, resp.Stdout, "install")
    assertContains(t, resp.Stdout, "uninstall")
    assertContains(t, resp.Stdout, "enable")
    assertContains(t, resp.Stdout, "disable")
}
```
