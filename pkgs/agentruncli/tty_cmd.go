package agentruncli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/xhd2015/agent-pro/pkgs/agentsend"
	"github.com/xhd2015/agent-pro/pkgs/agenttty"
	"github.com/xhd2015/less-gen/flags"
	"github.com/xhd2015/tty-watch/pkgs/ttywatch"
)

const ttyHelp = `
Usage: agent-run tty <subcommand> [ARGS]

Subcommands:
  status    show status of a TTY session
  attach    attach to a live TTY session interactively
  send      send a message to a live TTY session
  snapshot  print a sanitized snapshot of a live TTY session
  watch     stream readonly output from a live TTY session
  kill      stop a live TTY session by registry id

Options:
  -h, --help   show help

Run agent-run tty <subcommand> --help for subcommand-specific options.
`

const ttyStatusHelp = `
Usage: agent-run tty status [OPTIONS] <session-id>

Show status of a TTY session from the registry.

Options:
  --json     output as JSON
  -h, --help show help
`

const ttyAttachHelp = `
Usage: agent-run tty attach <session-id>

Attach to a live grok-tty or codex-tty session by registry id (printed on stderr during run).

Options:
  -h, --help   show help
`

const ttySendHelp = `
Usage: agent-run tty send [OPTIONS] <session-id> "message"

Send a follow-up message to a live TTY session via the per-session send queue.
On successful enqueue, prints the session-local message id (msg_1, msg_2, …) to stdout.

Options:
  --no-wait            enqueue and exit immediately without waiting for delivery
  --max-wait DURATION  enqueue, print id, then wait up to DURATION for delivery
  --no-submit          inject without trailing Enter (no auto-submit); stored on queue entry
  -h, --help           show help
`

const ttySnapshotHelp = `
Usage: agent-run tty snapshot <session-id>

Print a sanitized snapshot of a live TTY session.

Options:
  -h, --help   show help
`

const ttyWatchHelp = `
Usage: agent-run tty watch <session-id>

Stream readonly output from a live TTY session.

Options:
  -h, --help   show help
`

type ttyStatusData struct {
	PID             int    `json:"pid"`
	Port            string `json:"port"`
	TTYType         string `json:"tty_type"`
	SessionID       string `json:"session_id"`
	SessionFilePath string `json:"session_file_path,omitempty"`
	StartTime       string `json:"start_time"`
	TCPReachable    bool   `json:"tcp_reachable"`
	ScreenStatus    string `json:"screen_status,omitempty"`
	Sendable        bool   `json:"sendable"`
	SendableReason  string `json:"sendable_reason,omitempty"`
	SendableState   string `json:"sendable_state,omitempty"`
	InputBox        string `json:"input_box"`
}

func runTty(args []string) error {
	if len(args) == 0 {
		fmt.Print(strings.TrimPrefix(ttyHelp, "\n"))
		return nil
	}
	cmd := args[0]
	sub := args[1:]
	switch cmd {
	case "-h", "--help":
		fmt.Print(strings.TrimPrefix(ttyHelp, "\n"))
		if len(sub) == 0 {
			return nil
		}
		return runTty(sub)
	case "status":
		return runTtyStatus(sub)
	case "attach":
		return runTtyAttach(sub)
	case "send":
		return runTtySend(sub)
	case "snapshot":
		return runTtySnapshot(sub)
	case "watch":
		return runTtyWatch(sub)
	case "kill":
		return runKill(sub)
	default:
		return fmt.Errorf("unknown tty subcommand: %s", cmd)
	}
}

func resolveRegistryEntry(sessionID string) (*ttywatch.RegistryEntry, string, error) {
	store, err := openStore()
	if err != nil {
		return nil, "", err
	}
	return agenttty.LookupSession(store.Home(), sessionID)
}

