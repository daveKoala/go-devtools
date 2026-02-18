package main

import (
	"fmt"
	"os"

	"go-devtools/internal/menu"
	"go-devtools/internal/modules"
	"go-devtools/internal/modules/authtoken"
	"go-devtools/internal/modules/chucknorris"
	"go-devtools/internal/modules/cloudcli"
	"go-devtools/internal/modules/envinfo"
	"go-devtools/internal/modules/helloworld"
)

func main() {
	toolModules := []modules.Tool{
		helloworld.New(),
		envinfo.New(),
		chucknorris.New(),
		authtoken.New(),
		cloudcli.New(),
	}

	items := modules.ToMenuItems(toolModules)
	items = append(items, menu.QuitItem("Exit"))
	root := menu.New("Developer Tools CLI", items)
	if err := menu.Run(root); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
