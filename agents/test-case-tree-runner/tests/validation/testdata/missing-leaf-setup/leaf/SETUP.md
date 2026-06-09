```go
import "testing"

func Run(t *testing.T, req *Request) (*Response, error) {
	return &Response{Result: req.Value + 1}, nil
}
```
