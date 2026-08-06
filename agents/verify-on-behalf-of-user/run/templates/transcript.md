# Verify transcript — {{TITLE}}

| Field | Value |
|-------|-------|
| Mode | **{{MODE}}** |
| Scope | {{SCOPE}} |
| Depth | **{{DEPTH}}** |
| Depth reason | {{DEPTH_REASON}} |
| Surface | {{SURFACE}} |
| Started | {{STARTED}} |
| HOME / bin | {{HOME_BIN_NOTE}} |
| Install method | {{INSTALL_METHOD}} |
| Targets | {{TARGETS}} |
| Verdict | **{{VERDICT}}** |

> **Mode:** `sandbox` (default) or `host` (only when user explicitly opted into
> host / outside sandbox). Host: warn + dry-run/plan before mutate; change-scoped
> install only. See topics sandbox and host.

---

## 1. Git sanity

> **Annotation:** {{GIT_ANNOTATION}}

```sh
$ pwd
{{PWD}}

$ git status --short
{{GIT_STATUS}}

$ git diff --stat
{{GIT_DIFF_STAT}}
```

> **Check:** {{GIT_CHECK}}

---

## 2. Build & install

> **Annotation:** {{BUILD_ANNOTATION}}
>
> Sandbox: binaries under ~/.sandbox/bin after enter-sandbox.sh.
> Host: wrk --reinstall-local (dry-run then apply) or change-scoped
> script/install / go install ./cmd/T only (T = claimed tool name).

```sh
{{BUILD_COMMANDS}}
```

> **Check:** {{BUILD_CHECK}}

---

## 3. Runtime bring-up

> **Annotation:** {{RUNTIME_ANNOTATION}}

```sh
{{RUNTIME_COMMANDS}}
```

> **Check:** {{RUNTIME_CHECK}}

---

## 4. Scenarios

> **Annotation:** {{SCENARIO_ANNOTATION}}
> Depth is **{{DEPTH}}** — smoke is only a labeled downgrade when justified.

### S1 — {{S1_TITLE}}

```sh
{{S1_COMMANDS}}
```

> **Check:** {{S1_CHECK}}

### S2 — {{S2_TITLE}} (optional)

```sh
{{S2_COMMANDS}}
```

> **Check:** {{S2_CHECK}}

---

## 5. Evidence & teardown

```sh
{{EVIDENCE_AND_TEARDOWN}}
```

> **Check:** {{EVIDENCE_CHECK}}

---

## 6. Browser-agent (UI surfaces)

> **Annotation:** UI verify uses **browser-agent only** (not playwright-debug). Missing UI path → **FAIL**.

```sh
{{BROWSER_AGENT_COMMANDS}}
```

> **Check:** {{BROWSER_CHECK}}

---

## 7. Doctest spot-check (optional)

```sh
{{DOCTEST_COMMANDS}}
```

> **Check:** {{DOCTEST_CHECK}}

---

## Summary

{{SUMMARY}}

| Meta | Value |
|------|-------|
| Mode | {{MODE}} |
| Depth | {{DEPTH}} ({{DEPTH_REASON}}) |
| Surface | {{SURFACE}} |
| Verdict | **{{VERDICT}}** |
| Transcript file | `{{TRANSCRIPT_PATH}}` |

Sandbox data (inspect after sandbox-mode verify): `~/.sandbox/default-home`

**Agent:** after saving this file, print its **full contents** in the reply for direct review.
