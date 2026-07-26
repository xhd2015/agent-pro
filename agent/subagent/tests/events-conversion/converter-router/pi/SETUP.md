## Steps
1. The converter-router handles pi events. See leaf SETUP.md for details.

```go
import ("testing")

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    _ = t
    _ = req
    return nil
}
```
