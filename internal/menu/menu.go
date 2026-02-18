package menu

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"go-devtools/internal/requirements"
)

const defaultMaxDepth = 4
const uiWidth = 72

const (
	ansiReset  = "\033[0m"
	ansiBold   = "\033[1m"
	ansiDim    = "\033[2m"
	ansiCyan   = "\033[36m"
	ansiBlue   = "\033[34m"
	ansiGreen  = "\033[32m"
	ansiYellow = "\033[33m"
	ansiRed    = "\033[31m"
	ansiWhite  = "\033[37m"
)

type Action int

const (
	ActionNone Action = iota
	ActionBack
	ActionQuit
)

type Item struct {
	Label        string
	Description  string
	NextMenu     *Menu
	Run          func() (string, error)
	Requirements []requirements.Check
	Action       Action
}

type Menu struct {
	Title string
	Items []Item
}

func New(title string, items []Item) *Menu {
	return &Menu{
		Title: title,
		Items: items,
	}
}

func WithExit(items []Item, label string, action Action) []Item {
	return append(items, Item{
		Label:  label,
		Action: action,
	})
}

func WithBack(items []Item) []Item {
	return WithExit(items, "Back", ActionBack)
}

func WithQuit(items []Item) []Item {
	return WithExit(items, "Exit", ActionQuit)
}

func BackItem(label string) Item {
	return Item{
		Label:  label,
		Action: ActionBack,
	}
}

func QuitItem(label string) Item {
	return Item{
		Label:  label,
		Action: ActionQuit,
	}
}

type Builder struct {
	title string
	items []Item
}

func NewBuilder(title string) *Builder {
	return &Builder{title: title}
}

func (b *Builder) Action(label, description string, run func() (string, error)) *Builder {
	b.items = append(b.items, Item{
		Label:       label,
		Description: description,
		Run:         run,
	})
	return b
}

func (b *Builder) SubMenu(label, description string, submenu *Menu, checks ...requirements.Check) *Builder {
	b.items = append(b.items, Item{
		Label:        label,
		Description:  description,
		NextMenu:     submenu,
		Requirements: checks,
	})
	return b
}

func (b *Builder) Custom(item Item) *Builder {
	b.items = append(b.items, item)
	return b
}

func (b *Builder) WithBack() *Builder {
	b.items = WithBack(b.items)
	return b
}

func (b *Builder) WithQuit() *Builder {
	b.items = WithQuit(b.items)
	return b
}

func (b *Builder) Build() *Menu {
	return New(b.title, b.items)
}

type key int

const (
	keyUnknown key = iota
	keyUp
	keyDown
	keyLeft
	keyEnter
	keyInstall
	keyQuit
)

type terminalState struct {
	previous string
}

type requirementFailure struct {
	err       error
	installer *requirements.InstallAction
}

type Runner struct {
	stack          []*Menu
	cursor         int
	status         string
	maxDepth       int
	term           *terminalState
	pendingInstall *requirements.InstallAction
	useColor       bool
}

func NewRunner(root *Menu) *Runner {
	return &Runner{
		stack:    []*Menu{root},
		maxDepth: defaultMaxDepth,
		useColor: os.Getenv("NO_COLOR") == "",
	}
}

func Run(root *Menu) error {
	return NewRunner(root).Run()
}

func (r *Runner) Run() error {
	state, err := setRawMode()
	if err != nil {
		return err
	}
	r.term = state
	defer restoreMode(state)

	reader := bufio.NewReader(os.Stdin)
	for {
		r.render()

		pressed, err := readKey(reader)
		if err != nil {
			return err
		}

		done, err := r.handleKey(pressed)
		if err != nil {
			return err
		}
		if done {
			fmt.Print("\033[H\033[2J\r")
			return nil
		}
	}
}

