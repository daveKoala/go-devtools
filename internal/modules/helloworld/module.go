package helloworld

import (
	"time"

	"go-devtools/internal/menu"
	"go-devtools/internal/modules"
	"go-devtools/internal/requirements"
)

type Tool struct{}

func New() modules.Tool {
	return Tool{}
}

func (Tool) ID() string { return "hello-tool" }

func (Tool) Label() string { return "Hello Tool" }

func (Tool) Description() string { return "Example standalone module with nested menus" }

func (Tool) Requirements() []requirements.Check { return nil }

func (Tool) Actions() []modules.Action {
	return []modules.Action{
		{
			ID:          "greet",
			Label:       "Print greeting",
			Description: "Simple example action",
			Usage:       "devtools run hello-tool greet",
			Run:         runGreeting,
		},
		{
			ID:          "timestamp",
			Label:       "Show current timestamp",
			Description: "Print the current RFC3339 timestamp",
			Usage:       "devtools run hello-tool timestamp",
			Run:         runTimestamp,
		},
	}
}

func (Tool) Menu() *menu.Menu {
	utilities := menu.NewBuilder("Hello Tool / Utilities").
		Action("Show current timestamp", "Runs a local action", func() (string, error) {
			return runTimestamp(modules.ActionContext{})
		}).
		WithBack().
		Build()

	return menu.NewBuilder("Hello Tool").
		Action("Print greeting", "Simple example action", func() (string, error) {
			return runGreeting(modules.ActionContext{})
		}).
		SubMenu("Utilities", "Nested submenu example", utilities).
		WithBack().
		Build()
}

func runGreeting(_ modules.ActionContext) (string, error) {
	return "Hello from the Hello Tool module.", nil
}

func runTimestamp(_ modules.ActionContext) (string, error) {
	return time.Now().Format(time.RFC3339), nil
}
