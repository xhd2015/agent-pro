package agenttty

import (
	"os"
	"strings"
)

// OpenCloseExits reports whether --open should tear down the TTY when the
// interactive window/attach ends (red-close ≈ exit).
//
// Production default: true.
//
// Mechanism (pkgs/agenttty/run.go):
//  1. AttachWriter uses attach_mode=screen (ptywrap roleWriter) so bare WS
//     disconnect without {"type":"detach_keep"} calls stopChild() — kills the
//     PTY agent when the user closes the iTerm window (attach client dies).
//  2. --open does not force KeepTerminalAlive, so after the child exits,
//     __serve__ shuts down (tty-watch ServeSession non-keep-alive path) instead
//     of waiting forever on ctx.Done().
//
// --detach is unchanged (daemon is intentional).
// Explicit --keep-tty still forces KeepTerminalAlive.
// Ctrl-] still sends detach_keep → child kept (intentional detach).
//
// Opt out (debug / old keep-alive-after-open behavior):
//
//	AGENT_RUN_OPEN_CLOSE_EXITS=0
func OpenCloseExits() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("AGENT_RUN_OPEN_CLOSE_EXITS")))
	switch v {
	case "0", "false", "no", "off":
		return false
	default:
		// empty, "1", "true", "yes", or any other value → production default on
		return true
	}
}

// OpenCloseExitsExperiment is a deprecated alias of OpenCloseExits.
// Prefer OpenCloseExits.
func OpenCloseExitsExperiment() bool {
	return OpenCloseExits()
}

// openCloseExitsExperiment is the unexported alias used inside this package.
func openCloseExitsExperiment() bool {
	return OpenCloseExits()
}
