# Scenario

**Feature**: ambiguity without --index is fatal

```
focus <id> -> two candidates -> list and stop
```

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = t
	_ = d
	req.Args = []string{req.SessionID}
	return nil
}
```
