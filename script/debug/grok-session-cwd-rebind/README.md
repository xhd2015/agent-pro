# grok-session-cwd-rebind

Experiment / inspect: can a Grok session’s **tool cwd** be rebound after the
original workspace directory is deleted?

## Setup

- `grok` on PATH
- `~/.grok/auth.json` (copied into an isolated `GROK_HOME`)

## Run

```sh
# REPRO: delete old workspace, resume with --cwd NEW, no meta patch → exit 1 + REPRO:
go run ./script/debug/grok-session-cwd-rebind --expect=repro

# VERIFY: apply patches then resume → exit 0 + PASS:/VERIFY:
go run ./script/debug/grok-session-cwd-rebind --expect=healthy --patch=move_dir
```

## Findings (2026-07-13)

| Patch | Tool cwd rebounds? |
|-------|--------------------|
| `summary` only | no |
| `prompt_context` only | no |
| `summary,prompt_context` | no |
| **`move_dir` only** | **yes** |
| `summary,move_dir` / `prompt_context,move_dir` / full set | yes |

**Root cause:** sessions live under  
`$GROK_HOME/sessions/<url-encoded-cwd>/<session-id>/`.  
Grok keys session storage by cwd. **Renaming/moving that directory** to the
encoded **new** cwd is what rebinds tools. Editing `info.cwd` /
`working_directory` in place (without move) is insufficient.

Loop doc: `doc/LOOP_2026-07-13_grok-session-cwd-rebind-meta.md`
