package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	codexcfg "github.com/xhd2015/agent-pro/agent/codex/config"
	codexsessions "github.com/xhd2015/agent-pro/agent/codex/sessions"
	codexskills "github.com/xhd2015/agent-pro/agent/codex/skills"
	"github.com/xhd2015/agent-pro/agent/opencode/commands"
	opencodecfg "github.com/xhd2015/agent-pro/agent/opencode/config"
	openskills "github.com/xhd2015/agent-pro/agent/opencode/skills"
	"github.com/xhd2015/agent-pro/agent/opencode/permissions"
	"github.com/xhd2015/agent-pro/agent/opencode/plugins"
	"github.com/xhd2015/agent-pro/frontend"
	"github.com/xhd2015/agent-pro/pkgs/agentconfig"
	"github.com/xhd2015/agent-pro/run"
	"github.com/xhd2015/agent-pro/server"
	"github.com/xhd2015/less-gen/flags"
)

const help = `
Usage: agent-pro <command> [ARGS]

Commands:
  opencode          manage opencode hooks, permissions, and config
  pi                manage pi configuration
  crush             manage crush configuration
  codex             manage codex configuration
  skills            list available skills (explore, reproduce)
  skill             show or install a skill
  traces            view agent trace sessions (web viewer)
  show-agent-files  collect known agent files under ~/.agent-pro/agent-files-collection/

Run agent-pro <command> --help for command-specific options.
`

const opencodeHelp = `
Usage: agent-pro opencode <command> [ARGS]

Commands:
  commands          list opencode slash commands
  config            manage opencode configuration (export/import)
  permissions       manage opencode permissions
  plugins           manage opencode plugins
  skills            list installed skills

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
	case "pi":
		return handlePi(args[1:])
	case "crush":
		return handleCrush(args[1:])
	case "skill":
		return handleSkill(args[1:])
	case "skills":
		return handleSkills(args[1:])
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
  sessions          list, brief, or log Codex CLI rollout sessions
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
Usage: agent-pro codex sessions [SESSION-ID] [OPTIONS]

List, briefly show, or print full logs for Codex CLI rollout sessions.

Modes:
  agent-pro codex sessions [--limit N] [--json]
      List recent sessions (default limit 20, max 100).

  agent-pro codex sessions <session-id> [--json]
      Show a brief summary with the last few displayable messages.

  agent-pro codex sessions <session-id> --log
      Print the full human-readable session log.

  agent-pro codex sessions <session-id> --tail <n>
      Print the last N displayable log events (--log is implied).

Options:
  --limit <n>   max sessions to list (default 20)
  --json        output JSON instead of a table or brief text
  --log         print full session log (requires session id)
  --tail <n>    print last N displayable log events (implies --log)
  -h,--help     show help
`

func handleCodexSessions(args []string) error {
	var limitFlag *int
	var jsonFlag *bool
	var logFlag *bool
	var tailFlag *int
	remaining, err := flags.Int("--limit", &limitFlag).
		Bool("--json", &jsonFlag).
		Bool("--log", &logFlag).
		Int("--tail", &tailFlag).
		Help("-h,--help", codexSessionsHelp).
		Parse(args)
	if err != nil {
		return err
	}

	home := homeDir()
	codexHome := codexsessions.CodexHomeFromEnv(home)
	useJSON := jsonFlag != nil && *jsonFlag
	tail := 0
	if tailFlag != nil && *tailFlag > 0 {
		tail = *tailFlag
	}
	useLog := (logFlag != nil && *logFlag) || tail > 0

	if len(remaining) == 0 {
		limit := 20
		if limitFlag != nil && *limitFlag > 0 {
			limit = *limitFlag
		}
		sessions, err := codexsessions.List(codexHome, limit)
		if err != nil {
			return fmt.Errorf("list codex sessions: %w", err)
		}
		if useJSON {
			data, err := codexsessions.FormatListJSON(sessions)
			if err != nil {
				return fmt.Errorf("format codex sessions json: %w", err)
			}
			fmt.Println(string(data))
			return nil
		}
		fmt.Println(codexsessions.FormatListTable(sessions, home))
		return nil
	}

	if len(remaining) != 1 {
		return fmt.Errorf("expected at most one session id, got %d arguments", len(remaining))
	}
	sessionID := strings.TrimSpace(remaining[0])
	if sessionID == "" {
		return fmt.Errorf("session id is required")
	}

	if useLog {
		if useJSON {
			return fmt.Errorf("--tail and --log cannot be used with --json")
		}
		path, err := codexsessions.Find(codexHome, sessionID)
		if err != nil {
			return err
		}
		return codexsessions.PrintLog(path, os.Stdout, tail)
	}

	brief, err := codexsessions.Brief(codexHome, sessionID, 3)
	if err != nil {
		return err
	}
	if useJSON {
		data, err := codexsessions.FormatBriefJSON(brief)
		if err != nil {
			return fmt.Errorf("format codex session brief json: %w", err)
		}
		fmt.Println(string(data))
		return nil
	}
	fmt.Println(codexsessions.FormatBriefText(brief, home))
	return nil
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
	var id, baseURL, apiShape, name, dir string
	var models []string
	_, err := flags.String("--id", &id).
		String("--base-url", &baseURL).
		String("--api-shape", &apiShape).
		StringSlice("--model", &models).
		String("--name", &name).
		String("--dir", &dir).
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
  --dir <project-dir>  project dir for local config (default: global ~/.config/opencode)
  -h,--help            show help
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

	providers[id] = map[string]interface{}{
		"npm":    npm,
		"name":   displayName,
		"options": map[string]interface{}{"baseURL": baseURL},
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
