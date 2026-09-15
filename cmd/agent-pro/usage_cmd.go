package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/xhd2015/agent-pro/agent/usage"
	"github.com/xhd2015/agent-pro/agent/usage/view"
	"github.com/xhd2015/agent-pro/pkgs/cronspec"
	"github.com/xhd2015/less-gen/flags"
)

const usageHelp = `
Usage: agent-pro usage <command> [ARGS]

Commands:
  collect   collect each provider's usage into ~/.agent-pro/usages (snapshots)
  list      list stored usage snapshots
  view      serve a read-only dashboard of the stored snapshots

Run agent-pro usage <command> --help for command-specific options.
`

const usageCollectHelp = `
Usage: agent-pro usage collect [options]

Read account usage from every provider over its HTTP API and append one
snapshot per provider under

  $AGENT_PRO_HOME/usages/<provider>/<YYYY-MM-DD>/<HH-MM-SS>-snapshot.jsonl
  (default root: ~/.agent-pro/usages)

Each snapshot holds the provider usage numbers plus the count of that
provider's sessions on disk (total, oldest, newest). A provider whose usage
fetch fails still gets a snapshot: usage.ok is false with the error, while the
session counts are unaffected because they only read local files.

Providers (always all three):

  grok          grok billing API   ($GROK_HOME/auth.json)
  codex         Codex usage API    ($CODEX_HOME/auth.json)
  commandcode   Command Code API   (~/.commandcode/auth.json)

Options:
  --grok-home <dir>         grok home for auth and sessions (default: $GROK_HOME or ~/.grok)
  --codex-home <dir>        codex home for auth and sessions (default: $CODEX_HOME or ~/.codex)
  --commandcode-home <dir>  Command Code home (default: ~/.commandcode)
  --json                    print one JSON line per provider (no ANSI, no extra text)
  --cron <spec>             repeat on a schedule: 5-field cron (min hour dom mon dow),
                            @hourly|@daily|@midnight|@weekly|@monthly, or @every <dur>
  --max-runs <n>            stop after n cycles with --cron (0 = unlimited)
  --timeout <dur>           per-provider usage fetch timeout (default 30s)
  -h,--help                 show help

Exit status:
  0 when at least one provider's usage was fetched (a failing provider is a
    warning on stderr and its snapshot records usage.ok = false)
  1 when every provider failed, the --cron spec is invalid, or the store
    could not be written

Environment:
  AGENT_PRO_HOME              store root (default ~/.agent-pro)
  GROK_HOME / CODEX_HOME      provider homes when the flags are omitted
  COMMANDCODE_SANDBOX         when true, COMMANDCODE_API_URL overrides the API base
  COMMANDCODE_API_URL         Command Code API base (sandbox/tests)
`

const usageListHelp = `
Usage: agent-pro usage list [options]

List stored usage snapshots, newest first, across every provider.

Options:
  --last <n>      show at most n snapshots (0 = all; default 10)
  --json          print one stored record per line (no ANSI)
  -h,--help       show help
`

const usageViewHelp = `
Usage: agent-pro usage view [options]

Serve a read-only local dashboard of the stored usage snapshots: one card per
provider, a chart per metric, and the newest snapshots. The store is only read,
nothing is ever written back.

  $AGENT_PRO_HOME/usages/<provider>/<YYYY-MM-DD>/<HH-MM-SS>-snapshot.jsonl
  (default root: ~/.agent-pro/usages)

Charts:
  primary_percent   grok/codex used_percent, commandcode usage_percent
  sessions_total    sessions found on disk per provider (cumulative)
  ...every other numeric key in the snapshots, from the metric menu

Options:
  --port <n>        preferred listen port; 0 binds any free port (default 8080).
                    A busy port is skipped for the next 99 ports
  --open            open the dashboard in a browser
  --no-open         never open a browser (default when stdout is not a terminal)
  --since <dur>     default chart range: 90m, 24h, 7d, 2w (default 7d)
  --static-dir <d>  serve dashboard.html from a directory instead of the embed
  --json            print one JSON line for the bound URL, then serve
  -h,--help         show help

Exit status:
  0 on a clean shutdown
  1 on an invalid flag, an unreadable store root, or no free port; a snapshot
    file that cannot be parsed is a warning, not a failure

Environment:
  AGENT_PRO_HOME    store root (default ~/.agent-pro)
`

func handleUsage(args []string) error {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		fmt.Print(strings.TrimPrefix(usageHelp, "\n"))
		return nil
	}
	switch args[0] {
	case "collect":
		return handleUsageCollect(args[1:])
	case "list":
		return handleUsageList(args[1:])
	case "view":
		return handleUsageView(args[1:])
	default:
		return fmt.Errorf("unknown usage command: %s", args[0])
	}
}

