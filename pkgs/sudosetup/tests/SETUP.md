# Scenario

**Feature**: sudosetup doctest harness with injectable FS and Runner

```
# leaf Setup seeds Request; root Run builds Manager with mapFS + recordingRunner
doctest Setup chain -> mapFS seed -> sudosetup.Manager -> Response + RunnerCalls

# no real sudo, visudo, or /etc writes
Runner fake -> records sudo/visudo/install/rm; FS map -> in-memory paths only
```

## Preconditions

- Package `github.com/xhd2015/agent-pro/pkgs/sudosetup` exposes `Manager`, `Config`,
  `Rule`, `Status`, injectable `FS` and `Runner`.
- Each test uses an isolated `mapFS` rooted at `t.TempDir()`; `/etc/sudoers.d` and
  user cache paths are redirected under that root.
- `Manager` panics under `go test` if `FS` or `Runner` is nil (production deps
  would touch `/etc/sudoers.d` and real `sudo`); tests must inject fakes.
- `StdinIsTerminal` on `Request` drives `Manager.StdinIsTerminal` (RootSetup
  defaults true); install/remove tests can set false to assert non-TTY errors.
- `recordingRunner` fakes `sudo -n`, `visudo`, `install`, `rm`, and `sudo -k` per
  `Request` flags — never spawns real subprocesses.

## Steps

1. Root `Setup` sets default `Config`, `Rule`, and username env (`testuser`).
2. Group/leaf `Setup` sets `Operation`, `Action`, and scenario-specific seeds.
3. `Run` builds `Manager`, executes operation, returns `Response`.
4. Leaf `Assert` checks status, filesystem state, and runner audit trail.

## Context

- Default POC config: `CacheDirName=remote-agent-sudo-poc`, `SudoersName=remote-agent-sudo-poc`.
- Default VPN rule variant uses `CacheDirName=remote-agent`, `SudoersName=remote-agent-sing-box`.
- Manifest filename: `sudo-setup-manifest.json`.
- `mapFS` implements package `FS` with path redirection for `/etc/sudoers.d/*` and
  `UserCacheDir/<CacheDirName>/*`.

```go
import (
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/sudosetup/sudosetuptest"
)

func Setup(t *testing.T, req *Request) error {
	return sudosetuptest.RootSetup(t, req)
}

func installedSeedLine(username, command, argsPattern string) string {
	return sudosetuptest.InstalledSeedLine(username, command, argsPattern)
}

func installedManifestSeed(username, command, argsPattern string) *ManifestSeed {
	return sudosetuptest.InstalledManifestSeed(username, command, argsPattern)
}

func assertNoError(t *testing.T, err error) {
	sudosetuptest.AssertNoError(t, err)
}

func assertError(t *testing.T, err error) {
	sudosetuptest.AssertError(t, err)
}

func assertEqual(t *testing.T, field string, got, want any) {
	sudosetuptest.AssertEqual(t, field, got, want)
}

func assertContains(t *testing.T, got, want string) {
	sudosetuptest.AssertContains(t, got, want)
}

func hasRunnerCall(calls []RunnerCall, name string, argPrefix ...string) bool {
	return sudosetuptest.HasRunnerCall(calls, name, argPrefix...)
}

func runnerCallCount(calls []RunnerCall, name string, argPrefix ...string) int {
	return sudosetuptest.RunnerCallCount(calls, name, argPrefix...)
}

func detailContains(detail, sub string) bool {
	return sudosetuptest.DetailContains(detail, sub)
}
```