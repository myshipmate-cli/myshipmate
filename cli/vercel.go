package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// VercelDeploy deploys the current project to Vercel
func VercelDeploy(project *ProjectInfo, envVars []EnvVar) error {
	token, err := EnsureLoggedIn("vercel")
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("🚀 Deploying to Vercel...")

	// Step 1: Create or get project
	projectID, err := vercelGetOrCreateProject(token, project.Name)
	if err != nil {
		return fmt.Errorf("project setup failed: %w", err)
	}
	fmt.Printf("   ✓ Project: %s\n", project.Name)

	// Step 2: Sync environment variables
	if len(envVars) > 0 {
		if err := vercelSyncEnvVars(token, projectID, envVars); err != nil {
			fmt.Printf("   ⚠ Warning: could not sync all env vars: %v\n", err)
		} else {
			fmt.Printf("   ✓ Synced %d environment variables\n", len(envVars))
		}
	}

	// Step 3: Upload files and create deployment
	deployURL, err := vercelCreateDeployment(token, projectID, project)
	if err != nil {
		return fmt.Errorf("deployment failed: %w", err)
	}

	fmt.Printf("   ✓ Deployed successfully!\n")
	fmt.Println()
	fmt.Printf("   🌐 Live at: %s\n", deployURL)

	// Save state
	SaveProjectState(project.Path, ProjectState{
		Platform:    "vercel",
		ProjectID:   projectID,
		ProjectName: project.Name,
		DeployURL:   deployURL,
	})

	return nil
}

