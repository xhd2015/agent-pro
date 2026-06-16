# Proof: HOME Env Overrides Home Directory

These tests demonstrate that setting the `HOME` environment variable overrides
the home directory across multiple runtimes:

- **bash** — `$HOME` variable, `echo $HOME`
- **golang** — `os.UserHomeDir()`
- **nodejs** — `require('os').homedir()`
- **bun** — `require('os').homedir()`

Each test sets `HOME` to a temporary directory and verifies the runtime
resolves that directory as the user's home — confirming that `HOME` is the
authoritative source for home directory discovery.

These tests do **not** test any code in this project; they are purely
demonstrations of platform behavior.

## How to Run

```sh
doctest test ./tests/proof-of-home-dir
```
