package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// MonorepoType represents the type of monorepo
type MonorepoType string

const (
	MonorepoTurborepo MonorepoType = "turborepo"
	MonorepoNx        MonorepoType = "nx"
	MonorepoPnpm      MonorepoType = "pnpm-workspace"
	MonorepoLerna     MonorepoType = "lerna"
	MonorepoYarn      MonorepoType = "yarn-workspace"
	MonorepoNone      MonorepoType = "none"
)

// MonorepoInfo holds monorepo detection results
type MonorepoInfo struct {
	Type     MonorepoType
	Apps     []MonorepoPackage
	Packages []MonorepoPackage
}

// MonorepoPackage represents a single package in the monorepo
type MonorepoPackage struct {
	Name   string
	Path   string
	Type   ProjectType
	HasApp bool // has an application entry point
}

// DetectMonorepo checks if the current directory is a monorepo
func DetectMonorepo() (*MonorepoInfo, error) {
	info := &MonorepoInfo{
		Type: MonorepoNone,
	}

	// Check for Turborepo
	if fileExists("turbo.json") {
		info.Type = MonorepoTurborepo
	}

	// Check for Nx
	if fileExists("nx.json") {
		info.Type = MonorepoNx
	}

	// Check for pnpm workspace
	if fileExists("pnpm-workspace.yaml") {
		info.Type = MonorepoPnpm
	}

	// Check for Lerna
	if fileExists("lerna.json") {
		info.Type = MonorepoLerna
	}

	// Check for Yarn workspaces in package.json
	if fileExists("package.json") {
		content, err := readFile("package.json")
		if err == nil && strings.Contains(content, "\"workspaces\"") {
			if info.Type == MonorepoNone {
				info.Type = MonorepoYarn
			}
		}
	}

	if info.Type == MonorepoNone {
		return nil, nil // not a monorepo
	}

	// Scan for apps and packages
	if err := scanMonorepoPackages(info); err != nil {
		return info, err
	}

	return info, nil
}

// scanMonorepoPackages finds all apps and packages in the monorepo
func scanMonorepoPackages(info *MonorepoInfo) error {
	// Common monorepo structures
	searchDirs := []string{
		"apps",
		"packages",
		"services",
		"modules",
	}

	for _, dir := range searchDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		}

		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			pkgPath := filepath.Join(dir, entry.Name())
			pkg := scanPackage(pkgPath)
			if pkg == nil {
				continue
			}

			if dir == "apps" || dir == "services" || pkg.HasApp {
				info.Apps = append(info.Apps, *pkg)
			} else {
				info.Packages = append(info.Packages, *pkg)
			}
		}
	}

	return nil
}

// scanPackage examines a single package directory
func scanPackage(path string) *MonorepoPackage {
	// Read package name from package.json if exists
	name := filepath.Base(path)
	if fileExists(filepath.Join(path, "package.json")) {
		content, err := os.ReadFile(filepath.Join(path, "package.json"))
		if err == nil {
			var pkg map[string]interface{}
			if json.Unmarshal(content, &pkg) == nil {
				if n, ok := pkg["name"].(string); ok {
					name = n
				}
			}
		}
	}

	// Detect project type by scanning the package directory
	origDir, _ := os.Getwd()
	os.Chdir(path)
	defer os.Chdir(origDir)

	projectType := detectType()
	hasApp := hasApplicationEntry()

	return &MonorepoPackage{
		Name:   name,
		Path:   path,
		Type:   projectType,
		HasApp: hasApp,
	}
}

// hasApplicationEntry checks if a package has an app entry point
func hasApplicationEntry() bool {
	entries := []string{
		"main.go", "main.py", "app.py", "manage.py", // Go, Python
		"src/index.ts", "src/index.js", "src/main.ts", "src/main.js", // Node
		"index.ts", "index.js", "server.ts", "server.js", // Node
		"src/App.tsx", "src/App.jsx", // React
	}

	for _, e := range entries {
		if _, err := os.Stat(e); err == nil {
			return true
		}
	}

	return false
}

// DisplayMonorepoInfo shows monorepo detection results
func DisplayMonorepoInfo(info *MonorepoInfo) {
	fmt.Printf("   📦 Detected %s monorepo\n", info.Type)
	fmt.Println()

	if len(info.Apps) > 0 {
		fmt.Println("   Apps:")
		for _, app := range info.Apps {
			fmt.Printf("      📁 %s (%s)\n", app.Path, app.Type)
		}
	}

	if len(info.Packages) > 0 {
		fmt.Println("   Packages:")
		for _, pkg := range info.Packages {
			fmt.Printf("      📁 %s (%s)\n", pkg.Path, pkg.Type)
		}
	}
	fmt.Println()
}

// SelectMonorepoApp prompts user to select which app to deploy
func SelectMonorepoApp(info *MonorepoInfo) (*MonorepoPackage, error) {
	if len(info.Apps) == 0 {
		return nil, fmt.Errorf("no deployable apps found in monorepo")
	}

	if len(info.Apps) == 1 {
		fmt.Printf("   Auto-selecting: %s (%s)\n", info.Apps[0].Path, info.Apps[0].Type)
		return &info.Apps[0], nil
	}

	fmt.Println("   Which app do you want to deploy?")
	for i, app := range info.Apps {
		fmt.Printf("     [%d] %s (%s)\n", i+1, app.Path, app.Type)
	}

	fmt.Print("   Select (number): ")
	var choice int
	fmt.Scanln(&choice)

	if choice < 1 || choice > len(info.Apps) {
		return nil, fmt.Errorf("invalid selection: %d", choice)
	}

	return &info.Apps[choice-1], nil
}
