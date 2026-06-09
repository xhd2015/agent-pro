```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.Result != 42 {
		t.Fatalf("expected 42, got %d", resp.Result)
	}
}
```
