package modules

import (
	"fmt"

	"go-devtools/internal/menu"
	"go-devtools/internal/requirements"
)

type ActionContext struct {
	Params      map[string]string
	Positionals []string
}

type Action struct {
	ID          string
	Label       string
	Description string
	Usage       string
	Run         func(ActionContext) (string, error)
}

type Tool interface {
	ID() string
	Label() string
	Description() string
	Menu() *menu.Menu
	Requirements() []requirements.Check
	Actions() []Action
}

func ToMenuItem(tool Tool) menu.Item {
	return menu.Item{
		Label:        tool.Label(),
		Description:  tool.Description(),
		NextMenu:     tool.Menu(),
		Requirements: tool.Requirements(),
	}
}

func ToMenuItems(tools []Tool) []menu.Item {
	items := make([]menu.Item, 0, len(tools))
	for _, tool := range tools {
		items = append(items, ToMenuItem(tool))
	}
	return items
}

func FindTool(tools []Tool, id string) (Tool, bool) {
	for _, tool := range tools {
		if tool.ID() == id {
			return tool, true
		}
	}
	return nil, false
}

func FindAction(tool Tool, actionID string) (Action, bool) {
	for _, action := range tool.Actions() {
		if action.ID == actionID {
			return action, true
		}
	}
	return Action{}, false
}

func ValidateRequirements(tool Tool) error {
	for _, check := range tool.Requirements() {
		if err := check.Run(); err != nil {
			if check.Installer != nil {
				return fmt.Errorf("%w (installer available: %s)", err, check.Installer.Label)
			}
			return err
		}
	}
	return nil
}
