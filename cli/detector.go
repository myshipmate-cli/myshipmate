package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ProjectType represents the detected project type
type ProjectType string

const (
	// JavaScript/TypeScript Frameworks
	ProjectNextJS    ProjectType = "nextjs"
	ProjectReact     ProjectType = "react"
	ProjectVue       ProjectType = "vue"
	ProjectAngular   ProjectType = "angular"
	ProjectSvelte    ProjectType = "svelte"
	ProjectSvelteKit ProjectType = "sveltekit"
	ProjectTanStack  ProjectType = "tanstack"
	ProjectRemix     ProjectType = "remix"
	ProjectAstro     ProjectType = "astro"
	ProjectNuxt      ProjectType = "nuxt"
	ProjectGatsby    ProjectType = "gatsby"
	ProjectSolidJS   ProjectType = "solidjs"
	ProjectQwik      ProjectType = "qwik"

	// Backend Languages
	ProjectGo     ProjectType = "go"
	ProjectNode   ProjectType = "node"
	ProjectPython ProjectType = "python"
	ProjectRuby   ProjectType = "ruby"
	ProjectJava   ProjectType = "java"
	ProjectKotlin ProjectType = "kotlin"
	ProjectRust   ProjectType = "rust"
	ProjectCSharp ProjectType = "csharp"
	ProjectPHP    ProjectType = "php"
	ProjectElixir ProjectType = "elixir"
	ProjectDart   ProjectType = "dart"

	// Python Frameworks
	ProjectDjango  ProjectType = "django"
	ProjectFlask   ProjectType = "flask"
	ProjectFastAPI ProjectType = "fastapi"

	// Ruby Frameworks
	ProjectRails   ProjectType = "rails"
	ProjectSinatra ProjectType = "sinatra"

	// Java Frameworks
	ProjectSpring ProjectType = "spring"

	// PHP Frameworks
	ProjectLaravel ProjectType = "laravel"
	ProjectSymfony ProjectType = "symfony"

	// Static/Other
	ProjectStatic  ProjectType = "static"
	ProjectDocker  ProjectType = "docker"
	ProjectUnknown ProjectType = "unknown"
)

// ProjectInfo holds detected project information
type ProjectInfo struct {
	Type            ProjectType
	Name            string
	Path            string
	HasDocker       bool
	HasEnvFile      bool
	HasDockerIgnore bool
	BuildCommand    string
	StartCommand    string
	Port            string
	Dependencies    []string
}

// DetectProject scans the current directory and detects the project type
func DetectProject() (*ProjectInfo, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	info := &ProjectInfo{
		Path: cwd,
		Name: filepath.Base(cwd),
	}

	// Check for Dockerfile first
	if fileExists("Dockerfile") {
		info.HasDocker = true
	}

	// Check for .env file
	if fileExists(".env") || fileExists(".env.local") || fileExists(".env.production") {
		info.HasEnvFile = true
	}

	// Check for .dockerignore
	if fileExists(".dockerignore") {
		info.HasDockerIgnore = true
	}

	// Detect project type
	info.Type = detectType()

	// If not found in root, scan subdirectories
	if info.Type == ProjectUnknown {
		subprojects := scanSubdirectories(cwd)
		if len(subprojects) > 0 {
			return handleSubprojects(subprojects)
		}
	}

	// Set default commands based on project type
	setDefaultCommands(info)

	return info, nil
}

