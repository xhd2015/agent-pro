# Scenario

**Bug**: interactive bash profile errors smear across terminal width when carriage-return redraw fails

```
# init file runs undefined PS1 helpers; bash prints errors then accepts input
harness PTY -> tty-watch run bash --init-file <fake> -i -> profile errors at column 0
harness PTY -> echo LAYOUT_OK -> LAYOUT_OK visible without padded bash: lines
```

Bare `shortpath` and `parse_git_branch` commands in the init file mimic a broken `.bashrc` that
runs undefined helpers during PS1 setup. Harness attaches with **pipe stdout** (non-TTY) so `attachStdoutWriter` exercises the
`\r`-stripping path that smears interactive bash profile errors.

## Steps

1. `writeFakeBashInit` writes the init file under `TTY_WATCH_HOME`.
2. Run `bash --init-file <init> -i` (via `exec bash` after seeding the `\r`-delimited
   profile-error burst the PTY emits before the interactive session).
3. After 1s, write `echo LAYOUT_OK\n` on the attach stdin pipe.
4. Read until `LAYOUT_OK` or 8s timeout.

Carriage-return redraw must keep `bash:` errors at the left margin; stripping `\r` smears them
far right.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "run-interactive-bash-layout"
	return nil
}
```