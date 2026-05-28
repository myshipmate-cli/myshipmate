package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

// HerokuDeploy deploys the current project to Heroku
func HerokuDeploy(project *ProjectInfo, envVars []EnvVar) error {
	token, err := EnsureLoggedIn("heroku")
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("🚀 Deploying to Heroku...")

	// Check if Heroku CLI is available (preferred method)
	if path, err := exec.LookPath("heroku"); err == nil {
		return herokuDeployViaCLI(path, project, envVars, token)
	}

	// Fall back to API deployment
	return herokuDeployViaAPI(token, project, envVars)
}

// herokuDeployViaCLI uses the Heroku CLI to deploy (preferred)
func herokuDeployViaCLI(herokuPath string, project *ProjectInfo, envVars []EnvVar, token string) error {
	// Authenticate
	fmt.Println("   🔐 Authenticating with Heroku CLI...")
	cmd := exec.Command(herokuPath, "auth:token")
	if _, err := cmd.CombinedOutput(); err != nil {
		// Not logged in, try to login
		cmd = exec.Command(herokuPath, "login", "--interactive")
		cmd.Env = append(os.Environ(), fmt.Sprintf("HEROKU_API_KEY=%s", token))
		if out, err := cmd.CombinedOutput(); err != nil {
			fmt.Printf("   ⚠ Heroku CLI auth failed: %s\n", string(out))
		}
	}
	fmt.Println("   ✓ Authenticated with Heroku CLI")

	// Check if app exists or create it
	fmt.Println("   📦 Setting up Heroku app...")
	cmd = exec.Command(herokuPath, "apps:info")
	cmd.Dir = project.Path
	out, err := cmd.CombinedOutput()

	if err != nil || strings.Contains(string(out), "no app") {
		// Create new app
		cmd = exec.Command(herokuPath, "create", project.Name)
		cmd.Dir = project.Path
		if _, err := cmd.CombinedOutput(); err != nil {
			// Name might be taken, create with random name
			cmd = exec.Command(herokuPath, "create")
			cmd.Dir = project.Path
			if out2, err := cmd.CombinedOutput(); err != nil {
				return fmt.Errorf("heroku create failed: %s", string(out2))
			}
		}
		fmt.Printf("   ✓ App created: %s\n", project.Name)
	} else {
		fmt.Printf("   ✓ Using existing app\n")
	}

	// Set environment variables
	if len(envVars) > 0 {
		fmt.Println("   📋 Setting config vars...")
		var configPairs []string
		for _, v := range envVars {
			configPairs = append(configPairs, fmt.Sprintf("%s=%s", v.Key, v.Value))
		}
		args := append([]string{"config:set"}, configPairs...)
		cmd = exec.Command(herokuPath, args...)
		cmd.Dir = project.Path
		if out, err := cmd.CombinedOutput(); err != nil {
			fmt.Printf("   ⚠ Warning: could not set all config vars: %s\n", string(out))
		} else {
			fmt.Printf("   ✓ Set %d config vars\n", len(envVars))
		}
	}

	// Set buildpacks based on project type
	setHerokuBuildpack(herokuPath, project)

	// Deploy via git push
	fmt.Println("   📤 Deploying via git push...")

	// Make sure git remote exists
	cmd = exec.Command("git", "remote", "get-url", "heroku")
	cmd.Dir = project.Path
	if _, err := cmd.CombinedOutput(); err != nil {
		// Add heroku remote
		cmd = exec.Command(herokuPath, "git:remote", "-a", project.Name)
		cmd.Dir = project.Path
		cmd.CombinedOutput()
	}

	// Commit any uncommitted changes (WITH USER CONFIRMATION)
	cmd = exec.Command("git", "status", "--porcelain")
	cmd.Dir = project.Path
	if out, _ := cmd.CombinedOutput(); len(out) > 0 {
		changedFiles := strings.TrimSpace(string(out))
		fileCount := len(strings.Split(changedFiles, "\n"))

		fmt.Printf("\n   ⚠ You have %d uncommitted file(s):\n", fileCount)
		for _, line := range strings.Split(changedFiles, "\n") {
			if line != "" {
				fmt.Printf("      %s\n", strings.TrimSpace(line))
			}
		}
		fmt.Println()
		fmt.Print("   Heroku deploys from git. Commit and push these changes? (y/n): ")

		reader := bufio.NewReader(os.Stdin)
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))

		if answer == "y" || answer == "yes" {
			fmt.Println("   ℹ Committing changes...")
			cmd = exec.Command("git", "add", ".")
			cmd.Dir = project.Path
			cmd.Run()
			cmd = exec.Command("git", "commit", "-m", "Deploy via Shipmate")
			cmd.Dir = project.Path
			cmd.Run()
		} else {
			fmt.Println("   ✗ Skipped commit. Heroku needs a clean git push to deploy.")
			fmt.Println("     Commit your changes manually, then run: git push heroku main")
			return nil
		}
	}

	// Push to Heroku
	cmd = exec.Command("git", "push", "heroku", "main")
	cmd.Dir = project.Path
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		// Try master branch
		cmd = exec.Command("git", "push", "heroku", "master")
		cmd.Dir = project.Path
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("git push to heroku failed")
		}
	}

	// Get app URL
	cmd = exec.Command(herokuPath, "apps:info", "--json")
	cmd.Dir = project.Path
	out, err = cmd.CombinedOutput()
	appURL := fmt.Sprintf("https://%s.herokuapp.com", project.Name)
	if err == nil {
		var appInfo struct {
			App struct {
				WebURL string `json:"web_url"`
			} `json:"app"`
		}
		if json.Unmarshal(out, &appInfo) == nil && appInfo.App.WebURL != "" {
			appURL = appInfo.App.WebURL
		}
	}

	fmt.Println()
	fmt.Printf("   ✓ Deployed successfully!\n")
	fmt.Printf("   🌐 Live at: %s\n", appURL)

	SaveProjectState(project.Path, ProjectState{
		Platform:    "heroku",
		ProjectName: project.Name,
		DeployURL:   appURL,
	})

	return nil
}

