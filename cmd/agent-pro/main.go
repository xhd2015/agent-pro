package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	codexcfg "github.com/xhd2015/agent-pro/agent/codex/config"
	codexsessions "github.com/xhd2015/agent-pro/agent/codex/sessions"
	codexskills "github.com/xhd2015/agent-pro/agent/codex/skills"
	grokfork "github.com/xhd2015/agent-pro/agent/grok/fork"
	groksessions "github.com/xhd2015/agent-pro/agent/grok/sessions"
	grokview "github.com/xhd2015/agent-pro/agent/grok/view"
	"github.com/xhd2015/agent-pro/agent/opencode/commands"
	opencodecfg "github.com/xhd2015/agent-pro/agent/opencode/config"
	"github.com/xhd2015/agent-pro/agent/opencode/permissions"
	"github.com/xhd2015/agent-pro/agent/opencode/plugins"
	opencodesessions "github.com/xhd2015/agent-pro/agent/opencode/sessions"
	openskills "github.com/xhd2015/agent-pro/agent/opencode/skills"
	"github.com/xhd2015/agent-pro/frontend"
	"github.com/xhd2015/agent-pro/pkgs/agentconfig"
	"github.com/xhd2015/agent-pro/pkgs/agenttty"
	"github.com/xhd2015/agent-pro/run"
	"github.com/xhd2015/agent-pro/server"
	lib "github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
	"github.com/xhd2015/less-gen/flags"
)

const help = `
Usage: agent-pro <command> [ARGS]

Commands:
  opencode          manage opencode hooks, permissions, and config
  pi                manage pi configuration
  crush             manage crush configuration
  codex             manage codex configuration
  grok              manage grok CLI sessions
  bookmark          list/show/remove session bookmarks (multi-runner catalog)
  proc              resolve agent session from a process id
  skills            list available skills; skills update refreshes installs
  skill             show or install a skill (--show / --install)
  traces            view agent trace sessions (web viewer)
  show-agent-files  collect known agent files under ~/.agent-pro/agent-files-collection/

Run agent-pro <command> --help for command-specific options.
Run agent-pro skill --help and agent-pro skill --install --help for skill options.
`

const opencodeHelp = `
Usage: agent-pro opencode <command> [ARGS]

Commands:
  commands          list opencode slash commands
  config            manage opencode configuration (export/import)
  permissions       manage opencode permissions
  plugins           manage opencode plugins
  skills            list installed skills
  sessions          list OpenCode CLI sessions
  session           show info for one OpenCode CLI session

Run agent-pro opencode <command> --help for command-specific options.
`

func main() {
	server.Init(frontend.DistFS, frontend.TemplateHTML)
	if err := handle(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "agent-pro: %v\n", err)
		os.Exit(1)
	}
}

func handle(args []string) error {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		fmt.Print(strings.TrimPrefix(help, "\n"))
		return nil
	}

	switch args[0] {
	case "opencode":
		return handleOpenCode(args[1:])
	case "codex":
		return handleCodex(args[1:])
	case "grok":
		return handleGrok(args[1:])
	case "bookmark", "bookmarks":
		return handleBookmark(args[1:])
	case "pi":
		return handlePi(args[1:])
	case "crush":
		return handleCrush(args[1:])
	case "skill":
		return handleSkill(args[1:])
	case "skills":
		return handleSkills(args[1:])
	case "proc":
		return handleProc(args[1:])
	case "traces":
		return handleTraces(args[1:])
	case "show-agent-files":
		return handleShowAgentFiles(args[1:])
	default:
		return fmt.Errorf("unknown command: %s", args[0])
	}
}

func handleOpenCode(args []string) error {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		fmt.Print(strings.TrimPrefix(opencodeHelp, "\n"))
		return nil
	}

	switch args[0] {
	case "permissions":
		return handlePermissions(args[1:])
	case "plugins":
		return handlePlugins(args[1:])
	case "commands":
		return handleCommands(args[1:])
	case "config":
		return handleOpenCodeConfig(args[1:])
	case "skills":
		return handleOpenCodeSkills(args[1:])
	case "sessions":
		return handleOpenCodeSessions(args[1:])
	case "session":
		return handleOpenCodeSession(args[1:])
	default:
		return fmt.Errorf("unknown opencode command: %s", args[0])
	}
}

const commandsHelp = `
Usage: agent-pro opencode commands <command> [ARGS]

Commands:
  list [--dir DIR]  list slash commands (global and local)
  doc               show opencode commands reference documentation

Options:
  --dir <dir>       project directory for local commands (default: current directory)
  -h,--help         show help
`

const commandsDoc = `
OpenCode Commands Reference
============================

Commands are custom slash commands that send a pre-defined prompt to the LLM.
They can be defined in JSON config or as markdown files.

Config locations (opencode.jsonc is preferred):

  Global: ~/.config/opencode/opencode.jsonc  (applies to all projects)
  Local:  .opencode/opencode.jsonc           (project-specific, overrides global)

  opencode.jsonc is preferred (supports comments and trailing commas).
  opencode.json is also supported as a fallback.

Definition methods:

  1. JSON config ("command" field in opencode.jsonc):

    {
      "command": {
        "test": {
          "template": "Run the full test suite with coverage report...",
          "description": "Run tests with coverage",
          "agent": "build",
          "model": "anthropic/claude-3-5-sonnet-20241022"
        }
      }
    }

  2. Markdown files (auto-discovered):

    Global: ~/.config/opencode/commands/*.md
    Local:  .opencode/commands/*.md

    Example .opencode/commands/test.md:

      ---
      description: Run tests with coverage
      agent: build
      model: anthropic/claude-3-5-sonnet-20241022
      ---
      Run the full test suite with coverage report and show any failures.
      Focus on the failing tests and suggest fixes.

    The markdown filename ("test.md") becomes the command name (/test).

Template placeholders:

  $ARGUMENTS   All arguments passed to the command
  $1, $2, $3   Positional arguments
  !` + "`command`" + `   Inject shell command output into the prompt
  @filepath    Include referenced file content in the prompt

Options:

  template      (required) The prompt sent to the LLM
  description   Brief description shown in TUI
  agent         Which agent executes this command
  subtask       Force subagent invocation (true/false)
  model         Override the default model for this command

Custom commands can override built-in commands (/init, /undo, /redo, /share, /help).

For more information: https://opencode.ai/docs/commands/
`

func handleCommands(args []string) error {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		fmt.Print(strings.TrimPrefix(commandsHelp, "\n"))
		return nil
	}

	switch args[0] {
	case "list":
		return handleCommandsList(args[1:])
	case "doc":
		return handleCommandsDoc(args[1:])
	default:
		return fmt.Errorf("unknown commands command: %s", args[0])
	}
}

func handleCommandsDoc(args []string) error {
	fmt.Print(strings.TrimPrefix(commandsDoc, "\n"))
	return nil
}

func handleCommandsList(args []string) error {
	var dirFlag *string
	_, err := flags.String("--dir", &dirFlag).
		Help("-h,--help", `
Usage: agent-pro opencode commands list [--dir DIR]

List slash commands from opencode config and markdown files.

Options:
  --dir <dir>   project directory (default: current directory)
  -h,--help     show help
`).
		Parse(args)
	if err != nil {
		return err
	}

	dir := "."
	if dirFlag != nil && strings.TrimSpace(*dirFlag) != "" {
		dir = strings.TrimSpace(*dirFlag)
	}
	dir, err = filepath.Abs(filepath.Clean(dir))
	if err != nil {
		return fmt.Errorf("resolve dir: %w", err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("find home dir: %w", err)
	}

	printCommandsLocation("Global", filepath.Join(home, ".config", "opencode"), dir, home)
	printCommandsLocation("Local", filepath.Join(dir, ".opencode"), dir, home)

	return nil
}

func printCommandsLocation(label, opencodeDir, baseDir, home string) {
	fmt.Printf("%s (%s):\n", label, shortenHome(opencodeDir, home))
	fmt.Printf("  sources: opencode.jsonc command field, {command,commands}/**/*.md\n")

	cmds, err := commands.List(opencodeDir)
	if err != nil {
		fmt.Printf("  (error: %v)\n\n", err)
		return
	}

	if len(cmds) == 0 {
		fmt.Println("  No commands found")
		return
	}

	fmt.Printf("  Found %d command(s):\n", len(cmds))
	for _, c := range cmds {
		loc := formatLocation(c.Path, c.Line, home)
		name := c.Name
		if c.NameAuto {
			name += " [auto]"
		}
		if c.Description != "" {
			fmt.Printf("    %-35s %s\n      %s\n", name, c.Description, loc)
		} else {
			fmt.Printf("    %-35s %s\n", name, loc)
		}
	}
	fmt.Println()
}

const pluginsHelp = `
Usage: agent-pro opencode plugins <command> [ARGS]

Commands:
  list [--dir DIR]    list installed plugins (global and local)
  add <plugin.ts>     install a plugin file to local .opencode/plugins/
  add --global <...>  install a plugin file to global ~/.config/opencode/plugins/
  doc                 show opencode plugins reference documentation

Options:
  --dir <dir>         project directory for local plugins (default: current directory)
  -h,--help           show help
`

const pluginsDoc = `
OpenCode Plugins Reference
===========================

Plugins extend OpenCode by hooking into events and customizing behavior.
They are JavaScript/TypeScript modules that export plugin functions.

Installation methods:

  1. Local files (auto-discovered at startup):
     Global: ~/.config/opencode/plugins/*.{ts,js}
     Local:  .opencode/plugins/*.{ts,js}

  2. npm packages (specified in opencode.jsonc):
     { "plugin": ["opencode-wakatime", "@my-org/custom-plugin"] }

     npm plugins auto-install with Bun; cached in ~/.cache/opencode/node_modules/

Load order:
  1. Global config (~/.config/opencode/opencode.jsonc)
  2. Project config (.opencode/opencode.jsonc)
  3. Global plugin directory (~/.config/opencode/plugins/)
  4. Project plugin directory (.opencode/plugins/)

Basic structure:

  export const MyPlugin = async ({ project, client, $, directory, worktree }) => {
    return {
      // Hook implementations go here
    }
  }

Context parameters:
  project    Current project information
  directory  Current working directory
  worktree   Git worktree path
  client     OpenCode SDK client for interacting with the AI
  $          Bun shell API for executing commands

Dependencies:
  For local plugins using npm packages, create .opencode/package.json:

    { "dependencies": { "shescape": "^2.1.0" } }

  OpenCode runs bun install at startup to install these dependencies.

Available events:
  command.executed        file.edited         file.watcher.updated
  installation.updated    lsp.client.diagnostics  lsp.updated
  message.part.removed    message.part.updated    message.removed
  message.updated         permission.asked        permission.replied
  server.connected        session.created         session.compacted
  session.deleted         session.diff            session.error
  session.idle            session.status          session.updated
  todo.updated            shell.env
  tool.execute.after      tool.execute.before
  tui.prompt.append       tui.command.execute     tui.toast.show
  experimental.session.compacting

Custom tools:
  Plugins can add custom tools via the tool() helper:

    import { type Plugin, tool } from "@opencode-ai/plugin"

    export const CustomToolsPlugin: Plugin = async (ctx) => {
      return {
        tool: {
          mytool: tool({
            description: "This is a custom tool",
            args: { foo: tool.schema.string() },
            async execute(args, context) {
              return "Hello " + args.foo
            },
          }),
        },
      }
    }

For more information: https://opencode.ai/docs/plugins/
`

func handlePlugins(args []string) error {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		fmt.Print(strings.TrimPrefix(pluginsHelp, "\n"))
		return nil
	}

	switch args[0] {
	case "list":
		return handlePluginsList(args[1:])
	case "add":
		return handlePluginsAdd(args[1:])
	case "doc":
		return handlePluginsDoc(args[1:])
	default:
		return fmt.Errorf("unknown plugins command: %s", args[0])
	}
}

func handlePluginsDoc(args []string) error {
	fmt.Print(strings.TrimPrefix(pluginsDoc, "\n"))
	return nil
}

