# Scenario

**Feature**: human rendering of the usage overlay at a fixed width

```
# the view arrives already projected; FormatUsage only lays it out
View + FormatOptions{Width: 80, Color: false} -> overlay text without ANSI
```

## Preconditions

- This branch only covers `req.Format = "human"`.

## Steps

1. Set `req.Format = "human"` and pin `req.Width = 80`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
req.Format = "human"
req.Width = 80
return nil
}
```
