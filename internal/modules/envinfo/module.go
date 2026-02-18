package envinfo

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"go-devtools/internal/menu"
	"go-devtools/internal/modules"
	"go-devtools/internal/requirements"
)

type Tool struct{}

func New() modules.Tool {
	return Tool{}
}

func (Tool) ID() string { return "env-info" }

func (Tool) Label() string { return "Environment Info" }

func (Tool) Description() string { return "Diagnostics module with requirement checks" }

func (Tool) Requirements() []requirements.Check {
	return []requirements.Check{
		requirements.CommandExistsWithBrew("go", "go"),
	}
}

func (Tool) Menu() *menu.Menu {
	paths := menu.NewBuilder("Environment Info / PATH").
		Action("Show PATH entries", "Displays PATH split into lines", func() (string, error) {
			path := os.Getenv("PATH")
			entries := strings.Split(path, ":")
			return strings.Join(entries, "\n"), nil
		}).
		WithBack().
		Build()

	return menu.NewBuilder("Environment Info Tool").
		Action("Show Go runtime info", "Prints local runtime metadata", func() (string, error) {
			return fmt.Sprintf("GOOS=%s GOARCH=%s GOVERSION=%s", runtime.GOOS, runtime.GOARCH, runtime.Version()), nil
		}).
		SubMenu("PATH details", "Nested submenu example", paths).
		WithBack().
		Build()
}
