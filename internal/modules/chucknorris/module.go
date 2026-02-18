package chucknorris

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go-devtools/internal/menu"
	"go-devtools/internal/modules"
	"go-devtools/internal/requirements"
)

type Tool struct{}

func New() modules.Tool {
	return Tool{}
}

func (Tool) ID() string { return "chuck-norris-facts" }

func (Tool) Label() string { return "Chuck Norris Fact" }

func (Tool) Description() string { return "Fetch a random fact from api.chucknorris.io" }

func (Tool) Requirements() []requirements.Check { return nil }

func (Tool) Actions() []modules.Action {
	return []modules.Action{
		{
			ID:          "random-fact",
			Label:       "Get random fact",
			Description: "Calls GET https://api.chucknorris.io/jokes/random",
			Usage:       "devtools run chuck-norris-facts random-fact",
			Run:         fetchFactAction,
		},
	}
}

func (Tool) Menu() *menu.Menu {
	return menu.NewBuilder("Chuck Norris Fact Tool").
		Action("Get random fact", "Calls GET https://api.chucknorris.io/jokes/random", func() (string, error) {
			return fetchFactAction(modules.ActionContext{})
		}).
		WithBack().
		Build()
}

type jokeResponse struct {
	ID    string `json:"id"`
	URL   string `json:"url"`
	Value string `json:"value"`
}

func fetchFactAction(_ modules.ActionContext) (string, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get("https://api.chucknorris.io/jokes/random")
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status: %s", resp.Status)
	}

	var joke jokeResponse
	if err := json.NewDecoder(resp.Body).Decode(&joke); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return fmt.Sprintf("%s\n\nSource: %s", joke.Value, joke.URL), nil
}
