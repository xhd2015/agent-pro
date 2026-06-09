```go
import "testing"

type Request struct {
	Bad bool
}

func Setup(t *testing.T, req *Request) error {
	req.Value = 5
	return nil
}
```