// setHerokuBuildpack sets the appropriate buildpack based on project type
func setHerokuBuildpack(herokuPath string, project *ProjectInfo) {
	var buildpack string

	switch project.Type {
	case ProjectNode, ProjectNextJS, ProjectReact:
		buildpack = "heroku/nodejs"
	case ProjectPython, ProjectDjango, ProjectFlask, ProjectFastAPI:
		buildpack = "heroku/python"
	case ProjectRuby, ProjectRails, ProjectSinatra:
		buildpack = "heroku/ruby"
	case ProjectJava, ProjectSpring, ProjectKotlin:
		buildpack = "heroku/java"
	case ProjectGo:
		buildpack = "heroku/go"
	case ProjectPHP, ProjectLaravel, ProjectSymfony:
		buildpack = "heroku/php"
	default:
		return // Let Heroku auto-detect
	}

	cmd := exec.Command(herokuPath, "buildpacks:set", buildpack)
	cmd.Dir = project.Path
	cmd.CombinedOutput()
	fmt.Printf("   ✓ Buildpack: %s\n", buildpack)
}

// herokuDeployViaAPI uses the Heroku Platform API
func herokuDeployViaAPI(token string, project *ProjectInfo, envVars []EnvVar) error {
	fmt.Println("   ℹ Heroku CLI not found. Using API deployment.")
	fmt.Println()

	// Create app
	appName, err := herokuCreateApp(token, project.Name)
	if err != nil {
		return fmt.Errorf("app creation failed: %w", err)
	}
	fmt.Printf("   ✓ App created: %s\n", appName)

	// Set config vars
	if len(envVars) > 0 {
		if err := herokuSetConfigVars(token, appName, envVars); err != nil {
			fmt.Printf("   ⚠ Warning: could not set config vars: %v\n", err)
		} else {
			fmt.Printf("   ✓ Set %d config vars\n", len(envVars))
		}
	}

	// Heroku API deployment requires source blob upload which is complex
	// Recommend CLI or git push
	fmt.Println()
	fmt.Println("   📤 To deploy your code:")
	fmt.Println()
	fmt.Println("   Option 1: Install Heroku CLI (recommended)")
	fmt.Println("     $ curl https://cli-assets.heroku.com/install.sh | sh")
	fmt.Println("     $ heroku login")
	fmt.Println("     $ heroku git:remote -a " + appName)
	fmt.Println("     $ git push heroku main")
	fmt.Println()
	fmt.Println("   Option 2: Connect GitHub repo")
	fmt.Printf("     https://dashboard.heroku.com/apps/%s/deploy/github\n", appName)
	fmt.Println()
	fmt.Printf("   🌐 App URL: https://%s.herokuapp.com\n", appName)
	fmt.Printf("   📋 Dashboard: https://dashboard.heroku.com/apps/%s\n", appName)

	SaveProjectState(project.Path, ProjectState{
		Platform:    "heroku",
		ProjectName: appName,
		DeployURL:   fmt.Sprintf("https://%s.herokuapp.com", appName),
	})

	return nil
}

func herokuCreateApp(token, name string) (string, error) {
	body := map[string]string{
		"name":   name,
		"region": "us",
	}
	bodyBytes, _ := json.Marshal(body)

	req, err := http.NewRequest("POST", "https://api.heroku.com/apps", bytes.NewReader(bodyBytes))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/vnd.heroku+json; version=3")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		data, _ := io.ReadAll(resp.Body)
		// If name taken, try without name
		if resp.StatusCode == 422 {
			body = map[string]string{"region": "us"}
			bodyBytes, _ = json.Marshal(body)
			req, _ = http.NewRequest("POST", "https://api.heroku.com/apps", bytes.NewReader(bodyBytes))
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Accept", "application/vnd.heroku+json; version=3")

			resp, err = http.DefaultClient.Do(req)
			if err != nil {
				return "", err
			}
			defer resp.Body.Close()

			if resp.StatusCode == 201 {
				var app struct {
					Name string `json:"name"`
				}
				json.NewDecoder(resp.Body).Decode(&app)
				return app.Name, nil
			}
		}
		return "", fmt.Errorf("heroku API error (%d): %s", resp.StatusCode, string(data))
	}

	var app struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&app); err != nil {
		return "", err
	}

	return app.Name, nil
}

func herokuSetConfigVars(token, appName string, envVars []EnvVar) error {
	configMap := make(map[string]string)
	for _, v := range envVars {
		configMap[v.Key] = v.Value
	}

	bodyBytes, _ := json.Marshal(configMap)

	req, err := http.NewRequest("PATCH",
		fmt.Sprintf("https://api.heroku.com/apps/%s/config-vars", appName),
		bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/vnd.heroku+json; version=3")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		data, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("config var update failed (%d): %s", resp.StatusCode, string(data))
	}

	return nil
}