func handlePluginsList(args []string) error {
	var dirFlag *string
	_, err := flags.String("--dir", &dirFlag).
		Help("-h,--help", `
Usage: agent-pro opencode plugins list [--dir DIR]

List installed plugins from auto-discovered plugin directories.

Options:
  --dir <dir>   project directory (default: current directory)
  -h,--help     show help
`).
		Parse(args)
	if err != nil {
		return err
	}

	dir := "."
	if dirFlag != nil && strings.TrimSpace(*dirFlag) != "" {
		dir = strings.TrimSpace(*dirFlag)
	}
	dir, err = filepath.Abs(filepath.Clean(dir))
	if err != nil {
		return fmt.Errorf("resolve dir: %w", err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("find home dir: %w", err)
	}

	printPluginLocation("Global", filepath.Join(home, ".config", "opencode"), dir, home)
	printPluginLocation("Local", filepath.Join(dir, ".opencode"), dir, home)

	return nil
}

func printPluginLocation(label, opencodeDir, baseDir, home string) {
	fmt.Printf("%s (%s):\n", label, shortenHome(opencodeDir, home))
	fmt.Printf("  scanned: plugins/*.ts, plugins/*.js, plugin/*.ts, plugin/*.js\n")

	result, err := plugins.List(opencodeDir)
	if err != nil {
		fmt.Printf("  (error: %v)\n\n", err)
		return
	}

	if len(result) == 0 {
		fmt.Println("  No plugins found")
		return
	}

	fmt.Printf("  Found %d plugin(s):\n", len(result))
	for _, p := range result {
		fmt.Printf("    %s  (%s)\n", p.Name, shortenHome(p.Path, home))
	}
	fmt.Println()
}

func handlePluginsAdd(args []string) error {
	var globalFlag bool
	remaining, err := flags.Bool("--global", &globalFlag).
		Help("-h,--help", `
Usage: agent-pro opencode plugins add <plugin.ts> [--global]

Install a plugin file to the opencode auto-discovered plugins directory.
Without --global, installs to .opencode/plugins/ in the current project.
With --global, installs to ~/.config/opencode/plugins/.

Options:
  --global   install to global plugins directory
  -h,--help  show help
`).
		Parse(args)
	if err != nil {
		return err
	}

	if len(remaining) == 0 {
		return fmt.Errorf("plugin file path is required")
	}
	pluginPath := remaining[0]

	if globalFlag {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("find home dir: %w", err)
		}
		destDir := filepath.Join(home, ".config", "opencode")
		dstPath, err := plugins.InstallToDir(destDir, pluginPath)
		if err != nil {
			return err
		}
		fmt.Printf("Installed plugin: %s\n", dstPath)
		return nil
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get current dir: %w", err)
	}
	destDir := filepath.Join(cwd, ".opencode")
	dstPath, err := plugins.InstallToDir(destDir, pluginPath)
	if err != nil {
		return err
	}
	fmt.Printf("Installed plugin: %s\n", dstPath)
	return nil
}

const permissionsHelp = `
Usage: agent-pro opencode permissions <command> [ARGS]

Commands:
  list [--dir DIR]  list configured permission (deny) rules (global and local)
  doc               show opencode permissions reference documentation

Options:
  --dir <dir>       project directory for local permissions (default: current directory)
  -h,--help         show help
`

const permissionsDoc = `
OpenCode Permissions Reference
==============================

Available permission keys and their value types:

  Key                    Value Type   Matches Against
  ─────────────────────────────────────────────────────
  read                   Rule         File path
  edit                   Rule         File path (covers edit, write, patch)
  glob                   Rule         Glob pattern
  grep                   Rule         Regex pattern
  bash                   Rule         Parsed commands (e.g. git status)
  task                   Rule         Subagent type
  skill                  Rule         Skill name
  lsp                    Rule         LSP queries
  question               Action       Questions during execution
  webfetch               Rule         URL
  websearch              Action       Search query
  external_directory     Rule         Paths outside project working directory
  doom_loop              Action       Repeated identical tool calls
  repo_clone             Rule         Repository cloning
  repo_overview          Rule         Repository overview/analysis
  list                   Rule         Listing files/directories
  todowrite              Action       Writing to the todo list

Actions: "allow", "ask", "deny"

Rule = a plain action string ("allow" shorthand for {"*": "allow"})
       or an object { "pattern": "action", ... }

Action only = a single action string (no per-pattern matching)

Rules are evaluated with the LAST matching rule winning.
Broad rules go first, narrow rules last.

Permission config can be set globally or per-project:

  Global: ~/.config/opencode/opencode.jsonc  (applies to all projects)
  Local:  .opencode/opencode.jsonc           (project-specific, overrides global)

  opencode.jsonc is preferred (supports comments and trailing commas).
  opencode.json is also supported as a fallback.

Example configuration (~/.config/opencode/opencode.jsonc or .opencode/opencode.jsonc):

  {
    "permission": {
      "bash": {
        "*": "ask",
        "git *": "allow",
        "npm run *": "allow",
        "rm *": "deny",
        "grep *": "allow"
      },
      "edit": {
        "*": "deny",
        "packages/web/**/*.mdx": "allow"
      },
      "external_directory": {
        "/tmp/**": "allow",
        "~/projects/**": "allow"
      }
    }
  }

Defaults:
  Most permissions default to "allow".
  doom_loop and external_directory default to "ask".
  .env files are denied for read by default.

Per-agent permissions:

  {
    "agent": {
      "build": {
        "permission": {
          "bash": { "git *": "allow", "*": "ask" },
          "edit": "deny"
        }
      }
    }
  }

For more information: https://opencode.ai/docs/permissions/
`

func handlePermissions(args []string) error {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		fmt.Print(strings.TrimPrefix(permissionsHelp, "\n"))
		return nil
	}

	switch args[0] {
	case "list":
		return handlePermissionsList(args[1:])
	case "doc":
		return handlePermissionsDoc(args[1:])
	default:
		return fmt.Errorf("unknown permissions command: %s", args[0])
	}
}

func handlePermissionsDoc(args []string) error {
	fmt.Print(strings.TrimPrefix(permissionsDoc, "\n"))
	return nil
}

func handlePermissionsList(args []string) error {
	var dirFlag *string
	_, err := flags.String("--dir", &dirFlag).
		Help("-h,--help", `
Usage: agent-pro opencode permissions list [--dir DIR]

List configured permission (deny) rules from the opencode config.

Options:
  --dir <dir>   project directory (default: current directory)
  -h,--help     show help
`).
		Parse(args)
	if err != nil {
		return err
	}

	dir := "."
	if dirFlag != nil && strings.TrimSpace(*dirFlag) != "" {
		dir = strings.TrimSpace(*dirFlag)
	}
	dir, err = filepath.Abs(filepath.Clean(dir))
	if err != nil {
		return fmt.Errorf("resolve dir: %w", err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("find home dir: %w", err)
	}

	printPermissionLocation("Global", filepath.Join(home, ".config", "opencode"), home)
	printPermissionLocation("Local", filepath.Join(dir, ".opencode"), home)

	return nil
}

func printPermissionLocation(label, opencodeDir, home string) {
	fmt.Printf("%s (%s):\n", label, shortenHome(opencodeDir, home))
	fmt.Printf("  scanned: opencode.jsonc, opencode.json\n")

	cfg, err := opencodecfg.ReadDir(opencodeDir)
	if err != nil {
		fmt.Printf("  (error: %v)\n\n", err)
		return
	}

	printPermissionRules(cfg, home)
}

func printPermissionRules(cfg *opencodecfg.Config, home string) {
	bashPerm := permissions.GetBash(cfg.Data)
	if bashPerm == nil {
		fmt.Println("  No deny rules configured")
		return
	}

	obj, ok := bashPerm.(map[string]interface{})
	if !ok {
		fmt.Println("  No deny rules configured")
		return
	}

	keys := opencodecfg.SortedKeys(obj)
	if len(keys) == 0 {
		fmt.Println("  No deny rules configured")
		return
	}

	fmt.Printf("  Found %d rule(s):\n", len(keys))
	for _, key := range keys {
		action, _ := obj[key].(string)
		loc := formatLocation(cfg.Path, opencodecfg.FindKeyLine(cfg.Path, key), home)
		fmt.Printf("    %-40s %s  (%s)\n", key, action, loc)
	}
	fmt.Println()
}

func shortenHome(path string, home string) string {
	if strings.HasPrefix(path, home+string(os.PathSeparator)) {
		return "~" + path[len(home):]
	}
	return path
}

func formatLocation(filePath string, line int, home string) string {
	s := shortenHome(filePath, home)
	if line > 0 {
		s = fmt.Sprintf("%s:%d", s, line)
	}
	return s
}

// --- grok ---

const grokHelp = `
Usage: agent-pro grok <command> [ARGS]

Commands:
  sessions          list recent Grok CLI sessions (table)
  session           per-session ops (info, view, prompts, …)
  sessions prompts  alias for: session prompts (user prompt history)

Run agent-pro grok <command> --help for command-specific options.
`

func handleGrok(args []string) error {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		fmt.Print(strings.TrimPrefix(grokHelp, "\n"))
		return nil
	}

	switch args[0] {
	case "sessions":
		return handleGrokSessions(args[1:])
	case "session":
		return handleGrokSession(args[1:])
	default:
		return fmt.Errorf("unknown grok command: %s", args[0])
	}
}

const grokSessionsHelp = `
Usage: agent-pro grok sessions [OPTIONS]
       agent-pro grok sessions prompts [OPTIONS]   (alias: session prompts)

List recent Grok CLI sessions from ~/.grok (or $GROK_HOME).

For user prompt history (compact [timestamp] lines), use:
  agent-pro grok session prompts …
  agent-pro grok sessions prompts …   (same)

Options:
  --here         only sessions whose cwd matches the process working directory
  --dir DIR      only sessions whose cwd matches DIR (repeatable; OR; must exist)
  --recent W     only sessions active within W (Nd|Nh|Nm, e.g. 1d, 2h, 30m)
  --active       only sessions listed in active_sessions.json
  --main-agent   only main-agent class sessions (mutually exclusive with --sub-agent)
  --sub-agent    only sub-agent class sessions (mutually exclusive with --main-agent)
  --forked       only forked sessions (kind fork/subagent_fork or forked_at set)
  --limit <n>    max sessions to list after filters (default 20, max 100)
  --grep <pat>   search session JSON; print indented hit lines under each match
  --color        always use ANSI colors (even when stdout is not a TTY)
  -h,--help      show help

Place filters (--here / --dir) OR together. Other filters AND with place.
Role filters (--main-agent / --sub-agent) are mutually exclusive; --forked ANDs with role.
`

func handleGrokSessions(args []string) error {
	// Alias: "grok sessions prompts …" → same as "grok session prompts …"
	// (without this, leftover "prompts" was silently ignored and the old table printed).
	if len(args) > 0 && args[0] == "prompts" {
		return handleGrokSessionPrompts(args[1:])
	}

	var limitFlag *int
	var grepFlag *string
	var colorFlag *bool
	var hereFlag *bool
	var dirFlag []string
	var recentFlag *string
	var activeFlag *bool
	var mainAgentFlag *bool
	var subAgentFlag *bool
	var forkedFlag *bool
	remaining, err := flags.Int("--limit", &limitFlag).
		String("--grep", &grepFlag).
		Bool("--color", &colorFlag).
		Bool("--here", &hereFlag).
		StringSlice("--dir", &dirFlag).
		String("--recent", &recentFlag).
		Bool("--active", &activeFlag).
		Bool("--main-agent", &mainAgentFlag).
		Bool("--sub-agent", &subAgentFlag).
		Bool("--forked", &forkedFlag).
		Help("-h,--help", grokSessionsHelp).
		Parse(args)
	if err != nil {
		return err
	}
	if len(remaining) > 0 {
		return fmt.Errorf("unexpected argument %q (did you mean 'grok session prompts' or 'grok sessions prompts'?)", remaining[0])
	}

	limit := 20
	if limitFlag != nil && *limitFlag > 0 {
		limit = *limitFlag
	}

	grokHome := agenttty.GrokHome()
	home := homeDir()
	now := time.Now()

	// Build place set from --here / --dir (OR).
	var placeCWDs []string
	if hereFlag != nil && *hereFlag {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("--here: getwd: %w", err)
		}
		abs, err := filepath.Abs(cwd)
		if err != nil {
			return fmt.Errorf("--here: abs: %w", err)
		}
		placeCWDs = append(placeCWDs, filepath.Clean(abs))
	}
	for _, d := range dirFlag {
		abs, err := filepath.Abs(d)
		if err != nil {
			return fmt.Errorf("--dir %q: %w", d, err)
		}
		abs = filepath.Clean(abs)
		st, err := os.Stat(abs)
		if err != nil {
			return fmt.Errorf("--dir %q: %w", d, err)
		}
		if !st.IsDir() {
			return fmt.Errorf("--dir %q is not a directory", d)
		}
		placeCWDs = append(placeCWDs, abs)
	}

	hasPlace := len(placeCWDs) > 0
	hasRecent := recentFlag != nil
	hasActive := activeFlag != nil && *activeFlag
	hasMainAgent := mainAgentFlag != nil && *mainAgentFlag
	hasSubAgent := subAgentFlag != nil && *subAgentFlag
	hasForked := forkedFlag != nil && *forkedFlag
	hasNewFilters := hasPlace || hasRecent || hasActive || hasMainAgent || hasSubAgent || hasForked

	// Pure --grep (no place/recent/active/role/forked): keep ListWithGrep hit-line path.
	if grepFlag != nil && !hasNewFilters {
		pattern := strings.TrimSpace(*grepFlag)
		if pattern == "" {
			return fmt.Errorf("--grep requires a non-empty pattern")
		}
		colorMode := "auto"
		if colorFlag != nil && *colorFlag {
			colorMode = "always"
		}
		matches, err := groksessions.ListWithGrep(grokHome, limit, pattern)
		if err != nil {
			return fmt.Errorf("list grok sessions: %w", err)
		}
		fmt.Println(groksessions.FormatListTableWithHits(matches, home, now, colorMode))
		return nil
	}

	opts := groksessions.ListOptions{
		Limit:     limit,
		PlaceCWDs: placeCWDs,
		Active:    hasActive,
		Now:       now,
		MainAgent: hasMainAgent,
		SubAgent:  hasSubAgent,
		Forked:    hasForked,
	}
	if hasRecent {
		w, err := groksessions.ParseRecentWindow(*recentFlag)
		if err != nil {
			return err
		}
		opts.Recent = w
		opts.RecentSet = true
	}
	if grepFlag != nil {
		pattern := strings.TrimSpace(*grepFlag)
		if pattern == "" {
			return fmt.Errorf("--grep requires a non-empty pattern")
		}
		opts.Grep = pattern
		opts.GrepSet = true
	}

	sessions, err := groksessions.ListWithOptions(grokHome, opts)
	if err != nil {
		return fmt.Errorf("list grok sessions: %w", err)
	}

	fmt.Println(groksessions.FormatListTable(sessions, home, now))
	return nil
}

