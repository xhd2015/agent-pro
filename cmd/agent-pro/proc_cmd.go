package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	groksessions "github.com/xhd2015/agent-pro/agent/grok/sessions"
	"github.com/xhd2015/agent-pro/pkgs/procresolve"
	"github.com/xhd2015/less-gen/flags"
)

const procHelp = `
Usage: agent-pro proc <command> [ARGS]

Commands:
  resolve <pid>   resolve agent session from a process id

Run agent-pro proc <command> --help for command-specific options.
`

const procResolveHelp = `
Usage: agent-pro proc resolve <pid> [options]

Resolve an agent session (grok/codex) from a process id by walking the
process tree and inspecting open files (not cmdline flags).

Options:
  --json          print Result as JSON (no tree glyphs)
  --ascii-tree    use ASCII connectors (+-- / ` + "`--" + ` / |) in human tree output
  --no-enrich     do not look up grok session title/model
  -h, --help      show help

Environment:
  AGENT_PRO_PROCRESOLVE_TEST_SNAPSHOT  optional JSON fixture for ListProcs/Lsof
                                       (tests only; when set, skips live ps/lsof)
`

// testSnapshot is the JSON body of AGENT_PRO_PROCRESOLVE_TEST_SNAPSHOT.
type testSnapshot struct {
	Procs     []testSnapProc       `json:"procs"`
	OpenFiles map[string][]string  `json:"open_files"`
	GrokHome  string               `json:"grok_home,omitempty"`
	CodexHome string               `json:"codex_home,omitempty"`
}

type testSnapProc struct {
	PID  int    `json:"pid"`
	PPID int    `json:"ppid"`
	Cmd  string `json:"cmd"`
}

func handleProc(args []string) error {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		fmt.Print(strings.TrimPrefix(procHelp, "\n"))
		// Also surface resolve flags on proc --help so either form documents --json.
		fmt.Print(strings.TrimPrefix(procResolveHelp, "\n"))
		return nil
	}
	switch args[0] {
	case "resolve":
		return handleProcResolve(args[1:])
	default:
		return fmt.Errorf("unknown proc command: %s", args[0])
	}
}

func handleProcResolve(args []string) error {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		fmt.Print(strings.TrimPrefix(procResolveHelp, "\n"))
		return nil
	}

	var jsonOut bool
	var asciiTree bool
	var noEnrich bool

	remaining, err := flags.New().
		Bool("--json", &jsonOut).
		Bool("--ascii-tree", &asciiTree).
		Bool("--no-enrich", &noEnrich).
		Help("-h,--help", strings.TrimPrefix(procResolveHelp, "\n")).
		Parse(args)
	if err != nil {
		return err
	}
	if len(remaining) == 0 {
		return fmt.Errorf("proc resolve: missing <pid>\n\n%s", strings.TrimPrefix(procResolveHelp, "\n"))
	}
	if len(remaining) > 1 {
		return fmt.Errorf("proc resolve: unexpected args %v", remaining[1:])
	}
	pid, err := strconv.Atoi(remaining[0])
	if err != nil {
		return fmt.Errorf("proc resolve: invalid pid %q", remaining[0])
	}

	opts, err := buildProcResolveOptions(noEnrich)
	if err != nil {
		return err
	}

	result, err := procresolve.ResolveFromPID(pid, opts)
	if err != nil {
		return err
	}

	if jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(result); err != nil {
			return err
		}
		// Encode already writes trailing newline.
		return nil
	}

	// Human mode: summary + tree.
	fmt.Printf("Kind: %s\n", result.Kind)
	if result.SessionID != "" {
		fmt.Printf("SessionID: %s\n", result.SessionID)
	}
	if result.Source != "" {
		fmt.Printf("Source: %s\n", result.Source)
	}
	if result.Confidence != "" {
		fmt.Printf("Confidence: %s\n", result.Confidence)
	}
	if result.RunnerPID != 0 {
		fmt.Printf("RunnerPID: %d\n", result.RunnerPID)
	}
	if result.GrokTitle != "" {
		fmt.Printf("GrokTitle: %s\n", result.GrokTitle)
	}
	if result.GrokModel != "" {
		fmt.Printf("GrokModel: %s\n", result.GrokModel)
	}
	if len(result.Tree) > 0 {
		fmt.Println("Tree:")
		fmt.Print(procresolve.FormatTree(result.Tree, procresolve.TreeFormatOptions{ASCII: asciiTree}))
	}
	return nil
}

func buildProcResolveOptions(noEnrich bool) (procresolve.Options, error) {
	opts := procresolve.Options{
		EnrichInfo: !noEnrich,
	}

	if raw := os.Getenv("AGENT_PRO_PROCRESOLVE_TEST_SNAPSHOT"); raw != "" {
		var snap testSnapshot
		if err := json.Unmarshal([]byte(raw), &snap); err != nil {
			return opts, fmt.Errorf("AGENT_PRO_PROCRESOLVE_TEST_SNAPSHOT: %w", err)
		}
		procs := make([]procresolve.Proc, 0, len(snap.Procs))
		for _, p := range snap.Procs {
			procs = append(procs, procresolve.Proc{PID: p.PID, PPID: p.PPID, Cmd: p.Cmd})
		}
		files := map[int][]string{}
		for k, v := range snap.OpenFiles {
			pid, err := strconv.Atoi(k)
			if err != nil {
				continue
			}
			files[pid] = v
		}
		snapProcs := procs
		opts.ListProcs = func() []procresolve.Proc { return snapProcs }
		opts.Lsof = func(pid int) []string { return files[pid] }
		if snap.GrokHome != "" {
			opts.GrokHome = snap.GrokHome
		}
		if snap.CodexHome != "" {
			opts.CodexHome = snap.CodexHome
		}
	} else {
		opts.ListProcs = procresolve.ListLiveProcs
		opts.Lsof = procresolve.LiveLsof
		if opts.GrokHome == "" {
			// Default home for enrich path; sessions.Info tolerates empty via Find.
			if home, err := os.UserHomeDir(); err == nil {
				opts.GrokHome = home
			}
		}
	}

	if opts.EnrichInfo {
		opts.LookupGrokInfo = func(home, sessionID string) (string, string, error) {
			info, err := groksessions.Info(home, sessionID)
			if err != nil {
				return "", "", err
			}
			return info.Title, info.CurrentModelID, nil
		}
	}
	return opts, nil
}
