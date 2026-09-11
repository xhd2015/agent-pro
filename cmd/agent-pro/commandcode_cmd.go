package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/xhd2015/agent-pro/agent/commandcode"
	"github.com/xhd2015/less-gen/flags"
	"golang.org/x/term"
)

const commandCodeHelp = `
Usage: agent-pro commandcode <command> [ARGS]

Read Command Code account data over the public HTTP API. Every command is a
read-only GET; nothing shells out to the cmd CLI or reads its TTY output.

Commands:
  usage         credit usage, rate limits, and renewal (composite)
  whoami        authenticated account and org
  credits       credit balance and rate-limit windows
  subscription  current subscription
  summary       usage totals for the billing period
  namespaces    personal and team namespaces

Common options:
  --home <dir>    Command Code config dir (default: ~/.commandcode)
  --api-url <u>   API base URL (default: https://api.commandcode.ai)
  --json          print the raw API payload as JSON
  -h,--help       show help

Environment:
  COMMAND_CODE_API_KEY  API key, overriding the auth.json the cmd CLI writes
  COMMANDCODE_SANDBOX   when true, COMMANDCODE_API_URL overrides the base URL

Run agent-pro commandcode <command> --help for command-specific options.
`

const commandCodeUsageHelp = `
Usage: agent-pro commandcode usage [options]

Show credit usage, rate-limit windows, and renewal, laid out like the cmd
CLI's usage overlay. Fetches whoami, credits, subscription, and the
billing-period usage summary.

Options:
  --home <dir>     Command Code config dir (default: ~/.commandcode)
  --api-url <u>    API base URL (default: https://api.commandcode.ai)
  --org <id>       pin the org scope (default: whoami org, else personal)
  --since <time>   pin the summary period start, RFC3339 (default: the
                   subscription's current period start)
  --width <n>      bar width basis (default: terminal width, else 80)
  --no-color       disable ANSI color (also honors NO_COLOR)
  --json           print the merged payload as JSON
  --open           open the Studio usage page in a browser
  --watch          refresh until interrupted
  --interval <d>   refresh interval for --watch (default: 60s)
  -h,--help        show help

Exit status:
  0 when the account was fetched, even if some endpoints failed (each failure
  is a warning on stderr and its section is omitted).
  1 when the account could not be fetched at all, for example when the API is
  unreachable or the credential was rejected.
`

const commandCodeWhoamiHelp = `
Usage: agent-pro commandcode whoami [options]

Show the authenticated account, its org, and its spend limits.

Options:
  --home <dir>    Command Code config dir (default: ~/.commandcode)
  --api-url <u>   API base URL (default: https://api.commandcode.ai)
  --json          print the raw API payload as JSON
  -h,--help       show help
`

const commandCodeCreditsHelp = `
Usage: agent-pro commandcode credits [options]

Show the credit balance and the 5-hour / weekly rate-limit windows.

Options:
  --home <dir>    Command Code config dir (default: ~/.commandcode)
  --api-url <u>   API base URL (default: https://api.commandcode.ai)
  --json          print the raw API payload as JSON
  -h,--help       show help
`

const commandCodeSubscriptionHelp = `
Usage: agent-pro commandcode subscription [options]

Show the current subscription, its plan, and the billing period.

Options:
  --home <dir>    Command Code config dir (default: ~/.commandcode)
  --api-url <u>   API base URL (default: https://api.commandcode.ai)
  --json          print the raw API payload as JSON
  -h,--help       show help
`

const commandCodeSummaryHelp = `
Usage: agent-pro commandcode summary [options]

Show usage totals (requests, cost, tokens) for the billing period.

Options:
  --home <dir>    Command Code config dir (default: ~/.commandcode)
  --api-url <u>   API base URL (default: https://api.commandcode.ai)
  --since <time>  period start, RFC3339 (default: whole account history)
  --json          print the raw API payload as JSON
  -h,--help       show help
`

const commandCodeNamespacesHelp = `
Usage: agent-pro commandcode namespaces [options]

Show the personal namespace and any team namespaces.

Options:
  --home <dir>    Command Code config dir (default: ~/.commandcode)
  --api-url <u>   API base URL (default: https://api.commandcode.ai)
  --json          print the raw API payload as JSON
  -h,--help       show help
`