const grokSessionHelp = `
Usage: agent-pro grok session <command> [ARGS]

Commands:
  list                  alias for: agent-pro grok sessions
  info   <session-id>   show detailed info for one Grok CLI session
  status <session-id>   show dual-signal liveness (file-active + live PIDs)
  files  <session-id>   list regular files in the session directory
  stats  <session-id>   analyse counts, latency, tools, and tasks for one session
  view   <session-id>   print or web-view session messages (in-memory convert)
  prompts [session-id]  list user prompts (one session or recent multi)
  fork   <session-id>   fork via grok --resume … --fork-session (no agent-run map)
  backup <session-id>   backup session tree to a self-describing directory
  bookmark <session-id> pin session into multi-runner bookmark catalog
  bookmarks             list grok bookmarks (alias: agent-pro bookmark list --runner grok)
  unbookmark <id>       remove grok bookmark (alias: agent-pro bookmark remove --runner grok)
  log    <session-id>   print session log (not implemented yet)

Run agent-pro grok session <command> --help for command-specific options.
`

const grokSessionPromptsHelp = `
Usage:
  agent-pro grok session prompts <session-id>
    [--grep P] [--exclude Q] [--head N | --tail N] [--max-body N]
    [--color|--no-color]
  agent-pro grok session prompts [--recent <window>] [--limit N]
    [--grep P] [--exclude Q] [--head N | --tail N] [--max-body N]
    [--color|--no-color]
  agent-pro grok sessions prompts …   (alias)

Show user prompts only as compact lines:
  [YYYY-MM-DD HH:MM:SS] prompt text…

Single mode: all user prompts for one session (full history), optional text filters.
Multi mode (no session id): newest sessions by last_active, with selection matrix:

  (no flags)              last 10 sessions that have prompts
  --limit N               last N sessions that have prompts (N >= 1)
  --recent Nd|Nh|Nm       all sessions with ≥1 in-window user prompt (no default cap)
  --recent W --limit N    in-window sessions only, stop at N

Filter pipeline (per session, after recent window): grep keep → exclude drop → head|tail.
Sessions with zero survivors are skipped and do not count toward --limit.
Head and tail are mutually exclusive; N >= 1. Empty --grep/--exclude patterns error.

Body length: full collapsed text by default. --max-body N soft-caps each body to
N runes + … (N >= 1). With --grep, full body + highlight unless --max-body, which
windows around the first match within N runes.

Multi layout: session header, prompt lines, separator rule between sessions, footer.
Output streams session-by-session (not buffered until the end).

Options:
  --recent WINDOW   time window: Nd, Nh, or Nm (e.g. 1d, 2h, 30m)
  --limit N         session limit (see matrix above; must be >= 1)
  --grep P          keep prompts whose text matches P (case-insensitive literal)
  --exclude Q       drop prompts whose text matches Q (case-insensitive literal)
  --head N          first N prompts per session after text filters (N >= 1)
  --tail N          last N prompts per session after text filters (N >= 1)
  --max-body N      soft-cap each prompt body to N runes + … (N >= 1; default: full)
  --color           force ANSI color on (even when stdout is not a TTY)
  --no-color        force ANSI color off
  -h,--help         show help

Color (auto by default): TTY on unless NO_COLOR is set; --color/--no-color override.
With --grep and color on, match spans are bold-red; omission markers are dim.
Headers, timestamps, separators, and footer are dim when color is on.

Notes:
  - session-id cannot be combined with --recent or --limit
  - sessions with zero user prompts (or zero in-window prompts) are skipped
`

const grokSessionForkHelp = `
Usage: agent-pro grok session fork <session-id> [OPTIONS]

Fork a Grok CLI session using the native flag:
  grok --resume <session-id> --fork-session

Does not use agent-run storage (avoids already-mapped import limits).

Options:
  -n, --new-terminal   open a new iTerm2 window and run the fork command there
  --dir DIR            workspace (default: session info.cwd from GROK_HOME)
  --session-id UUID    optional id for the forked Grok session
  --dry-run            print plan only; do not launch
  -h,--help            show help
`

const grokSessionInfoHelp = `
Usage: agent-pro grok session info <session-id> [OPTIONS]

Show detailed info for one Grok CLI session from ~/.grok (or $GROK_HOME).
Appends a dual-signal Active block (file-active + live PIDs).

Options:
  --no-pid      skip live PID scan; Active state from file-active only
  -h,--help     show help
`

const grokSessionStatusHelp = `
Usage: agent-pro grok session status <session-id> [OPTIONS]

Show dual-signal liveness for one Grok CLI session:
  file-active (active_sessions.json) + live PIDs (open-file hard hits on grok runners).

State: running | marked-active | inactive

Options:
  --no-pid      skip live PID scan; state from file-active only
  --json        print SessionStatus as JSON (no ANSI)
  -h,--help     show help
`

const grokSessionFilesHelp = `
Usage: agent-pro grok session files <session-id> [OPTIONS]

List regular files in a Grok CLI session directory.

Options:
  --json        print files as a JSON array (no ANSI)
  -h,--help     show help
`

const grokSessionStatsHelp = `
Usage: agent-pro grok session stats <session-id> [OPTIONS]

Analyse one Grok CLI session: counts and latency (signals.json), per-tool
handler times (events.jsonl), thinking blocks / background tasks / subagents
(updates.jsonl).

Options:
  --json        print SessionStats as JSON (raw ms, no ANSI)
  --by-tool     include per-tool table (default: included when tools exist)
  --top N       top-N tools/tasks/subagents sections (default 5; 0 hides)
  --color       always use ANSI colors (even when stdout is not a TTY)
  --no-color    disable ANSI colors
  -h,--help     show help
`

const grokSessionViewHelp = `
Usage: agent-pro grok session view <session-id> [OPTIONS]

View a Grok CLI session transcript. Converts updates.jsonl to standard
events fully in memory (does not write agent-run storage).

Without --web: print human-readable events to stdout (print mode).
With --web:    start a local read-only web viewer using the agent-run UI.

Options:
  --web         serve read-only web UI (live follow via file watch; prints URL)
  --follow      print mode: after existing events, keep printing until Ctrl+C
  --port N      with --web: preferred listen port (default 61781; tries next
                up to 100 ports if taken)
  --open        with --web: open the viewer in a browser
  -h,--help     show help
`

const grokSessionBackupHelp = `
Usage: agent-pro grok session backup <session-id> [OPTIONS]

Backup one Grok CLI session to a self-describing directory
(manifest.json + payload/). Optionally create a .tar.gz archive.

Busy sessions (file-active or live PIDs) are refused. Linked child sessions
are included by default.

Options:
  --out-dir DIR     write backup directory here (default: temp dir, kept)
  -o,--output PATH  also create archive at PATH (must end with .tar.gz)
  --no-children     skip copying linked child session directories
  --dry-run         plan only; print what would be backed up (no writes)
  --json            print BackupResult as JSON
  -h,--help         show help
`

func handleGrokSession(args []string) error {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		fmt.Print(strings.TrimPrefix(grokSessionHelp, "\n"))
		return nil
	}

	switch args[0] {
	case "list", "ls":
		// Alias for agent-pro grok sessions (same flags: --here, --dir, --recent,
		// --active, --main-agent, --sub-agent, --forked, --limit, --grep, --color).
		return handleGrokSessions(args[1:])
	case "info":
		return handleGrokSessionInfo(args[1:])
	case "status":
		return handleGrokSessionStatus(args[1:])
	case "files":
		return handleGrokSessionFiles(args[1:])
	case "stats":
		return handleGrokSessionStats(args[1:])
	case "view":
		return handleGrokSessionView(args[1:])
	case "prompts":
		return handleGrokSessionPrompts(args[1:])
	case "fork":
		return handleGrokSessionFork(args[1:])
	case "backup":
		return handleGrokSessionBackup(args[1:])
	case "bookmark":
		return handleGrokSessionBookmark(args[1:])
	case "bookmarks":
		return handleBookmarkList(append([]string{"--runner", "grok"}, args[1:]...))
	case "unbookmark":
		return handleBookmarkRemove(append([]string{"--runner", "grok"}, args[1:]...))
	case "log":
		return fmt.Errorf("not implemented yet")
	default:
		return fmt.Errorf("unknown grok session command: %s", args[0])
	}
}

func handleGrokSessionPrompts(args []string) error {
	var recentFlag *string
	var limitFlag *int
	var grepFlag *string
	var excludeFlag *string
	var headFlag *int
	var tailFlag *int
	var maxBodyFlag *int
	var colorFlag bool
	var noColorFlag bool
	remaining, err := flags.String("--recent", &recentFlag).
		Int("--limit", &limitFlag).
		String("--grep", &grepFlag).
		String("--exclude", &excludeFlag).
		Int("--head", &headFlag).
		Int("--tail", &tailFlag).
		Int("--max-body", &maxBodyFlag).
		Bool("--color", &colorFlag).
		Bool("--no-color", &noColorFlag).
		Help("-h,--help", grokSessionPromptsHelp).
		Parse(args)
	if err != nil {
		return err
	}
	if colorFlag && noColorFlag {
		return fmt.Errorf("--color and --no-color cannot be specified together")
	}
	colorMode := "auto"
	if colorFlag {
		colorMode = "always"
	}
	if noColorFlag {
		colorMode = "never"
	}

	recentSet := recentFlag != nil
	limitSet := limitFlag != nil
	grepSet := grepFlag != nil
	excludeSet := excludeFlag != nil
	headSet := headFlag != nil
	tailSet := tailFlag != nil
	maxBodySet := maxBodyFlag != nil

	var recent time.Duration
	if recentSet {
		w, err := groksessions.ParseRecentWindow(*recentFlag)
		if err != nil {
			return err
		}
		recent = w
	}
	limit := 0
	if limitSet {
		limit = *limitFlag
		if limit < 1 {
			return fmt.Errorf("--limit must be >= 1")
		}
	}

	grep := ""
	if grepSet {
		grep = *grepFlag
		if grep == "" {
			return fmt.Errorf("--grep pattern must not be empty")
		}
	}
	exclude := ""
	if excludeSet {
		exclude = *excludeFlag
		if exclude == "" {
			return fmt.Errorf("--exclude pattern must not be empty")
		}
	}
	if headSet && tailSet {
		return fmt.Errorf("--head and --tail are mutually exclusive")
	}
	head := 0
	if headSet {
		head = *headFlag
		if head < 1 {
			return fmt.Errorf("--head must be >= 1")
		}
	}
	tail := 0
	if tailSet {
		tail = *tailFlag
		if tail < 1 {
			return fmt.Errorf("--tail must be >= 1")
		}
	}
	maxBodyRunes := 0
	if maxBodySet {
		maxBodyRunes = *maxBodyFlag
		if maxBodyRunes < 1 {
			return fmt.Errorf("--max-body must be >= 1 (got %d)", maxBodyRunes)
		}
	}
	filterOpts := groksessions.FilterUserPromptsOptions{
		Grep:       grep,
		GrepSet:    grepSet,
		Exclude:    exclude,
		ExcludeSet: excludeSet,
		Head:       head,
		HeadSet:    headSet,
		Tail:       tail,
		TailSet:    tailSet,
	}
	hasFilter := grepSet || excludeSet || headSet || tailSet

	// Single-session mode: exactly one id, no --recent/--limit.
	if len(remaining) > 1 {
		return fmt.Errorf("expected at most one session id, got %d arguments", len(remaining))
	}
	if len(remaining) == 1 {
		if recentSet || limitSet {
			return fmt.Errorf("session id cannot be combined with --recent or --limit")
		}
		sessionID := strings.TrimSpace(remaining[0])
		if sessionID == "" {
			return fmt.Errorf("session id is required")
		}
		grokHome := agenttty.GrokHome()
		sp, err := groksessions.Prompts(grokHome, sessionID)
		if err != nil {
			return err
		}
		if hasFilter {
			kept, ob, oa, err := groksessions.FilterUserPrompts(sp.UserPrompts, filterOpts)
			if err != nil {
				return err
			}
			sp.UserPrompts = kept
			sp.OmittedBefore = ob
			sp.OmittedAfter = oa
		}
		bw := bufio.NewWriter(os.Stdout)
		defer bw.Flush()
		return groksessions.WritePromptsText(bw, sp, groksessions.FormatPromptsOptions{
			Now:          time.Now(),
			Home:         homeDir(),
			ColorMode:    colorMode,
			Grep:         grep,
			GrepSet:      grepSet,
			MaxBodyRunes: maxBodyRunes,
			MaxBodySet:   maxBodySet,
		})
	}

	// Multi mode: discover summaries, then load+print each session before the next
	// (true progressive stdout — do not ListPrompts-buffer then dump).
	now := time.Now()
	// Small buffer + explicit Flush after each session in StreamPromptsList so the
	// terminal shows the first block without waiting for later updates.jsonl reads.
	bw := bufio.NewWriterSize(os.Stdout, 1024)
	defer bw.Flush()
	fmt.Fprintln(os.Stderr, "scanning sessions…")
	return groksessions.StreamPromptsList(bw, agenttty.GrokHome(), groksessions.ListPromptsOptions{
		Now:        now,
		Recent:     recent,
		RecentSet:  recentSet,
		Limit:      limit,
		LimitSet:   limitSet,
		Home:       homeDir(),
		Grep:       grep,
		GrepSet:    grepSet,
		Exclude:    exclude,
		ExcludeSet: excludeSet,
		Head:       head,
		HeadSet:    headSet,
		Tail:       tail,
		TailSet:    tailSet,
	}, groksessions.FormatPromptsOptions{
		Now:          now,
		Home:         homeDir(),
		Window:       recent,
		Limit:        limit,
		RecentSet:    recentSet,
		LimitSet:     limitSet,
		ColorMode:    colorMode,
		Grep:         grep,
		GrepSet:      grepSet,
		MaxBodyRunes: maxBodyRunes,
		MaxBodySet:   maxBodySet,
	})
}

