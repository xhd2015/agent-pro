package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	lessflags "github.com/xhd2015/less-flags"
	"github.com/slack-go/slack"
	"github.com/xhd2015/agent-pro/pkgs/slackutil"
)

const channelsHelpText = `slack-msg channels: list or search workspace channels.

Usage:
  slack-msg channels <command> [options]

Commands:
  list    List visible channels
  search  Search channels by name

Options:
  -h, --help  Show help
`

const channelsListHelpText = `slack-msg channels list: list workspace channels.

Usage:
  slack-msg channels list [options]

Options:
  --token TOKEN   Bot token (env: SLACK_BOT_TOKEN)
  --config PATH   JSON config file (env: SLACK_CONFIG)
  --types TYPES   Channel types (default: public,private)
  --limit N       Max channels to print
  --json          Structured JSON output
  -h, --help      Show help
`

const channelsSearchHelpText = `slack-msg channels search: search workspace channels by name.

Usage:
  slack-msg channels search [options] QUERY

Options:
  --token TOKEN   Bot token (env: SLACK_BOT_TOKEN)
  --config PATH   JSON config file (env: SLACK_CONFIG)
  --types TYPES   Channel types (default: public,private)
  --limit N       Max channels to print
  --exact         Match channel name exactly
  --prefix        Match channel name by prefix
  --json          Structured JSON output
  -h, --help      Show help
`

type channelRow struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	IsPrivate  bool   `json:"is_private"`
	IsMember   bool   `json:"is_member"`
	IsArchived bool   `json:"is_archived"`
}

type channelsDoc struct {
	Channels []channelRow `json:"channels"`
}

func runChannels(args []string) error {
	if len(args) == 0 {
		fmt.Print(channelsHelpText)
		return nil
	}

	switch args[0] {
	case "-h", "--help":
		fmt.Print(channelsHelpText)
		return nil
	case "list":
		return runChannelsList(args[1:])
	case "search":
		return runChannelsSearch(args[1:])
	default:
		if strings.HasPrefix(args[0], "-") {
			return fmt.Errorf("unknown option: %s", args[0])
		}
		return fmt.Errorf("unknown command: %s", args[0])
	}
}

func runChannelsList(args []string) error {
	var (
		tokenFlag  *string
		configFlag *string
		typesFlag  *string
		limitFlag  *int
		jsonFlag   bool
	)

	remain, err := lessflags.String("--token", &tokenFlag).
		String("--config", &configFlag).
		String("--types", &typesFlag).
		Int("--limit", &limitFlag).
		Bool("--json", &jsonFlag).
		Help("-h,--help", channelsListHelpText).
		HelpNoExit().
		StopOnFirstArg().
		Parse(args)
	if errors.Is(err, lessflags.ErrHelp) {
		return nil
	}
	if err != nil {
		return err
	}
	if len(remain) > 0 {
		return fmt.Errorf("unexpected arguments: %v", remain)
	}

	token, cfg, types := resolveChannelsAuthAndTypes(tokenFlag, configFlag, typesFlag)
	_ = cfg

	channels, listErr := listChannels(token, types)
	if listErr != nil {
		fmt.Fprintf(os.Stderr, "channels failed: %v\n", listErr)
		os.Exit(1)
	}

	channels = excludeArchived(channels)
	sortChannelsByName(channels)
	channels = applyChannelLimit(channels, limitFlag)
	return printChannels(channels, jsonFlag)
}

func runChannelsSearch(args []string) error {
	var (
		tokenFlag  *string
		configFlag *string
		typesFlag  *string
		limitFlag  *int
		jsonFlag   bool
		exactFlag  bool
		prefixFlag bool
	)

	remain, err := lessflags.String("--token", &tokenFlag).
		String("--config", &configFlag).
		String("--types", &typesFlag).
		Int("--limit", &limitFlag).
		Bool("--json", &jsonFlag).
		Bool("--exact", &exactFlag).
		Bool("--prefix", &prefixFlag).
		Help("-h,--help", channelsSearchHelpText).
		HelpNoExit().
		StopOnFirstArg().
		Parse(args)
	if errors.Is(err, lessflags.ErrHelp) {
		return nil
	}
	if err != nil {
		return err
	}

	if exactFlag && prefixFlag {
		fmt.Fprintln(os.Stderr, "--exact and --prefix are mutually exclusive")
		os.Exit(1)
	}

	if len(remain) == 0 {
		fmt.Fprintln(os.Stderr, "query required")
		os.Exit(1)
	}
	if len(remain) > 1 {
		return fmt.Errorf("unexpected arguments: %v", remain[1:])
	}
	query := remain[0]

	token, cfg, types := resolveChannelsAuthAndTypes(tokenFlag, configFlag, typesFlag)
	_ = cfg

	channels, listErr := listChannels(token, types)
	if listErr != nil {
		fmt.Fprintf(os.Stderr, "channels failed: %v\n", listErr)
		os.Exit(1)
	}

	channels = excludeArchived(channels)
	channels = filterChannelsByName(channels, query, exactFlag, prefixFlag)
	sortChannelsByName(channels)
	channels = applyChannelLimit(channels, limitFlag)
	return printChannels(channels, jsonFlag)
}