func handleCommandCode(args []string) error {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		fmt.Print(strings.TrimPrefix(commandCodeHelp, "\n"))
		return nil
	}

	switch args[0] {
	case "usage":
		return handleCommandCodeUsage(args[1:])
	case "whoami":
		return handleCommandCodeWhoami(args[1:])
	case "credits":
		return handleCommandCodeCredits(args[1:])
	case "subscription", "subscriptions":
		return handleCommandCodeSubscription(args[1:])
	case "summary":
		return handleCommandCodeSummary(args[1:])
	case "namespaces":
		return handleCommandCodeNamespaces(args[1:])
	default:
		return fmt.Errorf("unknown commandcode command: %s", args[0])
	}
}

// commandCodeClientFlags are the flags every commandcode subcommand shares.
type commandCodeClientFlags struct {
	home   string
	apiURL string
	json   *bool
}

func (f *commandCodeClientFlags) register(b *flags.Builder) *flags.Builder {
	return b.String("--home", &f.home).
		String("--api-url", &f.apiURL).
		Bool("--json", &f.json)
}

func (f *commandCodeClientFlags) client() (*commandcode.Client, error) {
	return commandcode.NewClient(f.home, f.apiURL)
}

func (f *commandCodeClientFlags) jsonRequested() bool {
	return f.json != nil && *f.json
}

func handleCommandCodeWhoami(args []string) error {
	var clientFlags commandCodeClientFlags
	remaining, err := clientFlags.register(flags.New()).
		Help("-h,--help", commandCodeWhoamiHelp).
		Parse(args)
	if err != nil {
		return err
	}
	if len(remaining) > 0 {
		return fmt.Errorf("unexpected arguments: %s", strings.Join(remaining, " "))
	}

	client, err := clientFlags.client()
	if err != nil {
		return err
	}
	whoami, err := client.Whoami(context.Background(), "")
	if err != nil {
		return err
	}
	if clientFlags.jsonRequested() {
		return printCommandCodeJSON(whoami)
	}
	fmt.Print(commandcode.FormatWhoami(whoami))
	field := "credential"
	fmt.Printf("%-18s %s\n", field, client.Auth.Source)
	return nil
}

func handleCommandCodeCredits(args []string) error {
	var clientFlags commandCodeClientFlags
	remaining, err := clientFlags.register(flags.New()).
		Help("-h,--help", commandCodeCreditsHelp).
		Parse(args)
	if err != nil {
		return err
	}
	if len(remaining) > 0 {
		return fmt.Errorf("unexpected arguments: %s", strings.Join(remaining, " "))
	}

	client, err := clientFlags.client()
	if err != nil {
		return err
	}
	credits, err := client.Credits(context.Background(), "")
	if err != nil {
		return err
	}
	if clientFlags.jsonRequested() {
		return printCommandCodeJSON(credits)
	}
	fmt.Print(commandcode.FormatCreditsView(credits))
	return nil
}

func handleCommandCodeSubscription(args []string) error {
	var clientFlags commandCodeClientFlags
	remaining, err := clientFlags.register(flags.New()).
		Help("-h,--help", commandCodeSubscriptionHelp).
		Parse(args)
	if err != nil {
		return err
	}
	if len(remaining) > 0 {
		return fmt.Errorf("unexpected arguments: %s", strings.Join(remaining, " "))
	}

	client, err := clientFlags.client()
	if err != nil {
		return err
	}
	sub, err := client.Subscription(context.Background(), "")
	if err != nil {
		return err
	}
	if clientFlags.jsonRequested() {
		return printCommandCodeJSON(sub)
	}
	fmt.Print(commandcode.FormatSubscriptionView(sub))
	return nil
}

func handleCommandCodeSummary(args []string) error {
	var clientFlags commandCodeClientFlags
	var since string
	remaining, err := clientFlags.register(flags.New()).
		String("--since", &since).
		Help("-h,--help", commandCodeSummaryHelp).
		Parse(args)
	if err != nil {
		return err
	}
	if len(remaining) > 0 {
		return fmt.Errorf("unexpected arguments: %s", strings.Join(remaining, " "))
	}

	client, err := clientFlags.client()
	if err != nil {
		return err
	}
	summary, err := client.Summary(context.Background(), "", since)
	if err != nil {
		return err
	}
	if clientFlags.jsonRequested() {
		return printCommandCodeJSON(summary)
	}
	fmt.Print(commandcode.FormatSummaryView(summary))
	return nil
}

func handleCommandCodeNamespaces(args []string) error {
	var clientFlags commandCodeClientFlags
	remaining, err := clientFlags.register(flags.New()).
		Help("-h,--help", commandCodeNamespacesHelp).
		Parse(args)
	if err != nil {
		return err
	}
	if len(remaining) > 0 {
		return fmt.Errorf("unexpected arguments: %s", strings.Join(remaining, " "))
	}

	client, err := clientFlags.client()
	if err != nil {
		return err
	}
	namespaces, err := client.Namespaces(context.Background())
	if err != nil {
		return err
	}
	if clientFlags.jsonRequested() {
		return printCommandCodeJSON(namespaces)
	}
	fmt.Print(commandcode.FormatNamespacesView(namespaces))
	return nil
}