// grokSessionOpenInNewTerminal is injectable for tests (default: iTerm ForceNew).
var grokSessionOpenInNewTerminal = defaultGrokSessionOpenInNewTerminal

func defaultGrokSessionOpenInNewTerminal(dir, followUp string) error {
	return lib.OpenConfig(dir, &lib.Config{
		Mode:             lib.ModeForceNew,
		FollowUpCommands: []string{followUp},
	})
}

func handleGrokSessionFork(args []string) error {
	var newTerminal bool
	var dryRun bool
	var dir string
	var newSessionID string
	remaining, err := flags.Bool("-n,--new-terminal", &newTerminal).
		Bool("--dry-run", &dryRun).
		String("--dir", &dir).
		String("--session-id", &newSessionID).
		Help("-h,--help", grokSessionForkHelp).
		Parse(args)
	if err != nil {
		return err
	}
	if len(remaining) != 1 {
		return fmt.Errorf("expected exactly one session id, got %d arguments", len(remaining))
	}
	sessionID := strings.TrimSpace(remaining[0])
	if sessionID == "" {
		return fmt.Errorf("session id is required")
	}

	grokHome := agenttty.GrokHome()
	info, err := groksessions.Info(grokHome, sessionID)
	if err != nil {
		return err
	}

	cwd := strings.TrimSpace(dir)
	if cwd == "" {
		cwd = strings.TrimSpace(info.CWD)
	}
	if cwd == "" {
		return fmt.Errorf("session %s has empty cwd; pass --dir", sessionID)
	}
	abs, absErr := filepath.Abs(cwd)
	if absErr != nil {
		return fmt.Errorf("workspace dir: %w", absErr)
	}
	if real, e := filepath.EvalSymlinks(abs); e == nil {
		abs = real
	}
	st, stErr := os.Stat(abs)
	if stErr != nil {
		return fmt.Errorf("workspace dir: %w", stErr)
	}
	if !st.IsDir() {
		return fmt.Errorf("workspace dir: not a directory: %s", abs)
	}
	cwd = abs

	launch := grokfork.SessionLaunch{
		SessionID:    sessionID,
		NewSessionID: newSessionID,
		Dir:          cwd,
		Bin:          "grok",
		Env:          os.Environ(),
	}
	cmdLine := launch.QuotedCommandLine()

	if dryRun {
		fmt.Println("Would fork grok session")
		fmt.Printf("  grok id:   %s\n", sessionID)
		fmt.Printf("  cwd:       %s\n", cwd)
		fmt.Printf("  command:   %s\n", cmdLine)
		if newTerminal {
			fmt.Println("  terminal:  new iTerm2 window")
		} else {
			fmt.Println("  terminal:  current")
		}
		return nil
	}

	if newTerminal {
		// Keep -n opening iTerm with grok --resume --fork-session (not grok-fork --session-id).
		if err := grokSessionOpenInNewTerminal(cwd, cmdLine); err != nil {
			return fmt.Errorf("open new terminal: %w", err)
		}
		fmt.Printf("Opened new window; forking grok session %s\n", sessionID)
		return nil
	}

	if err := grokfork.Fork(&grokfork.Options{
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Env:    os.Environ(),
	}, sessionID, newSessionID, cwd); err != nil {
		return fmt.Errorf("grok fork: %w", err)
	}
	return nil
}

func handleGrokSessionInfo(args []string) error {
	var noPID bool
	remaining, err := flags.Bool("--no-pid", &noPID).
		Help("-h,--help", grokSessionInfoHelp).
		Parse(args)
	if err != nil {
		return err
	}
	if len(remaining) != 1 {
		return fmt.Errorf("expected exactly one session id, got %d arguments", len(remaining))
	}

	sessionID := strings.TrimSpace(remaining[0])
	if sessionID == "" {
		return fmt.Errorf("session id is required")
	}

	grokHome := agenttty.GrokHome()
	info, err := groksessions.Info(grokHome, sessionID)
	if err != nil {
		return err
	}

	fmt.Println(groksessions.FormatInfoText(info, homeDir(), time.Now()))

	st, err := groksessions.Status(grokHome, sessionID, !noPID, &groksessions.LiveOptions{
		ListProcs: nil, // production defaults inside LivePIDsForSession
		Lsof:      nil,
	})
	if err != nil {
		return err
	}
	fmt.Println()
	fmt.Println(groksessions.FormatActiveBlock(st))
	return nil
}

func handleGrokSessionStatus(args []string) error {
	var noPID bool
	var jsonFlag *bool
	remaining, err := flags.Bool("--no-pid", &noPID).
		Bool("--json", &jsonFlag).
		Help("-h,--help", grokSessionStatusHelp).
		Parse(args)
	if err != nil {
		return err
	}
	if len(remaining) != 1 {
		return fmt.Errorf("expected exactly one session id, got %d arguments", len(remaining))
	}

	sessionID := strings.TrimSpace(remaining[0])
	if sessionID == "" {
		return fmt.Errorf("session id is required")
	}

	grokHome := agenttty.GrokHome()
	st, err := groksessions.Status(grokHome, sessionID, !noPID, nil)
	if err != nil {
		return err
	}

	if jsonFlag != nil && *jsonFlag {
		out, err := groksessions.FormatStatusJSON(st)
		if err != nil {
			return fmt.Errorf("format session status json: %w", err)
		}
		fmt.Println(out)
		return nil
	}

	fmt.Println(groksessions.FormatStatusText(st))
	return nil
}

func handleGrokSessionFiles(args []string) error {
	var jsonFlag *bool
	remaining, err := flags.Bool("--json", &jsonFlag).
		Help("-h,--help", grokSessionFilesHelp).
		Parse(args)
	if err != nil {
		return err
	}
	if len(remaining) != 1 {
		return fmt.Errorf("expected exactly one session id, got %d arguments", len(remaining))
	}

	sessionID := strings.TrimSpace(remaining[0])
	if sessionID == "" {
		return fmt.Errorf("session id is required")
	}

	grokHome := agenttty.GrokHome()
	_, files, err := groksessions.ListSessionFiles(grokHome, sessionID)
	if err != nil {
		return err
	}

	if jsonFlag != nil && *jsonFlag {
		out, err := groksessions.FormatSessionFilesJSON(files)
		if err != nil {
			return fmt.Errorf("format session files json: %w", err)
		}
		fmt.Println(out)
		return nil
	}

	fmt.Println(groksessions.FormatSessionFilesTable(files))
	return nil
}

func handleGrokSessionBackup(args []string) error {
	var outDir string
	var archivePath string
	var noChildren bool
	var dryRun bool
	var jsonFlag *bool
	remaining, err := flags.String("--out-dir", &outDir).
		String("-o,--output", &archivePath).
		Bool("--no-children", &noChildren).
		Bool("--dry-run", &dryRun).
		Bool("--json", &jsonFlag).
		Help("-h,--help", grokSessionBackupHelp).
		Parse(args)
	if err != nil {
		return err
	}
	if len(remaining) != 1 {
		return fmt.Errorf("expected exactly one session id, got %d arguments", len(remaining))
	}
	sessionID := strings.TrimSpace(remaining[0])
	if sessionID == "" {
		return fmt.Errorf("session id is required")
	}

	include := !noChildren
	opts := &groksessions.BackupOptions{
		GrokHome:        agenttty.GrokHome(),
		OutDir:          outDir,
		ArchivePath:     archivePath,
		IncludeChildren: &include,
		Live:            nil, // production defaults
		DryRun:          dryRun,
	}
	result, err := groksessions.Backup(sessionID, opts)
	if err != nil {
		return err
	}

	if jsonFlag != nil && *jsonFlag {
		data, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return fmt.Errorf("format backup result json: %w", err)
		}
		fmt.Println(string(data))
		return nil
	}

	if result.DryRun {
		fmt.Println("Would backup grok session")
		fmt.Printf("  session:  %s\n", result.SessionID)
		fmt.Printf("  cwd:      %s\n", result.CWD)
		fmt.Printf("  related:  %s\n", strings.Join(result.RelatedSessions, ", "))
		fmt.Printf("  files:    %d\n", result.PlannedFiles)
		fmt.Printf("  bytes:    %d\n", result.PlannedBytes)
		if outDir != "" {
			fmt.Printf("  out-dir:  %s\n", outDir)
		}
		if archivePath != "" {
			fmt.Printf("  archive:  %s\n", archivePath)
		}
		return nil
	}

	fmt.Printf("Backup written to %s\n", result.Dir)
	fmt.Printf("  session:  %s\n", result.SessionID)
	fmt.Printf("  manifest: %s\n", result.ManifestPath)
	if result.ArchivePath != "" {
		fmt.Printf("  archive:  %s\n", result.ArchivePath)
	}
	return nil
}

func handleGrokSessionStats(args []string) error {
	var jsonFlag *bool
	var byToolFlag *bool
	var colorFlag *bool
	var noColorFlag *bool
	topN := 5
	remaining, err := flags.Bool("--json", &jsonFlag).
		Bool("--by-tool", &byToolFlag).
		Bool("--color", &colorFlag).
		Bool("--no-color", &noColorFlag).
		Int("--top", &topN).
		Help("-h,--help", grokSessionStatsHelp).
		Parse(args)
	if err != nil {
		return err
	}
	if colorFlag != nil && *colorFlag && noColorFlag != nil && *noColorFlag {
		return fmt.Errorf("--color and --no-color are mutually exclusive")
	}
	if len(remaining) != 1 {
		return fmt.Errorf("expected exactly one session id, got %d arguments", len(remaining))
	}

	sessionID := strings.TrimSpace(remaining[0])
	if sessionID == "" {
		return fmt.Errorf("session id is required")
	}

	grokHome := agenttty.GrokHome()
	stats, err := groksessions.Stats(grokHome, sessionID)
	if err != nil {
		return err
	}

	// Surface source warnings on stderr (human + JSON).
	for _, w := range stats.Sources.Warnings {
		fmt.Fprintf(os.Stderr, "warning: %s\n", w)
	}

	if jsonFlag != nil && *jsonFlag {
		// Optional: --by-tool false could strip tools, but default keep full struct.
		_ = byToolFlag
		data, err := json.MarshalIndent(stats, "", "  ")
		if err != nil {
			return fmt.Errorf("format session stats json: %w", err)
		}
		fmt.Println(string(data))
		return nil
	}

	// Text mode: FormatStatsTextOpts prints tools when non-empty.
	// --by-tool is reserved for future force-include; default matches that.
	_ = byToolFlag
	colorMode := "auto"
	if colorFlag != nil && *colorFlag {
		colorMode = "always"
	}
	if noColorFlag != nil && *noColorFlag {
		colorMode = "never"
	}
	fmt.Println(groksessions.FormatStatsTextOpts(stats, groksessions.FormatStatsOptions{
		Home:      homeDir(),
		Now:       time.Now(),
		ColorMode: colorMode,
		TopN:      topN,
	}))
	return nil
}

