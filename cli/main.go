package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var (
	// version is set at build time via -ldflags
	version = "dev"
)

var rootCmd = &cobra.Command{
	Use:   "shipmate",
	Short: "Shipmate - The smart deployer for developers",
	Long: `Shipmate detects your project type, generates a Dockerfile, and deploys to the right platform.

Supported languages and frameworks:
  JavaScript/TypeScript: Next.js, React, Vue, Angular, Svelte, Nuxt, Remix, Astro, TanStack, Gatsby, SolidJS, Qwik
  Backend: Go, Node.js, Python (Django/Flask/FastAPI), Ruby (Rails/Sinatra), Java, Kotlin, Spring Boot, Rust, C#/.NET, PHP (Laravel/Symfony), Elixir, Dart/Flutter
  Static: Plain HTML/CSS/JS

Supported platforms:
  - Vercel (Next.js, React, static sites)
  - Railway (Go, Node.js, Python, Docker)
  - Render (Go, Node.js, Python, Ruby)
  - Netlify (static sites, Next.js)
  - Fly.io (Docker, Go, Node.js)
  - Heroku (Go, Node.js, Python, Ruby, Java, PHP)

Example:
  cd my-project
  shipmate`,
	Version: version,
}

var reviewBeforeDeploy bool

var shipCmd = &cobra.Command{
	Use:     "ship",
	Aliases: []string{"deploy", "s"},
	Short:   "Deploy your project (main command)",
	Long: `Scans your project, detects the type, generates Dockerfile if needed,
and deploys to the recommended platform with one command.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		start := time.Now()

		displayBanner()
		fmt.Printf("  Version: %s\n\n", version)

		// Create deployer (handles monorepo detection, project detection, env loading)
		deployer, err := NewDeployer()
		if err != nil {
			return err
		}

		if deployer.Project.Type == ProjectUnknown {
			fmt.Println("   ⚠ Could not detect project type")
			fmt.Println("   Make sure you're in a project directory with recognizable files.")
			fmt.Println()
			fmt.Println("   Supported files:")
			fmt.Println("     go.mod, Cargo.toml, pom.xml, Gemfile, requirements.txt,")
			fmt.Println("     package.json, composer.json, mix.exs, pubspec.yaml, index.html")
			return nil
		}

		fmt.Println("🔍 Project detected:")

		// Optional: Run code review before deploying
		if reviewBeforeDeploy {
			fmt.Println()
			fmt.Println("📖 Running AI code review before deploy...")
			result, err := ReviewCode(deployer.Project)
			if err != nil {
				fmt.Printf("   ⚠ Review failed: %v\n", err)
				fmt.Println("   Continuing with deployment...")
			} else {
				GenerateReviewReport(result)
				DisplayReviewSummary(result)

				// Ask if user wants to continue deploying
				fmt.Print("   Continue with deployment? (y/n): ")
				var answer string
				fmt.Scanln(&answer)
				answer = strings.TrimSpace(strings.ToLower(answer))
				if answer != "y" && answer != "yes" {
					fmt.Println("   ✗ Deployment cancelled.")
					return nil
				}
				fmt.Println()
			}
		}

		// Run the full deployment workflow
		if err := deployer.Run(); err != nil {
			fmt.Printf("\n✗ Deployment failed: %v\n", err)
			return err
		}

		elapsed := time.Since(start)
		fmt.Println()
		fmt.Printf("✓ Done in %s\n", elapsed.Round(time.Millisecond))
		fmt.Println()
		fmt.Println("🎉 Thanks for using Shipmate!")
		fmt.Println("   Share your experience: https://myshipmate.cc")

		return nil
	},
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize Shipmate in your project (generate Dockerfile only)",
	Long: `Generates a Dockerfile and .dockerignore for your project without deploying.
Useful if you want to review the generated files before deploying.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("🔧 Initializing Shipmate...")
		fmt.Println()

		// Detect project
		project, err := DetectProject()
		if err != nil {
			return fmt.Errorf("project detection failed: %w", err)
		}

		if project.Type == ProjectUnknown {
			fmt.Println("   ⚠ Could not detect project type")
			return nil
		}

		fmt.Printf("   ✓ Detected: %s project\n", project.Type)
		fmt.Printf("   ✓ Build command: %s\n", project.BuildCommand)
		fmt.Printf("   ✓ Start command: %s\n", project.StartCommand)
		fmt.Printf("   ✓ Port: %s\n", project.Port)
		fmt.Println()

		// Generate Dockerfile
		if !project.HasDocker {
			if err := GenerateDockerfile(project); err != nil {
				return fmt.Errorf("Dockerfile generation failed: %w", err)
			}
			fmt.Println("   ✓ Dockerfile created")
		} else {
			fmt.Println("   ℹ Dockerfile already exists (skipped)")
		}

		// Generate .dockerignore
		if !project.HasDockerIgnore {
			if err := GenerateDockerIgnore(project); err != nil {
				fmt.Printf("   ⚠ Warning: could not create .dockerignore: %v\n", err)
			} else {
				fmt.Println("   ✓ .dockerignore created")
			}
		} else {
			fmt.Println("   ℹ .dockerignore already exists (skipped)")
		}

		fmt.Println()
		fmt.Println("✓ Shipmate initialized!")
		fmt.Println()
		fmt.Println("Next: Review the Dockerfile, then run:")
		fmt.Println("  $ shipmate ship")
		return nil
	},
}

