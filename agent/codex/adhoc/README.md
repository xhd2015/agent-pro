# Process-scoped Codex Responses connections

`Connection` is a versioned, non-secret session binding: provider identity, base URL, **raw** upstream model, reasoning policy, and a credential reference. It is independent of Marcus, model-directory discovery, and llm-proxy.

```go
connection := &adhoc.Connection{
    Version: adhoc.Version,
    Provider: "my-gateway",
    BaseURL: "http://127.0.0.1:9000/v1",
    Model: "raw-upstream-model",
    Reasoning: adhoc.ReasoningPolicy{
        Requested: "high", Effort: "high",
        Supported: []string{"low", "high"}, Origin: "api",
    },
    Auth: adhoc.Credential{File: "/existing/private/token"},
}
// Set agentrunapi.Opts.CodexConnection or agentrunapi.RunOpts.CodexConnection.
// The host application must dispatch agentruncli.Handle for re-exec, including
// the internal __codex-adhoc-exec receiver. Native sessions need no connection.
```

The normal lifecycle reserves the binding before launch, restores it on retry/resume, and preserves existing native sessions. `Apply(argv, interactive)` is the lower-level pure compiler: call it before appending prompts/separators. It returns process-only provider overrides, `responses`, a raw model, and TUI `--no-daemon`; it does not read a credential or write config. Caller model/provider/effort conflicts fail closed.

`SupportedEfforts` means direct wire-compatible values verified with Codex 0.166.0, not a claim that every provider/model supports them. `max` passes through. CLI `ultra` maps to wire `xhigh`, so is excluded from this direct-effort vocabulary. Newer/older Codex versions need re-verification; unsupported flags/values must fail rather than route elsewhere.

Credentials may reference a process environment key, existing token file, or JSON pointer with non-secret identity guards. Optional SHA256 pins are available when value identity must also remain fixed. Use `PublicToken` only for an explicitly public dummy marker, never a real credential. Do not put secrets in guards, connection fields, session Env, argv, logs, or UI DTOs. `ValidatePersistedEnv` enforces the owned credential-variable boundary; the final receiver resolves the reference immediately before replacing the Codex process. On Windows it uses an stdio-connected child instead of Unix exec.

The actual host driver must support `--require-codex-connection-v1`; use `agentdriver.VerifyCodexAdhoc` rather than relying on the parent SDK's version. Follow-up commands must carry the guard and the same store home. A stale receiver or missing/corrupt binding fails closed.

Tests: public compile/reference tests, agentrunapi lifecycle hooks, real receiver subprocess boundary, plus opt-in `AGENT_PRO_TEST_REAL_CODEX=1 go test -v ./agent/codex/adhoc`. Real Codex only contacts local SSE mocks. GPT-5.5 covers conventional shell tools; GPT-6.1-Sol covers text/SSE and effort mapping, not code-mode tool execution. Temporary config fixtures are test isolation, not a production configuration writer.
