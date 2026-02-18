package modules

import (
	"go-devtools/internal/menu"
	"go-devtools/internal/requirements"
)

type Tool interface {
	ID() string
	Label() string
	Description() string
	Menu() *menu.Menu
	Requirements() []requirements.Check
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
