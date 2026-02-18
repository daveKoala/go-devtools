package cloudcli

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"go-devtools/internal/menu"
	"go-devtools/internal/modules"
	"go-devtools/internal/requirements"
)

type Tool struct{}

func New() modules.Tool {
	return Tool{}
}

func (Tool) ID() string { return "cloud-cli-checks" }

func (Tool) Label() string { return "Cloud CLI Checks" }

func (Tool) Description() string { return "AWS/Azure CLI checks with install actions" }

func (Tool) Requirements() []requirements.Check { return nil }

func (Tool) Actions() []modules.Action {
	return []modules.Action{
		{
			ID:          "aws-version",
			Label:       "Show aws version",
			Description: "Runs aws --version",
			Usage:       "devtools run cloud-cli-checks aws-version",
			Run:         showAWSVersion,
		},
		{
			ID:          "azure-version",
			Label:       "Show az version",
			Description: "Runs az version",
			Usage:       "devtools run cloud-cli-checks azure-version",
			Run:         showAzureVersion,
		},
	}
}

func (Tool) Menu() *menu.Menu {
	awsMenu := menu.NewBuilder("Cloud CLI / AWS").
		Action("Show aws version", "Runs aws --version", func() (string, error) {
			return showAWSVersion(modules.ActionContext{})
		}).
		WithBack().
		Build()

	azureMenu := menu.NewBuilder("Cloud CLI / Azure").
		Action("Show az version", "Runs az version", func() (string, error) {
			return showAzureVersion(modules.ActionContext{})
		}).
		WithBack().
		Build()

	return menu.NewBuilder("Cloud CLI Checks").
		SubMenu("AWS CLI", "Requires aws command", awsMenu, requirements.CommandExistsWithBrew("aws", "awscli")).
		SubMenu("Azure CLI", "Requires az command", azureMenu, requirements.CommandExistsWithBrew("az", "azure-cli")).
		WithBack().
		Build()
}

func showAWSVersion(_ modules.ActionContext) (string, error) {
	return runCommand("aws", "--version")
}

func showAzureVersion(_ modules.ActionContext) (string, error) {
	return runCommand("az", "version")
}

func runCommand(name string, args ...string) (string, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd := exec.Command(name, args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%s %s failed: %w (%s)", name, strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}

	out := strings.TrimSpace(stdout.String())
	if out == "" {
		out = strings.TrimSpace(stderr.String())
	}
	if out == "" {
		out = "Command completed with no output."
	}
	return out, nil
}
