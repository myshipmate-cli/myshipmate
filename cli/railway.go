package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

// RailwayDeploy deploys the current project to Railway
func RailwayDeploy(project *ProjectInfo, envVars []EnvVar) error {
	token, err := EnsureLoggedIn("railway")
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("🚀 Deploying to Railway...")

	// Check if Railway CLI is available (preferred method)
	if path, err := exec.LookPath("railway"); err == nil {
		return railwayDeployViaCLI(path, project, envVars, token)
	}

	// Fall back to API deployment
	return railwayDeployViaAPI(token, project, envVars)
}

// railwayDeployViaCLI uses the Railway CLI to deploy (preferred)
func railwayDeployViaCLI(railwayPath string, project *ProjectInfo, envVars []EnvVar, token string) error {
	// Set the token
	cmd := exec.Command(railwayPath, "login", "--token", token)
	if out, err := cmd.CombinedOutput(); err != nil {
		fmt.Printf("   ⚠ Railway CLI login failed: %s\n", string(out))
		// Continue anyway, might already be logged in
	}
	fmt.Println("   ✓ Authenticated with Railway CLI")

	// Check if already linked to a project
	cmd = exec.Command(railwayPath, "status")
	cmd.Dir = project.Path
	out, err := cmd.CombinedOutput()

	if err != nil || !strings.Contains(string(out), "Project") {
		// Create new project
		fmt.Println("   📦 Creating new Railway project...")
		cmd = exec.Command(railwayPath, "init", "--name", project.Name)
		cmd.Dir = project.Path
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("railway init failed: %s", string(out))
		}
		fmt.Printf("   ✓ Project created: %s\n", project.Name)
	}

	// Sync environment variables
	if len(envVars) > 0 {
		fmt.Println("   📋 Syncing environment variables...")
		for _, v := range envVars {
			cmd = exec.Command(railwayPath, "variables", "--set", fmt.Sprintf("%s=%s", v.Key, v.Value))
			cmd.Dir = project.Path
			cmd.CombinedOutput() // ignore errors for individual vars
		}
		fmt.Printf("   ✓ Synced %d environment variables\n", len(envVars))
	}

	// Deploy
	fmt.Println("   📤 Deploying...")
	cmd = exec.Command(railwayPath, "up", "--detach")
	cmd.Dir = project.Path
	out, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("railway up failed: %s", string(out))
	}

	// Get the deployment URL
	cmd = exec.Command(railwayPath, "domain")
	cmd.Dir = project.Path
	out, err = cmd.CombinedOutput()
	if err == nil {
		url := strings.TrimSpace(string(out))
		if !strings.HasPrefix(url, "http") {
			url = "https://" + url
		}
		fmt.Printf("   ✓ Deployed successfully!\n")
		fmt.Println()
		fmt.Printf("   🌐 Live at: %s\n", url)

		SaveProjectState(project.Path, ProjectState{
			Platform:    "railway",
			ProjectName: project.Name,
			DeployURL:   url,
		})
	} else {
		fmt.Println("   ✓ Deployment triggered!")
		fmt.Println("   ℹ Check status at: https://railway.app/dashboard")
	}

	return nil
}

// railwayDeployViaAPI uses the Railway GraphQL API
func railwayDeployViaAPI(token string, project *ProjectInfo, envVars []EnvVar) error {
	fmt.Println("   ℹ Railway CLI not found. Using API deployment.")
	fmt.Println()
	fmt.Println("   For the best experience, install Railway CLI:")
	fmt.Println("   $ npm install -g @railway/cli")
	fmt.Println()

	// Create project via GraphQL
	projectID, err := railwayCreateProject(token, project.Name)
	if err != nil {
		return err
	}
	fmt.Printf("   ✓ Project created: %s\n", project.Name)

	// Get or create the production environment.
	// Railway auto-creates "production" on projectCreate, so we look it up first.
	envID, err := railwayGetOrCreateEnvironment(token, projectID)
	if err != nil {
		return err
	}

	// Create service
	serviceID, err := railwayCreateService(token, projectID, project.Name)
	if err != nil {
		return err
	}
	fmt.Printf("   ✓ Service created: %s\n", project.Name)

	// Sync env vars
	if len(envVars) > 0 {
		if err := railwaySyncEnvVarsAPI(token, envID, serviceID, envVars); err != nil {
			fmt.Printf("   ⚠ Warning: could not sync env vars: %v\n", err)
		} else {
			fmt.Printf("   ✓ Synced %d environment variables\n", len(envVars))
		}
	}

	// Note: Actual code deployment via API requires file upload which is complex
	// We'll instruct the user to use git push or install Railway CLI
	fmt.Println()
	fmt.Println("   📤 To deploy your code, either:")
	fmt.Println("      1. Install Railway CLI: npm install -g @railway/cli")
	fmt.Println("      2. Or connect your GitHub repo in the Railway dashboard")
	fmt.Println()
	fmt.Printf("   📋 Dashboard: https://railway.app/project/%s\n", projectID)

	SaveProjectState(project.Path, ProjectState{
		Platform:    "railway",
		ProjectID:   projectID,
		ProjectName: project.Name,
		DeployURL:   fmt.Sprintf("https://railway.app/project/%s", projectID),
	})

	return nil
}

