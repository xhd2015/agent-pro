// Package cli implements read-only DeepSeek Harness session commands.
package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/xhd2015/agent-pro/agent/dsh/sessions"
	flags "github.com/xhd2015/less-flags"
)

const Help = `Usage: agent-pro dsh <command> [ARGS]

Commands:
  sessions          list stored DSH sessions
  session           inspect stored DSH sessions

Run agent-pro dsh <command> --help for options.
`
const sessionHelp = `Usage: agent-pro dsh session <command> [ARGS]

Commands:
  list              alias for: agent-pro dsh sessions
  info ID           show durable session metadata
  status ID         show activity, liveness, and durable outcome
  log ID            print durable transcript
  messages ID       print user and assistant messages
  prompts [ID]      print human prompts (or recent sessions)

Run agent-pro dsh session <command> --help for options.
`

// Run dispatches arguments after "agent-pro dsh". Diagnostics never enter stdout.
func Run(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		_, err := io.WriteString(stdout, Help)
		return err
	}
	action := args[0]
	args = args[1:]
	if action == "session" {
		if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
			_, err := io.WriteString(stdout, sessionHelp)
			return err
		}
		action = args[0]
		args = args[1:]
		if action == "list" {
			action = "sessions"
		}
	} else if action != "sessions" {
		return fmt.Errorf("unknown dsh command: %s", action)
	}
	switch action {
	case "sessions", "info", "status", "log", "messages", "prompts":
	default:
		return fmt.Errorf("unknown dsh session command: %s", action)
	}
	var opts sessions.Options
	var jsonMode, here, help bool
	var dirs []string
	limit, tail := 20, 0
	parser := flags.String("--home", &opts.Home).String("--sessions-dir", &opts.SessionsDir).Bool("--json", &jsonMode).Bool("-h,--help", &help)
	if action == "sessions" || action == "prompts" {
		parser.Bool("--here", &here).StringSlice("--dir", &dirs).Int("--limit", &limit)
	}
	if action == "log" || action == "messages" || action == "prompts" {
		parser.Int("--tail", &tail)
	}
	var recorded flags.Flags
	rest, err := parser.CollectParsedFlags(&recorded).Parse(args)
	if err != nil {
		return err
	}
	if help {
		_, err = io.WriteString(stdout, leafHelp(action))
		return err
	}
	if limit < 1 {
		return fmt.Errorf("--limit must be positive")
	}
	if tail < 0 {
		return fmt.Errorf("--tail must be non-negative")
	}
	if (action == "sessions" && len(rest) != 0) || (action == "prompts" && len(rest) > 1) || (action != "sessions" && action != "prompts" && len(rest) != 1) {
		return fmt.Errorf("invalid arguments; see agent-pro dsh %s --help", commandName(action))
	}
	if action == "prompts" && len(rest) == 1 {
		for _, arg := range recorded.Reconstruct() {
			if arg == "--here" || arg == "--dir" || arg == "--limit" || strings.HasPrefix(arg, "--dir=") || strings.HasPrefix(arg, "--limit=") {
				return fmt.Errorf("--here, --dir and --limit apply only to prompts without a session id")
			}
		}
	}
	if action == "sessions" || action == "prompts" {
		if _, err := filter(nil, here, dirs, limit); err != nil {
			return err
		}
	}
	store, err := sessions.NewStore(opts)
	if err != nil {
		return err
	}
	if action == "sessions" || (action == "prompts" && len(rest) == 0) {
		rows, warnings, err := store.List()
		if err != nil {
			return err
		}
		for _, warning := range warnings {
			if _, err = fmt.Fprintf(stderr, "warning: %s\n", warning); err != nil {
				return err
			}
		}
		rows, err = filter(rows, here, dirs, limit)
		if err != nil {
			return err
		}
		if action == "sessions" {
			if jsonMode {
				return writeJSON(stdout, rows)
			}
			return printList(stdout, rows)
		}
		all := []Message{}
		for _, row := range rows {
			inspection, readErr := store.Inspect(row.ID)
			if readErr != nil {
				if _, err = fmt.Fprintf(stderr, "warning: %s\n", readErr); err != nil {
					return err
				}
				continue
			}
			items, messageErr := messages(inspection, true)
			if messageErr != nil {
				return messageErr
			}
			all = append(all, items...)
		}
		sort.SliceStable(all, func(i, j int) bool {
			if all[i].Time != all[j].Time {
				return all[i].Time < all[j].Time
			}
			if all[i].SessionID != all[j].SessionID {
				return all[i].SessionID < all[j].SessionID
			}
			return all[i].Seq < all[j].Seq
		})
		return printMessages(stdout, last(all, tail), jsonMode)
	}
	inspection, err := store.Inspect(rest[0])
	if err != nil {
		return err
	}
	switch action {
	case "info":
		if jsonMode {
			return writeJSON(stdout, inspection.Summary)
		}
		return printInfo(stdout, inspection.Summary, false)
	case "status":
		status := sessionStatus(inspection)
		if jsonMode {
			return writeJSON(stdout, status)
		}
		return printStatus(stdout, status)
	case "messages", "prompts":
		items, messageErr := messages(inspection, action == "prompts")
		if messageErr != nil {
			return messageErr
		}
		return printMessages(stdout, last(items, tail), jsonMode)
	case "log":
		return printLog(stdout, inspection, tail, jsonMode)
	}
	return nil
}
func commandName(action string) string {
	if action == "sessions" {
		return action
	}
	return "session " + action
}
func leafHelp(action string) string {
	positional := " <session-id>"
	if action == "sessions" {
		positional = ""
	}
	if action == "prompts" {
		positional = " [session-id]"
	}
	text := "Usage: agent-pro dsh " + commandName(action) + positional + " [OPTIONS]\n\nOptions:\n  --home DIR          DSH home ($DSH_HOME or ~/.dsh)\n  --sessions-dir DIR  explicit session root; overrides --home\n  --json              write JSON (log: durable event NDJSON)\n"
	if action == "sessions" || action == "prompts" {
		text += "  --here              filter by current working directory\n  --dir DIR           filter by working directory (repeatable; OR with --here)\n  --limit N           recent session limit (default 20)\n"
	}
	if action == "log" || action == "messages" || action == "prompts" {
		text += "  --tail N            last N entries (0 = all)\n"
	}
	text += "  -h, --help          show help\n"
	if action == "status" {
		text += "\nReports live activity (thinking / responding / tool_calling / error / idle),\nprocess liveness, and the durable-log outcome.\n"
	}
	return text
}
func filter(rows []sessions.Summary, here bool, dirs []string, limit int) ([]sessions.Summary, error) {
	places := map[string]bool{}
	if here {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		dirs = append(dirs, cwd)
	}
	for _, dir := range dirs {
		abs, err := filepath.Abs(dir)
		if err != nil {
			return nil, err
		}
		st, err := os.Stat(abs)
		if err != nil {
			return nil, fmt.Errorf("--dir %q: %w", dir, err)
		}
		if !st.IsDir() {
			return nil, fmt.Errorf("--dir %q is not a directory", dir)
		}
		places[filepath.Clean(abs)] = true
	}
	result := []sessions.Summary{}
	for _, row := range rows {
		if len(places) == 0 || places[filepath.Clean(row.Cwd)] {
			result = append(result, row)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].UpdatedAt != result[j].UpdatedAt {
			return result[i].UpdatedAt > result[j].UpdatedAt
		}
		return result[i].ID < result[j].ID
	})
	if len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}
func writeJSON(w io.Writer, value any) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}
func last[T any](items []T, n int) []T {
	if n > 0 && len(items) > n {
		return items[len(items)-n:]
	}
	return items
}
func clean(text string) string {
	return strings.Map(func(r rune) rune {
		if r < 32 || r == 127 {
			return ' '
		}
		return r
	}, text)
}
