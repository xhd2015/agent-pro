package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tidwall/jsonc"
	"github.com/xhd2015/agent-pro/agent/opencode/commands"
	"github.com/xhd2015/agent-pro/agent/opencode/config"
	"github.com/xhd2015/agent-pro/agent/opencode/permissions"
	"github.com/xhd2015/agent-pro/agent/opencode/plugins"
	"github.com/xhd2015/less-gen/flags"
)

const help = `
Usage: agent-pro <command> [ARGS]

Commands:
  opencode          manage opencode hooks and permissions

Run agent-pro <command> --help for command-specific options.
`

const opencodeHelp = `
Usage: agent-pro opencode <command> [ARGS]

Commands:
  commands          list opencode slash commands
  permissions       manage opencode permissions
  plugins           manage opencode plugins

Run agent-pro opencode <command> --help for command-specific options.
`

func main() {
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

	printCommandsLocation("Global", filepath.Join(home, ".config", "opencode"), dir)
	printCommandsLocation("Local", filepath.Join(dir, ".opencode"), dir)

	return nil
}

func printCommandsLocation(label, opencodeDir, baseDir string) {
	fmt.Printf("%s (%s):\n", label, opencodeDir)
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
		if c.Path != "" {
			rel, _ := filepath.Rel(baseDir, c.Path)
			fmt.Printf("    %-35s %s\n      file: %s  source: %s\n", c.Name, c.Description, rel, c.Source)
		} else {
			fmt.Printf("    %-35s %s  source: %s\n", c.Name, c.Description, c.Source)
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

	printPluginLocation("Global", filepath.Join(home, ".config", "opencode"), dir)
	printPluginLocation("Local", filepath.Join(dir, ".opencode"), dir)

	return nil
}

func printPluginLocation(label, opencodeDir, baseDir string) {
	fmt.Printf("%s (%s):\n", label, opencodeDir)
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
		rel, _ := filepath.Rel(baseDir, p.Path)
		if rel == "" {
			rel = p.Path
		}
		fmt.Printf("    %s  (%s)\n", p.Name, rel)
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

	printPermissionLocation("Global", filepath.Join(home, ".config", "opencode"))
	printPermissionLocation("Local", filepath.Join(dir, ".opencode"))

	return nil
}

func printPermissionLocation(label, opencodeDir string) {
	fmt.Printf("%s (%s):\n", label, opencodeDir)
	fmt.Printf("  scanned: opencode.jsonc, opencode.json\n")

	cfg, err := readConfigDir(opencodeDir)
	if err != nil {
		fmt.Printf("  (error: %v)\n\n", err)
		return
	}

	printPermissionRules(cfg)
}

func readConfigDir(opencodeDir string) (*config.Config, error) {
	var configPath string
	var raw []byte
	for _, name := range []string{"opencode.jsonc", "opencode.json"} {
		p := filepath.Join(opencodeDir, name)
		data, err := os.ReadFile(p)
		if err == nil {
			configPath = p
			raw = data
			break
		}
	}
	if configPath == "" {
		configPath = filepath.Join(opencodeDir, "opencode.json")
		return &config.Config{Path: configPath, Data: config.Data{}}, nil
	}

	cleaned := jsonc.ToJSON(raw)
	var data config.Data
	if err := json.Unmarshal(cleaned, &data); err != nil {
		return nil, fmt.Errorf("parse %s: %w", configPath, err)
	}
	return &config.Config{Path: configPath, Data: data}, nil
}

func printPermissionRules(cfg *config.Config) {
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

	keys := config.SortedKeys(obj)
	if len(keys) == 0 {
		fmt.Println("  No deny rules configured")
		return
	}

	fmt.Printf("  Found %d rule(s):\n", len(keys))
	for _, key := range keys {
		action, _ := obj[key].(string)
		fmt.Printf("    %-40s %s\n", key, action)
	}
	fmt.Println()
}
