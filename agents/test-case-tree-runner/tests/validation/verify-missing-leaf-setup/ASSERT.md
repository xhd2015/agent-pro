## Expected
- Compile passes — leaf has `func Run` which satisfies the "Setup or Run" requirement

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if !resp.Passed {
		t.Fatalf("expected compile to pass (leaf has Run), got:\n%s", resp.Output)
	}
}
```