func vercelGetOrCreateProject(token, projectName string) (string, error) {
	// Paginate through all projects to find existing one
	var allProjects []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	nextTimestamp := int64(0)
	pageSize := 100

	for {
		url := fmt.Sprintf("https://api.vercel.com/v9/projects?limit=%d", pageSize)
		if nextTimestamp > 0 {
			url += fmt.Sprintf("&until=%d", nextTimestamp)
		}

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return "", err
		}
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return "", err
		}

		if resp.StatusCode != 200 {
			resp.Body.Close()
			return "", fmt.Errorf("vercel API error: %d", resp.StatusCode)
		}

		var result struct {
			Projects []struct {
				ID        string `json:"id"`
				Name      string `json:"name"`
				CreatedAt int64  `json:"createdAt"`
			} `json:"projects"`
			Pagination struct {
				Next int64 `json:"next"`
			} `json:"pagination"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			resp.Body.Close()
			return "", err
		}
		resp.Body.Close()

		// Check each project as we go (early exit if found)
		for _, p := range result.Projects {
			if strings.EqualFold(p.Name, projectName) {
				return p.ID, nil
			}
			allProjects = append(allProjects, struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			}{p.ID, p.Name})
		}

		// No more pages
		if result.Pagination.Next == 0 || len(result.Projects) < pageSize {
			break
		}
		nextTimestamp = result.Pagination.Next
	}

	// Project not found — create it
	body := map[string]string{"name": projectName}
	bodyBytes, _ := json.Marshal(body)

	req, err := http.NewRequest("POST", "https://api.vercel.com/v10/projects", bytes.NewReader(bodyBytes))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		bodyData, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("create project failed (%d): %s", resp.StatusCode, string(bodyData))
	}

	var created struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		return "", err
	}

	return created.ID, nil
}
func vercelSyncEnvVars(token, projectID string, envVars []EnvVar) error {
	for _, v := range envVars {
		body := map[string]interface{}{
			"key":    v.Key,
			"value":  v.Value,
			"target": []string{"production", "preview", "development"},
			"type":   "encrypted",
		}
		bodyBytes, _ := json.Marshal(body)

		req, err := http.NewRequest("POST",
			fmt.Sprintf("https://api.vercel.com/v10/projects/%s/env", projectID),
			bytes.NewReader(bodyBytes))
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		resp.Body.Close()

		if resp.StatusCode != 200 && resp.StatusCode != 201 && resp.StatusCode != 409 {
			return fmt.Errorf("env var %s failed: %d", v.Key, resp.StatusCode)
		}
	}
	return nil
}

func vercelCreateDeployment(token, projectID string, project *ProjectInfo) (string, error) {
	fmt.Println("   📤 Uploading project files...")

	files, err := collectFiles(project.Path)
	if err != nil {
		return "", fmt.Errorf("file collection failed: %w", err)
	}

	fmt.Printf("   ✓ Found %d files to upload\n", len(files))

	var fileDigests []map[string]string
	for _, f := range files {
		digest, err := vercelUploadFile(token, f.relPath, f.data)
		if err != nil {
			return "", fmt.Errorf("upload %s failed: %w", f.relPath, err)
		}
		fileDigests = append(fileDigests, map[string]string{
			"file": f.relPath,
			"sha":  digest,
		})
	}

	deployBody := map[string]interface{}{
		"name":    project.Name,
		"project": projectID,
		"target":  "production",
		"files":   fileDigests,
	}

	if project.Type == ProjectNextJS {
		deployBody["buildCommand"] = "npm run build"
		deployBody["outputDirectory"] = ".next"
		deployBody["framework"] = "nextjs"
	} else if project.BuildCommand != "" {
		deployBody["buildCommand"] = project.BuildCommand
	}

	bodyBytes, _ := json.Marshal(deployBody)

	req, err := http.NewRequest("POST", "https://api.vercel.com/v13/deployments", bytes.NewReader(bodyBytes))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		bodyData, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("deployment failed (%d): %s", resp.StatusCode, string(bodyData))
	}

	var deployment struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&deployment); err != nil {
		return "", err
	}

	url := deployment.URL
	if !strings.HasPrefix(url, "http") {
		url = "https://" + url
	}

	return url, nil
}

// VercelFileInfo represents a file to upload
type VercelFileInfo struct {
	path    string
	relPath string
	data    []byte
}

func collectFiles(rootDir string) ([]VercelFileInfo, error) {
	var files []VercelFileInfo
	ignorePatterns := loadIgnorePatterns(rootDir)

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			if shouldIgnoreDir(info.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if info.Size() > 100*1024*1024 {
			return nil
		}

		relPath, err := filepath.Rel(rootDir, path)
		if err != nil {
			return nil
		}
		if shouldIgnore(relPath, ignorePatterns) {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		files = append(files, VercelFileInfo{
			path:    path,
			relPath: relPath,
			data:    data,
		})
		return nil
	})

	return files, err
}

func shouldIgnoreDir(name string) bool {
	ignored := []string{
		"node_modules", ".git", ".next", ".nuxt", "dist", "build",
		"__pycache__", ".venv", "venv", "target", ".gradle",
		".idea", ".vscode", "vendor", ".cache",
	}
	for _, ig := range ignored {
		if name == ig {
			return true
		}
	}
	return false
}

func loadIgnorePatterns(rootDir string) []string {
	var patterns []string
	files := []string{".gitignore", ".dockerignore", ".vercelignore"}
	for _, f := range files {
		content, err := os.ReadFile(filepath.Join(rootDir, f))
		if err != nil {
			continue
		}
		for _, line := range strings.Split(string(content), "\n") {
			line = strings.TrimSpace(line)
			if line != "" && !strings.HasPrefix(line, "#") {
				patterns = append(patterns, line)
			}
		}
	}
	return patterns
}

func shouldIgnore(path string, patterns []string) bool {
	for _, pattern := range patterns {
		if matched, _ := filepath.Match(pattern, path); matched {
			return true
		}
		if matched, _ := filepath.Match(pattern, filepath.Base(path)); matched {
			return true
		}
		if strings.HasPrefix(path, pattern+"/") {
			return true
		}
	}
	return false
}

func vercelUploadFile(token, filePath string, data []byte) (string, error) {
	h := sha1.New()
	h.Write(data)
	digest := hex.EncodeToString(h.Sum(nil))

	// Check if already uploaded
	checkURL := fmt.Sprintf("https://api.vercel.com/v2/now/files/%s", digest)
	req, _ := http.NewRequest("GET", checkURL, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err == nil {
		resp.Body.Close()
		if resp.StatusCode == 200 {
			return digest, nil
		}
	}

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile("file", filePath)
	if err != nil {
		return "", err
	}
	part.Write(data)
	writer.Close()

	req, err = http.NewRequest("POST", "https://api.vercel.com/v2/now/files", &buf)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("x-vercel-digest", digest)

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		bodyData, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("upload failed (%d): %s", resp.StatusCode, string(bodyData))
	}

	return digest, nil
}
