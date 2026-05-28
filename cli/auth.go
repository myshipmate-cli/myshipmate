package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

// Config holds the shipmate configuration stored at ~/.shipmate/config.json
type Config struct {
	Platforms map[string]PlatformAuth `json:"platforms"`
	Projects  map[string]ProjectState `json:"projects"`
}

// PlatformAuth stores authentication info for a platform
type PlatformAuth struct {
	Token     string `json:"token"`
	TeamID    string `json:"team_id,omitempty"`
	ProjectID string `json:"project_id,omitempty"`
}

// ProjectState stores state for a previously deployed project
type ProjectState struct {
	Platform    string `json:"platform"`
	ProjectID   string `json:"project_id"`
	ProjectName string `json:"project_name"`
	DeployURL   string `json:"deploy_url"`
	LastDeploy  string `json:"last_deploy"`
}

var configPath string

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	configPath = filepath.Join(home, ".shipmate", "config.json")
}

// LoadConfig loads the configuration from disk
func LoadConfig() (*Config, error) {
	cfg := &Config{
		Platforms: make(map[string]PlatformAuth),
		Projects:  make(map[string]ProjectState),
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil // no config yet, return empty
		}
		return nil, err
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("corrupt config at %s: %w", configPath, err)
	}

	if cfg.Platforms == nil {
		cfg.Platforms = make(map[string]PlatformAuth)
	}
	if cfg.Projects == nil {
		cfg.Projects = make(map[string]ProjectState)
	}

	return cfg, nil
}

// SaveConfig persists the configuration to disk
func SaveConfig(cfg *Config) error {
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0600)
}

// GetToken returns the stored token for a platform, or empty string if not logged in
func GetToken(platform string) (string, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return "", err
	}

	auth, ok := cfg.Platforms[platform]
	if !ok {
		return "", nil
	}
	return auth.Token, nil
}

// SetToken stores a token for a platform
func SetToken(platform, token string) error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	auth := cfg.Platforms[platform]
	auth.Token = token
	cfg.Platforms[platform] = auth

	return SaveConfig(cfg)
}

// SaveProjectState saves deployment state for a project
func SaveProjectState(projectPath string, state ProjectState) error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	if state.LastDeploy == "" {
		state.LastDeploy = time.Now().UTC().Format(time.RFC3339)
	}

	cfg.Projects[projectPath] = state
	return SaveConfig(cfg)
}

// GetProjectState retrieves deployment state for a project
func GetProjectState(projectPath string) (*ProjectState, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return nil, err
	}

	state, ok := cfg.Projects[projectPath]
	if !ok {
		return nil, nil
	}
	return &state, nil
}

// OpenBrowser opens the default browser to the given URL
func OpenBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return fmt.Errorf("unsupported platform")
	}
	return cmd.Start()
}

// LoginToPlatform handles the login flow for a platform
func LoginToPlatform(platform string) error {
	var tokenURL string
	var instructions string

	switch platform {
	case "vercel":
		tokenURL = "https://vercel.com/account/tokens"
		instructions = `Vercel Login:
  1. Opening browser to Vercel token page...
  2. Click "Create Token"
  3. Give it a name like "shipmate"
  4. Copy the token and paste it below`
	case "railway":
		tokenURL = "https://railway.app/account/tokens"
		instructions = `Railway Login:
  1. Opening browser to Railway token page...
  2. Click "Create Token"
  3. Give it a name like "shipmate"
  4. Copy the token and paste it below`
	case "render":
		tokenURL = "https://dashboard.render.com/u/0/apikeys"
		instructions = `Render Login:
  1. Opening browser to Render API keys page...
  2. Click "Create API Key"
  3. Give it a name like "shipmate"
  4. Copy the key and paste it below`
	case "netlify":
		tokenURL = "https://app.netlify.com/user/applications#personal-access-tokens"
		instructions = `Netlify Login:
  1. Opening browser to Netlify tokens page...
  2. Click "New access token"
  3. Give it a name like "shipmate"
  4. Copy the token and paste it below`
	case "flyio":
		tokenURL = "https://fly.io/user/personal_access_tokens"
		instructions = `Fly.io Login:
  1. Opening browser to Fly.io tokens page...
  2. Click "Create token"
  3. Copy the token and paste it below`
	case "heroku":
		tokenURL = "https://dashboard.heroku.com/account"
		instructions = `Heroku Login:
  1. Opening browser to Heroku account page...
  2. Scroll down to "API Key" section
  3. Click "Reveal" to show your API key
  4. Copy the key and paste it below`
	default:
		return fmt.Errorf("unknown platform: %s", platform)
	}

	fmt.Println()
	fmt.Println(instructions)
	fmt.Println()

	// Open browser
	if err := OpenBrowser(tokenURL); err != nil {
		fmt.Printf("   ⚠ Could not open browser. Visit manually: %s\n", tokenURL)
	} else {
		fmt.Printf("   ✓ Opened: %s\n", tokenURL)
	}

	fmt.Println()
	fmt.Print("   Paste token: ")

	var token string
	fmt.Scanln(&token)

	if token == "" {
		return fmt.Errorf("no token provided")
	}

	if err := SetToken(platform, token); err != nil {
		return err
	}

	fmt.Printf("   ✓ Token saved for %s\n", platform)
	return nil
}

// EnsureLoggedIn checks if user is logged in to a platform, prompts if not
func EnsureLoggedIn(platform string) (string, error) {
	token, err := GetToken(platform)
	if err != nil {
		return "", err
	}

	if token != "" {
		return token, nil
	}

	fmt.Printf("\n🔐 You need to log in to %s first.\n", platform)
	if err := LoginToPlatform(platform); err != nil {
		return "", err
	}

	token, err = GetToken(platform)
	if err != nil {
		return "", err
	}
	return token, nil
}
