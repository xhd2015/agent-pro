```go
type Request struct{ Value int }
type Response struct{ Result int }

func Setup(t *testing.T, req *Request) error { _ = req; return nil }
func Run(t *testing.T, req *Request) (*Response, error) { return &Response{Result: req.Value}, nil }
```
