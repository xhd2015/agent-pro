## Steps
1. The converter-router handles codex events. See leaf SETUP.md for details.

```go
import ("testing")

func Setup(t *testing.T, req *Request) error {
    _ = t
    _ = req
    return nil
}
```