func handleCommandCodeUsage(args []string) error {
	var home, apiURL, org, since string
	var width int
	var noColor, jsonFlag, openFlag, watchFlag *bool
	var interval time.Duration
	remaining, err := flags.New().
		String("--home", &home).
		String("--api-url", &apiURL).
		String("--org", &org).
		String("--since", &since).
		Int("--width", &width).
		Bool("--no-color", &noColor).
		Bool("--json", &jsonFlag).
		Bool("--open", &openFlag).
		Bool("--watch", &watchFlag).
		Duration("--interval", &interval).
		Help("-h,--help", commandCodeUsageHelp).
		Parse(args)
	if err != nil {
		return err
	}
	if len(remaining) > 0 {
		return fmt.Errorf("unexpected arguments: %s", strings.Join(remaining, " "))
	}

	isJSON := jsonFlag != nil && *jsonFlag
	isWatch := watchFlag != nil && *watchFlag
	if isJSON && isWatch {
		return fmt.Errorf("--json and --watch are mutually exclusive")
	}

	client, err := commandcode.NewClient(home, apiURL)
	if err != nil {
		return err
	}

	opts := commandcode.FormatOptions{
		Width: width,
		Color: commandCodeUseColor(noColor),
	}
	if opts.Width <= 0 {
		opts.Width = commandCodeTerminalWidth()
	}

	render := func(ctx context.Context) error {
		data, err := client.FetchUsageWithOptions(ctx, commandcode.UsageOptions{OrgID: org, Since: since})
		if err != nil {
			return err
		}
		// No identity means every scoped endpoint went unqueried, so there is
		// nothing to render: report the fetch as failed rather than printing an
		// empty overlay.
		if data.Whoami == nil {
			return errors.New(commandCodeUsageFailure(data))
		}
		for _, warning := range data.Errors {
			fmt.Fprintf(os.Stderr, "warning: %s\n", warning)
		}

		if openFlag != nil && *openFlag {
			view := commandcode.ProjectUsageView(data, time.Now())
			if view.UsageURL == "" {
				return fmt.Errorf("no Studio usage URL for this account")
			}
			if err := openInBrowser(view.UsageURL); err != nil {
				return err
			}
			fmt.Printf("opened %s\n", view.UsageURL)
		}

		if isJSON {
			out, err := commandcode.FormatUsageJSON(data)
			if err != nil {
				return fmt.Errorf("format usage json: %w", err)
			}
			fmt.Println(string(out))
		} else {
			fmt.Print(commandcode.FormatUsage(commandcode.ProjectUsageView(data, time.Now()), opts))
		}
		return nil
	}

	if !isWatch {
		return render(context.Background())
	}

	if interval <= 0 {
		interval = time.Minute
	}
	return commandCodeWatchUsage(render, interval)
}

// commandCodeUsageFailure explains a fetch that produced no identity, which
// leaves every scoped endpoint unqueried.
func commandCodeUsageFailure(data *commandcode.UsageData) string {
	if len(data.Errors) > 0 {
		return data.Errors[0]
	}
	return "no account data returned"
}

// commandCodeWatchUsage refreshes the usage view until interrupted, clearing
// the screen between refreshes when stdout is a terminal. Failures are
// reported and retried so one blip does not end the watch.
func commandCodeWatchUsage(render func(context.Context) error, interval time.Duration) error {
	clearScreen := term.IsTerminal(int(os.Stdout.Fd()))
	for {
		if clearScreen {
			fmt.Print("\x1b[2J\x1b[H")
		}
		if err := render(context.Background()); err != nil {
			fmt.Fprintf(os.Stderr, "warning: %s\n", err)
		}
		fmt.Printf("\nrefreshing in %s (Ctrl-C to stop)\n", interval)
		time.Sleep(interval)
	}
}

func printCommandCodeJSON(value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("format commandcode json: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

func commandCodeUseColor(noColor *bool) bool {
	if noColor != nil && *noColor {
		return false
	}
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	return term.IsTerminal(int(os.Stdout.Fd()))
}

func commandCodeTerminalWidth() int {
	if width, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && width > 0 {
		return width
	}
	return commandcode.DefaultTerminalWidth
}

func openInBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("open %s: %w", url, err)
	}
	return nil
}
