# Scenario

**Feature**: focus-specific help documents selection

```
User -> focus --help -> focus usage
```

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Args = []string{"--help"}
	return nil
}
```