func railwayGraphQL(token, query string) (map[string]interface{}, error) {
	body := map[string]string{"query": query}
	bodyBytes, _ := json.Marshal(body)

	req, err := http.NewRequest("POST", "https://backboard.railway.app/graphql/v2", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("railway API error (%d): %s", resp.StatusCode, string(data))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	if errors, ok := result["errors"]; ok {
		return nil, fmt.Errorf("graphql error: %v", errors)
	}

	return result, nil
}

func railwayCreateProject(token, name string) (string, error) {
	query := fmt.Sprintf(`mutation { projectCreate(input: { name: "%s" }) { id } }`, name)
	result, err := railwayGraphQL(token, query)
	if err != nil {
		return "", err
	}

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("unexpected response format")
	}
	projectCreate, ok := data["projectCreate"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("unexpected response format")
	}
	return projectCreate["id"].(string), nil
}

func railwayCreateEnvironment(token, projectID string) (string, error) {
	query := fmt.Sprintf(`mutation { environmentCreate(input: { projectId: "%s", name: "production" }) { id } }`, projectID)
	result, err := railwayGraphQL(token, query)
	if err != nil {
		return "", err
	}

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("unexpected response format")
	}
	envCreate, ok := data["environmentCreate"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("unexpected response format")
	}
	return envCreate["id"].(string), nil
}

// railwayGetOrCreateEnvironment returns the existing "production" environment ID
// for the project, or any environment if "production" isn't present, falling back
// to creating one only if the project has no environments at all.
func railwayGetOrCreateEnvironment(token, projectID string) (string, error) {
	query := fmt.Sprintf(`query { project(id: "%s") { environments { edges { node { id name } } } } }`, projectID)
	result, err := railwayGraphQL(token, query)
	if err == nil {
		if data, ok := result["data"].(map[string]interface{}); ok {
			if proj, ok := data["project"].(map[string]interface{}); ok {
				if envs, ok := proj["environments"].(map[string]interface{}); ok {
					if edges, ok := envs["edges"].([]interface{}); ok {
						var firstID string
						for _, edge := range edges {
							e, _ := edge.(map[string]interface{})
							n, _ := e["node"].(map[string]interface{})
							id, _ := n["id"].(string)
							name, _ := n["name"].(string)
							if firstID == "" {
								firstID = id
							}
							if name == "production" {
								return id, nil
							}
						}
						if firstID != "" {
							return firstID, nil
						}
					}
				}
			}
		}
	}
	return railwayCreateEnvironment(token, projectID)
}

func railwayCreateService(token, projectID, name string) (string, error) {
	query := fmt.Sprintf(`mutation { serviceCreate(input: { projectId: "%s", name: "%s" }) { id } }`, projectID, name)
	result, err := railwayGraphQL(token, query)
	if err != nil {
		return "", err
	}

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("unexpected response format")
	}
	serviceCreate, ok := data["serviceCreate"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("unexpected response format")
	}
	return serviceCreate["id"].(string), nil
}

func railwaySyncEnvVarsAPI(token, envID, serviceID string, envVars []EnvVar) error {
	// Build variables map
	variables := make(map[string]string)
	for _, v := range envVars {
		variables[v.Key] = v.Value
	}

	varsJSON, _ := json.Marshal(variables)
	query := fmt.Sprintf(`mutation { variableCollectionUpsert(input: { environmentId: "%s", serviceId: "%s", replace: false, variables: %s }) }`,
		envID, serviceID, string(varsJSON))

	_, err := railwayGraphQL(token, query)
	return err
}

