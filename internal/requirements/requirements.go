package requirements

import (
	"fmt"
	"os"
	"os/exec"
)

type InstallAction struct {
	Label string
	Run   func() (string, error)
}

type Check struct {
	Name      string
	Validate  func() error
	Installer *InstallAction
}

func (c Check) Run() error {
	if c.Validate == nil {
		return nil
	}
	return c.Validate()
}

func CommandExists(name string) Check {
	return Check{
		Name: name,
		Validate: func() error {
			if _, err := exec.LookPath(name); err != nil {
				return fmt.Errorf("required command %q is not installed", name)
			}
			return nil
		},
	}
}

func CommandExistsWithBrew(name, formula string) Check {
	check := CommandExists(name)
	check.Installer = &InstallAction{
		Label: fmt.Sprintf("Install %s with Homebrew", name),
		Run: func() (string, error) {
			if _, err := exec.LookPath("brew"); err != nil {
				return "", fmt.Errorf("homebrew is required for auto-install of %q", formula)
			}
			cmd := exec.Command("brew", "install", formula)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return "", fmt.Errorf("brew install %s failed: %w", formula, err)
			}
			return fmt.Sprintf("Installed %q with Homebrew.", formula), nil
		},
	}
	return check
}

func EnvVarSet(name string) Check {
	return Check{
		Name: name,
		Validate: func() error {
			if os.Getenv(name) == "" {
				return fmt.Errorf("required environment variable %q is not set", name)
			}
			return nil
		},
	}
}