// scanSubdirectories walks subdirectories (up to 2 levels) and detects projects
func scanSubdirectories(root string) []*ProjectInfo {
	var projects []*ProjectInfo
	maxDepth := 2

	// Directories to skip
	skipDirs := map[string]bool{
		".git": true, "node_modules": true, "vendor": true,
		"__pycache__": true, ".venv": true, "venv": true,
		"build": true, "dist": true, ".next": true,
		"target": true, ".cache": true, "coverage": true,
	}

	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Skip non-directories
		if !info.IsDir() {
			return nil
		}

		// Calculate depth
		relPath, _ := filepath.Rel(root, path)
		depth := strings.Count(relPath, string(filepath.Separator))
		if relPath == "." {
			depth = 0
		}

		// Stop if too deep
		if depth > maxDepth {
			return filepath.SkipDir
		}

		// Skip certain directories
		if skipDirs[info.Name()] {
			return filepath.SkipDir
		}

		// Skip root directory (already checked)
		if path == root {
			return nil
		}

		// Try to detect project in this directory
		origDir, _ := os.Getwd()
		os.Chdir(path)
		defer os.Chdir(origDir)

		projectType := detectType()
		if projectType != ProjectUnknown {
			projectInfo := &ProjectInfo{
				Type:            projectType,
				Name:            filepath.Base(path),
				Path:            path,
				HasDocker:       fileExists("Dockerfile"),
				HasEnvFile:      fileExists(".env") || fileExists(".env.local") || fileExists(".env.production"),
				HasDockerIgnore: fileExists(".dockerignore"),
			}
			setDefaultCommands(projectInfo)
			projects = append(projects, projectInfo)
		}

		return nil
	})

	return projects
}

// handleSubprojects presents subprojects to user and returns selected one
func handleSubprojects(projects []*ProjectInfo) (*ProjectInfo, error) {
	if len(projects) == 1 {
		// Auto-select the only project
		fmt.Printf("\n📁 Found project in subdirectory: %s/%s\n", filepath.Base(filepath.Dir(projects[0].Path)), projects[0].Name)
		fmt.Printf("   Type: %s\n\n", projects[0].Type)
		
		// Change to that directory
		os.Chdir(projects[0].Path)
		return projects[0], nil
	}

	// Multiple projects found - let user choose
	fmt.Printf("\n📁 Found %d projects in subdirectories:\n\n", len(projects))
	for i, p := range projects {
		relPath, _ := filepath.Rel(mustGetwd(), p.Path)
		fmt.Printf("   [%d] %s (%s)\n", i+1, relPath, p.Type)
	}

	fmt.Print("\n   Select project (number): ")
	var choice int
	fmt.Scanln(&choice)

	if choice < 1 || choice > len(projects) {
		return nil, fmt.Errorf("invalid selection")
	}

	selected := projects[choice-1]
	os.Chdir(selected.Path)
	return selected, nil
}

func mustGetwd() string {
	wd, _ := os.Getwd()
	return wd
}

