```go
import "testing"

type Response struct {
	Result int
}

func Run(t *testing.T, req *Request) (*Response, error) {
	return &Response{Result: 1}, nil
}
```
