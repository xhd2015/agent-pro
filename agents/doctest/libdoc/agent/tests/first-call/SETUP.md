## Preconditions
- Tests in this group verify the first invocation of `doctest agent implement`.
- No `CODEX_THREAD_ID` is set beforehand.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.Env = append(req.Env, "TEST_GROUP=first-call")
    return nil
}
```