func handleGrokSessionView(args []string) error {
	var webFlag bool
	var followFlag bool
	var openFlag bool
	port := 0
	remaining, err := flags.Bool("--web", &webFlag).
		Bool("--follow", &followFlag).
		Bool("--open", &openFlag).
		Int("--port", &port).
		Help("-h,--help", grokSessionViewHelp).
		Parse(args)
	if err != nil {
		return err
	}
	if len(remaining) != 1 {
		return fmt.Errorf("expected exactly one session id, got %d arguments", len(remaining))
	}
	sessionID := strings.TrimSpace(remaining[0])
	if sessionID == "" {
		return fmt.Errorf("session id is required")
	}
	if openFlag && !webFlag {
		return fmt.Errorf("--open requires --web")
	}
	if port != 0 && !webFlag {
		return fmt.Errorf("--port requires --web")
	}

	grokHome := agenttty.GrokHome()
	viewer, err := grokview.Open(grokHome, sessionID)
	if err != nil {
		return err
	}
	if err := viewer.Bootstrap(); err != nil {
		return fmt.Errorf("load session updates: %w", err)
	}

	if webFlag {
		return grokview.ServeWeb(context.Background(), viewer, grokview.WebOptions{
			Port:   port,
			Open:   openFlag,
			Stderr: os.Stderr,
		})
	}

	if followFlag {
		return grokview.PrintFollow(context.Background(), os.Stdout, viewer)
	}
	grokview.PrintSnapshot(os.Stdout, viewer)
	return nil
}

// --- codex ---

const codexHelp = `
Usage: agent-pro codex <command> [ARGS]

Commands:
  model             show configured model settings
  model-providers   list custom model providers
  mcp               list MCP server configurations
  projects          list trusted project paths
  plugins           list installed plugins
  skills            list installed skills
  sessions          list Codex CLI rollout sessions
  session           show info or log for one Codex CLI session
  features          show enabled/disabled feature flags
  doc               show codex config.toml reference documentation

Run agent-pro codex <command> --help for command-specific options.
`

func handleCodex(args []string) error {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		fmt.Print(strings.TrimPrefix(codexHelp, "\n"))
		return nil
	}

	switch args[0] {
	case "model":
		return handleCodexModel(args[1:])
	case "model-providers":
		return handleCodexModelProviders(args[1:])
	case "mcp":
		return handleCodexMCP(args[1:])
	case "projects":
		return handleCodexProjects(args[1:])
	case "plugins":
		return handleCodexPlugins(args[1:])
	case "skills":
		return handleCodexSkills(args[1:])
	case "sessions":
		return handleCodexSessions(args[1:])
	case "session":
		return handleCodexSession(args[1:])
	case "features":
		return handleCodexFeatures(args[1:])
	case "doc":
		return handleCodexDoc(args[1:])
	default:
		return fmt.Errorf("unknown codex command: %s", args[0])
	}
}

func loadCodexConfig() (*codexcfg.Config, error) {
	return codexcfg.ReadDefault()
}

// --- codex model ---

const codexModelHelp = `
Usage: agent-pro codex model

Show the configured model and reasoning settings from ~/.codex/config.toml.

Options:
  -h,--help     show help
`

func handleCodexModel(args []string) error {
	_, err := flags.
		Help("-h,--help", codexModelHelp).
		Parse(args)
	if err != nil {
		return err
	}

	cfg, err := loadCodexConfig()
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}

	fmt.Printf("Config: %s\n\n", shortenHome(cfg.Path, homeDir()))

	if cfg.Model != "" {
		fmt.Printf("  model:                         %s\n", cfg.Model)
	}
	if cfg.ModelReasoningEffort != "" {
		fmt.Printf("  model_reasoning_effort:        %s\n", cfg.ModelReasoningEffort)
	}
	if cfg.ModelReasoningSummary != "" {
		fmt.Printf("  model_reasoning_summary:       %s\n", cfg.ModelReasoningSummary)
	}
	if cfg.ModelVerbosity != "" {
		fmt.Printf("  model_verbosity:               %s\n", cfg.ModelVerbosity)
	}
	if cfg.ModelProvider != "" {
		fmt.Printf("  model_provider:                %s\n", cfg.ModelProvider)
	}
	if cfg.DefaultPermissions != "" {
		fmt.Printf("  default_permissions:           %s\n", cfg.DefaultPermissions)
	}
	if cfg.WebSearch != "" {
		fmt.Printf("  web_search:                    %s\n", cfg.WebSearch)
	}
	if cfg.Personality != "" {
		fmt.Printf("  personality:                   %s\n", cfg.Personality)
	}
	if cfg.ApprovalPolicy != "" {
		fmt.Printf("  approval_policy:               %s\n", cfg.ApprovalPolicy)
	}
	if cfg.SandboxMode != "" {
		fmt.Printf("  sandbox_mode:                  %s\n", cfg.SandboxMode)
	}
	if cfg.ModelInstructionsFile != "" {
		fmt.Printf("  model_instructions_file:       %s\n", cfg.ModelInstructionsFile)
	}
	if cfg.DeveloperInstructions != "" {
		fmt.Printf("  developer_instructions:        %s\n", truncate(cfg.DeveloperInstructions, 80))
	}

	return nil
}

// --- codex model-providers ---

const codexModelProvidersHelp = `
Usage: agent-pro codex model-providers [list]

List custom model providers from ~/.codex/config.toml.

Options:
  -h,--help     show help
`

func handleCodexModelProviders(args []string) error {
	_, err := flags.
		Help("-h,--help", codexModelProvidersHelp).
		Parse(args)
	if err != nil {
		return err
	}

	cfg, err := loadCodexConfig()
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}

	fmt.Printf("Config: %s\n\n", shortenHome(cfg.Path, homeDir()))

	if len(cfg.ModelProviders) == 0 {
		fmt.Println("No custom model providers configured.")
		return nil
	}

	fmt.Printf("Found %d provider(s):\n\n", len(cfg.ModelProviders))
	for _, entry := range sortedModelProviders(cfg.ModelProviders) {
		mp := entry.ModelProvider
		fmt.Printf("  [%s]\n", entry.ID)
		if mp.Name != "" {
			fmt.Printf("    name:                %s\n", mp.Name)
		}
		if mp.BaseURL != "" {
			fmt.Printf("    base_url:            %s\n", mp.BaseURL)
		}
		if mp.RequiresOpenAIAuth != nil {
			fmt.Printf("    requires_openai_auth: %v\n", *mp.RequiresOpenAIAuth)
		}
		if mp.WireAPI != "" {
			fmt.Printf("    wire_api:            %s\n", mp.WireAPI)
		}
		if mp.EnvKey != "" {
			fmt.Printf("    env_key:             %s\n", mp.EnvKey)
		}
		fmt.Println()
	}
	return nil
}

// --- codex mcp ---

const codexMCPHelp = `
Usage: agent-pro codex mcp [list]

List MCP server configurations from ~/.codex/config.toml.

Options:
  -h,--help     show help
`

func handleCodexMCP(args []string) error {
	_, err := flags.
		Help("-h,--help", codexMCPHelp).
		Parse(args)
	if err != nil {
		return err
	}

	cfg, err := loadCodexConfig()
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}

	fmt.Printf("Config: %s\n\n", shortenHome(cfg.Path, homeDir()))

	if len(cfg.MCPServers) == 0 {
		fmt.Println("No MCP servers configured.")
		return nil
	}

	fmt.Printf("Found %d MCP server(s):\n\n", len(cfg.MCPServers))
	for _, entry := range sortedMCPServers(cfg.MCPServers) {
		srv := entry.MCPServer
		fmt.Printf("  [%s]\n", entry.ID)
		if srv.Command != "" {
			fmt.Printf("    command:  %s\n", srv.Command)
			if len(srv.Args) > 0 {
				fmt.Printf("    args:     %s\n", strings.Join(srv.Args, " "))
			}
		}
		if srv.URL != "" {
			fmt.Printf("    url:      %s\n", srv.URL)
		}
		if srv.Enabled != nil {
			fmt.Printf("    enabled:  %v\n", *srv.Enabled)
		}
		fmt.Println()
	}
	return nil
}

// --- codex projects ---

const codexProjectsHelp = `
Usage: agent-pro codex projects [list]

List trusted project paths from ~/.codex/config.toml.

Options:
  -h,--help     show help
`

func handleCodexProjects(args []string) error {
	_, err := flags.
		Help("-h,--help", codexProjectsHelp).
		Parse(args)
	if err != nil {
		return err
	}

	cfg, err := loadCodexConfig()
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}

	fmt.Printf("Config: %s\n\n", shortenHome(cfg.Path, homeDir()))

	if len(cfg.Projects) == 0 {
		fmt.Println("No projects configured.")
		return nil
	}

	home := homeDir()
	fmt.Printf("Found %d project(s):\n\n", len(cfg.Projects))
	for _, p := range sortedProjects(cfg.Projects) {
		fmt.Printf("  %-60s trust_level = %s\n", shortenHome(p.Path, home), p.TrustLevel)
	}
	return nil
}

// --- codex plugins ---

const codexPluginsHelp = `
Usage: agent-pro codex plugins [list]

List installed plugins from ~/.codex/config.toml.

Options:
  -h,--help     show help
`

func handleCodexPlugins(args []string) error {
	_, err := flags.
		Help("-h,--help", codexPluginsHelp).
		Parse(args)
	if err != nil {
		return err
	}

	cfg, err := loadCodexConfig()
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}

	fmt.Printf("Config: %s\n\n", shortenHome(cfg.Path, homeDir()))

	if len(cfg.Plugins) == 0 {
		fmt.Println("No plugins configured.")
		return nil
	}

	fmt.Printf("Found %d plugin(s):\n\n", len(cfg.Plugins))
	for _, entry := range sortedPlugins(cfg.Plugins) {
		p := entry.Plugin
		status := "enabled"
		if p.Enabled != nil && !*p.Enabled {
			status = "disabled"
		}
		fmt.Printf("  %-50s %s\n", entry.ID, status)
	}
	return nil
}

// --- codex features ---

const codexFeaturesHelp = `
Usage: agent-pro codex features [list]

Show enabled/disabled feature flags from ~/.codex/config.toml.

Options:
  -h,--help     show help
`

func handleCodexFeatures(args []string) error {
	_, err := flags.
		Help("-h,--help", codexFeaturesHelp).
		Parse(args)
	if err != nil {
		return err
	}

	cfg, err := loadCodexConfig()
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}

	fmt.Printf("Config: %s\n\n", shortenHome(cfg.Path, homeDir()))

	f := cfg.Features
	features := []struct {
		name    string
		enabled *bool
	}{
		{"apps", f.Apps},
		{"codex_git_commit", f.CodexGitCommit},
		{"fast_mode", f.FastMode},
		{"hooks", f.Hooks},
		{"memories", f.Memories},
		{"multi_agent", f.MultiAgent},
		{"personality", f.Personality},
		{"shell_snapshot", f.ShellSnapshot},
		{"shell_tool", f.ShellTool},
		{"unified_exec", f.UnifiedExec},
		{"undo", f.Undo},
		{"prevent_idle_sleep", f.PreventIdleSleep},
		{"skill_mcp_dependency_install", f.SkillMCPDependencyInstall},
	}

	fmt.Printf("Feature flags:\n\n")
	for _, feat := range features {
		status := "default"
		if feat.enabled != nil {
			if *feat.enabled {
				status = "enabled"
			} else {
				status = "disabled"
			}
		}
		fmt.Printf("  %-35s %s\n", feat.name, status)
	}
	return nil
}

// --- codex doc ---

