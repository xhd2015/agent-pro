package agenttty

import "os"

// openCloseExitsExperiment enables conceptual close→exit for --open:
//
//	AGENT_RUN_OPEN_CLOSE_EXITS=1
//
// Mechanism (see pkgs/agenttty/run.go):
//  1. AttachWriter uses attach_mode=screen (ptywrap roleWriter) so bare WS
//     disconnect without {"type":"detach_keep"} calls stopChild() — kills the
//     PTY agent when the user closes the iTerm window (attach client dies).
//  2. --open does not force KeepTerminalAlive, so after the child exits,
//     __serve__ shuts down (tty-watch ServeSession non-keep-alive path) instead
//     of waiting forever on ctx.Done().
//
// --detach is unchanged (daemon is intentional).
// Ctrl-] still sends detach_keep → child kept (intentional detach).
//
// This is an experiment gate for human debug; not production default.
func openCloseExitsExperiment() bool {
	return OpenCloseExitsExperiment()
}

// OpenCloseExitsExperiment is the exported env gate (AGENT_RUN_OPEN_CLOSE_EXITS=1).
// Used by agentui / agentrunapi so KeepTerminalAlive is not re-forced for --open.
func OpenCloseExitsExperiment() bool {
	v := os.Getenv("AGENT_RUN_OPEN_CLOSE_EXITS")
	return v == "1" || v == "true" || v == "yes"
}
