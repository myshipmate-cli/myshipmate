package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Deployer orchestrates the full deployment workflow
type Deployer struct {
	Project  *ProjectInfo
	EnvVars  []EnvVar
	Monorepo *MonorepoInfo
}

// NewDeployer creates a new Deployer with full project analysis
func NewDeployer() (*Deployer, error) {
	d := &Deployer{}

	// Step 1: Detect monorepo first
	mono, err := DetectMonorepo()
	if err != nil {
		return nil, fmt.Errorf("monorepo detection failed: %w", err)
	}
	d.Monorepo = mono

	// Step 2: If monorepo, let user select an app
	if mono != nil {
		DisplayMonorepoInfo(mono)

		selectedApp, err := SelectMonorepoApp(mono)
		if err != nil {
			return nil, err
		}

		// Change into the selected app directory
		origDir, _ := os.Getwd()
		if err := os.Chdir(selectedApp.Path); err != nil {
			return nil, fmt.Errorf("could not change to %s: %w", selectedApp.Path, err)
		}
		defer os.Chdir(origDir)

		// Re-detect project in the selected app
		project, err := DetectProject()
		if err != nil {
			return nil, err
		}
		d.Project = project
	} else {
		// Step 3: Detect project normally
		project, err := DetectProject()
		if err != nil {
			return nil, err
		}
		d.Project = project
	}

	// Step 4: Load environment variables
	envVars, err := LoadAllEnvVars()
	if err != nil {
		return nil, fmt.Errorf("env loading failed: %w", err)
	}
	d.EnvVars = envVars

	return d, nil
}

// Run executes the full deployment workflow
func (d *Deployer) Run() error {
	project := d.Project

	// Display project info
	fmt.Printf("   ✓ Project: %s\n", project.Name)
	fmt.Printf("   ✓ Type: %s\n", project.Type)
	if project.BuildCommand != "" {
		fmt.Printf("   ✓ Build: %s\n", project.BuildCommand)
	}
	if project.StartCommand != "" {
		fmt.Printf("   ✓ Start: %s\n", project.StartCommand)
	}
	if project.Port != "" {
		fmt.Printf("   ✓ Port: %s\n", project.Port)
	}
	if project.HasDocker {
		fmt.Println("   ✓ Existing Dockerfile found")
	}
	if project.HasEnvFile {
		fmt.Println("   ✓ Environment file detected")
	}
	fmt.Println()

	// Generate Dockerfile if needed
	if !project.HasDocker {
		fmt.Println("📦 Generating Dockerfile...")
		if err := GenerateDockerfile(project); err != nil {
			return fmt.Errorf("Dockerfile generation failed: %w", err)
		}
		fmt.Println("   ✓ Dockerfile created")

		if !project.HasDockerIgnore {
			if err := GenerateDockerIgnore(project); err != nil {
				fmt.Printf("   ⚠ Warning: could not create .dockerignore: %v\n", err)
			} else {
				fmt.Println("   ✓ .dockerignore created")
			}
		}
		fmt.Println()
	}

	// Check Docker (informational only)
	hasDocker := checkDockerInstalled()
	if hasDocker {
		fmt.Println("🐳 Docker detected locally (optional — platforms build remotely)")
	} else {
		fmt.Println("ℹ️  No local Docker needed — deployment platforms build your Dockerfile")
	}
	fmt.Println()

	// Display env vars
	if len(d.EnvVars) > 0 {
		DisplayEnvSummary(d.EnvVars)
		fmt.Println()
	}

	// Check for previous deployment state
	prevState, _ := GetProjectState(project.Path)

	// Select platform
	platform, err := d.selectPlatform(prevState)
	if err != nil {
		return err
	}

	// Check git status (important for most platforms)
	if !isGitRepo(project.Path) {
		fmt.Println()
		fmt.Println("⚠️  Most platforms deploy from your Git repository.")
		fmt.Print("   Initialize git now? (y/n): ")

		reader := bufio.NewReader(os.Stdin)
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))

		if answer == "y" || answer == "yes" {
			if err := initGitRepo(project.Path); err != nil {
				fmt.Printf("   ⚠ Git init failed: %v\n", err)
			} else {
				fmt.Println("   ✓ Git repository initialized")
			}
		}
	}

	// Deploy!
	switch platform {
	case "vercel":
		return VercelDeploy(project, d.EnvVars)
	case "railway":
		return RailwayDeploy(project, d.EnvVars)
	case "render":
		return RenderDeploy(project, d.EnvVars)
	case "netlify":
		return NetlifyDeploy(project, d.EnvVars)
	case "flyio":
		return FlyioDeploy(project, d.EnvVars)
	case "heroku":
		return HerokuDeploy(project, d.EnvVars)
	default:
		return fmt.Errorf("unknown platform: %s", platform)
	}
}

// selectPlatform shows platform options and returns the selected one
func (d *Deployer) selectPlatform(prevState *ProjectState) (string, error) {
	platforms := GetRecommendedPlatforms(d.Project.Type)

	if len(platforms) == 0 {
		return "", fmt.Errorf("no platforms available for project type: %s", d.Project.Type)
	}

	fmt.Println("🎯 Recommended platforms:")
	for i, p := range platforms {
		marker := "  "
		if i == 0 {
			marker = "❯ "
		}
		extra := ""
		if prevState != nil && prevState.Platform == string(p.Name) {
			extra = " (last deployed here)"
		}
		fmt.Printf("  %s[%d] %s%s\n", marker, i+1, p.DisplayName, extra)
	}

	// Add option to redeploy to previous platform
	if prevState != nil {
		fmt.Println()
		fmt.Printf("   [0] Redeploy to %s (previous)\n", prevState.Platform)
	}

	fmt.Println()
	fmt.Print("   Select platform (number): ")

	var choice int
	fmt.Scanln(&choice)

	// Re-deploy to previous
	if choice == 0 && prevState != nil {
		return prevState.Platform, nil
	}

	if choice < 1 || choice > len(platforms) {
		// Default to first recommended
		fmt.Printf("   ℹ Invalid choice, using %s\n", platforms[0].DisplayName)
		return string(platforms[0].Name), nil
	}

	return string(platforms[choice-1].Name), nil
}

// initGitRepo initializes a git repository
func initGitRepo(path string) error {
	cmds := [][]string{
		{"git", "-C", path, "init"},
		{"git", "-C", path, "add", "."},
		{"git", "-C", path, "commit", "-m", "Initial commit (by Shipmate)"},
	}

	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("%s failed: %s", args[0], string(out))
		}
	}

	return nil
}

// SelectPlatformForDeploy handles the `shipmate deploy <platform>` command
func SelectPlatformForDeploy(platformArg string, project *ProjectInfo, envVars []EnvVar) error {
	platform := strings.ToLower(platformArg)

	switch platform {
	case "vercel":
		return VercelDeploy(project, envVars)
	case "railway":
		return RailwayDeploy(project, envVars)
	case "render":
		return RenderDeploy(project, envVars)
	case "netlify":
		return NetlifyDeploy(project, envVars)
	case "flyio", "fly":
		return FlyioDeploy(project, envVars)
	case "heroku":
		return HerokuDeploy(project, envVars)
	default:
		return fmt.Errorf("unknown platform: %s\n\nAvailable platforms:\n  - vercel\n  - railway\n  - render\n  - netlify\n  - flyio\n  - heroku", platform)
	}
}