const codexDoc = `
Codex config.toml Reference
============================

Codex stores user-level configuration at ~/.codex/config.toml.
Project overrides go in .codex/config.toml (trusted projects only).

Config precedence (highest first):
  1. CLI flags and --config overrides
  2. Project .codex/config.toml
  3. Profile files (~/.codex/<name>.config.toml)
  4. User config (~/.codex/config.toml)
  5. System config (/etc/codex/config.toml)
  6. Built-in defaults

Key top-level settings:

  model                      Model to use (e.g. "gpt-5.5")
  model_reasoning_effort     Reasoning effort (minimal|low|medium|high|xhigh)
  model_reasoning_summary    Summary detail (auto|concise|detailed|none)
  model_verbosity            Verbosity (low|medium|high)
  default_permissions        Named permissions profile name
  web_search                 Web search mode (cached|live|disabled)
  personality                Communication style
  approval_policy            When to pause for approval
  sandbox_mode               Filesystem/network access level
  notify                     Notification commands
  log_dir                    Log output directory

Sections:

  [model_providers.<id>]     Custom model provider definition
  [mcp_servers.<id>]         MCP server (stdio or HTTP)
  [projects."<path>"]        Per-project trust level
  [plugins."<name>"]         Plugin enable/disable
  [marketplaces.<id>]        Plugin marketplace registry
  [features]                 Feature flag toggles
  [permissions.<name>]       Named permission profiles
  [hooks]                    Lifecycle hook definitions
  [agents.<name>]            Subagent definitions
  [tui]                      TUI keymap settings
  [windows]                  Windows-specific settings

For full reference: https://developers.openai.com/codex/config-reference
`

func handleCodexDoc(args []string) error {
	fmt.Print(strings.TrimPrefix(codexDoc, "\n"))
	return nil
}

// --- codex sessions ---

const codexSessionsHelp = `
Usage: agent-pro codex sessions [--limit N] [--json]

List recent Codex CLI rollout sessions.

Options:
  --limit <n>   max sessions to list (default 20, max 100)
  --json        output JSON instead of a table
  -h,--help     show help
`

const codexSessionHelp = `
Usage: agent-pro codex session <command> [ARGS]

Commands:
  info <session-id>   show detailed info for one Codex CLI session
  log  <session-id>   print human-readable session log

Run agent-pro codex session <command> --help for command-specific options.
`

const codexSessionInfoHelp = `
Usage: agent-pro codex session info <session-id> [--json]

Show detailed info for one Codex CLI rollout session.

Options:
  --json        output JSON instead of text
  -h,--help     show help
`

const codexSessionLogHelp = `
Usage: agent-pro codex session log <session-id> [--tail N]

Print the human-readable session log from a rollout JSONL file.

Options:
  --tail <n>    print last N displayable log events (0 = full log)
  -h,--help     show help
`

func handleCodexSessions(args []string) error {
	var limitFlag *int
	var jsonFlag *bool
	remaining, err := flags.Int("--limit", &limitFlag).
		Bool("--json", &jsonFlag).
		Help("-h,--help", codexSessionsHelp).
		Parse(args)
	if err != nil {
		return err
	}
	if len(remaining) > 0 {
		return fmt.Errorf("unexpected arguments: %v", remaining)
	}

	home := homeDir()
	codexHome := codexsessions.CodexHomeFromEnv(home)
	limit := 20
	if limitFlag != nil && *limitFlag > 0 {
		limit = *limitFlag
	}
	sessions, err := codexsessions.List(codexHome, limit)
	if err != nil {
		return fmt.Errorf("list codex sessions: %w", err)
	}
	if jsonFlag != nil && *jsonFlag {
		data, err := codexsessions.FormatListJSON(sessions)
		if err != nil {
			return fmt.Errorf("format codex sessions json: %w", err)
		}
		fmt.Println(string(data))
		return nil
	}
	fmt.Println(codexsessions.FormatListTable(sessions, home, time.Now()))
	return nil
}

func handleCodexSession(args []string) error {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		fmt.Print(strings.TrimPrefix(codexSessionHelp, "\n"))
		return nil
	}

	switch args[0] {
	case "info":
		return handleCodexSessionInfo(args[1:])
	case "log":
		return handleCodexSessionLog(args[1:])
	default:
		return fmt.Errorf("unknown codex session command: %s", args[0])
	}
}

func handleCodexSessionInfo(args []string) error {
	var jsonFlag *bool
	remaining, err := flags.Bool("--json", &jsonFlag).
		Help("-h,--help", codexSessionInfoHelp).
		Parse(args)
	if err != nil {
		return err
	}
	if len(remaining) != 1 {
		return fmt.Errorf("expected exactly one session id, got %d arguments", len(remaining))
	}

	sessionID := strings.TrimSpace(remaining[0])
	if sessionID == "" {
		return fmt.Errorf("session id is required")
	}

	home := homeDir()
	codexHome := codexsessions.CodexHomeFromEnv(home)
	if jsonFlag != nil && *jsonFlag {
		brief, err := codexsessions.Brief(codexHome, sessionID, 3)
		if err != nil {
			return err
		}
		data, err := codexsessions.FormatBriefJSON(brief)
		if err != nil {
			return fmt.Errorf("format codex session brief json: %w", err)
		}
		fmt.Println(string(data))
		return nil
	}

	info, err := codexsessions.Info(codexHome, sessionID, 3)
	if err != nil {
		return err
	}
	fmt.Println(codexsessions.FormatInfoText(info, home, time.Now()))
	return nil
}

func handleCodexSessionLog(args []string) error {
	var tailFlag *int
	remaining, err := flags.Int("--tail", &tailFlag).
		Help("-h,--help", codexSessionLogHelp).
		Parse(args)
	if err != nil {
		return err
	}
	if len(remaining) != 1 {
		return fmt.Errorf("expected exactly one session id, got %d arguments", len(remaining))
	}

	sessionID := strings.TrimSpace(remaining[0])
	if sessionID == "" {
		return fmt.Errorf("session id is required")
	}

	tail := 0
	if tailFlag != nil && *tailFlag > 0 {
		tail = *tailFlag
	}

	home := homeDir()
	codexHome := codexsessions.CodexHomeFromEnv(home)
	path, err := codexsessions.Find(codexHome, sessionID)
	if err != nil {
		return err
	}
	return codexsessions.PrintLog(path, os.Stdout, tail)
}

// --- codex skills ---

const codexSkillsHelp = `
Usage: agent-pro codex skills <command> [ARGS]

Commands:
  list              list installed skills
  doc               show codex skills reference documentation

Options:
  -h,--help         show help
`

func handleCodexSkills(args []string) error {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		fmt.Print(strings.TrimPrefix(codexSkillsHelp, "\n"))
		return nil
	}

	switch args[0] {
	case "list":
		return handleCodexSkillsList(args[1:])
	case "doc":
		return handleCodexSkillsDoc(args[1:])
	default:
		return fmt.Errorf("unknown codex skills command: %s", args[0])
	}
}

func handleCodexSkillsList(args []string) error {
	var dirFlag *string
	_, err := flags.String("--dir", &dirFlag).
		Help("-h,--help", `
Usage: agent-pro codex skills list [--dir DIR]

List installed skills from Codex and agent skills standard locations.

Searches:
  Global: ~/.codex/skills/, ~/.agents/skills/
  Local:  .agents/skills/ (project-specific)

Options:
  --dir <dir>   project directory for local skills (default: current directory)
  -h,--help     show help
`).
		Parse(args)
	if err != nil {
		return err
	}

	dir := "."
	if dirFlag != nil && strings.TrimSpace(*dirFlag) != "" {
		dir = strings.TrimSpace(*dirFlag)
	}
	dir, err = filepath.Abs(filepath.Clean(dir))
	if err != nil {
		return fmt.Errorf("resolve dir: %w", err)
	}

	result, err := codexskills.List(dir)
	if err != nil {
		return fmt.Errorf("list skills: %w", err)
	}

	home := homeDir()
	printCodexSkillGroup("Global", result.Global, home)
	printCodexSkillGroup("Local", result.Local, home)

	return nil
}

func printCodexSkillGroup(label string, skills []codexskills.SkillInfo, home string) {
	fmt.Printf("%s:\n", label)
	fmt.Printf("  scanned: ~/.codex/skills/, ~/.agents/skills/ for global, .agents/skills/ for local\n")

	if len(skills) == 0 {
		fmt.Println("  No skills found")
		fmt.Println()
		return
	}

	fmt.Printf("  Found %d skill(s):\n", len(skills))
	for _, s := range skills {
		loc := shortenHome(s.Path, home)
		if s.Description != "" {
			fmt.Printf("    %-30s %s\n      %s\n", s.Name, s.Description, loc)
		} else {
			fmt.Printf("    %-30s %s\n", s.Name, loc)
		}
	}
	fmt.Println()
}

const codexSkillsDoc = `
Codex Skills Reference
======================

Codex uses the open agent skills standard: https://agentskills.io/home

Skills are auto-discovered from these locations:

  ~/.codex/skills/          Codex legacy directory
  ~/.agents/skills/         Open agent skills standard (user-level)
  .agents/skills/           Open agent skills standard (project-level)

Each skill is a directory containing a SKILL.md file with YAML frontmatter.

Directory structure:

  ~/.codex/skills/
  ├── my-skill/
  │   ├── SKILL.md          # Required: name + description in YAML frontmatter
  │   ├── agents/           # Optional: openai.yaml for UI metadata
  │   ├── scripts/          # Optional: executable code
  │   ├── references/       # Optional: reference docs loaded on demand
  │   └── assets/           # Optional: template files

SKILL.md format:

  ---
  name: my-skill             # Required, kebab-case, ≤64 chars
  description: >-            # Required, explains what + when to trigger
    Description text here.
  metadata:                  # Optional
    short-description: ...   # Optional subfield
  license: MIT               # Optional
  ---
  # Markdown body (skill instructions for the AI)

Specification: https://agentskills.io/specification
Codex docs: https://developers.openai.com/codex/skills
`

func handleCodexSkillsDoc(args []string) error {
	fmt.Print(strings.TrimPrefix(codexSkillsDoc, "\n"))
	return nil
}

// --- opencode skills ---

const opencodeSkillsDoc = `
OpenCode Skills Reference
==========================

OpenCode uses the open agent skills standard: https://agentskills.io/home

Skills are auto-discovered from skill directories containing SKILL.md files.
OpenCode searches these locations (global and project-local):

  OpenCode-native:
    Global: ~/.config/opencode/skills/<name>/SKILL.md
    Local:  .opencode/skills/<name>/SKILL.md

  Claude-compatible:
    Global: ~/.claude/skills/<name>/SKILL.md
    Local:  .claude/skills/<name>/SKILL.md

  Agent-compatible:
    Global: ~/.agents/skills/<name>/SKILL.md
    Local:  .agents/skills/<name>/SKILL.md

For project-local paths, OpenCode walks up from the current working directory
to the git worktree root, loading all matching skills along the way.

SKILL.md format:

  ---
  name: git-release              # Required, must match directory name
  description: Create consistent releases...  # Required, 1-1024 chars
  license: MIT                   # Optional
  compatibility: opencode        # Optional
  metadata:                      # Optional, string-to-string map
    audience: maintainers
  ---
  # Markdown body (skill instructions for the AI)

Naming rules: ^[a-z0-9]+(-[a-z0-9]+)*$ (lowercase kebab-case), 1-64 chars,
must match the directory name.

Skills are loaded on-demand via the skill tool: skill({ name: "skill-name" }).

Specification: https://agentskills.io/specification
OpenCode docs: https://opencode.ai/docs/skills/
`

const opencodeSkillsHelp = `
Usage: agent-pro opencode skills <command> [ARGS]

Commands:
  list [--dir DIR]  list installed skills (global and local)
  doc               show opencode skills reference documentation

Options:
  --dir <dir>       project directory for local skills (default: current directory)
  -h,--help         show help
`

func handleOpenCodeSkills(args []string) error {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		fmt.Print(strings.TrimPrefix(opencodeSkillsHelp, "\n"))
		return nil
	}

	switch args[0] {
	case "list":
		return handleOpenCodeSkillsList(args[1:])
	case "doc":
		return handleOpenCodeSkillsDoc(args[1:])
	default:
		return fmt.Errorf("unknown opencode skills command: %s", args[0])
	}
}

func handleOpenCodeSkillsList(args []string) error {
	var dirFlag *string
	_, err := flags.String("--dir", &dirFlag).
		Help("-h,--help", `
Usage: agent-pro opencode skills list [--dir DIR]

List installed skills from opencode auto-discovered skill directories.

Searches:
  Global: ~/.config/opencode/skills/
  Local:  .opencode/skills/ (project-specific)

Options:
  --dir <dir>   project directory for local skills (default: current directory)
  -h,--help     show help
`).
		Parse(args)
	if err != nil {
		return err
	}

	dir := "."
	if dirFlag != nil && strings.TrimSpace(*dirFlag) != "" {
		dir = strings.TrimSpace(*dirFlag)
	}
	dir, err = filepath.Abs(filepath.Clean(dir))
	if err != nil {
		return fmt.Errorf("resolve dir: %w", err)
	}

	result, err := openskills.List(dir)
	if err != nil {
		return fmt.Errorf("list skills: %w", err)
	}

	home := homeDir()
	printSkillGroup("Global", result.Global, home)
	printSkillGroup("Local", result.Local, home)

	return nil
}