func (r *Runner) handleKey(pressed key) (bool, error) {
	current := r.currentMenu()
	switch pressed {
	case keyQuit:
		return true, nil
	case keyInstall:
		if r.pendingInstall == nil {
			r.status = "No install action is available for the current requirement error."
			return false, nil
		}

		out, err := r.runAction(r.pendingInstall.Run)
		if err != nil {
			r.status = fmt.Sprintf("Install error: %v", err)
		} else {
			r.status = out
			if out == "" {
				r.status = "Install action completed. Re-enter the module to retry checks."
			}
		}
		r.pendingInstall = nil
		return false, nil
	case keyUp:
		if len(current.Items) == 0 {
			return false, nil
		}
		r.cursor--
		if r.cursor < 0 {
			r.cursor = len(current.Items) - 1
		}
	case keyDown:
		if len(current.Items) == 0 {
			return false, nil
		}
		r.cursor++
		if r.cursor >= len(current.Items) {
			r.cursor = 0
		}
	case keyLeft:
		if len(r.stack) > 1 {
			r.stack = r.stack[:len(r.stack)-1]
			r.cursor = 0
			r.status = ""
			r.pendingInstall = nil
		}
	case keyEnter:
		if len(current.Items) == 0 {
			return false, nil
		}

		selected := current.Items[r.cursor]
		switch selected.Action {
		case ActionQuit:
			return true, nil
		case ActionBack:
			if len(r.stack) > 1 {
				r.stack = r.stack[:len(r.stack)-1]
				r.cursor = 0
				r.status = ""
				r.pendingInstall = nil
			}
			return false, nil
		}

		if selected.NextMenu != nil {
			if len(r.stack) >= r.maxDepth {
				r.status = fmt.Sprintf("Max menu depth reached (%d). Go back before opening more submenus.", r.maxDepth)
				return false, nil
			}

			if failure := firstFailedRequirement(selected.Requirements); failure != nil {
				r.status = fmt.Sprintf("Requirement failed: %v", failure.err)
				r.pendingInstall = failure.installer
				if r.pendingInstall != nil {
					r.status = fmt.Sprintf("%s | Press i to %s.", r.status, r.pendingInstall.Label)
				}
				return false, nil
			}

			r.stack = append(r.stack, selected.NextMenu)
			r.cursor = 0
			r.status = ""
			r.pendingInstall = nil
			return false, nil
		}

		if selected.Run != nil {
			out, err := r.runAction(selected.Run)
			if err != nil {
				r.status = fmt.Sprintf("Error: %v", err)
			} else {
				r.status = out
			}
		}
	}
	return false, nil
}

func (r *Runner) runAction(action func() (string, error)) (string, error) {
	if action == nil {
		return "", nil
	}

	restoreMode(r.term)
	out, runErr := action()
	rawErr := enableRawMode()
	if rawErr != nil {
		if runErr != nil {
			return out, fmt.Errorf("%v; failed to restore raw mode: %w", runErr, rawErr)
		}
		return out, fmt.Errorf("failed to restore raw mode: %w", rawErr)
	}
	return out, runErr
}