func handleUsageCollect(args []string) error {
	var (
		jsonFlag            *bool
		cronFlag            *string
		maxRunsFlag         *int
		timeoutFlag         *time.Duration
		grokHomeFlag        *string
		codexHomeFlag       *string
		commandCodeHomeFlag *string
	)
	remaining, err := flags.Bool("--json", &jsonFlag).
		String("--cron", &cronFlag).
		Int("--max-runs", &maxRunsFlag).
		Duration("--timeout", &timeoutFlag).
		String("--grok-home", &grokHomeFlag).
		String("--codex-home", &codexHomeFlag).
		String("--commandcode-home", &commandCodeHomeFlag).
		Help("-h,--help", usageCollectHelp).
		Parse(args)
	if err != nil {
		return err
	}
	if len(remaining) > 0 {
		return fmt.Errorf("unexpected arguments: %s", strings.Join(remaining, " "))
	}

	asJSON := jsonFlag != nil && *jsonFlag
	maxRuns := 0
	if maxRunsFlag != nil {
		maxRuns = *maxRunsFlag
		if maxRuns < 0 {
			return fmt.Errorf("invalid --max-runs %d: must not be negative", maxRuns)
		}
	}
	timeout := usage.DefaultTimeout
	if timeoutFlag != nil {
		if *timeoutFlag <= 0 {
			return fmt.Errorf("invalid --timeout %s: must be greater than 0", timeoutFlag)
		}
		timeout = *timeoutFlag
	}

	collectOpts := usage.CollectOptions{
		Timeout:         timeout,
		GrokHome:        derefString(grokHomeFlag),
		CodexHome:       derefString(codexHomeFlag),
		CommandCodeHome: derefString(commandCodeHomeFlag),
		FixtureDir:      strings.TrimSpace(os.Getenv(usage.EnvFixtureDir)),
	}
	store := usage.NewStore(filepath.Join(resolveAgentProHome(), "usages"))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// runCycle reports how many providers produced usage this cycle.
	runCycle := func(cycleCtx context.Context) (int, error) {
		return writeUsageSnapshots(store, usage.Collect(cycleCtx, collectOpts), asJSON)
	}

	cronSpec := derefString(cronFlag)
	if cronSpec == "" {
		collected, err := runCycle(ctx)
		if err != nil {
			return err
		}
		if collected == 0 {
			return errors.New("no usage collected: every provider failed")
		}
		return nil
	}

	schedule, err := cronspec.Parse(cronSpec)
	if err != nil {
		return fmt.Errorf("invalid --cron spec %q: %w", cronSpec, err)
	}
	return cronspec.Run(ctx, schedule, cronspec.RunOptions{
		MaxRuns: maxRuns,
		Logf: func(format string, args ...any) {
			fmt.Fprintf(os.Stderr, "usage collect: "+format+"\n", args...)
		},
	}, func(cycleCtx context.Context) error {
		_, err := runCycle(cycleCtx)
		return err
	})
}

// writeUsageSnapshots stores one record per provider and reports them.
// It returns the number of providers whose usage was fetched.
func writeUsageSnapshots(store *usage.Store, records []usage.Record, asJSON bool) (int, error) {
	collected := 0
	lines := make([][]byte, 0, len(records))
	for _, rec := range records {
		if rec.Usage.OK {
			collected++
		} else {
			fmt.Fprintf(os.Stderr, "warning: %s: %s\n", rec.Provider, rec.Usage.Error)
		}
		line, err := rec.MarshalJSONLine()
		if err != nil {
			return 0, fmt.Errorf("encode %s snapshot: %w", rec.Provider, err)
		}
		lines = append(lines, line)
		if _, err := store.Append(rec); err != nil {
			return 0, err
		}
	}

	if asJSON {
		for _, line := range lines {
			fmt.Println(string(line))
		}
		return collected, nil
	}

	printUsageTable(records)
	fmt.Printf("\n%d snapshots appended under %s\n", len(records), shortenHome(store.Root, homeDir()))
	return collected, nil
}

func printUsageTable(records []usage.Record) {
	fmt.Printf("%-12s %-34s %-9s %s\n", "provider", "usage", "sessions", "took")
	for _, rec := range records {
		fmt.Printf("%-12s %-34s %-9s %s\n",
			rec.Provider,
			usageHeadline(rec),
			strconv.Itoa(rec.Sessions.Total),
			formatDurationMS(rec.Usage.DurationMS),
		)
		if detail := rec.Usage.Display["detail"]; detail != "" {
			fmt.Printf("%-12s %s\n", "", detail)
		}
	}
}

