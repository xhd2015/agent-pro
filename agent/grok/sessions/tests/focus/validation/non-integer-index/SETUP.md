# Scenario

**Feature**: a non-numeric --index is a parse error

```
User -> focus <id> --index x -> parse error
```

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Args = []string{fixtureFocusSessionID, "--index", "x"}
	return nil
}
```