func (r *Runner) render() {
	current := r.currentMenu()
	var b strings.Builder

	b.WriteString("\033[H\033[2J\r")
	topRule := strings.Repeat("=", uiWidth)
	bottomRule := strings.Repeat("-", uiWidth)

	fmt.Fprintf(&b, "%s\r\n", r.paint(topRule, ansiCyan))
	fmt.Fprintf(&b, "%s\r\n", r.paint("DEV TOOLS CLI", ansiBold+ansiWhite))
	fmt.Fprintf(&b, "%s %s\r\n", r.paint("Menu:", ansiBold+ansiBlue), r.paint(current.Title, ansiYellow))
	fmt.Fprintf(&b, "%s %d/%d\r\n", r.paint("Depth:", ansiBold+ansiBlue), len(r.stack), r.maxDepth)
	fmt.Fprintf(&b, "%s\r\n\r\n", r.paint(topRule, ansiCyan))

	for i, item := range current.Items {
		cursor := "  "
		if r.cursor == i {
			cursor = "▶ "
		}

		label := fmt.Sprintf("%s%s", cursor, item.Label)
		if r.cursor == i {
			label = r.paint(label, ansiBold+ansiGreen)
		} else {
			label = r.paint(label, ansiWhite)
		}
		fmt.Fprint(&b, label)

		if item.Description != "" {
			fmt.Fprintf(&b, " %s %s", r.paint("-", ansiDim), r.paint(item.Description, ansiDim))
		}
		b.WriteString("\r\n")
	}

	fmt.Fprintf(&b, "\r\n%s\r\n", r.paint(bottomRule, ansiCyan))
	controls := "↑/↓ move | Enter select | ← back | q quit"
	if r.pendingInstall != nil {
		controls += " | i install"
	}
	fmt.Fprintf(&b, "%s\r\n", r.paint(controls, ansiDim))
	if r.status != "" {
		fmt.Fprintf(&b, "\r\n%s\r\n", r.formatStatus(r.status))
	}

	fmt.Print(b.String())
}

func (r *Runner) currentMenu() *Menu {
	return r.stack[len(r.stack)-1]
}

func firstFailedRequirement(checks []requirements.Check) *requirementFailure {
	for _, check := range checks {
		if err := check.Run(); err != nil {
			return &requirementFailure{
				err:       err,
				installer: check.Installer,
			}
		}
	}
	return nil
}

func readKey(reader *bufio.Reader) (key, error) {
	b, err := reader.ReadByte()
	if err != nil {
		return keyUnknown, err
	}

	switch b {
	case 'q', 'Q':
		return keyQuit, nil
	case 'i', 'I':
		return keyInstall, nil
	case '\r', '\n':
		return keyEnter, nil
	case 27:
		b1, err := reader.ReadByte()
		if err != nil {
			return keyUnknown, nil
		}
		b2, err := reader.ReadByte()
		if err != nil {
			return keyUnknown, nil
		}
		if b1 == '[' {
			switch b2 {
			case 'A':
				return keyUp, nil
			case 'B':
				return keyDown, nil
			case 'D':
				return keyLeft, nil
			}
		}
	}

	return keyUnknown, nil
}

func setRawMode() (*terminalState, error) {
	get := exec.Command("stty", "-g")
	get.Stdin = os.Stdin
	output, err := get.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to read terminal state: %w", err)
	}

	prev := strings.TrimSpace(string(output))
	if err := enableRawMode(); err != nil {
		return nil, err
	}
	return &terminalState{previous: prev}, nil
}

func enableRawMode() error {
	raw := exec.Command("stty", "raw", "-echo")
	raw.Stdin = os.Stdin
	if err := raw.Run(); err != nil {
		return fmt.Errorf("failed to enable raw mode: %w", err)
	}
	return nil
}

func restoreMode(state *terminalState) {
	if state == nil {
		return
	}
	cmd := exec.Command("stty", state.previous)
	cmd.Stdin = os.Stdin
	_ = cmd.Run()
}

func normalizeCRLF(input string) string {
	normalized := strings.ReplaceAll(input, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	return strings.ReplaceAll(normalized, "\n", "\r\n")
}

func (r *Runner) paint(text, style string) string {
	if !r.useColor || style == "" {
		return text
	}
	return style + text + ansiReset
}

func (r *Runner) formatStatus(status string) string {
	color := ansiGreen
	if strings.HasPrefix(status, "Requirement failed:") {
		color = ansiYellow
	}
	if strings.HasPrefix(status, "Error:") || strings.HasPrefix(status, "Install error:") {
		color = ansiRed
	}

	lines := strings.Split(normalizeCRLF(status), "\r\n")
	for i, line := range lines {
		lines[i] = r.paint(line, color)
	}
	return strings.Join(lines, "\r\n")
}
