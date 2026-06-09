```go
import "testing"

type Request struct {
	Value int
}

type Response struct {
	Result int
}

func Run(t *testing.T, req *Request) (*Response, error) {
	return &Response{Result: req.Value}, nil
}
```