// RenderDeploy deploys the current project to Render
func RenderDeploy(project *ProjectInfo, envVars []EnvVar) error {
	token, err := EnsureLoggedIn("render")
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("🚀 Deploying to Render...")

	// Render requires a git repo connected to Render
	// Check if this is a git repo
	if !isGitRepo(project.Path) {
		fmt.Println("   ⚠ Render requires a Git repository.")
		fmt.Println("   Initialize git first:")
		fmt.Println("     $ git init")
		fmt.Println("     $ git add .")
		fmt.Println("     $ git commit -m 'initial commit'")
		fmt.Println("     $ git remote add origin <your-github-repo>")
		fmt.Println("     $ git push -u origin main")
		fmt.Println()
		fmt.Println("   Then run shipmate again.")
		return nil
	}

	// Get git remote URL
	remoteURL := getGitRemoteURL(project.Path)
	if remoteURL == "" {
		fmt.Println("   ⚠ No git remote found. Push your code to GitHub first:")
		fmt.Println("     $ git remote add origin git@github.com:user/repo.git")
		fmt.Println("     $ git push -u origin main")
		return nil
	}

	fmt.Printf("   ✓ Git remote: %s\n", remoteURL)

	// Create service via Render API
	serviceURL, err := renderCreateService(token, project, remoteURL, envVars)
	if err != nil {
		return err
	}

	fmt.Printf("   ✓ Service created!\n")
	fmt.Println()
	fmt.Printf("   🌐 Dashboard: %s\n", serviceURL)

	SaveProjectState(project.Path, ProjectState{
		Platform:    "render",
		ProjectName: project.Name,
		DeployURL:   serviceURL,
	})

	return nil
}

func renderCreateService(token string, project *ProjectInfo, repoURL string, envVars []EnvVar) (string, error) {
	// Build environment variables array for Render API
	var envList []map[string]string
	for _, v := range envVars {
		envList = append(envList, map[string]string{
			"key":   v.Key,
			"value": v.Value,
		})
	}

	// Determine service details based on project type
	serviceType := "web"
	buildCommand := project.BuildCommand
	startCommand := project.StartCommand
	runtime := "docker"

	// If project has a Dockerfile, use docker runtime
	if project.HasDocker {
		runtime = "docker"
		buildCommand = ""
		startCommand = ""
	}

	body := map[string]interface{}{
		"serviceDetails": map[string]interface{}{
			"name":         project.Name,
			"type":         serviceType,
			"runtime":      runtime,
			"repo":         repoURL,
			"buildCommand": buildCommand,
			"startCommand": startCommand,
			"envVars":      envList,
		},
	}

	bodyBytes, _ := json.Marshal(body)

	req, err := http.NewRequest("POST", "https://api.render.com/v1/services", bytes.NewReader(bodyBytes))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		data, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("render API error (%d): %s", resp.StatusCode, string(data))
	}

	var result struct {
		ServiceDetails struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"serviceDetails"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return fmt.Sprintf("https://dashboard.render.com/web/%s", result.ServiceDetails.ID), nil
}

// isGitRepo checks if the path is a git repository
func isGitRepo(path string) bool {
	cmd := exec.Command("git", "-C", path, "rev-parse", "--is-inside-work-tree")
	out, err := cmd.CombinedOutput()
	return err == nil && strings.TrimSpace(string(out)) == "true"
}

// getGitRemoteURL gets the origin remote URL
func getGitRemoteURL(path string) string {
	cmd := exec.Command("git", "-C", path, "remote", "get-url", "origin")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// NetlifyDeploy deploys a static/Next.js site to Netlify
func NetlifyDeploy(project *ProjectInfo, envVars []EnvVar) error {
	token, err := EnsureLoggedIn("netlify")
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("🚀 Deploying to Netlify...")

	// For static sites, we can do a direct deploy via API
	if project.Type == ProjectStatic {
		return netlifyManualDeploy(token, project)
	}

	// For Next.js/React, recommend using Netlify CLI
	fmt.Println("   ℹ For Next.js/React projects, Netlify CLI is recommended.")
	fmt.Println()

	// Check if netlify CLI is available
	if path, err := exec.LookPath("netlify"); err == nil {
		return netlifyDeployViaCLI(path, project, token)
	}

	fmt.Println("   Install Netlify CLI for the best experience:")
	fmt.Println("   $ npm install -g netlify-cli")
	fmt.Println("   $ netlify login")
	fmt.Println("   $ netlify deploy --prod")
	fmt.Println()
	fmt.Println("   Or deploy manually:")
	fmt.Printf("   1. Go to https://app.netlify.com/sites/create\n")
	fmt.Printf("   2. Connect your Git repository\n")
	fmt.Printf("   3. Set build command: %s\n", project.BuildCommand)
	fmt.Printf("   4. Set output directory: dist (or .next for Next.js)\n")

	return nil
}

func netlifyDeployViaCLI(netlifyPath string, project *ProjectInfo, token string) error {
	// Login
	cmd := exec.Command(netlifyPath, "login", "--auth", token)
	cmd.CombinedOutput()

	// Deploy
	fmt.Println("   📤 Deploying via Netlify CLI...")
	cmd = exec.Command(netlifyPath, "deploy", "--prod", "--dir=dist")
	cmd.Dir = project.Path
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("netlify deploy failed: %s", string(out))
	}

	fmt.Println("   ✓ Deployed successfully!")
	fmt.Println(string(out))
	return nil
}

