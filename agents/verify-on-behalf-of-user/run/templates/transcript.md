# Verify transcript — {{TITLE}}

| Field | Value |
|-------|-------|
| Scope | {{SCOPE}} |
| Depth | **{{DEPTH}}** |
| Depth reason | {{DEPTH_REASON}} |
| Surface | {{SURFACE}} |
| Started | {{STARTED}} |
| Sandbox HOME | ~/.sandbox/default-home |
| Sandbox bin | ~/.sandbox/bin |
| Verdict | **{{VERDICT}}** |

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

## 2. Build & install (sandbox bin)

> **Annotation:** Binaries install to ~/.sandbox/bin — not ~/go/bin.

```sh
$ source .agents/skills/verify-on-behalf-of-user/scripts/enter-sandbox.sh
{{ENTER_SANDBOX_OUTPUT}}

$ go build -o "$SANDBOX_BIN/{{BINARY}}" {{BUILD_TARGET}}
{{BUILD_OUTPUT}}
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
| Depth | {{DEPTH}} ({{DEPTH_REASON}}) |
| Surface | {{SURFACE}} |
| Verdict | **{{VERDICT}}** |
| Transcript file | `{{TRANSCRIPT_PATH}}` |

Sandbox data (inspect after verify): `~/.sandbox/default-home`

**Agent:** after saving this file, print its **full contents** in the reply for direct review.