var loginCmd = &cobra.Command{
	Use:   "login [platform]",
	Short: "Log in to a deployment platform",
	Long: `Authenticate with a deployment platform.
Supported platforms: vercel, railway, render, netlify, flyio, heroku

Example:
  shipmate login vercel
  shipmate login railway`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		platform := args[0]
		return LoginToPlatform(platform)
	},
}

var logoutCmd = &cobra.Command{
	Use:   "logout [platform]",
	Short: "Log out from a deployment platform",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		platform := args[0]
		cfg, err := LoadConfig()
		if err != nil {
			return err
		}

		delete(cfg.Platforms, platform)
		if err := SaveConfig(cfg); err != nil {
			return err
		}

		fmt.Printf("✓ Logged out from %s\n", platform)
		return nil
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current project and deployment status",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("📋 Shipmate Status")
		fmt.Println()

		// Show logged-in platforms
		cfg, err := LoadConfig()
		if err != nil {
			return err
		}

		fmt.Println("Logged in platforms:")
		if len(cfg.Platforms) == 0 {
			fmt.Println("  (none)")
		}
		for platform, auth := range cfg.Platforms {
			maskedToken := maskValue(auth.Token)
			fmt.Printf("  ✓ %s (token: %s)\n", platform, maskedToken)
		}
		fmt.Println()

		// Show current project
		project, err := DetectProject()
		if err != nil {
			fmt.Println("Current project: (error detecting)")
			return nil
		}

		if project.Type == ProjectUnknown {
			fmt.Println("Current project: (unknown type)")
			return nil
		}

		fmt.Printf("Current project:\n")
		fmt.Printf("  Name: %s\n", project.Name)
		fmt.Printf("  Type: %s\n", project.Type)
		fmt.Printf("  Path: %s\n", project.Path)

		// Check for previous deployment
		state, _ := GetProjectState(project.Path)
		if state != nil {
			when := state.LastDeploy
			if t, err := time.Parse(time.RFC3339, state.LastDeploy); err == nil {
				when = t.Local().Format("2006-01-02 15:04 MST")
			}
			if when == "" {
				fmt.Printf("  Last deployed: %s\n", state.Platform)
			} else {
				fmt.Printf("  Last deployed: %s (%s)\n", state.Platform, when)
			}
			fmt.Printf("  URL: %s\n", state.DeployURL)
		}

		// Check monorepo
		mono, _ := DetectMonorepo()
		if mono != nil {
			fmt.Println()
			fmt.Printf("Monorepo: %s\n", mono.Type)
			fmt.Printf("  Apps: %d\n", len(mono.Apps))
			fmt.Printf("  Packages: %d\n", len(mono.Packages))
		}

		return nil
	},
}

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Show detected environment variables",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("📋 Environment Variables")
		fmt.Println()

		files := FindEnvFiles()
		if len(files) == 0 {
			fmt.Println("  No .env files found in current directory")
			return nil
		}

		fmt.Printf("  Found files: %s\n", strings.Join(files, ", "))
		fmt.Println()

		vars, err := LoadAllEnvVars()
		if err != nil {
			return err
		}

		if len(vars) == 0 {
			fmt.Println("  No variables found")
			return nil
		}

		DisplayEnvSummary(vars)
		return nil
	},
}

var reviewCmd = &cobra.Command{
	Use:   "review",
	Short: "AI code review — find bugs, security issues, and bad logic",
	Long: `Reviews your project code using AI to find:
  - Bugs and nil/null pointer issues
  - Security vulnerabilities (SQL injection, XSS, etc.)
  - Performance problems
  - Bad logic patterns and race conditions

Creates a SHIPMATE_REVIEW.md file at the project root with all findings.

Currently FREE during beta. Will require Shipmate Cloud account in the future.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		displayBanner()

		// Detect project
		project, err := DetectProject()
		if err != nil {
			return fmt.Errorf("project detection failed: %w", err)
		}

		if project.Type == ProjectUnknown {
			fmt.Println("   ⚠ Could not detect project type")
			fmt.Println("   Make sure you're in a project directory.")
			return nil
		}

		fmt.Println("🔍 Code Review")
		fmt.Printf("   Project: %s (%s)\n", project.Name, project.Type)
		fmt.Println()

		// Run the review
		result, err := ReviewCode(project)
		if err != nil {
			return fmt.Errorf("review failed: %w", err)
		}

		// Generate report file
		if err := GenerateReviewReport(result); err != nil {
			return fmt.Errorf("report generation failed: %w", err)
		}

		// Display terminal summary
		DisplayReviewSummary(result)

		return nil
	},
}

func main() {
	// Register flags
	shipCmd.Flags().BoolVarP(&reviewBeforeDeploy, "review", "r", false, "Run AI code review before deploying")

	rootCmd.AddCommand(shipCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(logoutCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(envCmd)
	rootCmd.AddCommand(reviewCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
