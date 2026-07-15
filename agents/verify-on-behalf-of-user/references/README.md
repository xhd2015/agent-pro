# Verify recipes (consumer repos)

This skill is generic. Per-project smoke steps live in **consumer repositories**:

```
<your-repo>/docs/verify-recipes/<feature>.md
```

Each recipe should include:

1. Scope and expected git paths
2. Build commands (`go build -o "$SANDBOX_BIN/..."`)
3. Smoke command sequence (after `enter-sandbox.sh`)
4. Expected stdout/stderr and on-disk checks
5. Browser or log steps when applicable

The agent copies recipe commands into the verify transcript verbatim.