func handleOpenCodeSkillsDoc(args []string) error {
	fmt.Print(strings.TrimPrefix(opencodeSkillsDoc, "\n"))
	return nil
}

func printSkillGroup(label string, skills []openskills.SkillInfo, home string) {
	fmt.Printf("%s:\n", label)
	fmt.Printf("  scanned: .opencode/skills/, .claude/skills/, .agents/skills/\n")

	if len(skills) == 0 {
		fmt.Println("  No skills found")
		fmt.Println()
		return
	}

	fmt.Printf("  Found %d skill(s):\n", len(skills))
	for _, s := range skills {
		loc := shortenHome(s.Path, home)
		if s.Description != "" {
			fmt.Printf("    %-30s %s\n      %s\n", s.Name, s.Description, loc)
		} else {
			fmt.Printf("    %-30s %s\n", s.Name, loc)
		}
	}
	fmt.Println()
}

const opencodeSessionsHelp = `
Usage: agent-pro opencode sessions [--limit N]

List recent OpenCode CLI sessions from ~/.local/share/opencode (or $XDG_DATA_HOME/opencode).

Options:
  --limit <n>   max sessions to list (default 20, max 100)
  -h,--help     show help
`

const opencodeSessionHelp = `
Usage: agent-pro opencode session <command> [ARGS]

Commands:
  info <session-id>   show detailed info for one OpenCode CLI session

Run agent-pro opencode session <command> --help for command-specific options.
`

const opencodeSessionInfoHelp = `
Usage: agent-pro opencode session info <session-id>

Show detailed info for one OpenCode CLI session.

Options:
  -h,--help     show help
`

func handleOpenCodeSessions(args []string) error {
	var limitFlag *int
	remaining, err := flags.Int("--limit", &limitFlag).
		Help("-h,--help", opencodeSessionsHelp).
		Parse(args)
	if err != nil {
		return err
	}
	if len(remaining) > 0 {
		return fmt.Errorf("unexpected arguments: %v", remaining)
	}

	home := homeDir()
	dataDir := opencodesessions.OpenCodeDataHome(home)
	limit := 20
	if limitFlag != nil && *limitFlag > 0 {
		limit = *limitFlag
	}
	sessions, err := opencodesessions.List(dataDir, limit)
	if err != nil {
		return fmt.Errorf("list opencode sessions: %w", err)
	}
	fmt.Println(opencodesessions.FormatListTable(sessions, home, time.Now()))
	return nil
}

func handleOpenCodeSession(args []string) error {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		fmt.Print(strings.TrimPrefix(opencodeSessionHelp, "\n"))
		return nil
	}

	switch args[0] {
	case "info":
		return handleOpenCodeSessionInfo(args[1:])
	default:
		return fmt.Errorf("unknown opencode session command: %s", args[0])
	}
}

func handleOpenCodeSessionInfo(args []string) error {
	remaining, err := flags.New().
		Help("-h,--help", opencodeSessionInfoHelp).
		Parse(args)
	if err != nil {
		return err
	}
	if len(remaining) != 1 {
		return fmt.Errorf("expected exactly one session id, got %d arguments", len(remaining))
	}

	sessionID := strings.TrimSpace(remaining[0])
	if sessionID == "" {
		return fmt.Errorf("session id is required")
	}

	home := homeDir()
	dataDir := opencodesessions.OpenCodeDataHome(home)
	info, err := opencodesessions.Info(dataDir, sessionID)
	if err != nil {
		return err
	}
	fmt.Println(opencodesessions.FormatInfoText(info, home, time.Now()))
	return nil
}

// --- opencode config ---

const opencodeConfigHelp = `
Usage: agent-pro opencode config <command> [ARGS]

Commands:
  export <file.zip>     export opencode configuration to a zip file
  import <file.zip>     import opencode configuration from a zip file
  add-provider          add a custom provider to the opencode config (v1 format)

Options:
  -h,--help             show help
`

func handleOpenCodeConfig(args []string) error {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		fmt.Print(strings.TrimPrefix(opencodeConfigHelp, "\n"))
		return nil
	}
	switch args[0] {
	case "export":
		return handleOpenCodeConfigExport(args[1:])
	case "import":
		return handleOpenCodeConfigImport(args[1:])
	case "add-provider":
		return handleOpenCodeConfigAddProvider(args[1:])
	default:
		return fmt.Errorf("unknown opencode config command: %s", args[0])
	}
}

func handleOpenCodeConfigExport(args []string) error {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		fmt.Print(strings.TrimPrefix(opencodeConfigHelp, "\n"))
		return nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("find home dir: %w", err)
	}
	zipPath := args[0]
	return agentconfig.Export("opencode", home, zipPath)
}

func handleOpenCodeConfigImport(args []string) error {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		fmt.Print(strings.TrimPrefix(opencodeConfigHelp, "\n"))
		return nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("find home dir: %w", err)
	}
	zipPath := args[0]
	return agentconfig.Import(home, zipPath)
}

func handleOpenCodeConfigAddProvider(args []string) error {
	var id, baseURL, apiShape, name, dir, apiKey string
	var models []string
	_, err := flags.String("--id", &id).
		String("--base-url", &baseURL).
		String("--api-shape", &apiShape).
		StringSlice("--model", &models).
		String("--name", &name).
		String("--dir", &dir).
		String("--api-key", &apiKey).
		Help("-h,--help", `
Usage: agent-pro opencode config add-provider [OPTIONS]

Add a custom provider entry (v1 format) to the opencode config file under the
top-level "provider" key, mirroring opencode's connect custom-provider flow.

Options:
  --id <id>            provider id (required)
  --base-url <url>     base URL written to options.baseURL (required)
  --api-shape <shape>  api shape: anthropic or openai (required)
  --model <id>         model id; repeatable, at least one required
  --name <name>        provider display name (default: --id)
  --api-key <key>      API key written inline to options.apiKey (optional)
  --dir <project-dir>  project dir for local config (default: global ~/.config/opencode)
  -h,--help            show help

Note: --base-url should be the full API base including the version path (e.g.
/v1), because opencode appends only the endpoint path (/messages for anthropic,
/chat/completions for openai-compatible). Example: --base-url http://127.0.0.1:15721/v1
`).
		Parse(args)
	if err != nil {
		return err
	}

	if id == "" {
		return fmt.Errorf("--id is required")
	}
	if baseURL == "" {
		return fmt.Errorf("--base-url is required")
	}
	if apiShape == "" {
		return fmt.Errorf("--api-shape is required")
	}
	var npm string
	switch apiShape {
	case "anthropic":
		npm = "@ai-sdk/anthropic"
	case "openai":
		npm = "@ai-sdk/openai-compatible"
	default:
		return fmt.Errorf("--api-shape %q is not valid; valid values are: anthropic, openai", apiShape)
	}
	if len(models) == 0 {
		return fmt.Errorf("at least one --model is required")
	}

	displayName := name
	if displayName == "" {
		displayName = id
	}

	var opencodeDir string
	if dir != "" {
		opencodeDir = filepath.Join(dir, ".opencode")
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("find home dir: %w", err)
		}
		opencodeDir = filepath.Join(home, ".config", "opencode")
	}

	cfg, err := opencodecfg.ReadDir(opencodeDir)
	if err != nil {
		return err
	}
	data := cfg.Data
	if data == nil {
		data = opencodecfg.Data{}
	}

	providers, _ := data["provider"].(map[string]interface{})
	if providers == nil {
		providers = map[string]interface{}{}
	}
	if _, exists := providers[id]; exists {
		return fmt.Errorf("provider %q already exists in %s", id, cfg.Path)
	}

	modelsMap := make(map[string]interface{}, len(models))
	for _, m := range models {
		modelsMap[m] = map[string]interface{}{"name": m}
	}

	options := map[string]interface{}{"baseURL": baseURL}
	if apiKey != "" {
		options["apiKey"] = apiKey
	}

	providers[id] = map[string]interface{}{
		"npm":     npm,
		"name":    displayName,
		"options": options,
		"models":  modelsMap,
	}
	data["provider"] = providers
	cfg.Data = data

	if err := cfg.Write(); err != nil {
		return err
	}

	fmt.Printf("Added provider %s to %s\n", id, cfg.Path)
	return nil
}

// --- pi ---

const piHelp = `
Usage: agent-pro pi <command> [ARGS]

Commands:
  config            manage pi configuration (export/import)

Run agent-pro pi <command> --help for command-specific options.
`

func handlePi(args []string) error {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		fmt.Print(strings.TrimPrefix(piHelp, "\n"))
		return nil
	}
	switch args[0] {
	case "config":
		return handlePiConfig(args[1:])
	default:
		return fmt.Errorf("unknown pi command: %s", args[0])
	}
}

const piConfigHelp = `
Usage: agent-pro pi config <command> [ARGS]

Commands:
  export <file.zip>   export pi configuration to a zip file
  import <file.zip>   import pi configuration from a zip file

Options:
  -h,--help           show help
`

func handlePiConfig(args []string) error {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		fmt.Print(strings.TrimPrefix(piConfigHelp, "\n"))
		return nil
	}
	switch args[0] {
	case "export":
		return handlePiConfigExport(args[1:])
	case "import":
		return handlePiConfigImport(args[1:])
	default:
		return fmt.Errorf("unknown pi config command: %s", args[0])
	}
}

func handlePiConfigExport(args []string) error {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		fmt.Print(strings.TrimPrefix(piConfigHelp, "\n"))
		return nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("find home dir: %w", err)
	}
	zipPath := args[0]
	return agentconfig.Export("pi", home, zipPath)
}

func handlePiConfigImport(args []string) error {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		fmt.Print(strings.TrimPrefix(piConfigHelp, "\n"))
		return nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("find home dir: %w", err)
	}
	zipPath := args[0]
	return agentconfig.Import(home, zipPath)
}

// --- crush ---

const crushHelp = `
Usage: agent-pro crush <command> [ARGS]

Commands:
  config            manage crush configuration (export/import)

Run agent-pro crush <command> --help for command-specific options.
`

func handleCrush(args []string) error {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		fmt.Print(strings.TrimPrefix(crushHelp, "\n"))
		return nil
	}
	switch args[0] {
	case "config":
		return handleCrushConfig(args[1:])
	default:
		return fmt.Errorf("unknown crush command: %s", args[0])
	}
}

const crushConfigHelp = `
Usage: agent-pro crush config <command> [ARGS]

Commands:
  export <file.zip>   export crush configuration to a zip file
  import <file.zip>   import crush configuration from a zip file

Options:
  -h,--help           show help
`

func handleCrushConfig(args []string) error {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		fmt.Print(strings.TrimPrefix(crushConfigHelp, "\n"))
		return nil
	}
	switch args[0] {
	case "export":
		return handleCrushConfigExport(args[1:])
	case "import":
		return handleCrushConfigImport(args[1:])
	default:
		return fmt.Errorf("unknown crush config command: %s", args[0])
	}
}

func handleCrushConfigExport(args []string) error {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		fmt.Print(strings.TrimPrefix(crushConfigHelp, "\n"))
		return nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("find home dir: %w", err)
	}
	zipPath := args[0]
	return agentconfig.Export("crush", home, zipPath)
}

func handleCrushConfigImport(args []string) error {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		fmt.Print(strings.TrimPrefix(crushConfigHelp, "\n"))
		return nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("find home dir: %w", err)
	}
	zipPath := args[0]
	return agentconfig.Import(home, zipPath)
}

// --- show-agent-files ---

const showAgentFilesHelp = `
Usage: agent-pro show-agent-files [--dir DIR]

Collect symlinks to well-known agent directories from $HOME into a target
directory. Default target: ~/.agent-pro/agent-files-collection/

The default target is always wiped and recreated. A custom target
specified with --dir must either not exist or be empty.

Agent directories scanned:

  ~/.codex             OpenAI Codex CLI
  ~/.claude            Claude Code (Anthropic)
  ~/.config/opencode   OpenCode
  ~/.agents            Agent Skills Standard
  ~/.gemini            Gemini CLI
  ~/.config/gemini-cli Gemini CLI
  ~/.cursor            Cursor Editor

Only existing directories are linked.

Options:
  --dir <dir>  target directory (default: ~/.agent-pro/agent-files-collection/)
  -h,--help    show help
`

