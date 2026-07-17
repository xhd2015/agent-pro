# Scenario

**Feature**: pkgs/agentruncli is a library package, not package main

```
# scan production sources
pkgs/agentruncli/*.go
  -> package agentruncli  (not package main)
```

## Preconditions

- Module layout: DOCTEST_ROOT = tests/agentruncli → `../../pkgs/agentruncli`.
- Production `.go` only (skip `_test.go`).

## Steps

1. Mode `not_package_main`.
2. Run parses package clause of first production source.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = "not_package_main"
	return nil
}
```