func resolveChannelsAuthAndTypes(tokenFlag, configFlag, typesFlag *string) (token string, cfg *slackutil.SlackConfig, types []string) {
	configPath := ""
	if configFlag != nil && *configFlag != "" {
		configPath = *configFlag
	} else if env := os.Getenv("SLACK_CONFIG"); env != "" {
		configPath = env
	}

	if configPath != "" {
		loaded, loadErr := slackutil.Load(configPath)
		if loadErr != nil {
			fmt.Fprintf(os.Stderr, "failed to load config: %v\n", loadErr)
			os.Exit(1)
		}
		cfg = loaded
	}

	token = slackutil.ResolveBotToken(flagString(tokenFlag), "SLACK_BOT_TOKEN", cfg)
	if token == "" {
		if cfg != nil && configPath != "" {
			fmt.Fprintf(os.Stderr, "botToken is empty in %s\n", slackutil.ConfigDisplayPath(configPath))
		} else {
			fmt.Fprintln(os.Stderr, "bot token required")
		}
		os.Exit(1)
	}

	typesStr := "public,private"
	if typesFlag != nil && *typesFlag != "" {
		typesStr = *typesFlag
	}
	types = mapChannelTypes(typesStr)
	return token, cfg, types
}

// mapChannelTypes maps CLI shorthand (public,private) to Slack API types.
func mapChannelTypes(typesStr string) []string {
	parts := strings.Split(typesStr, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		switch p {
		case "public":
			out = append(out, "public_channel")
		case "private":
			out = append(out, "private_channel")
		default:
			// Pass through full API type names (public_channel, private_channel, im, mpim).
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return []string{"public_channel", "private_channel"}
	}
	return out
}

func listChannels(token string, types []string) ([]channelRow, error) {
	if len(types) == 0 {
		types = []string{"public_channel", "private_channel"}
	}
	api := slackutil.NewAPIClient(token)
	multi := len(types) > 1

	var all []channelRow
	var warnings []string
	var lastMissing error
	success := 0

	for _, typ := range types {
		rows, err := listChannelsOfType(api, typ)
		if err != nil {
			code, needed := slackListErrorInfo(err, typ)
			if code == "missing_scope" {
				if multi {
					// Soft-skip this type; continue so other types can succeed.
					warnings = append(warnings, softSkipWarning(typ, needed))
					lastMissing = missingScopeError(needed)
					continue
				}
				// Sole requested type: hard fail with needed= when known.
				return nil, missingScopeError(needed)
			}
			// Non-missing_scope errors hard-fail immediately (no soft-degrade).
			return nil, err
		}
		success++
		all = append(all, rows...)
	}

	if success == 0 {
		if lastMissing != nil {
			return nil, lastMissing
		}
		return nil, fmt.Errorf("missing_scope")
	}

	// Soft-skip warnings only when at least one type succeeded (exit 0 path).
	for _, w := range warnings {
		fmt.Fprintln(os.Stderr, w)
	}
	// Deduplicate by ID: multi-type merges can overlap if an API/mock returns
	// the same conversation for more than one types request.
	return dedupeChannelsByID(all), nil
}

func dedupeChannelsByID(channels []channelRow) []channelRow {
	if len(channels) < 2 {
		return channels
	}
	seen := make(map[string]struct{}, len(channels))
	out := make([]channelRow, 0, len(channels))
	for _, ch := range channels {
		if _, ok := seen[ch.ID]; ok {
			continue
		}
		seen[ch.ID] = struct{}{}
		out = append(out, ch)
	}
	return out
}

func listChannelsOfType(api *slack.Client, typ string) ([]channelRow, error) {
	params := &slack.GetConversationsParameters{
		Types: []string{typ},
		Limit: 200,
	}
	var all []channelRow
	for {
		channels, nextCursor, err := api.GetConversations(params)
		if err != nil {
			return nil, err
		}
		for _, ch := range channels {
			all = append(all, channelRow{
				ID:         ch.ID,
				Name:       ch.Name,
				IsPrivate:  ch.IsPrivate,
				IsMember:   ch.IsMember,
				IsArchived: ch.IsArchived,
			})
		}
		if nextCursor == "" {
			break
		}
		params.Cursor = nextCursor
	}
	return all, nil
}

// slackListErrorInfo extracts Slack error code and needed scope (when known).
// slack-go's SlackErrorResponse does not surface `needed`; fall back to the
// conventional scope for the conversation type.
func slackListErrorInfo(err error, typ string) (code, needed string) {
	code = err.Error()
	var ser slack.SlackErrorResponse
	if errors.As(err, &ser) {
		code = ser.Err
	} else if se, ok := err.(slack.SlackErrorResponse); ok {
		code = se.Err
	}
	if code == "missing_scope" {
		needed = neededScopeForType(typ)
	}
	return code, needed
}

func neededScopeForType(typ string) string {
	switch typ {
	case "private_channel":
		return "groups:read"
	case "public_channel":
		return "channels:read"
	case "im":
		return "im:history"
	case "mpim":
		return "mpim:history"
	default:
		return ""
	}
}

func channelTypeLabel(typ string) string {
	switch typ {
	case "private_channel":
		return "private channels"
	case "public_channel":
		return "public channels"
	case "im":
		return "direct messages"
	case "mpim":
		return "group direct messages"
	default:
		return typ
	}
}

const missingScopeSeeHelp = "; see: slack-msg --help --topic add-missing-scope"

func softSkipWarning(typ, needed string) string {
	label := channelTypeLabel(typ)
	if needed != "" {
		return fmt.Sprintf("warning: skipped %s (missing %s)%s", label, needed, missingScopeSeeHelp)
	}
	return fmt.Sprintf("warning: skipped %s (missing_scope)%s", label, missingScopeSeeHelp)
}

func missingScopeError(needed string) error {
	if needed != "" {
		return fmt.Errorf("missing_scope (needed %s)%s", needed, missingScopeSeeHelp)
	}
	return fmt.Errorf("missing_scope%s", missingScopeSeeHelp)
}

func excludeArchived(channels []channelRow) []channelRow {
	out := make([]channelRow, 0, len(channels))
	for _, ch := range channels {
		if ch.IsArchived {
			continue
		}
		out = append(out, ch)
	}
	return out
}

func sortChannelsByName(channels []channelRow) {
	sort.SliceStable(channels, func(i, j int) bool {
		return channels[i].Name < channels[j].Name
	})
}

func applyChannelLimit(channels []channelRow, limitFlag *int) []channelRow {
	if limitFlag == nil || *limitFlag <= 0 {
		return channels
	}
	if *limitFlag >= len(channels) {
		return channels
	}
	return channels[:*limitFlag]
}

func stripChannelHash(s string) string {
	return strings.TrimPrefix(s, "#")
}

func filterChannelsByName(channels []channelRow, query string, exact, prefix bool) []channelRow {
	q := strings.ToLower(stripChannelHash(query))
	out := make([]channelRow, 0)
	for _, ch := range channels {
		name := strings.ToLower(stripChannelHash(ch.Name))
		match := false
		switch {
		case exact:
			match = name == q
		case prefix:
			match = strings.HasPrefix(name, q)
		default:
			match = strings.Contains(name, q)
		}
		if match {
			out = append(out, ch)
		}
	}
	return out
}

func printChannels(channels []channelRow, asJSON bool) error {
	if asJSON {
		doc := channelsDoc{Channels: channels}
		if doc.Channels == nil {
			doc.Channels = []channelRow{}
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetEscapeHTML(false)
		if err := enc.Encode(doc); err != nil {
			fmt.Fprintf(os.Stderr, "channels failed: %v\n", err)
			os.Exit(1)
		}
		// json.Encoder.Encode already appends a trailing newline.
		return nil
	}

	for _, ch := range channels {
		kind := "public"
		if ch.IsPrivate {
			kind = "private"
		}
		member := "-"
		if ch.IsMember {
			member = "member"
		}
		fmt.Printf("%s  #%s  %s  %s\n", ch.ID, ch.Name, kind, member)
	}
	// Empty human list: empty stdout (no forced blank line), exit 0.
	return nil
}
