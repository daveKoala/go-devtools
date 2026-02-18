package cli

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"go-devtools/internal/modules"
)

func Run(args []string, stdout io.Writer, stderr io.Writer, tools []modules.Tool, runTUI func() error) error {
	if len(args) == 0 {
		return runTUI()
	}

	switch args[0] {
	case "tui":
		return runTUI()
	case "help", "--help", "-h":
		return printHelp(stdout, tools, args[1:])
	case "list":
		return printList(stdout, tools)
	case "run":
		return runAction(stdout, stderr, tools, args[1:])
	default:
		// Shortcut form: devtools <module-id> <action-id> [args...]
		return runAction(stdout, stderr, tools, args)
	}
}

func printHelp(stdout io.Writer, tools []modules.Tool, topic []string) error {
	if len(topic) == 0 {
		fmt.Fprintln(stdout, "Developer Tools CLI")
		fmt.Fprintln(stdout, "")
		fmt.Fprintln(stdout, "Usage:")
		fmt.Fprintln(stdout, "  devtools                      Launch interactive TUI")
		fmt.Fprintln(stdout, "  devtools tui                  Launch interactive TUI")
		fmt.Fprintln(stdout, "  devtools list                 List modules and actions")
		fmt.Fprintln(stdout, "  devtools help                 Show this help")
		fmt.Fprintln(stdout, "  devtools help <module-id>     Show module actions")
		fmt.Fprintln(stdout, "  devtools help <module-id> <action-id>")
		fmt.Fprintln(stdout, "  devtools run <module-id> <action-id> [--key value|--key=value|key=value]")
		fmt.Fprintln(stdout, "  devtools <module-id> <action-id> [args]   Shortcut for run")
		fmt.Fprintln(stdout, "")
		fmt.Fprintln(stdout, "Examples:")
		fmt.Fprintln(stdout, "  devtools run chuck-norris-facts random-fact")
		fmt.Fprintln(stdout, "  devtools run auth-token-generator userpass-token --username alice --password secret")
		fmt.Fprintln(stdout, "  devtools help auth-token-generator google-token")
		fmt.Fprintln(stdout, "")
		return printList(stdout, tools)
	}

	tool, ok := modules.FindTool(tools, topic[0])
	if !ok {
		return fmt.Errorf("unknown module %q", topic[0])
	}

	if len(topic) == 1 {
		fmt.Fprintf(stdout, "Module: %s (%s)\n", tool.Label(), tool.ID())
		fmt.Fprintf(stdout, "%s\n\n", tool.Description())
		return printModuleActions(stdout, tool)
	}

	action, ok := modules.FindAction(tool, topic[1])
	if !ok {
		return fmt.Errorf("unknown action %q for module %q", topic[1], topic[0])
	}

	fmt.Fprintf(stdout, "Module: %s (%s)\n", tool.Label(), tool.ID())
	fmt.Fprintf(stdout, "Action: %s (%s)\n", action.Label, action.ID)
	fmt.Fprintf(stdout, "Description: %s\n", action.Description)
	if action.Usage != "" {
		fmt.Fprintf(stdout, "Usage: %s\n", action.Usage)
	}
	return nil
}

func printList(stdout io.Writer, tools []modules.Tool) error {
	fmt.Fprintln(stdout, "Available modules and actions:")
	for _, tool := range tools {
		fmt.Fprintf(stdout, "- %s (%s)\n", tool.Label(), tool.ID())
		actions := tool.Actions()
		sort.Slice(actions, func(i, j int) bool { return actions[i].ID < actions[j].ID })
		for _, action := range actions {
			fmt.Fprintf(stdout, "  - %s: %s\n", action.ID, action.Description)
		}
	}
	return nil
}

func printModuleActions(stdout io.Writer, tool modules.Tool) error {
	fmt.Fprintln(stdout, "Actions:")
	actions := tool.Actions()
	sort.Slice(actions, func(i, j int) bool { return actions[i].ID < actions[j].ID })
	for _, action := range actions {
		fmt.Fprintf(stdout, "- %s: %s\n", action.ID, action.Description)
		if action.Usage != "" {
			fmt.Fprintf(stdout, "  usage: %s\n", action.Usage)
		}
	}
	return nil
}

func runAction(stdout io.Writer, stderr io.Writer, tools []modules.Tool, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: devtools run <module-id> <action-id> [--key value|--key=value|key=value]")
	}

	moduleID := args[0]
	actionID := args[1]
	rest := args[2:]

	tool, ok := modules.FindTool(tools, moduleID)
	if !ok {
		return fmt.Errorf("unknown module %q", moduleID)
	}

	action, ok := modules.FindAction(tool, actionID)
	if !ok {
		return fmt.Errorf("unknown action %q for module %q", actionID, moduleID)
	}

	if hasHelpFlag(rest) {
		fmt.Fprintf(stdout, "Module: %s (%s)\n", tool.Label(), tool.ID())
		fmt.Fprintf(stdout, "Action: %s (%s)\n", action.Label, action.ID)
		fmt.Fprintf(stdout, "Description: %s\n", action.Description)
		if action.Usage != "" {
			fmt.Fprintf(stdout, "Usage: %s\n", action.Usage)
		}
		return nil
	}

	if err := modules.ValidateRequirements(tool); err != nil {
		return fmt.Errorf("requirements failed for module %q: %w", moduleID, err)
	}

	params, positionals, err := parseArgs(rest)
	if err != nil {
		return err
	}

	out, err := action.Run(modules.ActionContext{
		Params:      params,
		Positionals: positionals,
	})
	if err != nil {
		return err
	}

	if out != "" {
		fmt.Fprintln(stdout, out)
	}
	_ = stderr
	return nil
}

func hasHelpFlag(args []string) bool {
	for _, arg := range args {
		if arg == "--help" || arg == "-h" || arg == "help" {
			return true
		}
	}
	return false
}

func parseArgs(args []string) (map[string]string, []string, error) {
	params := map[string]string{}
	positionals := make([]string, 0)

	i := 0
	for i < len(args) {
		token := args[i]

		switch {
		case strings.HasPrefix(token, "--"):
			trimmed := strings.TrimPrefix(token, "--")
			if strings.Contains(trimmed, "=") {
				parts := strings.SplitN(trimmed, "=", 2)
				if parts[0] == "" {
					return nil, nil, fmt.Errorf("invalid flag %q", token)
				}
				params[parts[0]] = parts[1]
				i++
				continue
			}

			if i+1 >= len(args) {
				return nil, nil, fmt.Errorf("missing value for flag %q", token)
			}
			if strings.HasPrefix(args[i+1], "-") {
				return nil, nil, fmt.Errorf("missing value for flag %q", token)
			}
			params[trimmed] = args[i+1]
			i += 2
			continue

		case strings.Contains(token, "="):
			parts := strings.SplitN(token, "=", 2)
			if parts[0] == "" {
				return nil, nil, fmt.Errorf("invalid argument %q", token)
			}
			params[parts[0]] = parts[1]
			i++
			continue

		default:
			positionals = append(positionals, token)
			i++
		}
	}

	return params, positionals, nil
}