func runTtyStatus(args []string) error {
	var jsonFlag bool
	remaining, err := flags.Bool("--json", &jsonFlag).
		Help("-h,--help", ttyStatusHelp).
		Parse(args)
	if err != nil {
		return err
	}
	if len(remaining) == 0 {
		return fmt.Errorf("tty status: requires <session-id>")
	}
	sessionID := remaining[0]

	store, err := openStore()
	if err != nil {
		return err
	}
	home := store.Home()

	ttySess, err := agenttty.ResolveByTerminalID(home, sessionID)
	if err != nil {
		return err
	}

	agentSessionID := ttySess.AgentSessionID
	if agentSessionID == "" {
		agentSessionID = sessionID
	}
	sessionFilePath := filepath.Join(home, "sessions", agentSessionID, "meta.json")
	tcpReachable := ttySess.TCPReachable
	screenStatus := ttySess.ScreenStatus
	if !tcpReachable && screenStatus == "" {
		screenStatus = "unknown"
	}

	var writable agenttty.WritableStatus
	var scrollbackText string
	provider, ok := agenttty.Get(ttySess.RunnerID)
	if tcpReachable && ok && provider.CheckWritable != nil {
		text, err := ttywatch.SnapshotText(ttySess.Registry.ListenAddr, sessionID)
		if err == nil && len(text) > 0 {
			scrollbackText = text
			scrollback := []byte(scrollbackText)
			writable = provider.CheckWritable(scrollback)
			if provider.DetectScreenStatus != nil &&
				(screenStatus == "" || screenStatus == "unknown") {
				if live := provider.DetectScreenStatus(scrollback); live != "" && live != "unknown" {
					screenStatus = live
				}
			}
		} else {
			writable = agenttty.WritableStatus{Reason: "no terminal output", State: "unknown"}
		}
	} else {
		writable = agenttty.WritableStatus{Reason: "terminal unreachable", State: "unreachable"}
	}

	inputToken := strings.TrimSpace(fmt.Sprint(agenttty.DetectInputBox(scrollbackText)))
	inputHuman, inputJSON := agenttty.InputBoxReport(inputToken)

	data := ttyStatusData{
		PID:             ttySess.Registry.PID,
		Port:            ttySess.Registry.ListenAddr,
		TTYType:         ttySess.RunnerID,
		SessionID:       ttySess.Registry.SessionID,
		SessionFilePath: sessionFilePath,
		StartTime:       ttySess.Registry.CreatedAt,
		TCPReachable:    tcpReachable,
		ScreenStatus:    screenStatus,
		Sendable:        writable.Ready,
		SendableReason:  writable.Reason,
		SendableState:   writable.State,
		InputBox:        inputJSON,
	}

	if jsonFlag {
		enc := json.NewEncoder(os.Stdout)
		return enc.Encode(data)
	}

	fmt.Printf("pid: %d\n", data.PID)
	fmt.Printf("port: %s\n", data.Port)
	fmt.Printf("tty type: %s\n", data.TTYType)
	fmt.Printf("session id: %s\n", data.SessionID)
	fmt.Printf("session file path: %s\n", data.SessionFilePath)
	fmt.Printf("start time: %s\n", data.StartTime)
	tcpLabel := "reachable"
	if !data.TCPReachable {
		tcpLabel = "unreachable"
	}
	fmt.Printf("tcp reachable: %s\n", tcpLabel)
	if data.ScreenStatus != "" {
		fmt.Printf("screen status: %s\n", data.ScreenStatus)
	}
	if data.Sendable {
		fmt.Printf("sendable: yes\n")
	} else if data.SendableReason != "" {
		fmt.Printf("sendable: no (%s)\n", data.SendableReason)
	} else {
		fmt.Printf("sendable: no\n")
	}
	fmt.Println(inputHuman)
	return nil
}