func detectType() ProjectType {
	// Check for Dockerfile first (if no other project files)
	hasDockerfile := fileExists("Dockerfile")
	hasOtherFiles := fileExists("go.mod") || fileExists("Cargo.toml") ||
		fileExists("pom.xml") || fileExists("build.gradle") || fileExists("Gemfile") ||
		fileExists("requirements.txt") || fileExists("package.json") || fileExists("index.html")

	if hasDockerfile && !hasOtherFiles {
		return ProjectDocker
	}

	// Check for Go
	if fileExists("go.mod") {
		return ProjectGo
	}

	// Check for Rust
	if fileExists("Cargo.toml") {
		return ProjectRust
	}

	// Check for Java/Kotlin
	if fileExists("pom.xml") || fileExists("build.gradle") || fileExists("build.gradle.kts") {
		// Check if it's Spring Boot
		if fileExists("pom.xml") {
			content, _ := readFile("pom.xml")
			if strings.Contains(content, "spring-boot") {
				return ProjectSpring
			}
		}
		if fileExists("build.gradle") || fileExists("build.gradle.kts") {
			content, _ := readFile("build.gradle")
			if content == "" {
				content, _ = readFile("build.gradle.kts")
			}
			if strings.Contains(content, "spring-boot") || strings.Contains(content, "org.springframework.boot") {
				return ProjectSpring
			}
		}

		// Check if Kotlin
		if fileExists("build.gradle.kts") {
			return ProjectKotlin
		}
		return ProjectJava
	}

	// Check for Ruby
	if fileExists("Gemfile") {
		content, _ := readFile("Gemfile")
		if strings.Contains(content, "rails") || strings.Contains(content, "railties") {
			return ProjectRails
		}
		if strings.Contains(content, "sinatra") {
			return ProjectSinatra
		}
		return ProjectRuby
	}

	// Check for Python
	if fileExists("requirements.txt") || fileExists("Pipfile") || fileExists("pyproject.toml") || fileExists("setup.py") {
		// Try to detect specific frameworks
		content := ""
		if fileExists("requirements.txt") {
			content, _ = readFile("requirements.txt")
		} else if fileExists("Pipfile") {
			content, _ = readFile("Pipfile")
		} else if fileExists("pyproject.toml") {
			content, _ = readFile("pyproject.toml")
		}

		if strings.Contains(content, "django") || strings.Contains(content, "Django") {
			return ProjectDjango
		}
		if strings.Contains(content, "fastapi") || strings.Contains(content, "FastAPI") {
			return ProjectFastAPI
		}
		if strings.Contains(content, "flask") || strings.Contains(content, "Flask") {
			return ProjectFlask
		}

		return ProjectPython
	}

	// Check for C#/.NET
	if fileExists("*.csproj") || fileExists("*.sln") {
		return ProjectCSharp
	}

	// Check for PHP
	if fileExists("composer.json") {
		content, _ := readFile("composer.json")
		if strings.Contains(content, "laravel/framework") {
			return ProjectLaravel
		}
		if strings.Contains(content, "symfony") {
			return ProjectSymfony
		}
		return ProjectPHP
	}

	// Check for Elixir
	if fileExists("mix.exs") {
		return ProjectElixir
	}

	// Check for Dart/Flutter
	if fileExists("pubspec.yaml") {
		return ProjectDart
	}

	// Check for Node.js/JavaScript frameworks
	if fileExists("package.json") {
		content, _ := readFile("package.json")

		// Parse package.json to check dependencies
		var pkg map[string]interface{}
		if err := json.Unmarshal([]byte(content), &pkg); err == nil {
			deps := make(map[string]interface{})

			// Merge dependencies and devDependencies
			if d, ok := pkg["dependencies"].(map[string]interface{}); ok {
				for k, v := range d {
					deps[k] = v
				}
			}
			if d, ok := pkg["devDependencies"].(map[string]interface{}); ok {
				for k, v := range d {
					deps[k] = v
				}
			}

			// Check for frameworks in order of specificity
			if _, ok := deps["next"]; ok {
				return ProjectNextJS
			}
			if _, ok := deps["nuxt"]; ok {
				return ProjectNuxt
			}
			if _, ok := deps["@remix-run/react"]; ok {
				return ProjectRemix
			}
			if _, ok := deps["astro"]; ok {
				return ProjectAstro
			}
			if _, ok := deps["@sveltejs/kit"]; ok {
				return ProjectSvelteKit
			}
			if _, ok := deps["svelte"]; ok {
				return ProjectSvelte
			}
			if _, ok := deps["@angular/core"]; ok {
				return ProjectAngular
			}
			if _, ok := deps["gatsby"]; ok {
				return ProjectGatsby
			}
			if _, ok := deps["solid-js"]; ok {
				return ProjectSolidJS
			}
			if _, ok := deps["@builder.io/qwik"]; ok {
				return ProjectQwik
			}
			if _, ok := deps["vue"]; ok {
				return ProjectVue
			}
			if _, ok := deps["@tanstack/react-start"]; ok {
				return ProjectTanStack
			}
			if _, ok := deps["@tanstack/router"]; ok {
				return ProjectTanStack
			}
			if _, ok := deps["react"]; ok {
				return ProjectReact
			}
		}

		return ProjectNode
	}

	// Check for static site
	if fileExists("index.html") {
		return ProjectStatic
	}

	return ProjectUnknown
}