func handleShowAgentFiles(args []string) error {
	var dirFlag *string
	_, err := flags.String("--dir", &dirFlag).
		Help("-h,--help", showAgentFilesHelp).
		Parse(args)
	if err != nil {
		return err
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("find home dir: %w", err)
	}

	useDefault := true
	targetDir := filepath.Join(homeDir, ".agent-pro", "agent-files-collection")
	if dirFlag != nil && strings.TrimSpace(*dirFlag) != "" {
		targetDir = *dirFlag
		targetDir, err = filepath.Abs(filepath.Clean(targetDir))
		if err != nil {
			return fmt.Errorf("resolve --dir: %w", err)
		}
		useDefault = false
	}

	if useDefault {
		if err := os.RemoveAll(targetDir); err != nil {
			return fmt.Errorf("wipe target dir: %w", err)
		}
	} else {
		if err := ValidateTargetDir(targetDir); err != nil {
			return err
		}
	}

	fmt.Printf("Collecting agent files into %s\n\n", shortenHome(targetDir, homeDir))
	if err := CollectAgentFiles(homeDir, targetDir); err != nil {
		return err
	}

	fmt.Printf("\nDone. View with: ls -laR %s\n", shortenHome(targetDir, homeDir))
	return nil
}

// --- traces ---

func handleTraces(args []string) error {
	return run.Run(args)
}

// --- session bookmarks (multi-runner catalog) ---

const bookmarkHelp = `
Usage: agent-pro bookmark [list|show|remove] [ARGS]

Manage the multi-runner session bookmark catalog stored under
$AGENT_PRO_HOME/session_bookmarks.json (default: ~/.agent-pro/).

Commands:
  list                  list bookmarks (default when no subcommand)
  show   <session-id>   show one bookmark
  remove <session-id>   remove a bookmark (aliases: rm, unbookmark)

Options (list):
  --runner <name>   filter by agent runner (e.g. grok, codex)
  -t,--tag <tag>    AND filter: bookmark must include all tags (repeatable)
  --limit <n>       max rows (0 = unlimited)
  --stale           catalog snapshot only (no FS refresh / orphan checks)
  --enrich          slow: walk GROK_HOME via Find when session_dir is stale
  --json            print JSON (no ANSI)
  -h,--help         show help

Options (show / remove):
  --runner <name>   disambiguate when the same session id is pinned under
                    multiple runners
  --stale           catalog snapshot only (show; same as list --stale)
  --enrich          slow Find recovery when session_dir is stale (show)
  --json            print JSON (show/remove ack; no ANSI)
  -h,--help         show help

Enrich modes (list/show; --stale and --enrich are mutually exclusive):
  default           cheap refresh from stored session_dir/summary.json only
  --stale           no live FS checks; Orphaned not computed
  --enrich          light first, then Find under GROK_HOME (slow on large trees)

Grok pin aliases:
  agent-pro grok session bookmark <id> …
  agent-pro grok session bookmarks
  agent-pro grok session unbookmark <id>
`

const grokSessionBookmarkHelp = `
Usage: agent-pro grok session bookmark <session-id> [OPTIONS]

Pin a Grok CLI session into the multi-runner bookmark catalog
($AGENT_PRO_HOME/session_bookmarks.json).

Options:
  -t,--tag <tag>         add/merge tag (repeatable); sorted unique on write
  -d,--description TEXT  set description (omit to keep on update)
  --clear-tags           wipe existing tags before merging any --tag values
  --json                 print bookmark as JSON (no ANSI)
  -h,--help              show help
`

func handleBookmark(args []string) error {
	if len(args) == 0 {
		return handleBookmarkList(nil)
	}
	if args[0] == "-h" || args[0] == "--help" {
		fmt.Print(strings.TrimPrefix(bookmarkHelp, "\n"))
		return nil
	}

	switch args[0] {
	case "list", "ls":
		return handleBookmarkList(args[1:])
	case "show":
		return handleBookmarkShow(args[1:])
	case "remove", "rm", "unbookmark":
		return handleBookmarkRemove(args[1:])
	default:
		// Bare flags after "bookmark" → list (e.g. bookmark --json --runner grok).
		if strings.HasPrefix(args[0], "-") {
			return handleBookmarkList(args)
		}
		return fmt.Errorf("unknown bookmark command: %s", args[0])
	}
}

func handleBookmarkList(args []string) error {
	var runner string
	var tags []string
	var limit int
	var staleFlag *bool
	var enrichFlag *bool
	var jsonFlag *bool
	remaining, err := flags.String("--runner", &runner).
		StringSlice("-t,--tag", &tags).
		Int("--limit", &limit).
		Bool("--stale", &staleFlag).
		Bool("--enrich", &enrichFlag).
		Bool("--json", &jsonFlag).
		Help("-h,--help", bookmarkHelp).
		Parse(args)
	if err != nil {
		return err
	}
	if len(remaining) > 0 {
		return fmt.Errorf("unexpected arguments: %s", strings.Join(remaining, " "))
	}

	mode, err := resolveBookmarkEnrichMode(staleFlag, enrichFlag)
	if err != nil {
		return err
	}

	agentHome := resolveAgentProHome()
	grokHome := agenttty.GrokHome()
	views, warnings, err := groksessions.ListBookmarks(agentHome, grokHome, groksessions.ListFilter{
		Runner: runner,
		Tags:   tags,
		Limit:  limit,
		Enrich: mode,
	})
	if err != nil {
		return err
	}
	for _, w := range warnings {
		fmt.Fprintf(os.Stderr, "warning: %s\n", w)
	}

	if jsonFlag != nil && *jsonFlag {
		out, err := groksessions.FormatBookmarkJSON(views)
		if err != nil {
			return fmt.Errorf("format bookmarks json: %w", err)
		}
		fmt.Println(out)
		return nil
	}

	fmt.Println(groksessions.FormatBookmarksTable(views))
	return nil
}

func handleBookmarkShow(args []string) error {
	var runner string
	var staleFlag *bool
	var enrichFlag *bool
	var jsonFlag *bool
	remaining, err := flags.String("--runner", &runner).
		Bool("--stale", &staleFlag).
		Bool("--enrich", &enrichFlag).
		Bool("--json", &jsonFlag).
		Help("-h,--help", bookmarkHelp).
		Parse(args)
	if err != nil {
		return err
	}
	if len(remaining) != 1 {
		return fmt.Errorf("expected exactly one session id, got %d arguments", len(remaining))
	}
	sessionID := strings.TrimSpace(remaining[0])
	if sessionID == "" {
		return fmt.Errorf("session id is required")
	}

	mode, err := resolveBookmarkEnrichMode(staleFlag, enrichFlag)
	if err != nil {
		return err
	}

	agentHome := resolveAgentProHome()
	grokHome := agenttty.GrokHome()
	view, err := groksessions.GetBookmark(agentHome, runner, sessionID, grokHome, mode)
	if err != nil {
		return err
	}

	if jsonFlag != nil && *jsonFlag {
		out, err := groksessions.FormatBookmarkJSON(view)
		if err != nil {
			return fmt.Errorf("format bookmark json: %w", err)
		}
		fmt.Println(out)
		return nil
	}

	fmt.Println(groksessions.FormatBookmarkShow(view))
	return nil
}

// resolveBookmarkEnrichMode maps --stale / --enrich CLI flags to EnrichMode.
// Default (neither) is EnrichLight. Both set is an error.
func resolveBookmarkEnrichMode(staleFlag, enrichFlag *bool) (groksessions.EnrichMode, error) {
	stale := staleFlag != nil && *staleFlag
	enrich := enrichFlag != nil && *enrichFlag
	if stale && enrich {
		return 0, fmt.Errorf("--stale and --enrich are mutually exclusive")
	}
	if stale {
		return groksessions.EnrichOff, nil
	}
	if enrich {
		return groksessions.EnrichHeavy, nil
	}
	return groksessions.EnrichLight, nil
}

func handleBookmarkRemove(args []string) error {
	var runner string
	var jsonFlag *bool
	remaining, err := flags.String("--runner", &runner).
		Bool("--json", &jsonFlag).
		Help("-h,--help", bookmarkHelp).
		Parse(args)
	if err != nil {
		return err
	}
	if len(remaining) != 1 {
		return fmt.Errorf("expected exactly one session id, got %d arguments", len(remaining))
	}
	sessionID := strings.TrimSpace(remaining[0])
	if sessionID == "" {
		return fmt.Errorf("session id is required")
	}

	agentHome := resolveAgentProHome()
	if err := groksessions.RemoveBookmark(agentHome, runner, sessionID); err != nil {
		return err
	}

	if jsonFlag != nil && *jsonFlag {
		out, err := groksessions.FormatBookmarkJSON(map[string]any{
			"removed":    true,
			"session_id": sessionID,
			"runner":     runner,
		})
		if err != nil {
			return err
		}
		fmt.Println(out)
		return nil
	}

	if runner != "" {
		fmt.Printf("Removed bookmark %s (runner %s)\n", sessionID, runner)
	} else {
		fmt.Printf("Removed bookmark %s\n", sessionID)
	}
	return nil
}

func handleGrokSessionBookmark(args []string) error {
	var tagsFlag *[]string
	var descFlag *string
	var clearTags bool
	var jsonFlag *bool
	remaining, err := flags.StringSlice("-t,--tag", &tagsFlag).
		String("-d,--description", &descFlag).
		Bool("--clear-tags", &clearTags).
		Bool("--json", &jsonFlag).
		Help("-h,--help", grokSessionBookmarkHelp).
		Parse(args)
	if err != nil {
		return err
	}
	if len(remaining) != 1 {
		return fmt.Errorf("expected exactly one session id, got %d arguments", len(remaining))
	}
	sessionID := strings.TrimSpace(remaining[0])
	if sessionID == "" {
		return fmt.Errorf("session id is required")
	}

	opts := &groksessions.PinOptions{
		Description: descFlag,
		ClearTags:   clearTags,
	}
	if tagsFlag != nil {
		opts.Tags = *tagsFlag
	}

	agentHome := resolveAgentProHome()
	grokHome := agenttty.GrokHome()
	bm, created, err := groksessions.BookmarkGrok(agentHome, grokHome, sessionID, opts)
	if err != nil {
		return err
	}

	if jsonFlag != nil && *jsonFlag {
		out, err := groksessions.FormatBookmarkJSON(bm)
		if err != nil {
			return fmt.Errorf("format bookmark json: %w", err)
		}
		fmt.Println(out)
		return nil
	}

	if created {
		fmt.Printf("Bookmarked session %s\n", sessionID)
	} else {
		fmt.Printf("Updated bookmark %s\n", sessionID)
	}
	if bm != nil {
		if len(bm.Tags) > 0 {
			fmt.Printf("  tags: %s\n", strings.Join(bm.Tags, ", "))
		}
		if bm.Description != "" {
			fmt.Printf("  description: %s\n", bm.Description)
		}
		if bm.Title != "" {
			fmt.Printf("  title: %s\n", bm.Title)
		}
	}
	return nil
}

// resolveAgentProHome returns AGENT_PRO_HOME if set, else ~/.agent-pro.
// CLI reads env once and passes the path into package APIs (no library Setenv).
func resolveAgentProHome() string {
	if v := strings.TrimSpace(os.Getenv("AGENT_PRO_HOME")); v != "" {
		return v
	}
	return filepath.Join(homeDir(), ".agent-pro")
}

// --- helpers ---

func homeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "~"
	}
	return home
}

func truncate(s string, maxLen int) string {
	firstLine := s
	if idx := strings.IndexByte(s, '\n'); idx >= 0 {
		firstLine = s[:idx]
	}
	if len(firstLine) > maxLen {
		return firstLine[:maxLen] + "..."
	}
	return firstLine
}

type projectEntry struct {
	Path       string
	TrustLevel string
}

func sortedProjects(m map[string]codexcfg.Project) []projectEntry {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var result []projectEntry
	for _, k := range keys {
		result = append(result, projectEntry{Path: k, TrustLevel: m[k].TrustLevel})
	}
	return result
}

func sortedModelProviders(m map[string]codexcfg.ModelProvider) []struct {
	ID string
	codexcfg.ModelProvider
} {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var result []struct {
		ID string
		codexcfg.ModelProvider
	}
	for _, k := range keys {
		result = append(result, struct {
			ID string
			codexcfg.ModelProvider
		}{ID: k, ModelProvider: m[k]})
	}
	return result
}

func sortedMCPServers(m map[string]codexcfg.MCPServer) []struct {
	ID string
	codexcfg.MCPServer
} {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var result []struct {
		ID string
		codexcfg.MCPServer
	}
	for _, k := range keys {
		result = append(result, struct {
			ID string
			codexcfg.MCPServer
		}{ID: k, MCPServer: m[k]})
	}
	return result
}

func sortedPlugins(m map[string]codexcfg.Plugin) []struct {
	ID string
	codexcfg.Plugin
} {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var result []struct {
		ID string
		codexcfg.Plugin
	}
	for _, k := range keys {
		result = append(result, struct {
			ID string
			codexcfg.Plugin
		}{ID: k, Plugin: m[k]})
	}
	return result
}