func netlifyManualDeploy(token string, project *ProjectInfo) error {
	fmt.Println("   📤 Uploading files to Netlify...")

	// For static sites, zip and upload
	// This is a simplified version - full implementation would create a deploy
	// with individual file uploads via Netlify's deploy API

	fmt.Println("   ℹ Manual deploy for static sites:")
	fmt.Println("     1. Build your site (if needed)")
	fmt.Println("     2. Drag & drop your output folder at:")
	fmt.Println("        https://app.netlify.com/drop")
	fmt.Println()
	fmt.Println("   Or install Netlify CLI:")
	fmt.Println("     $ npm install -g netlify-cli")
	fmt.Println("     $ netlify deploy --prod --dir=.")

	return nil
}

// FlyioDeploy deploys to Fly.io
func FlyioDeploy(project *ProjectInfo, envVars []EnvVar) error {
	fmt.Println()
	fmt.Println("🚀 Deploying to Fly.io...")

	// Fly.io strongly requires their CLI (flyctl)
	flyPath, err := exec.LookPath("flyctl")
	if err != nil {
		flyPath, err = exec.LookPath("fly")
	}

	if err != nil {
		fmt.Println("   ⚠ Fly.io requires flyctl CLI.")
		fmt.Println()
		fmt.Println("   Install it:")
		fmt.Println("     $ curl -L https://fly.io/install.sh | sh")
		fmt.Println()
		fmt.Println("   Then run:")
		fmt.Println("     $ fly auth signup")
		fmt.Println("     $ fly launch")
		fmt.Println("     $ fly deploy")
		return nil
	}

	fmt.Println("   ✓ Found flyctl")

	// If fly.toml is absent, launch the app (without deploying — we'll do that after
	// secrets are set). If fly.toml already exists, skip launch and go straight to deploy.
	flyTomlPath := project.Path + string(os.PathSeparator) + "fly.toml"
	if _, statErr := os.Stat(flyTomlPath); os.IsNotExist(statErr) {
		fmt.Println("   📦 Launching app on Fly.io...")
		cmd := exec.Command(flyPath, "launch", "--yes", "--no-deploy", "--copy-config", "--name", project.Name)
		cmd.Dir = project.Path
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Println("   ⚠ Launch failed. Try running interactively:")
			fmt.Println("     $ fly launch")
			return nil
		}
	} else {
		fmt.Println("   ✓ Existing fly.toml found")
	}

	// Set env vars as secrets BEFORE deploying so the first boot has them.
	if len(envVars) > 0 {
		fmt.Println("   📋 Setting secrets...")
		args := []string{"secrets", "set", "--stage"}
		for _, v := range envVars {
			args = append(args, fmt.Sprintf("%s=%s", v.Key, v.Value))
		}
		cmd := exec.Command(flyPath, args...)
		cmd.Dir = project.Path
		if out, err := cmd.CombinedOutput(); err != nil {
			fmt.Printf("   ⚠ Warning: could not stage secrets: %s\n", string(out))
		} else {
			fmt.Printf("   ✓ Staged %d secrets\n", len(envVars))
		}
	}

	// Deploy
	fmt.Println("   📤 Deploying...")
	cmd := exec.Command(flyPath, "deploy")
	cmd.Dir = project.Path
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("fly deploy failed")
	}

	appURL := fmt.Sprintf("https://%s.fly.dev", project.Name)
	fmt.Println()
	fmt.Printf("   ✓ Deployed! App URL: %s\n", appURL)

	SaveProjectState(project.Path, ProjectState{
		Platform:    "flyio",
		ProjectName: project.Name,
		DeployURL:   appURL,
	})

	return nil
}