func setDefaultCommands(info *ProjectInfo) {
	switch info.Type {
	case ProjectNextJS:
		info.BuildCommand = "npm run build"
		info.StartCommand = "npm start"
		info.Port = "3000"
	case ProjectReact, ProjectVue, ProjectAngular, ProjectSvelte, ProjectTanStack, ProjectAstro, ProjectGatsby, ProjectSolidJS, ProjectQwik:
		info.BuildCommand = "npm run build"
		info.StartCommand = "npm run preview"
		info.Port = "3000"
	case ProjectNuxt:
		info.BuildCommand = "npm run build"
		info.StartCommand = "node .output/server/index.mjs"
		info.Port = "3000"
	case ProjectRemix:
		info.BuildCommand = "npm run build"
		info.StartCommand = "npm start"
		info.Port = "3000"
	case ProjectSvelteKit:
		info.BuildCommand = "npm run build"
		info.StartCommand = "node build/index.js"
		info.Port = "3000"
	case ProjectGo:
		info.BuildCommand = "go build -o server ."
		info.StartCommand = "./server"
		info.Port = "8080"
	case ProjectNode:
		info.BuildCommand = ""
		info.StartCommand = "node index.js"
		info.Port = "3000"
	case ProjectPython:
		info.BuildCommand = ""
		info.StartCommand = "python app.py"
		info.Port = "8000"
	case ProjectDjango:
		info.BuildCommand = "python manage.py collectstatic --noinput"
		info.StartCommand = "gunicorn config.wsgi:application"
		info.Port = "8000"
	case ProjectFlask:
		info.BuildCommand = ""
		info.StartCommand = "gunicorn app:app"
		info.Port = "8000"
	case ProjectFastAPI:
		info.BuildCommand = ""
		info.StartCommand = "uvicorn main:app --host 0.0.0.0 --port 8000"
		info.Port = "8000"
	case ProjectRuby:
		info.BuildCommand = ""
		info.StartCommand = "ruby app.rb"
		info.Port = "4567"
	case ProjectRails:
		info.BuildCommand = "bundle exec rails assets:precompile"
		info.StartCommand = "bundle exec rails server -b 0.0.0.0"
		info.Port = "3000"
	case ProjectSinatra:
		info.BuildCommand = ""
		info.StartCommand = "ruby app.rb"
		info.Port = "4567"
	case ProjectJava, ProjectKotlin:
		info.BuildCommand = "mvn clean package"
		info.StartCommand = "java -jar target/app.jar"
		info.Port = "8080"
	case ProjectSpring:
		info.BuildCommand = "mvn clean package"
		info.StartCommand = "java -jar target/app.jar"
		info.Port = "8080"
	case ProjectRust:
		info.BuildCommand = "cargo build --release"
		info.StartCommand = "./target/release/app"
		info.Port = "8080"
	case ProjectCSharp:
		info.BuildCommand = "dotnet build"
		info.StartCommand = "dotnet run"
		info.Port = "5000"
	case ProjectPHP:
		info.BuildCommand = ""
		info.StartCommand = "php -S 0.0.0.0:8000"
		info.Port = "8000"
	case ProjectLaravel:
		info.BuildCommand = "php artisan optimize"
		info.StartCommand = "php artisan serve --host=0.0.0.0 --port=8000"
		info.Port = "8000"
	case ProjectSymfony:
		info.BuildCommand = ""
		info.StartCommand = "php bin/console server:start 0.0.0.0:8000"
		info.Port = "8000"
	case ProjectElixir:
		info.BuildCommand = "mix deps.get && mix compile"
		info.StartCommand = "mix phx.server"
		info.Port = "4000"
	case ProjectDart:
		info.BuildCommand = "flutter build web"
		info.StartCommand = "dhttpd --path build/web --port 8080"
		info.Port = "8080"
	case ProjectStatic:
		info.BuildCommand = ""
		info.StartCommand = ""
		info.Port = "80"
	}
}

func fileExists(filename string) bool {
	// Handle glob patterns
	if strings.Contains(filename, "*") {
		matches, _ := filepath.Glob(filename)
		return len(matches) > 0
	}

	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

func readFile(filename string) (string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
