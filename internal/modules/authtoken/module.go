package authtoken

import (
	"bufio"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"go-devtools/internal/menu"
	"go-devtools/internal/modules"
	"go-devtools/internal/requirements"
)

type Tool struct{}

func New() modules.Tool {
	return Tool{}
}

func (Tool) ID() string { return "auth-token-generator" }

func (Tool) Label() string { return "Auth Token Generator" }

func (Tool) Description() string { return "Generate mock auth tokens for local development" }

func (Tool) Requirements() []requirements.Check { return nil }

func (Tool) Menu() *menu.Menu {
	userPassMenu := menu.NewBuilder("Auth Token / Username + Password").
		Action("Generate token", "Prompt for username and password", generateUserPassToken).
		WithBack().
		Build()

	googleMenu := menu.NewBuilder("Auth Token / Google").
		Action("Generate token", "Prompt for Google email identity", generateGoogleToken).
		WithBack().
		Build()

	return menu.NewBuilder("Auth Token Generator").
		SubMenu("Username + Password", "Token for classic credential flow", userPassMenu).
		SubMenu("Google", "Token for Google OAuth flow", googleMenu).
		WithBack().
		Build()
}

func generateUserPassToken() (string, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Username: ")
	username, err := readLine(reader)
	if err != nil {
		return "", err
	}

	fmt.Print("Password: ")
	password, err := readLine(reader)
	if err != nil {
		return "", err
	}
	if password == "" {
		return "", fmt.Errorf("password cannot be empty")
	}

	return buildToken(map[string]string{
		"method":   "password",
		"username": username,
	})
}

func generateGoogleToken() (string, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Google email: ")
	email, err := readLine(reader)
	if err != nil {
		return "", err
	}
	if email == "" {
		return "", fmt.Errorf("email cannot be empty")
	}

	return buildToken(map[string]string{
		"method": "google",
		"email":  email,
	})
}

func readLine(reader *bufio.Reader) (string, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

func buildToken(claims map[string]string) (string, error) {
	claims["iat"] = fmt.Sprintf("%d", time.Now().Unix())
	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("failed to marshal claims: %w", err)
	}

	signature := make([]byte, 24)
	if _, err := rand.Read(signature); err != nil {
		return "", fmt.Errorf("failed to generate random signature: %w", err)
	}

	payload := base64.RawURLEncoding.EncodeToString(claimsJSON)
	sig := base64.RawURLEncoding.EncodeToString(signature)
	return fmt.Sprintf("dev.%s.%s", payload, sig), nil
}
