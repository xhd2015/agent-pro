# Scenario

**Feature**: update banner + model loading is not ready for `/status` injection

```
snapshot has Update available banner + "model: loading" + prompt placeholder ›
  -> CheckWritable returns ready=false, state=loading
```

## Preconditions

- Fixture `codex-update-plus-model-loading.txt` from live capture
  (`/tmp/codex-status-fixtures-for-req/update-plus-model-loading.txt`).
- Existing `model:loading` detection may already yield `loading`; leaf locks that outcome.

## Steps

1. Set `req.FixtureFile` to the update+model-loading fixture.

## Context

- F3 guards boot-incomplete screens where model is still loading (with or without update banner).

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.FixtureFile = fixtureModelLoading
	return nil
}
```