func runTtyAttach(args []string) error {
	remaining, err := flags.Help("-h,--help", ttyAttachHelp).Parse(args)
	if err != nil {
		return err
	}
	if len(remaining) != 1 {
		return fmt.Errorf("attach: requires <session-id>")
	}
	sessionID := remaining[0]

	entry, _, err := resolveRegistryEntry(sessionID)
	if err != nil {
		return err
	}

	_, err = ttywatch.AttachWriter(entry.ListenAddr, sessionID, "attach")
	return err
}

func runTtySend(args []string) error {
	var noWait bool
	var maxWait time.Duration
	var noSubmit bool
	remaining, err := flags.Bool("--no-wait", &noWait).
		Duration("--max-wait", &maxWait).
		Bool("--no-submit", &noSubmit).
		Help("-h,--help", ttySendHelp).
		Parse(args)
	if err != nil {
		return err
	}
	if noWait && maxWait > 0 {
		return fmt.Errorf("tty send: --no-wait and --max-wait are mutually exclusive")
	}
	if len(remaining) < 1 {
		return fmt.Errorf("tty send: requires <session-id> and <message>")
	}
	if len(remaining) < 2 {
		return fmt.Errorf("tty send: requires <session-id> and <message>")
	}
	sessionID := remaining[0]
	message := strings.Join(remaining[1:], " ")

	store, err := openStore()
	if err != nil {
		return err
	}
	home := store.Home()

	ttySess, err := agenttty.ResolveByTerminalID(home, sessionID)
	if err != nil {
		return err
	}

	if !ttySess.TCPReachable {
		return fmt.Errorf("terminal unreachable at %s", ttySess.Registry.ListenAddr)
	}

	provider, ok := agenttty.Get(ttySess.RunnerID)
	if !ok {
		return fmt.Errorf("unknown tty runner: %s", ttySess.RunnerID)
	}

	sess := agentsend.Session{
		Home:              home,
		Runner:            ttySess.RunnerID,
		TerminalSessionID: sessionID,
		ListenAddr:        ttySess.Registry.ListenAddr,
	}

	enqueuedAt := time.Now()
	id, err := agentsend.EnqueueWith(home, sess, message, agentsend.EnqueueOptions{NoSubmit: noSubmit})
	if err != nil {
		return err
	}
	fmt.Println(id)

	waitOpts := agentsend.WaitOptions{EnqueuedAt: enqueuedAt}
	switch {
	case noWait:
		waitOpts.Mode = agentsend.WaitNoWait
	case maxWait > 0:
		waitOpts.Mode = agentsend.WaitMaxWait
		waitOpts.MaxWait = maxWait
		waitOpts.StartDrainer = true
	default:
		waitOpts.Mode = agentsend.WaitDefault
		waitOpts.StartDrainer = true
	}

	if waitOpts.StartDrainer {
		agentsend.StartDrainer(home, sess, provider)
	}

	return agentsend.WaitForDelivery(home, sess, id, waitOpts)
}

func runTtySnapshot(args []string) error {
	remaining, err := flags.Help("-h,--help", ttySnapshotHelp).Parse(args)
	if err != nil {
		return err
	}
	if len(remaining) != 1 {
		return fmt.Errorf("snapshot: requires <session-id>")
	}
	sessionID := remaining[0]

	entry, _, err := resolveRegistryEntry(sessionID)
	if err != nil {
		return err
	}

	text, err := ttywatch.SnapshotText(entry.ListenAddr, sessionID)
	if err != nil {
		return err
	}
	if text != "" {
		fmt.Println(text)
	}
	return nil
}

func runTtyWatch(args []string) error {
	remaining, err := flags.Help("-h,--help", ttyWatchHelp).Parse(args)
	if err != nil {
		return err
	}
	if len(remaining) != 1 {
		return fmt.Errorf("watch: requires <session-id>")
	}
	sessionID := remaining[0]

	entry, _, err := resolveRegistryEntry(sessionID)
	if err != nil {
		return err
	}

	return ttywatch.StreamObserver(entry.ListenAddr, sessionID, os.Stdout)
}