// usageHeadline is the usage column: the provider summary, flagged when the
// provider's session tree could not be read.
func usageHeadline(rec usage.Record) string {
	if rec.Sessions.Error != "" {
		return rec.Headline() + " (sessions unreadable)"
	}
	return rec.Headline()
}

func formatDurationMS(ms int64) string {
	if ms < 1000 {
		return fmt.Sprintf("%dms", ms)
	}
	return fmt.Sprintf("%.1fs", float64(ms)/1000)
}

func handleUsageList(args []string) error {
	var (
		lastFlag *int
		jsonFlag *bool
	)
	remaining, err := flags.Int("--last", &lastFlag).
		Bool("--json", &jsonFlag).
		Help("-h,--help", usageListHelp).
		Parse(args)
	if err != nil {
		return err
	}
	if len(remaining) > 0 {
		return fmt.Errorf("unexpected arguments: %s", strings.Join(remaining, " "))
	}

	last := 10
	if lastFlag != nil {
		last = *lastFlag
		if last < 0 {
			return fmt.Errorf("invalid --last %d: must not be negative", last)
		}
	}

	store := usage.NewStore(filepath.Join(resolveAgentProHome(), "usages"))
	stored, err := store.List(last)
	if err != nil {
		return err
	}
	if len(stored) == 0 {
		fmt.Printf("no usage snapshots under %s\n", shortenHome(store.Root, homeDir()))
		return nil
	}

	if jsonFlag != nil && *jsonFlag {
		for _, item := range stored {
			line, err := item.Record.MarshalJSONLine()
			if err != nil {
				return fmt.Errorf("encode stored snapshot: %w", err)
			}
			fmt.Println(string(line))
		}
		return nil
	}

	fmt.Printf("%-24s %-12s %-34s %-9s %s\n", "ts", "provider", "usage", "sessions", "file")
	for _, item := range stored {
		fmt.Printf("%-24s %-12s %-34s %-9s %s\n",
			item.Record.TS.UTC().Format(time.RFC3339),
			item.Record.Provider,
			usageHeadline(item.Record),
			strconv.Itoa(item.Record.Sessions.Total),
			shortenHome(item.Path, homeDir()),
		)
	}
	return nil
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func handleUsageView(args []string) error {
	var (
		portFlag   *int
		openFlag   *bool
		noOpenFlag *bool
		sinceFlag  *string
		staticFlag *string
		jsonFlag   *bool
	)
	remaining, err := flags.Int("--port", &portFlag).
		Bool("--open", &openFlag).
		Bool("--no-open", &noOpenFlag).
		String("--since", &sinceFlag).
		String("--static-dir", &staticFlag).
		Bool("--json", &jsonFlag).
		Help("-h,--help", usageViewHelp).
		Parse(args)
	if err != nil {
		return err
	}
	if len(remaining) > 0 {
		return fmt.Errorf("unexpected arguments: %s", strings.Join(remaining, " "))
	}

	opts := view.Options{
		Port:      view.DefaultPort,
		Since:     view.DefaultSince,
		StaticDir: derefString(staticFlag),
		JSON:      jsonFlag != nil && *jsonFlag,
		Stdout:    os.Stdout,
		Stderr:    os.Stderr,
	}
	if portFlag != nil {
		opts.Port = *portFlag
		opts.PortExplicit = true
		if opts.Port < 0 || opts.Port > 65535 {
			return fmt.Errorf("invalid --port %d: must be between 0 and 65535", opts.Port)
		}
	}
	if sinceFlag != nil {
		since, err := view.ParseRange(*sinceFlag)
		if err != nil {
			return fmt.Errorf("invalid --since %q: %w", *sinceFlag, err)
		}
		opts.Since = since
	}
	if opts.StaticDir != "" {
		if _, err := os.Stat(filepath.Join(opts.StaticDir, "dashboard.html")); err != nil {
			return fmt.Errorf("invalid --static-dir %s: %w", opts.StaticDir, err)
		}
	}
	switch {
	case openFlag != nil && *openFlag && noOpenFlag != nil && *noOpenFlag:
		return errors.New("--open and --no-open are mutually exclusive")
	case openFlag != nil && *openFlag:
		opts.Open = true
	case noOpenFlag != nil && *noOpenFlag:
		opts.Open = false
	default:
		opts.Open = stdoutIsTerminal()
	}

	store := usage.NewStore(filepath.Join(resolveAgentProHome(), "usages"))
	return view.Serve(context.Background(), store, opts)
}

// stdoutIsTerminal reports whether the dashboard URL would be seen by a person,
// which is when opening a browser is helpful.
func stdoutIsTerminal() bool {
	info, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}
