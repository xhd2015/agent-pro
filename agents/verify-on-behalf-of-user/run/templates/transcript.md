# Verify transcript — {{TITLE}}

| Field | Value |
|-------|-------|
| Scope | {{SCOPE}} |
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

## 3. Smoke commands

> **Annotation:** {{SMOKE_ANNOTATION}}

```sh
{{SMOKE_COMMANDS}}
```

> **Check:** {{SMOKE_CHECK}}

---

## 4. On-disk / artifacts (optional)

```sh
{{ARTIFACT_COMMANDS}}
```

> **Check:** {{ARTIFACT_CHECK}}

---

## 5. Browser / logs (optional)

> **Annotation:** {{BROWSER_ANNOTATION}}

```sh
{{BROWSER_OR_LOGS}}
```

> **Check:** {{BROWSER_CHECK}}

---

## 6. Doctest spot-check (optional)

```sh
{{DOCTEST_COMMANDS}}
```

> **Check:** {{DOCTEST_CHECK}}

---

## Summary

{{SUMMARY}}

Transcript file: `{{TRANSCRIPT_PATH}}`

Sandbox data (inspect after verify): `~/.sandbox/default-home`